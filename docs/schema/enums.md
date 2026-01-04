# Enum Reference

This document lists all enum values used in the Go Money database.

## TransactionType

Stored in `transactions.transaction_type` as integer.

| Value | Name | Description |
|-------|------|-------------|
| 0 | UNSPECIFIED | Not used in practice |
| 1 | TRANSFER | Transfer between Asset/Liability accounts |
| 2 | INCOME | Money received into an account |
| 3 | EXPENSE | Money spent from an account |
| 4 | REVERSAL | Void/reversal of another transaction |
| 5 | ADJUSTMENT | Balance adjustment entry |

### Transaction Type Usage

```sql
-- Get all expense transactions
SELECT * FROM transactions
WHERE transaction_type = 3 AND deleted_at IS NULL;

-- Get income and expense for a period
SELECT
    CASE transaction_type
        WHEN 2 THEN 'Income'
        WHEN 3 THEN 'Expense'
    END as type,
    SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type IN (2, 3)
  AND deleted_at IS NULL
  AND transaction_date_only BETWEEN '2024-01-01' AND '2024-12-31'
GROUP BY transaction_type;
```

## AccountType

Stored in `accounts.type` as integer.

| Value | Name | Description | Normal Balance |
|-------|------|-------------|----------------|
| 0 | UNSPECIFIED | Not used in practice | - |
| 1 | ASSET | Cash, bank accounts, investments | Debit (positive) |
| 4 | LIABILITY | Credit cards, loans, debts | Credit (negative) |
| 5 | EXPENSE | Spending categories (groceries, utilities) | Debit (positive) |
| 6 | INCOME | Revenue sources (salary, dividends) | Credit (negative) |
| 7 | ADJUSTMENT | Special account for balance adjustments | Variable |

### Account Type Usage

```sql
-- Get all asset accounts with balances
SELECT name, current_balance, currency
FROM accounts
WHERE type = 1 AND deleted_at IS NULL;

-- Calculate net worth (Assets - Liabilities)
SELECT
    SUM(CASE WHEN type = 1 THEN current_balance ELSE 0 END) as assets,
    SUM(CASE WHEN type = 4 THEN current_balance ELSE 0 END) as liabilities,
    SUM(CASE
        WHEN type = 1 THEN current_balance
        WHEN type = 4 THEN -current_balance
        ELSE 0
    END) as net_worth
FROM accounts
WHERE type IN (1, 4) AND deleted_at IS NULL;
```

## AccountFlags

Stored in `accounts.flags` as bigint. Uses bitmask pattern.

| Bit | Value | Name | Description |
|-----|-------|------|-------------|
| 0 | 1 | IsDefault | Default account for its type |

### AccountFlags Usage

```sql
-- Get default accounts
SELECT * FROM accounts
WHERE (flags & 1) = 1 AND deleted_at IS NULL;

-- Check if account is default
SELECT
    name,
    CASE WHEN (flags & 1) = 1 THEN true ELSE false END as is_default
FROM accounts
WHERE deleted_at IS NULL;
```

## RuleInterpreterType

Stored in `rules.interpreter_type` and `schedule_rules.interpreter_type` as integer.

| Value | Name | Description |
|-------|------|-------------|
| 0 | UNSPECIFIED | Not used in practice |
| 1 | LUA | Lua scripting language |

### Interpreter Type Usage

```sql
-- Get all Lua rules
SELECT * FROM rules
WHERE interpreter_type = 1 AND deleted_at IS NULL
ORDER BY sort_order;
```

## ImportSource

Stored in `import_deduplication.import_source` as integer.

| Value | Name | Description |
|-------|------|-------------|
| 0 | UNSPECIFIED | Not used in practice |
| 1 | FIREFLY | Imported from Firefly III |

### Import Source Usage

```sql
-- Check imported transactions from Firefly
SELECT i.*, t.title, t.transaction_date_only
FROM import_deduplication i
JOIN transactions t ON t.id = i.transaction_id
WHERE i.import_source = 1
ORDER BY i.created_at DESC;
```

## Boolean Fields

Several tables use boolean fields:

| Table | Field | Purpose |
|-------|-------|---------|
| currencies | is_active | Whether currency is available for use |
| rules | enabled | Whether rule is active |
| rules | is_final_rule | Stop processing further rules after this one |
| schedule_rules | enabled | Whether scheduled rule is active |
| double_entries | is_debit | True for debit entries, false for credit |

### Boolean Usage Examples

```sql
-- Active currencies only
SELECT * FROM currencies WHERE is_active = true;

-- Enabled rules in execution order
SELECT * FROM rules
WHERE enabled = true AND deleted_at IS NULL
ORDER BY sort_order;

-- Debit entries for an account
SELECT * FROM double_entries
WHERE account_id = 1 AND is_debit = true AND deleted_at IS NULL;
```
