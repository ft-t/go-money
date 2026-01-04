# Go Money Documentation Index

> **For AI Agents:** Start here. This index maps all documentation with keywords for efficient navigation.

## Quick References (Read These First)

| File | Description |
|------|-------------|
| [Schema Quick-Ref](schema/QUICK-REF.md) | All tables, columns, types, enums in one file |
| [Business Logic Quick-Ref](business-logic/QUICK-REF.md) | Formulas, transaction rules, account rules |
| [Analytics Quick-Ref](analytics/QUICK-REF.md) | Query templates with descriptions |

---

## By Use Case

### "I need to write a SQL query"
| Document | Keywords |
|----------|----------|
| [Analytics Overview](analytics/overview.md) | query basics, table selection, data sources |
| [Common Queries](analytics/common-queries.md) | 25+ ready SQL examples, spending, income, net worth |
| [MCP Examples](mcp/examples.md) | natural language to SQL, 30+ mappings |
| [Performance Tips](analytics/performance-tips.md) | indexes, optimization, daily_stat vs transactions |

### "I need to understand the database schema"
| Document | Keywords |
|----------|----------|
| [Schema Overview](schema/overview.md) | ER diagram, table relationships, foreign keys |
| [Enums Reference](schema/enums.md) | TransactionType, AccountType, AccountFlags, all enum values |
| [Indexes](schema/indexes.md) | database indexes, query optimization |

### "I need details about a specific table"
| Table | Document | Key Fields |
|-------|----------|------------|
| accounts | [accounts.md](schema/tables/accounts.md) | id, name, type, currency, current_balance |
| transactions | [transactions.md](schema/tables/transactions.md) | source/destination amounts, dates, category_id, tag_ids |
| categories | [categories.md](schema/tables/categories.md) | id, name |
| tags | [tags.md](schema/tables/tags.md) | id, name, tag_ids array |
| currencies | [currencies.md](schema/tables/currencies.md) | id, rate, decimal_places |
| daily_stat | [stats.md](schema/tables/stats.md) | account_id, date, amount (running balance) |
| double_entries | [double_entry.md](schema/tables/double_entry.md) | is_debit, amount, ledger |
| rules | [rules.md](schema/tables/rules.md) | Lua scripts, sort_order, group |
| users | [users.md](schema/tables/users.md) | login, password (bcrypt) |

### "I need to understand how transactions work"
| Document | Keywords |
|----------|----------|
| [Transaction Overview](business-logic/transactions/overview.md) | creation flow, processing pipeline |
| [Transaction Types](business-logic/transactions/types.md) | expense, income, transfer, adjustment behavior |
| [Amount Calculations](business-logic/transactions/amount-calculations.md) | base currency conversion, FX, formulas |
| [Double-Entry](business-logic/double-entry/overview.md) | debit/credit rules, ledger entries |

### "I need to understand accounts"
| Document | Keywords |
|----------|----------|
| [Account Types](business-logic/accounts/types.md) | asset, liability, expense, income, which accounts for which tx |
| [Balance Tracking](business-logic/accounts/balance-tracking.md) | current_balance updates, daily_stat recalculation |

### "I need to understand the API"
| Document | Keywords |
|----------|----------|
| [Authentication](api/authentication.md) | JWT, RS256, service tokens, login flow |
| [Endpoints](api/endpoints.md) | all endpoints, request/response schemas, Connect protocol |

### "I need to understand the MCP server"
| Document | Keywords |
|----------|----------|
| [MCP Overview](mcp/overview.md) | read-only queries, AI integration |
| [Tool Reference](mcp/tool-reference.md) | query tool spec, parameters, output format |
| [Query Safety](mcp/query-safety.md) | blocked statements, limits, allowed tables |
| [Examples](mcp/examples.md) | natural language → SQL mappings |

---

## Query Patterns by Topic

| Topic | Document | Example Query |
|-------|----------|---------------|
| Net Worth | [balance-queries.md](analytics/query-patterns/balance-queries.md) | `SUM(CASE WHEN type=1 THEN balance...)` |
| Monthly Spending | [transaction-queries.md](analytics/query-patterns/transaction-queries.md) | `WHERE transaction_type=3 AND DATE_TRUNC...` |
| Category Breakdown | [category-analysis.md](analytics/query-patterns/category-analysis.md) | `GROUP BY category_id` |
| Tag Filtering | [tag-analysis.md](analytics/query-patterns/tag-analysis.md) | `WHERE tag_id = ANY(tag_ids)` |
| Trends Over Time | [time-series.md](analytics/query-patterns/time-series.md) | `DATE_TRUNC, LAG, rolling averages` |

---

## Key Constants

### Transaction Types
| Value | Name | Use |
|-------|------|-----|
| 1 | Transfer | Between asset/liability accounts |
| 2 | Income | Money received |
| 3 | Expense | Money spent |
| 5 | Adjustment | Balance correction |

### Account Types
| Value | Name | Use |
|-------|------|-----|
| 1 | Asset | Bank accounts, cash, investments |
| 4 | Liability | Credit cards, loans |
| 5 | Expense | Spending categories |
| 6 | Income | Revenue sources |
| 7 | Adjustment | Balance adjustments |

### Essential WHERE Clauses
```sql
WHERE deleted_at IS NULL           -- Active records only
WHERE transaction_type = 3         -- Expenses only
WHERE type IN (1, 4)               -- Assets and liabilities (for net worth)
```

---

## Code Examples

| File | Description |
|------|-------------|
| [lua/set_category_by_title.lua](lua/set_category_by_title.lua) | Auto-categorize by transaction title |
| [lua/convert_from_withdrawal_to_transfer.lua](lua/convert_from_withdrawal_to_transfer.lua) | Convert withdrawal to transfer |
| [lua/scheduled_withdrawal.lua](lua/scheduled_withdrawal.lua) | Scheduled recurring transaction |
| [reporting/queries/withdrawals_by_tag.sql](reporting/queries/withdrawals_by_tag.sql) | SQL report example |

---

## Utility Documentation

| File | Description |
|------|-------------|
| [debug.md](debug.md) | Frontend debugging (custom API host) |

---

## Cross-Reference Map

| If you're reading... | Also see... |
|---------------------|-------------|
| schema/tables/transactions.md | business-logic/transactions/types.md |
| schema/tables/accounts.md | business-logic/accounts/types.md |
| schema/tables/double_entry.md | business-logic/double-entry/overview.md |
| schema/tables/stats.md | business-logic/statistics/daily-stats.md |
| analytics/common-queries.md | mcp/examples.md |
| api/authentication.md | api/endpoints.md |
