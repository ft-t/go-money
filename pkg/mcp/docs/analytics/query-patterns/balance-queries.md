# Balance Queries

Queries for account balances, net worth, and historical balance tracking.

## Current Balances

### Single Account Balance

```sql
SELECT current_balance, currency
FROM accounts
WHERE id = :account_id
  AND deleted_at IS NULL;
```

### All Account Balances

```sql
SELECT
    id,
    name,
    type,
    current_balance,
    currency
FROM accounts
WHERE deleted_at IS NULL
ORDER BY type, display_order NULLS LAST;
```

### Balances by Account Type

```sql
SELECT
    CASE type
        WHEN 1 THEN 'Asset'
        WHEN 4 THEN 'Liability'
        WHEN 5 THEN 'Expense'
        WHEN 6 THEN 'Income'
        WHEN 7 THEN 'Adjustment'
    END as account_type,
    COUNT(*) as account_count,
    SUM(current_balance) as total_balance
FROM accounts
WHERE deleted_at IS NULL
GROUP BY type
ORDER BY type;
```

## Net Worth

### Simple Net Worth (Same Currency)

```sql
SELECT
    SUM(CASE WHEN type = 1 THEN current_balance ELSE 0 END) as total_assets,
    SUM(CASE WHEN type = 4 THEN current_balance ELSE 0 END) as total_liabilities,
    SUM(CASE
        WHEN type = 1 THEN current_balance
        WHEN type = 4 THEN -current_balance
        ELSE 0
    END) as net_worth
FROM accounts
WHERE type IN (1, 4)
  AND deleted_at IS NULL;
```

### Net Worth by Currency

```sql
SELECT
    currency,
    SUM(CASE WHEN type = 1 THEN current_balance ELSE 0 END) as assets,
    SUM(CASE WHEN type = 4 THEN current_balance ELSE 0 END) as liabilities,
    SUM(CASE
        WHEN type = 1 THEN current_balance
        WHEN type = 4 THEN -current_balance
        ELSE 0
    END) as net_worth
FROM accounts
WHERE type IN (1, 4)
  AND deleted_at IS NULL
GROUP BY currency
ORDER BY net_worth DESC;
```

### Multi-Currency Net Worth (Converted to Base)

```sql
SELECT
    SUM(CASE
        WHEN a.type = 1 THEN a.current_balance / c.rate
        WHEN a.type = 4 THEN -a.current_balance / c.rate
        ELSE 0
    END) as net_worth_in_base
FROM accounts a
JOIN currencies c ON c.id = a.currency
WHERE a.type IN (1, 4)
  AND a.deleted_at IS NULL;
```

## Historical Balances

### Balance at Specific Date

Using daily_stat (fastest):

```sql
SELECT amount as balance
FROM daily_stat
WHERE account_id = :account_id
  AND date <= :target_date
ORDER BY date DESC
LIMIT 1;
```

### Balance History (Last 30 Days)

```sql
SELECT date, amount as balance
FROM daily_stat
WHERE account_id = :account_id
  AND date >= CURRENT_DATE - INTERVAL '30 days'
ORDER BY date;
```

### Monthly End Balances

```sql
SELECT DISTINCT ON (DATE_TRUNC('month', date))
    DATE_TRUNC('month', date) as month,
    amount as end_of_month_balance
FROM daily_stat
WHERE account_id = :account_id
ORDER BY DATE_TRUNC('month', date), date DESC;
```

## Net Worth Over Time

### Daily Net Worth History

```sql
WITH daily_balances AS (
    SELECT
        ds.date,
        a.type,
        ds.amount
    FROM daily_stat ds
    JOIN accounts a ON a.id = ds.account_id
    WHERE a.type IN (1, 4)
      AND a.deleted_at IS NULL
)
SELECT
    date,
    SUM(CASE WHEN type = 1 THEN amount ELSE 0 END) as assets,
    SUM(CASE WHEN type = 4 THEN amount ELSE 0 END) as liabilities,
    SUM(CASE
        WHEN type = 1 THEN amount
        WHEN type = 4 THEN -amount
        ELSE 0
    END) as net_worth
FROM daily_balances
GROUP BY date
ORDER BY date;
```

### Monthly Net Worth Trend

```sql
WITH monthly_balances AS (
    SELECT DISTINCT ON (a.id, DATE_TRUNC('month', ds.date))
        DATE_TRUNC('month', ds.date) as month,
        a.id as account_id,
        a.type,
        ds.amount
    FROM daily_stat ds
    JOIN accounts a ON a.id = ds.account_id
    WHERE a.type IN (1, 4)
      AND a.deleted_at IS NULL
    ORDER BY a.id, DATE_TRUNC('month', ds.date), ds.date DESC
)
SELECT
    month,
    SUM(CASE WHEN type = 1 THEN amount ELSE -amount END) as net_worth
FROM monthly_balances
GROUP BY month
ORDER BY month;
```

## Balance Changes

### Daily Balance Change

```sql
SELECT
    ds1.date,
    ds1.amount - COALESCE(ds2.amount, 0) as daily_change
FROM daily_stat ds1
LEFT JOIN daily_stat ds2
    ON ds2.account_id = ds1.account_id
    AND ds2.date = ds1.date - INTERVAL '1 day'
WHERE ds1.account_id = :account_id
  AND ds1.date >= CURRENT_DATE - INTERVAL '30 days'
ORDER BY ds1.date;
```

### Month-over-Month Change

```sql
WITH monthly AS (
    SELECT DISTINCT ON (DATE_TRUNC('month', date))
        DATE_TRUNC('month', date) as month,
        amount
    FROM daily_stat
    WHERE account_id = :account_id
    ORDER BY DATE_TRUNC('month', date), date DESC
)
SELECT
    m1.month,
    m1.amount as balance,
    m1.amount - COALESCE(m2.amount, 0) as change,
    ROUND((m1.amount - COALESCE(m2.amount, 0)) /
          NULLIF(ABS(m2.amount), 0) * 100, 2) as percent_change
FROM monthly m1
LEFT JOIN monthly m2 ON m2.month = m1.month - INTERVAL '1 month'
ORDER BY m1.month DESC;
```

## Account Comparison

### Top 5 Asset Accounts by Balance

```sql
SELECT name, current_balance, currency
FROM accounts
WHERE type = 1  -- Asset
  AND deleted_at IS NULL
ORDER BY current_balance DESC
LIMIT 5;
```

### Accounts with Highest Growth (30 Days)

```sql
WITH balance_change AS (
    SELECT
        ds.account_id,
        MAX(CASE WHEN ds.date = CURRENT_DATE THEN ds.amount END) as current_balance,
        MAX(CASE WHEN ds.date = CURRENT_DATE - INTERVAL '30 days' THEN ds.amount END) as old_balance
    FROM daily_stat ds
    WHERE ds.date IN (CURRENT_DATE, CURRENT_DATE - INTERVAL '30 days')
    GROUP BY ds.account_id
)
SELECT
    a.name,
    bc.current_balance,
    bc.old_balance,
    bc.current_balance - COALESCE(bc.old_balance, 0) as change
FROM balance_change bc
JOIN accounts a ON a.id = bc.account_id
WHERE a.type = 1  -- Assets only
  AND a.deleted_at IS NULL
ORDER BY change DESC
LIMIT 10;
```

## Verification Queries

### Verify Balance Consistency

Check that current_balance matches latest daily_stat:

```sql
SELECT
    a.id,
    a.name,
    a.current_balance,
    ds.amount as stat_balance,
    a.current_balance - COALESCE(ds.amount, 0) as difference
FROM accounts a
LEFT JOIN (
    SELECT DISTINCT ON (account_id)
        account_id, amount
    FROM daily_stat
    ORDER BY account_id, date DESC
) ds ON ds.account_id = a.id
WHERE a.deleted_at IS NULL
  AND ABS(a.current_balance - COALESCE(ds.amount, 0)) > 0.01;
-- Should return no rows
```

### Verify Ledger Balance

```sql
SELECT
    SUM(CASE WHEN is_debit THEN amount_in_base_currency ELSE 0 END) as total_debits,
    SUM(CASE WHEN NOT is_debit THEN amount_in_base_currency ELSE 0 END) as total_credits,
    SUM(CASE WHEN is_debit THEN amount_in_base_currency
             ELSE -amount_in_base_currency END) as difference
FROM double_entries
WHERE deleted_at IS NULL;
-- difference should be 0
```

## Performance Notes

- Use `daily_stat` for historical balance queries - it's pre-computed
- Use `accounts.current_balance` for current balance - it's always up-to-date
- Avoid calculating balances by summing transactions - slow for large datasets
- The `daily_stat.amount` field is a running total, not a daily change
