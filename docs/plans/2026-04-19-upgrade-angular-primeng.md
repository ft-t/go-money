# Upgrade Angular, PrimeNG, and All Frontend Dependencies

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Upgrade `frontend/` from Angular 19 + PrimeNG 19 to the latest stable (Angular 21.2.9, PrimeNG 21.1.6) plus every other transitive dev/runtime dependency, while guaranteeing zero known vulnerabilities (`npm audit` + `trivy`) before any compilation or runtime work.

**Architecture:**
**Interleaved major hops**: Angular 19→20, then PrimeNG 19→20, then Angular 20→21, then PrimeNG 20→21. Interleaving matters because PrimeNG major N has a peer-dependency on Angular major N — bumping both Angular majors before touching PrimeNG would leave the install in a broken peer-dep state. Angular moves via `ng update` schematics; PrimeNG moves via the `primeng` MCP migration tools (`migrate_v19_to_v20`, `migrate_v20_to_v21`). All remaining deps bumped via `npm-check-updates`. Every stage gated on a clean `npm audit` + `trivy fs` scan — **compilation is blocked until both reports show zero vulnerabilities and the user approves**.

**Tech Stack:** Angular 21, PrimeNG 21, Tailwind 3, ConnectRPC, TypeScript 5.x, Node (current LTS), trivy 0.69+, npm 11.

**Branch:** `upgrade-angular` (created off `master`).

**Non-goals:**
- No Go backend changes.
- No feature/behavior changes.
- No PrimeNG theme redesign — keep current `@primeng/themes` config, only update to the new API where migration demands it.
- No TypeScript strict-mode tightening beyond what Angular schematics apply.

---

## Ground Rules (apply to every task)

1. **Never run `ng build`, `ng serve`, or `ng test` until audits pass and user approves.** If either `npm audit` or `trivy fs` reports a vulnerability at any severity ≥ *low*, STOP, print both reports, wait for user instruction.
2. Commit after every green task. Conventional Commits. Signed (`git commit -S`).
3. Never edit generated files (`package-lock.json` is OK to regenerate; `node_modules/` never committed).
4. Always cd into `frontend/` for npm/ng/ncu commands.
5. Consult `primeng` MCP (`get_component`, `validate_props`, `get_migration_guide`) before manually patching any PrimeNG template/TS changes flagged by the migration tool.
6. Consult `angular-cli` MCP (`get_best_practices`, `search_documentation`) for any schematic failure or idiomatic question.

---

## Task 1: Create upgrade branch

**Files:** none (git only).

**Step 1: Verify clean tree**

Run: `git status`
Expected: `nothing to commit, working tree clean` on `master`.

**Step 2: Create and switch to branch**

Run:
```bash
git checkout -b upgrade-angular
```
Expected: `Switched to a new branch 'upgrade-angular'`.

**Step 3: Confirm**

Run: `git branch --show-current`
Expected: `upgrade-angular`.

*(No commit yet — branch only.)*

---

## Task 2: Baseline security audit (BEFORE any change)

**Files:**
- Create: `docs/plans/artifacts/2026-04-19-audit-baseline-npm.json`
- Create: `docs/plans/artifacts/2026-04-19-audit-baseline-trivy.json`

**Step 1: Create artifacts dir**

Run: `mkdir -p docs/plans/artifacts`

**Step 2: Capture `npm audit` baseline**

Run:
```bash
cd frontend && npm audit --json > ../docs/plans/artifacts/2026-04-19-audit-baseline-npm.json || true
```
`|| true` because npm audit exits non-zero when vulns exist. We want the report regardless.

**Step 3: Capture `trivy fs` baseline**

Run (from repo root):
```bash
trivy fs --scanners vuln --include-dev-deps --format json --output docs/plans/artifacts/2026-04-19-audit-baseline-trivy.json frontend/
```
`--include-dev-deps` is required: default behavior suppresses dev-dep CVEs, which would make pre/post comparison misleading and would let the user's "zero vulnerabilities" requirement be satisfied while karma/webpack transitives still carry known issues.

**Step 4: Summarize both reports**

Run:
```bash
jq '.metadata.vulnerabilities' frontend/../docs/plans/artifacts/2026-04-19-audit-baseline-npm.json
jq '[.Results[]?.Vulnerabilities[]?] | group_by(.Severity) | map({severity: .[0].Severity, count: length})' docs/plans/artifacts/2026-04-19-audit-baseline-trivy.json
```
Print both summaries to console for the user.

**Step 5: Commit baseline reports**

```bash
git add docs/plans/artifacts/2026-04-19-audit-baseline-*.json docs/plans/2026-04-19-upgrade-angular-primeng.md
git commit -S -m "chore(frontend): capture baseline npm audit + trivy reports"
```

---

## Task 3: Upgrade Angular 19 → 20

**Files:**
- Modify: `frontend/package.json`
- Modify: `frontend/package-lock.json`
- Modify: any `*.ts`/`*.html` files that schematics rewrite.

**Step 1: Consult Angular MCP for v20 upgrade notes**

Invoke `mcp__angular-cli__search_documentation` with query `"v19 to v20 migration"` and `mcp__angular-cli__get_best_practices` with `workspacePath=frontend/`. Record any manual migration steps.

**Step 2: Run Angular schematics**

```bash
cd frontend
npx ng update @angular/core@20 @angular/cli@20 --force=false
```
If schematics prompt interactively, re-run non-interactively:
```bash
npx --yes @angular/cli@20 update @angular/core@20 @angular/cli@20
```

**Step 3: Resolve schematic reported changes**

- Read `ng update` stdout. Apply any manual migration it points out.
- Run `mcp__angular-cli__onpush_zoneless_migration` only if schematics require it.

**Step 4: Commit**

```bash
cd ..
git add frontend/package.json frontend/package-lock.json frontend/
git commit -S -m "chore(frontend): ng update @angular/core @angular/cli to v20"
```

*(No build yet. Build is gated behind Task 8 audit + approval.)*

---

## Task 4: Upgrade PrimeNG 19 → 20

**Files:**
- Modify: `frontend/package.json` (`primeng`, `@primeng/themes`, `primeicons`, `tailwindcss-primeui`).
- Modify: any `*.html`/`*.ts` files flagged by the migration tool.

**Why now (after Angular 20, before Angular 21):**
PrimeNG 20 has a peer-dep on `@angular/core ^20`. Running the PrimeNG migration before Angular 20 would fail; running it after Angular 21 would install against an unsupported peer. Sync majors.

**Step 1: Run migration tool**

Invoke `mcp__primeng__migrate_v19_to_v20`. Read migration output — lists renamed components, token changes, removed props.

**Step 2: Bump PrimeNG packages**

```bash
cd frontend
npm install primeng@20 @primeng/themes@20 primeicons@latest tailwindcss-primeui@latest
```

**Step 3: Fix any migration-flagged code**

- For each flagged component/prop: call `mcp__primeng__get_component` and `mcp__primeng__validate_props` before editing.
- Use `replace_text_in_file` (JetBrains MCP) for edits.

**Step 4: Commit**

```bash
cd ..
git add frontend/package.json frontend/package-lock.json frontend/src
git commit -S -m "chore(frontend): migrate primeng v19 to v20"
```

---

## Task 5: Upgrade Angular 20 → 21

**Files:** same surface as Task 3.

**Step 1: Consult Angular MCP for v21 upgrade notes**

`mcp__angular-cli__search_documentation` with `"v20 to v21 migration"`.

**Step 2: Run Angular schematics**

```bash
cd frontend
npx ng update @angular/core@21 @angular/cli@21
```

**Step 3: Resolve schematic-reported changes**

Apply any manual migration flagged by `ng update` output.

**Step 4: Commit**

```bash
cd ..
git add frontend/package.json frontend/package-lock.json frontend/
git commit -S -m "chore(frontend): ng update @angular/core @angular/cli to v21"
```

---

## Task 6: Upgrade PrimeNG 20 → 21

**Files:** same surface as Task 4.

**Why now (after Angular 21):**
PrimeNG 21 peer-dep on `@angular/core ^21`. Sync majors again.

**Step 1: Run migration tool**

Invoke `mcp__primeng__migrate_v20_to_v21`.

**Step 2: Bump PrimeNG packages**

```bash
cd frontend
npm install primeng@21 @primeng/themes@21 primeicons@latest tailwindcss-primeui@latest
```

**Step 3: Fix any migration-flagged code**

Same procedure as Task 4 Step 3.

**Step 4: Commit**

```bash
cd ..
git add frontend/package.json frontend/package-lock.json frontend/src
git commit -S -m "chore(frontend): migrate primeng v20 to v21"
```

---

## Task 7: Upgrade remaining dependencies

**Files:** `frontend/package.json`, `frontend/package-lock.json`.

**Step 1: Install `npm-check-updates` (transient)**

```bash
cd frontend
npx --yes npm-check-updates@latest --doctor --upgrade --target=latest \
  --reject="@angular/*,@angular-devkit/*,primeng,@primeng/themes,primeicons,tailwindcss-primeui"
```
`--reject` list: already bumped above; don't let ncu override. `--doctor` runs an install+test loop — **if tests trigger here, skip `--doctor`** and run plain `ncu -u` instead:
```bash
npx --yes npm-check-updates@latest -u --target=latest \
  --reject="@angular/*,@angular-devkit/*,primeng,@primeng/themes,primeicons,tailwindcss-primeui"
npm install
```

**Step 2: Review the diff**

Run: `git diff package.json`
Print the diff. Flag any major-version jumps beyond Angular/PrimeNG so the user can approve before continuing.

**Step 3: Remove suspicious packages**

`package.json` currently pins `install@^0.13.0` and `npm@^11.4.2` as dependencies — these are *almost certainly accidental* (publishing an app with `npm` as a runtime dep is wrong). Flag to user: propose removal. DO NOT remove without explicit "yes".

**Step 4: Commit**

```bash
cd ..
git add frontend/package.json frontend/package-lock.json
git commit -S -m "chore(frontend): bump remaining deps to latest"
```

---

## Task 7b: Prune unused dependencies

**Files:** `frontend/package.json`, `frontend/package-lock.json`.

**Step 1: Static import scan**

Run `depcheck` non-destructively:
```bash
cd frontend
npx --yes depcheck --json > /tmp/depcheck.json
cat /tmp/depcheck.json | jq '{unused: .dependencies, unusedDev: .devDependencies, missing: .missing}'
```
`depcheck` flags deps whose specifiers never appear in `src/` imports. False positives possible — validate each.

**Step 2: Cross-check each flagged package before removal**

For every package `X` flagged as unused:
```bash
Grep(pattern="X", path="frontend/src", output_mode="files_with_matches")
Grep(pattern="X", path="frontend/angular.json")
Grep(pattern="X", path="frontend/eslint.config.js")
```
Only propose removal if all three return empty AND the package is not a build-tool peer (e.g., `postcss`, `autoprefixer`, `@types/*`).

**Step 3: Known suspects (flag proactively to user)**

Based on current `package.json` inspection, these are strong candidates for removal — ASK before dropping each:
- `install@^0.13.0` — the `install` CLI is never imported from Angular app code; almost certainly accidental.
- `npm@^11.4.2` — npm itself as a runtime dep is wrong.
- `util.promisify@^1.1.3` — native `util.promisify` exists in Node; in browser code this is typically dragged in by a polyfill. Verify with grep first.
- `primeclt@^0.1.5` — obscure low-version package; verify actual usage.
- `highlightjs-copy@^1.0.6` — only needed if `ngx-highlightjs` copy button is wired; grep first.

**Step 4: Propose removal list to user**

Print the final list as:
```
Proposed removals:
  <name>@<version>  (reason: <why>)
Keep anyway? [list to keep]
```
Wait for explicit "yes, drop all" or per-package ack.

**Step 5: Remove approved packages**

```bash
cd frontend
npm uninstall <pkg1> <pkg2> ...
```

**Step 6: Verify build graph intact**

```bash
npm ls --all > /dev/null
```
Expected: no `UNMET DEPENDENCY` errors. If errors appear, the removal was premature — reinstall the flagged package and revisit.

**Step 7: Commit**

```bash
cd ..
git add frontend/package.json frontend/package-lock.json
git commit -S -m "chore(frontend): drop unused dependencies"
```

---

## Task 8: Post-upgrade security audit (GATE — no compile before this passes)

**Files:**
- Create: `docs/plans/artifacts/2026-04-19-audit-post-npm.json`
- Create: `docs/plans/artifacts/2026-04-19-audit-post-trivy.json`

**Step 1: Regenerate lockfile cleanly**

```bash
cd frontend
rm -rf node_modules
npm install
```
(Clean install surfaces any peer-dep conflicts the earlier incremental installs hid.)

**Step 2: Capture `npm audit`**

```bash
npm audit --json > ../docs/plans/artifacts/2026-04-19-audit-post-npm.json || true
npm audit
```
Second call is human-readable for the report shown to user.

**Step 3: Capture `trivy fs`**

```bash
cd ..
trivy fs --scanners vuln --include-dev-deps --format json --output docs/plans/artifacts/2026-04-19-audit-post-trivy.json frontend/
trivy fs --scanners vuln --include-dev-deps --severity LOW,MEDIUM,HIGH,CRITICAL frontend/
```
`--include-dev-deps` MUST match the baseline flag (Task 2 Step 3) for the comparison to be valid.
Second call prints human-readable table.

**Step 4: Summarize**

```bash
jq '.metadata.vulnerabilities' docs/plans/artifacts/2026-04-19-audit-post-npm.json
jq '[.Results[]?.Vulnerabilities[]?] | group_by(.Severity) | map({severity: .[0].Severity, count: length})' docs/plans/artifacts/2026-04-19-audit-post-trivy.json
```

**Step 5: Gate decision**

- If both reports show **zero vulnerabilities of any severity**: proceed to Task 9 (show user + ask to compile).
- If **any vulnerability** remains:
  - Print both reports in full.
  - STOP.
  - Open a sub-task: for each vuln, check if there is a fixed version upstream; if yes, pin it. If not, report to user with severity + CVE IDs and ask whether to accept risk, add `overrides`, or abort.
  - Do NOT run `npm audit fix --force` automatically — it can downgrade peers.

**Step 6: Commit reports**

```bash
git add docs/plans/artifacts/2026-04-19-audit-post-*.json
git commit -S -m "chore(frontend): capture post-upgrade npm audit + trivy reports"
```

---

## Task 9: Show reports, ask user, then compile (ONLY on approval)

**Step 1: Present reports**

Output to user, in this exact order:
1. Baseline npm audit summary (Task 2).
2. Baseline trivy summary (Task 2).
3. Post-upgrade npm audit summary (Task 8).
4. Post-upgrade trivy summary (Task 8).
5. `git diff master -- frontend/package.json` (show the full version delta).
6. Any manually-applied migration edits list.

**Step 2: Explicit ask**

Verbatim prompt:
> "Audits clean (or listed above). Ready to compile (`ng build`) and run tests (`ng test`). Approve? (yes/no)"

**Step 3: On "yes" — compile**

```bash
cd frontend
npx ng build --configuration production
```
Expected: build succeeds. Record any warnings.

**Step 4: On "yes" — run unit tests**

```bash
npx ng test --watch=false --browsers=ChromeHeadless
```
Expected: all pass.

**Step 5: On "no" — STOP**

Do nothing further. Leave branch as-is for user inspection.

**Step 6: Commit build/test artifacts if required**

Usually none. If any files changed (e.g., generated config), commit:
```bash
cd ..
git add -A frontend/
git commit -S -m "chore(frontend): apply post-upgrade config updates"
```

---

## Task 10: Smoke test in browser (only after Task 9 approved)

**Files:** none.

**Step 1: Start dev server**

```bash
cd frontend && npm start
```
Leave running.

**Step 2: Ensure backend reachable**

Per project `CLAUDE.md` local-testing section: backend assumed on `http://localhost:52057`.

**Step 3: Use Claude-in-Chrome MCP**

- `mcp__claude-in-chrome__tabs_context_mcp` — snapshot current tabs.
- `mcp__claude-in-chrome__tabs_create_mcp` with `url=http://localhost:4200` — open app.
- Set cookie:
  ```js
  document.cookie = "customApiHost=http://localhost:52057; path=/; samesite=lax"
  ```
- Refresh page.
- Login `test` / `test`.
- Navigate: Dashboard → Transactions → Rules (Monaco/Lua editor) → Reports. Verify no console errors via `mcp__claude-in-chrome__read_console_messages`.

**Step 4: Capture GIF of walkthrough**

`mcp__claude-in-chrome__gif_creator` → save as `upgrade-smoke-test.gif` under `docs/plans/artifacts/`.

**Step 5: Report result**

Summarize: pages that rendered cleanly, console errors (if any), visual regressions (if any). Ask user whether to open a PR or hand back to them.

---

## Task 11: Done-gate + handoff

**Checklist (all must be true):**
- [ ] Branch `upgrade-angular` holds every commit from Tasks 1–10.
- [ ] `npm audit` in `frontend/` reports `found 0 vulnerabilities`.
- [ ] `trivy fs frontend/` reports zero vulnerabilities at LOW and above.
- [ ] `npx ng build --configuration production` succeeds.
- [ ] `npx ng test --watch=false` all pass.
- [ ] Browser smoke test GIF attached.
- [ ] No files deleted or renamed outside schematic output.

**Handoff:** Do NOT push or open PR without explicit user approval (project rule: never push to `master`; even feature branch push requires confirmation per global `CLAUDE.md`).

Prompt user:
> "All tasks complete. Push `upgrade-angular` to origin and open PR? (yes/no)"

On "yes": `git push -u origin upgrade-angular` then `gh pr create`.
On "no": STOP.

---

## Rollback strategy

If anything irreversible breaks mid-upgrade:
```bash
git reset --hard master
git branch -D upgrade-angular
```
Only with explicit user approval (destructive). Prefer `git checkout master -- frontend/package.json frontend/package-lock.json && npm install` for a partial rollback of just the lockfile while keeping investigation commits.
