# categories Table

The `categories` table stores transaction categories for organizing expenses.

## Schema

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | integer | NO | auto-increment | Primary key |
| name | text | NO | - | Category name |
| created_at | timestamp | NO | - | Record creation time |
| deleted_at | timestamp | YES | - | Soft delete timestamp |

## Primary Key

- `id` (integer, auto-increment)

## Indexes

| Index | Definition | Purpose |
|-------|------------|---------|
| categories_pk | UNIQUE (id) | Primary key |
| categories__uindex | UNIQUE (name) WHERE deleted_at IS NULL | Unique names for active categories |

## Common Queries

### All Active Categories

```sql
SELECT id, name, created_at
FROM categories
WHERE deleted_at IS NULL
ORDER BY name;
```

### Category with Transaction Count

```sql
SELECT
    c.id,
    c.name,
    COUNT(t.id) as transaction_count
FROM categories c
LEFT JOIN transactions t ON t.category_id = c.id AND t.deleted_at IS NULL
WHERE c.deleted_at IS NULL
GROUP BY c.id, c.name
ORDER BY transaction_count DESC;
```

### Category Spending Summary

```sql
SELECT
    c.id,
    c.name,
    COUNT(t.id) as transaction_count,
    SUM(t.destination_amount_in_base_currency) as total_spent
FROM categories c
LEFT JOIN transactions t ON t.category_id = c.id
    AND t.deleted_at IS NULL
    AND t.transaction_type = 3  -- Expense
WHERE c.deleted_at IS NULL
GROUP BY c.id, c.name
ORDER BY total_spent DESC NULLS LAST;
```

### Monthly Spending by Category

```sql
SELECT
    c.name as category,
    DATE_TRUNC('month', t.transaction_date_only) as month,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3  -- Expense
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= DATE_TRUNC('year', CURRENT_DATE)
GROUP BY c.name, DATE_TRUNC('month', t.transaction_date_only)
ORDER BY month, total DESC;
```

### Uncategorized Transactions

```sql
SELECT COUNT(*) as uncategorized_count
FROM transactions
WHERE category_id IS NULL
  AND transaction_type = 3  -- Expense
  AND deleted_at IS NULL;
```

### Search Categories

```sql
SELECT * FROM categories
WHERE deleted_at IS NULL
  AND name ILIKE '%' || :search || '%';
```

## Notes

- Categories are optional - transactions can have `category_id = NULL`
- Unique constraint on name prevents duplicate category names
- Soft delete allows historical data preservation
- Use with expense transactions (type=3) for spending analysis
