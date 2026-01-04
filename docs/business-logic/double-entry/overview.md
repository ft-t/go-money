# Double-Entry Bookkeeping

How Go Money implements double-entry accounting principles.

## Concept Overview

Double-entry bookkeeping is an accounting system where every transaction creates two entries:
- **Debit entry**: Increases asset/expense accounts, decreases liability/income accounts
- **Credit entry**: Decreases asset/expense accounts, increases liability/income accounts

**Fundamental Rule**: Total Debits = Total Credits (always balanced)

## Account Type Classifications

### Debit-Normal Accounts

Accounts that increase with debits:
- **Asset** (type=1): Bank accounts, cash, investments
- **Expense** (type=5): Spending categories

```go
func isDebitNormal(accountType AccountType) bool {
    switch accountType {
    case ASSET, EXPENSE:
        return true
    default:
        return false
    }
}
```

### Credit-Normal Accounts

Accounts that increase with credits:
- **Liability** (type=4): Credit cards, loans
- **Income** (type=6): Salary, dividends
- **Adjustment** (type=7): Balance corrections

## Debit/Credit Determination

The system determines whether an entry is debit or credit based on:
1. Account type (debit-normal vs credit-normal)
2. Sign of the amount

```go
func isDebit(accountType AccountType, amount decimal.Decimal) bool {
    if isDebitNormal(accountType) {
        return amount.IsPositive()  // debit-normal: + => debit, - => credit
    }
    return amount.IsNegative()      // credit-normal: - => debit, + => credit
}
```

### Decision Matrix

| Account Type | Amount Sign | Entry Type |
|--------------|-------------|------------|
| Asset | Positive | Debit |
| Asset | Negative | Credit |
| Expense | Positive | Debit |
| Expense | Negative | Credit |
| Liability | Positive | Credit |
| Liability | Negative | Debit |
| Income | Positive | Credit |
| Income | Negative | Debit |
| Adjustment | Positive | Credit |
| Adjustment | Negative | Debit |

**Code Reference:** `pkg/transactions/double_entry/double_entry.go:171-187`

## Entry Creation

For each transaction, exactly two entries are created:

```go
entries := []*DoubleEntry{
    {
        TransactionID:        tx.ID,
        IsDebit:              isDebit,
        AccountID:            tx.SourceAccountID,
        AmountInBaseCurrency: baseAmount.Abs(),
    },
    {
        TransactionID:        tx.ID,
        IsDebit:              !isDebit,  // Opposite of source
        AccountID:            tx.DestinationAccountID,
        AmountInBaseCurrency: baseAmount.Abs(),
    },
}
```

**Key Points:**
- One entry is always debit, the other is always credit
- Both entries have the same absolute amount
- Amount is always positive (direction indicated by is_debit flag)

**Code Reference:** `pkg/transactions/double_entry/double_entry.go:147-166`

## Transaction Type Examples

### Expense: Groceries Purchase ($50)

```
Source: Checking Account (Asset)
  source_amount_in_base_currency: -50.00

Destination: Groceries (Expense)
  destination_amount_in_base_currency: 50.00
```

**Double Entries:**
| Account | Type | Amount | Is Debit |
|---------|------|--------|----------|
| Groceries (Expense) | Debit-normal | $50.00 | true |
| Checking (Asset) | Debit-normal | $50.00 | false (credit) |

### Income: Salary ($5000)

```
Source: Salary (Income)
  source_amount_in_base_currency: 5000.00

Destination: Checking Account (Asset)
  destination_amount_in_base_currency: 5000.00
```

**Double Entries:**
| Account | Type | Amount | Is Debit |
|---------|------|--------|----------|
| Salary (Income) | Credit-normal | $5000.00 | false (credit) |
| Checking (Asset) | Debit-normal | $5000.00 | true (debit) |

### Transfer: Checking to Savings ($1000)

```
Source: Checking Account (Asset)
  source_amount_in_base_currency: -1000.00

Destination: Savings Account (Asset)
  destination_amount_in_base_currency: 1000.00
```

**Double Entries:**
| Account | Type | Amount | Is Debit |
|---------|------|--------|----------|
| Checking (Asset) | Debit-normal | $1000.00 | false (credit) |
| Savings (Asset) | Debit-normal | $1000.00 | true (debit) |

### Credit Card Payment ($500)

```
Source: Checking Account (Asset)
  source_amount_in_base_currency: -500.00

Destination: Credit Card (Liability)
  destination_amount_in_base_currency: 500.00
```

**Double Entries:**
| Account | Type | Amount | Is Debit |
|---------|------|--------|----------|
| Checking (Asset) | Debit-normal | $500.00 | false (credit) |
| Credit Card (Liability) | Credit-normal | $500.00 | true (debit) |

## Validation Rules

Before creating entries, the system validates:

1. **Source account exists**: Required for determining debit/credit
2. **Both account IDs set**: source_account_id and destination_account_id required
3. **Amounts balanced**: source and destination amounts in base currency must be equal (absolute value)
4. **Opposite signs**: source and destination amounts must have opposite signs

```go
if !tx.SourceAmountInBaseCurrency.Abs().Equal(tx.DestinationAmountInBaseCurrency.Abs()) {
    return nil, errors.New("amounts must be equal")
}

if tx.SourceAmountInBaseCurrency.Sign() == tx.DestinationAmountInBaseCurrency.Sign() {
    return nil, errors.New("amounts must have opposite signs")
}
```

**Code Reference:** `pkg/transactions/double_entry/double_entry.go:133-141`

## Database Schema

### double_entries Table

| Column | Type | Description |
|--------|------|-------------|
| id | bigint | Primary key |
| transaction_id | bigint | FK to transactions |
| account_id | integer | FK to accounts |
| is_debit | boolean | true=debit, false=credit |
| amount_in_base_currency | numeric | Always positive |
| base_currency | text | Currency code |
| transaction_date | timestamp | For sorting |
| created_at | timestamp | Record creation |
| deleted_at | timestamp | Soft delete |

### Unique Constraint

```sql
UNIQUE (transaction_id, is_debit) WHERE deleted_at IS NULL
```

Ensures exactly one debit and one credit per transaction.

## SQL Queries

### Account Ledger

```sql
SELECT
    de.transaction_date,
    t.title,
    CASE WHEN de.is_debit THEN de.amount_in_base_currency ELSE NULL END as debit,
    CASE WHEN NOT de.is_debit THEN de.amount_in_base_currency ELSE NULL END as credit
FROM double_entries de
JOIN transactions t ON t.id = de.transaction_id
WHERE de.account_id = :account_id
  AND de.deleted_at IS NULL
ORDER BY de.transaction_date DESC;
```

### Trial Balance

```sql
SELECT
    a.name,
    a.type,
    SUM(CASE WHEN de.is_debit THEN de.amount_in_base_currency ELSE 0 END) as total_debits,
    SUM(CASE WHEN NOT de.is_debit THEN de.amount_in_base_currency ELSE 0 END) as total_credits
FROM double_entries de
JOIN accounts a ON a.id = de.account_id
WHERE de.deleted_at IS NULL
  AND a.deleted_at IS NULL
GROUP BY a.id, a.name, a.type
ORDER BY a.type, a.name;
```

### Verify Balance (Debits = Credits)

```sql
SELECT
    SUM(CASE WHEN is_debit THEN amount_in_base_currency ELSE 0 END) as total_debits,
    SUM(CASE WHEN NOT is_debit THEN amount_in_base_currency ELSE 0 END) as total_credits
FROM double_entries
WHERE deleted_at IS NULL;
-- These two values should always be equal
```

### Find Unbalanced Transactions

```sql
SELECT
    transaction_id,
    SUM(CASE WHEN is_debit THEN amount_in_base_currency ELSE -amount_in_base_currency END) as balance
FROM double_entries
WHERE deleted_at IS NULL
GROUP BY transaction_id
HAVING SUM(CASE WHEN is_debit THEN amount_in_base_currency ELSE -amount_in_base_currency END) != 0;
-- Should return no rows
```

## Update/Delete Handling

### Transaction Update

When a transaction is updated:
1. Soft-delete existing double entries
2. Create new entries with updated amounts

```go
// Soft-delete existing
dbTx.Exec("UPDATE double_entries SET deleted_at = now()
           WHERE transaction_id IN ? AND deleted_at IS NULL", txIds)

// Create new
dbTx.CreateInBatches(entries, batchSize)
```

### Transaction Delete

When a transaction is deleted:
1. Soft-delete corresponding double entries

```go
dbTx.Exec("UPDATE double_entries SET deleted_at = now()
           WHERE transaction_id IN ? AND deleted_at IS NULL", txIds)
```

**Code Reference:** `pkg/transactions/double_entry/double_entry.go:38-56`

## Precision

The system uses 18 decimal places for amount comparison:

```go
const roundPlaces = 18

if !sourceAmount.Abs().Round(roundPlaces).Equal(destAmount.Abs().Round(roundPlaces)) {
    return nil, errors.New("amounts must be equal")
}
```

This handles floating-point precision issues in currency conversions.

## Use Cases

### When to Query double_entries

| Use Case | Preferred Table |
|----------|-----------------|
| Formal accounting reports | double_entries |
| Account reconciliation | double_entries |
| Audit trail | double_entries |
| Quick balance lookup | daily_stat or accounts |
| Spending analysis | transactions |

### Advantages Over Transactions Table

- Pre-computed debit/credit classification
- Amounts always in base currency
- Clean accounting view of each account
- Proper ledger format for reports


---

## See Also

- [double_entries Table](../../schema/tables/double_entry.md) - Schema definition
- [Transaction Types](../transactions/types.md) - Debit/credit rules per type
- [Business Logic Quick-Ref](../QUICK-REF.md) - Key formulas
