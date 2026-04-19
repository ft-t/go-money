# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Repository

**Go Money** — self-hosted personal finance manager. Go backend (ConnectRPC) + Angular 19 / PrimeNG 19 frontend. PostgreSQL via GORM. Lua scripting for transaction rules. Grafana for reporting. Embedded MCP server exposed at `/mcp`.

## Layout

```
cmd/                    entrypoints (server, mcp-client, sync-exchange-rates, jwt-key-generator)
pkg/                    domain packages — each owns its interfaces.go + generated mocks
frontend/               Angular 19 SPA (PrimeNG + Tailwind)
docs/                   AI-oriented docs — start at docs/INDEX.md (schema, business-logic, mcp, analytics)
build/                  Dockerfiles + SAM for AWS exchange-rate lambda
compose/                local docker-compose stacks
helm/                   k8s chart
.mcp.json               project-scoped MCP servers (angular-cli, primeng)
```

Server wiring lives in `cmd/server/main.go` — it constructs every service with explicit deps (no DI container). Read it first to trace how a new dep threads through.

## Commands

### Go
```
make lint                            # golangci-lint
make generate                        # regenerate mocks (go generate ./...)
make test                            # AUTO_CREATE_CI_DB=true go test ./...
go test -p 1 -timeout 60s ./pkg/...  # single package (respect -p 1 for DB tests)
make build-docker                    # server image
make mcp-client                      # build bin/mcp-client
make key-gen                         # build bin/jwt-key-generator
make update-pb                       # pull latest go-money-pb from buf
```

Tests hit a **real** Postgres. Set `AUTO_CREATE_CI_DB=true` or provide `Db_Host`/`ReadonlyDb_Host`/`Redis_Host` env vars. Use `-p 1` to serialize DB-touching tests.

### Frontend (`cd frontend`)
```
npm start         # ng serve on :4200
npm run build     # ng build → dist/go-money
npm test          # karma + jasmine
npm run format    # prettier
```

### Local backend deploy (docker-compose)

Local backend runs from `compose/docker-compose-backend.yaml` (uses `.env` in the same dir). To rebuild the Go binary and restart:

```
make build-docker                                              # build go-money-server:latest
cd compose && docker compose -f docker-compose-backend.yaml down
cd compose && docker compose -f docker-compose-backend.yaml up -d
```

Verify: `docker ps` should show `app` bound to the configured `GRPC_PORT`. Migrations run on app boot.

### Local browser testing (Claude-in-Chrome)

1. Ensure `npm start` is running (port 4200). If port busy, likely already running — skip.
2. Backend is assumed already reachable on `http://localhost:52057` (ssh tunnel or local binary).
3. Open `http://localhost:4200` in a Chrome tab.
4. Point the frontend at the backend by setting the cookie, then **refresh**:
   ```js
   window.cookieStore.set("customApiHost", "http://localhost:52057")
   ```
   Refresh the page after this — the transport is wired on app init.
5. Login with `test` / `test`.
6. Exercise the feature.

Notes:
- The app reads the cookie via `document.cookie` (ngx-cookie-service). `cookieStore.set` defaults to `secure:true` which hides the cookie from `document.cookie` on plain `http://localhost`. If login or API calls still go to `:4200` after refresh, re-set with `document.cookie = "customApiHost=http://localhost:52057; path=/; samesite=lax"` and refresh.
- `mcp__claude-in-chrome__read_network_requests` confirms the correct host is hit — look for `localhost:52057/gomoneypb....`.
- **After `npm install` of `@buf/...` pb** — restart ng serve. Webpack caches module resolution at startup; new proto fields on existing messages silently get stripped on the wire (the request body becomes `{}`), because the old schema in the bundle has no field descriptor for them.

## Architecture Notes

- **Protobuf contracts** live externally at `buf.build/xskydev/go-money-pb`. Never edit generated pb code locally — run `make update-pb` to bump.
- **Transport**: ConnectRPC (gRPC + JSON) via `boilerplate.NewDefaultGrpcServerBuild`. Handlers in `cmd/server/internal/handlers/`. Auth middleware in `cmd/server/internal/middlewares/`.
- **Double-entry ledger**: every transaction produces `double_entries` rows. Logic in `pkg/transactions/double_entry/`. Amounts carry both native + base-currency fields — base currency set via `CurrencyConfig.BaseCurrency`.
- **Transaction pipeline**: validation → applicable-account resolution → Lua rules → double-entry post → stats update. See `pkg/transactions/service.go` + `docs/business-logic/transactions/`.
- **Lua rules**: `pkg/transactions/rules/` — interpreter uses `gopher-lua` + `gopher-luar`. Scheduler re-inits at boot from DB rows.
- **MCP server** (embedded, Go side): `pkg/mcp/` exposes read-only SQL `query` tool + tag/category/rule/currency/transaction tools, mounted at `/mcp` when `MCP.Disable=false`. Docs loaded from `MCP.DocsDir` at startup.
- **Importers**: Firefly, Privat24 (xlsx), Mono, Paribas, Revolut — all share `importers.BaseParser`.
- **Mocks**: `//go:generate mockgen` produces `*_mocks_test.go` (package-local) using `github.com/golang/mock/gomock`. Never hand-edit. Regenerate after interface changes.
- **Frontend**: Angular 19 standalone-style, PrimeNG 19 components, Tailwind + `tailwindcss-primeui`, ConnectRPC client, Monaco for Lua editing. App shell in `frontend/src/app/layout/`, pages in `frontend/src/app/pages/`. **Read `docs/frontend-architecture.md` before touching `frontend/`** — documents shared patterns (per-page config, transport wiring, MCP server usage).

## Docs System

`docs/INDEX.md` is the entry point — it indexes schema tables, business-logic flows, analytics query patterns, and MCP tool specs with keywords. Consult it before guessing DB column meanings or transaction semantics.

## MCP Tool Usage Guide

Two project-scoped MCP servers are wired in `.mcp.json`. Prefer them over ad-hoc `npm`/grep when working in `frontend/`.

### `angular-cli` — when to use
- **Mandatory first call** when touching `frontend/`: `list_projects` (discovers workspace path) then `get_best_practices` (version-specific standards for Angular 19).
- Conceptual Angular questions ("what is a signal", "how do injectors work"): `search_documentation`.
- Code snippets for a feature ("show standalone component with route"): `find_examples`.
- Migrating to zoneless / OnPush: `onpush_zoneless_migration`.
- **Skip** for: pure TypeScript/RxJS, PrimeNG-specific work, backend work.

### `primeng` — when to use
- Before adding/editing any PrimeNG component in templates or TS.
- Look up component API before writing: `get_component`, `get_component_props`, `get_component_events`, `get_component_slots`.
- Verify a prop name/type before using it: `validate_props`.
- Finding the right component: `suggest_component`, `search_components`, `find_by_prop`, `find_by_event`, `find_components_with_feature`.
- Theming / styling / passthrough: `get_theming_guide`, `get_tailwind_guide`, `get_passthrough_guide`, `get_component_tokens`.
- Working examples: `get_example`, `get_usage_example`, `list_examples`.
- Migration: `migrate_v18_to_v19` (pinned at v19 — do not run newer migrations without approval).
- **Skip** for: non-PrimeNG UI libs, pure Angular framework questions (use `angular-cli` instead).

Combine the two: Angular structural questions → `angular-cli`; PrimeNG component details → `primeng`.
