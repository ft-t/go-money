# MCP Client Setup

How to connect an AI agent (Claude Desktop, Claude Code, Cursor, etc.) to the Go Money MCP server.

## Background

Go Money's MCP server is an **HTTP** endpoint at `/mcp` (StreamableHTTP transport). Most local AI agents speak MCP over **stdio** only. Bridge the two with `go-money-mcp-client` — a small Go binary that proxies stdio calls to the remote HTTP server, forwarding a service token for auth.

```
Agent (stdio) ──► go-money-mcp-client ──HTTPS──► Go Money server /mcp
```

## Server-side prerequisites

1. Go Money server is running with MCP enabled. Default — it is. To disable, set:
   ```
   MCP_DISABLE=true
   ```
2. Create a **service token** (the auth credential the bridge will use). Use the Configuration API or the web UI under *Settings → Service Tokens*. The token is a JWT — copy it **once**; it cannot be retrieved later.

Configuration env vars (see `pkg/configuration/types.go`):

| Variable            | Default   | Purpose                                    |
|---------------------|-----------|--------------------------------------------|
| `MCP_DISABLE`       | `false`   | Turn the embedded MCP server on/off.       |
| `MCP_DOCS_DIR`      | `./mcp`   | Folder with MCP guidance docs loaded at boot (shipped with Docker image). |

## Install the client

### From GitHub Releases

Each release publishes `go-money-mcp-client_<version>_<os>_<arch>.{tar.gz|zip}` for:

- `linux_amd64`, `linux_arm64`
- `darwin_amd64`, `darwin_arm64`
- `windows_amd64`

Grab the archive from <https://github.com/ft-t/go-money/releases>, extract, and place the binary on your `PATH`.

### From source

```bash
git clone https://github.com/ft-t/go-money
cd go-money
make mcp-client          # builds ./bin/mcp-client
```

Or directly:

```bash
go install github.com/ft-t/go-money/cmd/mcp-client@latest
```

## CLI flags

```
go-money-mcp-client \
  -server http://localhost:8080/mcp/ \
  -token  <service-token>            \
  -header "X-Something: value"       # optional, repeatable
```

| Flag       | Default                         | Description                                                 |
|------------|---------------------------------|-------------------------------------------------------------|
| `-server`  | `http://localhost:8080/mcp/`    | Full URL to the Go Money MCP endpoint. Include trailing `/`. |
| `-token`   | — (required)                    | Service token. Falls back to `$GOMONEY_TOKEN` when unset.   |
| `-header`  | none (repeatable)               | Extra HTTP header, `Key: Value` format.                     |
| `-version` | —                               | Print version and exit.                                     |

`$GOMONEY_TOKEN` lets you keep the secret out of shell history and config files.

## Wiring into agents

### Claude Desktop / Claude Code

Add to your MCP config (`~/Library/Application Support/Claude/claude_desktop_config.json` on macOS, or `~/.claude/mcp_servers.json` for Claude Code):

```json
{
  "mcpServers": {
    "go-money": {
      "command": "go-money-mcp-client",
      "args": [
        "-server", "https://your-go-money-host/mcp/"
      ],
      "env": {
        "GOMONEY_TOKEN": "eyJhbGciOi..."
      }
    }
  }
}
```

Restart the agent. Tools should appear under the `go-money` server.

### Cursor / Windsurf / other stdio MCP clients

Any client that accepts a stdio MCP server command works the same way — point `command` at the binary, pass `-server` in `args`, put the token in `env.GOMONEY_TOKEN`.

## Available tools

The bridge forwards every tool exposed by the server. Current set:

- `query` — read-only SQL against the Postgres DB (see [tool-reference.md](tool-reference.md)).
- Tags: `list_tags`, `create_tag`, `update_tag`, `delete_tag`.
- Categories: `list_categories`, `create_category`, `update_category`, `delete_category`.
- Rules: `list_rules`, `create_rule`, `update_rule`, `delete_rule`, `test_rule`.
- Currencies: `list_currencies`, `upsert_currency`.
- Transactions: `list_transactions`, `create_transaction`.

See [tool-reference.md](tool-reference.md) for the authoritative per-tool spec. See [GOLDEN-RULES.md](GOLDEN-RULES.md) for agent guidance before issuing queries.

## Troubleshooting

| Symptom                                      | Cause / fix                                                                 |
|----------------------------------------------|-----------------------------------------------------------------------------|
| `service token is required`                  | Set `-token` or `$GOMONEY_TOKEN`.                                           |
| `failed to create transport: ... x509`       | TLS cert not trusted. Import the CA into your OS trust store.               |
| 401 Unauthorized                             | Token expired, revoked, or wrong server. Regenerate a service token.        |
| Client connects but no tools                 | Server's `MCP_DISABLE=true`. Toggle it off and restart.                     |
| Tools missing after server upgrade           | Restart the agent — stdio client caches tool list on handshake.             |
| `invalid header format`                      | `-header` must be `Key: Value` with the colon.                              |

## Security notes

- Service tokens inherit permissions of the user that created them — scope them accordingly.
- Rotate tokens periodically. Revoke compromised ones via the Configuration API.
- `query` is read-only (SELECT only); mutation tools respect the token's permissions.
- All agent activity is logged server-side for audit.
