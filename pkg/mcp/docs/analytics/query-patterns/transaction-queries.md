# Transaction Queries

Queries for listing, filtering, and searching transactions.

## Basic Listing

### Recent Transactions

```sql
SELECT
    t.id,
    t.title,
    t.transaction_date_only,
    t.transaction_type,
    t.destination_amount,
    t.destination_currency,
    sa.name as source_account,
    da.name as destination_account,
    c.name as category
FROM transactions t
LEFT JOIN accounts sa ON sa.id = t.source_account_id
LEFT JOIN accounts da ON da.id = t.destination_account_id
LEFT JOIN categories c ON c.id = t.category_id
WHERE t.deleted_at IS NULL
ORDER BY t.transaction_date_time DESC
LIMIT 50;
```

### Transactions by Type

```sql
-- Expenses only
SELECT * FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
ORDER BY transaction_date_time DESC;

-- Income only
SELECT * FROM transactions
WHERE transaction_type = 2
  AND deleted_at IS NULL;

-- Transfers only
SELECT * FROM transactions
WHERE transaction_type = 1
  AND deleted_at IS NULL;
```

### Transactions by Account

```sql
-- All transactions for an account (as source or destination)
SELECT
    t.*,
    CASE
        WHEN t.source_account_id = :account_id THEN 'Outflow'
        ELSE 'Inflow'
    END as direction
FROM transactions t
WHERE (t.source_account_id = :account_id
    OR t.destination_account_id = :account_id)
  AND t.deleted_at IS NULL
ORDER BY t.transaction_date_time DESC;
```

## Date Filtering

### This Month

```sql
SELECT * FROM transactions
WHERE DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE)
  AND deleted_at IS NULL;
```

### Date Range

```sql
SELECT * FROM transactions
WHERE transaction_date_only BETWEEN :start_date AND :end_date
  AND deleted_at IS NULL
ORDER BY transaction_date_time;
```

### Last N Days

```sql
SELECT * FROM transactions
WHERE transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'
  AND deleted_at IS NULL
ORDER BY transaction_date_time DESC;
```

### This Year

```sql
SELECT * FROM transactions
WHERE EXTRACT(YEAR FROM transaction_date_only) = EXTRACT(YEAR FROM CURRENT_DATE)
  AND deleted_at IS NULL;
```

## Search Queries

### By Title (Case-Insensitive)

```sql
SELECT * FROM transactions
WHERE LOWER(title) LIKE LOWER('%' || :search_term || '%')
  AND deleted_at IS NULL
ORDER BY transaction_date_time DESC;
```

### By Notes

```sql
SELECT * FROM transactions
WHERE notes ILIKE '%' || :search_term || '%'
  AND deleted_at IS NULL;
```

### By Reference Number

```sql
SELECT * FROM transactions
WHERE reference_number = :reference
  AND deleted_at IS NULL;
```

### By Internal Reference

```sql
SELECT * FROM transactions
WHERE :ref = ANY(internal_reference_numbers)
  AND deleted_at IS NULL;
```

## Amount Filtering

### Transactions Over Amount

```sql
SELECT * FROM transactions
WHERE ABS(destination_amount_in_base_currency) > :min_amount
  AND deleted_at IS NULL
ORDER BY ABS(destination_amount_in_base_currency) DESC;
```

### Amount Range

```sql
SELECT * FROM transactions
WHERE destination_amount_in_base_currency BETWEEN :min AND :max
  AND deleted_at IS NULL;
```

### Largest Transactions

```sql
SELECT
    id,
    title,
    transaction_date_only,
    destination_amount_in_base_currency
FROM transactions
WHERE transaction_type = 3  -- Expenses
  AND deleted_at IS NULL
ORDER BY destination_amount_in_base_currency DESC
LIMIT 20;
```

## Category Filtering

### By Specific Category

```sql
SELECT t.*, c.name as category_name
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.category_id = :category_id
  AND t.deleted_at IS NULL;
```

### Uncategorized Transactions

```sql
SELECT * FROM transactions
WHERE category_id IS NULL
  AND transaction_type = 3  -- Expenses
  AND deleted_at IS NULL
ORDER BY transaction_date_time DESC;
```

### Multiple Categories

```sql
SELECT t.*, c.name as category_name
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.category_id IN (:cat1, :cat2, :cat3)
  AND t.deleted_at IS NULL;
```

## Tag Filtering

### By Single Tag

```sql
SELECT * FROM transactions
WHERE :tag_id = ANY(tag_ids)
  AND deleted_at IS NULL;
```

### Any of Multiple Tags (OR)

```sql
SELECT * FROM transactions
WHERE tag_ids && ARRAY[:tag1, :tag2, :tag3]::integer[]
  AND deleted_at IS NULL;
```

### All of Multiple Tags (AND)

```sql
SELECT * FROM transactions
WHERE tag_ids @> ARRAY[:tag1, :tag2]::integer[]
  AND deleted_at IS NULL;
```

### Untagged Transactions

```sql
SELECT * FROM transactions
WHERE (tag_ids IS NULL OR ARRAY_LENGTH(tag_ids, 1) IS NULL)
  AND deleted_at IS NULL;
```

### With Tag Names

```sql
SELECT
    t.id,
    t.title,
    t.transaction_date_only,
    ARRAY_AGG(tg.name ORDER BY tg.name) as tags
FROM transactions t
LEFT JOIN tags tg ON tg.id = ANY(t.tag_ids)
WHERE t.deleted_at IS NULL
  AND (tg.deleted_at IS NULL OR tg.id IS NULL)
GROUP BY t.id, t.title, t.transaction_date_only
ORDER BY t.transaction_date_time DESC
LIMIT 50;
```

## Pagination

### Offset-Based

```sql
SELECT * FROM transactions
WHERE deleted_at IS NULL
ORDER BY transaction_date_time DESC
LIMIT :page_size
OFFSET :page_size * (:page_number - 1);
```

### Cursor-Based (Better for Large Datasets)

```sql
-- First page
SELECT * FROM transactions
WHERE deleted_at IS NULL
ORDER BY transaction_date_time DESC, id DESC
LIMIT :page_size;

-- Next page (using last item's values)
SELECT * FROM transactions
WHERE deleted_at IS NULL
  AND (transaction_date_time, id) < (:last_datetime, :last_id)
ORDER BY transaction_date_time DESC, id DESC
LIMIT :page_size;
```

## Aggregations

### Transaction Count by Type

```sql
SELECT
    CASE transaction_type
        WHEN 1 THEN 'Transfer'
        WHEN 2 THEN 'Income'
        WHEN 3 THEN 'Expense'
        WHEN 5 THEN 'Adjustment'
    END as type,
    COUNT(*) as count
FROM transactions
WHERE deleted_at IS NULL
GROUP BY transaction_type
ORDER BY count DESC;
```

### Average Transaction Amount

```sql
SELECT
    transaction_type,
    COUNT(*) as count,
    AVG(ABS(destination_amount_in_base_currency)) as avg_amount,
    MIN(ABS(destination_amount_in_base_currency)) as min_amount,
    MAX(ABS(destination_amount_in_base_currency)) as max_amount
FROM transactions
WHERE deleted_at IS NULL
GROUP BY transaction_type;
```

## Duplicate Detection

### Potential Duplicates

```sql
SELECT
    title,
    transaction_date_only,
    destination_amount,
    destination_currency,
    COUNT(*) as count
FROM transactions
WHERE deleted_at IS NULL
GROUP BY title, transaction_date_only, destination_amount, destination_currency
HAVING COUNT(*) > 1
ORDER BY count DESC, transaction_date_only DESC;
```

### Same Amount, Same Day

```sql
SELECT
    t1.id as id1,
    t2.id as id2,
    t1.title,
    t1.transaction_date_only,
    t1.destination_amount
FROM transactions t1
JOIN transactions t2 ON t1.id < t2.id
    AND t1.transaction_date_only = t2.transaction_date_only
    AND t1.destination_amount = t2.destination_amount
    AND t1.destination_currency = t2.destination_currency
WHERE t1.deleted_at IS NULL
  AND t2.deleted_at IS NULL;
```

## Combined Filters

### Complex Search Example

```sql
SELECT
    t.id,
    t.title,
    t.transaction_date_only,
    t.destination_amount_in_base_currency,
    c.name as category,
    sa.name as from_account
FROM transactions t
LEFT JOIN categories c ON c.id = t.category_id
LEFT JOIN accounts sa ON sa.id = t.source_account_id
WHERE t.transaction_type = 3  -- Expense
  AND t.transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'
  AND t.destination_amount_in_base_currency > 50
  AND t.category_id = :category_id
  AND t.deleted_at IS NULL
ORDER BY t.destination_amount_in_base_currency DESC
LIMIT 20;
```

## Performance Notes

- Always include `deleted_at IS NULL`
- Use indexes: `ix_transaction_date_only`, `ix_category_id`, `ix_source_account`
- For large result sets, prefer cursor-based pagination
- Filter by date range to limit scan scope
- Use `LIMIT` for exploratory queries
