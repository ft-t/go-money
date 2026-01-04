# rules and schedule_rules Tables

These tables store Lua scripts for transaction processing and scheduled automation.

## rules Table

Transaction processing rules executed when transactions are created.

### Schema

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | integer | NO | auto-increment | Primary key |
| title | text | YES | - | Rule display name |
| script | text | NO | - | Lua script code |
| interpreter_type | integer | NO | - | Script interpreter (0=unspecified, 1=Lua) |
| sort_order | integer | NO | - | Execution order (lower runs first) |
| enabled | boolean | NO | - | Whether rule is active |
| is_final_rule | boolean | NO | - | Stop processing after this rule |
| group_name | text | NO | - | Logical grouping |
| created_at | timestamp | NO | - | Record creation time |
| updated_at | timestamp | NO | - | Record update time |
| deleted_at | timestamp | YES | - | Soft delete timestamp |

### Indexes

| Index | Definition | Purpose |
|-------|------------|---------|
| rules_pk | UNIQUE (id) | Primary key |

## schedule_rules Table

Scheduled rules executed on a cron schedule.

### Schema

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | integer | NO | auto-increment | Primary key |
| title | text | NO | - | Rule display name |
| script | text | NO | - | Lua script code |
| interpreter_type | integer | NO | - | Script interpreter (0=unspecified, 1=Lua) |
| cron_expression | text | NO | - | Cron schedule expression |
| enabled | boolean | NO | - | Whether rule is active |
| group_name | text | NO | - | Logical grouping |
| last_run_at | timestamp | YES | - | Last execution timestamp |
| created_at | timestamp | NO | - | Record creation time |
| updated_at | timestamp | NO | - | Record update time |
| deleted_at | timestamp | YES | - | Soft delete timestamp |

## Interpreter Types

| Value | Name | Description |
|-------|------|-------------|
| 0 | UNSPECIFIED | Not used |
| 1 | LUA | Lua scripting language |

## Common Queries

### All Active Transaction Rules

```sql
SELECT
    id,
    title,
    script,
    sort_order,
    is_final_rule,
    group_name
FROM rules
WHERE enabled = true
  AND deleted_at IS NULL
ORDER BY sort_order;
```

### Rules by Group

```sql
SELECT
    group_name,
    COUNT(*) as rule_count
FROM rules
WHERE deleted_at IS NULL
GROUP BY group_name
ORDER BY group_name;
```

### Scheduled Rules Status

```sql
SELECT
    id,
    title,
    cron_expression,
    enabled,
    last_run_at,
    CURRENT_TIMESTAMP - last_run_at as time_since_last_run
FROM schedule_rules
WHERE deleted_at IS NULL
ORDER BY last_run_at DESC NULLS LAST;
```

### Active Scheduled Rules

```sql
SELECT
    id,
    title,
    cron_expression,
    group_name
FROM schedule_rules
WHERE enabled = true
  AND deleted_at IS NULL
ORDER BY title;
```

### Rules That Haven't Run Recently

```sql
SELECT *
FROM schedule_rules
WHERE enabled = true
  AND deleted_at IS NULL
  AND (last_run_at IS NULL OR last_run_at < CURRENT_TIMESTAMP - INTERVAL '24 hours');
```

## Rule Execution Order

Transaction rules are executed in `sort_order` sequence:

1. Lower `sort_order` values run first
2. If `is_final_rule = true`, no further rules execute
3. Rules can modify transaction properties (title, category, tags, etc.)

## Cron Expression Format

Standard cron format with 5 fields:
```
* * * * *
│ │ │ │ │
│ │ │ │ └── Day of week (0-6, Sunday = 0)
│ │ │ └──── Month (1-12)
│ │ └────── Day of month (1-31)
│ └──────── Hour (0-23)
└────────── Minute (0-59)
```

Examples:
- `0 0 * * *` - Daily at midnight
- `0 */6 * * *` - Every 6 hours
- `0 0 1 * *` - First day of each month

## Notes

- Only interpreter_type = 1 (Lua) is currently supported
- Transaction rules have access to transaction context
- Scheduled rules run independently on their schedule
- `is_final_rule` prevents subsequent rules from running
- Group names help organize related rules
