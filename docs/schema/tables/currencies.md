# currencies Table

The `currencies` table stores currency definitions and exchange rates.

## Schema

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | text | NO | - | ISO 4217 currency code (e.g., "USD", "EUR") |
| rate | numeric | NO | 1 | Exchange rate relative to base currency |
| is_active | boolean | NO | false | Whether currency is available for use |
| decimal_places | integer | NO | 2 | Number of decimal places for display |
| updated_at | timestamp | NO | - | Rate update timestamp |
| deleted_at | timestamp | YES | - | Soft delete timestamp |

## Primary Key

- `id` (text - ISO 4217 currency code)

## Indexes

| Index | Definition | Purpose |
|-------|------------|---------|
| currencies_pk | UNIQUE (id) | Primary key |

## Exchange Rate Concept

The `rate` field represents the exchange rate relative to your **base currency**:

- Base currency always has `rate = 1`
- Other currencies: `amount_in_base = amount * rate`
- Example: If USD is base and EUR rate is 0.92, then 100 EUR = 92 USD

## Common Queries

### All Active Currencies

```sql
SELECT id, rate, decimal_places
FROM currencies
WHERE is_active = true
  AND deleted_at IS NULL
ORDER BY id;
```

### Get Base Currency

```sql
SELECT * FROM currencies
WHERE rate = 1
  AND is_active = true
  AND deleted_at IS NULL;
```

### Convert Amount to Base Currency

```sql
SELECT
    100 as original_amount,
    'EUR' as original_currency,
    100 * rate as amount_in_base
FROM currencies
WHERE id = 'EUR';
```

### Currency Usage Statistics

```sql
SELECT
    c.id as currency,
    c.rate,
    COUNT(DISTINCT a.id) as account_count
FROM currencies c
LEFT JOIN accounts a ON a.currency = c.id AND a.deleted_at IS NULL
WHERE c.deleted_at IS NULL
GROUP BY c.id, c.rate
ORDER BY account_count DESC;
```

### Accounts by Currency

```sql
SELECT
    a.currency,
    c.rate,
    COUNT(*) as account_count,
    SUM(a.current_balance) as total_in_currency,
    SUM(a.current_balance * c.rate) as total_in_base
FROM accounts a
JOIN currencies c ON c.id = a.currency
WHERE a.type IN (1, 4)  -- Asset, Liability
  AND a.deleted_at IS NULL
  AND c.deleted_at IS NULL
GROUP BY a.currency, c.rate;
```

### Net Worth in Base Currency

```sql
SELECT
    SUM(
        CASE WHEN a.type = 1 THEN a.current_balance * c.rate
             WHEN a.type = 4 THEN -a.current_balance * c.rate
             ELSE 0
        END
    ) as net_worth_in_base
FROM accounts a
JOIN currencies c ON c.id = a.currency
WHERE a.type IN (1, 4)
  AND a.deleted_at IS NULL;
```

### Currency Exposure

```sql
SELECT
    a.currency,
    SUM(
        CASE WHEN a.type = 1 THEN a.current_balance * c.rate
             WHEN a.type = 4 THEN -a.current_balance * c.rate
             ELSE 0
        END
    ) as exposure_in_base,
    ROUND(
        SUM(
            CASE WHEN a.type = 1 THEN a.current_balance * c.rate
                 WHEN a.type = 4 THEN -a.current_balance * c.rate
                 ELSE 0
            END
        ) * 100.0 / NULLIF((
            SELECT SUM(
                CASE WHEN type = 1 THEN current_balance * cr.rate
                     WHEN type = 4 THEN -current_balance * cr.rate
                     ELSE 0
                END
            )
            FROM accounts
            JOIN currencies cr ON cr.id = currency
            WHERE type IN (1, 4) AND deleted_at IS NULL
        ), 0),
    2) as percentage
FROM accounts a
JOIN currencies c ON c.id = a.currency
WHERE a.type IN (1, 4)
  AND a.deleted_at IS NULL
GROUP BY a.currency
ORDER BY exposure_in_base DESC;
```

## Notes

- Currency codes follow ISO 4217 standard (3-letter codes)
- Rate is relative to base currency (base currency has rate=1)
- `decimal_places` determines precision for display and calculations
- Rates should be updated regularly for accurate conversions
- Transaction amounts are stored in original currency AND converted to base


---

## See Also

- [Currency Conversion](../../business-logic/currencies/conversion.md) - Conversion formulas
- [Amount Calculations](../../business-logic/transactions/amount-calculations.md) - Multi-currency transactions
- [Schema Quick-Ref](../QUICK-REF.md) - All tables at a glance
