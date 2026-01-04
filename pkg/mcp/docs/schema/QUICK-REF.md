# Schema Quick Reference

> Compact reference for all tables, columns, and enums. For details, see individual table docs.

## Tables Overview

| Table | Primary Key | Description |
|-------|-------------|-------------|
| accounts | id (int) | All account types: assets, liabilities, categories |
| transactions | id (int) | All financial transactions |
| categories | id (int) | Transaction categories |
| tags | id (int) | Transaction tags |
| currencies | id (text) | Currency codes and exchange rates |
| daily_stat | composite | Pre-computed daily balances |
| double_entries | id (int) | Double-entry ledger |
| rules | id (int) | Lua automation rules |
| schedule_rules | id (int) | Cron-scheduled rules |
| users | id (int) | User authentication |
| import_deduplication | composite | Import duplicate detection |
| service_tokens | id (uuid) | API service tokens |
| jti_revocations | id (text) | Revoked token tracking |

---

## accounts

```sql
id              integer PRIMARY KEY
name            text NOT NULL
type            integer NOT NULL        -- See AccountType enum
currency        text NOT NULL           -- e.g., "USD", "EUR"
current_balance numeric(20,8)           -- Live balance
display_order   integer
flags           integer                 -- Bitset, see AccountFlags
extra           jsonb                   -- Custom data
created_at      timestamp
deleted_at      timestamp               -- Soft delete
```

## transactions

```sql
id                                  integer PRIMARY KEY
title                               text
transaction_type                    integer NOT NULL    -- See TransactionType enum
source_account_id                   integer             -- FK → accounts
destination_account_id              integer             -- FK → accounts
source_amount                       numeric(20,8)
source_currency                     text
destination_amount                  numeric(20,8)
destination_currency                text
source_amount_in_base_currency      numeric(20,8)       -- Converted amount
destination_amount_in_base_currency numeric(20,8)       -- Converted amount
fx_source_amount                    numeric(20,8)       -- FX transaction original
fx_source_currency                  text
transaction_date_time               timestamp           -- Full datetime
transaction_date_only               date                -- Date only (for grouping)
category_id                         integer             -- FK → categories
tag_ids                             integer[]           -- Array of tag IDs
notes                               text
created_at                          timestamp
deleted_at                          timestamp           -- Soft delete
```

## categories

```sql
id         integer PRIMARY KEY
name       text NOT NULL
created_at timestamp
deleted_at timestamp
```

## tags

```sql
id         integer PRIMARY KEY
name       text NOT NULL
created_at timestamp
deleted_at timestamp
```

## currencies

```sql
id             text PRIMARY KEY        -- "USD", "EUR", etc.
rate           numeric(20,10) NOT NULL -- Units per 1 base currency
decimal_places integer DEFAULT 2
is_active      boolean DEFAULT true
created_at     timestamp
deleted_at     timestamp
```

**Conversion:** `base_amount = amount / rate`

## daily_stat

```sql
account_id integer NOT NULL         -- FK → accounts
date       date NOT NULL
amount     numeric(20,8)            -- Running balance at end of day
PRIMARY KEY (account_id, date)
```

## double_entries

```sql
id                      integer PRIMARY KEY
transaction_id          integer NOT NULL    -- FK → transactions
account_id              integer NOT NULL    -- FK → accounts
is_debit                boolean NOT NULL
amount                  numeric(20,8)
amount_in_base_currency numeric(20,8)
transaction_date        timestamp
created_at              timestamp
deleted_at              timestamp
```

## rules

```sql
id          integer PRIMARY KEY
name        text NOT NULL
script      text NOT NULL           -- Lua code
enabled     boolean DEFAULT true
sort_order  integer DEFAULT 0
is_final    boolean DEFAULT false   -- Stop processing after this rule
group       text                    -- Rule group name
created_at  timestamp
deleted_at  timestamp
```

## schedule_rules

```sql
id              integer PRIMARY KEY
name            text NOT NULL
cron_expression text NOT NULL       -- Cron syntax
script          text NOT NULL       -- Lua code
enabled         boolean DEFAULT true
last_run        timestamp
next_run        timestamp
created_at      timestamp
deleted_at      timestamp
```

## users

```sql
id         integer PRIMARY KEY
login      text NOT NULL UNIQUE
password   text NOT NULL           -- Bcrypt hash
created_at timestamp
deleted_at timestamp
```

---

## Enums

### TransactionType
| Value | Name | Source Account | Destination Account |
|-------|------|----------------|---------------------|
| 1 | Transfer | Asset/Liability | Asset/Liability |
| 2 | Income | Income | Asset |
| 3 | Expense | Asset/Liability | Expense |
| 5 | Adjustment | Adjustment | Asset/Liability |

### AccountType
| Value | Name | Normal Balance | Use |
|-------|------|----------------|-----|
| 1 | Asset | Debit | Bank, cash, investments |
| 4 | Liability | Credit | Credit cards, loans |
| 5 | Expense | Debit | Spending categories |
| 6 | Income | Credit | Revenue sources |
| 7 | Adjustment | - | Balance corrections |

### AccountFlags (Bitset)
| Bit | Value | Name |
|-----|-------|------|
| 0 | 1 | IsDefault |
| 1 | 2 | ExcludeFromNetWorth |
| 2 | 4 | ExcludeFromReports |

---

## Key Relationships

```
transactions.source_account_id      → accounts.id
transactions.destination_account_id → accounts.id
transactions.category_id            → categories.id
transactions.tag_ids                → tags.id (array)
double_entries.transaction_id       → transactions.id
double_entries.account_id           → accounts.id
daily_stat.account_id               → accounts.id
```

---

## Common Filters

```sql
-- Active records only (always include)
WHERE deleted_at IS NULL

-- Expenses only
WHERE transaction_type = 3

-- Income only
WHERE transaction_type = 2

-- Transfers only
WHERE transaction_type = 1

-- Assets and liabilities (for net worth)
WHERE type IN (1, 4)

-- This month
WHERE DATE_TRUNC('month', transaction_date_only) = DATE_TRUNC('month', CURRENT_DATE)

-- Has specific tag
WHERE :tag_id = ANY(tag_ids)
```

---

## See Also
- [Schema Overview](overview.md) - ER diagram
- [Enums Reference](enums.md) - Full enum details
- [Indexes](indexes.md) - Index definitions
