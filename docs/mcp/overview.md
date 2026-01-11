# MCP Server Overview

This document describes how to use the Go Money PostgreSQL MCP server for AI-assisted financial analytics.

## Purpose

The MCP (Model Context Protocol) server provides read-only access to the Go Money PostgreSQL database, enabling AI agents to:
- Query financial data using natural language
- Generate analytics reports
- Answer questions about spending, income, and net worth
- Create data visualizations

## Available Tool

### query

Execute read-only SQL queries against the database.

**Parameters:**
- `sql` (string, required): The SQL query to execute

**Example:**
```json
{
  "sql": "SELECT * FROM accounts WHERE deleted_at IS NULL"
}
```

## Query Safety

The MCP server enforces read-only access:
- Only SELECT statements are allowed
- INSERT, UPDATE, DELETE, DROP, etc. are blocked
- Queries have timeout limits
- Large result sets may be truncated

## Database Context for AI Agents

When generating queries, consider:

### Key Tables

| Table | Purpose | Key Fields |
|-------|---------|------------|
| `accounts` | All financial accounts | id, name, current_balance, type, currency |
| `transactions` | All transactions | id, title, transaction_type, amounts, dates |
| `categories` | Expense categories | id, name |
| `tags` | Transaction tags | id, name |
| `currencies` | Currency rates | id, rate, is_active |
| `daily_stat` | Pre-computed daily changes | account_id, date, amount |
| `double_entries` | Double-entry ledger | transaction_id, account_id, is_debit, amount |

### Essential Filters

Always include these filters for accurate data:

```sql
-- Active records only
WHERE deleted_at IS NULL

-- Specific transaction types
WHERE transaction_type = 3  -- Expenses
WHERE transaction_type = 2  -- Income
WHERE transaction_type = 1  -- Transfers

-- Specific account types
WHERE type = 1  -- Asset
WHERE type = 4  -- Liability
WHERE type = 5  -- Expense
WHERE type = 6  -- Income
```

### Common Patterns

**Date ranges:**
```sql
-- This month
WHERE DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE)

-- Last 30 days
WHERE transaction_date_only >= CURRENT_DATE - INTERVAL '30 days'

-- This year
WHERE transaction_date_only >= DATE_TRUNC('year', CURRENT_DATE)
```

**Amounts:**
```sql
-- Use base currency amounts for comparisons
SELECT destination_amount_in_base_currency as amount ...

-- Original currency amounts
SELECT destination_amount, destination_currency ...
```

**Tags (array field):**
```sql
-- Has specific tag
WHERE :tag_id = ANY(tag_ids)

-- Has all tags
WHERE tag_ids @> ARRAY[1, 2]::integer[]

-- Has any of tags
WHERE tag_ids && ARRAY[1, 2]::integer[]
```

## Example Queries for Common Questions

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
WHERE type IN (1, 4) AND deleted_at IS NULL;
```

### "How much did I spend this month?"

```sql
SELECT SUM(destination_amount_in_base_currency) as total_spending
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE);
```

### "What are my top expense categories?"

```sql
SELECT c.name, SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE t.transaction_type = 3 AND t.deleted_at IS NULL
GROUP BY c.name
ORDER BY total DESC
LIMIT 10;
```

### "Show my recent transactions"

```sql
SELECT
    t.title,
    t.transaction_date_only,
    t.destination_amount,
    t.destination_currency,
    c.name as category
FROM transactions t
LEFT JOIN categories c ON c.id = t.category_id
WHERE t.deleted_at IS NULL
ORDER BY t.transaction_date_time DESC
LIMIT 20;
```

### "How much did I spend on [category]?"

```sql
SELECT SUM(t.destination_amount_in_base_currency) as total
FROM transactions t
JOIN categories c ON c.id = t.category_id
WHERE c.name ILIKE '%groceries%'  -- Replace with category
  AND t.transaction_type = 3
  AND t.deleted_at IS NULL;
```

## Performance Tips

1. **Use daily_stat for balance calculations** - faster than scanning transactions
2. **Limit results** - add LIMIT for exploratory queries
3. **Use indexed columns** - filter by transaction_date_only, account_id, transaction_type
4. **Avoid full table scans** - always include WHERE conditions

## Response Guidelines

When presenting query results:
1. Format currency amounts appropriately
2. Include currency codes when relevant
3. Round percentages to 2 decimal places
4. Present dates in user-friendly format
5. Explain any calculations or assumptions

## Security Notes

- Database credentials are managed by the MCP server
- All queries are logged for audit purposes
- Sensitive data (passwords) is never returned in queries
- Read-only access prevents data modification
