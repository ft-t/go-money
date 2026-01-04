# API Endpoints Reference

Go Money uses [Connect](https://connectrpc.com/) protocol, providing both gRPC and HTTP/JSON compatibility.

## Protocol

All endpoints use POST requests with JSON bodies:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"key": "value"}' \
  http://localhost:8080/<service>/<method>
```

## Services Overview

| Service | Package | Description |
|---------|---------|-------------|
| UsersService | users.v1 | User authentication |
| ConfigurationService | configuration.v1 | App configuration, service tokens |
| AccountsService | accounts.v1 | Account management |
| TransactionsService | transactions.v1 | Transaction CRUD |
| CategoriesService | categories.v1 | Category management |
| TagsService | tags.v1 | Tag management |
| CurrencyService | currency.v1 | Currency and exchange |
| RulesService | rules.v1 | Automation rules |
| ImportService | import.v1 | Data import |
| AnalyticsService | analytics.v1 | Financial analytics |
| MaintenanceService | maintenance.v1 | System maintenance |

---

## UsersService

Package: `gomoneypb.users.v1`

### Login

Authenticate and receive a JWT token.

```
POST /gomoneypb.users.v1.UsersService/Login
```

**Auth Required:** No

**Request:**
```json
{
  "login": "admin",
  "password": "secret"
}
```

**Response:**
```json
{
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Create

Create the first admin user.

```
POST /gomoneypb.users.v1.UsersService/Create
```

**Auth Required:** No (only works when no users exist)

**Request:**
```json
{
  "login": "admin",
  "password": "secure_password"
}
```

**Response:**
```json
{
  "id": 1
}
```

---

## ConfigurationService

Package: `gomoneypb.configuration.v1`

### GetConfiguration

Get public application configuration.

```
POST /gomoneypb.configuration.v1.ConfigurationService/GetConfiguration
```

**Auth Required:** No

**Response:**
```json
{
  "base_currency": "USD",
  "should_create_admin": false
}
```

### GetConfigsByKeys

Get configuration values by keys.

```
POST /gomoneypb.configuration.v1.ConfigurationService/GetConfigsByKeys
```

**Auth Required:** Yes

**Request:**
```json
{
  "keys": ["theme", "language"]
}
```

### SetConfigByKey

Set a configuration value.

```
POST /gomoneypb.configuration.v1.ConfigurationService/SetConfigByKey
```

**Auth Required:** Yes

**Request:**
```json
{
  "key": "theme",
  "value": "dark"
}
```

### GetServiceTokens

List service tokens.

```
POST /gomoneypb.configuration.v1.ConfigurationService/GetServiceTokens
```

**Auth Required:** Yes

**Request:**
```json
{
  "ids": []
}
```

**Response:**
```json
{
  "service_tokens": [
    {
      "id": "uuid-here",
      "name": "MCP Integration",
      "expires_at": "2025-12-31T23:59:59Z",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

### CreateServiceToken

Create a new service token.

```
POST /gomoneypb.configuration.v1.ConfigurationService/CreateServiceToken
```

**Auth Required:** Yes

**Request:**
```json
{
  "name": "API Access",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

**Response:**
```json
{
  "service_token": { ... },
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### RevokeServiceToken

Revoke a service token.

```
POST /gomoneypb.configuration.v1.ConfigurationService/RevokeServiceToken
```

**Auth Required:** Yes

**Request:**
```json
{
  "id": "uuid-here"
}
```

---

## AccountsService

Package: `gomoneypb.accounts.v1`

### ListAccounts

List all accounts.

```
POST /gomoneypb.accounts.v1.AccountsService/ListAccounts
```

**Auth Required:** Yes

**Request:**
```json
{
  "include_deleted": false
}
```

**Response:**
```json
{
  "accounts": [
    {
      "id": 1,
      "name": "Checking",
      "type": "ACCOUNT_TYPE_ASSET",
      "currency": "USD",
      "current_balance": "1500.00",
      "display_order": 1
    }
  ]
}
```

### CreateAccount

Create a new account.

```
POST /gomoneypb.accounts.v1.AccountsService/CreateAccount
```

**Auth Required:** Yes

**Request:**
```json
{
  "account": {
    "name": "Savings",
    "type": "ACCOUNT_TYPE_ASSET",
    "currency": "USD"
  }
}
```

### CreateAccountsBulk

Create multiple accounts at once.

```
POST /gomoneypb.accounts.v1.AccountsService/CreateAccountsBulk
```

**Auth Required:** Yes

**Request:**
```json
{
  "accounts": [
    { "name": "Account 1", "type": "ACCOUNT_TYPE_ASSET", "currency": "USD" },
    { "name": "Account 2", "type": "ACCOUNT_TYPE_ASSET", "currency": "EUR" }
  ]
}
```

### UpdateAccount

Update an existing account.

```
POST /gomoneypb.accounts.v1.AccountsService/UpdateAccount
```

**Auth Required:** Yes

**Request:**
```json
{
  "account": {
    "id": 1,
    "name": "Updated Name"
  }
}
```

### DeleteAccount

Soft-delete an account.

```
POST /gomoneypb.accounts.v1.AccountsService/DeleteAccount
```

**Auth Required:** Yes

**Request:**
```json
{
  "id": 1
}
```

---

## TransactionsService

Package: `gomoneypb.transactions.v1`

### ListTransactions

List transactions with filtering and pagination.

```
POST /gomoneypb.transactions.v1.TransactionsService/ListTransactions
```

**Auth Required:** Yes

**Request:**
```json
{
  "page_size": 50,
  "page_token": "",
  "filter": {
    "account_ids": [1, 2],
    "category_ids": [5],
    "tag_ids": [3],
    "date_from": "2024-01-01",
    "date_to": "2024-12-31",
    "transaction_types": ["TRANSACTION_TYPE_EXPENSE"]
  }
}
```

**Response:**
```json
{
  "transactions": [
    {
      "id": 123,
      "title": "Grocery Store",
      "transaction_type": "TRANSACTION_TYPE_EXPENSE",
      "source_account_id": 1,
      "destination_account_id": 5,
      "destination_amount": "45.99",
      "destination_currency": "USD",
      "transaction_date_only": "2024-01-15",
      "category_id": 3,
      "tag_ids": [1, 2]
    }
  ],
  "next_page_token": "abc123"
}
```

### CreateTransaction

Create a new transaction.

```
POST /gomoneypb.transactions.v1.TransactionsService/CreateTransaction
```

**Auth Required:** Yes

**Request:**
```json
{
  "transaction": {
    "title": "Coffee",
    "transaction_type": "TRANSACTION_TYPE_EXPENSE",
    "source_account_id": 1,
    "destination_account_id": 5,
    "destination_amount": "4.50",
    "destination_currency": "USD",
    "transaction_date_time": "2024-01-15T10:30:00Z",
    "category_id": 3,
    "tag_ids": [1]
  }
}
```

### CreateTransactionsBulk

Create multiple transactions.

```
POST /gomoneypb.transactions.v1.TransactionsService/CreateTransactionsBulk
```

**Auth Required:** Yes

**Request:**
```json
{
  "transactions": [
    { "title": "Transaction 1", ... },
    { "title": "Transaction 2", ... }
  ]
}
```

### UpdateTransaction

Update an existing transaction.

```
POST /gomoneypb.transactions.v1.TransactionsService/UpdateTransaction
```

**Auth Required:** Yes

**Request:**
```json
{
  "transaction": {
    "id": 123,
    "title": "Updated Title",
    "category_id": 5
  }
}
```

### DeleteTransactions

Soft-delete transactions.

```
POST /gomoneypb.transactions.v1.TransactionsService/DeleteTransactions
```

**Auth Required:** Yes

**Request:**
```json
{
  "ids": [123, 124, 125]
}
```

### GetApplicableAccounts

Get valid source/destination accounts per transaction type.

```
POST /gomoneypb.transactions.v1.TransactionsService/GetApplicableAccounts
```

**Auth Required:** Yes

**Response:**
```json
{
  "applicable_records": [
    {
      "transaction_type": "TRANSACTION_TYPE_EXPENSE",
      "source_accounts": [...],
      "destination_accounts": [...]
    }
  ]
}
```

### GetTitleSuggestions

Get title autocomplete suggestions.

```
POST /gomoneypb.transactions.v1.TransactionsService/GetTitleSuggestions
```

**Auth Required:** Yes

**Request:**
```json
{
  "prefix": "Gro",
  "limit": 10
}
```

---

## CategoriesService

Package: `gomoneypb.categories.v1`

### ListCategories

List all categories.

```
POST /gomoneypb.categories.v1.CategoriesService/ListCategories
```

**Auth Required:** Yes

**Response:**
```json
{
  "categories": [
    { "id": 1, "name": "Groceries" },
    { "id": 2, "name": "Transportation" }
  ]
}
```

### CreateCategory

Create a new category.

```
POST /gomoneypb.categories.v1.CategoriesService/CreateCategory
```

**Auth Required:** Yes

**Request:**
```json
{
  "category": {
    "name": "Entertainment"
  }
}
```

### UpdateCategory

Update a category.

```
POST /gomoneypb.categories.v1.CategoriesService/UpdateCategory
```

**Auth Required:** Yes

**Request:**
```json
{
  "category": {
    "id": 1,
    "name": "Updated Name"
  }
}
```

### DeleteCategory

Delete a category.

```
POST /gomoneypb.categories.v1.CategoriesService/DeleteCategory
```

**Auth Required:** Yes

**Request:**
```json
{
  "id": 1
}
```

---

## TagsService

Package: `gomoneypb.tags.v1`

### ListTags

List all tags.

```
POST /gomoneypb.tags.v1.TagsService/ListTags
```

**Auth Required:** Yes

### CreateTag

Create a new tag.

```
POST /gomoneypb.tags.v1.TagsService/CreateTag
```

**Auth Required:** Yes

**Request:**
```json
{
  "tag": {
    "name": "vacation"
  }
}
```

### UpdateTag

Update a tag.

```
POST /gomoneypb.tags.v1.TagsService/UpdateTag
```

**Auth Required:** Yes

### DeleteTag

Delete a tag.

```
POST /gomoneypb.tags.v1.TagsService/DeleteTag
```

**Auth Required:** Yes

**Request:**
```json
{
  "id": 1
}
```

### ImportTags

Bulk import tags.

```
POST /gomoneypb.tags.v1.TagsService/ImportTags
```

**Auth Required:** Yes

**Request:**
```json
{
  "tags": [
    { "name": "tag1" },
    { "name": "tag2" }
  ]
}
```

---

## CurrencyService

Package: `gomoneypb.currency.v1`

### GetCurrencies

List all currencies with rates.

```
POST /gomoneypb.currency.v1.CurrencyService/GetCurrencies
```

**Auth Required:** Yes

**Response:**
```json
{
  "currencies": [
    { "id": "USD", "rate": "1.0", "decimal_places": 2, "is_active": true },
    { "id": "EUR", "rate": "0.92", "decimal_places": 2, "is_active": true }
  ]
}
```

### CreateCurrency

Create a new currency.

```
POST /gomoneypb.currency.v1.CurrencyService/CreateCurrency
```

**Auth Required:** Yes

**Request:**
```json
{
  "currency": {
    "id": "GBP",
    "rate": "0.79",
    "decimal_places": 2
  }
}
```

### UpdateCurrency

Update currency rate.

```
POST /gomoneypb.currency.v1.CurrencyService/UpdateCurrency
```

**Auth Required:** Yes

**Request:**
```json
{
  "currency": {
    "id": "EUR",
    "rate": "0.91"
  }
}
```

### DeleteCurrency

Delete a currency.

```
POST /gomoneypb.currency.v1.CurrencyService/DeleteCurrency
```

**Auth Required:** Yes

**Request:**
```json
{
  "id": "GBP"
}
```

### Exchange

Convert amount between currencies.

```
POST /gomoneypb.currency.v1.CurrencyService/Exchange
```

**Auth Required:** Yes

**Request:**
```json
{
  "from_currency": "USD",
  "to_currency": "EUR",
  "amount": "100.00"
}
```

**Response:**
```json
{
  "amount": "92.00"
}
```

---

## RulesService

Package: `gomoneypb.rules.v1`

### ListRules

List transaction processing rules.

```
POST /gomoneypb.rules.v1.RulesService/ListRules
```

**Auth Required:** Yes

### CreateRule

Create a Lua automation rule.

```
POST /gomoneypb.rules.v1.RulesService/CreateRule
```

**Auth Required:** Yes

**Request:**
```json
{
  "rule": {
    "name": "Auto-categorize groceries",
    "script": "if string.match(tx.title, 'grocery') then tx.category_id = 5 end",
    "enabled": true,
    "sort_order": 10,
    "group": "categorization"
  }
}
```

### UpdateRule

Update a rule.

```
POST /gomoneypb.rules.v1.RulesService/UpdateRule
```

**Auth Required:** Yes

### DeleteRule

Delete a rule.

```
POST /gomoneypb.rules.v1.RulesService/DeleteRule
```

**Auth Required:** Yes

### DryRunRule

Test a rule without saving.

```
POST /gomoneypb.rules.v1.RulesService/DryRunRule
```

**Auth Required:** Yes

**Request:**
```json
{
  "script": "tx.category_id = 5",
  "transaction_id": 123
}
```

### ListScheduleRules

List scheduled (cron) rules.

```
POST /gomoneypb.rules.v1.RulesService/ListScheduleRules
```

**Auth Required:** Yes

### CreateScheduleRule

Create a scheduled rule.

```
POST /gomoneypb.rules.v1.RulesService/CreateScheduleRule
```

**Auth Required:** Yes

**Request:**
```json
{
  "rule": {
    "name": "Monthly Rent",
    "cron_expression": "0 0 1 * *",
    "script": "create_transaction({title='Rent', amount=1500, ...})",
    "enabled": true
  }
}
```

### UpdateScheduleRule

Update a scheduled rule.

```
POST /gomoneypb.rules.v1.RulesService/UpdateScheduleRule
```

**Auth Required:** Yes

### DeleteScheduleRule

Delete a scheduled rule.

```
POST /gomoneypb.rules.v1.RulesService/DeleteScheduleRule
```

**Auth Required:** Yes

### ValidateCronExpression

Validate a cron expression.

```
POST /gomoneypb.rules.v1.RulesService/ValidateCronExpression
```

**Auth Required:** Yes

**Request:**
```json
{
  "cron_expression": "0 0 * * *"
}
```

**Response:**
```json
{
  "valid": true
}
```

---

## ImportService

Package: `gomoneypb.import.v1`

### ParseTransactions

Parse transactions from file without importing.

```
POST /gomoneypb.import.v1.ImportService/ParseTransactions
```

**Auth Required:** Yes

**Request:**
```json
{
  "source": "IMPORT_SOURCE_REVOLUT",
  "data": "<base64-encoded-file>"
}
```

**Response:**
```json
{
  "transactions": [...],
  "errors": []
}
```

### ImportTransactions

Import transactions from file.

```
POST /gomoneypb.import.v1.ImportService/ImportTransactions
```

**Auth Required:** Yes

**Request:**
```json
{
  "source": "IMPORT_SOURCE_REVOLUT",
  "data": "<base64-encoded-file>",
  "account_mapping": {
    "EUR": 1,
    "USD": 2
  }
}
```

### Supported Import Sources

| Source | Description |
|--------|-------------|
| IMPORT_SOURCE_FIREFLY | Firefly III export |
| IMPORT_SOURCE_PRIVAT24 | PrivatBank (Ukraine) |
| IMPORT_SOURCE_MONO | Monobank (Ukraine) |
| IMPORT_SOURCE_PARIBAS | BNP Paribas |
| IMPORT_SOURCE_REVOLUT | Revolut |

---

## AnalyticsService

Package: `gomoneypb.analytics.v1`

### GetDebitsAndCreditsSummary

Get debit/credit summary for accounts.

```
POST /gomoneypb.analytics.v1.AnalyticsService/GetDebitsAndCreditsSummary
```

**Auth Required:** Yes

**Request:**
```json
{
  "account_ids": [1, 2],
  "date_from": "2024-01-01",
  "date_to": "2024-12-31"
}
```

**Response:**
```json
{
  "summaries": [
    {
      "account_id": 1,
      "total_debits": "5000.00",
      "total_credits": "3500.00"
    }
  ]
}
```

---

## MaintenanceService

Package: `gomoneypb.maintenance.v1`

### RecalculateAll

Recalculate all balances and statistics.

```
POST /gomoneypb.maintenance.v1.MaintenanceService/RecalculateAll
```

**Auth Required:** Yes

**Response:**
```json
{
  "success": true
}
```

---

## Error Handling

### Connect Error Codes

| Code | Description |
|------|-------------|
| PERMISSION_DENIED | Invalid or missing authentication |
| UNAUTHENTICATED | Token expired or invalid |
| INTERNAL | Server error |
| INVALID_ARGUMENT | Bad request parameters |
| NOT_FOUND | Resource not found |

### Error Response Format

```json
{
  "code": "PERMISSION_DENIED",
  "message": "invalid token"
}
```

---

## Enums Reference

### TransactionType

| Value | Meaning |
|-------|---------|
| TRANSACTION_TYPE_UNSPECIFIED | Not set |
| TRANSACTION_TYPE_TRANSFER | Transfer between accounts |
| TRANSACTION_TYPE_INCOME | Income |
| TRANSACTION_TYPE_EXPENSE | Expense |
| TRANSACTION_TYPE_ADJUSTMENT | Balance adjustment |

### AccountType

| Value | Meaning |
|-------|---------|
| ACCOUNT_TYPE_UNSPECIFIED | Not set |
| ACCOUNT_TYPE_ASSET | Asset account (checking, savings) |
| ACCOUNT_TYPE_LIABILITY | Liability (credit card, loan) |
| ACCOUNT_TYPE_EXPENSE | Expense category account |
| ACCOUNT_TYPE_INCOME | Income category account |
| ACCOUNT_TYPE_ADJUSTMENT | Adjustment account |
