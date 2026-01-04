# accounts Table

The `accounts` table stores all financial accounts including bank accounts, credit cards, expense categories, and income sources.

## Schema

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | integer | NO | auto-increment | Primary key |
| name | text | YES | - | Account display name |
| current_balance | numeric | NO | - | Current account balance |
| currency | text | NO | - | Currency code (ISO 4217) |
| type | integer | YES | - | Account type enum |
| flags | bigint | NO | - | Bitmask flags |
| extra | jsonb | NO | '{}' | Additional metadata |
| note | text | NO | - | Account notes |
| account_number | text | NO | - | Bank account number |
| iban | text | NO | - | IBAN for bank accounts |
| liability_percent | numeric | YES | - | Credit utilization tracking |
| display_order | integer | YES | - | UI sort order |
| first_transaction_at | timestamp | YES | - | Date of first transaction |
| last_updated_at | timestamp | NO | - | Balance update timestamp |
| created_at | timestamp | NO | - | Record creation time |
| deleted_at | timestamp | YES | - | Soft delete timestamp |

## Primary Key

- `id` (integer, auto-increment)

## Indexes

| Index | Definition | Purpose |
|-------|------------|---------|
| accounts_pk | UNIQUE (id) | Primary key |

## Account Types

| Value | Name | Description | Examples |
|-------|------|-------------|----------|
| 1 | Asset | Accounts you own | Checking, Savings, Cash, Investments |
| 4 | Liability | Accounts you owe | Credit Cards, Loans, Mortgages |
| 5 | Expense | Spending categories | Groceries, Utilities, Entertainment |
| 6 | Income | Revenue sources | Salary, Dividends, Freelance |
| 7 | Adjustment | Balance corrections | Opening Balance, Adjustments |

## Flags Bitmask

| Bit | Value | Name | Description |
|-----|-------|------|-------------|
| 0 | 1 | IsDefault | Default account for its type |

## Balance Interpretation

- **Asset accounts**: Positive = money you have
- **Liability accounts**: Positive = money you owe
- **Expense accounts**: Total spent in this category
- **Income accounts**: Total received from this source

## Common Queries

### All Active Accounts

```sql
SELECT * FROM accounts
WHERE deleted_at IS NULL
ORDER BY type, display_order, name;
```

### Asset and Liability Accounts

```sql
SELECT
    id,
    name,
    current_balance,
    currency,
    type
FROM accounts
WHERE type IN (1, 4)  -- Asset, Liability
  AND deleted_at IS NULL
ORDER BY type, name;
```

### Net Worth Calculation

```sql
SELECT
    SUM(CASE WHEN type = 1 THEN current_balance ELSE 0 END) as total_assets,
    SUM(CASE WHEN type = 4 THEN current_balance ELSE 0 END) as total_liabilities,
    SUM(CASE
        WHEN type = 1 THEN current_balance
        WHEN type = 4 THEN -current_balance
        ELSE 0
    END) as net_worth
FROM accounts
WHERE type IN (1, 4)
  AND deleted_at IS NULL;
```

### Net Worth by Currency

```sql
SELECT
    currency,
    SUM(CASE WHEN type = 1 THEN current_balance ELSE 0 END) as assets,
    SUM(CASE WHEN type = 4 THEN current_balance ELSE 0 END) as liabilities,
    SUM(CASE
        WHEN type = 1 THEN current_balance
        WHEN type = 4 THEN -current_balance
        ELSE 0
    END) as net_worth
FROM accounts
WHERE type IN (1, 4)
  AND deleted_at IS NULL
GROUP BY currency;
```

### Default Accounts

```sql
SELECT * FROM accounts
WHERE (flags & 1) = 1  -- IsDefault flag set
  AND deleted_at IS NULL;
```

### Expense Categories

```sql
SELECT
    id,
    name,
    current_balance as total_spent
FROM accounts
WHERE type = 5  -- Expense
  AND deleted_at IS NULL
ORDER BY current_balance DESC;
```

### Income Sources

```sql
SELECT
    id,
    name,
    current_balance as total_received
FROM accounts
WHERE type = 6  -- Income
  AND deleted_at IS NULL
ORDER BY current_balance DESC;
```

### Accounts with Transaction Activity

```sql
SELECT
    a.id,
    a.name,
    a.current_balance,
    a.first_transaction_at,
    a.last_updated_at
FROM accounts a
WHERE a.deleted_at IS NULL
  AND a.first_transaction_at IS NOT NULL
ORDER BY a.last_updated_at DESC;
```

### Search Accounts by Name

```sql
SELECT * FROM accounts
WHERE deleted_at IS NULL
  AND name ILIKE '%' || :search || '%';
```

## Notes

- `current_balance` is automatically updated when transactions are created
- Use `display_order` for custom sorting in UI
- `first_transaction_at` is set when the first transaction uses this account
- `liability_percent` can track credit utilization (balance / credit limit * 100)
- Always filter with `deleted_at IS NULL` for active accounts
