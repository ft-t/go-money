# MCP Query Examples

Natural language questions mapped to SQL queries for the Go Money MCP server.

## Account Queries

### "What accounts do I have?"

```sql
SELECT id, name, type, current_balance, currency
FROM accounts
WHERE deleted_at IS NULL
ORDER BY type, display_order NULLS LAST;
```

### "What's my checking account balance?"

```sql
SELECT name, current_balance, currency
FROM accounts
WHERE name ILIKE '%checking%'
  AND deleted_at IS NULL;
```

### "Show my credit cards"

```sql
SELECT id, name, current_balance, currency
FROM accounts
WHERE type = 4  -- Liability
  AND deleted_at IS NULL
ORDER BY current_balance DESC;
```

## Net Worth Queries

### "What's my net worth?"

```sql
SELECT
    SUM(CASE WHEN type = 1 THEN current_balance ELSE 0 END) as assets,
    SUM(CASE WHEN type = 4 THEN current_balance ELSE 0 END) as liabilities,
    SUM(CASE
        WHEN type = 1 THEN current_balance
        WHEN type = 4 THEN -current_balance
        ELSE 0
    END) as net_worth
FROM accounts
WHERE type IN (1, 4)
  AND deleted_at IS NULL;
```

### "How has my net worth changed this year?"

```sql
WITH monthly_net_worth AS (
    SELECT DISTINCT ON (DATE_TRUNC('month', ds.date))
        DATE_TRUNC('month', ds.date) as month,
        SUM(CASE WHEN a.type = 1 THEN ds.amount ELSE -ds.amount END) as net_worth
    FROM daily_stat ds
    JOIN accounts a ON a.id = ds.account_id
    WHERE a.type IN (1, 4)
      AND a.deleted_at IS NULL
      AND ds.date >= DATE_TRUNC('year', CURRENT_DATE)
    GROUP BY DATE_TRUNC('month', ds.date), ds.date
    ORDER BY DATE_TRUNC('month', ds.date), ds.date DESC
)
SELECT month, net_worth
FROM monthly_net_worth
ORDER BY month;
```

## Spending Queries

### "How much did I spend this month?"

```sql
SELECT SUM(destination_amount_in_base_currency) as total_spending
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE);
```

### "How much did I spend last month vs this month?"

```sql
SELECT
    SUM(CASE
        WHEN DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE)
        THEN destination_amount_in_base_currency ELSE 0
    END) as this_month,
    SUM(CASE
        WHEN DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')
        THEN destination_amount_in_base_currency ELSE 0
    END) as last_month
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND transaction_date_only >= DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month');
```

### "What's my daily spending average?"

```sql
SELECT
    ROUND(AVG(daily_total), 2) as avg_daily_spending
FROM (
    SELECT transaction_date_only, SUM(destination_amount_in_base_currency) as daily_total
    FROM transactions
    WHERE transaction_type = 3
      AND deleted_at IS NULL
      AND transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'
    GROUP BY transaction_date_only
) daily;
```

## Category Queries

### "What are my top expense categories?"

```sql
SELECT
    c.name as category,
    SUM(t.destination_amount_in_base_currency) as total,
    COUNT(*) as transaction_count
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= DATE_TRUNC('month', CURRENT_DATE)
GROUP BY c.name
ORDER BY total DESC
LIMIT 10;
```

### "How much did I spend on groceries?"

```sql
SELECT
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE c.name ILIKE '%grocer%'
  AND t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= DATE_TRUNC('month', CURRENT_DATE);
```

### "Show category breakdown for last month"

```sql
SELECT
    COALESCE(c.name, 'Uncategorized') as category,
    SUM(t.destination_amount_in_base_currency) as total,
    ROUND(SUM(t.destination_amount_in_base_currency) * 100.0 /
          SUM(SUM(t.destination_amount_in_base_currency)) OVER (), 2) as percentage
FROM transactions t
LEFT JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND DATE_TRUNC('month', t.transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE - INTERVAL '1 month')
GROUP BY c.name
ORDER BY total DESC;
```

## Income Queries

### "How much income did I receive this month?"

```sql
SELECT SUM(destination_amount_in_base_currency) as total_income
FROM transactions
WHERE transaction_type = 2
  AND deleted_at IS NULL
  AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE);
```

### "What are my income sources?"

```sql
SELECT
    a.name as source,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN accounts a ON a.id = t.source_account_id
WHERE t.transaction_type = 2
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= DATE_TRUNC('year', CURRENT_DATE)
GROUP BY a.name
ORDER BY total DESC;
```

### "Am I saving money? (Income vs Expenses)"

```sql
SELECT
    SUM(CASE WHEN transaction_type = 2 THEN destination_amount_in_base_currency ELSE 0 END) as income,
    SUM(CASE WHEN transaction_type = 3 THEN destination_amount_in_base_currency ELSE 0 END) as expenses,
    SUM(CASE WHEN transaction_type = 2 THEN destination_amount_in_base_currency ELSE 0 END) -
    SUM(CASE WHEN transaction_type = 3 THEN destination_amount_in_base_currency ELSE 0 END) as savings
FROM transactions
WHERE transaction_type IN (2, 3)
  AND deleted_at IS NULL
  AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE);
```

## Transaction Queries

### "Show my recent transactions"

```sql
SELECT
    t.title,
    t.transaction_date_only as date,
    t.destination_amount,
    t.destination_currency,
    c.name as category,
    CASE t.transaction_type
        WHEN 1 THEN 'Transfer'
        WHEN 2 THEN 'Income'
        WHEN 3 THEN 'Expense'
        WHEN 5 THEN 'Adjustment'
    END as type
FROM transactions t
LEFT JOIN categories c ON c.id = t.category_id
WHERE t.deleted_at IS NULL
ORDER BY t.transaction_date_time DESC
LIMIT 20;
```

### "Find transactions containing 'amazon'"

```sql
SELECT
    title,
    transaction_date_only,
    destination_amount,
    destination_currency
FROM transactions
WHERE title ILIKE '%amazon%'
  AND deleted_at IS NULL
ORDER BY transaction_date_time DESC
LIMIT 20;
```

### "What were my largest purchases this month?"

```sql
SELECT
    title,
    transaction_date_only,
    destination_amount_in_base_currency as amount
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE)
ORDER BY destination_amount_in_base_currency DESC
LIMIT 10;
```

## Tag Queries

### "Show spending by tag"

```sql
SELECT
    tg.name as tag,
    COUNT(*) as count,
    SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN tags tg ON tg.id = ANY(t.tag_ids)
WHERE t.transaction_type = 3
  AND t.deleted_at IS NULL
  AND t.transaction_date_only >= DATE_TRUNC('month', CURRENT_DATE)
GROUP BY tg.name
ORDER BY total DESC;
```

### "Find transactions tagged 'vacation'"

```sql
SELECT
    t.title,
    t.transaction_date_only,
    t.destination_amount,
    t.destination_currency
FROM transactions t
JOIN tags tg ON tg.id = ANY(t.tag_ids)
WHERE tg.name ILIKE '%vacation%'
  AND t.deleted_at IS NULL
ORDER BY t.transaction_date_time DESC;
```

### "How many transactions are untagged?"

```sql
SELECT COUNT(*) as untagged_count
FROM transactions
WHERE (tag_ids IS NULL OR ARRAY_LENGTH(tag_ids, 1) IS NULL)
  AND transaction_type = 3
  AND deleted_at IS NULL;
```

## Time Series Queries

### "Show my daily spending for the past week"

```sql
SELECT
    transaction_date_only as date,
    SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND transaction_date_only >= CURRENT_DATE - INTERVAL '7 days'
GROUP BY transaction_date_only
ORDER BY date;
```

### "What's my monthly spending trend?"

```sql
SELECT
    DATE_TRUNC('month', transaction_date_only) as month,
    SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND transaction_date_only >= CURRENT_DATE - INTERVAL '12 months'
GROUP BY DATE_TRUNC('month', transaction_date_only)
ORDER BY month;
```

### "Compare this year to last year"

```sql
SELECT
    SUM(CASE WHEN EXTRACT(YEAR FROM transaction_date_only) = EXTRACT(YEAR FROM CURRENT_DATE)
        THEN destination_amount_in_base_currency ELSE 0 END) as this_year,
    SUM(CASE WHEN EXTRACT(YEAR FROM transaction_date_only) = EXTRACT(YEAR FROM CURRENT_DATE) - 1
        THEN destination_amount_in_base_currency ELSE 0 END) as last_year
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL;
```

## Balance History Queries

### "Show my account balance history"

```sql
SELECT date, amount as balance
FROM daily_stat
WHERE account_id = 1  -- Replace with account ID
  AND date >= CURRENT_DATE - INTERVAL '30 days'
ORDER BY date;
```

### "What was my balance on a specific date?"

```sql
SELECT amount as balance
FROM daily_stat
WHERE account_id = 1  -- Replace with account ID
  AND date <= '2024-01-15'  -- Replace with target date
ORDER BY date DESC
LIMIT 1;
```

## Data Quality Queries

### "Find uncategorized expenses"

```sql
SELECT
    title,
    transaction_date_only,
    destination_amount_in_base_currency
FROM transactions
WHERE transaction_type = 3
  AND category_id IS NULL
  AND deleted_at IS NULL
ORDER BY destination_amount_in_base_currency DESC
LIMIT 20;
```

### "Find potential duplicate transactions"

```sql
SELECT
    title,
    transaction_date_only,
    destination_amount,
    COUNT(*) as occurrences
FROM transactions
WHERE deleted_at IS NULL
GROUP BY title, transaction_date_only, destination_amount
HAVING COUNT(*) > 1
ORDER BY occurrences DESC
LIMIT 20;
```

## Currency Queries

### "What currencies do I hold?"

```sql
SELECT
    currency,
    SUM(current_balance) as total_balance,
    COUNT(*) as account_count
FROM accounts
WHERE type = 1  -- Assets only
  AND deleted_at IS NULL
GROUP BY currency
ORDER BY total_balance DESC;
```

### "Show exchange rates"

```sql
SELECT id as currency, rate, decimal_places
FROM currencies
WHERE is_active = true
  AND deleted_at IS NULL
ORDER BY id;
```

## Ledger Queries

### "Show the double-entry ledger for an account"

```sql
SELECT
    de.transaction_date,
    t.title,
    CASE WHEN de.is_debit THEN de.amount_in_base_currency ELSE NULL END as debit,
    CASE WHEN NOT de.is_debit THEN de.amount_in_base_currency ELSE NULL END as credit
FROM double_entries de
JOIN transactions t ON t.id = de.transaction_id
WHERE de.account_id = 1  -- Replace with account ID
  AND de.deleted_at IS NULL
ORDER BY de.transaction_date DESC
LIMIT 50;
```

### "Verify the ledger is balanced"

```sql
SELECT
    SUM(CASE WHEN is_debit THEN amount_in_base_currency ELSE 0 END) as total_debits,
    SUM(CASE WHEN NOT is_debit THEN amount_in_base_currency ELSE 0 END) as total_credits
FROM double_entries
WHERE deleted_at IS NULL;
-- These two values should be equal
```
