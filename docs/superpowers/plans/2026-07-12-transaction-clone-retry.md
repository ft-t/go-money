# Transaction Clone and Safe Retry Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add blank and clone transaction actions while preventing duplicate creates after a partial sequential save failure.

**Architecture:** `TransactionUpsertComponent` continues to own editor ordering and UI feedback. A focused `TransactionSaveSession` owns saved request snapshots, sequential create/update calls, response validation, and stop-on-first-failure behavior. `TransactionEditorComponent` exports a complete live draft so cloning does not depend on stale transaction inputs.

**Tech Stack:** Angular 21, TypeScript 5.9, Jasmine/Karma, Protobuf ES 2, Connect RPC client.

## Global Constraints

- Existing backend and protobuf APIs remain unchanged.
- Saving remains sequential and stops at first failure.
- New transaction is fully blank.
- Clone preserves every editable live form value, including `skipRules`, but removes server-managed identity and timestamps.
- Frontend tests mock service boundaries and perform no real API calls.
- Direct frontend npm test/lint/build commands are explicitly allowed for this task despite the root Makefile.

---

### Task 1: Sequential Save Session

**Files:**
- Create: `frontend/src/app/pages/transactions/transaction-save-session.ts`
- Test: `frontend/src/app/pages/transactions/transaction-save-session.spec.ts`

**Interfaces:**
- Consumes: generated `CreateTransactionRequest`, `Transaction`, and `UpdateTransactionRequestSchema`.
- Produces: `TransactionSaveSession`, `TransactionSaveWriter`, `TransactionSaveItem`, `TransactionSaveResult`, and `TransactionCardState`.

- [ ] **Step 1: Write failing save-session tests**

Cover these independent behaviors with Jasmine tests and complete protobuf messages:

```typescript
it('stores returned identity and skips unchanged create on retry', async () => {
    const writer = createWriter();
    writer.createTransaction.and.resolveTo(create(CreateTransactionResponseSchema, {
        transaction: create(TransactionSchema, { id: 41n })
    }));
    const session = new TransactionSaveSession(writer);
    const request = createRequest('first');
    const applied: bigint[] = [];

    await session.save([{ id: 0n, request }], (_, transaction) => applied.push(transaction.id));
    await session.save([{ id: 41n, request }], (_, transaction) => applied.push(transaction.id));

    expect(writer.createTransaction).toHaveBeenCalledTimes(1);
    expect(writer.updateTransaction).not.toHaveBeenCalled();
    expect(applied).toEqual([41n]);
    expect(session.getCardState(0, request)).toBe('saved');
});
```

Also test changed saved requests use update, first failure stops later items, create response without non-zero ID fails, update response without transaction fails, `isSaving` blocks concurrent execution, and `remove()` keeps index-aligned state.

- [ ] **Step 2: Run tests and verify RED**

Run:

```bash
cd frontend && npm test -- --watch=false --browsers=ChromeHeadless --include='src/app/pages/transactions/transaction-save-session.spec.ts'
```

Expected: compilation failure because `transaction-save-session.ts` does not exist.

- [ ] **Step 3: Implement minimal save session**

Implement these exact public shapes:

```typescript
export type TransactionCardState = 'pending' | 'saved' | 'unsaved' | 'failed';

export interface TransactionSaveWriter {
    createTransaction(request: CreateTransactionRequest): Promise<CreateTransactionResponse>;
    updateTransaction(request: UpdateTransactionRequest): Promise<UpdateTransactionResponse>;
}

export interface TransactionSaveItem {
    id: bigint;
    request: CreateTransactionRequest;
}

export interface TransactionSaveResult {
    savedCount: number;
    failedIndex?: number;
    error?: unknown;
}

export class TransactionSaveSession {
    public isSaving = false;
    private savedRequests: Array<CreateTransactionRequest | undefined> = [];
    private states: TransactionCardState[] = [];

    constructor(private readonly writer: TransactionSaveWriter) {}

    async save(items: TransactionSaveItem[], apply: (index: number, transaction: Transaction) => void): Promise<TransactionSaveResult>;
    getCardState(index: number, request: CreateTransactionRequest): TransactionCardState;
    remove(index: number): void;
}
```

Use `equals(CreateTransactionRequestSchema, saved, current)` for snapshot comparison. Clear old failure state at save start. Skip equal saved requests. Create when `id === 0n`; update otherwise. Validate create response has `transaction.id !== 0n`, validate update response has `transaction`, apply each successful response immediately, then clone/store the successful request with `clone(CreateTransactionRequestSchema, request)`. Catch the first error and return its index without attempting later items. Reset `isSaving` in `finally`.

- [ ] **Step 4: Run focused tests and verify GREEN**

Run the Step 2 command. Expected: all save-session tests pass.

- [ ] **Step 5: Commit save session**

```bash
git add frontend/src/app/pages/transactions/transaction-save-session.ts frontend/src/app/pages/transactions/transaction-save-session.spec.ts
git commit -S -m "fix(frontend): track partial transaction saves"
```

### Task 2: Live Transaction Draft Export

**Files:**
- Modify: `frontend/src/app/shared/components/transaction-editor/transaction-editor.component.ts`
- Test: `frontend/src/app/shared/components/transaction-editor/transaction-editor.component.spec.ts`

**Interfaces:**
- Consumes: existing editor `FormGroup` and `TransactionSchema`.
- Produces: `TransactionDraft`, `initialSkipRules`, and `buildTransactionDraft()`.

- [ ] **Step 1: Write failing draft tests**

Build a real `FormGroup`, attach it to an editor instance created with `Object.create(TransactionEditorComponent.prototype)`, then call the wished-for API:

```typescript
const draft = editor.buildTransactionDraft();

expect(draft.transaction.id).toBe(0n);
expect(draft.transaction.title).toBe('Clone me');
expect(draft.transaction.tagIds).toEqual([2, 5]);
expect(draft.transaction.internalReferenceNumbers).toEqual(['reference']);
expect(draft.skipRules).toBeTrue();
```

Assert all form-backed transaction fields: type, both accounts, both amounts and currencies, FX fields, date, title, notes, category, tags, and internal references. Separately test `buildForm()` seeds `skipRules` from `initialSkipRules`.

- [ ] **Step 2: Run tests and verify RED**

```bash
cd frontend && npm test -- --watch=false --browsers=ChromeHeadless --include='src/app/shared/components/transaction-editor/transaction-editor.component.spec.ts'
```

Expected: compilation failure because `TransactionDraft`, `initialSkipRules`, and `buildTransactionDraft()` do not exist.

- [ ] **Step 3: Implement draft export**

Add:

```typescript
export interface TransactionDraft {
    transaction: Transaction;
    skipRules: boolean;
}

@Input() public initialSkipRules = false;

buildTransactionDraft(): TransactionDraft {
    return {
        transaction: create(TransactionSchema, {
            id: 0n,
            sourceAmount: NumberHelper.toPositiveNumber(this.form.get('sourceAmount')!.value),
            sourceCurrency: this.form.get('sourceCurrency')!.value,
            sourceAccountId: this.form.get('sourceAccountId')!.value,
            destinationAmount: NumberHelper.toPositiveNumber(this.form.get('destinationAmount')!.value),
            destinationCurrency: this.form.get('destinationCurrency')!.value,
            destinationAccountId: this.form.get('destinationAccountId')!.value,
            notes: this.form.get('notes')!.value,
            title: this.form.get('title')!.value,
            categoryId: this.form.get('categoryId')!.value,
            transactionDate: create(TimestampSchema, TimestampHelper.dateToTimestamp(this.form.get('transactionDate')!.value)),
            type: this.form.get('type')!.value,
            tagIds: [...(this.form.get('tagIds')!.value || [])],
            fxSourceAmount: NumberHelper.toPositiveNumber(this.form.get('fxSourceAmount')!.value),
            fxSourceCurrency: this.form.get('fxSourceCurrency')!.value,
            internalReferenceNumbers: [...(this.form.get('internalReferenceNumbers')!.value || [])]
        }),
        skipRules: this.form.get('skipRules')!.value
    };
}
```

Initialize the form's `skipRules` control from `this.initialSkipRules` instead of hard-coded `false`.

- [ ] **Step 4: Run focused tests and verify GREEN**

Run the Step 2 command. Expected: all editor draft tests pass.

- [ ] **Step 5: Commit draft export**

```bash
git add frontend/src/app/shared/components/transaction-editor/transaction-editor.component.ts frontend/src/app/shared/components/transaction-editor/transaction-editor.component.spec.ts
git commit -S -m "feat(frontend): export transaction editor drafts"
```

### Task 3: Upsert UI and Retry Integration

**Files:**
- Modify: `frontend/src/app/pages/transactions/transactions-upsert.component.ts`
- Modify: `frontend/src/app/pages/transactions/transactions-upsert.component.html`
- Test: `frontend/src/app/pages/transactions/transactions-upsert.component.spec.ts`
- Include: `docs/superpowers/plans/2026-07-12-transaction-clone-retry.md`

**Interfaces:**
- Consumes: `TransactionSaveSession`, `TransactionEditorComponent.buildTransactionDraft()`, and existing transaction service client.
- Produces: `addTransaction()`, `cloneTransaction(index)`, retry-aware `saveAll()`, and card status rendering.

- [ ] **Step 1: Write failing component behavior tests**

Use a component created with `Object.create(TransactionUpsertComponent.prototype)`, real transaction arrays, and editor doubles that return real forms/requests. Cover:

```typescript
component.addTransaction();
expect(component.targetTransaction[1]).toEqual(create(TransactionSchema, {}));
expect(component.initialSkipRules[1]).toBeFalse();

component.cloneTransaction(0);
expect(component.targetTransaction[1]).toEqual(draft.transaction);
expect(component.initialSkipRules[1]).toBeTrue();
```

Also cover validation preventing all requests, partial failure toast text, values retained after failure, full-success navigation, and delete splicing transaction, skip-rules, and save-session state together.

- [ ] **Step 2: Run tests and verify RED**

```bash
cd frontend && npm test -- --watch=false --browsers=ChromeHeadless --include='src/app/pages/transactions/transactions-upsert.component.spec.ts'
```

Expected: compilation failure because new transaction, clone, and save-session integration APIs do not exist.

- [ ] **Step 3: Implement page integration**

Add `public initialSkipRules: boolean[] = [false]` and a declared `public readonly saveSession: TransactionSaveSession`. Initialize `saveSession` in the constructor immediately after `createClient(TransactionsService, this.transport)` assigns `transactionService`; a field initializer would run before that assignment and capture `undefined`. Replace `addSplit()` with:

```typescript
addTransaction(): void {
    this.targetTransaction.push(create(TransactionSchema, {}));
    this.initialSkipRules.push(false);
}

cloneTransaction(index: number): void {
    const editor = this.components.get(index);
    if (!editor) return;
    const draft = editor.buildTransactionDraft();
    this.targetTransaction.push(draft.transaction);
    this.initialSkipRules.push(draft.skipRules);
}
```

Refactor `saveAll()` to validate all editors first, build `TransactionSaveItem[]`, call `saveSession.save()`, immediately replace each successful `targetTransaction[index]`, and update `initialSkipRules[index]` from its successful request. On partial failure show `${savedCount} transaction(s) saved. Transaction ${failedIndex + 1} failed: ${ErrorHelper.getMessage(error)}` and stay on page. Navigate only when no failure exists. Guard entry with `saveSession.isSaving`.

Rename split deletion helpers to `canDeleteDraft()` and `deleteDraft()`, and splice `initialSkipRules` plus `saveSession.remove(index)` when deleting.

- [ ] **Step 4: Update template**

Replace Split button with New transaction and Clone transaction buttons. Pass `[initialSkipRules]="initialSkipRules[idx]"` into each editor. Add Saved, Unsaved changes, and Failed status text from `saveSession.getCardState(idx, editor.buildTransactionRequest())`. Disable Save while `saveSession.isSaving`. Keep existing commission, refresh, snippets, and server-delete controls unchanged.

- [ ] **Step 5: Run focused tests and verify GREEN**

Run the Step 2 command. Expected: all upsert component tests pass.

- [ ] **Step 6: Run frontend verification**

```bash
cd frontend && npm test -- --watch=false --browsers=ChromeHeadless
cd frontend && npm run lint
cd frontend && npm run build
```

Expected: tests, lint, and production build pass without errors.

- [ ] **Step 7: Run repository done gate**

Check root `config.dev.json`; when absent, use required database hosts. Run:

```bash
make lint
Db_Host=tools.lan ReadonlyDb_Host=tools.lan Redis_Host=tools.lan go test -p 1 -timeout 60s ./...
go build ./...
```

Expected: lint, Go tests, and Go build pass.

- [ ] **Step 8: Commit integration**

```bash
git add docs/superpowers/plans/2026-07-12-transaction-clone-retry.md frontend/src/app/pages/transactions/transactions-upsert.component.ts frontend/src/app/pages/transactions/transactions-upsert.component.html frontend/src/app/pages/transactions/transactions-upsert.component.spec.ts
git commit -S -m "feat(frontend): add safe transaction cloning"
```
