# Source baseline (frozen)

## Upstream

**Repository:** https://github.com/Wei-Shaw/sub2api  
**Tag:** `v0.2.5`  
**Commit:** `86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea`  
**Date:** 2026-09-15 16:20:31 +0800  
**Prior tag:** `v0.2.4` (`5de5e2bed035d43591a2e10e51f420ef6a84eb98`, 2026-09-09)

**Commit range analyzed:** `v0.2.4..v0.2.5` = 115 non-merge commits across 79 real PRs.

## Fork (this repository)

**Baseline commit:** `fe6bc318800ca86d2b658f0e8f46e83d32626e8b`  
**Tag:** `v1.1.12`  
**Date:** 2026-09-13 (release commit)  
**Max migration:** `226`

## 第一档 PRs (15 items, upstream commit order)

| PR | Upstream merge commit | Title |
|---|---|---|
| #6904 | `685e97a3a` | scheduler cache omits account_scheduling_threshold/session_window |
| #6912 | `67d3a896b` | redeem duration max 365→36500 |
| #6924 | `2493b0dc3` | validate Codex User-Agent before identity pairing |
| #6843 | `8efe2fd8c` | usage cost tooltip toFixed(6)→toFixed(8) |
| #6969 | `75b7dd1e0` | group peak-rate validation returns 500 not 400 |
| #6973 | `9c30951e4` | renewal modal overflow |
| #7012 | `e67ffda7a` | easypay upstreamType regex allow dots |
| #7022 | `5948988aa` | header_util duplicate accept-encoding |
| #7024 | `d2067668d` | subscription assignment stale selectedUser race |
| #7050 | `67845665d` | balance notify email splice race |
| #7052 | `ba57ea914` | transient auth errors force logout |
| #7053 | `99b93b298` | proxies filter pagination |
| #7073 | `a4c517917` | batch_image priority sort reversed |
| #7082 | `1e1d15cca` | antigravity token cache project_id collision |
| #7162 | `b8275209e` | antigravity gemini SSE blank separator |

## Four-state summary (per-file across all 15 PRs)

- **CLEAN:** 47 files (forward-apply to fork HEAD succeeds in temp index)
- **CONFLICT:** 3 files (context drift: `SubscriptionsView.vue`, `admin_auth_test.go`, `jwt_auth_test.go`)
- **NOBASE:** 0
- **ALREADY:** 0

**Note:** The two `*_test.go` conflicts are new test files referencing upstream test doubles this fork doesn't carry. The `SubscriptionsView.vue` conflict (#7024) is minor debounce context drift — manual resolution trivial.

## Evidence artifacts

Frozen in `docs/upstream-sync/evidence-0.2.5/`:
- `candidates.tsv` — 79 PRs × (tier, batch, four_state, rationale)
- `commits.tsv` — 115 commits in `v0.2.4..v0.2.5`
- `files.tsv` — 542 per-file four-state records across all 79 PRs

## Verification snapshot

Fork compiles clean at baseline:
```
$ cd backend && go build ./...
(exit 0)
$ cd frontend && pnpm run typecheck
(no errors)
```

Integration test build tags confirmed: `backend/internal/**/*_integration_test.go` uses `//go:build integration` (miniredis-compatible) or `integration && postgres` (skip in fork). None of the 15 PRs' unit tests are postgres-gated.
