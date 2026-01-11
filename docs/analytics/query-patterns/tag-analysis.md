# Tag Analysis

Queries for analyzing transactions using PostgreSQL array operators for tags.

## PostgreSQL Array Operators

| Operator | Meaning | Example |
|----------|---------|---------|
| `= ANY(array)` | Element in array | `5 = ANY(tag_ids)` |
| `&&` | Arrays overlap (OR) | `tag_ids && ARRAY[1,2]` |
| `@>` | Contains all (AND) | `tag_ids @> ARRAY[1,2]` |
| `<@` | Contained by | `ARRAY[1] <@ tag_ids` |

## Basic Tag Queries

### All Tags with Usage

```sql
SELECT
    t.id,
    t.name,
    t.color,
    COUNT(tx.id) as usage_count,
    SUM(tx.destination_amount_in_base_currency) as total_amount
FROM tags t
LEFT JOIN transactions tx ON t.id = ANY(tx.tag_ids)
    AND tx.deleted_at IS NULL
    AND tx.transaction_type = 3
WHERE t.deleted_at IS NULL
GROUP BY t.id, t.name, t.color
ORDER BY usage_count DESC;
```

### Top Tags by Spending

```sql
SELECT
    tg.name as tag,
    COUNT(*) as count,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN tags tg ON tg.id = ANY(t.tag_ids)
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND tg.deleted_at IS NULL
GROUP BY tg.name
ORDER BY total DESC
LIMIT 10;
```

### Unused Tags

```sql
SELECT t.id, t.name
FROM tags t
WHERE t.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1 FROM transactions tx
      WHERE t.id = ANY(tx.tag_ids)
        AND tx.deleted_at IS NULL
  );
```

## Finding Tagged Transactions

### Single Tag

```sql
SELECT
    t.id,
    t.title,
    t.transaction_date_only,
    t.destination_amount_in_base_currency
FROM transactions t
WHERE :tag_id = ANY(t.tag_ids)
  AND t.deleted_at IS NULL
ORDER BY t.transaction_date_time DESC;
```

### Any of Multiple Tags (OR)

```sql
SELECT * FROM transactions
WHERE tag_ids && ARRAY[1, 2, 3]::integer[]
  AND deleted_at IS NULL;
```

### All of Multiple Tags (AND)

```sql
SELECT * FROM transactions
WHERE tag_ids @> ARRAY[1, 2]::integer[]
  AND deleted_at IS NULL;
```

### Transactions with Specific Tag Count

```sql
-- Exactly 2 tags
SELECT * FROM transactions
WHERE ARRAY_LENGTH(tag_ids, 1) = 2
  AND deleted_at IS NULL;

-- More than 3 tags
SELECT * FROM transactions
WHERE ARRAY_LENGTH(tag_ids, 1) > 3
  AND deleted_at IS NULL;
```

## Untagged Analysis

### Untagged Transactions

```sql
SELECT
    id,
    title,
    transaction_date_only,
    destination_amount_in_base_currency
FROM transactions
WHERE (tag_ids IS NULL OR ARRAY_LENGTH(tag_ids, 1) IS NULL)
  AND transaction_type = 3
  AND deleted_at IS NULL
ORDER BY destination_amount_in_base_currency DESC;
```

### Untagged Count and Total

```sql
SELECT
    COUNT(*) as untagged_count,
    SUM(destination_amount_in_base_currency) as untagged_total
FROM transactions
WHERE (tag_ids IS NULL OR ARRAY_LENGTH(tag_ids, 1) IS NULL)
  AND transaction_type = 3
  AND deleted_at IS NULL;
```

### Tagged vs Untagged Comparison

```sql
SELECT
    CASE
        WHEN tag_ids IS NULL OR ARRAY_LENGTH(tag_ids, 1) IS NULL
        THEN 'Untagged'
        ELSE 'Tagged'
    END as status,
    COUNT(*) as count,
    SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
GROUP BY status;
```

## Tag Combinations

### Most Common Tag Pairs

```sql
SELECT
    t1.name as tag1,
    t2.name as tag2,
    COUNT(*) as count
FROM transactions tx
JOIN tags t1 ON t1.id = tx.tag_ids[1]
JOIN tags t2 ON t2.id = tx.tag_ids[2]
WHERE ARRAY_LENGTH(tx.tag_ids, 1) >= 2
  AND tx.deleted_at IS NULL
  AND t1.id < t2.id  -- Avoid duplicates
GROUP BY t1.name, t2.name
ORDER BY count DESC
LIMIT 10;
```

### Tag Combination Analysis

```sql
SELECT
    tx.tag_ids,
    ARRAY_AGG(DISTINCT tg.name ORDER BY tg.name) as tag_names,
    COUNT(*) as count,
    SUM(tx.destination_amount_in_base_currency) as total
FROM transactions tx
JOIN tags tg ON tg.id = ANY(tx.tag_ids)
WHERE tx.deleted_at IS NULL
  AND tx.transaction_type = 3
  AND ARRAY_LENGTH(tx.tag_ids, 1) > 1
GROUP BY tx.tag_ids
ORDER BY count DESC
LIMIT 20;
```

## Tag Trends Over Time

### Monthly Tag Usage

```sql
SELECT
    DATE_TRUNC('month', t.transaction_date_only) as month,
    tg.name as tag,
    COUNT(*) as count,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN tags tg ON tg.id = ANY(t.tag_ids)
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= CURRENT_DATE - INTERVAL '12 months'
GROUP BY month, tg.name
ORDER BY month DESC, total DESC;
```

### Tag Trend (Specific Tag)

```sql
SELECT
    DATE_TRUNC('month', t.transaction_date_only) as month,
    COUNT(*) as count,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
WHERE :tag_id = ANY(t.tag_ids)
  AND t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= CURRENT_DATE - INTERVAL '12 months'
GROUP BY month
ORDER BY month;
```

## Tag + Category Analysis

### Tags by Category

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
GROUP BY c.name, tg.name
ORDER BY c.name, total DESC;
```

### Most Used Tags per Category

```sql
WITH ranked AS (
    SELECT
        c.name as category,
        tg.name as tag,
        COUNT(*) as count,
        ROW_NUMBER() OVER (PARTITION BY c.name ORDER BY COUNT(*) DESC) as rn
    FROM transactions t
    JOIN categories c ON c.id = t.category_id
    JOIN tags tg ON tg.id = ANY(t.tag_ids)
    WHERE t.transaction_type = 3
      AND t.deleted_at IS NULL
    GROUP BY c.name, tg.name
)
SELECT category, tag, count
FROM ranked
WHERE rn <= 3
ORDER BY category, count DESC;
```

## Expanding Tag IDs to Names

### Transaction List with Tag Names

```sql
SELECT
    t.id,
    t.title,
    t.transaction_date_only,
    t.destination_amount_in_base_currency,
    ARRAY_AGG(tg.name ORDER BY tg.name) FILTER (WHERE tg.name IS NOT NULL) as tags
FROM transactions t
LEFT JOIN tags tg ON tg.id = ANY(t.tag_ids) AND tg.deleted_at IS NULL
WHERE t.deleted_at IS NULL
GROUP BY t.id, t.title, t.transaction_date_only, t.destination_amount_in_base_currency
ORDER BY t.transaction_date_time DESC
LIMIT 50;
```

### Tag String Concatenation

```sql
SELECT
    t.id,
    t.title,
    STRING_AGG(tg.name, ', ' ORDER BY tg.name) as tags
FROM transactions t
LEFT JOIN tags tg ON tg.id = ANY(t.tag_ids)
WHERE t.deleted_at IS NULL
GROUP BY t.id, t.title
ORDER BY t.id DESC
LIMIT 50;
```

## Tag Statistics

### Tag Usage Statistics

```sql
SELECT
    tg.name as tag,
    COUNT(*) as usage_count,
    SUM(t.destination_amount_in_base_currency) as total_amount,
    AVG(t.destination_amount_in_base_currency) as avg_amount,
    MIN(t.transaction_date_only) as first_used,
    MAX(t.transaction_date_only) as last_used
FROM transactions t
JOIN tags tg ON tg.id = ANY(t.tag_ids)
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND tg.deleted_at IS NULL
GROUP BY tg.name
ORDER BY usage_count DESC;
```

### Tags per Transaction Distribution

```sql
SELECT
    COALESCE(ARRAY_LENGTH(tag_ids, 1), 0) as tag_count,
    COUNT(*) as transaction_count,
    SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
GROUP BY COALESCE(ARRAY_LENGTH(tag_ids, 1), 0)
ORDER BY tag_count;
```

## Performance Notes

- PostgreSQL array operators are efficient with GIN index
- Consider creating index: `CREATE INDEX ix_tag_ids ON transactions USING GIN(tag_ids)`
- For large datasets, filter by date first, then by tags
- `ANY()` is faster than `@>` for single element checks
