# MCP Golden Rules

> **CRITICAL**: Follow these rules to minimize token usage and maximize efficiency.

## Rule 1: Use Aggregates, Not Row-by-Row

```sql
-- GOOD: Single query with aggregate
SELECT SUM(destination_amount_in_base_currency) as total
FROM transactions WHERE transaction_type = 3 AND deleted_at IS NULL;

-- BAD: Fetching all rows then summing in code
SELECT * FROM transactions WHERE transaction_type = 3;
```

## Rule 2: Use Pre-computed Tables

| Need | Use This | NOT This |
|------|----------|----------|
| Current balance | `accounts.current_balance` | SUM of transactions |
| Balance history | `daily_stat.amount` | SUM of transactions by date |
| Net worth | `accounts WHERE type IN (1,4)` | Complex transaction sums |

## Rule 3: Always Include Base Filters

```sql
-- ALWAYS add this filter:
WHERE deleted_at IS NULL              -- Required for ALL queries
```

## Rule 4: Select Only Needed Columns

```sql
-- GOOD
SELECT id, title, destination_amount_in_base_currency FROM transactions;

-- BAD
SELECT * FROM transactions;
```

## Rule 5: Single Query for Multiple Metrics

```sql
-- GOOD: One query for income AND expenses
SELECT
    SUM(CASE WHEN transaction_type = 2 THEN destination_amount_in_base_currency END) as income,
    SUM(CASE WHEN transaction_type = 3 THEN destination_amount_in_base_currency END) as expenses
FROM transactions WHERE deleted_at IS NULL;

-- BAD: Two separate queries
```

## Rule 6: Use GROUP BY for Breakdowns

```sql
-- GOOD: Category totals in one query
SELECT category_id, SUM(destination_amount_in_base_currency) as total
FROM transactions
WHERE transaction_type = 3 AND deleted_at IS NULL
GROUP BY category_id;

-- BAD: Query per category
```

## Rule 7: NEVER Guess Values - Always Lookup First

**CRITICAL**: Never use hardcoded or guessed values for categories, tags, or accounts.
Always query the lookup tables first to get real IDs and names.

```sql
-- GOOD: First fetch actual categories
SELECT id, name FROM categories WHERE deleted_at IS NULL;
-- Then use real IDs: WHERE category_id = 5

-- GOOD: First fetch actual tags
SELECT id, name FROM tags WHERE deleted_at IS NULL;
-- Then use real IDs: WHERE 3 = ANY(tag_ids)

-- GOOD: First fetch actual accounts
SELECT id, name, type FROM accounts WHERE deleted_at IS NULL;
-- Then use real IDs: WHERE source_account_id = 12

-- BAD: Guessing with LIKE patterns
WHERE title LIKE '%netflix%'           -- Don't guess!
WHERE category_id = 999                -- Made up ID!
```

### Categorization/Tagging Workflow

When asked to analyze or filter by category/tag/account:

1. **First**: Query lookup table to show available options
2. **Then**: Use actual IDs from step 1 in your analysis query

```sql
-- Step 1: Show categories
SELECT id, name FROM categories WHERE deleted_at IS NULL;
-- Returns: (1, 'Food'), (2, 'Transport'), (3, 'Entertainment')

-- Step 2: Use real ID for analysis
SELECT SUM(destination_amount_in_base_currency)
FROM transactions
WHERE category_id = 3 AND transaction_type = 3 AND deleted_at IS NULL;
```

## Rule 8: LIMIT Usage

```sql
-- For LOOKUPS (categories, tags, accounts): No LIMIT needed, tables are small
SELECT id, name FROM categories WHERE deleted_at IS NULL;

-- For LISTINGS (showing transactions): Use LIMIT
SELECT * FROM transactions WHERE deleted_at IS NULL ORDER BY transaction_date_time DESC LIMIT 20;

-- For ANALYTICS (aggregates): NO LIMIT - aggregates return single/few rows
SELECT SUM(...) FROM transactions WHERE ...;
SELECT category_id, SUM(...) FROM transactions GROUP BY category_id;
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

### Lookup Tables (Query These First!)
| Table | Use For |
|-------|---------|
| `categories` | Get real category IDs and names |
| `tags` | Get real tag IDs and names |
| `accounts` | Get real account IDs, names, types |

### Data Tables
| Table | Use For |
|-------|---------|
| `transactions` | Transaction details, spending/income |
| `daily_stat` | Balance history, trends |

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

-- Top categories (with names)
SELECT c.name, SUM(t.destination_amount_in_base_currency) as total
FROM transactions t JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3 AND t.deleted_at IS NULL
GROUP BY c.name ORDER BY total DESC;

-- Transactions by tag (first lookup tag id!)
SELECT t.* FROM transactions t
WHERE 5 = ANY(t.tag_ids) AND t.deleted_at IS NULL;
```
