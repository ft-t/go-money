# import_deduplication Table

The `import_deduplication` table prevents duplicate transaction imports.

## Schema

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| import_source | integer | NO | - | Source system identifier (composite PK) |
| key | text | NO | - | Unique transaction identifier from source (composite PK) |
| transaction_id | bigint | NO | - | FK to imported transaction |
| created_at | timestamp | NO | - | Import timestamp |

## Primary Key

Composite primary key: `(import_source, key)`

## Indexes

| Index | Definition | Purpose |
|-------|------------|---------|
| import_deduplication_pk | UNIQUE (import_source, key) | Primary key / dedup check |

## Import Sources

| Value | Name | Description |
|-------|------|-------------|
| 0 | UNSPECIFIED | Not used |
| 1 | FIREFLY | Firefly III finance manager |

## Common Queries

### Check if Already Imported

```sql
SELECT EXISTS(
    SELECT 1 FROM import_deduplication
    WHERE import_source = :source
      AND key = :unique_key
) as already_imported;
```

### Get Imported Transaction

```sql
SELECT t.*
FROM import_deduplication id
JOIN transactions t ON t.id = id.transaction_id
WHERE id.import_source = :source
  AND id.key = :unique_key;
```

### Import Statistics

```sql
SELECT
    import_source,
    COUNT(*) as imported_count,
    MIN(created_at) as first_import,
    MAX(created_at) as last_import
FROM import_deduplication
GROUP BY import_source;
```

### Recent Imports

```sql
SELECT
    id.key,
    id.created_at,
    t.title,
    t.transaction_date_only
FROM import_deduplication id
JOIN transactions t ON t.id = id.transaction_id
WHERE id.import_source = :source
ORDER BY id.created_at DESC
LIMIT 50;
```

## Notes

- Key should be a unique identifier from the source system
- Prevents re-importing the same transaction multiple times
- Useful for incremental imports
- Transaction relationship helps trace imported data
