# Database Indexes

This document lists all indexes in the Go Money database and their optimization purposes.

## transactions Table

| Index | Definition | Best For |
|-------|------------|----------|
| `transactions_pkey` | UNIQUE (id) | Primary key lookups |
| `idx_transactions_active_date` | (transaction_date_time DESC) WHERE deleted_at IS NULL | Recent transactions, date filtering |
| `idx_transactions_active_date_type` | (transaction_date_time DESC, transaction_type) WHERE deleted_at IS NULL | Filter by type and date |
| `ix_source_tx` | (source_account_id, transaction_date_only) INCLUDE (source_amount, destination_amount) | Source account history |
| `ix_dest_tx` | (destination_account_id, transaction_date_only) INCLUDE (source_amount, destination_amount) | Destination account history |
| `ix_source_dest_tx` | (source_account_id, destination_account_id, transaction_date_only) | Transfer queries |
| `idx_transactions_internal_ref_numbers` | GIN (internal_reference_numbers) WHERE deleted_at IS NULL | Reference number search |

### Optimized Query Patterns

```sql
-- Uses idx_transactions_active_date
SELECT * FROM transactions
WHERE deleted_at IS NULL
ORDER BY transaction_date_time DESC
LIMIT 50;

-- Uses idx_transactions_active_date_type
SELECT * FROM transactions
WHERE deleted_at IS NULL
  AND transaction_type = 3
ORDER BY transaction_date_time DESC;

-- Uses ix_source_tx
SELECT * FROM transactions
WHERE source_account_id = 1
  AND transaction_date_only >= '2024-01-01'
ORDER BY transaction_date_only;

-- Uses idx_transactions_internal_ref_numbers
SELECT * FROM transactions
WHERE 'REF123' = ANY(internal_reference_numbers)
  AND deleted_at IS NULL;
```

## double_entries Table

| Index | Definition | Best For |
|-------|------------|----------|
| `double_entries_pk` | UNIQUE (id) | Primary key |
| `ix_transaction` | (transaction_id) | Find entries by transaction |
| `ix_uniq_record` | UNIQUE (transaction_id, is_debit) WHERE deleted_at IS NULL | Data integrity |
| `ix_double_entries_transaction_date` | (account_id, transaction_date) WHERE deleted_at IS NULL | Account ledger queries |

### Optimized Query Patterns

```sql
-- Uses ix_double_entries_transaction_date
SELECT * FROM double_entries
WHERE account_id = 1
  AND transaction_date >= '2024-01-01'
  AND deleted_at IS NULL
ORDER BY transaction_date;

-- Uses ix_transaction
SELECT * FROM double_entries
WHERE transaction_id = 12345;
```

## daily_stat Table

| Index | Definition | Best For |
|-------|------------|----------|
| `daily_stat_pk` | UNIQUE (account_id, date) | Primary key, exact lookups |
| `daily_stat_account_id_index` | (account_id) | Account filtering |
| `ix_latest_stat` | (account_id, date DESC) | Most recent stats |

### Optimized Query Patterns

```sql
-- Uses daily_stat_pk
SELECT * FROM daily_stat
WHERE account_id = 1 AND date = '2024-01-15';

-- Uses ix_latest_stat
SELECT * FROM daily_stat
WHERE account_id = 1
ORDER BY date DESC
LIMIT 30;
```

## categories Table

| Index | Definition | Best For |
|-------|------------|----------|
| `categories_pk` | UNIQUE (id) | Primary key |
| `categories__uindex` | UNIQUE (name) WHERE deleted_at IS NULL | Name lookups, duplicates prevention |

## tags Table

| Index | Definition | Best For |
|-------|------------|----------|
| `tags_pk` | UNIQUE (id) | Primary key |
| `ix_tag_name` | UNIQUE (name) WHERE deleted_at IS NULL | Name lookups |

## users Table

| Index | Definition | Best For |
|-------|------------|----------|
| `users_pk` | UNIQUE (id) | Primary key |
| `users_login_uindex` | UNIQUE (login) WHERE deleted_at IS NULL | Login lookups |

## accounts Table

| Index | Definition | Best For |
|-------|------------|----------|
| `accounts_pk` | UNIQUE (id) | Primary key |

## currencies Table

| Index | Definition | Best For |
|-------|------------|----------|
| `currencies_pk` | UNIQUE (id) | Primary key (currency code) |

## import_deduplication Table

| Index | Definition | Best For |
|-------|------------|----------|
| `import_deduplication_pk` | UNIQUE (import_source, key) | Duplicate detection |

## Query Optimization Tips

### 1. Use Partial Indexes

Most indexes include `WHERE deleted_at IS NULL` to only index active records:
```sql
-- This query uses the partial index efficiently
SELECT * FROM transactions WHERE deleted_at IS NULL ...

-- This query cannot use partial index
SELECT * FROM transactions WHERE deleted_at IS NOT NULL ...
```

### 2. Covering Indexes

The transaction indexes include amounts in INCLUDE clause for covering queries:
```sql
-- Uses covering index, no table lookup needed
SELECT source_amount, destination_amount
FROM transactions
WHERE source_account_id = 1
  AND transaction_date_only >= '2024-01-01';
```

### 3. Date Range Queries

Always use `transaction_date_only` for date filtering (indexed):
```sql
-- Good: uses index
WHERE transaction_date_only BETWEEN '2024-01-01' AND '2024-01-31'

-- Avoid: may not use index efficiently
WHERE DATE(transaction_date_time) = '2024-01-15'
```

### 4. Use daily_stat for Aggregates

For balance calculations, prefer `daily_stat` over scanning transactions:
```sql
-- Fast: pre-computed
SELECT SUM(amount) FROM daily_stat WHERE account_id = 1;

-- Slower: scans transactions
SELECT SUM(destination_amount) FROM transactions WHERE destination_account_id = 1;
```

### 5. Tag Array Queries

The `tag_ids` array doesn't have a dedicated index. For heavy tag filtering:
```sql
-- Consider using EXISTS for better performance
SELECT * FROM transactions t
WHERE EXISTS (SELECT 1 FROM tags WHERE id = ANY(t.tag_ids) AND name = 'vacation')
  AND t.deleted_at IS NULL;
```
