# users Table

The `users` table stores user accounts for authentication.

## Schema

| Column | Type | Nullable | Default | Description |
|--------|------|----------|---------|-------------|
| id | integer | NO | auto-increment | Primary key |
| login | text | NO | - | Username for login |
| password | text | NO | - | Hashed password |
| created_at | timestamp | NO | - | Account creation time |
| deleted_at | timestamp | YES | - | Soft delete timestamp |

## Primary Key

- `id` (integer, auto-increment)

## Indexes

| Index | Definition | Purpose |
|-------|------------|---------|
| users_pk | UNIQUE (id) | Primary key |
| users_login_uindex | UNIQUE (login) WHERE deleted_at IS NULL | Unique usernames for active users |

## Security Notes

- Passwords are stored hashed (never plain text)
- The unique index on login prevents duplicate usernames
- Soft deletes allow account recovery while preventing username reuse

## Common Queries

### Find User by Login

```sql
SELECT id, login, created_at
FROM users
WHERE login = :username
  AND deleted_at IS NULL;
```

### List Active Users

```sql
SELECT id, login, created_at
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC;
```

### Check if Username Exists

```sql
SELECT EXISTS(
    SELECT 1 FROM users
    WHERE login = :username
      AND deleted_at IS NULL
) as username_exists;
```

## Notes

- This is a single-user system in typical deployment
- Password verification happens in application code
- Consider adding service tokens for API access
