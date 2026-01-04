# Authentication

Go Money uses JWT (JSON Web Tokens) with RSA-256 signing for authentication.

## Overview

| Feature | Description |
|---------|-------------|
| Algorithm | RS256 (RSA with SHA-256) |
| Token Format | JWT (RFC 7519) |
| Token Types | Web tokens (short-lived), Service tokens (long-lived) |
| Header | `Authorization: Bearer <token>` |

## Token Types

### Web Tokens

Short-lived tokens for interactive sessions.

| Property | Value |
|----------|-------|
| TTL | 24 hours |
| Token Type | `web` |
| Use Case | Web UI, interactive sessions |
| Refresh | Re-login required |

### Service Tokens

Long-lived tokens for programmatic access.

| Property | Value |
|----------|-------|
| TTL | Custom (set at creation) |
| Token Type | `service_token` |
| Use Case | API integrations, automation, MCP server |
| Revocable | Yes |

## JWT Claims

```json
{
  "exp": 1704067200,
  "jti": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": 1,
  "token_type": "web"
}
```

| Claim | Type | Description |
|-------|------|-------------|
| exp | number | Expiration timestamp (Unix) |
| jti | string | Unique token identifier (UUID) |
| user_id | int32 | User ID |
| token_type | string | `web` or `service_token` |

## Authentication Flow

### Initial Login

```
POST /gomoneypb.users.v1.UsersService/Login
```

Request:
```json
{
  "login": "admin",
  "password": "secret"
}
```

Response:
```json
{
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

### Using the Token

Include the token in the Authorization header:

```
Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...
```

## User Management

### First User Creation

The first user is created via the Create endpoint. This only works when no users exist.

```
POST /gomoneypb.users.v1.UsersService/Create
```

Request:
```json
{
  "login": "admin",
  "password": "secure_password"
}
```

Response:
```json
{
  "id": 1
}
```

**Note:** After the first user is created, this endpoint returns an error.

### Password Storage

Passwords are hashed using bcrypt with a cost factor of 5.

## Service Tokens

### Creating a Service Token

Requires authentication.

```
POST /gomoneypb.configuration.v1.ConfigurationService/CreateServiceToken
```

Request:
```json
{
  "name": "MCP Integration",
  "expires_at": "2025-12-31T23:59:59Z"
}
```

Response:
```json
{
  "service_token": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "MCP Integration",
    "expires_at": "2025-12-31T23:59:59Z",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**Important:** The `token` field is only returned once at creation time. Store it securely.

### Listing Service Tokens

```
POST /gomoneypb.configuration.v1.ConfigurationService/GetServiceTokens
```

Request:
```json
{
  "ids": []
}
```

Response:
```json
{
  "service_tokens": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "name": "MCP Integration",
      "expires_at": "2025-12-31T23:59:59Z",
      "created_at": "2024-01-15T10:30:00Z"
    }
  ]
}
```

### Revoking a Service Token

```
POST /gomoneypb.configuration.v1.ConfigurationService/RevokeServiceToken
```

Request:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

Response:
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000"
}
```

## Token Revocation

Service tokens can be revoked before expiration. The system maintains a revocation list.

### Revocation Process

1. Token JTI is added to `jti_revocations` table
2. Revocation entry expires 7 days after original token expiry
3. Token validation checks revocation list with 5-minute cache

### Revocation Cache

| Setting | Value |
|---------|-------|
| Cache Size | 1000 entries |
| Cache TTL | 5 minutes |
| Storage | In-memory LRU |

## Database Tables

### users

| Column | Type | Description |
|--------|------|-------------|
| id | integer | Primary key (auto-increment) |
| login | text | Username |
| password | text | Bcrypt hash |
| created_at | timestamp | Creation time |
| deleted_at | timestamp | Soft delete marker |

### service_tokens

| Column | Type | Description |
|--------|------|-------------|
| id | uuid | Primary key (token JTI) |
| name | text | User-provided name |
| expires_at | timestamp | Expiration time |
| deleted_at | timestamp | Soft delete (revocation) |
| created_at | timestamp | Creation time |

### jti_revocations

| Column | Type | Description |
|--------|------|-------------|
| id | text | Primary key (JTI) |
| expires_at | timestamp | Entry expiration (7 days after token expiry) |

## Error Codes

| Error | Connect Code | Description |
|-------|--------------|-------------|
| Invalid token | UNAUTHENTICATED | Token parsing or validation failed |
| Token expired | UNAUTHENTICATED | Token past expiration |
| Token revoked | UNAUTHENTICATED | Service token was revoked |
| Permission denied | PERMISSION_DENIED | Valid token but missing user_id |

## Security Considerations

### Private Key Management

- RSA private key is configured via `JWT_PRIVATE_KEY` environment variable
- If not provided, a temporary key is generated (not recommended for production)
- Key format: PEM-encoded PKCS#1 RSA private key

### Best Practices

1. **Use HTTPS** - Always use TLS in production
2. **Rotate Keys** - Periodically rotate RSA keys
3. **Short TTL for Web** - 24-hour TTL limits exposure
4. **Revoke Unused Tokens** - Regularly audit and revoke service tokens
5. **Secure Storage** - Store service tokens securely (secrets manager, encrypted storage)

## Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| JWT_PRIVATE_KEY | RSA private key (PEM) | Auto-generated |

### Key Generation

To generate a new RSA key:

```bash
openssl genrsa -out private.pem 2048
```

Then set the environment variable:

```bash
export JWT_PRIVATE_KEY="$(cat private.pem)"
```

## API Protocol

Go Money uses [Connect](https://connectrpc.com/) protocol (gRPC-compatible HTTP/JSON).

### Request Format

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"key": "value"}' \
  http://localhost:8080/gomoneypb.users.v1.UsersService/Login
```

### Unauthenticated Endpoints

| Endpoint | Description |
|----------|-------------|
| UsersService/Login | User authentication |
| UsersService/Create | First user creation |
| ConfigurationService/GetConfiguration | Public configuration |

### Authenticated Endpoints

All other endpoints require a valid JWT token in the Authorization header.
