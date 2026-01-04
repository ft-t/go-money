# Analytics Quick Reference

> Query templates and patterns. For full examples, see [common-queries.md](common-queries.md).

## Table Selection Guide

| Question Type | Use Table | Why |
|---------------|-----------|-----|
| Current balance | `accounts` | `current_balance` is live |
| Historical balance | `daily_stat` | Pre-computed daily snapshots |
| Transaction details | `transactions` | Full transaction data |
| Spending by category | `transactions` + `categories` | JOIN for category names |
| Spending by tag | `transactions` | `tag_ids` array column |
| Ledger entries | `double_entries` | Formal accounting view |

---

## Essential Filters (Always Include)

```sql
WHERE deleted_at IS NULL           -- Exclude soft-deleted records
```

---

## Query Templates

### Net Worth
```sql
SELECT SUM(CASE
    WHEN type = 1 THEN current_balance
    WHEN type = 4 THEN -current_balance
    ELSE 0
END) as net_worth
FROM accounts
WHERE type IN (1, 4) AND deleted_at IS NULL;
```

### Total Spending This Month
```sql
SELECT SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE);
```

### Total Income This Month
```sql
SELECT SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 2
  AND deleted_at IS NULL
  AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE);
```

### Spending by Category
```sql
SELECT c.name, SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3 AND t.deleted_at IS NULL
GROUP BY c.name
ORDER BY total DESC;
```

### Transactions by Tag
```sql
SELECT t.*
FROM transactions t
WHERE :tag_id = ANY(t.tag_ids)
  AND t.deleted_at IS NULL;
```

### Recent Transactions
```sql
SELECT t.*, c.name as category_name
FROM transactions t
LEFT JOIN categories c ON c.id = t.category_id
WHERE t.deleted_at IS NULL
ORDER BY t.transaction_date_time DESC
LIMIT 20;
```

### Account Balance at Date
```sql
SELECT amount as balance
FROM daily_stat
WHERE account_id = :account_id
  AND date <= :target_date
ORDER BY date DESC
LIMIT 1;
```

### Monthly Spending Trend
```sql
SELECT DATE_TRUNC('month', transaction_date_only) as month,
       SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND transaction_date_only >= CURRENT_DATE - INTERVAL '12 months'
GROUP BY DATE_TRUNC('month', transaction_date_only)
ORDER BY month;
```

### Daily Spending (Last 7 Days)
```sql
SELECT transaction_date_only as date,
       SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND transaction_date_only >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY transaction_date_only
ORDER BY date;
```

---

## Date Patterns

| Period | SQL Pattern |
|--------|-------------|
| This month | `DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE)` |
| Last month | `DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')` |
| This year | `transaction_date_only >= DATE_TRUNC('year', CURRENT_DATE)` |
| Last 30 days | `transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'` |
| Last 7 days | `transaction_date_only >= CURRENT_DATE - INTERVAL '7 days'` |
| Specific date | `transaction_date_only = '2024-01-15'` |
| Date range | `transaction_date_only BETWEEN '2024-01-01' AND '2024-12-31'` |

---

## Tag Array Operations

```sql
-- Has specific tag
WHERE :tag_id = ANY(tag_ids)

-- Has ALL of these tags
WHERE tag_ids @> ARRAY[1, 2]::integer[]

-- Has ANY of these tags
WHERE tag_ids && ARRAY[1, 2]::integer[]

-- Has no tags
WHERE tag_ids IS NULL OR ARRAY_LENGTH(tag_ids, 1) IS NULL
```

---

## Amount Fields

| Field | Use For |
|-------|---------|
| `source_amount` | Original amount in source currency |
| `destination_amount` | Original amount in destination currency |
| `source_amount_in_base_currency` | Converted for aggregation |
| `destination_amount_in_base_currency` | Converted for aggregation |

**For aggregations, always use `*_in_base_currency` fields.**

---

## Performance Tips

1. **Use `daily_stat` for balance history** - faster than summing transactions
2. **Filter by `transaction_date_only`** - indexed, faster than `transaction_date_time`
3. **Add `LIMIT`** - always limit exploratory queries
4. **Use `transaction_type` filter** - indexed, reduces scan
5. **Avoid functions on indexed columns** - `EXTRACT(YEAR FROM date)` prevents index use

---

## See Also
- [Common Queries](common-queries.md) - 25+ complete examples
- [Performance Tips](performance-tips.md) - Optimization guide
- [Balance Queries](query-patterns/balance-queries.md) - Net worth patterns
- [Time Series](query-patterns/time-series.md) - Trend analysis
