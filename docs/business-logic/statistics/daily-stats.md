# Daily Statistics

Pre-computed daily balance snapshots for efficient analytics.

## Purpose

The `daily_stat` table stores running account balances at the end of each day. This enables:
- O(1) current balance lookups
- Efficient historical balance queries
- Fast time-series analytics
- No need to scan all transactions for balance calculations

## Table Schema

| Column | Type | Description |
|--------|------|-------------|
| account_id | integer | FK to accounts (composite PK) |
| date | date | Date (composite PK) |
| amount | numeric | Running balance at end of day |

**Primary Key:** `(account_id, date)`

## Key Concept: Running Balance

The `amount` field stores the **cumulative running balance**, not daily change:

```
Day 1: 100.00   (initial deposit)
Day 2: 150.00   (deposited 50)
Day 3: 125.00   (spent 25)
```

**Not** daily changes like:
```
Day 1: +100.00
Day 2: +50.00
Day 3: -25.00
```

## Recalculation Algorithm

### When Triggered

Statistics are recalculated when:
1. Transaction created
2. Transaction updated
3. Transaction deleted

### Step-by-Step Process

#### 1. Identify Impacted Accounts

For each transaction, both source and destination accounts are impacted:

```go
func getAccountsForTx(tx *Transaction) []int32 {
    return []int32{tx.SourceAccountID, tx.DestinationAccountID}
}
```

#### 2. Find Earliest Impact Date

When multiple transactions are processed, find the earliest date per account:

```go
impactedAccounts := map[int32]time.Time{}  // account_id -> earliest date

for _, tx := range transactions {
    for _, accountID := range getAccountsForTx(tx) {
        if existing, ok := impactedAccounts[accountID]; !ok {
            impactedAccounts[accountID] = tx.TransactionDateTime
        } else if existing.After(tx.TransactionDateTime) {
            impactedAccounts[accountID] = tx.TransactionDateTime
        }
    }
}
```

**Code Reference:** `pkg/transactions/stats.go:37-53`

#### 3. Determine Recalculation Start

Find the actual starting point by looking at existing stats:

```sql
SELECT LEAST(
    @startDate,
    COALESCE(
        (SELECT date FROM daily_stat
         WHERE account_id = @accountID AND date < @startDate
         ORDER BY date DESC LIMIT 1),
        (SELECT MIN(transaction_date_only) FROM transactions
         WHERE source_account_id = @accountID OR destination_account_id = @accountID
         AND deleted_at IS NULL)
    )
)
```

#### 4. Generate Date Series

Create continuous dates from start to today:

```sql
generate_series(
    min_date,
    GREATEST(NOW()::DATE, max_transaction_date) + 1,
    '1 day'::INTERVAL
)
```

#### 5. Sum Daily Transactions

Calculate net change per day:

```sql
SELECT
    transaction_date_only as tx_date,
    SUM(
        CASE
            WHEN source_account_id = @accountID THEN source_amount
            ELSE destination_amount
        END
    ) as amount
FROM transactions
WHERE (source_account_id = @accountID OR destination_account_id = @accountID)
  AND deleted_at IS NULL
GROUP BY transaction_date_only
```

#### 6. Calculate Running Balance

Apply window function for cumulative sum:

```sql
SELECT
    date,
    SUM(COALESCE(daily_amount, 0) + COALESCE(initial_balance, 0))
        OVER (ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) as amount
FROM date_series
LEFT JOIN daily_sums ON tx_date = date
LEFT JOIN initial_value ON initial_date = date
ORDER BY date ASC
```

#### 7. Update Tables

```sql
-- Upsert daily_stat
INSERT INTO daily_stat(account_id, date, amount)
SELECT @accountID, date, amount
FROM running_totals
ON CONFLICT ON CONSTRAINT daily_stat_pk
DO UPDATE SET amount = excluded.amount;

-- Update current balance
UPDATE accounts
SET current_balance = (SELECT amount FROM running WHERE date = MAX(date)),
    last_updated_at = NOW()
WHERE id = @accountID;
```

**Code Reference:** `pkg/transactions/scripts/daily_recalculate.sql`

## SQL Queries

### Current Balance

```sql
SELECT amount as current_balance
FROM daily_stat
WHERE account_id = :account_id
ORDER BY date DESC
LIMIT 1;
```

### Balance at Specific Date

```sql
SELECT amount as balance
FROM daily_stat
WHERE account_id = :account_id
  AND date <= :target_date
ORDER BY date DESC
LIMIT 1;
```

### Balance History (Last 30 Days)

```sql
SELECT date, amount as balance
FROM daily_stat
WHERE account_id = :account_id
  AND date >= CURRENT_DATE - INTERVAL '30 days'
ORDER BY date;
```

### Net Worth Over Time

```sql
SELECT
    ds.date,
    SUM(CASE WHEN a.type = 1 THEN ds.amount ELSE 0 END) -
    SUM(CASE WHEN a.type = 4 THEN ds.amount ELSE 0 END) as net_worth
FROM daily_stat ds
JOIN accounts a ON a.id = ds.account_id
WHERE a.type IN (1, 4)  -- Asset, Liability
  AND a.deleted_at IS NULL
GROUP BY ds.date
ORDER BY ds.date;
```

### Daily Balance Change

To get daily changes instead of running totals:

```sql
SELECT
    ds1.date,
    ds1.amount - COALESCE(ds2.amount, 0) as daily_change
FROM daily_stat ds1
LEFT JOIN daily_stat ds2
    ON ds2.account_id = ds1.account_id
    AND ds2.date = ds1.date - INTERVAL '1 day'
WHERE ds1.account_id = :account_id
ORDER BY ds1.date DESC;
```

### Accounts with Activity in Date Range

```sql
SELECT DISTINCT ds.account_id, a.name
FROM daily_stat ds
JOIN accounts a ON a.id = ds.account_id
WHERE ds.date BETWEEN :start_date AND :end_date
  AND a.deleted_at IS NULL;
```

### Monthly Balance Snapshots

```sql
SELECT DISTINCT ON (DATE_TRUNC('month', date))
    DATE_TRUNC('month', date) as month,
    amount as end_of_month_balance
FROM daily_stat
WHERE account_id = :account_id
ORDER BY DATE_TRUNC('month', date), date DESC;
```

## Performance Characteristics

### Why Pre-compute?

| Approach | Balance Query | Historical Query |
|----------|---------------|------------------|
| Sum transactions | O(n) | O(n * days) |
| Pre-computed stats | O(1) | O(days) |

### Indexes

```sql
CREATE UNIQUE INDEX daily_stat_pk ON daily_stat (account_id, date);
CREATE INDEX daily_stat_account_id_index ON daily_stat (account_id);
CREATE INDEX ix_latest_stat ON daily_stat (account_id, date DESC);
```

### Memory Optimization

StatService uses an LRU cache to avoid redundant calculations:

```go
type StatService struct {
    noGapTillTime *expirable.LRU[string, time.Time]
}

// Cache expires after 10 minutes
noGapTillTime: expirable.NewLRU[string, time.Time](1000, nil, 10*time.Minute)
```

**Code Reference:** `pkg/transactions/stats.go:18-26`

## Data Integrity

### No Gaps Guarantee

The recalculation generates all dates from first transaction to today:

```sql
generate_series(first_date, today, '1 day')
```

Every day has an entry, even if no transactions occurred.

### Consistency with accounts.current_balance

The `accounts.current_balance` is updated atomically with daily_stat:

```sql
UPDATE accounts
SET current_balance = (SELECT amount FROM running WHERE date = MAX(date))
WHERE id = @accountID
```

### Verification Query

Check for inconsistencies:

```sql
SELECT
    a.id,
    a.name,
    a.current_balance,
    ds.amount as stat_balance,
    a.current_balance - COALESCE(ds.amount, 0) as difference
FROM accounts a
LEFT JOIN (
    SELECT DISTINCT ON (account_id)
        account_id, amount
    FROM daily_stat
    ORDER BY account_id, date DESC
) ds ON ds.account_id = a.id
WHERE a.deleted_at IS NULL
  AND ABS(a.current_balance - COALESCE(ds.amount, 0)) > 0.01;
-- Should return no rows
```

## Edge Cases

### First Transaction

When an account has no prior daily_stat entries, the system:
1. Finds the earliest transaction date
2. Uses that as the starting point
3. Initial balance is 0

### Backdated Transaction

If a transaction is added with a past date:
1. Recalculation starts from that past date
2. All subsequent daily_stat entries are updated
3. Running balances cascade forward correctly

### Transaction Deletion

When deleting a transaction:
1. Recalculation starts from the deleted transaction's date
2. All entries from that date forward are recomputed
3. The "hole" is filled automatically

## Usage Recommendations

| Query Type | Use daily_stat | Use transactions |
|------------|----------------|------------------|
| Current balance | ✓ | |
| Balance at date | ✓ | |
| Balance trends | ✓ | |
| Net worth history | ✓ | |
| Transaction details | | ✓ |
| Category breakdown | | ✓ |
| Search by title | | ✓ |
| Individual records | | ✓ |
