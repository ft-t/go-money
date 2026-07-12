# Transaction Clone and Safe Retry Design

## Goal

Replace the existing Split action with explicit New transaction and Clone transaction actions. Prevent duplicate transactions when a multi-transaction save partially succeeds and the user retries.

## Scope

This change affects the transaction upsert page and its component tests. Existing transaction APIs remain unchanged. Saving remains sequential and stops at the first server failure.

## User Interface

Each transaction card exposes these actions:

- New transaction appends a fully blank transaction.
- Clone transaction appends a new transaction populated from the current card's live form values.

Clone copies every editable field, including transaction type, accounts, amounts, currencies, date, title, notes, category, tags, and reference fields. It excludes ID, creation and update timestamps, history, and other server-managed state.

New and cloned transactions can be removed until their first successful save.

## Client State

The page records a normalized saved snapshot for each transaction that succeeds during the current save session. The snapshot represents the request accepted by the server.

After a successful create, the page replaces the local transaction with the transaction returned by the server. This immediately applies its server ID and prevents a later retry from creating it again.

Card states are:

- Pending: new or not yet attempted.
- Saved: server request succeeded and current normalized form matches its saved snapshot.
- Unsaved changes: request previously succeeded but current form differs from its saved snapshot.
- Failed: most recent request failed.

## Save Flow

Before sending any request, the page marks every editor as touched and validates every form. Any client validation failure prevents all server calls.

The page then processes transactions in display order:

1. Build the normalized request and snapshot from the editor's current form.
2. Skip a transaction whose snapshot matches its last successful snapshot.
3. Create transactions without server IDs.
4. Update transactions with server IDs.
5. On success, apply the returned server transaction and store the successful snapshot.
6. On failure, mark that card failed, stop processing, and leave the page open.

The Save button is disabled while processing. This prevents concurrent saves and double-click duplicates.

After every transaction succeeds or is skipped, existing post-save navigation runs. After a partial failure, all cards and entered values remain available for correction and retry.

## Retry Behavior

On retry:

- Unchanged successful transactions are skipped.
- Edited successful transactions are updated.
- The failed transaction is retried.
- Transactions after the failed transaction are attempted only after it succeeds.

This prevents duplicate creates without adding idempotency keys or changing backend APIs.

## Errors and Feedback

A partial failure identifies both progress and failed card, for example: `1 transaction saved. Transaction 2 failed: <reason>`.

Successfully saved cards display Saved. Editing such a card changes its indicator to Unsaved changes. The failed card receives a visible error state and remains editable. Pending cards remain unchanged.

A create response without a transaction or assigned ID is treated as a failure. An update response without a transaction is also treated as a failure. Prior failure state is cleared when a new save attempt begins.

## Testing

Component tests use a mocked transaction service and cover:

- New transaction creates a fully blank card.
- Clone copies all editable live form values and removes server-managed fields.
- Successful create applies the returned ID and stores a saved snapshot.
- Retry skips unchanged successful transactions.
- Editing a successful transaction causes an update on retry.
- Failure stops later requests and preserves form values.
- Save is disabled during an active request.
- Missing returned transaction or ID is handled as failure.
- Full success retains existing navigation behavior.

Tests do not call real APIs.

## Out of Scope

- Backend or protobuf changes.
- Atomic multi-transaction upsert.
- Idempotency keys.
- Changes to transaction import behavior.
