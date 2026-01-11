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
