# double_entries Table

The `double_entries` table implements double-entry bookkeeping, creating debit and credit entries for each transaction.

## Schema

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | bigint | NO | auto-increment | Primary key |
| transaction_id | bigint | NO | - | FK to transactions.id |
| account_id | integer | NO | - | FK to accounts.id |
| is_debit | boolean | NO | - | True=debit, False=credit |
| amount_in_base_currency | numeric | YES | - | Amount in base currency (always positive) |
| base_currency | text | YES | - | Base currency code |
| transaction_date | timestamp | YES | - | Transaction date for sorting |
| created_at | timestamp | NO | - | Record creation time |
| deleted_at | timestamp | YES | - | Soft delete timestamp |

## Primary Key

- `id` (bigint, auto-increment)

## Indexes

| Index | Definition | Purpose |
|-------|------------|---------|
| double_entries_pk | UNIQUE (id) | Primary key |
| ix_transaction | (transaction_id) | Find entries by transaction |
| ix_uniq_record | UNIQUE (transaction_id, is_debit) WHERE deleted_at IS NULL | One debit, one credit per transaction |
| ix_double_entries_transaction_date | (account_id, transaction_date) WHERE deleted_at IS NULL | Account ledger queries |

## Double-Entry Bookkeeping Concept

Every transaction creates exactly two entries:
1. **Debit entry** (`is_debit = true`) - increases asset/expense or decreases liability/income
2. **Credit entry** (`is_debit = false`) - decreases asset/expense or increases liability/income

The fundamental rule: **Total Debits = Total Credits**

## Entry Creation by Transaction Type

| Transaction Type | Debit Account | Credit Account |
|------------------|---------------|----------------|
| Expense (3) | Expense account | Asset/Liability account |
| Income (2) | Asset/Liability account | Income account |
| Transfer (1) | Destination account | Source account |
| Adjustment (5) | Target account or Adjustment | Adjustment or Target |

## Common Queries

### Account Ledger

```sql
SELECT
    de.id,
    de.transaction_date,
    t.title,
    CASE WHEN de.is_debit THEN de.amount_in_base_currency ELSE NULL END as debit,
    CASE WHEN NOT de.is_debit THEN de.amount_in_base_currency ELSE NULL END as credit,
    de.base_currency
FROM double_entries de
JOIN transactions t ON t.id = de.transaction_id
WHERE de.account_id = :account_id
  AND de.deleted_at IS NULL
ORDER BY de.transaction_date DESC, de.id DESC;
```

### Account Balance from Ledger

```sql
SELECT
    SUM(CASE WHEN is_debit THEN amount_in_base_currency ELSE -amount_in_base_currency END) as balance
FROM double_entries
WHERE account_id = :account_id
  AND deleted_at IS NULL;
```

### Trial Balance

```sql
SELECT
    a.name as account_name,
    a.type as account_type,
    SUM(CASE WHEN de.is_debit THEN de.amount_in_base_currency ELSE 0 END) as total_debit,
    SUM(CASE WHEN NOT de.is_debit THEN de.amount_in_base_currency ELSE 0 END) as total_credit,
    SUM(CASE WHEN de.is_debit THEN de.amount_in_base_currency ELSE -de.amount_in_base_currency END) as balance
FROM double_entries de
JOIN accounts a ON a.id = de.account_id
WHERE de.deleted_at IS NULL
  AND a.deleted_at IS NULL
GROUP BY a.id, a.name, a.type
ORDER BY a.type, a.name;
```

### Verify Debits Equal Credits

```sql
SELECT
    SUM(CASE WHEN is_debit THEN amount_in_base_currency ELSE 0 END) as total_debits,
    SUM(CASE WHEN NOT is_debit THEN amount_in_base_currency ELSE 0 END) as total_credits,
    SUM(CASE WHEN is_debit THEN amount_in_base_currency ELSE -amount_in_base_currency END) as difference
FROM double_entries
WHERE deleted_at IS NULL;
-- difference should always be 0
```

### Entries for a Transaction

```sql
SELECT
    de.id,
    a.name as account_name,
    de.is_debit,
    de.amount_in_base_currency,
    de.base_currency
FROM double_entries de
JOIN accounts a ON a.id = de.account_id
WHERE de.transaction_id = :transaction_id
  AND de.deleted_at IS NULL;
```

### Daily Account Activity

```sql
SELECT
    DATE(transaction_date) as date,
    SUM(CASE WHEN is_debit THEN amount_in_base_currency ELSE 0 END) as debits,
    SUM(CASE WHEN NOT is_debit THEN amount_in_base_currency ELSE 0 END) as credits
FROM double_entries
WHERE account_id = :account_id
  AND deleted_at IS NULL
  AND transaction_date >= CURRENT_DATE - INTERVAL '30 days'
GROUP BY DATE(transaction_date)
ORDER BY date;
```

### Running Balance

```sql
SELECT
    de.transaction_date,
    t.title,
    de.is_debit,
    de.amount_in_base_currency,
    SUM(CASE WHEN de2.is_debit THEN de2.amount_in_base_currency
             ELSE -de2.amount_in_base_currency END) as running_balance
FROM double_entries de
JOIN transactions t ON t.id = de.transaction_id
JOIN double_entries de2 ON de2.account_id = de.account_id
    AND de2.deleted_at IS NULL
    AND (de2.transaction_date < de.transaction_date
         OR (de2.transaction_date = de.transaction_date AND de2.id <= de.id))
WHERE de.account_id = :account_id
  AND de.deleted_at IS NULL
GROUP BY de.id, de.transaction_date, t.title, de.is_debit, de.amount_in_base_currency
ORDER BY de.transaction_date, de.id;
```

## Notes

- `amount_in_base_currency` is always positive; direction determined by `is_debit`
- Each transaction should have exactly one debit and one credit entry
- Use this table for formal accounting reports and reconciliation
- The unique index ensures data integrity (no duplicate entries)
- Soft deletes preserve audit trail


---

## See Also

- [Double-Entry Overview](../../business-logic/double-entry/overview.md) - Bookkeeping concepts
- [Schema Quick-Ref](../QUICK-REF.md) - All tables at a glance
