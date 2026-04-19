# Frontend Architecture

Shared patterns for the Angular 19 / PrimeNG 19 frontend. Read this before touching `frontend/`.

## Transport

ConnectRPC client built once in `frontend/src/app/consts/transport.ts` and injected via the `TRANSPORT_TOKEN` opaque DI token. All service-level clients are created with `createClient(XxxService, transport)`. The backend host is read from the `customApiHost` cookie at bootstrap — see `CLAUDE.md` "Local browser testing" for setup.

## Per-Page Configuration

Any page that needs to persist small UI customizations (filters, saved searches, quick shortcuts, pinned items) uses the **per-page configuration** pattern. No new RPCs, no new tables.

### Storage

Reuses the existing `AppConfig` table behind `ConfigurationService.GetConfigsByKeys` / `SetConfigByKey`:

- **Key format:** `page.<page-id>` — flat namespace, kebab-case id (e.g. `page.accounts-list`).
- **Value:** opaque JSON string. Shape owned by the consuming page.
- **Scope:** **global** — one row per page, shared across all users. Any authenticated user may read and write.
- **No versioning.** Schema evolution is the page's responsibility; shallow default-merge provides forward-compat for added fields.

### Frontend helper

`frontend/src/app/services/page-config.service.ts` — `PageConfigService`:

```typescript
async get<T>(pageId: string, defaults: T): Promise<T>
async set<T>(pageId: string, value: T): Promise<void>
```

- `get` returns `{ ...defaults, ...parsed }` (shallow merge). Parse failures log and fall back to defaults. RPC errors **propagate** — callers decide how to handle them.
- `set` stringifies and writes. RPC errors propagate.

### Page-side conventions

1. **Colocate types with the component:** `<page>.config.ts` next to `<page>.component.ts`.
2. **Export a page-id constant** (`ACCOUNTS_LIST_PAGE_ID = 'accounts-list'`) — never a raw string literal.
3. **Export defaults** (`ACCOUNTS_LIST_DEFAULTS`). Ship empty collections, not opinionated seed data — global scope means seeds leak to every user.
4. **Load failures must not break the page.** Wrap the `get()` call in try/catch; log and fall back to a fresh spread copy of defaults. Optional feature — silent fallback on load.
5. **Save failures toast.** Use `MessageService` with `severity: 'error'`. Users must know when their change didn't persist.
6. **Never mutate in place.** Build new `pageConfig` with spread/filter, then call `savePageConfig()`.

### Adding a new page config

1. Create `<page>.config.ts` with the interface, page-id constant, and defaults.
2. In the component: inject `PageConfigService`, declare `pageConfig` initialized from defaults, add `loadPageConfig()` (guarded) and `savePageConfig()` (toasts on failure).
3. Call `loadPageConfig()` in `ngOnInit` — after critical loaders (auth/server config) and before render-driving loaders.
4. Document the new key under "Existing keys" in `docs/api/endpoints.md` so it stays discoverable.

### Reference implementation

- Service: `frontend/src/app/services/page-config.service.ts`
- Types: `frontend/src/app/pages/accounts/accounts-list.config.ts`
- Consumer: `frontend/src/app/pages/accounts/accounts-list.component.ts` (quick-tag chips)

## MCP servers for frontend work

Two project-scoped MCP servers (`.mcp.json`) must be consulted before writing component code:

- **`angular-cli`** — call `list_projects` then `get_best_practices` first. Use for Angular-19-specific patterns (signals, standalone, zoneless, OnPush).
- **`primeng`** — consult before adding or editing any PrimeNG component. `get_component`, `get_component_props`, `validate_props` catch prop typos and version drift.

Skip these only for pure TypeScript / RxJS / backend work.
