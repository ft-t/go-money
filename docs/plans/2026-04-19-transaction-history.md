# Transaction History Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Append-only audit log for every transaction mutation (create, update, delete, rule-applied, bulk, import, scheduler), surfaced as a timeline on the transaction details page with per-event JSON diff.

**Architecture:**
- New table `transaction_history` stores one row per mutation event with `snapshot` (full tx JSONB) + `diff` (RFC 6902 patch vs previous event).
- Actor passed through `context.Context` (USER set by gRPC middleware, IMPORTER/SCHEDULER set at entry, BULK at handler, RULE set inline by rule executor).
- New ConnectRPC service `TransactionHistoryService.ListHistory(transaction_id)` returns ordered events.
- Frontend extends existing `TransactionsDetailsComponent` with a Timeline tab; uses `jsondiffpatch` for visual diff rendering.

**Tech Stack:**
- Go: `github.com/wI2L/jsondiff` (RFC 6902 producer), GORM, gormigrate (existing).
- Proto: new `gomoneypb/transactions/history/v1/history.proto` in `xskydev/go-money-pb` (separate repo).
- Frontend: Angular 21, PrimeNG 21 (`Timeline`, `Tabs`), `jsondiffpatch` for HTML diff.

**Important conventions (from CLAUDE.md):**
- Mocks: `//go:generate mockgen` per package. Run `make generate` after interface edits.
- Tests: real DB, `-p 1`, `-timeout 60s`, `Db_Host=tools.lan ReadonlyDb_Host=tools.lan Redis_Host=tools.lan` if no `config.dev.json`.
- No `if` in tests. Separate success/failure tables.
- Use minimal local interfaces (consumer side), not fat shared `interfaces.go`.
- Verification gate: `make lint`, `go test -p 1 ./modified/...`, `go build ./...`.

---

## Phase 0 — Proto definitions (out-of-band)

**Goal:** Define proto messages in `buf.build/xskydev/go-money-pb`. **User pushes this repo separately**, then plan continues with `make update-pb`.

**File to add in `go-money-pb` repo:** `gomoneypb/transactions/history/v1/history.proto`

```proto
syntax = "proto3";

package gomoneypb.transactions.history.v1;

import "google/protobuf/timestamp.proto";
import "google/protobuf/struct.proto";

enum TransactionHistoryEventType {
  TRANSACTION_HISTORY_EVENT_TYPE_UNSPECIFIED = 0;
  TRANSACTION_HISTORY_EVENT_TYPE_CREATED      = 1;
  TRANSACTION_HISTORY_EVENT_TYPE_UPDATED      = 2;
  TRANSACTION_HISTORY_EVENT_TYPE_DELETED      = 3;
  TRANSACTION_HISTORY_EVENT_TYPE_RULE_APPLIED = 4;
}

enum TransactionHistoryActorType {
  TRANSACTION_HISTORY_ACTOR_TYPE_UNSPECIFIED = 0;
  TRANSACTION_HISTORY_ACTOR_TYPE_USER        = 1;
  TRANSACTION_HISTORY_ACTOR_TYPE_RULE        = 2;
  TRANSACTION_HISTORY_ACTOR_TYPE_SCHEDULER   = 3;
  TRANSACTION_HISTORY_ACTOR_TYPE_IMPORTER    = 4;
  TRANSACTION_HISTORY_ACTOR_TYPE_BULK        = 5;
}

message TransactionHistoryEvent {
  int64                          id              = 1;
  int64                          transaction_id  = 2;
  TransactionHistoryEventType    event_type      = 3;
  TransactionHistoryActorType    actor_type      = 4;
  optional int32                 actor_user_id   = 5;
  optional int32                 actor_rule_id   = 6;
  optional string                actor_extra     = 7;  // importer name / bulk op kind / scheduler ctx
  google.protobuf.Struct         snapshot        = 8;  // tx state at event time (allow-listed fields)
  // RFC 6902 JSON Patch ops vs previous event's snapshot.
  // Empty for first CREATED event.
  google.protobuf.Struct         diff            = 9;  // {"ops": [...]} wrapper for Struct compatibility
  google.protobuf.Timestamp      occurred_at     = 10;
}

message ListTransactionHistoryRequest {
  int64 transaction_id = 1;
}

message ListTransactionHistoryResponse {
  repeated TransactionHistoryEvent events = 1;  // ascending occurred_at
}

service TransactionHistoryService {
  rpc ListHistory(ListTransactionHistoryRequest) returns (ListTransactionHistoryResponse);
}
```

**Step 1:** Hand the proto block above to the user. Wait for them to push to `xskydev/go-money-pb` and tell us the new module version.

**Step 2:** Update Go module:
```bash
make update-pb
```
Expected: `go.mod` reports new `buf.build/gen/go/xskydev/go-money-pb` version. Build still passes.

**Step 3:** Update frontend module:
```bash
cd frontend && npm install @buf/xskydev_go-money-pb.bufbuild_es@latest @buf/xskydev_go-money-pb.connectrpc_es@latest
```
Expected: `package.json` updated. Restart `ng serve` after this (per `CLAUDE.md` — webpack caches resolution).

**STOP.** Plan resumes after pb is merged + pulled.

---

## Task 1 — DB migration

**Files:**
- Modify: `pkg/database/migrations.go` (append migration block at end of `getMigrations`).

**Step 1: Add migration block**

Append after the last entry (`2026-04-19-AddAccountTagIds`):

```go
{
    ID: "2026-04-19-AddTransactionHistory",
    Migrate: func(db *gorm.DB) error {
        return boilerplate.ExecuteSql(db,
            `CREATE TABLE IF NOT EXISTS transaction_history (
                id              BIGSERIAL PRIMARY KEY,
                transaction_id  BIGINT      NOT NULL,
                event_type      SMALLINT    NOT NULL,
                actor_type      SMALLINT    NOT NULL,
                actor_user_id   INT,
                actor_rule_id   INT,
                actor_extra     TEXT,
                snapshot        JSONB       NOT NULL,
                diff            JSONB,
                occurred_at     TIMESTAMP   NOT NULL DEFAULT now()
            );`,
            `CREATE INDEX IF NOT EXISTS idx_tx_history_tx_id_occurred
                ON transaction_history (transaction_id, occurred_at);`,
            `CREATE INDEX IF NOT EXISTS idx_tx_history_actor_rule
                ON transaction_history (actor_rule_id) WHERE actor_rule_id IS NOT NULL;`,
        )
    },
},
```

**Step 2: Build**

Run: `go build ./...`
Expected: PASS.

**Step 3: Verify migration applies on a fresh DB**

Run: `Db_Host=tools.lan ReadonlyDb_Host=tools.lan Redis_Host=tools.lan AUTO_CREATE_CI_DB=true go test -p 1 -timeout 60s -run TestMigrations ./pkg/database/...`
Expected: PASS, no migration errors.
*If no `TestMigrations` exists, run any existing `pkg/database` test that opens the DB — the migration runs on connect.*

**Step 4: Commit**

```bash
git add pkg/database/migrations.go
git commit -S -m "feat(db): add transaction_history table"
```

---

## Task 2 — Database model

**Files:**
- Create: `pkg/database/transaction_history.go`

**Step 1: Write model**

```go
package database

import "time"

type TransactionHistoryEventType int16

const (
    TransactionHistoryEventTypeCreated     TransactionHistoryEventType = 1
    TransactionHistoryEventTypeUpdated     TransactionHistoryEventType = 2
    TransactionHistoryEventTypeDeleted     TransactionHistoryEventType = 3
    TransactionHistoryEventTypeRuleApplied TransactionHistoryEventType = 4
)

type TransactionHistoryActorType int16

const (
    TransactionHistoryActorTypeUser      TransactionHistoryActorType = 1
    TransactionHistoryActorTypeRule      TransactionHistoryActorType = 2
    TransactionHistoryActorTypeScheduler TransactionHistoryActorType = 3
    TransactionHistoryActorTypeImporter  TransactionHistoryActorType = 4
    TransactionHistoryActorTypeBulk      TransactionHistoryActorType = 5
)

type TransactionHistory struct {
    ID            int64
    TransactionID int64
    EventType     TransactionHistoryEventType `gorm:"type:smallint"`
    ActorType     TransactionHistoryActorType `gorm:"type:smallint"`
    ActorUserID   *int32
    ActorRuleID   *int32
    ActorExtra    *string
    Snapshot      map[string]any `gorm:"type:jsonb;serializer:json"`
    Diff          map[string]any `gorm:"type:jsonb;serializer:json"` // {"ops": [...]}, nil for first event
    OccurredAt    time.Time
}

func (TransactionHistory) TableName() string { return "transaction_history" }
```

**Step 2: Build**

Run: `go build ./...`
Expected: PASS.

**Step 3: Commit**

```bash
git add pkg/database/transaction_history.go
git commit -S -m "feat(db): add TransactionHistory model"
```

---

## Task 3 — History package skeleton (interfaces + actor context)

**Files:**
- Create: `pkg/transactions/history/interfaces.go`
- Create: `pkg/transactions/history/actor.go`
- Create: `pkg/transactions/history/types.go`

**Step 1: Define actor + context helpers** (`actor.go`)

```go
package history

import (
    "context"

    "github.com/ft-t/go-money/pkg/database"
)

type Actor struct {
    Type        database.TransactionHistoryActorType
    UserID      *int32
    RuleID      *int32
    Extra       string // importer name / bulk op kind / scheduler ctx
}

type actorCtxKey struct{}

func WithActor(ctx context.Context, a Actor) context.Context {
    return context.WithValue(ctx, actorCtxKey{}, a)
}

func ActorFromContext(ctx context.Context) (Actor, bool) {
    a, ok := ctx.Value(actorCtxKey{}).(Actor)
    return a, ok
}

func UserActor(userID int32) Actor {
    return Actor{Type: database.TransactionHistoryActorTypeUser, UserID: &userID}
}

func ImporterActor(name string) Actor {
    return Actor{Type: database.TransactionHistoryActorTypeImporter, Extra: name}
}

func SchedulerActor(ruleID int32) Actor {
    return Actor{Type: database.TransactionHistoryActorTypeScheduler, RuleID: &ruleID}
}

func BulkActor(userID int32, op string) Actor {
    return Actor{Type: database.TransactionHistoryActorTypeBulk, UserID: &userID, Extra: op}
}

func RuleActor(ruleID int32) Actor {
    return Actor{Type: database.TransactionHistoryActorTypeRule, RuleID: &ruleID}
}
```

**Step 2: Define types** (`types.go`)

```go
package history

import (
    "github.com/ft-t/go-money/pkg/database"
)

// HistoryExcludedFields lists tx columns NOT captured in snapshot/diff.
// updated_at: noisy. created_at/id: immutable. *_in_base_currency: FX-recompute noise.
// deleted_at: delete is its own event.
var HistoryExcludedFields = map[string]struct{}{
    "id":                                   {},
    "created_at":                           {},
    "updated_at":                           {},
    "deleted_at":                           {},
    "source_amount_in_base_currency":       {},
    "destination_amount_in_base_currency":  {},
}

type RecordRequest struct {
    Tx        *database.Transaction
    Previous  *database.Transaction // nil for CREATED
    EventType database.TransactionHistoryEventType
    Actor     Actor
}
```

**Step 3: Define interfaces** (`interfaces.go`)

```go
package history

import (
    "context"

    "github.com/ft-t/go-money/pkg/database"
    "gorm.io/gorm"
)

//go:generate mockgen -destination interfaces_mocks_test.go -package history_test -source=interfaces.go

// Recorder writes a single mutation event to history. Idempotent per call.
type Recorder interface {
    Record(ctx context.Context, tx *gorm.DB, req RecordRequest) error
}

// Reader returns events for a transaction in chronological order.
type Reader interface {
    List(ctx context.Context, transactionID int64) ([]*database.TransactionHistory, error)
}
```

**Step 4: Build**

Run: `go build ./...`
Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/transactions/history/
git commit -S -m "feat(history): scaffold history package interfaces + actor context"
```

---

## Task 4 — Snapshot + diff computation

**Files:**
- Create: `pkg/transactions/history/snapshot.go`
- Create: `pkg/transactions/history/snapshot_test.go`

**Step 1: Add jsondiff dep**

Run: `go get github.com/wI2L/jsondiff@latest && go mod tidy`
Expected: dep added to go.mod.

**Step 2: Write failing tests for snapshot field exclusion** (`snapshot_test.go`)

```go
package history_test

import (
    "testing"
    "time"

    "github.com/ft-t/go-money/pkg/database"
    "github.com/ft-t/go-money/pkg/transactions/history"
    "github.com/lib/pq"
    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

type snapshotCase struct {
    name     string
    tx       *database.Transaction
    expectIn []string
    expectOut []string
}

func TestSnapshot_Success(t *testing.T) {
    cases := []snapshotCase{
        {
            name: "all fields populated",
            tx: &database.Transaction{
                ID:                  42,
                Title:               "lunch",
                SourceAmount:        decimal.NewNullDecimal(decimal.NewFromInt(1)),
                SourceCurrency:      "USD",
                CreatedAt:           time.Now(),
                UpdatedAt:           time.Now(),
                TagIDs:              pq.Int32Array{1, 2},
                TransactionDateTime: time.Now(),
                TransactionDateOnly: time.Now(),
            },
            expectIn:  []string{"title", "source_amount", "source_currency", "tag_ids"},
            expectOut: []string{"id", "created_at", "updated_at", "deleted_at",
                "source_amount_in_base_currency", "destination_amount_in_base_currency"},
        },
    }
    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            snap, err := history.Snapshot(tc.tx)
            require.NoError(t, err)
            for _, k := range tc.expectIn {
                _, ok := snap[k]
                assert.True(t, ok, "expected key %q in snapshot", k)
            }
            for _, k := range tc.expectOut {
                _, ok := snap[k]
                assert.False(t, ok, "expected key %q NOT in snapshot", k)
            }
        })
    }
}

func TestDiff_Success(t *testing.T) {
    a := map[string]any{"title": "old", "notes": "same"}
    b := map[string]any{"title": "new", "notes": "same"}
    diff, err := history.Diff(a, b)
    require.NoError(t, err)
    require.NotNil(t, diff)
    ops, ok := diff["ops"].([]any)
    require.True(t, ok)
    assert.Len(t, ops, 1)
}

func TestDiff_Empty(t *testing.T) {
    a := map[string]any{"title": "same"}
    diff, err := history.Diff(a, a)
    require.NoError(t, err)
    assert.Nil(t, diff)
}
```

**Step 3: Run failing tests**

Run: `go test -p 1 -timeout 60s ./pkg/transactions/history/...`
Expected: FAIL — undefined `history.Snapshot`, `history.Diff`.

**Step 4: Implement** (`snapshot.go`)

```go
package history

import (
    "encoding/json"

    "github.com/cockroachdb/errors"
    "github.com/ft-t/go-money/pkg/database"
    "github.com/wI2L/jsondiff"
)

// Snapshot serialises a Transaction to a map, dropping HistoryExcludedFields.
// Uses GORM/JSON column names (snake_case) by going through json.Marshal on the
// struct first, then dropping excluded keys.
func Snapshot(tx *database.Transaction) (map[string]any, error) {
    raw, err := json.Marshal(toMarshallable(tx))
    if err != nil {
        return nil, errors.Wrap(err, "marshal tx")
    }
    var m map[string]any
    if err := json.Unmarshal(raw, &m); err != nil {
        return nil, errors.Wrap(err, "unmarshal tx")
    }
    for k := range HistoryExcludedFields {
        delete(m, k)
    }
    return m, nil
}

// toMarshallable mirrors database.Transaction with snake_case json tags.
// Centralising the wire shape here makes the snapshot stable even if the
// GORM struct gains tags or fields later.
type marshallableTx struct {
    ID                       int64             `json:"id"`
    SourceAmount             any               `json:"source_amount"`
    SourceCurrency           string            `json:"source_currency"`
    SourceAmountInBase       any               `json:"source_amount_in_base_currency"`
    FxSourceAmount           any               `json:"fx_source_amount"`
    FxSourceCurrency         string            `json:"fx_source_currency"`
    DestinationAmount        any               `json:"destination_amount"`
    DestinationCurrency      string            `json:"destination_currency"`
    DestinationAmountInBase  any               `json:"destination_amount_in_base_currency"`
    SourceAccountID          int32             `json:"source_account_id"`
    DestinationAccountID     int32             `json:"destination_account_id"`
    TagIDs                   []int32           `json:"tag_ids"`
    CreatedAt                any               `json:"created_at"`
    UpdatedAt                any               `json:"updated_at"`
    DeletedAt                any               `json:"deleted_at"`
    Notes                    string            `json:"notes"`
    Extra                    map[string]string `json:"extra"`
    TransactionDateTime      any               `json:"transaction_date_time"`
    TransactionDateOnly      any               `json:"transaction_date_only"`
    TransactionType          int32             `json:"transaction_type"`
    Flags                    int64             `json:"flags"`
    VoidedByTransactionID    *int64            `json:"voided_by_transaction_id"`
    Title                    string            `json:"title"`
    ReferenceNumber          *string           `json:"reference_number"`
    InternalReferenceNumbers []string          `json:"internal_reference_numbers"`
    CategoryID               *int32            `json:"category_id"`
}

func toMarshallable(tx *database.Transaction) marshallableTx {
    var srcAmt, srcAmtBase, fxAmt, dstAmt, dstAmtBase any
    if tx.SourceAmount.Valid {
        srcAmt = tx.SourceAmount.Decimal.String()
    }
    if tx.SourceAmountInBaseCurrency.Valid {
        srcAmtBase = tx.SourceAmountInBaseCurrency.Decimal.String()
    }
    if tx.FxSourceAmount.Valid {
        fxAmt = tx.FxSourceAmount.Decimal.String()
    }
    if tx.DestinationAmount.Valid {
        dstAmt = tx.DestinationAmount.Decimal.String()
    }
    if tx.DestinationAmountInBaseCurrency.Valid {
        dstAmtBase = tx.DestinationAmountInBaseCurrency.Decimal.String()
    }
    var deletedAt any
    if tx.DeletedAt.Valid {
        deletedAt = tx.DeletedAt.Time
    }
    return marshallableTx{
        ID:                       tx.ID,
        SourceAmount:             srcAmt,
        SourceCurrency:           tx.SourceCurrency,
        SourceAmountInBase:       srcAmtBase,
        FxSourceAmount:           fxAmt,
        FxSourceCurrency:         tx.FxSourceCurrency,
        DestinationAmount:        dstAmt,
        DestinationCurrency:      tx.DestinationCurrency,
        DestinationAmountInBase:  dstAmtBase,
        SourceAccountID:          tx.SourceAccountID,
        DestinationAccountID:     tx.DestinationAccountID,
        TagIDs:                   []int32(tx.TagIDs),
        CreatedAt:                tx.CreatedAt,
        UpdatedAt:                tx.UpdatedAt,
        DeletedAt:                deletedAt,
        Notes:                    tx.Notes,
        Extra:                    tx.Extra,
        TransactionDateTime:      tx.TransactionDateTime,
        TransactionDateOnly:      tx.TransactionDateOnly,
        TransactionType:          int32(tx.TransactionType),
        Flags:                    int64(tx.Flags),
        VoidedByTransactionID:    tx.VoidedByTransactionID,
        Title:                    tx.Title,
        ReferenceNumber:          tx.ReferenceNumber,
        InternalReferenceNumbers: []string(tx.InternalReferenceNumbers),
        CategoryID:               tx.CategoryID,
    }
}

// Diff returns RFC 6902 patch wrapped as {"ops": [...]} for jsonb storage,
// or nil if a == b.
func Diff(prev, curr map[string]any) (map[string]any, error) {
    patch, err := jsondiff.Compare(prev, curr)
    if err != nil {
        return nil, errors.Wrap(err, "compute json patch")
    }
    if len(patch) == 0 {
        return nil, nil
    }
    raw, err := json.Marshal(patch)
    if err != nil {
        return nil, errors.Wrap(err, "marshal patch")
    }
    var ops []any
    if err := json.Unmarshal(raw, &ops); err != nil {
        return nil, errors.Wrap(err, "unmarshal patch")
    }
    return map[string]any{"ops": ops}, nil
}
```

**Step 5: Run tests**

Run: `go test -p 1 -timeout 60s ./pkg/transactions/history/...`
Expected: PASS.

**Step 6: Commit**

```bash
git add pkg/transactions/history/snapshot.go pkg/transactions/history/snapshot_test.go go.mod go.sum
git commit -S -m "feat(history): snapshot + RFC 6902 diff helpers"
```

---

## Task 5 — History service (Recorder + Reader)

**Files:**
- Create: `pkg/transactions/history/service.go`
- Create: `pkg/transactions/history/service_test.go`

**Step 1: Write failing tests** (`service_test.go`)

```go
package history_test

import (
    "context"
    "testing"

    "github.com/ft-t/go-money/pkg/database"
    "github.com/ft-t/go-money/pkg/testingutils"
    "github.com/ft-t/go-money/pkg/transactions/history"
    "github.com/shopspring/decimal"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
    testingutils.RunTests(m)
}

func TestService_Record_Success(t *testing.T) {
    cases := []struct {
        name      string
        eventType database.TransactionHistoryEventType
        actor     history.Actor
        expectDiff bool
    }{
        {
            name:      "created has nil diff",
            eventType: database.TransactionHistoryEventTypeCreated,
            actor:     history.UserActor(7),
            expectDiff: false,
        },
        {
            name:      "updated has diff",
            eventType: database.TransactionHistoryEventTypeUpdated,
            actor:     history.UserActor(7),
            expectDiff: true,
        },
    }

    for _, tc := range cases {
        t.Run(tc.name, func(t *testing.T) {
            testingutils.FlushAllTables()
            db := database.GetDb(database.DbTypeMaster)
            svc := history.NewService()

            curr := &database.Transaction{ID: 1, Title: "new title", SourceCurrency: "USD"}
            var prev *database.Transaction
            if tc.eventType == database.TransactionHistoryEventTypeUpdated {
                prev = &database.Transaction{ID: 1, Title: "old title", SourceCurrency: "USD"}
            }

            err := svc.Record(context.Background(), db, history.RecordRequest{
                Tx:        curr,
                Previous:  prev,
                EventType: tc.eventType,
                Actor:     tc.actor,
            })
            require.NoError(t, err)

            var rows []database.TransactionHistory
            require.NoError(t, db.Where("transaction_id = ?", 1).Find(&rows).Error)
            require.Len(t, rows, 1)
            assert.Equal(t, tc.eventType, rows[0].EventType)
            assert.Equal(t, tc.actor.Type, rows[0].ActorType)
            assert.Equal(t, tc.expectDiff, rows[0].Diff != nil)
        })
    }
}

func TestService_List_Success(t *testing.T) {
    testingutils.FlushAllTables()
    db := database.GetDb(database.DbTypeMaster)
    svc := history.NewService()

    tx1 := &database.Transaction{ID: 9, Title: "v1", SourceCurrency: "USD"}
    tx2 := &database.Transaction{ID: 9, Title: "v2", SourceCurrency: "USD"}

    require.NoError(t, svc.Record(context.Background(), db, history.RecordRequest{
        Tx: tx1, EventType: database.TransactionHistoryEventTypeCreated, Actor: history.UserActor(1),
    }))
    require.NoError(t, svc.Record(context.Background(), db, history.RecordRequest{
        Tx: tx2, Previous: tx1, EventType: database.TransactionHistoryEventTypeUpdated, Actor: history.UserActor(1),
    }))

    rows, err := svc.List(context.Background(), 9)
    require.NoError(t, err)
    require.Len(t, rows, 2)
    assert.Equal(t, database.TransactionHistoryEventTypeCreated, rows[0].EventType)
    assert.Equal(t, database.TransactionHistoryEventTypeUpdated, rows[1].EventType)
}

// Bonus: zero-diff update is still recorded with nil diff (e.g. forced rule run that didn't change anything is filtered upstream, but defensive test ensures Service won't drop events).
func TestService_Record_ZeroDiff_Update(t *testing.T) {
    decimal.DivisionPrecision = 16 // touch decimal package to avoid unused import
    testingutils.FlushAllTables()
    db := database.GetDb(database.DbTypeMaster)
    svc := history.NewService()
    tx := &database.Transaction{ID: 5, Title: "same"}
    require.NoError(t, svc.Record(context.Background(), db, history.RecordRequest{
        Tx: tx, Previous: tx, EventType: database.TransactionHistoryEventTypeUpdated,
        Actor: history.UserActor(1),
    }))
    var rows []database.TransactionHistory
    require.NoError(t, db.Where("transaction_id = ?", 5).Find(&rows).Error)
    require.Len(t, rows, 1)
    assert.Nil(t, rows[0].Diff)
}
```

**Step 2: Run failing tests**

Run: `Db_Host=tools.lan ReadonlyDb_Host=tools.lan Redis_Host=tools.lan go test -p 1 -timeout 60s ./pkg/transactions/history/...`
Expected: FAIL — undefined `history.NewService`.

**Step 3: Implement** (`service.go`)

```go
package history

import (
    "context"
    "time"

    "github.com/cockroachdb/errors"
    "github.com/ft-t/go-money/pkg/database"
    "gorm.io/gorm"
)

type Service struct{}

func NewService() *Service { return &Service{} }

func (s *Service) Record(ctx context.Context, tx *gorm.DB, req RecordRequest) error {
    snap, err := Snapshot(req.Tx)
    if err != nil {
        return errors.Wrap(err, "snapshot")
    }

    var diff map[string]any
    if req.Previous != nil {
        prevSnap, err := Snapshot(req.Previous)
        if err != nil {
            return errors.Wrap(err, "snapshot prev")
        }
        diff, err = Diff(prevSnap, snap)
        if err != nil {
            return errors.Wrap(err, "diff")
        }
    }

    row := &database.TransactionHistory{
        TransactionID: req.Tx.ID,
        EventType:     req.EventType,
        ActorType:     req.Actor.Type,
        ActorUserID:   req.Actor.UserID,
        ActorRuleID:   req.Actor.RuleID,
        Snapshot:      snap,
        Diff:          diff,
        OccurredAt:    time.Now().UTC(),
    }
    if req.Actor.Extra != "" {
        e := req.Actor.Extra
        row.ActorExtra = &e
    }
    return errors.WithStack(tx.WithContext(ctx).Create(row).Error)
}

func (s *Service) List(ctx context.Context, transactionID int64) ([]*database.TransactionHistory, error) {
    var rows []*database.TransactionHistory
    if err := database.GetDbWithContext(ctx, database.DbTypeReadonly).
        Where("transaction_id = ?", transactionID).
        Order("occurred_at ASC, id ASC").
        Find(&rows).Error; err != nil {
        return nil, errors.WithStack(err)
    }
    return rows, nil
}
```

**Step 4: Run tests**

Run: `Db_Host=tools.lan ReadonlyDb_Host=tools.lan Redis_Host=tools.lan go test -p 1 -timeout 60s ./pkg/transactions/history/...`
Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/transactions/history/service.go pkg/transactions/history/service_test.go
git commit -S -m "feat(history): Service with Record + List"
```

---

## Task 6 — Hook history into transactions.Service (CRUD)

**Files:**
- Modify: `pkg/transactions/interfaces.go` (add `HistorySvc`)
- Modify: `pkg/transactions/service.go` — add `HistorySvc` to `ServiceConfig`, call from `CreateBulkInternal`, `Update`, `DeleteTransaction`.

**Step 1: Add interface in `pkg/transactions/interfaces.go`**

```go
type HistorySvc interface {
    Record(ctx context.Context, tx *gorm.DB, req history.RecordRequest) error
}
```

(Add `gorm` and `history` imports.)

**Step 2: Add field to `ServiceConfig`**

```go
type ServiceConfig struct {
    // ...existing fields...
    HistorySvc HistorySvc
}
```

**Step 3: Hook `CreateBulkInternal` (after `tx.CreateInBatches` and `tx.Updates`, in `pkg/transactions/service.go`)**

Use the actor from context; if absent, fall back to no actor (event still recorded with `ActorTypeUnspecified`-equivalent — set to `User` with nil UserID for now is wrong; better: skip recording when no actor → log warn). Decide: **require actor**. Rationale: handlers/middleware/scheduler/importer must set it. Missing actor = bug → log warn + skip record (don't fail the tx).

After `toCreate` succeeded:

```go
for _, t := range toCreate {
    s.recordHistory(ctx, tx, t, nil, database.TransactionHistoryEventTypeCreated)
}
```

After per-tx `tx.Updates(newTx)` succeeded, for entries in `toUpdate` paired with `originalTxs`:

```go
// In the toUpdate loop, find matching original by ID:
origByID := map[int64]*database.Transaction{}
for _, o := range originalTxs {
    origByID[o.ID] = o
}
for _, newTx := range toUpdate {
    if err := tx.Updates(newTx).Error; err != nil { ... }
    s.recordHistory(ctx, tx, newTx, origByID[newTx.ID], database.TransactionHistoryEventTypeUpdated)
}
```

Add helper at bottom of service.go:

```go
func (s *Service) recordHistory(
    ctx context.Context,
    tx *gorm.DB,
    curr *database.Transaction,
    prev *database.Transaction,
    eventType database.TransactionHistoryEventType,
) {
    actor, ok := history.ActorFromContext(ctx)
    if !ok {
        zerolog.Ctx(ctx).Warn().
            Int64("tx_id", curr.ID).
            Int16("event_type", int16(eventType)).
            Msg("history actor missing from context; skipping history record")
        return
    }
    if err := s.cfg.HistorySvc.Record(ctx, tx, history.RecordRequest{
        Tx: curr, Previous: prev, EventType: eventType, Actor: actor,
    }); err != nil {
        zerolog.Ctx(ctx).Error().Err(err).
            Int64("tx_id", curr.ID).
            Msg("failed to record transaction history")
    }
}
```

(Add `history` import.)

**Step 4: Hook `DeleteTransaction`**

Inside the for loop, after `tx.Delete(txToDelete)` succeeds:

```go
s.recordHistory(ctx, tx, txToDelete, txToDelete, database.TransactionHistoryEventTypeDeleted)
```

(`Previous == Current` is fine — diff will be nil, snapshot captures final state. Event type alone signals the delete.)

**Step 5: Add tests covering history hook in `pkg/transactions/service_history_test.go`**

(One success table per code path: create, update, delete. Use mocked `HistorySvc` via `gomock`.)

```go
package transactions_test

// minimal sketch — full file follows existing service_test.go patterns.
// Each test:
// 1. Build service with mock HistorySvc.
// 2. Set ctx via history.WithActor(ctx, history.UserActor(1)).
// 3. Expect Recorder.Record called with matching event_type + actor.
// 4. Expect Record NOT called when ctx has no actor.
```

**Step 6: Regenerate mocks**

Run: `make generate`
Expected: `pkg/transactions/interfaces_mocks_test.go` updated with `MockHistorySvc`.

**Step 7: Run tests**

Run: `Db_Host=tools.lan ReadonlyDb_Host=tools.lan Redis_Host=tools.lan go test -p 1 -timeout 60s ./pkg/transactions/...`
Expected: PASS.

**Step 8: Commit**

```bash
git add pkg/transactions/interfaces.go pkg/transactions/service.go pkg/transactions/service_history_test.go pkg/transactions/interfaces_mocks_test.go
git commit -S -m "feat(transactions): emit history on create/update/delete"
```

---

## Task 7 — Hook rules.Executor (per-rule events)

**Files:**
- Modify: `pkg/transactions/rules/interfaces.go` — add `HistorySvc`.
- Modify: `pkg/transactions/rules/executor.go` — accept HistorySvc; emit per-rule events when rule changed tx.

**Decision:** Executor records history *inline* (does not require ctx actor). Each rule that mutated tx → one event with `ActorType=RULE`, `RuleID=rule.ID`. Diff vs the tx state *before* this rule ran (not before the whole pipeline).

**Step 1: Update interfaces.go**

```go
type HistorySvc interface {
    Record(ctx context.Context, tx *gorm.DB, req history.RecordRequest) error
}
```

**Step 2: Modify executor.go** — `NewExecutor` takes optional history svc; `executeInternal` records on each rule that returned `result == true` AND produced a non-empty diff (jsondiff returns nil for no-op).

```go
type Executor struct {
    interpreter Interpreter
    history     HistorySvc // may be nil for callers that don't need it (e.g. dry-run)
    db          DBProvider // see below
}

type DBProvider interface {
    GetDB(ctx context.Context) *gorm.DB
}

func NewExecutor(interpreter Interpreter, hist HistorySvc, db DBProvider) *Executor {
    return &Executor{interpreter: interpreter, history: hist, db: db}
}
```

`db` is needed because rule execution doesn't yet have an open `*gorm.DB` — but `database.FromContext` already returns the active tx. Use that directly instead of a `DBProvider`:

```go
// inside executeInternal, after loop or per-rule:
if result {
    if err := s.recordRule(ctx, before, clonedTx, rule); err != nil {
        return nil, err
    }
    tx = clonedTx
}
```

```go
func (s *Executor) recordRule(ctx context.Context, before, after *database.Transaction, rule *database.Rule) error {
    if s.history == nil || after.ID == 0 {
        // tx not yet persisted (CREATE path) — defer rule events to caller.
        // See note below.
        return nil
    }
    db := database.FromContext(ctx, database.GetDbWithContext(ctx, database.DbTypeMaster))
    return s.history.Record(ctx, db, history.RecordRequest{
        Tx: after, Previous: before, EventType: database.TransactionHistoryEventTypeRuleApplied,
        Actor: history.RuleActor(rule.ID),
    })
}
```

**Important nuance:** rules run *before* the tx is persisted on CREATE (no `tx.ID`). For CREATE, rule events must be flushed *after* the parent CREATED event when the tx has an ID. Two options:

- **Option A (recommended):** Buffer rule events on the `*database.Transaction` (new transient field `RuleEvents []RuleEvent` in `database.Transaction`, `gorm:"-"`). After `transactions.Service.CreateBulkInternal` records the parent CREATED event, it iterates `tx.RuleEvents` and calls `historySvc.Record` for each.
- **Option B:** Persist tx, then run rules, then update — splits the existing single-pass flow.

Pick **A**. Implementation:

1. Add to `database.Transaction`:
   ```go
   RuleAppliedEvents []RuleAppliedEvent `gorm:"-"`
   ```
   ```go
   type RuleAppliedEvent struct {
       RuleID int32
       Before *Transaction
       After  *Transaction
   }
   ```
2. Executor appends to `tx.RuleAppliedEvents` instead of writing immediately.
3. `transactions.Service.recordHistory` — after writing CREATED/UPDATED, drains `tx.RuleAppliedEvents` and writes one RULE_APPLIED event per entry.

**Step 3: Tests** — `pkg/transactions/rules/executor_history_test.go`. Two tables: success (one rule changed, two rules changed, no rule changed → no events) and failure (record returns error → executor surfaces error).

**Step 4: Update wiring (deferred to Task 12 main.go).**

**Step 5: Build + test**

Run: `make generate && Db_Host=tools.lan ReadonlyDb_Host=tools.lan Redis_Host=tools.lan go test -p 1 -timeout 60s ./pkg/transactions/...`
Expected: PASS.

**Step 6: Commit**

```bash
git add pkg/transactions/rules/interfaces.go pkg/transactions/rules/executor.go pkg/transactions/rules/executor_history_test.go pkg/database/transaction.go pkg/transactions/service.go pkg/transactions/service_history_test.go
git commit -S -m "feat(rules): buffer rule-applied events for history sink"
```

---

## Task 8 — Hook BulkSetCategory / BulkSetTags

**Files:**
- Modify: `pkg/transactions/service.go` — `BulkSetCategory` and `BulkSetTags` now read previous tx state, write update, then record history with `BulkActor`.

**Step 1: Refactor BulkSetCategory**

```go
func (s *Service) BulkSetCategory(ctx context.Context, assignments []CategoryAssignment) error {
    if len(assignments) == 0 {
        return nil
    }
    tx := database.GetDbWithContext(ctx, database.DbTypeMaster).Begin()
    defer tx.Rollback()

    for _, a := range assignments {
        var prev database.Transaction
        if err := tx.Where("id = ? AND deleted_at IS NULL", a.TransactionID).First(&prev).Error; err != nil {
            return errors.Wrapf(err, "load tx %d", a.TransactionID)
        }
        next := prev
        next.CategoryID = &a.CategoryID
        if err := tx.Model(&database.Transaction{}).
            Where("id = ?", a.TransactionID).
            Update("category_id", a.CategoryID).Error; err != nil {
            return errors.Wrapf(err, "update tx %d", a.TransactionID)
        }
        s.recordHistoryBulk(ctx, tx, &next, &prev, "set_category")
    }
    return errors.WithStack(tx.Commit().Error)
}

func (s *Service) recordHistoryBulk(ctx context.Context, tx *gorm.DB, curr, prev *database.Transaction, op string) {
    // BULK requires user_id from ctx. Fall back to ActorFromContext else warn.
    actor, ok := history.ActorFromContext(ctx)
    if !ok || actor.UserID == nil {
        zerolog.Ctx(ctx).Warn().Msg("bulk op without user actor; skipping history")
        return
    }
    bulkActor := history.BulkActor(*actor.UserID, op)
    if err := s.cfg.HistorySvc.Record(ctx, tx, history.RecordRequest{
        Tx: curr, Previous: prev, EventType: database.TransactionHistoryEventTypeUpdated, Actor: bulkActor,
    }); err != nil {
        zerolog.Ctx(ctx).Error().Err(err).Int64("tx_id", curr.ID).Msg("failed to record bulk history")
    }
}
```

**Step 2: Mirror for BulkSetTags** (same shape, op = `"set_tags"`).

**Step 3: Tests** — `pkg/transactions/service_bulk_history_test.go`. Two tables: success per op, failure (load fails → no history).

**Step 4: Run**

Run: `Db_Host=tools.lan ReadonlyDb_Host=tools.lan Redis_Host=tools.lan go test -p 1 -timeout 60s ./pkg/transactions/...`
Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/transactions/service.go pkg/transactions/service_bulk_history_test.go
git commit -S -m "feat(transactions): record history on bulk category/tag ops"
```

---

## Task 9 — Hook importers (actor=IMPORTER)

**Files:**
- Modify: `pkg/importers/importer.go` (and per-importer entrypoints if they call `transactionSvc.Create*` directly).

**Step 1: Identify importer entrypoints**

Run: `Grep "TransactionSvc\." in pkg/importers/`. List the call sites (likely in `importer.go` shared bridge). For each, wrap ctx:

```go
ctx = history.WithActor(ctx, history.ImporterActor("firefly")) // or "mono", "privat24", etc.
```

**Step 2: Add helper to choose name**

If importer type is already an enum, map it. Otherwise, accept a string parameter on the importer constructor.

**Step 3: Tests** — extend each importer test to assert ctx propagation by spying on a wrapped TransactionSvc.

**Step 4: Run**

Run: `Db_Host=tools.lan ReadonlyDb_Host=tools.lan Redis_Host=tools.lan go test -p 1 -timeout 60s ./pkg/importers/...`
Expected: PASS.

**Step 5: Commit**

```bash
git add pkg/importers/
git commit -S -m "feat(importers): tag transactions with IMPORTER history actor"
```

---

## Task 10 — Hook scheduler (actor=SCHEDULER)

**Files:**
- Modify: `pkg/transactions/rules/scheduler.go` — `ExecuteTask` wraps ctx with `SchedulerActor(rule.ID)` before calling `TransactionSvc.CreateRawTransaction`.

**Step 1: Wrap ctx**

In `ExecuteTask`:

```go
ctx = history.WithActor(ctx, history.SchedulerActor(rule.ID))
```

**Step 2: Test** — `scheduler_test.go` add success case asserting actor in ctx propagates.

**Step 3: Run**

Run: `go test -p 1 -timeout 60s ./pkg/transactions/rules/...`
Expected: PASS.

**Step 4: Commit**

```bash
git add pkg/transactions/rules/scheduler.go pkg/transactions/rules/scheduler_test.go
git commit -S -m "feat(scheduler): tag scheduled transactions with SCHEDULER actor"
```

---

## Task 11 — gRPC handler

**Files:**
- Create: `cmd/server/internal/handlers/transaction_history.go`
- Create: `cmd/server/internal/handlers/transaction_history_test.go`
- Modify: `cmd/server/internal/handlers/interfaces.go` — add `TransactionHistorySvc`.

**Step 1: Add service interface** (`interfaces.go`)

```go
type TransactionHistorySvc interface {
    List(ctx context.Context, transactionID int64) ([]*database.TransactionHistory, error)
}
```

**Step 2: Implement handler**

```go
package handlers

import (
    "context"

    "buf.build/gen/go/xskydev/go-money-pb/connectrpc/go/gomoneypb/transactions/history/v1/historyv1connect"
    historyv1 "buf.build/gen/go/xskydev/go-money-pb/protocolbuffers/go/gomoneypb/transactions/history/v1"
    "connectrpc.com/connect"
    "github.com/ft-t/go-money/cmd/server/internal/middlewares"
    "github.com/ft-t/go-money/pkg/auth"
    "github.com/ft-t/go-money/pkg/boilerplate"
)

type TransactionHistoryApi struct {
    svc    TransactionHistorySvc
    mapper TransactionHistoryMapper
}

type TransactionHistoryMapper interface {
    MapEvent(row *database.TransactionHistory) *historyv1.TransactionHistoryEvent
}

func NewTransactionHistoryApi(grpc *boilerplate.GrpcServer, svc TransactionHistorySvc, mapper TransactionHistoryMapper) (*TransactionHistoryApi, error) {
    a := &TransactionHistoryApi{svc: svc, mapper: mapper}
    grpc.Register(historyv1connect.NewTransactionHistoryServiceHandler(a))
    return a, nil
}

func (a *TransactionHistoryApi) ListHistory(
    ctx context.Context,
    c *connect.Request[historyv1.ListTransactionHistoryRequest],
) (*connect.Response[historyv1.ListTransactionHistoryResponse], error) {
    jwt := middlewares.FromContext(ctx)
    if jwt.UserID == 0 {
        return nil, connect.NewError(connect.CodePermissionDenied, auth.ErrInvalidToken)
    }
    rows, err := a.svc.List(ctx, c.Msg.TransactionId)
    if err != nil {
        return nil, connect.NewError(connect.CodeInternal, err)
    }
    out := make([]*historyv1.TransactionHistoryEvent, 0, len(rows))
    for _, r := range rows {
        out = append(out, a.mapper.MapEvent(r))
    }
    return connect.NewResponse(&historyv1.ListTransactionHistoryResponse{Events: out}), nil
}
```

**Step 3: Mapper** — extend `pkg/mapper/mapper.go` (or wherever `MapTransaction` lives) with `MapTransactionHistoryEvent`. Convert `Snapshot`/`Diff` maps to `*structpb.Struct` via `structpb.NewStruct`.

**Step 4: Tests** — auth check (no jwt → CodePermissionDenied), success path returns mapped events.

**Step 5: Build**

Run: `make generate && go build ./...`
Expected: PASS.

**Step 6: Commit**

```bash
git add cmd/server/internal/handlers/transaction_history.go cmd/server/internal/handlers/transaction_history_test.go cmd/server/internal/handlers/interfaces.go pkg/mapper/
git commit -S -m "feat(api): TransactionHistoryService.ListHistory handler"
```

---

## Task 12 — Wire in main.go + middleware actor

**Files:**
- Modify: `cmd/server/main.go` — instantiate `history.NewService()`; pass to `transactions.ServiceConfig{HistorySvc: ...}`, `rules.NewExecutor(..., historySvc, ...)`; register `TransactionHistoryApi`.
- Modify: `cmd/server/internal/middlewares/middleware.go` — auth middleware sets `history.WithActor(ctx, history.UserActor(jwtData.UserID))` after JWT verification (before delegating to handler chain).

**Step 1: Update middleware**

In the auth middleware where `ctx = WithContext(ctx, claims)`:

```go
import historypkg "github.com/ft-t/go-money/pkg/transactions/history"

// after claims attached:
if claims.UserID != 0 {
    ctx = historypkg.WithActor(ctx, historypkg.UserActor(claims.UserID))
}
```

**Step 2: Update main.go wiring**

```go
historySvc := history.NewService()

ruleEngine := rules.NewExecutor(ruleInterpreter) // existing — needs new signature
// becomes:
ruleEngine := rules.NewExecutor(ruleInterpreter, historySvc)

transactionSvc := transactions.NewService(&transactions.ServiceConfig{
    // ...existing...
    HistorySvc: historySvc,
})

// register handler:
if _, err := handlers.NewTransactionHistoryApi(grpcServer, historySvc, mapper); err != nil {
    log.Logger.Fatal().Err(err).Msg("failed to create transaction history handler")
}
```

**Step 3: Build**

Run: `go build ./...`
Expected: PASS.

**Step 4: Boot end-to-end smoke test**

```bash
make build-docker
cd compose && docker compose -f docker-compose-backend.yaml down && docker compose -f docker-compose-backend.yaml up -d
docker logs --tail=50 compose-app-1
```
Expected: clean boot, migration applied, no panics.

**Step 5: Commit**

```bash
git add cmd/server/main.go cmd/server/internal/middlewares/middleware.go
git commit -S -m "feat(server): wire history svc + handler; middleware sets USER actor"
```

---

## Task 13 — Frontend service client

**Files:**
- Create: `frontend/src/app/services/transaction-history.service.ts`

**Step 1: Implement**

```typescript
import { Inject, Injectable } from '@angular/core';
import { createClient, Transport } from '@connectrpc/connect';
import { TRANSPORT_TOKEN } from '../consts/transport';
import {
  TransactionHistoryService,
  ListTransactionHistoryRequestSchema,
} from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/transactions/history/v1/history_pb';
import { create } from '@bufbuild/protobuf';

@Injectable({ providedIn: 'root' })
export class TransactionHistoryClient {
  private client;
  constructor(@Inject(TRANSPORT_TOKEN) transport: Transport) {
    this.client = createClient(TransactionHistoryService, transport);
  }

  listHistory(transactionId: bigint) {
    return this.client.listHistory(create(ListTransactionHistoryRequestSchema, { transactionId }));
  }
}
```

**Step 2: Smoke test in Karma** — minimal `expect(client).toBeTruthy()`.

**Step 3: Commit**

```bash
git add frontend/src/app/services/transaction-history.service.ts
git commit -S -m "feat(frontend): TransactionHistory client"
```

---

## Task 14 — Frontend timeline component

**Files:**
- Create: `frontend/src/app/pages/transactions/history/transaction-history.component.ts`
- Create: `frontend/src/app/pages/transactions/history/transaction-history.component.html`
- Modify: `frontend/src/app/pages/transactions/transactions-details.component.{ts,html}` — add `<p-tabs>` with "Details" + "History" tabs.

**Step 1: Add jsondiffpatch dep**

```bash
cd frontend && npm install jsondiffpatch
```

**Step 2: Component scaffold**

Use PrimeNG `Timeline` (`p-timeline`). Per event:
- Marker: icon by event type (plus/pencil/trash/cogs).
- Header: actor label (user name / rule title / importer name / "scheduler" / "bulk").
- Body: human field-level diff list (rendered from `event.diff.ops`) — each op = `path` → `op` → `from`/`value`. For arrays/objects use `jsondiffpatch.formatters.html`.

Component public API: `@Input() transactionId: bigint;` plus internal `events: TransactionHistoryEvent[]`. Fetch via `TransactionHistoryClient.listHistory(transactionId)`.

**Step 3: Template** uses PrimeNG `Tabs` in details page. The History tab lazy-loads `<app-transaction-history [transactionId]="transaction.id">`.

**Step 4: Diff renderer** — small util `formatPatchOp(op)` returning `{label, severity, before, after}` where severity tags for color: green (add), amber (replace), red (remove).

For complex nested values (e.g. `tag_ids` array changes), fall back to `jsondiffpatch.formatters.html.format(jsondiffpatch.diff(before, after))` — embed as `[innerHTML]` with `DomSanitizer`.

**Step 5: Use PrimeNG MCP for component validation** — before writing template, call `mcp__primeng__get_component({ name: "Timeline" })` and `mcp__primeng__get_component({ name: "Tabs" })` to confirm v21 API.

**Step 6: Run frontend**

```bash
cd frontend && npm test -- --watch=false
```
Expected: PASS.

**Step 7: Manual smoke** — follow `CLAUDE.md` "Local browser testing" steps. Open `/transactions/<id>`, switch to History tab, verify timeline renders create + any rule events.

**Step 8: Commit**

```bash
git add frontend/src/app/pages/transactions/ frontend/package.json frontend/package-lock.json
git commit -S -m "feat(frontend): transaction history timeline tab"
```

---

## Task 15 — End-to-end verification

**Step 1: Full lint**

Run: `make lint`
Expected: PASS.

**Step 2: Full test (modified packages)**

```bash
Db_Host=tools.lan ReadonlyDb_Host=tools.lan Redis_Host=tools.lan \
  go test -p 1 -timeout 60s \
  ./pkg/transactions/... ./pkg/importers/... ./cmd/server/internal/handlers/...
```
Expected: PASS.

**Step 3: Full build**

Run: `go build ./...`
Expected: PASS.

**Step 4: Frontend build**

```bash
cd frontend && npm run build
```
Expected: PASS.

**Step 5: Manual matrix**

In the running stack, verify each event source produces a row:

| Action | Expected event |
|--------|---------------|
| Create tx via UI | 1 CREATED + N RULE_APPLIED |
| Edit tx via UI | 1 UPDATED + N RULE_APPLIED (if rules trigger on update) |
| Delete tx via UI | 1 DELETED |
| Bulk set category from list | 1 UPDATED (actor=BULK, extra=set_category) |
| Bulk set tags from list | 1 UPDATED (actor=BULK, extra=set_tags) |
| Import CSV | N CREATED (actor=IMPORTER, extra=<importer name>) + RULE_APPLIED |
| Scheduled rule fires | 1 CREATED (actor=SCHEDULER, rule_id set) + any further RULE_APPLIED |

Inspect via SQL:
```sql
SELECT id, transaction_id, event_type, actor_type, actor_user_id, actor_rule_id, actor_extra, occurred_at
FROM transaction_history ORDER BY occurred_at DESC LIMIT 50;
```

**Step 6: Final commit**

If any drift fixes were needed, commit them.

**Step 7: Check pre-completion gate**

- [ ] No assets removed.
- [ ] Tests have no branching, separate success/failure tables.
- [ ] Mocks regenerated (`make generate`).
- [ ] All hooks call `recordHistory` defensively (warn on missing actor, never fail tx).
- [ ] `make lint` PASS.
- [ ] `go test -p 1 ./modified/...` PASS.
- [ ] `go build ./...` PASS.
- [ ] Manual matrix from Step 5 passes.

---

## Open notes / non-goals

- **Rule events on CREATE:** rule executor cannot persist events before tx has an ID; uses transient `RuleAppliedEvents` field (Task 7). On UPDATE the executor *could* write inline since tx.ID exists, but for symmetry it always buffers.
- **No UI for "revert to event N"** — out of scope. History is read-only display.
- **No retention policy** — append-only forever. Add a separate cleanup job if/when volume becomes a concern.
- **No bulk delete history yet** — bulk deletes are out-of-band; if added later, mirror Task 8 pattern.
- **FX recompute (`BaseAmountService`) deliberately not logged** — see Q4 in design discussion.
