# Query Performance Tips

Guidelines for writing efficient queries against the Go Money database.

## Table Selection Guide

### When to Use Each Table

| Query Type | Best Table | Why |
|------------|------------|-----|
| Current balance | `accounts.current_balance` | Pre-computed, always current |
| Balance at date | `daily_stat` | Pre-computed running totals |
| Balance history/trends | `daily_stat` | Indexed by date, no JOINs needed |
| Net worth over time | `daily_stat` | Aggregate by date, very fast |
| Transaction details | `transactions` | Has all fields |
| Transaction search | `transactions` | Title, notes, references |
| Category breakdown | `transactions` + `categories` | Need JOIN for names |
| Tag analysis | `transactions` + `tags` | Array operators + JOIN |
| Formal ledger | `double_entries` | Debit/credit format |
| Trial balance | `double_entries` | Accounting reports |

### Avoid These Patterns

| ❌ Slow Pattern | ✅ Fast Alternative |
|----------------|---------------------|
| Sum transactions for balance | Use `accounts.current_balance` |
| Scan transactions for balance history | Use `daily_stat` |
| Full table scan without date filter | Add date range filter |
| `SELECT *` for large tables | Select only needed columns |

## Index Utilization

### Key Indexes

```sql
-- Transactions
ix_transaction_date_only (transaction_date_only)
ix_type_date_only (transaction_type, transaction_date_only)
ix_source_account (source_account_id)
ix_dest_account (destination_account_id)
ix_category_id (category_id)

-- Daily Stats
daily_stat_pk (account_id, date)
ix_latest_stat (account_id, date DESC)

-- Double Entries
ix_double_entries_transaction_date (account_id, transaction_date)
```

### Writing Index-Friendly Queries

```sql
-- ✅ Uses ix_transaction_date_only
WHERE transaction_date_only >= '2024-01-01'

-- ❌ Function on column prevents index use
WHERE EXTRACT(YEAR FROM transaction_date_only) = 2024

-- ✅ Better alternative
WHERE transaction_date_only >= '2024-01-01'
  AND transaction_date_only < '2025-01-01'
```

## Date Filtering Best Practices

### Always Filter by Date

```sql
-- ✅ Good: Limits scan to 30 days
SELECT * FROM transactions
WHERE deleted_at IS NULL
  AND transaction_date_only >= CURRENT_DATE - INTERVAL '30 days';

-- ❌ Bad: Scans entire table
SELECT * FROM transactions
WHERE deleted_at IS NULL;
```

### Use DATE_TRUNC for Grouping

```sql
-- ✅ Efficient: Groups at database level
SELECT
    DATE_TRUNC('month', transaction_date_only) as month,
    SUM(amount)
FROM transactions
GROUP BY DATE_TRUNC('month', transaction_date_only);
```

### Date Range vs Year Extraction

```sql
-- ❌ Slow: Function prevents index use
WHERE EXTRACT(YEAR FROM transaction_date_only) = 2024

-- ✅ Fast: Uses index
WHERE transaction_date_only >= '2024-01-01'
  AND transaction_date_only < '2025-01-01'
```

## Soft Delete Handling

### Always Include deleted_at Filter

```sql
-- Every query must include this
WHERE deleted_at IS NULL
```

### Index Impact

Indexes often include `WHERE deleted_at IS NULL`:
```sql
CREATE INDEX ix_active_transactions
ON transactions (transaction_date_only)
WHERE deleted_at IS NULL;
```

## JOIN Optimization

### Order Matters

```sql
-- ✅ Filter before JOIN when possible
SELECT t.*, c.name
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3
  AND t.transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'
  AND t.deleted_at IS NULL;
```

### Avoid Unnecessary JOINs

```sql
-- ❌ JOINs when not needed
SELECT t.id, t.title, t.destination_amount
FROM transactions t
JOIN accounts sa ON sa.id = t.source_account_id
JOIN accounts da ON da.id = t.destination_account_id
WHERE t.deleted_at IS NULL;

-- ✅ No JOINs when account names not needed
SELECT id, title, destination_amount
FROM transactions
WHERE deleted_at IS NULL;
```

## Aggregation Tips

### Filter Early, Aggregate Late

```sql
-- ✅ Filters before aggregation
SELECT category_id, SUM(destination_amount_in_base_currency)
FROM transactions
WHERE transaction_type = 3
  AND transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'
  AND deleted_at IS NULL
GROUP BY category_id;
```

### Use Conditional Aggregates

```sql
-- ✅ Single scan for multiple metrics
SELECT
    SUM(CASE WHEN transaction_type = 2 THEN destination_amount_in_base_currency ELSE 0 END) as income,
    SUM(CASE WHEN transaction_type = 3 THEN destination_amount_in_base_currency ELSE 0 END) as expenses
FROM transactions
WHERE deleted_at IS NULL;

-- ❌ Multiple scans
SELECT SUM(destination_amount_in_base_currency) FROM transactions WHERE transaction_type = 2;
SELECT SUM(destination_amount_in_base_currency) FROM transactions WHERE transaction_type = 3;
```

## Pagination

### Offset-Based (Simple but Slow for Large Offsets)

```sql
SELECT * FROM transactions
WHERE deleted_at IS NULL
ORDER BY transaction_date_time DESC
LIMIT 50 OFFSET 1000;  -- Slow: scans 1050 rows
```

### Cursor-Based (Better for Deep Pagination)

```sql
-- First page
SELECT * FROM transactions
WHERE deleted_at IS NULL
ORDER BY transaction_date_time DESC, id DESC
LIMIT 50;

-- Subsequent pages
SELECT * FROM transactions
WHERE deleted_at IS NULL
  AND (transaction_date_time, id) < (:last_datetime, :last_id)
ORDER BY transaction_date_time DESC, id DESC
LIMIT 50;
```

## Array Field (tag_ids) Performance

### GIN Index for Arrays

```sql
CREATE INDEX ix_tag_ids ON transactions USING GIN(tag_ids);
```

### Efficient Array Operators

```sql
-- ✅ Uses GIN index
WHERE 5 = ANY(tag_ids)
WHERE tag_ids && ARRAY[1, 2, 3]
WHERE tag_ids @> ARRAY[1, 2]

-- Filter by date first, then by tags
WHERE transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'
  AND 5 = ANY(tag_ids)
```

## Using EXPLAIN ANALYZE

### Check Query Plans

```sql
EXPLAIN ANALYZE
SELECT * FROM transactions
WHERE transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'
  AND deleted_at IS NULL;
```

### What to Look For

| Indicator | Good | Bad |
|-----------|------|-----|
| Index Scan | ✅ | |
| Seq Scan (small table) | ✅ | |
| Seq Scan (large table) | | ❌ |
| Rows estimated vs actual | Close | Far apart |
| Execution time | <100ms | >1000ms |

## Common Optimization Patterns

### Pre-aggregate When Possible

```sql
-- Use daily_stat instead of summing transactions
SELECT amount FROM daily_stat
WHERE account_id = 1
ORDER BY date DESC
LIMIT 1;
```

### Limit Result Sets

```sql
-- Always limit exploratory queries
SELECT * FROM transactions
WHERE deleted_at IS NULL
ORDER BY transaction_date_time DESC
LIMIT 100;
```

### Select Only Needed Columns

```sql
-- ✅ Only select what you need
SELECT id, title, destination_amount_in_base_currency
FROM transactions
WHERE deleted_at IS NULL;

-- ❌ Avoid SELECT *
SELECT * FROM transactions WHERE deleted_at IS NULL;
```

## Summary Checklist

- [ ] Always include `deleted_at IS NULL`
- [ ] Add date range filters
- [ ] Use `daily_stat` for balance/trend queries
- [ ] Use `accounts.current_balance` for current balance
- [ ] Avoid functions on indexed columns in WHERE
- [ ] Use cursor pagination for deep offsets
- [ ] Test with EXPLAIN ANALYZE
- [ ] Select only needed columns
- [ ] Use conditional aggregates for multiple metrics
