# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this repo is

Fork of [Wei-Shaw/sub2api](https://github.com/Wei-Shaw/sub2api) optimized for **personal / single-node** use:

| Fork capability | Detail |
|-----------------|--------|
| SQLite | Only supported database; `backend/migrations/*.sql` is **SQLite dialect** |
| Redis optional | `redis.enabled: false` → in-process miniredis (single node only) |
| Hide menus | Setting `hidden_menu_keys` + admin page `/admin/menu` |
| Usage billing | `usage_billing_dedup` + indexes so successful requests still write usage |
| Simple mode | `run_mode: simple` weakens SaaS/billing UI and hides several sidebar items |

This fork is SQLite-only and prioritizes 1C1G native deploy. Upstream multi-instance production still uses PostgreSQL + external Redis. Deeper notes: [REFACTOR.md](./REFACTOR.md), [DEV_GUIDE.md](./DEV_GUIDE.md), [REMOVED_PAGES.md](./REMOVED_PAGES.md), [deploy/START_NATIVE.md](./deploy/START_NATIVE.md).

**Merging upstream:** git history was rewritten (no shared ancestor). Do **not** `git merge upstream/main`. Cherry-pick or port by feature. Process + current checklists: [docs/upstream-sync/README.md](./docs/upstream-sync/README.md), [docs/upstream-sync/PORTING-0.2.4.md](./docs/upstream-sync/PORTING-0.2.4.md) (newest intake, 2026-09-12: v0.2.1–v0.2.4 including the v0.2.2 tag; 13 第一档 items + 12 第二档 clusters **landed on main and verified** at `a26b6b55d`, including the API Key Responses namespace follow-up fix; release target `v1.1.12`; OpenSpec changes, frozen intake evidence and final implementation report linked there; earlier leftovers rechecked), [docs/upstream-sync/PORTING-0.2.0.md](./docs/upstream-sync/PORTING-0.2.0.md) (0.1.185 + 0.2.0 first/second-tier main batches merged into main at `c9d9bebe8`; DOMPurify still pending), [docs/upstream-sync/PORTING-0.1.184.md](./docs/upstream-sync/PORTING-0.1.184.md) (0.1.184, P0/P1 merged), [docs/upstream-sync/PORTING-0.1.183.md](./docs/upstream-sync/PORTING-0.1.183.md) (0.1.181–0.1.183 bugfixes), [docs/upstream-sync/PORTING-0.1.180.md](./docs/upstream-sync/PORTING-0.1.180.md).

**Go version:** `1.27.0` (from `backend/go.mod`). CI asserts this string; bump go.mod and workflow version checks together.

**Frontend package manager:** **pnpm only** (not npm). Commit `frontend/pnpm-lock.yaml` after dependency changes. pnpm v11 needs `frontend/pnpm-workspace.yaml` `allowBuilds` for `esbuild` / `vue-demi` postinstall.

## Upstream release intake (run this whenever a new upstream release lands)

Triggered by "check upstream", "evaluate the new release", or a new tag showing on <https://github.com/Wei-Shaw/sub2api/releases>. Produce artifacts, not a chat-only answer.

**1. Fix the baseline.** `git fetch upstream --tags --prune`; list `<last-evaluated-tag>..<newest-tag>` non-merge commits. The last round's *upstream main* may already sit inside the next tag, so part of the range can already be ported — run a **reverse** `git apply --check` over the whole candidate set first (`ALREADY` state) before the forward one.

**2. Judge per candidate, four states per file**: `CLEAN` (forward applies) / `ALREADY` (reverse applies = already here) / `CONFLICT` / `NOBASE`. For a feature upstream split into many micro-commits, take the whole PR diff (`git diff <merge>^1 <merge>`), never the individual commits. Then, for every new call the patch introduces, `grep -rn "func .*<symbol>"` — three states never prove it compiles. `CONFLICT` may mean the changed function does not exist here, i.e. **this fork does not have the bug**; verify before filing it as work.

**3. Rank into four tiers** (this is the priority model — keep the tier numbers, downstream artifacts reference them):

| Tier | Meaning | Artifact |
|---|---|---|
| 第一档 | Confirmed live defect here, cheap, low risk | OpenSpec change |
| 第二档 | Worth taking, real work (manual hunks / migration / frontend) | OpenSpec change |
| 第三档 | On-demand — only pays off if the feature is actually used | **Ask the user, do not start** |
| 第四档 | Do not merge (PG-only, needs an absent platform, bug not applicable) | **Ask the user, do not start** |

Never silently promote 第三档/第四档 into work. If measuring a base changes a tier (a "missing base" that turns out to be 11 lines), say so explicitly rather than moving it quietly.

**4. Write `docs/upstream-sync/PORTING-<version>.md`**, same shape as the existing ones: §1 version table + per-cluster stats, §2 method (channels / four-state / migration numbering), §3 P0 with patch sites and base-symbol greps, §4 P1, §5 not-doing with the judgement basis, §6 N/A-or-already-merged, §7 landing order, §8 self-test, §9 new general lessons. Then update `docs/upstream-sync/README.md` (top checklist + lessons list) and the "Merging upstream" pointer above.

**5. Report the older rounds' leftovers.** Every time, restate what is still unmigrated or undecided in the earlier `PORTING-*.md` files and why (取决于使用方式 / 架构上不需要 / 需要新增迁移或改 schema), so an old "needs a base" label does not get inherited unmeasured for another round. The oldest-parked item is usually the one nobody has re-measured.

**6. Generate an OpenSpec change per 第一档 / 第二档 batch** under `openspec/changes/port-upstream-<version>-<slug>/`, following the newest existing change file-for-file: `README.md`, `proposal.md`, `source-baseline.md` (SHAs frozen — never edited during implementation), `source-feature-map.md`, `design.md`, `specs/<capability>/spec.md` (ADDED Requirements + Scenarios), `tasks.md`, `verification.md`. Patch sites live only in the PORTING doc; the change defines the behaviour that must hold and the acceptance evidence. Leave every checkbox unchecked and every evidence slot empty until the work is actually done.

**7. Verification is part of the intake, not a follow-up.** Confirm the tests a doc points at actually exist and pass on the current baseline (`go test -list`), and `head -1` any `_integration_test.go` — a batch of them here are `//go:build integration && postgres` and never compile in this fork.

## Common commands

### Frontend (`frontend/`)

```bash
cd frontend
pnpm install
pnpm run dev              # Vite dev server
pnpm run build            # vue-tsc -b && vite build → backend/internal/web/dist
pnpm run lint:check
pnpm run typecheck
pnpm run test:run         # all vitest
pnpm exec vitest run path/to/file.spec.ts   # single test file
```

Root Makefile critical frontend suite:

```bash
make test-frontend-critical
```

### Backend (`backend/`)

```bash
cd backend
go build -o bin/server ./cmd/server
go test ./...
go test -tags=unit ./...
go test -tags=integration ./...
go test -tags=unit ./internal/repository/ -run 'TestName' -count=1   # single test
golangci-lint run ./...   # CI uses v2.13
```

Codegen after schema / wire changes:

```bash
cd backend
go generate ./ent          # after backend/ent/schema/*.go changes — commit generated ent/
go generate ./cmd/server   # Wire DI → wire_gen.go
# or: make -C backend generate
```

### 1C1G native single binary (do not build on 1G VPS)

```bash
cd frontend && pnpm install && pnpm run build && cd ..
cd backend
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
  go build -tags embed -ldflags="-s -w" \
  -o ../sub2api ./cmd/server
```

- `-tags embed` embeds `backend/internal/web/dist` (see `internal/web/embed_on.go`).
- Root `/sub2api` is gitignored; deploy artifact only.
- Personal config template: `deploy/config.personal.sqlite.yaml`
- systemd unit: `deploy/sub2api-sqlite.service`

Docker (SQLite compose): see README / `deploy/docker-compose.sqlite.yml`.

### Root

```bash
make build                # backend + frontend
make test                 # backend tests + frontend lint/typecheck/critical vitest
make -C backend test-unit
```

## Architecture

```
cmd/server          Entry, flags (-setup, -version), setup wizard vs main server
internal/config     Viper config; run_mode, database.driver, redis.enabled
internal/setup      First-run install (CLI / web / AUTO_SETUP); creates admin user once
internal/server     Gin router, middleware, route registration
internal/handler    HTTP handlers (admin, auth, gateway, payment, …)
internal/service    Business logic
internal/repository Ent client + raw SQL repos; migrations runner; SQLite aux tables
internal/web        Embedded SPA (build tag embed)
ent/schema          Ent schemas → generated ent/*
migrations/         Ordered SQL migrations (currently SQLite dialect)
```

**DI:** Google Wire (`cmd/server/wire.go` → `wire_gen.go`). Prefer regenerating rather than hand-editing `wire_gen.go`.

**Request path (API gateway):** client → Gin gateway handlers → account/group scheduling + upstream providers (`internal/service`, `internal/platform`) → usage/billing write path (dedup tables matter on SQLite).

**Frontend:** Vue 3 + Pinia + Vue Router + Tailwind. Admin views under `frontend/src/views/admin/`, user views under `views/user/`. Sidebar filters by `run_mode` (`hideInSimpleMode`) and `hidden_menu_keys`. API clients under `frontend/src/api/`.

**Auth / admin bootstrap:**

- `default.admin_email` / `default.admin_password` in YAML are **not** live account updates. Login uses `users` table (bcrypt). Setup / `AUTO_SETUP` creates admin once; later password changes need DB or admin UI (UI may be hidden in simple mode).
- Direct URL `/admin/users` still works in simple mode even when sidebar hides it; router blocks other simple-restricted paths (`/admin/groups`, subscriptions, redeem, …).

## SQLite-specific constraints (this fork)

1. **Migrations are SQLite SQL.** PG features (partitioning, trgm, plpgsql, some backfills) are no-ops (`SELECT 1`). Do not assume PG-only SQL runs. Converter: `backend/scripts/pg_sql_to_sqlite.py`.
2. **Timestamps:** use `DATETIME` (not `TEXT`) for `*_at` columns so `modernc.org/sqlite` can scan into `time.Time`. DSN must include `_time_format=sqlite` (`DatabaseConfig.sqliteDSN`, setup DSN builder). **Readers must not assume the write-side rule held**: existing DBs may carry `TEXT` columns created by older builds (aux table vs migration disagreement), and modernc returns `string` for those, so `Scan(*time.Time)` fails. Scan such columns type-agnostically (`scanSchedulerOutboxTime`).
3. **Statement splitting:** `splitSQLStatements` in `migrations_runner.go` must ignore `;` inside comments/strings. Naive split breaks large files (e.g. `033_ops_monitoring_vnext.sql`).
4. **Checksum immutability:** changing an already-applied migration file fails checksum on existing DBs. Prefer new migration files; for personal wipe-and-recreate, deleting `*.db` is acceptable.
5. **Safety-net tables:** `EnsureSQLiteAuxTables` creates tables some no-op migrations omitted (e.g. `user_allowed_groups`). Login loads allowed groups — missing table → 503 on login. Its DDL must match the migration column-for-column: both use `IF NOT EXISTS`, whichever runs first wins (usually the aux table), so a mismatch makes the migration file misleading.
6. **Embedded Redis:** fine for one process; multi-instance needs real Redis.
7. **Disabling a background service is not a SQLite adaptation.** `skipSQLiteBackgroundJobs` (`internal/service/wire.go`) skips a service's whole `Start()`. Justify it per-SQL, never by "we run SQLite" — a skipped service throws no error, so the feature just silently stops working (this caused the `c35a482` scheduler outage). Still skipped and unaudited: `UsageCleanupService`, `AccountExpiryService`, `ScheduledTestRunnerService` (so `auto_pause_on_expired` does not fire). Verified clean as of 2026-08-16: no `COPY` / `pg_dump` / PG-only SQL in production code, and `$1`-style placeholders are native SQLite parameter syntax, not a PG remnant. PG→SQLite dialect reference: [docs/upstream-sync/README.md §5](./docs/upstream-sync/README.md).

## Development pitfalls (repo-specific)

- Changing a Go interface → update all test stubs/mocks (`DEV_GUIDE` 坑 6).
- Ent schema change without `go generate ./ent` + committing `ent/` → dead code paths.
- Frontend: if `pnpm install` blocks on build scripts, set `allowBuilds.esbuild` / `vue-demi` to `true` in `frontend/pnpm-workspace.yaml`.
- bcrypt hashes contain `$` — shell/PowerShell expands them; write SQL/scripts via files.
- Bulk-editing accounts across platforms can wipe model mappings (OpenAI Codex, etc.); group bulk ops by platform (`DEV_GUIDE` 坑 10).

## Config touchpoints

| Concern | Where |
|---------|--------|
| Personal 1C1G defaults | `deploy/config.personal.sqlite.yaml` |
| Full example | `deploy/config.example.yaml` |
| Data dir | `DATA_DIR` env (config + sqlite path + install lock) |
| Install lock | `$DATA_DIR/.installed` + `config.yaml` → skips setup wizard |
| Simple mode | `run_mode: simple` |
| Hide ops / batch image | `ops.enabled`, `batch_image.enabled` |

## Releases / tags

This fork uses its **own** `1.1.x` numbers. Do **not** set `VERSION` or tags to upstream `0.1.x`.

**Source of truth**

- File: `backend/cmd/server/VERSION` (plain `X.Y.Z`, one line)
- `backend/scripts/resolve-version.sh`: exact `vX.Y.Z` checkout → tag without `v`; otherwise the file. `deploy/deploy-remote.sh` bakes that into the binary.

**Sequence (do not reorder)**

1. Set `backend/cmd/server/VERSION` to the new number and **commit** it (so the tag commit itself carries the matching file). `v1.1.8` tagged `f8f257ad7` while the file was still `1.1.2` — do not repeat that.
2. Annotated tag only (`git tag -a`), never lightweight.
3. `git push origin main` then `git push origin vX.Y.Z`. Pushing `v*` runs `.github/workflows/release.yml`.

**Annotated tag body (Chinese, same shape as `v1.1.9`)**

Relative to the previous tag SHA. Always include:

1. **同步上游** — Wei-Shaw/sub2api versions included, and **what landed** (P0/P1/decisions, with this-fork commit SHAs). Split by upstream version (`v0.1.180`, `v0.1.181–v0.1.183`, …).
2. **本仓库自上一 tag 起另外新增** — fork-only docs, VERSION alignment, lint cleanup, etc.
3. **明确未合** — deferred/skipped clusters so the tag is not read as “we took the whole upstream release” (tool-bridge, Grok 429/Realtime, monitor-v2 composite, frontend lockfile upgrades, plugins, CN providers, …).

Do not rewrite historical PORTING / OpenSpec notes that record “VERSION was 1.1.8 at port time”.

## Admin CLI skill

Repo skill `skills/sub2api-admin` wraps admin HTTP API (`node scripts/sub2api-admin.js …`) with `SUB2API_BASE_URL` + `SUB2API_ADMIN_API_KEY` or `SUB2API_JWT`. Prefer it over ad-hoc curl for account/group bulk ops.
