# transactions Table

The `transactions` table is the core table storing all financial transactions in the system.

## Schema

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | bigint | NO | auto-increment | Primary key |
| source_amount | numeric | YES | - | Amount leaving source account |
| source_currency | text | NO | - | Currency code of source amount |
| source_amount_in_base_currency | numeric | YES | - | Source amount converted to base currency |
| destination_amount | numeric | YES | - | Amount entering destination account |
| destination_currency | text | NO | - | Currency code of destination amount |
| destination_amount_in_base_currency | numeric | YES | - | Destination amount in base currency |
| fx_source_amount | numeric | YES | - | Original foreign currency amount (expenses only) |
| fx_source_currency | text | YES | - | Original foreign currency code (expenses only) |
| source_account_id | integer | YES | - | FK to accounts.id |
| destination_account_id | integer | YES | - | FK to accounts.id |
| category_id | integer | YES | - | FK to categories.id |
| tag_ids | integer[] | YES | - | Array of tag IDs |
| transaction_type | integer | NO | - | See TransactionType enum |
| flags | bigint | NO | - | Bitmask for transaction flags |
| title | text | YES | - | Transaction description/title |
| notes | text | YES | - | Additional notes |
| reference_number | text | YES | - | External reference number |
| internal_reference_numbers | text[] | YES | - | Array of internal reference numbers |
| extra | jsonb | NO | '{}' | Additional metadata |
| transaction_date_time | timestamp | NO | - | Full transaction timestamp |
| transaction_date_only | date | NO | - | Date portion only (for grouping) |
| voided_by_transaction_id | bigint | YES | - | ID of reversal transaction |
| created_at | timestamp | NO | now() | Record creation time |
| updated_at | timestamp | NO | now() | Record update time |
| deleted_at | timestamp | YES | - | Soft delete timestamp |

## Primary Key

- `id` (bigint, auto-increment)

## Indexes

| Index | Definition | Purpose |
|-------|------------|---------|
| transactions_pkey | UNIQUE (id) | Primary key |
| idx_transactions_active_date | (transaction_date_time DESC) WHERE deleted_at IS NULL | Active transactions by date |
| idx_transactions_active_date_type | (transaction_date_time DESC, transaction_type) WHERE deleted_at IS NULL | Filter by date and type |
| ix_source_tx | (source_account_id, transaction_date_only) INCLUDE (amounts) | Source account queries |
| ix_dest_tx | (destination_account_id, transaction_date_only) INCLUDE (amounts) | Destination account queries |
| ix_source_dest_tx | (source_account_id, destination_account_id, transaction_date_only) | Both accounts |
| idx_transactions_internal_ref_numbers | GIN (internal_reference_numbers) WHERE deleted_at IS NULL | Reference number search |

## Amount Fields Explained

### For Expense Transactions (type=3)
- `source_amount` / `source_currency`: Amount from your account
- `destination_amount` / `destination_currency`: Amount to expense category (usually same)
- `fx_source_amount` / `fx_source_currency`: Original foreign currency if paid abroad
- Base currency fields: Amounts converted for reporting

### For Income Transactions (type=2)
- `source_amount` / `source_currency`: Amount from income source
- `destination_amount` / `destination_currency`: Amount deposited to your account

### For Transfer Transactions (type=1)
- `source_amount` / `source_currency`: Amount leaving source account
- `destination_amount` / `destination_currency`: Amount entering destination account
- Different currencies = currency exchange

## Transaction Type Behavior

| Type | Source Account | Destination Account |
|------|----------------|---------------------|
| 1 (Transfer) | Asset or Liability | Asset or Liability |
| 2 (Income) | Income account | Asset or Liability |
| 3 (Expense) | Asset or Liability | Expense account |
| 5 (Adjustment) | Adjustment account | Target account |

## Common Queries

### List Recent Transactions

```sql
SELECT
    t.id,
    t.title,
    t.transaction_date_only,
    t.transaction_type,
    t.source_amount,
    t.source_currency,
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

### Transactions by Account

```sql
SELECT * FROM transactions
WHERE deleted_at IS NULL
  AND (source_account_id = :account_id OR destination_account_id = :account_id)
ORDER BY transaction_date_time DESC;
```

### Transactions by Category

```sql
SELECT
    t.title,
    t.transaction_date_only,
    t.destination_amount,
    t.destination_currency
FROM transactions t
WHERE t.category_id = :category_id
  AND t.deleted_at IS NULL
ORDER BY t.transaction_date_time DESC;
```

### Transactions with Tag

```sql
SELECT * FROM transactions
WHERE :tag_id = ANY(tag_ids)
  AND deleted_at IS NULL
ORDER BY transaction_date_time DESC;
```

### Transactions with Multiple Tags (AND)

```sql
SELECT * FROM transactions
WHERE tag_ids @> ARRAY[:tag1, :tag2]::integer[]
  AND deleted_at IS NULL;
```

### Transactions with Any Tag (OR)

```sql
SELECT * FROM transactions
WHERE tag_ids && ARRAY[:tag1, :tag2]::integer[]
  AND deleted_at IS NULL;
```

### Monthly Spending by Category

```sql
SELECT
    c.name as category,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3  -- Expense
  AND t.deleted_at IS NULL
  AND DATE_TRUNC('month', t.transaction_date_only) = '2024-01-01'
GROUP BY c.name
ORDER BY total DESC;
```

### Find Duplicate Transactions

```sql
SELECT
    title,
    transaction_date_only,
    source_amount,
    COUNT(*) as count
FROM transactions
WHERE deleted_at IS NULL
GROUP BY title, transaction_date_only, source_amount
HAVING COUNT(*) > 1;
```

### Search by Reference Number

```sql
SELECT * FROM transactions
WHERE :ref = ANY(internal_reference_numbers)
  AND deleted_at IS NULL;
```

## Notes

- Always filter with `deleted_at IS NULL` for active transactions
- Use `transaction_date_only` for date grouping to avoid timezone issues
- The `tag_ids` array requires PostgreSQL array operators for queries
- Base currency amounts are computed at transaction time and cached


---

## See Also

- [Transaction Types](../../business-logic/transactions/types.md) - Type-specific behavior
- [Transaction Overview](../../business-logic/transactions/overview.md) - Processing pipeline
- [Amount Calculations](../../business-logic/transactions/amount-calculations.md) - Currency conversion
- [Transaction Queries](../../analytics/query-patterns/transaction-queries.md) - Query patterns
- [Schema Quick-Ref](../QUICK-REF.md) - All tables at a glance
