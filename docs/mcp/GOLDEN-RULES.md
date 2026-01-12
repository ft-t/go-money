# MCP Golden Rules

> **CRITICAL**: Follow these rules to minimize token usage and maximize efficiency.

## Rule 1: Use Aggregates, Not Row-by-Row

```sql
-- ✅ GOOD: Single query with aggregate
SELECT SUM(destination_amount_in_base_currency) as total
FROM transactions WHERE transaction_type = 3 AND deleted_at IS NULL;

-- ❌ BAD: Fetching all rows then summing in code
SELECT * FROM transactions WHERE transaction_type = 3;
```

## Rule 2: Use Pre-computed Tables

| Need | Use This | NOT This |
|------|----------|----------|
| Current balance | `accounts.current_balance` | SUM of transactions |
| Balance history | `daily_stat.amount` | SUM of transactions by date |
| Net worth | `accounts WHERE type IN (1,4)` | Complex transaction sums |

## Rule 3: Always Include Filters

```sql
-- ALWAYS add these:
WHERE deleted_at IS NULL              -- Required for all queries
  AND transaction_date_only >= ...    -- Limit date range
LIMIT 100                             -- Cap results
```

## Rule 4: Select Only Needed Columns

```sql
-- ✅ GOOD
SELECT id, title, destination_amount_in_base_currency FROM transactions;

-- ❌ BAD
SELECT * FROM transactions;
```

## Rule 5: Single Query for Multiple Metrics

```sql
-- ✅ GOOD: One query for income AND expenses
SELECT
    SUM(CASE WHEN transaction_type = 2 THEN destination_amount_in_base_currency END) as income,
    SUM(CASE WHEN transaction_type = 3 THEN destination_amount_in_base_currency END) as expenses
FROM transactions WHERE deleted_at IS NULL;

-- ❌ BAD: Two separate queries
```

## Rule 6: Use GROUP BY for Breakdowns

```sql
-- ✅ GOOD: Category totals in one query
SELECT category_id, SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3 AND deleted_at IS NULL
GROUP BY category_id;

-- ❌ BAD: Query per category
```

## Rule 7: Limit Results Aggressively

```sql
-- For listings: LIMIT 10-20
-- For analytics: Use aggregates (no LIMIT needed)
-- For exploration: LIMIT 5 first, expand if needed
```

## Quick Reference

### Transaction Types
| Value | Type |
|-------|------|
| 1 | Transfer |
| 2 | Income |
| 3 | Expense |
| 5 | Adjustment |

### Account Types
| Value | Type |
|-------|------|
| 1 | Asset |
| 4 | Liability |
| 5 | Expense |
| 6 | Income |

### Essential Tables
| Table | Use For |
|-------|---------|
| `accounts` | Current balances, account list |
| `transactions` | Transaction details, spending/income |
| `daily_stat` | Balance history, trends |
| `categories` | Category names (JOIN) |
| `tags` | Tag names (JOIN with ANY) |

### Common Patterns

```sql
-- Net worth
SELECT SUM(CASE WHEN type=1 THEN current_balance WHEN type=4 THEN -current_balance END)
FROM accounts WHERE type IN (1,4) AND deleted_at IS NULL;

-- Monthly spending
SELECT SUM(destination_amount_in_base_currency)
FROM transactions
WHERE transaction_type = 3 AND deleted_at IS NULL
  AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE);

-- Top categories
SELECT c.name, SUM(t.destination_amount_in_base_currency) as total
FROM transactions t JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3 AND t.deleted_at IS NULL
GROUP BY c.name ORDER BY total DESC LIMIT 10;
```
