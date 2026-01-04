# Currency Conversion

How Go Money handles multi-currency transactions and base currency conversion.

## Base Currency Concept

Go Money uses a **base currency** (configured per installation) for:
- Unified reporting across accounts
- Consistent analytics
- Net worth calculations
- Double-entry bookkeeping entries

All amounts are converted to base currency for these purposes while preserving original currency amounts.

## Currency Table Schema

| Column | Type | Description |
|--------|------|-------------|
| id | text | Currency code (e.g., "USD", "EUR") - Primary Key |
| rate | numeric | Exchange rate vs base currency |
| is_active | boolean | Whether currency is enabled |
| decimal_places | integer | Precision for this currency |
| updated_at | timestamp | Last rate update |
| deleted_at | timestamp | Soft delete |

## Rate Storage Convention

Rates are stored as: **units of currency per 1 base currency unit**

Example with USD as base currency:
| Currency | Rate | Meaning |
|----------|------|---------|
| USD | 1.0000 | Base currency |
| EUR | 0.8529 | 1 USD = 0.8529 EUR |
| PLN | 3.5924 | 1 USD = 3.5924 PLN |
| UAH | 42.3078 | 1 USD = 42.3078 UAH |

## Conversion Formula

### To Base Currency

```
base_amount = foreign_amount / rate
```

**Example**: Convert 100 EUR to USD
```
100 EUR / 0.8529 = 117.25 USD
```

### From Base Currency

```
foreign_amount = base_amount * rate
```

**Example**: Convert 100 USD to EUR
```
100 USD * 0.8529 = 85.29 EUR
```

## Decimal Precision

Conversions respect currency-specific decimal places:

```sql
ROUND(amount / rate, decimal_places)
```

With fallback for near-zero results:
```sql
COALESCE(
    NULLIF(ROUND(amount / rate, decimal_places), 0),
    amount / rate
)
```

This prevents rounding to zero for small amounts.

## Transaction Amount Fields

Multi-currency transactions use these fields:

| Field | Purpose |
|-------|---------|
| source_amount | Amount in source account currency |
| source_currency | Source currency code |
| destination_amount | Amount in destination currency |
| destination_currency | Destination currency code |
| fx_source_amount | Original foreign amount (expenses only) |
| fx_source_currency | Original foreign currency |
| source_amount_in_base_currency | Converted source amount |
| destination_amount_in_base_currency | Converted destination amount |

## Conversion Priority Logic

The system uses smart conversion to avoid unnecessary rate lookups:

### Priority 1: Source Already Base Currency
```sql
WHEN source_currency = @baseCurrency
THEN source_amount
```

### Priority 2: FX Fields in Base Currency (Expenses)
```sql
WHEN transaction_type = 3  -- Expense
 AND fx_source_currency = @baseCurrency
 AND fx_source_amount IS NOT NULL
THEN fx_source_amount
```

### Priority 3: Destination Already Base Currency
```sql
WHEN destination_currency = @baseCurrency
THEN destination_amount
```

### Priority 4: Rate Table Conversion
```sql
WHEN source_currency != @baseCurrency
THEN source_amount / currency.rate
```

**Code Reference:** `pkg/transactions/scripts/update_amount_in_base_currency.sql`

## Multi-Currency Scenarios

### Same Currency Transaction

```
USD Account → USD Expense
source: -50.00 USD
destination: 50.00 USD
source_in_base: -50.00 USD
destination_in_base: 50.00 USD
```

No conversion needed.

### Currency Exchange (Transfer)

```
USD Account → EUR Account
source: -100.00 USD
destination: 85.29 EUR

source_in_base: -100.00 USD (already base)
destination_in_base: 100.00 USD (uses source since it's base)
```

### Foreign Currency Expense

```
EUR Card → Expense Category
source: -85.29 EUR
destination: 85.29 EUR
fx_source: -100.00 USD (actual charge)
fx_source_currency: USD

source_in_base: -100.00 USD (uses fx_source)
destination_in_base: 100.00 USD
```

### Foreign Expense with Conversion

```
EUR Card → Expense Category
source: -85.29 EUR
destination: 85.29 EUR
(no fx fields)

EUR rate: 0.8529
source_in_base: -100.00 USD (85.29 / 0.8529)
destination_in_base: 100.00 USD
```

## SQL Queries

### All Active Currencies

```sql
SELECT id, rate, decimal_places
FROM currencies
WHERE is_active = true
  AND deleted_at IS NULL
ORDER BY id;
```

### Convert Amount to Base Currency

```sql
SELECT
    amount / c.rate as amount_in_base
FROM currencies c
WHERE c.id = :currency_code;
```

### Transactions with Currency Conversion

```sql
SELECT
    t.id,
    t.title,
    t.source_amount,
    t.source_currency,
    t.source_amount_in_base_currency,
    ABS(t.source_amount / t.source_amount_in_base_currency) as effective_rate
FROM transactions t
WHERE t.source_currency != 'USD'  -- Replace with base currency
  AND t.deleted_at IS NULL
ORDER BY t.transaction_date_time DESC;
```

### Spending by Currency

```sql
SELECT
    t.source_currency,
    COUNT(*) as transaction_count,
    SUM(ABS(t.source_amount)) as total_in_currency,
    SUM(t.destination_amount_in_base_currency) as total_in_base
FROM transactions t
WHERE t.transaction_type = 3  -- Expense
  AND t.deleted_at IS NULL
GROUP BY t.source_currency
ORDER BY total_in_base DESC;
```

### Accounts by Currency

```sql
SELECT
    currency,
    COUNT(*) as account_count,
    SUM(current_balance) as total_balance
FROM accounts
WHERE type IN (1, 4)  -- Asset, Liability
  AND deleted_at IS NULL
GROUP BY currency;
```

## Rate Updates

When currency rates are updated, transactions can be recalculated:

```go
// Recalculate all transactions
baseAmountService.RecalculateAmountInBaseCurrencyForAll(ctx, tx)

// Recalculate specific transactions
baseAmountService.RecalculateAmountInBaseCurrency(ctx, tx, transactions)
```

**Code Reference:** `pkg/transactions/base_amount.go:30-82`

## FX Source Fields (Foreign Currency Tracking)

The `fx_source_*` fields are specifically for expense transactions where:
- Payment was made in one currency
- But charged to an account in a different currency

**Use Case**: Credit card in EUR, purchase abroad in USD

```
Card charges: 85.29 EUR
Actual purchase: 100.00 USD

Fields:
source_amount: -85.29
source_currency: EUR
fx_source_amount: -100.00
fx_source_currency: USD
```

This preserves both:
- What was deducted from your account (EUR)
- What you actually paid at the merchant (USD)

## Multi-Currency Net Worth

Calculate net worth across all currencies:

```sql
SELECT
    SUM(CASE WHEN a.type = 1
        THEN a.current_balance / COALESCE(c.rate, 1)
        ELSE 0 END) as total_assets_in_base,
    SUM(CASE WHEN a.type = 4
        THEN a.current_balance / COALESCE(c.rate, 1)
        ELSE 0 END) as total_liabilities_in_base
FROM accounts a
LEFT JOIN currencies c ON c.id = a.currency
WHERE a.deleted_at IS NULL
  AND a.type IN (1, 4);
```

## Important Notes

1. **Rate Timing**: Rates at transaction time are used; rate changes don't retroactively affect past transactions unless explicitly recalculated

2. **Base Currency Change**: Changing base currency requires recalculating all base currency amounts

3. **Precision**: Use the currency's `decimal_places` for display; internal calculations use higher precision

4. **Missing Rates**: If a currency rate is missing, conversion may fail or use a default

5. **Rate Sources**: Rates are typically updated from external sources (APIs) and stored in the currencies table
