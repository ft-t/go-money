# MCP Tool Reference

Detailed documentation of the Go Money MCP server tool.

## query Tool

Execute read-only SQL queries against the Go Money PostgreSQL database.

### Specification

```json
{
  "name": "query",
  "description": "Run a read-only SQL query",
  "parameters": {
    "type": "object",
    "properties": {
      "sql": {
        "type": "string",
        "description": "The SQL query to execute"
      }
    },
    "required": ["sql"]
  }
}
```

### Input Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| sql | string | Yes | Valid PostgreSQL SELECT statement |

### Output Format

Returns JSON array of result rows:

```json
[
  {
    "column1": "value1",
    "column2": 123,
    "column3": true
  },
  ...
]
```

### Data Types in Results

| PostgreSQL Type | JSON Type | Example |
|-----------------|-----------|---------|
| text, varchar | string | `"Hello"` |
| integer, bigint | number | `123` |
| numeric, decimal | string | `"123.45"` |
| boolean | boolean | `true` |
| timestamp | string (ISO 8601) | `"2024-01-15T10:30:00Z"` |
| date | string | `"2024-01-15"` |
| array | array | `[1, 2, 3]` |
| jsonb | object | `{"key": "value"}` |
| NULL | null | `null` |

### Usage Examples

#### Simple Query

```json
{
  "sql": "SELECT id, name FROM accounts WHERE deleted_at IS NULL LIMIT 5"
}
```

Response:
```json
[
  {"id": 1, "name": "Checking Account"},
  {"id": 2, "name": "Savings"},
  {"id": 3, "name": "Credit Card"},
  {"id": 4, "name": "Groceries"},
  {"id": 5, "name": "Salary"}
]
```

#### Aggregate Query

```json
{
  "sql": "SELECT SUM(current_balance) as total FROM accounts WHERE type = 1 AND deleted_at IS NULL"
}
```

Response:
```json
[
  {"total": "15234.56"}
]
```

#### Query with JOINs

```json
{
  "sql": "SELECT t.title, c.name as category FROM transactions t LEFT JOIN categories c ON c.id = t.category_id WHERE t.deleted_at IS NULL LIMIT 10"
}
```

### Error Responses

#### Syntax Error

```json
{
  "error": "ERROR: syntax error at or near \"SELEC\" (SQLSTATE 42601)"
}
```

#### Invalid Table

```json
{
  "error": "ERROR: relation \"invalid_table\" does not exist (SQLSTATE 42P01)"
}
```

#### Non-SELECT Query (Blocked)

```json
{
  "error": "Only SELECT queries are allowed"
}
```

#### Timeout

```json
{
  "error": "Query timeout exceeded"
}
```

### Limitations

| Limit | Value | Description |
|-------|-------|-------------|
| Query Type | SELECT only | Modifications blocked |
| Timeout | 30 seconds | Long queries terminated |
| Result Size | 10,000 rows | Larger results truncated |
| Column Count | No limit | All columns returned |

## Best Practices

### 1. Always Include Soft Delete Filter

```sql
-- Correct
SELECT * FROM transactions WHERE deleted_at IS NULL

-- Incorrect (includes deleted records)
SELECT * FROM transactions
```

### 2. Limit Results

```sql
-- Good: Limits results
SELECT * FROM transactions WHERE deleted_at IS NULL LIMIT 100

-- Bad: May return millions of rows
SELECT * FROM transactions WHERE deleted_at IS NULL
```

### 3. Use Indexed Columns in WHERE

```sql
-- Fast: Uses index on transaction_date_only
WHERE transaction_date_only >= '2024-01-01'

-- Slow: Function prevents index use
WHERE EXTRACT(YEAR FROM transaction_date_only) = 2024
```

### 4. Select Only Needed Columns

```sql
-- Good: Minimal data transfer
SELECT id, title, destination_amount FROM transactions

-- Avoid: Unnecessary data
SELECT * FROM transactions
```

### 5. Use Pre-computed Tables

```sql
-- Fast: Pre-computed balance
SELECT amount FROM daily_stat WHERE account_id = 1 ORDER BY date DESC LIMIT 1

-- Slow: Calculates from transactions
SELECT SUM(destination_amount) FROM transactions WHERE destination_account_id = 1
```

## Available Tables

| Table | Description | Key Use Cases |
|-------|-------------|---------------|
| accounts | All accounts | Balance queries, account lists |
| transactions | All transactions | Spending analysis, search |
| categories | Expense categories | Category breakdown |
| tags | Transaction tags | Tag analysis |
| currencies | Exchange rates | Multi-currency queries |
| daily_stat | Daily balance snapshots | Balance history, trends |
| double_entries | Debit/credit ledger | Formal accounting |
| rules | Automation rules | Rule listing |
| users | User accounts | User info |

## Quick Reference

### Transaction Types

```sql
WHERE transaction_type = 1  -- Transfer
WHERE transaction_type = 2  -- Income
WHERE transaction_type = 3  -- Expense
WHERE transaction_type = 5  -- Adjustment
```

### Account Types

```sql
WHERE type = 1  -- Asset
WHERE type = 4  -- Liability
WHERE type = 5  -- Expense
WHERE type = 6  -- Income
WHERE type = 7  -- Adjustment
```

### Amount Fields

```sql
-- Original currency amounts
source_amount, source_currency
destination_amount, destination_currency

-- Converted to base currency (for comparison/aggregation)
source_amount_in_base_currency
destination_amount_in_base_currency
```

### Date Fields

```sql
-- Date + time (for ordering)
transaction_date_time

-- Date only (for grouping/filtering)
transaction_date_only
```

## Transaction Creation

MCP exposes four create tools and four update tools, one per transaction type.
All tools operate via the transaction service, which runs the full pipeline
(rules engine, validation, stats, double-entry accounting).

### Required Fields by Transaction Type

| Field | Expense | Income | Transfer | Adjustment |
|-------|:-------:|:------:|:--------:|:----------:|
| transaction_date (RFC3339) | yes | yes | yes | yes |
| title | yes | yes | yes | yes |
| source_account_id | yes | yes | yes | auto |
| source_amount | yes (negative) | yes | yes (negative) | auto |
| source_currency | yes | yes | yes | auto |
| destination_account_id | yes | yes | yes | yes |
| destination_amount | yes (positive) | yes | yes (positive) | yes |
| destination_currency | yes | yes | yes | yes |

Legend: `yes` = required · `auto` = service resolves it (adjustments use the default adjustment account and derive the source amount via the currency converter) · sign hints in parentheses (negative/positive) are enforced by the transaction service. Income has no sign constraint.

Sign enforcement reference (from `pkg/transactions/service.go`):

- **Expense**: `destination_amount` must be positive. If `fx_source_amount` is supplied it must be negative and `fx_source_currency` becomes mandatory.
- **Transfer**: `source_amount` must be negative and `destination_amount` must be positive.
- **Income / Adjustment**: no sign constraint enforced at service layer.

All create tools also accept these common optional fields:
`notes`, `extra` (object, string-to-string map), `tag_ids` (number array),
`reference_number`, `internal_reference_numbers` (string array), `group_key`,
`skip_rules` (bool), `category_id`.

All `update_*` variants additionally require `id` (number, the transaction id
to replace) and otherwise take the same fields as their `create_*` counterpart.
All amounts are decimal strings (e.g. `"123.45"`), using `shopspring/decimal`
parse semantics.

### create_expense

Create an expense transaction. Required: transaction_date (RFC3339), title, source_account_id, source_amount (negative decimal string), source_currency, destination_account_id, destination_amount (positive decimal string), destination_currency. Optional: fx_source_amount (negative), fx_source_currency, notes, extra (map), tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id.

Example request:

```json
{
  "transaction_date": "2026-04-17T12:00:00Z",
  "title": "Groceries",
  "source_account_id": 1,
  "source_amount": "-42.50",
  "source_currency": "USD",
  "destination_account_id": 5,
  "destination_amount": "42.50",
  "destination_currency": "USD"
}
```

Example response:

```
Transaction created with id 1234
```

Error responses:

- `"title is required"` / `"transaction_date is required (RFC3339)"` / `"invalid transaction_date: ..."`
- `"source_account_id is required"` / `"source_amount is required"` / `"source_currency is required"`
- `"destination_account_id is required"` / `"destination_amount is required"` / `"destination_currency is required"`
- `"invalid source_amount: ..."` / `"invalid destination_amount: ..."` / `"invalid fx_source_amount: ..."`
- `"failed to create expense: <service error>"` — wraps validation/accounting failures (for example `"destination amount must be positive"`, `"foreign amount must be negative"`).

### create_income

Create an income transaction. Required: transaction_date (RFC3339), title, source_account_id, source_amount (decimal string), source_currency, destination_account_id, destination_amount (decimal string), destination_currency. Optional: notes, extra, tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id.

Example request:

```json
{
  "transaction_date": "2026-04-17T09:00:00Z",
  "title": "Salary",
  "source_account_id": 7,
  "source_amount": "3000.00",
  "source_currency": "USD",
  "destination_account_id": 1,
  "destination_amount": "3000.00",
  "destination_currency": "USD"
}
```

Example response:

```
Transaction created with id 1235
```

Error responses:

- Same field-required / parse errors as `create_expense` (except no fx_*).
- `"failed to create income: <service error>"`.

### create_transfer

Create a transfer between accounts. Required: transaction_date (RFC3339), title, source_account_id, source_amount (negative decimal string), source_currency, destination_account_id, destination_amount (positive decimal string), destination_currency. Optional: notes, extra, tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id.

Example request:

```json
{
  "transaction_date": "2026-04-17T18:00:00Z",
  "title": "Move to savings",
  "source_account_id": 1,
  "source_amount": "-500.00",
  "source_currency": "USD",
  "destination_account_id": 2,
  "destination_amount": "500.00",
  "destination_currency": "USD"
}
```

Example response:

```
Transaction created with id 1236
```

Error responses:

- Same field-required / parse errors as `create_expense` (no fx_*).
- `"failed to create transfer: <service error>"` — wraps `"source amount must be negative"` or `"destination amount must be positive"`.

### create_adjustment

Create a balance adjustment. The source account is resolved automatically (default adjustment account) and the source amount is derived via the currency converter. Required: transaction_date (RFC3339), title, destination_account_id, destination_amount (decimal string), destination_currency. Optional: notes, extra, tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id.

Example request:

```json
{
  "transaction_date": "2026-04-17T23:59:00Z",
  "title": "Reconcile cash wallet",
  "destination_account_id": 3,
  "destination_amount": "12.34",
  "destination_currency": "USD"
}
```

Example response:

```
Transaction created with id 1237
```

Error responses:

- `"destination_account_id is required"` / `"destination_amount is required"` / `"destination_currency is required"`
- `"invalid destination_amount: ..."`
- `"failed to create adjustment: <service error>"` — wraps `"failed to get default adjustment account"` or `"failed to convert destination amount to adjustment account currency"`.

### update_expense

Update an existing expense transaction by replacing all fields. Required: transaction_date (RFC3339), title, source_account_id, source_amount (negative decimal string), source_currency, destination_account_id, destination_amount (positive decimal string), destination_currency. Optional: fx_source_amount (negative), fx_source_currency, notes, extra (map), tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id. Also requires `id` (int, the transaction id to replace).

Example request:

```json
{
  "id": 1234,
  "transaction_date": "2026-04-17T12:00:00Z",
  "title": "Groceries (corrected)",
  "source_account_id": 1,
  "source_amount": "-43.00",
  "source_currency": "USD",
  "destination_account_id": 5,
  "destination_amount": "43.00",
  "destination_currency": "USD"
}
```

Example response:

```
Transaction 1234 updated
```

Error responses:

- `"id is required"` when `id` is missing or not a number.
- Same field-required / parse errors as `create_expense`.
- `"failed to update expense: <service error>"`.

### update_income

Update an existing income transaction by replacing all fields. Required: transaction_date (RFC3339), title, source_account_id, source_amount (decimal string), source_currency, destination_account_id, destination_amount (decimal string), destination_currency. Optional: notes, extra, tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id. Also requires `id` (int, the transaction id to replace).

Example request:

```json
{
  "id": 1235,
  "transaction_date": "2026-04-17T09:00:00Z",
  "title": "Salary (April)",
  "source_account_id": 7,
  "source_amount": "3100.00",
  "source_currency": "USD",
  "destination_account_id": 1,
  "destination_amount": "3100.00",
  "destination_currency": "USD"
}
```

Example response:

```
Transaction 1235 updated
```

Error responses:

- `"id is required"`.
- Same field-required / parse errors as `create_income`.
- `"failed to update income: <service error>"`.

### update_transfer

Update an existing transfer between accounts by replacing all fields. Required: transaction_date (RFC3339), title, source_account_id, source_amount (negative decimal string), source_currency, destination_account_id, destination_amount (positive decimal string), destination_currency. Optional: notes, extra, tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id. Also requires `id` (int, the transaction id to replace).

Example request:

```json
{
  "id": 1236,
  "transaction_date": "2026-04-17T18:00:00Z",
  "title": "Move to savings (fixed)",
  "source_account_id": 1,
  "source_amount": "-600.00",
  "source_currency": "USD",
  "destination_account_id": 2,
  "destination_amount": "600.00",
  "destination_currency": "USD"
}
```

Example response:

```
Transaction 1236 updated
```

Error responses:

- `"id is required"`.
- Same field-required / parse errors as `create_transfer`.
- `"failed to update transfer: <service error>"`.

### update_adjustment

Update an existing balance adjustment by replacing all fields. The source account is resolved automatically (default adjustment account) and the source amount is derived via the currency converter. Required: transaction_date (RFC3339), title, destination_account_id, destination_amount (decimal string), destination_currency. Optional: notes, extra, tag_ids, reference_number, internal_reference_numbers, group_key, skip_rules, category_id. Also requires `id` (int, the transaction id to replace).

Example request:

```json
{
  "id": 1237,
  "transaction_date": "2026-04-17T23:59:00Z",
  "title": "Reconcile cash wallet (corrected)",
  "destination_account_id": 3,
  "destination_amount": "15.00",
  "destination_currency": "USD"
}
```

Example response:

```
Transaction 1237 updated
```

Error responses:

- `"id is required"`.
- Same field-required / parse errors as `create_adjustment`.
- `"failed to update adjustment: <service error>"`.

## Currency Conversion

The server keeps exchange rates in the `currencies` table. Each row stores
`rate = units_per_base` relative to a configured base currency. Rates are
cached in-process behind an LRU (capacity 100) with
`configuration.DefaultCacheTTL`; the converter refreshes missing entries from
the readonly database on demand.

Conversion formula:

```
converted = amount / from_rate * to_rate
```

Same-currency calls short-circuit: `from_rate = to_rate = 1` and
`converted = amount`, and no rate lookup is performed.

### convert_currency

Convert an amount between two currencies using stored exchange rates. Rates are denominated vs base currency: amount / from_rate → base → × to_rate. Same-currency calls pass through (both rates = 1). Returns converted amount plus from_rate, to_rate, and base_currency so the caller can verify the math.

Input parameters:

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| from | string | Yes | ISO-4217 source currency code |
| to | string | Yes | ISO-4217 target currency code |
| amount | string | Yes | Decimal amount as string |

Response: JSON object with keys `from`, `to`, `amount`, `converted`,
`from_rate`, `to_rate`, `base_currency`. All values are strings; decimals use
the canonical `shopspring/decimal` string representation.

Example request:

```json
{
  "from": "EUR",
  "to": "USD",
  "amount": "100"
}
```

Example response:

```json
{
  "from": "EUR",
  "to": "USD",
  "amount": "100",
  "converted": "108.5",
  "from_rate": "0.92",
  "to_rate": "1",
  "base_currency": "USD"
}
```

Error responses:

- `"from, to, and amount are required"` when any argument is missing or empty.
- `"invalid amount: <wrap>"` when the `amount` string fails decimal parsing.
- `"failed to convert: <wrap>"` when a rate lookup fails (for example `"rate for XYZ not found"` when the currency is not present in the `currencies` table).

