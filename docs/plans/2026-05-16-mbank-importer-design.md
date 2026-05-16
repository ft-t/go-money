# mBank CSV importer — design

Date: 2026-05-16

## Goal

Add an mBank (Poland) CSV statement importer, parity with the existing
`privat24` / `revolut` importers. New import source `IMPORT_SOURCE_MBANK`.

## Source format

mBank "Lista operacji" CSV export.

- Encoding: **UTF-8 with BOM**.
- Line terminator: **CRLF**.
- Delimiter: `;`.
- ~25 metadata header lines, then the transaction table.
- Account number (26-digit Polish NRB) appears in the metadata line
  `mKonto Intensive - <NRB>;`.
- Account currency in the `#Waluta;#Wpływy;#Wydatki;` block, next row
  `PLN;<inflow>;<outflow>;`.
- Table anchor row: `#Data operacji;#Opis operacji;#Rachunek;#Kategoria;#Kwota;`.
- Data row columns:
  - `0` operation date — `2006-01-02` (date only, no time).
  - `1` description — quoted, space-padded, multiple sub-fields run together.
  - `2` masked own account label (e.g. `mKonto Intensive 7311 ... 9687`) — not
    used for matching (masked).
  - `3` mBank category (Polish, e.g. `Zdrowie i uroda`, `Bez kategorii`).
  - `4` amount — `-44,91 PLN`: optional sign, space/NBSP thousands grouping,
    `,` decimal separator, trailing currency code.
  - `5` trailing empty column.
- Data ends at the first blank/short row after the table.

## Approach

Mirror `privat24`: whole file parsed as one `Record` (no per-row CSV split —
the account number lives in the header and must survive).

`pkg/importers/mbank.go`:

- `type Mbank struct { *BaseParser }`, `NewMbank(base *BaseParser) *Mbank`.
- `Type()` → `importv1.ImportSource_IMPORT_SOURCE_MBANK`.
- `Parse` — identical shape to `Privat24.Parse` (decode files → one `Record`
  per file → `ParseMessages` → `GetAccountMapByNumbers` → `ToCreateRequests`).
- `ParseMessages`:
  - strip UTF-8 BOM,
  - `csv.Reader` with `Comma = ';'`, `FieldsPerRecord = -1`,
  - extract account NRB from the `mKonto ... - <NRB>` metadata line,
  - extract account currency from the `#Waluta` block,
  - locate anchor row (`Data operacji`),
  - iterate data rows until a blank/short row.

Per row → `*Transaction`:

- date: `time.Parse("2006-01-02", col0)`.
- description: collapse whitespace runs (`strings.Fields` + join), trim.
- category → `OriginalTxType` (raw string only; **no** `CategoryId` mapping —
  parity with privat24/revolut).
- amount: take the trailing currency token as currency; strip spaces, NBSP
  (` `), thin space (` `); replace `,` with `.`;
  `decimal.NewFromString`.
- single-currency PLN account (source currency == destination currency ==
  account currency):
  - negative → `TransactionTypeExpense`, `SourceAccount = NRB`.
  - positive → `TransactionTypeIncome`, `DestinationAccount = NRB`.
- `DeduplicationKeys`: `join(RFC3339 date, NRB, rawAmount, currency,
  description, category, "$$")` — one key, mirrors privat24 join style.

Parse failures set `tx.ParsingError` (row still emitted, like privat24) so the
user can review in the import UI.

Account match: header NRB → user's `Account.AccountNumber` via the existing
`BaseParser.GetAccountMapByNumbers`. No synthetic account name.

## Wiring

- `cmd/server/main.go`: add `importers.NewMbank(baseParser)` to the
  `importers.NewImporter(...)` implementations list.
- `importer.go` `importerSourceName`: `IMPORT_SOURCE_MBANK` → `"mbank"`
  (persisted to `transaction_history.actor_extra` — stable string).
- frontend `src/app/services/enum.service.ts`: add
  `{ name: 'mBank', value: ImportSource.MBANK, icon: '' }`.

## Protobuf

`go-money-pb` `proto/gomoneypb/import/v1/import.proto`:

```
IMPORT_SOURCE_MBANK = 6;
```

The proto edit is done in the `go-money-pb` repo. Publishing to BSR, bumping
the Go `buf.build/gen/go` module, and updating the frontend `@buf` package are
handled by the maintainer (out of band).

## Testing

`pkg/importers/mbank_test.go` — table-driven, separate success/failure
functions, zero branching. Fixtures under `pkg/importers/testdata/mbank/`
with masked data (no real name / NRB / IBAN):

- `expense.csv` — single negative PLN row.
- `income.csv` — single positive PLN row.
- `mixed.csv` — multiple rows, expense + income.
- `bad_date.csv` — unparseable date → `ParsingError`.
- `bad_amount.csv` — unparseable amount → `ParsingError`.

`TestMbank_Type` asserts the enum (gated on pb publish).
`ParseMessages` tests run independently of the enum.

## Verification gate (honest)

- `pkg/importers/mbank.go` references `IMPORT_SOURCE_MBANK`; frontend
  references `ImportSource.MBANK`. Neither compiles / type-checks until the pb
  is published, `make update-pb` is run, and the frontend `@buf` pb is
  reinstalled.
- Parsing-logic tests (`ParseMessages`, no enum dependency) are runnable now.
- `Type()`, `Parse`, full `go build ./...`, frontend build: gated on pb.

## Out of scope

- mBank category → go-money category auto-mapping.
- Multi-currency / FX rows (this export is single-currency PLN).
- Intraday time (export carries date only).
