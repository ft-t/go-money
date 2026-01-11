# Go Money Database Schema Overview

This document provides an overview of the Go Money database schema, designed for personal finance tracking with double-entry bookkeeping support.

## Entity Relationship Diagram

```mermaid
erDiagram
    accounts ||--o{ transactions : "source_account_id"
    accounts ||--o{ transactions : "destination_account_id"
    accounts ||--o{ double_entries : "account_id"
    accounts ||--o{ daily_stat : "account_id"
    categories ||--o{ transactions : "category_id"
    transactions ||--o{ double_entries : "transaction_id"
    transactions }o--o{ tags : "tag_ids (array)"
    currencies ||--o{ accounts : "currency"
    transactions ||--o{ import_deduplication : "transaction_id"

    accounts {
        int id PK
        text name
        numeric current_balance
        text currency FK
        int type
        bigint flags
        jsonb extra
        text note
        text account_number
        text iban
        numeric liability_percent
        int display_order
        timestamp first_transaction_at
        timestamp last_updated_at
        timestamp created_at
        timestamp deleted_at
    }

    transactions {
        bigint id PK
        numeric source_amount
        text source_currency
        numeric source_amount_in_base_currency
        numeric destination_amount
        text destination_currency
        numeric destination_amount_in_base_currency
        numeric fx_source_amount
        text fx_source_currency
        int source_account_id FK
        int destination_account_id FK
        int category_id FK
        int[] tag_ids
        int transaction_type
        bigint flags
        text title
        text notes
        text reference_number
        text[] internal_reference_numbers
        jsonb extra
        timestamp transaction_date_time
        date transaction_date_only
        bigint voided_by_transaction_id
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    double_entries {
        bigint id PK
        bigint transaction_id FK
        int account_id FK
        boolean is_debit
        numeric amount_in_base_currency
        text base_currency
        timestamp transaction_date
        timestamp created_at
        timestamp deleted_at
    }

    categories {
        int id PK
        text name
        timestamp created_at
        timestamp deleted_at
    }

    tags {
        int id PK
        text name
        text color
        text icon
        timestamp created_at
        timestamp deleted_at
    }

    currencies {
        text id PK
        numeric rate
        boolean is_active
        int decimal_places
        timestamp updated_at
        timestamp deleted_at
    }

    daily_stat {
        int account_id PK_FK
        date date PK
        numeric amount
    }

    rules {
        int id PK
        text title
        text script
        int interpreter_type
        int sort_order
        boolean enabled
        boolean is_final_rule
        text group_name
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    schedule_rules {
        int id PK
        text title
        text script
        int interpreter_type
        text cron_expression
        boolean enabled
        text group_name
        timestamp last_run_at
        timestamp created_at
        timestamp updated_at
        timestamp deleted_at
    }

    users {
        int id PK
        text login
        text password
        timestamp created_at
        timestamp deleted_at
    }

    import_deduplication {
        int import_source PK
        text key PK
        bigint transaction_id FK
        timestamp created_at
    }
```

## Table Summary

| Table | Purpose | Key Features |
|-------|---------|--------------|
| `accounts` | Stores all financial accounts | Supports Asset, Liability, Expense, Income, and Adjustment types |
| `transactions` | Records all financial transactions | Multi-currency support, links source and destination accounts |
| `double_entries` | Double-entry bookkeeping ledger | Every transaction creates debit and credit entries |
| `categories` | Transaction categorization | Simple hierarchy with soft delete |
| `tags` | Flexible transaction tagging | Many-to-many via array field in transactions |
| `currencies` | Currency definitions and rates | ISO 4217 codes, exchange rates relative to base currency |
| `daily_stat` | Daily account balance changes | Pre-computed for fast analytics |
| `rules` | Transaction processing rules | Lua scripts executed on transaction creation |
| `schedule_rules` | Scheduled automation rules | Cron-based Lua script execution |
| `users` | User accounts | Authentication credentials |
| `import_deduplication` | Import duplicate detection | Prevents re-importing the same transactions |

## Core Concepts

### Account Types

Accounts are classified into five types following double-entry bookkeeping principles:

| Type Value | Name | Purpose | Normal Balance |
|------------|------|---------|----------------|
| 1 | Asset | Bank accounts, cash, investments | Debit |
| 4 | Liability | Credit cards, loans, debts | Credit |
| 5 | Expense | Spending categories | Debit |
| 6 | Income | Revenue sources | Credit |
| 7 | Adjustment | Balance adjustments | Debit/Credit |

### Transaction Types

| Type Value | Name | Description |
|------------|------|-------------|
| 1 | Transfer | Money movement between Asset/Liability accounts |
| 2 | Income | Money received (Income account -> Asset/Liability account) |
| 3 | Expense | Money spent (Asset/Liability account -> Expense account) |
| 5 | Adjustment | Balance corrections or adjustments |

### Multi-Currency Support

The system supports multiple currencies with:
- **Source Currency**: Original currency of funds leaving an account
- **Destination Currency**: Currency of funds entering an account
- **Base Currency**: Common currency for reporting (amounts converted via currency rates)
- **FX Fields**: Track original foreign currency amounts in expense transactions

### Double-Entry Bookkeeping

Every transaction creates two entries in `double_entries`:
- One **debit** entry (IsDebit = true)
- One **credit** entry (IsDebit = false)

The sum of all debits must equal the sum of all credits (accounting equation).

### Soft Deletes

Most tables use soft deletes via `deleted_at` timestamp:
- Records with `deleted_at IS NULL` are active
- Records with `deleted_at IS NOT NULL` are logically deleted
- Always filter with `WHERE deleted_at IS NULL` for active records

## Key Relationships

1. **Transactions -> Accounts**: Each transaction links to source and/or destination accounts
2. **Transactions -> Categories**: Optional category for expense classification
3. **Transactions -> Tags**: Many-to-many via `tag_ids` array (requires array operations for queries)
4. **Double Entries -> Transactions**: Two entries per transaction for bookkeeping
5. **Double Entries -> Accounts**: Links ledger entries to specific accounts
6. **Daily Stats -> Accounts**: Pre-aggregated daily balance changes per account

## Indexes for Query Optimization

Key indexes to leverage for efficient queries:

| Index | Purpose |
|-------|---------|
| `idx_transactions_active_date` | Filter active transactions by date |
| `idx_transactions_active_date_type` | Filter by date and transaction type |
| `ix_source_tx` | Query by source account and date |
| `ix_dest_tx` | Query by destination account and date |
| `ix_source_dest_tx` | Query by both accounts and date |
| `daily_stat_pk` | Fast lookup by account and date |
| `ix_latest_stat` | Get most recent stats per account |
| `ix_double_entries_transaction_date` | Ledger queries by account and date |

## Quick Reference

### Common Query Patterns

```sql
-- Active accounts only
SELECT * FROM accounts WHERE deleted_at IS NULL;

-- Transactions for an account
SELECT * FROM transactions
WHERE deleted_at IS NULL
  AND (source_account_id = :id OR destination_account_id = :id);

-- Transactions with specific tag
SELECT * FROM transactions
WHERE deleted_at IS NULL
  AND :tag_id = ANY(tag_ids);

-- Account balance from daily stats
SELECT SUM(amount) as balance
FROM daily_stat
WHERE account_id = :id;
```

See individual table documentation for detailed column descriptions and query examples.
