# Category Analysis

Queries for analyzing spending and income by category.

## Basic Category Reports

### All Categories with Totals

```sql
SELECT
    c.id,
    c.name,
    COUNT(t.id) as transaction_count,
    SUM(t.destination_amount_in_base_currency) as total_amount
FROM categories c
LEFT JOIN transactions t ON t.category_id = c.id
    AND t.deleted_at IS NULL
    AND t.transaction_type = 3  -- Expense
WHERE c.deleted_at IS NULL
GROUP BY c.id, c.name
ORDER BY total_amount DESC NULLS LAST;
```

### Top 10 Expense Categories

```sql
SELECT
    c.name as category,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3  -- Expense
  AND t.deleted_at IS NULL
  AND c.deleted_at IS NULL
GROUP BY c.name
ORDER BY total DESC
LIMIT 10;
```

### Categories with No Transactions

```sql
SELECT c.id, c.name
FROM categories c
LEFT JOIN transactions t ON t.category_id = c.id
    AND t.deleted_at IS NULL
WHERE c.deleted_at IS NULL
GROUP BY c.id, c.name
HAVING COUNT(t.id) = 0;
```

## Time-Based Category Analysis

### Monthly Spending by Category

```sql
SELECT
    DATE_TRUNC('month', t.transaction_date_only) as month,
    c.name as category,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= CURRENT_DATE - INTERVAL '12 months'
GROUP BY month, c.name
ORDER BY month DESC, total DESC;
```

### Category Trend (Specific Category Over Time)

```sql
SELECT
    DATE_TRUNC('month', t.transaction_date_only) as month,
    SUM(t.destination_amount_in_base_currency) as total,
    COUNT(*) as transaction_count,
    AVG(t.destination_amount_in_base_currency) as avg_amount
FROM transactions t
WHERE t.category_id = :category_id
  AND t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= CURRENT_DATE - INTERVAL '12 months'
GROUP BY month
ORDER BY month;
```

### This Month vs Last Month by Category

```sql
WITH current_month AS (
    SELECT
        c.name,
        SUM(t.destination_amount_in_base_currency) as total
    FROM transactions t
    JOIN categories c ON c.id = t.category_id
    WHERE t.transaction_type = 3
      AND t.deleted_at IS NULL
      AND DATE_TRUNC('month', t.transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE)
    GROUP BY c.name
),
last_month AS (
    SELECT
        c.name,
        SUM(t.destination_amount_in_base_currency) as total
    FROM transactions t
    JOIN categories c ON c.id = t.category_id
    WHERE t.transaction_type = 3
      AND t.deleted_at IS NULL
      AND DATE_TRUNC('month', t.transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')
    GROUP BY c.name
)
SELECT
    COALESCE(cm.name, lm.name) as category,
    COALESCE(cm.total, 0) as current_month,
    COALESCE(lm.total, 0) as last_month,
    COALESCE(cm.total, 0) - COALESCE(lm.total, 0) as change,
    CASE
        WHEN lm.total IS NULL OR lm.total = 0 THEN NULL
        ELSE ROUND((cm.total - lm.total) / lm.total * 100, 1)
    END as percent_change
FROM current_month cm
FULL OUTER JOIN last_month lm ON cm.name = lm.name
ORDER BY current_month DESC NULLS LAST;
```

## Percentage Analysis

### Category as Percentage of Total

```sql
WITH category_totals AS (
    SELECT
        c.name,
        SUM(t.destination_amount_in_base_currency) as total
    FROM transactions t
    JOIN categories c ON c.id = t.category_id
    WHERE t.transaction_type = 3
      AND t.deleted_at IS NULL
      AND t.transaction_date_only >= DATE_TRUNC('month', CURRENT_DATE)
    GROUP BY c.name
),
grand_total AS (
    SELECT SUM(total) as total FROM category_totals
)
SELECT
    ct.name,
    ct.total,
    ROUND(ct.total / gt.total * 100, 2) as percentage
FROM category_totals ct, grand_total gt
WHERE gt.total > 0
ORDER BY percentage DESC;
```

### Monthly Category Breakdown (Percentages)

```sql
WITH monthly_data AS (
    SELECT
        DATE_TRUNC('month', t.transaction_date_only) as month,
        c.name as category,
        SUM(t.destination_amount_in_base_currency) as amount
    FROM transactions t
    JOIN categories c ON c.id = t.category_id
    WHERE t.transaction_type = 3
      AND t.deleted_at IS NULL
      AND t.transaction_date_only >= CURRENT_DATE - INTERVAL '6 months'
    GROUP BY month, c.name
),
monthly_totals AS (
    SELECT month, SUM(amount) as total
    FROM monthly_data
    GROUP BY month
)
SELECT
    md.month,
    md.category,
    md.amount,
    ROUND(md.amount / mt.total * 100, 1) as percentage
FROM monthly_data md
JOIN monthly_totals mt ON mt.month = md.month
WHERE mt.total > 0
ORDER BY md.month DESC, percentage DESC;
```

## Uncategorized Analysis

### Uncategorized Spending Total

```sql
SELECT
    COUNT(*) as transaction_count,
    SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3
  AND category_id IS NULL
  AND deleted_at IS NULL;
```

### Uncategorized by Month

```sql
SELECT
    DATE_TRUNC('month', transaction_date_only) as month,
    COUNT(*) as count,
    SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3
  AND category_id IS NULL
  AND deleted_at IS NULL
GROUP BY month
ORDER BY month DESC;
```

### Uncategorized Transactions List

```sql
SELECT
    id,
    title,
    transaction_date_only,
    destination_amount_in_base_currency
FROM transactions
WHERE transaction_type = 3
  AND category_id IS NULL
  AND deleted_at IS NULL
ORDER BY destination_amount_in_base_currency DESC
LIMIT 50;
```

## Category Statistics

### Category Spending Statistics

```sql
SELECT
    c.name as category,
    COUNT(t.id) as count,
    SUM(t.destination_amount_in_base_currency) as total,
    AVG(t.destination_amount_in_base_currency) as average,
    MIN(t.destination_amount_in_base_currency) as min_amount,
    MAX(t.destination_amount_in_base_currency) as max_amount,
    STDDEV(t.destination_amount_in_base_currency) as std_dev
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
GROUP BY c.name
ORDER BY total DESC;
```

### Most Active Categories (by Transaction Count)

```sql
SELECT
    c.name,
    COUNT(*) as transaction_count,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY c.name
ORDER BY transaction_count DESC
LIMIT 10;
```

## Year-over-Year Comparison

### Category YoY Change

```sql
WITH this_year AS (
    SELECT
        c.name,
        SUM(t.destination_amount_in_base_currency) as total
    FROM transactions t
    JOIN categories c ON c.id = t.category_id
    WHERE t.transaction_type = 3
      AND t.deleted_at IS NULL
      AND EXTRACT(YEAR FROM t.transaction_date_only) = EXTRACT(YEAR FROM CURRENT_DATE)
    GROUP BY c.name
),
last_year AS (
    SELECT
        c.name,
        SUM(t.destination_amount_in_base_currency) as total
    FROM transactions t
    JOIN categories c ON c.id = t.category_id
    WHERE t.transaction_type = 3
      AND t.deleted_at IS NULL
      AND EXTRACT(YEAR FROM t.transaction_date_only) = EXTRACT(YEAR FROM CURRENT_DATE) - 1
    GROUP BY c.name
)
SELECT
    COALESCE(ty.name, ly.name) as category,
    COALESCE(ty.total, 0) as this_year,
    COALESCE(ly.total, 0) as last_year,
    ROUND((COALESCE(ty.total, 0) - COALESCE(ly.total, 0)) /
          NULLIF(ly.total, 0) * 100, 1) as percent_change
FROM this_year ty
FULL OUTER JOIN last_year ly ON ty.name = ly.name
ORDER BY this_year DESC NULLS LAST;
```

## Combined with Tags

### Category + Tag Analysis

```sql
SELECT
    c.name as category,
    tg.name as tag,
    COUNT(*) as count,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN categories c ON c.id = t.category_id
JOIN tags tg ON tg.id = ANY(t.tag_ids)
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND tg.deleted_at IS NULL
GROUP BY c.name, tg.name
ORDER BY total DESC
LIMIT 20;
```

## Performance Notes

- Create index on `category_id` for faster JOINs
- Filter by date range to limit data scanned
- Use `destination_amount_in_base_currency` for consistent currency reporting
- Categories are soft-deleted, always check `deleted_at IS NULL`
