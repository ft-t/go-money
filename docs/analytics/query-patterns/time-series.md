# Time Series Queries

Queries for trend analysis, comparisons, and time-based aggregations.

## Aggregation Periods

### Daily Totals

```sql
SELECT
    transaction_date_only as date,
    SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3  -- Expense
  AND deleted_at IS NULL
  AND transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY transaction_date_only
ORDER BY date;
```

### Weekly Totals

```sql
SELECT
    DATE_TRUNC('week', transaction_date_only) as week_start,
    SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND transaction_date_only >= CURRENT_DATE - INTERVAL '12 weeks'
GROUP BY DATE_TRUNC('week', transaction_date_only)
ORDER BY week_start;
```

### Monthly Totals

```sql
SELECT
    DATE_TRUNC('month', transaction_date_only) as month,
    SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND transaction_date_only >= CURRENT_DATE - INTERVAL '12 months'
GROUP BY DATE_TRUNC('month', transaction_date_only)
ORDER BY month;
```

### Quarterly Totals

```sql
SELECT
    DATE_TRUNC('quarter', transaction_date_only) as quarter,
    SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND transaction_date_only >= CURRENT_DATE - INTERVAL '2 years'
GROUP BY DATE_TRUNC('quarter', transaction_date_only)
ORDER BY quarter;
```

### Yearly Totals

```sql
SELECT
    EXTRACT(YEAR FROM transaction_date_only) as year,
    SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
GROUP BY EXTRACT(YEAR FROM transaction_date_only)
ORDER BY year;
```

## Income vs Expenses Over Time

### Monthly Comparison

```sql
SELECT
    DATE_TRUNC('month', transaction_date_only) as month,
    SUM(CASE WHEN transaction_type = 2 THEN destination_amount_in_base_currency ELSE 0 END) as income,
    SUM(CASE WHEN transaction_type = 3 THEN destination_amount_in_base_currency ELSE 0 END) as expenses,
    SUM(CASE
        WHEN transaction_type = 2 THEN destination_amount_in_base_currency
        WHEN transaction_type = 3 THEN -destination_amount_in_base_currency
        ELSE 0
    END) as net_savings
FROM transactions
WHERE transaction_type IN (2, 3)
  AND deleted_at IS NULL
  AND transaction_date_only >= CURRENT_DATE - INTERVAL '12 months'
GROUP BY DATE_TRUNC('month', transaction_date_only)
ORDER BY month;
```

### Savings Rate Over Time

```sql
WITH monthly AS (
    SELECT
        DATE_TRUNC('month', transaction_date_only) as month,
        SUM(CASE WHEN transaction_type = 2 THEN destination_amount_in_base_currency ELSE 0 END) as income,
        SUM(CASE WHEN transaction_type = 3 THEN destination_amount_in_base_currency ELSE 0 END) as expenses
    FROM transactions
    WHERE transaction_type IN (2, 3)
      AND deleted_at IS NULL
      AND transaction_date_only >= CURRENT_DATE - INTERVAL '12 months'
    GROUP BY DATE_TRUNC('month', transaction_date_only)
)
SELECT
    month,
    income,
    expenses,
    income - expenses as savings,
    CASE WHEN income > 0
        THEN ROUND((income - expenses) / income * 100, 1)
        ELSE 0
    END as savings_rate_percent
FROM monthly
ORDER BY month;
```

## Period Comparisons

### This Month vs Last Month

```sql
WITH current_month AS (
    SELECT SUM(destination_amount_in_base_currency) as total
    FROM transactions
    WHERE transaction_type = 3
      AND deleted_at IS NULL
      AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE)
),
last_month AS (
    SELECT SUM(destination_amount_in_base_currency) as total
    FROM transactions
    WHERE transaction_type = 3
      AND deleted_at IS NULL
      AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')
)
SELECT
    cm.total as this_month,
    lm.total as last_month,
    cm.total - lm.total as difference,
    ROUND((cm.total - lm.total) / NULLIF(lm.total, 0) * 100, 1) as percent_change
FROM current_month cm, last_month lm;
```

### Year-over-Year

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
    cy.total as this_year,
    py.total as last_year,
    ROUND((cy.total - py.total) / NULLIF(py.total, 0) * 100, 1) as yoy_percent_change
FROM current_year cy, previous_year py;
```

### Same Month Last Year

```sql
WITH this_month AS (
    SELECT SUM(destination_amount_in_base_currency) as total
    FROM transactions
    WHERE transaction_type = 3
      AND deleted_at IS NULL
      AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE)
),
last_year_same_month AS (
    SELECT SUM(destination_amount_in_base_currency) as total
    FROM transactions
    WHERE transaction_type = 3
      AND deleted_at IS NULL
      AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 year')
)
SELECT
    tm.total as this_month,
    lm.total as same_month_last_year,
    ROUND((tm.total - lm.total) / NULLIF(lm.total, 0) * 100, 1) as percent_change
FROM this_month tm, last_year_same_month lm;
```

## Rolling Averages

### 7-Day Rolling Average

```sql
WITH daily AS (
    SELECT
        transaction_date_only as date,
        SUM(destination_amount_in_base_currency) as daily_total
    FROM transactions
    WHERE transaction_type = 3
      AND deleted_at IS NULL
      AND transaction_date_only >= CURRENT_DATE - INTERVAL '37 days'
    GROUP BY transaction_date_only
)
SELECT
    date,
    daily_total,
    AVG(daily_total) OVER (
        ORDER BY date
        ROWS BETWEEN 6 PRECEDING AND CURRENT ROW
    ) as rolling_7day_avg
FROM daily
WHERE date >= CURRENT_DATE - INTERVAL '30 days'
ORDER BY date;
```

### 30-Day Rolling Total

```sql
WITH daily AS (
    SELECT
        transaction_date_only as date,
        SUM(destination_amount_in_base_currency) as daily_total
    FROM transactions
    WHERE transaction_type = 3
      AND deleted_at IS NULL
      AND transaction_date_only >= CURRENT_DATE - INTERVAL '60 days'
    GROUP BY transaction_date_only
)
SELECT
    date,
    SUM(daily_total) OVER (
        ORDER BY date
        ROWS BETWEEN 29 PRECEDING AND CURRENT ROW
    ) as rolling_30day_total
FROM daily
WHERE date >= CURRENT_DATE - INTERVAL '30 days'
ORDER BY date;
```

## Cumulative Totals

### Year-to-Date Cumulative Spending

```sql
WITH daily AS (
    SELECT
        transaction_date_only as date,
        SUM(destination_amount_in_base_currency) as daily_total
    FROM transactions
    WHERE transaction_type = 3
      AND deleted_at IS NULL
      AND EXTRACT(YEAR FROM transaction_date_only) = EXTRACT(YEAR FROM CURRENT_DATE)
    GROUP BY transaction_date_only
)
SELECT
    date,
    daily_total,
    SUM(daily_total) OVER (ORDER BY date) as cumulative_ytd
FROM daily
ORDER BY date;
```

### Monthly Cumulative

```sql
WITH monthly AS (
    SELECT
        DATE_TRUNC('month', transaction_date_only) as month,
        SUM(destination_amount_in_base_currency) as monthly_total
    FROM transactions
    WHERE transaction_type = 3
      AND deleted_at IS NULL
      AND EXTRACT(YEAR FROM transaction_date_only) = EXTRACT(YEAR FROM CURRENT_DATE)
    GROUP BY DATE_TRUNC('month', transaction_date_only)
)
SELECT
    month,
    monthly_total,
    SUM(monthly_total) OVER (ORDER BY month) as cumulative
FROM monthly
ORDER BY month;
```

## Day of Week Analysis

### Spending by Day of Week

```sql
SELECT
    EXTRACT(DOW FROM transaction_date_only) as day_num,
    TO_CHAR(transaction_date_only, 'Day') as day_name,
    COUNT(*) as transaction_count,
    SUM(destination_amount_in_base_currency) as total,
    AVG(destination_amount_in_base_currency) as avg_per_transaction
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND transaction_date_only >= CURRENT_DATE - INTERVAL '90 days'
GROUP BY EXTRACT(DOW FROM transaction_date_only), TO_CHAR(transaction_date_only, 'Day')
ORDER BY day_num;
```

## Using daily_stat for Performance

### Net Worth History (Fast)

```sql
WITH daily_net_worth AS (
    SELECT
        ds.date,
        SUM(CASE
            WHEN a.type = 1 THEN ds.amount
            WHEN a.type = 4 THEN -ds.amount
            ELSE 0
        END) as net_worth
    FROM daily_stat ds
    JOIN accounts a ON a.id = ds.account_id
    WHERE a.type IN (1, 4)
      AND a.deleted_at IS NULL
    GROUP BY ds.date
)
SELECT date, net_worth
FROM daily_net_worth
ORDER BY date;
```

### Balance Trend (Fast)

```sql
SELECT date, amount as balance
FROM daily_stat
WHERE account_id = :account_id
  AND date >= CURRENT_DATE - INTERVAL '365 days'
ORDER BY date;
```

## Filling Date Gaps

### Generate Continuous Date Series

```sql
WITH date_series AS (
    SELECT generate_series(
        CURRENT_DATE - INTERVAL '30 days',
        CURRENT_DATE,
        '1 day'::interval
    )::date as date
),
daily_totals AS (
    SELECT
        transaction_date_only as date,
        SUM(destination_amount_in_base_currency) as total
    FROM transactions
    WHERE transaction_type = 3
      AND deleted_at IS NULL
      AND transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'
    GROUP BY transaction_date_only
)
SELECT
    ds.date,
    COALESCE(dt.total, 0) as spending
FROM date_series ds
LEFT JOIN daily_totals dt ON dt.date = ds.date
ORDER BY ds.date;
```

## Performance Notes

- Use `DATE_TRUNC()` for efficient grouping
- Use `daily_stat` for balance/net worth trends (pre-computed)
- Create index on `transaction_date_only` for date-based queries
- Limit date ranges to avoid full table scans
- Window functions can be expensive on large datasets
