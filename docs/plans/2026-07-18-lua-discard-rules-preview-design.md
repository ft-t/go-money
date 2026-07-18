# Lua `tx.discard()` + rules-in-import-preview — Design

Date: 2026-07-18
Status: approved (brainstormed in-session)
Proto: `go-money-pb` commits `de7c43c` (fields) — BSR build pinned via `make update-pb`.

## Problem

1. No way for a Lua rule to drop an unwanted transaction (e.g. `if tx.title() == "bcd"`) — every tx entering the rule engine is persisted.
2. Import preview (`ParseTransactions`) shows **pre-rule** state while the real import (`ImportTransactions` → `CreateBulkInternal` → `RuleSvc.ProcessTransactions`) persists **post-rule** state. Preview lies today; a discard verb would make the lie worse.

## Decisions (user-approved)

- **Discard = drop only** (option A): no row created, counted in response. Stateless — re-import runs rules again. No `import_ignored_transactions` write, no soft-flag rows.
- **Discard honored everywhere** rules run (option A): import, manual create, API, scheduled. Response carries `discarded=true`, no transaction payload.
- **Wire shape** (option A): new lightweight `AppliedRule { rule_id, diff_json }` on `ParsedTransaction`; header summary computed client-side.

## Design

### 1. Lua verb

- `LuaTransactionWrapper.Discard` registered as `tx:discard()`; sets `modified` + `tx.Discarded` directly on the wrapped transaction.
- `Interpreter.Run` signature unchanged (`(bool, error)`) — the discard verdict travels on the transient field, no interface/mock churn. (Simplification vs. the originally drafted `RunResult`.)
- `database.Transaction` gains transient `Discarded bool` (`gorm:"-"`), same pattern as `RuleAppliedEvents`.

### 2. Executor semantics

- On discard: mark `tx.Discarded`, **halt all remaining rules and groups** for that tx (discard is final; `is_final_rule` irrelevant after it).
- `RuleAppliedEvents` accumulated before discard are kept (preview shows what ran).
- Discarded tx **stays in `ProcessTransactions` output** — callers need it for counts/responses.

### 3. Persist path — `CreateBulkInternal` (`pkg/transactions/service.go`)

- Partition `Discarded` txs out of `toCreate`/`toUpdate`; never persisted; update-discard leaves existing row untouched.
- Response for a discarded req: `CreateTransactionResponse{discarded: true}` (no transaction).
- `Import` (`pkg/importers/importer.go`): `ImportTransactionsResponse.discarded_count` from responses; not conflated with `duplicate_count`/`skipped_count`.

### 4. Parse preview — rules dry-run

- `ParseTransactionsRequest.skip_rules` (field 3; proto3 default false ⇒ rules run by default, mirrors `ImportTransactionsRequest.skip_rules`).
- `Importer` gains `RuleSvc` dep (`ProcessTransactions`), wired in `cmd/server/main.go`.
- In `ConvertRequestsToTransactions`: for rows that are not duplicate/ignored, run the engine on the converted txs **in-memory only** — nothing persisted, no history rows.
- Per row: `discarded bool` (rule verdict; distinct from dedup `ignored`), `repeated AppliedRule applied_rules` — `diff_json` from existing `history.Snapshot`/`history.Diff` over each `RuleAppliedEvent.Before/After`.

### 5. Proto (published)

- `AppliedRule { int32 rule_id = 1; string diff_json = 2; }` (import/v1)
- `ParsedTransaction`: `discarded = 4`, `applied_rules = 5`
- `ImportTransactionsResponse.discarded_count = 4`
- `CreateTransactionResponse.discarded = 2`

### 6. Frontend (import review screen)

- DISCARDED badge (distinct from orange DUPLICATE); discarded rows auto-deselected and excluded from import gates (server drops them anyway; UI reflects).
- Per-row applied-rules popover/expander: rule title via `listRules`, diff rendered by diff-render logic **extracted from `transaction-history-timeline`** into a shared component (timeline keeps using it).
- Header summary: "N rules applied · M rows changed · K discarded" — computed from rows client-side.

### 7. Testing

- Executor: discard halts remaining rules/groups; events kept; output retains discarded tx.
- Service: partition (create/update/discard), counts, `discarded` response flag; manual create discard.
- Importers: preview populates `applied_rules`/`discarded`; `skip_rules=true` skips engine; real import excludes discarded and counts them.
- Handlers: passthrough.
- Frontend: badge/gating/summary specs.
- All: zero-branching tests, separate success/failure tables, real PG (`-p 1`).

## Out of scope

- No `tx.isDiscarded()` getter (YAGNI).
- No sticky ignore from rules (dedup table untouched by discard).
- No un-discard UI — deleting/fixing the rule is the escape hatch.
