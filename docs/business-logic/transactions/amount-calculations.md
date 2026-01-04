# Amount Calculations

How transaction amounts are stored, calculated, and converted to base currency.

## Amount Fields Overview

Each transaction has multiple amount fields:

| Field | Purpose | Sign Convention |
|-------|---------|-----------------|
| source_amount | Amount leaving source account | Usually negative |
| source_currency | Currency of source account | - |
| destination_amount | Amount entering destination | Usually positive |
| destination_currency | Currency of destination | - |
| fx_source_amount | Original foreign amount (expenses) | Negative |
| fx_source_currency | Original foreign currency | - |
| source_amount_in_base_currency | Converted source amount | Always negative |
| destination_amount_in_base_currency | Converted destination | Always positive |

## Currency Rate Storage

Currencies store rates relative to base currency:

```
rate = units of currency per 1 base currency unit
```

Example with USD as base currency:
- USD: rate = 1.0
- EUR: rate = 0.92 (1 USD = 0.92 EUR)
- GBP: rate = 0.79 (1 USD = 0.79 GBP)

## Base Currency Conversion Formula

```
base_amount = amount / rate
```

**Example**: 100 EUR to USD (base)
```
100 EUR / 0.92 = 108.70 USD
```

## Conversion Priority Logic

The system uses a priority-based approach to determine base currency amounts:

```
Priority Order:
1. If source_currency = base_currency → use source_amount directly
2. If expense (type=3) AND fx_source_currency = base_currency → use fx_source_amount
3. If destination_currency = base_currency → use destination_amount
4. Otherwise → convert using currency rate table
```

### Priority 1: Source Already in Base Currency

```sql
WHEN source_currency = @baseCurrency
THEN source_amount
```

No conversion needed when source is already in base currency.

### Priority 2: Foreign Currency Expense with Base FX

```sql
WHEN transaction_type = 3
 AND fx_source_currency = @baseCurrency
 AND fx_source_amount IS NOT NULL
THEN fx_source_amount
```

For expenses where the foreign currency tracking (`fx_source_*`) is in base currency, use that amount directly.

### Priority 3: Destination in Base Currency

```sql
WHEN source_amount IS NOT NULL
 AND destination_amount IS NOT NULL
 AND destination_currency = @baseCurrency
THEN destination_amount
```

If the other side of the transaction is in base currency, use it to avoid conversion.

### Priority 4: Rate Table Conversion

```sql
WHEN source_currency != @baseCurrency
THEN source_amount / currency.rate
```

Convert using the stored exchange rate.

## Sign Conventions for Base Amounts

After conversion, amounts are normalized:

```sql
destination_amount_in_base_currency = ABS(calculated_amount)  -- Always positive
source_amount_in_base_currency = -ABS(calculated_amount)      -- Always negative
```

This ensures consistent signs regardless of how the original amounts were stored.

## Decimal Precision

Converted amounts respect currency decimal places:

```sql
ROUND(amount / rate, currency.decimal_places)
```

Falls back to unrounded value if rounding produces zero:

```sql
COALESCE(
    NULLIF(ROUND(amount / rate, decimal_places), 0),
    amount / rate
)
```

## Multi-Currency Transaction Examples

### Same Currency Transfer

```
USD Account → USD Account
source_amount: -100.00 USD
destination_amount: 100.00 USD

Base (USD):
source_amount_in_base_currency: -100.00
destination_amount_in_base_currency: 100.00
```

### Currency Exchange Transfer

```
EUR Account → USD Account
source_amount: -92.00 EUR
destination_amount: 100.00 USD

Base (USD):
source_amount_in_base_currency: -100.00  (uses destination since it's base)
destination_amount_in_base_currency: 100.00
```

### Foreign Currency Expense

```
EUR Card → Expense Category
source_amount: -50.00 EUR
destination_amount: 50.00 EUR
fx_source_amount: -55.00 USD  (what you were charged in USD)
fx_source_currency: USD

Base (USD):
source_amount_in_base_currency: -55.00  (uses fx_source since it's base)
destination_amount_in_base_currency: 55.00
```

### Expense Requiring Conversion

```
EUR Card → Expense Category
source_amount: -50.00 EUR
destination_amount: 50.00 EUR

Base (USD), EUR rate = 0.92:
source_amount_in_base_currency: -54.35  (50 / 0.92)
destination_amount_in_base_currency: 54.35
```

## Recalculation Triggers

Base currency amounts are recalculated when:

1. **Transaction created** - Initial calculation during creation
2. **Transaction updated** - Recalculated on any modification
3. **Currency rates updated** - Batch recalculation via `RecalculateAmountInBaseCurrencyForAll`
4. **Base currency changed** - Full recalculation required

## Implementation Details

**Code Reference:** `pkg/transactions/base_amount.go`

The `BaseAmountService` handles conversions:

```go
type BaseAmountService struct {
    baseCurrency string
}

// Recalculate specific transactions
func (s *BaseAmountService) RecalculateAmountInBaseCurrency(
    ctx context.Context,
    tx *gorm.DB,
    specificTxIDs []*database.Transaction,
) error

// Recalculate all transactions
func (s *BaseAmountService) RecalculateAmountInBaseCurrencyForAll(
    ctx context.Context,
    tx *gorm.DB,
) error
```

## SQL Queries for Amount Analysis

### Sum Expenses in Base Currency

```sql
SELECT
    DATE_TRUNC('month', transaction_date_only) as month,
    SUM(destination_amount_in_base_currency) as total_spent
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
GROUP BY DATE_TRUNC('month', transaction_date_only)
ORDER BY month DESC;
```

### Compare Original vs Base Amounts

```sql
SELECT
    title,
    source_amount,
    source_currency,
    source_amount_in_base_currency,
    ROUND(source_amount / source_amount_in_base_currency, 4) as effective_rate
FROM transactions
WHERE source_currency != 'USD'
  AND deleted_at IS NULL
ORDER BY transaction_date_time DESC
LIMIT 20;
```

### Transactions with Currency Mismatch

```sql
SELECT *
FROM transactions
WHERE source_currency != destination_currency
  AND deleted_at IS NULL
ORDER BY transaction_date_time DESC;
```

## Notes

- Base currency amounts enable consistent reporting across multi-currency accounts
- The fx_source fields are specifically for tracking original foreign currency on expenses (e.g., credit card charges abroad)
- Rate changes don't retroactively affect historical base amounts unless explicitly recalculated
- For accurate historical reporting, consider storing the rate used at transaction time
