# Go Money

![build workflow](https://github.com/ft-t/go-money/actions/workflows/general.yaml/badge.svg?branch=master)
[![codecov](https://codecov.io/gh/ft-t/go-money/graph/badge.svg?token=pas79tP0Dr)](https://codecov.io/gh/ft-t/go-money)
[![go-report](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)](https://img.shields.io/badge/go%20report-A+-brightgreen.svg?style=flat)
[![PkgGoDev](https://pkg.go.dev/badge/github.com/ft-t/go-money)](https://pkg.go.dev/github.com/ft-t/go-money?tab=doc)

**Go Money** is an open-source personal finance manager written in Go, built for speed, flexibility, and advanced customization.

Unlike most finance apps, Go Money is designed for more advanced users who want full control over their transaction data and storage.

It enables customizations through Lua scripting and external reporting with Grafana.

## Key Features

- Multi-currency transactions with base-currency tracking
- Double-entry ledger for audit-grade bookkeeping
- Custom Lua hooks + transaction rules engine to auto-tag, categorize, enrich transactions
- Grafana-based reporting (bring your own dashboards)
- Tags, categories, hierarchical accounts, daily stats
- CSV/XLSX imports: Firefly III, Monobank, Privat24, Paribas, Revolut
- Embedded **MCP server** for AI agents — read-only SQL + domain tools exposed at `/mcp`
- Scriptable and developer-friendly architecture
- High test coverage and stable API
- Multiple client libraries via [ConnectRPC](https://buf.build/xskydev/go-money-pb/sdks/main:protobuf)
- Prebuilt multi-arch Docker images and standalone binaries (linux/darwin/windows)

## Demo
A demo instance of Go Money is available at [https://demo.go-money.top](https://demo.go-money.top) and grafana dashboards at [https://grafana.go-money.top](https://grafana.go-money.top).

Login credentials for the demo instance:
- **Username**: `demo`
- **Password**: `demo4vcxsdfss231`

`Note`: The demo instance is reset every 3 hours, so any data you enter will be lost after that time, also possible downtime during the reset process.

`Note2`: The demo instance is not intended for production use, it is provided for demonstration purposes only.

`Note3`: The demo instances is running on cheapest 1$ VPS, so it may be slow or unstable at times.

## Installation

Go Money is available as:

- **Docker image** — `ghcr.io/ft-t/go-money/go-money-full:latest` (bundles backend + UI).
- **Helm chart** — published to `gh-pages` branch.
- **Standalone binaries** — `go-money-server` + `go-money-mcp-client` for linux/darwin (amd64 + arm64) and windows (amd64), attached to each [GitHub Release](https://github.com/ft-t/go-money/releases).

For detailed installation instructions, please refer to the [Installation guide](https://github.com/ft-t/go-money/wiki/Installation).

## UI
GO Money provides a simple web UI for managing transactions, accounts, and other financial data.

## API
Go Money provides multi-protocol API (gRPC, JSON-RPC) for more details and documentation, please refer to the [API documentation](https://github.com/ft-t/go-money/wiki/Api)

## Reporting
Go Money does not come with built-in reports. Instead, it allows you to use Grafana to create custom dashboards and reports based on your transaction data.

[Grafana guide](https://github.com/ft-t/go-money/wiki/Grafana)

[Database schema](https://github.com/ft-t/go-money/wiki/Database-structure-and-entities-rules)

[Query examples](https://github.com/ft-t/go-money/tree/master/docs/reporting/queries)

[//]: # ([Grafana dashboards]&#40;https://github.com/ft-t/go-money/tree/master/docs/reporting/dashboards&#41;.)

## Scripting
Go Money runs Lua per transaction (via [gopher-lua](https://github.com/yuin/gopher-lua)) to enrich, re-tag, re-categorize, split, or annotate based on your own logic. Scripts live in the DB and hot-reload — no restarts.

Rules engine details: [docs/business-logic/rules-engine/overview.md](docs/business-logic/rules-engine/overview.md).

[Lua scripting guide](https://github.com/ft-t/go-money/wiki/Lua)

[Lua scripts examples](https://github.com/ft-t/go-money/tree/master/docs/lua)

## Imports
Bring your existing data from other finance apps. Supported formats:

- [Firefly III](https://firefly-iii.org/) export
- Monobank statement
- Privat24 xlsx statement
- Paribas xlsx statement
- Revolut statement

See [pkg/importers](https://github.com/ft-t/go-money/tree/master/pkg/importers) for the parsers.

## MCP Integration (AI agents)

Go Money ships an embedded [Model Context Protocol](https://modelcontextprotocol.io) server at `/mcp`. Connect any MCP-compatible AI agent (Claude Desktop, Claude Code, Cursor, Windsurf, Zed, etc.) and query your finances in natural language — "how much did I spend on groceries last month?", "what's my net worth trend?", "auto-tag my Uber transactions".

### Tools exposed

| Tool | Purpose |
|------|---------|
| `query` | Read-only SQL over the full ledger. Agent-facing docs loaded at boot; see [Query Safety](docs/mcp/query-safety.md). |
| Tag tools | `list_tags`, `create_tag`, `update_tag`, `delete_tag`. |
| Category tools | `list_categories`, `create_category`, `update_category`, `delete_category`. |
| Rule tools | `list_rules`, `create_rule`, `update_rule`, `delete_rule`, `test_rule` — manage Lua transaction rules. |
| Currency tools | `list_currencies`, `upsert_currency`. |
| Transaction tools | `list_transactions`, `create_transaction`. |

### Quick start (Claude Desktop / Claude Code)

1. Create a **service token** in *Settings → Service Tokens*.
2. Download `go-money-mcp-client` from the [latest release](https://github.com/ft-t/go-money/releases).
3. Add to your agent's MCP config:

   ```json
   {
     "mcpServers": {
       "go-money": {
         "command": "go-money-mcp-client",
         "args": ["-server", "https://your-go-money-host/mcp/"],
         "env": { "GOMONEY_TOKEN": "<paste-token>" }
       }
     }
   }
   ```

4. Restart the agent. Tools appear under the `go-money` namespace.

### Documentation

- [MCP overview](docs/mcp/overview.md) — purpose, query safety, example SQL.
- [MCP client setup](docs/mcp/client-setup.md) — full `go-money-mcp-client` flags, env vars, troubleshooting.
- [MCP tool reference](docs/mcp/tool-reference.md) — per-tool request/response spec.
- [MCP golden rules](docs/mcp/GOLDEN-RULES.md) — must-read for agents before generating queries.
- [MCP examples](docs/mcp/examples.md) — 30+ natural-language → SQL mappings.

## Multi currency support
Go Money supports multiple currencies, allowing you to manage transactions in different currencies seamlessly.

Withdrawal transaction always has source_currency set to the account currency, but at same time you can set foreign currency for the transaction, so you can track exchange rates and conversions.

Go Money stores additional fields to track amounts in primary currency, so its much easier to work with that information in reports. For more details refer to [Currencies guide](https://github.com/ft-t/go-money/wiki/Currencies)

## Documentation

- [Wiki](https://github.com/ft-t/go-money/wiki) — user-facing guides (install, config, Lua, Grafana).
- [`docs/`](docs/) — architecture, schema, business logic, MCP. Start at [`docs/INDEX.md`](docs/INDEX.md).
- [`docs/frontend-architecture.md`](docs/frontend-architecture.md) — Angular SPA patterns.
- [API documentation](https://github.com/ft-t/go-money/wiki/Api) — ConnectRPC (gRPC + JSON-RPC).
