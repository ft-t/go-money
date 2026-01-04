# Go Money Documentation

Comprehensive documentation for the Go Money personal finance system.

## Quick Links

### Schema Reference

- [Schema Overview](schema/overview.md) - ER diagram and table summary
- [Enum Reference](schema/enums.md) - All enum values and meanings
- [Index Reference](schema/indexes.md) - Database indexes for optimization

### Table Documentation

| Table | Description |
|-------|-------------|
| [accounts](schema/tables/accounts.md) | Bank accounts, credit cards, categories |
| [transactions](schema/tables/transactions.md) | All financial transactions |
| [categories](schema/tables/categories.md) | Transaction categories |
| [tags](schema/tables/tags.md) | Flexible transaction tags |
| [currencies](schema/tables/currencies.md) | Currency definitions and rates |
| [rules](schema/tables/rules.md) | Transaction processing rules |
| [double_entry](schema/tables/double_entry.md) | Double-entry bookkeeping ledger |
| [stats](schema/tables/stats.md) | Pre-computed daily statistics |
| [users](schema/tables/users.md) | User authentication |
| [import_deduplication](schema/tables/import_deduplication.md) | Import duplicate detection |

### Analytics

- [Analytics Overview](analytics/overview.md) - Data sources and concepts
- [Common Queries](analytics/common-queries.md) - Ready-to-use SQL queries
- [Performance Tips](analytics/performance-tips.md) - Query optimization

#### Query Patterns
- [Balance Queries](analytics/query-patterns/balance-queries.md) - Net worth, account balances
- [Transaction Queries](analytics/query-patterns/transaction-queries.md) - Filtering, pagination
- [Category Analysis](analytics/query-patterns/category-analysis.md) - Spending by category
- [Tag Analysis](analytics/query-patterns/tag-analysis.md) - Tag-based queries
- [Time Series](analytics/query-patterns/time-series.md) - Trends and comparisons

### MCP Server

- [MCP Overview](mcp/overview.md) - AI-assisted database queries
- [Tool Reference](mcp/tool-reference.md) - Query tool specification
- [Query Safety](mcp/query-safety.md) - Security and validation rules
- [Query Examples](mcp/examples.md) - Natural language to SQL mappings

### API Reference

- [Authentication](api/authentication.md) - JWT tokens and service tokens
- [Endpoints](api/endpoints.md) - Complete API reference

## Core Concepts

### Account Types

| Type | Value | Purpose |
|------|-------|---------|
| Asset | 1 | Cash, bank accounts, investments |
| Liability | 4 | Credit cards, loans, debts |
| Expense | 5 | Spending categories |
| Income | 6 | Revenue sources |
| Adjustment | 7 | Balance adjustments |

### Transaction Types

| Type | Value | Description |
|------|-------|-------------|
| Transfer | 1 | Between Asset/Liability accounts |
| Income | 2 | Money received |
| Expense | 3 | Money spent |
| Adjustment | 5 | Balance correction |

### Multi-Currency

- Each account has a currency
- Transactions store amounts in original and base currency
- `currencies.rate` converts to base currency: `base_amount = amount * rate`

### Double-Entry Bookkeeping

Every transaction creates:
- One debit entry (is_debit = true)
- One credit entry (is_debit = false)
- Total debits always equal total credits

### Soft Deletes

Most tables use soft deletes:
- `deleted_at IS NULL` = active record
- `deleted_at IS NOT NULL` = deleted record
- Always filter with `WHERE deleted_at IS NULL`

## Common Query Patterns

### Active Records

```sql
SELECT * FROM accounts WHERE deleted_at IS NULL;
```

### This Month's Expenses

```sql
SELECT SUM(destination_amount_in_base_currency)
FROM transactions
WHERE transaction_type = 3
  AND deleted_at IS NULL
  AND DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE);
```

### Net Worth

```sql
SELECT SUM(CASE
    WHEN type = 1 THEN current_balance
    WHEN type = 4 THEN -current_balance
    ELSE 0
END) as net_worth
FROM accounts
WHERE type IN (1, 4) AND deleted_at IS NULL;
```

### Transactions by Tag

```sql
SELECT * FROM transactions
WHERE :tag_id = ANY(tag_ids)
  AND deleted_at IS NULL;
```

## Getting Started

1. Start with [Schema Overview](schema/overview.md) for the big picture
2. Review [Enum Reference](schema/enums.md) to understand field values
3. Explore [Common Queries](analytics/common-queries.md) for analytics patterns
4. Use [MCP Overview](mcp/overview.md) for AI-assisted queries
