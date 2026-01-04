# Transaction Types

Detailed documentation of each transaction type and their behavior.

## Type Summary

| Value | Name | Source Account | Destination Account |
|-------|------|----------------|---------------------|
| 1 | TRANSFER | Asset/Liability | Asset/Liability |
| 2 | INCOME | Income | Asset/Liability |
| 3 | EXPENSE | Asset/Liability | Expense |
| 5 | ADJUSTMENT | Adjustment | Any |

## Expense Transactions (type=3)

Records money spent from an account.

### Account Requirements
- **Source**: Asset (bank account) or Liability (credit card)
- **Destination**: Expense category account

### Amount Fields
| Field | Sign | Description |
|-------|------|-------------|
| source_amount | Negative | Amount deducted from account |
| source_currency | - | Currency of source account |
| destination_amount | Positive | Amount recorded as expense |
| destination_currency | - | Usually same as source |
| fx_source_amount | Negative | Original foreign currency amount (if different) |
| fx_source_currency | - | Original foreign currency code |

### Example: Grocery Purchase

```
Source: Checking Account (USD)
  source_amount: -50.00
  source_currency: USD

Destination: Groceries (Expense)
  destination_amount: 50.00
  destination_currency: USD
```

### Example: Foreign Currency Purchase

```
Source: Euro Account (EUR)
  source_amount: -46.00
  source_currency: EUR

Destination: Groceries (Expense)
  destination_amount: 46.00
  destination_currency: EUR

Foreign Currency Tracking:
  fx_source_amount: -50.00
  fx_source_currency: USD
```

## Income Transactions (type=2)

Records money received into an account.

### Account Requirements
- **Source**: Income account (salary, dividends, etc.)
- **Destination**: Asset or Liability account

### Amount Fields
| Field | Sign | Description |
|-------|------|-------------|
| source_amount | Positive | Amount from income source |
| source_currency | - | Currency of income |
| destination_amount | Positive | Amount deposited |
| destination_currency | - | Currency of receiving account |

### Example: Salary Deposit

```
Source: Salary (Income)
  source_amount: 5000.00
  source_currency: USD

Destination: Checking Account (Asset)
  destination_amount: 5000.00
  destination_currency: USD
```

### Example: Multi-Currency Income

```
Source: Freelance USD (Income)
  source_amount: 1000.00
  source_currency: USD

Destination: Euro Account (Asset)
  destination_amount: 920.00
  destination_currency: EUR
```

## Transfer Transactions (type=1)

Moves money between two Asset/Liability accounts.

### Account Requirements
- **Source**: Asset or Liability account
- **Destination**: Asset or Liability account (different from source)

### Amount Fields
| Field | Sign | Description |
|-------|------|-------------|
| source_amount | Negative | Amount leaving source |
| source_currency | - | Currency of source account |
| destination_amount | Positive | Amount entering destination |
| destination_currency | - | Currency of destination account |

### Example: Same Currency Transfer

```
Source: Checking Account (Asset)
  source_amount: -1000.00
  source_currency: USD

Destination: Savings Account (Asset)
  destination_amount: 1000.00
  destination_currency: USD
```

### Example: Currency Exchange

```
Source: USD Account (Asset)
  source_amount: -1000.00
  source_currency: USD

Destination: EUR Account (Asset)
  destination_amount: 920.00
  destination_currency: EUR
```

### Example: Credit Card Payment

```
Source: Checking Account (Asset)
  source_amount: -500.00
  source_currency: USD

Destination: Credit Card (Liability)
  destination_amount: 500.00
  destination_currency: USD
```

## Adjustment Transactions (type=5)

Corrects account balances without affecting expense/income totals.

### Account Requirements
- **Source**: Default Adjustment account (auto-selected)
- **Destination**: Account being adjusted

### Amount Fields
| Field | Sign | Description |
|-------|------|-------------|
| source_amount | Inverted | Opposite of destination |
| destination_amount | Variable | Adjustment amount |

### Example: Positive Adjustment

```
Adjustment: +100 to Checking Account

Source: Adjustment Account
  source_amount: -100.00

Destination: Checking Account
  destination_amount: 100.00
```

### Example: Negative Adjustment

```
Adjustment: -50 to Savings Account

Source: Adjustment Account
  source_amount: 50.00

Destination: Savings Account
  destination_amount: -50.00
```

## Double Entry Impact

Each transaction type creates two ledger entries:

| Transaction Type | Debit Account | Credit Account |
|------------------|---------------|----------------|
| Expense | Expense account | Asset/Liability account |
| Income | Asset/Liability account | Income account |
| Transfer | Destination account | Source account |
| Adjustment | Target or Adjustment | Adjustment or Target |

## Validation Rules

### Common Rules
- Transaction date is required
- At least one account must be specified
- Amounts must be valid decimals

### Type-Specific Rules

**Expense:**
- Source amount must be negative
- Destination amount must be positive
- Destination currency required

**Income:**
- Destination amount must be positive

**Transfer:**
- Source and destination accounts required
- Source amount must be negative
- Destination amount must be positive

**Adjustment:**
- Destination account and amount required
- Source account auto-selected
