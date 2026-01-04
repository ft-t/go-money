# Query Safety Rules

Security and validation rules enforced by the Go Money MCP server.

## Read-Only Enforcement

### Allowed Statements

Only SELECT queries are permitted:

```sql
-- ✅ Allowed
SELECT * FROM accounts WHERE deleted_at IS NULL;
SELECT COUNT(*) FROM transactions;
SELECT id, name FROM categories;
```

### Blocked Statements

All data modification statements are blocked:

```sql
-- ❌ Blocked: INSERT
INSERT INTO accounts (name) VALUES ('New Account');

-- ❌ Blocked: UPDATE
UPDATE accounts SET name = 'Changed' WHERE id = 1;

-- ❌ Blocked: DELETE
DELETE FROM accounts WHERE id = 1;

-- ❌ Blocked: DROP
DROP TABLE accounts;

-- ❌ Blocked: TRUNCATE
TRUNCATE TABLE transactions;

-- ❌ Blocked: ALTER
ALTER TABLE accounts ADD COLUMN test TEXT;

-- ❌ Blocked: CREATE
CREATE TABLE test (id INT);
```

### Blocked Clauses in SELECT

```sql
-- ❌ Blocked: SELECT INTO
SELECT * INTO new_table FROM accounts;

-- ❌ Blocked: SELECT FOR UPDATE
SELECT * FROM accounts FOR UPDATE;

-- ❌ Blocked: WITH ... INSERT/UPDATE/DELETE
WITH cte AS (DELETE FROM accounts RETURNING *)
SELECT * FROM cte;
```

## Query Validation

### Statement Type Check

The server validates that queries begin with SELECT (case-insensitive):

```
Query: "SELECT ..."  → ✅ Allowed
Query: "select ..."  → ✅ Allowed
Query: "  SELECT ..." → ✅ Allowed (whitespace trimmed)
Query: "UPDATE ..."  → ❌ Blocked
Query: "DELETE ..."  → ❌ Blocked
```

### Common Table Expressions (CTEs)

CTEs are allowed for complex read queries:

```sql
-- ✅ Allowed: Read-only CTE
WITH monthly AS (
    SELECT DATE_TRUNC('month', transaction_date_only) as month,
           SUM(destination_amount_in_base_currency) as total
    FROM transactions
    WHERE transaction_type = 3 AND deleted_at IS NULL
    GROUP BY DATE_TRUNC('month', transaction_date_only)
)
SELECT * FROM monthly ORDER BY month;
```

### Subqueries

Subqueries are allowed:

```sql
-- ✅ Allowed: Subquery in WHERE
SELECT * FROM accounts
WHERE id IN (SELECT source_account_id FROM transactions WHERE deleted_at IS NULL);

-- ✅ Allowed: Subquery in FROM
SELECT * FROM (
    SELECT category_id, COUNT(*) as count
    FROM transactions
    WHERE deleted_at IS NULL
    GROUP BY category_id
) sub
WHERE count > 10;
```

## Resource Limits

### Query Timeout

| Setting | Value |
|---------|-------|
| Default timeout | 30 seconds |
| Maximum timeout | 60 seconds |

Queries exceeding the timeout are terminated:

```json
{
  "error": "Query timeout exceeded"
}
```

### Result Size Limits

| Setting | Value |
|---------|-------|
| Max rows | 10,000 |
| Max row size | No limit |
| Max total size | 10 MB |

Results exceeding limits are truncated with a warning.

### Query Complexity

No explicit complexity limits, but:
- Very complex JOINs may timeout
- Large Cartesian products may exceed memory
- Deep recursive CTEs may fail

## Data Access

### Accessible Tables

All application tables are accessible:

| Table | Access |
|-------|--------|
| accounts | ✅ Full read |
| transactions | ✅ Full read |
| categories | ✅ Full read |
| tags | ✅ Full read |
| currencies | ✅ Full read |
| daily_stat | ✅ Full read |
| double_entries | ✅ Full read |
| rules | ✅ Full read |
| schedule_rules | ✅ Full read |
| users | ✅ Full read |
| import_deduplication | ✅ Full read |

### Protected Data

No columns are explicitly hidden, but some contain no sensitive data:

| Column | Table | Notes |
|--------|-------|-------|
| password | users | Hashed, not useful |
| extra | accounts | May contain custom data |
| script | rules | Lua code, safe to read |

### System Tables

PostgreSQL system tables are accessible:

```sql
-- ✅ Allowed: Information schema
SELECT * FROM information_schema.tables WHERE table_schema = 'public';

-- ✅ Allowed: Column info
SELECT column_name, data_type FROM information_schema.columns
WHERE table_name = 'accounts';
```

## Error Handling

### Query Errors

Errors are returned with details:

```json
{
  "error": "ERROR: column \"invalid_column\" does not exist (SQLSTATE 42703)"
}
```

### Error Types

| Error | Cause | Resolution |
|-------|-------|------------|
| 42601 | Syntax error | Fix SQL syntax |
| 42P01 | Table not found | Check table name |
| 42703 | Column not found | Check column name |
| 57014 | Query timeout | Simplify query, add LIMIT |
| 22P02 | Invalid input | Check data types |

## Logging and Audit

### Query Logging

All queries are logged:
- Query text
- Execution time
- Result row count
- Timestamp
- Error (if any)

### No Personal Data in Logs

Logs contain query text but:
- No result data is logged
- No user data is captured
- Connection info is minimal

## Best Practices for Safe Queries

### 1. Use Parameterized Patterns

```sql
-- Good: Clear filter values
SELECT * FROM transactions
WHERE transaction_type = 3
  AND transaction_date_only >= '2024-01-01';

-- Avoid: Dynamic SQL construction
-- (MCP server doesn't support parameters, but use clear values)
```

### 2. Limit Results

```sql
-- Good: Limited results
SELECT * FROM transactions WHERE deleted_at IS NULL LIMIT 100;

-- Risk: Unbounded results
SELECT * FROM transactions WHERE deleted_at IS NULL;
```

### 3. Avoid Heavy Operations

```sql
-- Good: Filtered aggregate
SELECT COUNT(*) FROM transactions
WHERE transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'
  AND deleted_at IS NULL;

-- Slow: Full table aggregate
SELECT COUNT(*) FROM transactions;
```

### 4. Use Indexed Filters

```sql
-- Good: Uses index
WHERE transaction_date_only >= '2024-01-01'

-- Slow: Prevents index use
WHERE EXTRACT(YEAR FROM transaction_date_only) = 2024
```

## Security Summary

| Protection | Implementation |
|------------|----------------|
| Read-only access | Only SELECT allowed |
| No data modification | INSERT/UPDATE/DELETE blocked |
| No schema changes | DDL statements blocked |
| Query timeout | 30-second limit |
| Result limits | 10,000 rows max |
| Audit logging | All queries logged |
