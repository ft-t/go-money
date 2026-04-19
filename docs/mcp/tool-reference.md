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

## Transaction Rules

Rules are Lua scripts executed against every transaction (on create and on
edit, unless `skip_rules=true`). Each script receives a cloned transaction as
the global `tx` and helpers as `helpers`. A rule is considered "applied" when
the script mutates at least one field; applied rules are persisted into the
outgoing transaction.

Rules are grouped by `group_name` and ordered by `sort_order` (ascending)
within each group. If a rule has `is_final_rule=true` and it mutates the
transaction, execution stops for that group.

Runtime: [gopher-lua](https://github.com/yuin/gopher-lua) with
[gopher-lua-libs](https://github.com/vadv/gopher-lua-libs) preloaded
(`string`, `table`, `math`, etc.). Scripts must not raise errors — use early
`return` to exit.

### Lua API Reference

Globals: `tx` (transaction), `helpers` (utilities).

#### Transaction — `tx:field()` get, `tx:field(value)` set

| Method | Type | Notes |
|---|---|---|
| `tx:title()` / `tx:title(s)` | string | may be nil on sparse imports |
| `tx:notes()` / `tx:notes(s)` | string | may be nil |
| `tx:sourceCurrency()` / `tx:sourceCurrency(s)` | string (ISO-4217) | |
| `tx:destinationCurrency()` / `tx:destinationCurrency(s)` | string (ISO-4217) | |
| `tx:referenceNumber()` / `tx:referenceNumber(s)` | string | |
| `tx:sourceAccountID()` / `tx:sourceAccountID(n)` | int | |
| `tx:destinationAccountID()` / `tx:destinationAccountID(n)` | int | |
| `tx:categoryID()` / `tx:categoryID(n)` / `tx:categoryID(nil)` | int or nil | `nil` clears |
| `tx:transactionType()` / `tx:transactionType(n)` | int enum | see table below |
| `tx:sourceAmount()` / `tx:sourceAmount(x)` / `tx:sourceAmount(nil)` | number or nil | |
| `tx:destinationAmount()` / `tx:destinationAmount(x)` / `tx:destinationAmount(nil)` | number or nil | |
| `tx:getSourceAmountWithDecimalPlaces(n)` | number or nil | rounded |
| `tx:getDestinationAmountWithDecimalPlaces(n)` | number or nil | rounded |

Tags:

| Method | Returns |
|---|---|
| `tx:addTag(tagID)` | — |
| `tx:removeTag(tagID)` | — |
| `tx:getTags()` | Lua array of tag IDs |
| `tx:removeAllTags()` | — |

Internal reference numbers:

| Method | Returns |
|---|---|
| `tx:getInternalReferenceNumbers()` | Lua array of strings |
| `tx:addInternalReferenceNumber(s)` | — |
| `tx:setInternalReferenceNumbers({"a","b"})` | — |
| `tx:removeInternalReferenceNumber(s)` | — |

Date/time:

| Method | Effect |
|---|---|
| `tx:transactionDateTimeSetTime(hour, minute)` | Replaces time-of-day |
| `tx:transactionDateTimeAddDate(years, months, days)` | Adds delta |

Transaction type enum:

| Value | Name |
|---|---|
| 0 | UNSPECIFIED |
| 1 | TRANSFER_BETWEEN_ACCOUNTS |
| 2 | INCOME |
| 3 | EXPENSE |
| 4 | REVERSAL |
| 5 | ADJUSTMENT |

#### Helpers

| Method | Returns | Notes |
|---|---|---|
| `helpers:getAccountByID(id)` | account object | fields: `ID`, `Name`, `Currency`, `CurrentBalance`, `Type`, `AccountNumber`, `Iban`, … |
| `helpers:convertCurrency("FROM","TO", amount)` | number | uses stored rates, rounded to target decimals |

#### Patterns & nil-safety

`tx:title()` / `tx:notes()` can be nil on sparse data. Guard with `or ""`:

```lua
local title = tx:title() or ""
local notes = tx:notes() or ""
```

For literal substring match, pass `true` as the 4th arg of `string.find` to
disable Lua patterns:

```lua
string.find(title, "GOOGLE -ADS", 1, true)
```

### Example Scripts

Categorize by title keyword list:

```lua
local keywords = { "GOOGLE -ADS", "MERCHANT X" }
for _, keyword in ipairs(keywords) do
    if string.find(tx:title(), keyword, 1, true) then
        tx:categoryID(10)
        break
    end
end
```

Match title OR notes (two sources):

```lua
local title_keywords = { "Alice S", "Alice" }
for _, keyword in ipairs(title_keywords) do
    if string.find(tx:title(), keyword, 1, true) then
        tx:categoryID(36)
        return
    end
end

local notes = tx:notes() or ""
if string.find(notes, "ALICE SURNAME", 1, true) or
   string.find(notes, "021600146217XXXXXXXXXXXXX", 1, true) then
    tx:categoryID(36)
end
```

Reclassify a matched transaction as a transfer to a specific account:

```lua
local title = tx:title() or ""
local notes = tx:notes() or ""
if string.find(title, "Trading 212", 1, true) or
   string.find(notes, "Trading 212", 1, true) then
    tx:transactionType(1)            -- TRANSFER_BETWEEN_ACCOUNTS
    tx:destinationAccountID(47)      -- target account id
end
```

### list_rules

List all transaction rules.

No input parameters. Response is a JSON array of rules (`id`, `title`,
`script`, `enabled`, `sort_order`, `is_final_rule`, `group_name`) or the text
`"No rules found"` when the list is empty.

### dry_run_rule

Test a Lua script against an existing transaction without persisting
changes. Returns the transaction before and after rule execution plus a
boolean indicating whether the script mutated anything.

Input parameters:

| Parameter | Type | Required | Description |
|---|---|---|---|
| `transaction_id` | number | yes | Transaction to test against (use `0` for scheduled rules that create transactions) |
| `script` | string | yes | Lua script to execute |
| `title` | string | no | Display name; defaults to `"Test Rule"` |

Example request:

```json
{
  "transaction_id": 12345,
  "title": "Trading 212 transfer",
  "script": "local title = tx:title() or ''\nif string.find(title, 'Trading 212', 1, true) then\n  tx:transactionType(1)\n  tx:destinationAccountID(47)\nend"
}
```

Example response:

```json
{
  "rule_applied": true,
  "before": { "id": 12345, "title": "Trading 212", "transaction_type": 3, "destination_account_id": 10, "...": "..." },
  "after":  { "id": 12345, "title": "Trading 212", "transaction_type": 1, "destination_account_id": 47, "...": "..." }
}
```

Error responses:

- `"transaction_id parameter is required"` when missing or not a number.
- `"script parameter is required"` when missing or empty.
- `"dry run failed: <wrap>"` — wraps Lua runtime errors and transaction-lookup failures.

### create_rule

Create a new transaction rule.

Input parameters:

| Parameter | Type | Required | Description |
|---|---|---|---|
| `title` | string | yes | Display name |
| `script` | string | yes | Lua script (see Lua API above) |
| `sort_order` | number | no | Execution order within a group (lower runs first, default 0) |
| `enabled` | boolean | no | Whether the rule is active (default `true`) |
| `is_final_rule` | boolean | no | If `true`, stops execution of the rule's group when this rule applies (default `false`) |
| `group_name` | string | no | Group to organize rules; groups run in alphabetical order |

Example request:

```json
{
  "title": "Categorize Google Ads",
  "script": "local keywords = { 'GOOGLE -ADS' }\nfor _, k in ipairs(keywords) do\n  if string.find(tx:title(), k, 1, true) then\n    tx:categoryID(10)\n    break\n  end\nend",
  "sort_order": 100,
  "group_name": "categorization"
}
```

Example response:

```
Rule created with id 42
```

Error responses:

- `"title parameter is required"` / `"script parameter is required"`.
- `"failed to create rule: <wrap>"` — wraps validation / persistence errors.

### update_rule

Update an existing rule. Replaces `title`, `script`, and any optional flags
provided. Always validate the new script with `dry_run_rule` first.

Input parameters:

| Parameter | Type | Required | Description |
|---|---|---|---|
| `id` | number | yes | Rule id to update |
| `title` | string | yes | Display name |
| `script` | string | yes | Lua script (see Lua API above) |
| `sort_order` | number | no | Execution order |
| `enabled` | boolean | no | Active flag |
| `is_final_rule` | boolean | no | Group-stop flag |
| `group_name` | string | no | Group name |

Example response:

```
Rule 42 updated
```

Error responses:

- `"id parameter is required"` / `"title parameter is required"` / `"script parameter is required"`.
- `"failed to update rule: <wrap>"`.

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

