# tags Table

The `tags` table stores tags that can be applied to transactions for flexible categorization.

## Schema

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | integer | NO | auto-increment | Primary key |
| name | text | NO | - | Tag name |
| color | text | YES | - | Hex color code for UI display |
| icon | text | YES | - | Icon identifier for UI display |
| created_at | timestamp | NO | - | Record creation time |
| deleted_at | timestamp | YES | - | Soft delete timestamp |

## Primary Key

- `id` (integer, auto-increment)

## Indexes

| Index | Definition | Purpose |
|-------|------------|---------|
| tags_pk | UNIQUE (id) | Primary key |
| ix_tag_name | UNIQUE (name) WHERE deleted_at IS NULL | Unique names for active tags |

## Relationship to Transactions

Tags have a many-to-many relationship with transactions via the `transactions.tag_ids` array field.

## Common Queries

### All Active Tags

```sql
SELECT id, name, color, icon
FROM tags
WHERE deleted_at IS NULL
ORDER BY name;
```

### Tags with Transaction Count

```sql
SELECT
    t.id,
    t.name,
    t.color,
    COUNT(tx.id) as usage_count
FROM tags t
LEFT JOIN transactions tx ON t.id = ANY(tx.tag_ids) AND tx.deleted_at IS NULL
WHERE t.deleted_at IS NULL
GROUP BY t.id, t.name, t.color
ORDER BY usage_count DESC;
```

### Transactions with a Specific Tag

```sql
SELECT
    tx.id,
    tx.title,
    tx.transaction_date_only,
    tx.destination_amount,
    tx.destination_currency
FROM transactions tx
WHERE :tag_id = ANY(tx.tag_ids)
  AND tx.deleted_at IS NULL
ORDER BY tx.transaction_date_time DESC;
```

### Transactions with Multiple Tags (ALL)

```sql
SELECT * FROM transactions
WHERE tag_ids @> ARRAY[1, 2, 3]::integer[]  -- Has ALL these tags
  AND deleted_at IS NULL;
```

### Transactions with Any Tags (OR)

```sql
SELECT * FROM transactions
WHERE tag_ids && ARRAY[1, 2, 3]::integer[]  -- Has ANY of these tags
  AND deleted_at IS NULL;
```

### Tag Spending Analysis

```sql
SELECT
    t.name as tag,
    COUNT(tx.id) as transaction_count,
    SUM(tx.destination_amount_in_base_currency) as total_spent
FROM tags t
JOIN transactions tx ON t.id = ANY(tx.tag_ids)
    AND tx.deleted_at IS NULL
    AND tx.transaction_type = 3  -- Expense
WHERE t.deleted_at IS NULL
GROUP BY t.id, t.name
ORDER BY total_spent DESC;
```

### Expand Tag IDs to Names

```sql
SELECT
    tx.id,
    tx.title,
    ARRAY_AGG(t.name ORDER BY t.name) as tag_names
FROM transactions tx
LEFT JOIN tags t ON t.id = ANY(tx.tag_ids)
WHERE tx.deleted_at IS NULL
  AND (t.deleted_at IS NULL OR t.id IS NULL)
GROUP BY tx.id, tx.title;
```

### Untagged Transactions

```sql
SELECT COUNT(*) as untagged_count
FROM transactions
WHERE (tag_ids IS NULL OR ARRAY_LENGTH(tag_ids, 1) IS NULL)
  AND transaction_type = 3  -- Expense
  AND deleted_at IS NULL;
```

### Tag Combinations

```sql
SELECT
    tx.tag_ids,
    COUNT(*) as count
FROM transactions tx
WHERE tx.tag_ids IS NOT NULL
  AND ARRAY_LENGTH(tx.tag_ids, 1) > 1
  AND tx.deleted_at IS NULL
GROUP BY tx.tag_ids
ORDER BY count DESC
LIMIT 10;
```

## Notes

- Tags are stored as an array in `transactions.tag_ids` for flexibility
- Use PostgreSQL array operators (`@>`, `&&`, `= ANY()`) for tag queries
- Color can be stored as hex code (e.g., "#FF5733")
- Icon can reference icon library identifiers
- Multiple tags can be applied to a single transaction
