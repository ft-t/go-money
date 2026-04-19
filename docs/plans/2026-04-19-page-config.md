# Per-Page Configuration Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Introduce a reusable frontend `PageConfigService` that lets any Angular page persist its own JSON configuration (global, shared across users) via the existing `ConfigurationService` RPCs, and apply it to the Accounts List page as the first consumer with "quick tag" chips placed left of the search input.

**Architecture:** Zero backend changes. Reuse `AppConfig` table through existing `GetConfigsByKeys` / `SetConfigByKey` RPCs. New generic Angular service wraps the RPC client, handles JSON parse/serialize, defaults merge, and exposes `get<T>(pageId, defaults)` / `set<T>(pageId, value)`. Keys are flat `page.<page-id>` strings. JSON shape is owned by each consumer page (TypeScript interface + defaults colocated with the component).

**Tech Stack:** Angular standalone components, Connect-ES client for protobuf RPCs, PrimeNG components, existing `AppConfig` Postgres table.

**Non-goals:**
- No new proto messages or RPCs.
- No new database table or migration.
- No per-user scope (global only).
- No admin-only write enforcement (any authenticated user can write — same as current `SetConfigByKey`).
- No caching layer (fresh fetch on mount).
- No FE unit tests (repo currently has zero FE specs; manual browser verification only).

**Reference pattern:** `frontend/src/app/services/snippet.service.ts` already uses `ConfigurationService` with `getConfigsByKeys` / `setConfigByKey` and a hardcoded key (`transaction_snippets`). The new service generalizes exactly that shape. After this plan lands, `SnippetService` could later be refactored onto `PageConfigService`, but **out of scope here** — do not touch it.

---

## Task 1: Create `PageConfigService`

**Files:**
- Create: `frontend/src/app/services/page-config.service.ts`

**Step 1: Write the service**

```typescript
import { Inject, Injectable } from '@angular/core';
import { createClient, Transport } from '@connectrpc/connect';
import { ConfigurationService } from '@buf/xskydev_go-money-pb.bufbuild_es/gomoneypb/configuration/v1/configuration_pb';
import { TRANSPORT_TOKEN } from '../consts/transport';

@Injectable({ providedIn: 'root' })
export class PageConfigService {
    private readonly configService;

    constructor(@Inject(TRANSPORT_TOKEN) transport: Transport) {
        this.configService = createClient(ConfigurationService, transport);
    }

    async get<T>(pageId: string, defaults: T): Promise<T> {
        const key = this.buildKey(pageId);
        const response = await this.configService.getConfigsByKeys({ keys: [key] });
        const raw = response.configs[key];
        if (!raw) {
            return defaults;
        }
        try {
            const parsed = JSON.parse(raw) as Partial<T>;
            return { ...defaults, ...parsed };
        } catch (e) {
            console.error(`PageConfigService: failed to parse config for ${pageId}`, e);
            return defaults;
        }
    }

    async set<T>(pageId: string, value: T): Promise<void> {
        const key = this.buildKey(pageId);
        await this.configService.setConfigByKey({
            key,
            value: JSON.stringify(value),
        });
    }

    private buildKey(pageId: string): string {
        return `page.${pageId}`;
    }
}
```

**Notes for the implementer:**
- The `providedIn: 'root'` + `TRANSPORT_TOKEN` inject pattern is identical to `snippet.service.ts:9-25`. Do not deviate.
- The defaults merge is a **shallow** `{ ...defaults, ...parsed }`. This is intentional — pages with nested objects must handle deep merge themselves inside their own typed wrapper. YAGNI until a real consumer needs deep merge.
- `try/catch` only wraps `JSON.parse`. Do not catch RPC errors — callers decide how to surface them (matches `snippet.service.ts:44-49` style where network failures log and fall back).
- **Do not** add a `delete` method. YAGNI — no consumer yet.

**Step 2: Build the frontend to verify it compiles**

Run:
```bash
cd frontend && npm run build
```

Expected: build succeeds. No `PageConfigService`-related errors.

**Step 3: Commit**

```bash
git add frontend/src/app/services/page-config.service.ts
git commit -S -m "feat(frontend): add generic PageConfigService"
```

---

## Task 2: Define quick-tags config type for Accounts List

**Files:**
- Create: `frontend/src/app/pages/accounts/accounts-list.config.ts`

**Step 1: Write the model**

```typescript
export interface QuickTag {
    label: string;
    search: string;
}

export interface AccountsListConfig {
    quickTags: QuickTag[];
}

export const ACCOUNTS_LIST_PAGE_ID = 'accounts-list';

export const ACCOUNTS_LIST_DEFAULTS: AccountsListConfig = {
    quickTags: [],
};
```

**Notes for the implementer:**
- `QuickTag.label` is what's rendered on the chip.
- `QuickTag.search` is the text that gets pushed into the global filter input when the chip is clicked (mirrors `onGlobalFilter` in `accounts-list.component.ts`).
- Defaults is an empty array. First-run users see no chips and can add their own.
- Page ID is a **constant**, not a string literal scattered through the component — keeps the key stable and greppable.

**Step 2: Build to verify it compiles**

Run:
```bash
cd frontend && npm run build
```

Expected: build succeeds.

**Step 3: Commit**

```bash
git add frontend/src/app/pages/accounts/accounts-list.config.ts
git commit -S -m "feat(frontend): add accounts-list page config types"
```

---

## Task 3: Load and apply quick tags in `AccountsListComponent`

**Files:**
- Modify: `frontend/src/app/pages/accounts/accounts-list.component.ts`

**Step 1: Add imports and inject the service**

At the top of `accounts-list.component.ts`, add:

```typescript
import { PageConfigService } from '../../services/page-config.service';
import {
    AccountsListConfig,
    ACCOUNTS_LIST_DEFAULTS,
    ACCOUNTS_LIST_PAGE_ID,
    QuickTag,
} from './accounts-list.config';
```

**Step 2: Add component state**

Inside the `AccountsListComponent` class, alongside the other public fields (around line 87 where `serverConfig` is declared), add:

```typescript
public pageConfig: AccountsListConfig = { ...ACCOUNTS_LIST_DEFAULTS };
public editingQuickTags = false;
public newQuickTagLabel = '';
public newQuickTagSearch = '';
```

**Step 3: Inject `PageConfigService`**

Modify the constructor (currently lines 89-107) to add the service:

```typescript
constructor(
    @Inject(TRANSPORT_TOKEN) private transport: Transport,
    private messageService: MessageService,
    public router: Router,
    route: ActivatedRoute,
    private selectedDateService: SelectedDateService,
    private pageConfigService: PageConfigService,
) {
    // ... existing body unchanged
}
```

**Step 4: Load config on init**

Find the existing `async ngOnInit()` method (starts ~line 113). Add a `loadPageConfig()` call **before** `loadAccounts()`:

```typescript
async ngOnInit() {
    await this.loadConfig();
    await this.loadPageConfig();
    await this.loadAccounts();
    await this.loadAnalytics();
    // ... rest unchanged
}
```

Then add the method body elsewhere on the class:

```typescript
async loadPageConfig(): Promise<void> {
    try {
        this.pageConfig = await this.pageConfigService.get<AccountsListConfig>(
            ACCOUNTS_LIST_PAGE_ID,
            ACCOUNTS_LIST_DEFAULTS,
        );
    } catch (e) {
        console.error('Failed to load accounts-list page config:', e);
        this.pageConfig = { ...ACCOUNTS_LIST_DEFAULTS };
    }
}
```

Optional feature — load failures silently fall back to defaults; save failures toast.

**Step 5: Add quick-tag handlers**

Add these methods to the class:

```typescript
applyQuickTag(tag: QuickTag): void {
    this.filter.nativeElement.value = tag.search;
    this.table.filterGlobal(tag.search, 'contains');
}

async addQuickTag(): Promise<void> {
    const label = this.newQuickTagLabel.trim();
    const search = this.newQuickTagSearch.trim();
    if (!label || !search) {
        return;
    }
    this.pageConfig = {
        ...this.pageConfig,
        quickTags: [...this.pageConfig.quickTags, { label, search }],
    };
    this.newQuickTagLabel = '';
    this.newQuickTagSearch = '';
    await this.savePageConfig();
}

async removeQuickTag(index: number): Promise<void> {
    this.pageConfig = {
        ...this.pageConfig,
        quickTags: this.pageConfig.quickTags.filter((_, i) => i !== index),
    };
    await this.savePageConfig();
}

private async savePageConfig(): Promise<void> {
    try {
        await this.pageConfigService.set(ACCOUNTS_LIST_PAGE_ID, this.pageConfig);
    } catch (e) {
        console.error('Failed to save accounts-list page config:', e);
        this.messageService.add({
            severity: 'error',
            summary: 'Save failed',
            detail: ErrorHelper.getMessage(e),
        });
    }
}
```

**Notes for the implementer:**
- `applyQuickTag` uses the same `filterGlobal` / `nativeElement.value` pattern as the existing `onGlobalFilter` method — do not reinvent.
- `ErrorHelper.getMessage` is already imported in this file.
- `messageService` is already injected; reuse it.
- All mutations go through the save method — never mutate `pageConfig.quickTags` in place.

**Step 6: Build to verify it compiles**

Run:
```bash
cd frontend && npm run build
```

Expected: build succeeds. No type errors.

**Step 7: Commit**

```bash
git add frontend/src/app/pages/accounts/accounts-list.component.ts
git commit -S -m "feat(accounts): load and manage quick-tag page config"
```

---

## Task 4: Render quick tags in the Accounts List template

**Files:**
- Modify: `frontend/src/app/pages/accounts/accounts-list.component.html`

**Step 1: Replace the caption block**

Locate the existing `<ng-template #caption>` block (lines 40-60). The current structure is:

```html
<div class="flex justify-between items-center flex-column sm:flex-row gap-2">
    <p-iconfield iconPosition="left" class="ml-auto"> ... search ... </p-iconfield>
    <div class="flex gap-2"> ... Refresh / Create buttons ... </div>
</div>
```

Quick-tag chips must appear **to the left of the search input**. Replace the outer `<div>` body with:

```html
<div class="flex justify-between items-center flex-column sm:flex-row gap-2">
    <div class="flex items-center gap-2 flex-wrap">
        <p-button
            *ngFor="let tag of pageConfig.quickTags; let i = index"
            [label]="tag.label"
            severity="secondary"
            size="small"
            [outlined]="true"
            (click)="applyQuickTag(tag)">
        </p-button>
        <p-button
            icon="pi pi-pencil"
            severity="secondary"
            size="small"
            [text]="true"
            pTooltip="Edit quick tags"
            tooltipPosition="top"
            (click)="editingQuickTags = !editingQuickTags">
        </p-button>
    </div>

    <p-iconfield iconPosition="left">
        <p-inputicon>
            <i class="pi pi-search"></i>
        </p-inputicon>
        <input #filter pInputText type="text" (input)="onGlobalFilter(dt1, $event)"
               placeholder="Search keyword" />
        <p-button class="p-button-outlined mb-2 " icon="pi pi-filter-slash"
                  (click)="clear(dt1)"></p-button>
    </p-iconfield>

    <div class="flex gap-2">
        <p-button icon="pi pi-refresh" severity="secondary"
                  pTooltip="Refresh table" tooltipPosition="top"
                  (click)="refreshTable()" />
        <p-button label="Create new account" class="p-button-link"
                  (click)="this.router.navigate(['/accounts', 'new'])" />
    </div>
</div>

<div *ngIf="editingQuickTags" class="mt-3 border border-gray-200 dark:border-gray-700 rounded p-3">
    <div class="text-sm font-semibold mb-2">Quick tags</div>
    <div class="flex flex-col gap-2">
        <div *ngFor="let tag of pageConfig.quickTags; let i = index"
             class="flex items-center gap-2">
            <span class="min-w-[120px] font-medium">{{ tag.label }}</span>
            <span class="text-gray-500 text-sm flex-1">{{ tag.search }}</span>
            <p-button icon="pi pi-trash" severity="danger" size="small" [text]="true"
                      (click)="removeQuickTag(i)"></p-button>
        </div>
        <div class="flex items-center gap-2 pt-2 border-t border-gray-200 dark:border-gray-700">
            <input pInputText type="text" placeholder="Label"
                   class="w-32"
                   [(ngModel)]="newQuickTagLabel" />
            <input pInputText type="text" placeholder="Search text"
                   class="flex-1"
                   [(ngModel)]="newQuickTagSearch" />
            <p-button icon="pi pi-plus" severity="success" size="small"
                      (click)="addQuickTag()"></p-button>
        </div>
    </div>
</div>
```

**Notes for the implementer:**
- Removed `class="ml-auto"` from `p-iconfield` since chips now occupy that space. Layout is `[chips + edit] — [search] — [refresh/create]` via `justify-between`.
- The editor panel sits below the toolbar and is toggled by the pencil button.
- `[(ngModel)]` requires `FormsModule` — already imported in the component's `imports:` array (line 41).
- No new PrimeNG modules required — `ButtonModule`, `TooltipModule`, `InputTextModule`, `IconFieldModule` are already imported.

**Step 2: Build + serve; manually verify in the browser**

Run:
```bash
cd frontend && npm run build
```

Expected: build succeeds.

Then start the dev server and verify each behavior:

```bash
cd frontend && npm start
```

**Manual verification checklist (golden path + edges):**

1. Navigate to `/accounts`. No chips render (defaults empty). Pencil button is visible left of search.
2. Click pencil → editor panel appears with empty add-row.
3. Type label `Cash`, search `USD`, click `+`. Chip `Cash` appears. Panel lists the new row.
4. Reload the page. Chip persists. (Round-trips through `SetConfigByKey`.)
5. Click chip `Cash` → search input fills with `USD`, account table filters.
6. Open editor, click trash icon → chip removed. Reload → still removed.
7. Open a second browser (incognito, different user). Same chips visible — **confirms global scope**.
8. Open DevTools → Network. Reload → one RPC to `GetConfigsByKeys` with `["page.accounts-list"]`. No other new requests.
9. Use the existing "Clear" filter button — must still work (regression check).
10. Use existing refresh, create, row interactions — no regressions.

If any check fails, stop and investigate before committing.

**Step 3: Commit**

```bash
git add frontend/src/app/pages/accounts/accounts-list.component.html
git commit -S -m "feat(accounts): render quick-tag chips with inline editor"
```

---

## Task 5: Document the pattern for future pages

**Files:**
- Modify: `docs/api/endpoints.md`

**Step 1: Add a short subsection under the existing `SetConfigByKey` entry**

After the `SetConfigByKey` block (around line 146), add:

```markdown
### Per-Page Configuration Convention

The frontend uses `GetConfigsByKeys` / `SetConfigByKey` to persist **per-page UI configuration** under keys of the form `page.<page-id>`. The value is a JSON blob whose shape is owned by the page component (TypeScript interface colocated with the page). Configuration is **global** (shared across all users) and any authenticated user may write.

Frontend helper: `PageConfigService` (`frontend/src/app/services/page-config.service.ts`) — exposes `get<T>(pageId, defaults)` and `set<T>(pageId, value)`.

Existing keys:
- `page.accounts-list` — quick-tag chips on the Accounts List page.
```

**Step 2: Commit**

```bash
git add docs/api/endpoints.md
git commit -S -m "docs(api): document per-page configuration convention"
```

---

## Verification gate before finishing

- [ ] `cd frontend && npm run build` passes.
- [ ] Manual browser checklist (Task 4 Step 2) all green.
- [ ] No backend files modified — `git diff master -- pkg/ cmd/` is empty.
- [ ] No new proto, migration, or `AppConfig` schema change — `git diff master -- '*.proto' 'pkg/database/*'` is empty.
- [ ] `SnippetService` untouched — `git diff master -- frontend/src/app/services/snippet.service.ts` is empty.

Do not claim done until all five boxes are verified.
