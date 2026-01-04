# Common Analytics Queries

Ready-to-use SQL queries for common financial analytics tasks.

## Balance Queries

### Current Net Worth

```sql
SELECT
    SUM(CASE WHEN a.type = 1 THEN a.current_balance ELSE 0 END) as total_assets,
    SUM(CASE WHEN a.type = 4 THEN a.current_balance ELSE 0 END) as total_liabilities,
    SUM(CASE
        WHEN a.type = 1 THEN a.current_balance
        WHEN a.type = 4 THEN -a.current_balance
        ELSE 0
    END) as net_worth
FROM accounts a
WHERE a.type IN (1, 4)
  AND a.deleted_at IS NULL;
```

### Net Worth in Base Currency (Multi-Currency)

```sql
SELECT
    SUM(CASE
        WHEN a.type = 1 THEN a.current_balance * c.rate
        WHEN a.type = 4 THEN -a.current_balance * c.rate
        ELSE 0
    END) as net_worth_base
FROM accounts a
JOIN currencies c ON c.id = a.currency
WHERE a.type IN (1, 4)
  AND a.deleted_at IS NULL;
```

### Account Balances by Type

```sql
SELECT
    CASE a.type
        WHEN 1 THEN 'Asset'
        WHEN 4 THEN 'Liability'
        WHEN 5 THEN 'Expense'
        WHEN 6 THEN 'Income'
        WHEN 7 THEN 'Adjustment'
    END as account_type,
    a.name,
    a.current_balance,
    a.currency
FROM accounts a
WHERE a.deleted_at IS NULL
ORDER BY a.type, a.current_balance DESC;
```

## Spending Analysis

### Total Spending This Month

```sql
SELECT SUM(t.destination_amount_in_base_currency) as total_spending
FROM transactions t
WHERE t.transaction_type = 3  -- Expense
  AND t.deleted_at IS NULL
  AND DATE_TRUNC('month', t.transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE);
```

### Top 10 Expense Categories

```sql
SELECT
    c.name as category,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= DATE_TRUNC('year', CURRENT_DATE)
GROUP BY c.name
ORDER BY total DESC
LIMIT 10;
```

### Monthly Spending by Category

```sql
SELECT
    DATE_TRUNC('month', t.transaction_date_only) as month,
    c.name as category,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
LEFT JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= CURRENT_DATE - INTERVAL '12 months'
GROUP BY month, c.name
ORDER BY month DESC, total DESC;
```

### Daily Spending Trend (Last 30 Days)

```sql
SELECT
    t.transaction_date_only as date,
    SUM(t.destination_amount_in_base_currency) as daily_spending
FROM transactions t
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY t.transaction_date_only
ORDER BY date;
```

### Spending by Tag

```sql
SELECT
    tg.name as tag,
    COUNT(t.id) as transaction_count,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN tags tg ON tg.id = ANY(t.tag_ids)
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
GROUP BY tg.name
ORDER BY total DESC;
```

## Income Analysis

### Total Income This Month

```sql
SELECT SUM(t.destination_amount_in_base_currency) as total_income
FROM transactions t
WHERE t.transaction_type = 2  -- Income
  AND t.deleted_at IS NULL
  AND DATE_TRUNC('month', t.transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE);
```

### Monthly Income vs Expenses

```sql
SELECT
    DATE_TRUNC('month', t.transaction_date_only) as month,
    SUM(CASE WHEN t.transaction_type = 2 THEN t.destination_amount_in_base_currency ELSE 0 END) as income,
    SUM(CASE WHEN t.transaction_type = 3 THEN t.destination_amount_in_base_currency ELSE 0 END) as expenses,
    SUM(CASE WHEN t.transaction_type = 2 THEN t.destination_amount_in_base_currency ELSE 0 END) -
    SUM(CASE WHEN t.transaction_type = 3 THEN t.destination_amount_in_base_currency ELSE 0 END) as savings
FROM transactions t
WHERE t.transaction_type IN (2, 3)
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= CURRENT_DATE - INTERVAL '12 months'
GROUP BY month
ORDER BY month DESC;
```

### Income by Source

```sql
SELECT
    a.name as source,
    SUM(t.source_amount_in_base_currency) as total
FROM transactions t
JOIN accounts a ON a.id = t.source_account_id
WHERE t.transaction_type = 2  -- Income
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= DATE_TRUNC('year', CURRENT_DATE)
GROUP BY a.name
ORDER BY total DESC;
```

## Time-Based Analysis

### Year-over-Year Comparison

```sql
WITH current_year AS (
    SELECT SUM(destination_amount_in_base_currency) as total
    FROM transactions
    WHERE transaction_type = 3
      AND deleted_at IS NULL
      AND EXTRACT(YEAR FROM transaction_date_only) = EXTRACT(YEAR FROM CURRENT_DATE)
),
previous_year AS (
    SELECT SUM(destination_amount_in_base_currency) as total
    FROM transactions
    WHERE transaction_type = 3
      AND deleted_at IS NULL
      AND EXTRACT(YEAR FROM transaction_date_only) = EXTRACT(YEAR FROM CURRENT_DATE) - 1
)
SELECT
    cy.total as current_year_spending,
    py.total as previous_year_spending,
    ROUND((cy.total - py.total) / NULLIF(py.total, 0) * 100, 2) as percent_change
FROM current_year cy, previous_year py;
```

### Weekly Spending Average

```sql
SELECT
    AVG(weekly_total) as avg_weekly_spending
FROM (
    SELECT
        DATE_TRUNC('week', transaction_date_only) as week,
        SUM(destination_amount_in_base_currency) as weekly_total
    FROM transactions
    WHERE transaction_type = 3
      AND deleted_at IS NULL
      AND transaction_date_only >= CURRENT_DATE - INTERVAL '12 weeks'
    GROUP BY DATE_TRUNC('week', transaction_date_only)
) weekly;
```

## Account Activity

### Account Transaction Summary

```sql
SELECT
    a.name as account,
    a.type as account_type,
    COUNT(t.id) as transaction_count,
    SUM(CASE WHEN t.source_account_id = a.id THEN t.source_amount ELSE 0 END) as total_outflow,
    SUM(CASE WHEN t.destination_account_id = a.id THEN t.destination_amount ELSE 0 END) as total_inflow
FROM accounts a
LEFT JOIN transactions t ON (t.source_account_id = a.id OR t.destination_account_id = a.id)
    AND t.deleted_at IS NULL
WHERE a.type IN (1, 4)
  AND a.deleted_at IS NULL
GROUP BY a.id, a.name, a.type
ORDER BY transaction_count DESC;
```

### Largest Transactions

```sql
SELECT
    t.id,
    t.title,
    t.transaction_date_only,
    t.destination_amount_in_base_currency as amount,
    sa.name as from_account,
    da.name as to_account
FROM transactions t
LEFT JOIN accounts sa ON sa.id = t.source_account_id
LEFT JOIN accounts da ON da.id = t.destination_account_id
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
ORDER BY t.destination_amount_in_base_currency DESC
LIMIT 20;
```

### Recent Transactions

```sql
SELECT
    t.id,
    t.title,
    t.transaction_date_only,
    t.transaction_type,
    t.destination_amount,
    t.destination_currency,
    c.name as category
FROM transactions t
LEFT JOIN categories c ON c.id = t.category_id
WHERE t.deleted_at IS NULL
ORDER BY t.transaction_date_time DESC
LIMIT 50;
```

## Data Quality

### Uncategorized Expenses

```sql
SELECT
    t.id,
    t.title,
    t.transaction_date_only,
    t.destination_amount_in_base_currency
FROM transactions t
WHERE t.transaction_type = 3
  AND t.category_id IS NULL
  AND t.deleted_at IS NULL
ORDER BY t.transaction_date_time DESC;
```

### Untagged Transactions

```sql
SELECT
    t.id,
    t.title,
    t.transaction_date_only
FROM transactions t
WHERE (t.tag_ids IS NULL OR ARRAY_LENGTH(t.tag_ids, 1) IS NULL)
  AND t.deleted_at IS NULL
ORDER BY t.transaction_date_time DESC
LIMIT 50;
```

### Potential Duplicates

```sql
SELECT
    t.title,
    t.transaction_date_only,
    t.destination_amount,
    COUNT(*) as count
FROM transactions t
WHERE t.deleted_at IS NULL
GROUP BY t.title, t.transaction_date_only, t.destination_amount
HAVING COUNT(*) > 1
ORDER BY count DESC;
```

## Currency Analysis

### Currency Exposure

```sql
SELECT
    a.currency,
    SUM(CASE WHEN a.type = 1 THEN a.current_balance ELSE 0 END) as assets,
    SUM(CASE WHEN a.type = 4 THEN a.current_balance ELSE 0 END) as liabilities,
    SUM(CASE
        WHEN a.type = 1 THEN a.current_balance
        WHEN a.type = 4 THEN -a.current_balance
        ELSE 0
    END) as net_position
FROM accounts a
WHERE a.type IN (1, 4)
  AND a.deleted_at IS NULL
GROUP BY a.currency
ORDER BY net_position DESC;
```

### Foreign Currency Transactions

```sql
SELECT
    t.fx_source_currency as original_currency,
    COUNT(*) as transaction_count,
    SUM(t.fx_source_amount) as total_in_original,
    SUM(t.destination_amount_in_base_currency) as total_in_base
FROM transactions t
WHERE t.fx_source_currency IS NOT NULL
  AND t.deleted_at IS NULL
GROUP BY t.fx_source_currency
ORDER BY transaction_count DESC;
```

## Using daily_stat for Performance

### Balance History (Fast)

```sql
SELECT
    date,
    SUM(amount) OVER (ORDER BY date) as running_balance
FROM daily_stat
WHERE account_id = :account_id
ORDER BY date;
```

### Net Worth History

```sql
WITH daily_changes AS (
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
FROM daily_changes
ORDER BY date;
```
