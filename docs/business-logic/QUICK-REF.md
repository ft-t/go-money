# Business Logic Quick Reference

> Key formulas, rules, and behaviors. For details, see individual docs.

## Transaction Processing Flow

```
1. Input validation
2. Apply Lua rules (sorted by group, then sort_order)
3. Calculate base currency amounts
4. Create double-entry ledger records
5. Update account balances
6. Update daily statistics
```

---

## Transaction Types

| Type | Value | Source → Destination | Balance Effect |
|------|-------|----------------------|----------------|
| Transfer | 1 | Asset/Liability → Asset/Liability | Source -, Dest + |
| Income | 2 | Income → Asset | Asset + |
| Expense | 3 | Asset/Liability → Expense | Asset - |
| Adjustment | 5 | Adjustment → Asset/Liability | Asset +/- |

---

## Account Type Rules

### Which accounts can be source/destination?

| Transaction Type | Valid Source Types | Valid Destination Types |
|------------------|--------------------|-----------------------|
| Transfer (1) | Asset (1), Liability (4) | Asset (1), Liability (4) |
| Income (2) | Income (6) | Asset (1) |
| Expense (3) | Asset (1), Liability (4) | Expense (5) |
| Adjustment (5) | Adjustment (7) | Asset (1), Liability (4) |

### Default Accounts (auto-created)
- "Cash" → Asset
- "Default Expense" → Expense
- "Default Income" → Income
- "Default Liability" → Liability
- "Default Adjustment" → Adjustment

---

## Amount Calculations

### Base Currency Conversion

```sql
-- Priority order:
1. source_currency = base_currency → use source_amount
2. destination_currency = base_currency → use destination_amount
3. fx_source_currency = base_currency → use fx_source_amount
4. Otherwise → source_amount / currency_rate
```

**Formula:** `base_amount = amount / rate`

### Currency Rate Storage
- Rate = units of currency per 1 base currency
- Example: If base is USD and EUR rate is 0.92, then 1 USD = 0.92 EUR

---

## Double-Entry Rules

### Entry Creation
Every transaction creates exactly 2 entries:
- One debit (is_debit = true)
- One credit (is_debit = false)

### Debit/Credit Assignment

| Account Type | Positive Amount | Negative Amount |
|--------------|-----------------|-----------------|
| Asset (1) | Debit | Credit |
| Expense (5) | Debit | Credit |
| Liability (4) | Credit | Debit |
| Income (6) | Credit | Debit |

### Balance Equation
```
Total Debits = Total Credits (always)
Assets + Expenses = Liabilities + Income + Equity
```

---

## Daily Statistics

### What is daily_stat?
Pre-computed running balance for each account at end of each day.

### Update Triggers
- Transaction created
- Transaction updated
- Transaction deleted

### Recalculation Query
```sql
SELECT account_id, date,
       SUM(amount) OVER (PARTITION BY account_id ORDER BY date) as running_balance
FROM daily_stat
```

---

## Lua Rule Engine

### Execution Order
1. Rules grouped by `group` field
2. Within group, sorted by `sort_order`
3. If rule has `is_final = true`, stop processing

### Available in Lua Scripts

```lua
-- Transaction object
tx.title                    -- string
tx.transaction_type         -- number (1,2,3,5)
tx.source_account_id        -- number
tx.destination_account_id   -- number
tx.source_amount            -- string (decimal)
tx.destination_amount       -- string (decimal)
tx.category_id              -- number or nil
tx.tag_ids                  -- table of numbers

-- Helper functions
set_category(name)          -- Set category by name
set_tag(name)               -- Add tag by name
get_account_by_name(name)   -- Get account ID
convert(amount, from, to)   -- Currency conversion
```

---

## Key Formulas

### Net Worth
```sql
SELECT SUM(CASE
    WHEN type = 1 THEN current_balance  -- Assets
    WHEN type = 4 THEN -current_balance -- Liabilities
    ELSE 0
END) as net_worth
FROM accounts
WHERE type IN (1, 4) AND deleted_at IS NULL;
```

### Monthly Spending
```sql
SELECT SUM(destination_amount_in_base_currency)
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE);
```

### Account Balance at Date
```sql
SELECT amount as balance
FROM daily_stat
WHERE account_id = :id AND date <= :target_date
ORDER BY date DESC
LIMIT 1;
```

---

## See Also
- [Transaction Types](transactions/types.md) - Detailed type documentation
- [Amount Calculations](transactions/amount-calculations.md) - Full conversion logic
- [Double-Entry Overview](double-entry/overview.md) - Bookkeeping details
- [Rules Engine](rules-engine/overview.md) - Lua API reference
