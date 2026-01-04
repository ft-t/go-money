# daily_stat Table

The `daily_stat` table stores pre-computed daily balance changes for fast analytics.

## Schema

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| account_id | integer | NO | - | FK to accounts.id (composite PK) |
| date | date | NO | - | Date (composite PK) |
| amount | numeric | YES | - | Net balance change for the day |

## Primary Key

Composite primary key: `(account_id, date)`

## Indexes

| Index | Definition | Purpose |
|-------|------------|---------|
| daily_stat_pk | UNIQUE (account_id, date) | Primary key |
| daily_stat_account_id_index | (account_id) | Filter by account |
| ix_latest_stat | (account_id, date DESC) | Get most recent stats |

## Amount Interpretation

The `amount` field represents the net change in account balance for that day:
- Positive values: balance increased
- Negative values: balance decreased
- Sum of all amounts = current account balance

## Common Queries

### Daily Balance Changes

```sql
SELECT date, amount
FROM daily_stat
WHERE account_id = :account_id
ORDER BY date DESC
LIMIT 30;
```

### Account Balance from Stats

```sql
SELECT SUM(amount) as balance
FROM daily_stat
WHERE account_id = :account_id;
```

### Balance at Specific Date

```sql
SELECT SUM(amount) as balance
FROM daily_stat
WHERE account_id = :account_id
  AND date <= :target_date;
```

### Daily Spending Trend (Last 30 Days)

```sql
SELECT
    ds.date,
    ds.amount as daily_change
FROM daily_stat ds
JOIN accounts a ON a.id = ds.account_id
WHERE a.type = 5  -- Expense accounts
  AND ds.date >= CURRENT_DATE - INTERVAL '30 days'
ORDER BY ds.date;
```

### Monthly Summary from Daily Stats

```sql
SELECT
    DATE_TRUNC('month', date) as month,
    SUM(amount) as monthly_total
FROM daily_stat
WHERE account_id = :account_id
GROUP BY DATE_TRUNC('month', date)
ORDER BY month DESC;
```

### Running Balance Over Time

```sql
SELECT
    date,
    SUM(amount) OVER (ORDER BY date) as running_balance
FROM daily_stat
WHERE account_id = :account_id
ORDER BY date;
```

### Net Worth Over Time

```sql
WITH daily_totals AS (
    SELECT
        ds.date,
        SUM(CASE WHEN a.type = 1 THEN ds.amount ELSE 0 END) as asset_change,
        SUM(CASE WHEN a.type = 4 THEN ds.amount ELSE 0 END) as liability_change
    FROM daily_stat ds
    JOIN accounts a ON a.id = ds.account_id
    WHERE a.type IN (1, 4)
    GROUP BY ds.date
)
SELECT
    date,
    SUM(asset_change - liability_change) OVER (ORDER BY date) as net_worth
FROM daily_totals
ORDER BY date;
```

### Most Recent Activity Date

```sql
SELECT MAX(date) as last_activity
FROM daily_stat
WHERE account_id = :account_id;
```

### Accounts with Recent Activity

```sql
SELECT DISTINCT ds.account_id, a.name
FROM daily_stat ds
JOIN accounts a ON a.id = ds.account_id
WHERE ds.date >= CURRENT_DATE - INTERVAL '7 days'
  AND a.deleted_at IS NULL;
```

### Days with Spending

```sql
SELECT
    date,
    SUM(amount) as total_spending
FROM daily_stat ds
JOIN accounts a ON a.id = ds.account_id
WHERE a.type = 5  -- Expense
  AND ds.date >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY date
HAVING SUM(amount) > 0
ORDER BY date;
```

## Performance Tips

**When to use daily_stat vs transactions:**

| Use daily_stat | Use transactions |
|----------------|------------------|
| Balance calculations | Transaction details |
| Trend analysis | Category/tag filtering |
| Time series charts | Search by title/notes |
| Summary statistics | Individual records |

The daily_stat table is optimized for aggregate queries and should be preferred over scanning the transactions table for balance calculations.

## Notes

- Pre-computed for fast balance and trend queries
- One row per account per day with activity
- No entry means no balance change that day
- Updated automatically when transactions are created/modified
- Does not include soft-deleted transactions


---

## See Also

- [Daily Stats](../../business-logic/statistics/daily-stats.md) - Recalculation logic
- [Balance Queries](../../analytics/query-patterns/balance-queries.md) - Historical balance queries
- [Schema Quick-Ref](../QUICK-REF.md) - All tables at a glance
