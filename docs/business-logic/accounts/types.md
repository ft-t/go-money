# Account Types

Detailed documentation of account types and their roles in the Go Money system.

## Account Type Overview

| Value | Name | Purpose | Balance Meaning |
|-------|------|---------|-----------------|
| 1 | ASSET | Bank accounts, cash, investments | Positive = money you have |
| 4 | LIABILITY | Credit cards, loans, debts | Positive = money you owe |
| 5 | EXPENSE | Spending categories | Positive = money spent |
| 6 | INCOME | Income sources | Positive = money earned |
| 7 | ADJUSTMENT | Balance corrections | N/A (zeroes out) |

## Asset Accounts (type=1)

### Description

Asset accounts represent money you own or have access to. Examples:
- Checking accounts
- Savings accounts
- Cash on hand
- Investment accounts
- Digital wallets

### Balance Behavior

- **Positive balance**: You have money in this account
- **Negative balance**: Account is overdrawn (should be rare)

### Transaction Interactions

| Transaction Type | Role | Amount Sign |
|------------------|------|-------------|
| Transfer | Source or Destination | Negative (out) or Positive (in) |
| Expense | Source | Negative (money leaving) |
| Income | Destination | Positive (money arriving) |
| Adjustment | Destination | Variable (correction) |

### Double-Entry Behavior

Asset accounts are **debit-normal**:
- Positive amounts → Debit entry (increases balance)
- Negative amounts → Credit entry (decreases balance)

## Liability Accounts (type=4)

### Description

Liability accounts represent money you owe. Examples:
- Credit cards
- Loans
- Mortgages
- Lines of credit

### Balance Behavior

- **Positive balance**: Amount you currently owe
- **Negative balance**: You've overpaid (rare)

### Transaction Interactions

| Transaction Type | Role | Amount Sign |
|------------------|------|-------------|
| Transfer | Source or Destination | Pays down or adds debt |
| Expense | Source | Adds to debt (credit card purchase) |

### Special Feature: Liability Percent

Liability accounts can have a `liability_percent` field for tracking partial ownership or responsibility splits.

### Double-Entry Behavior

Liability accounts are **credit-normal**:
- Positive amounts → Credit entry (increases liability)
- Negative amounts → Debit entry (decreases liability)

## Expense Accounts (type=5)

### Description

Expense accounts categorize spending. Examples:
- Groceries
- Utilities
- Rent
- Entertainment
- Transportation

### Balance Behavior

- **Positive balance**: Total spent in this category
- Balance accumulates over time (no decreases except adjustments)

### Transaction Interactions

| Transaction Type | Role | Amount Sign |
|------------------|------|-------------|
| Expense | Destination only | Positive (expense recorded) |

### Note

Expense accounts can only be destinations for expense transactions. They cannot be sources or participate in transfers.

### Double-Entry Behavior

Expense accounts are **debit-normal**:
- Positive amounts → Debit entry (increases expense total)

## Income Accounts (type=6)

### Description

Income accounts track money sources. Examples:
- Salary
- Freelance income
- Investment dividends
- Gifts received
- Refunds

### Balance Behavior

- **Positive balance**: Total income received from this source

### Transaction Interactions

| Transaction Type | Role | Amount Sign |
|------------------|------|-------------|
| Income | Source only | Positive (income source) |

### Note

Income accounts can only be sources for income transactions. They cannot be destinations or participate in transfers.

### Double-Entry Behavior

Income accounts are **credit-normal**:
- Positive amounts → Credit entry (increases income total)

## Adjustment Accounts (type=7)

### Description

Adjustment accounts exist for balance corrections without affecting expense/income totals. Used when:
- Bank statement doesn't match recorded balance
- Initial balance setup
- Error corrections

### Balance Behavior

The default adjustment account accumulates all corrections. Its balance represents net adjustments made.

### Transaction Interactions

| Transaction Type | Role | Amount Sign |
|------------------|------|-------------|
| Adjustment | Source only | Auto-selected, inverted amount |

### Note

Adjustment accounts are automatically selected as the source for adjustment transactions. Users only specify the destination account and correction amount.

## Account Applicability Matrix

Which account types can be used for each transaction type:

| Transaction Type | Valid Source Types | Valid Destination Types |
|------------------|-------------------|------------------------|
| Transfer | Asset, Liability | Asset, Liability |
| Expense | Asset, Liability | Expense |
| Income | Income | Asset |
| Adjustment | Adjustment | Any (except Adjustment) |

**Code Reference:** `pkg/transactions/applicable_accounts/applicable_account.go:34-89`

## Default Accounts

The system requires one default account per type (except Asset which uses "Cash"):

| Default Name | Type |
|--------------|------|
| Cash | Asset (1) |
| Default Liability | Liability (4) |
| Default Expense | Expense (5) |
| Default Income | Income (6) |
| Default Adjustment | Adjustment (7) |

**Code Reference:** `pkg/boilerplate/default.go:5-11`

## Account Flags

Accounts use bitwise flags for properties:

| Flag | Value | Purpose |
|------|-------|---------|
| IsDefault | 1 (bit 0) | Marks account as default for its type |

### Checking Default Status

```sql
SELECT id, name, type
FROM accounts
WHERE flags & 1 = 1  -- IsDefault flag set
  AND deleted_at IS NULL;
```

## Account Fields

| Field | Type | Purpose |
|-------|------|---------|
| id | int32 | Primary key |
| name | text | Display name |
| currency | text | Currency code (e.g., "USD") |
| current_balance | numeric | Current balance |
| type | int | Account type (1, 4, 5, 6, 7) |
| flags | int64 | Bitwise flags |
| note | text | User notes |
| account_number | text | External account number |
| iban | text | IBAN for bank accounts |
| liability_percent | numeric | Ownership percentage (liabilities) |
| display_order | int32 | UI sort order |
| extra | jsonb | Custom key-value data |

## SQL Queries by Account Type

### All Asset and Liability Accounts (Net Worth)

```sql
SELECT id, name, type, current_balance, currency
FROM accounts
WHERE type IN (1, 4)  -- Asset, Liability
  AND deleted_at IS NULL
ORDER BY display_order NULLS LAST;
```

### All Expense Categories

```sql
SELECT id, name, current_balance as total_spent
FROM accounts
WHERE type = 5  -- Expense
  AND deleted_at IS NULL
ORDER BY current_balance DESC;
```

### All Income Sources

```sql
SELECT id, name, current_balance as total_earned
FROM accounts
WHERE type = 6  -- Income
  AND deleted_at IS NULL
ORDER BY current_balance DESC;
```

### Default Account for Each Type

```sql
SELECT id, name, type
FROM accounts
WHERE flags & 1 = 1
  AND deleted_at IS NULL
ORDER BY type;
```

### Accounts by Currency

```sql
SELECT currency,
       COUNT(*) as account_count,
       SUM(CASE WHEN type = 1 THEN current_balance ELSE 0 END) as total_assets,
       SUM(CASE WHEN type = 4 THEN current_balance ELSE 0 END) as total_liabilities
FROM accounts
WHERE type IN (1, 4)
  AND deleted_at IS NULL
GROUP BY currency;
```

## Validation Rules

1. **Unique combination**: Name + Type + Currency must be unique
2. **Default required**: Each type must have at least one default account
3. **Soft delete**: Accounts are soft-deleted (deleted_at timestamp)
4. **Currency immutable**: Currency should not change after creation (transactions reference it)
