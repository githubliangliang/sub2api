# AGENTS.md

This file provides guidance to coding agents working in this repository.

**User-selected v1.1.21 release (2026-10-30):** Current VERSION is 1.1.21, Go is 1.27.3 and the maximum SQLite migration is 227. [Detailed release notes](./docs/releases/v1.1.21.md) cover the actual increment after v1.1.20: historical 0.2.0/0.2.5 tail completion, SQLite personal maintenance, v0.2.19 first/second tiers, the selected GPT-6.1 Terra capability, and frontend security maintenance. Product content is based on `1b1c06ac6931947faa7cf3d2b0a215d5aac05c5f`; [release readiness](./docs/releases/evidence-v1.1.21/readiness.md) distinguishes existing implementation/CI acceptance from this release's own checks. The existing v1.1.20 tag and frozen historical source/verification records retain their original values. Publication results are verified against remote refs and the Release workflow; no business-instance deployment is included.

**User-selected GPT-6.1 Terra implemented (2026-10-30):** [port-upstream-0.2.19-terra](./openspec/changes/port-upstream-0.2.19-terra/verification.md) is 14/14 complete, committed on local main as `a541d0b611f4c2da5cab9811c3c4e02ae8344136`. Terra has a canonical identity, six efforts with default medium, preserved none/max and Responses tools; only-Chat tool requests fail before forwarding. Account metadata/raw catalogs retain precedence and cache/ETag integrity. Frozen resource and both fallback prices, explicit zero/overrides, strict 272k pricing boundaries, HTTP/SSE/WS usage and SQLite dedup are verified; existing whitelist/mapping presets/OpenCode configuration include Terra. Backend 54-package unit, embed build, five-package targeted race, SQLite227/quota/fixed-window/dedup and lint 0 issues passed; frontend 2106 tests passed with 2 existing skips, plus typecheck/lint/build/critical and real-browser save/reopen/copy acceptance. The new SQLite fixture migrates once and copies isolated databases for race checks; production dedup, quota, clipboard and form code remain unchanged. VERSION 1.1.20, Go 1.27.3, migration 227, dependencies and Ent/Wire are preserved. [Implementation evidence](./docs/upstream-sync/implementation-0.2.19-terra/README.md) records the actual starting SHA, code hashes and process failures; frozen intake/source/proposal/spec files remain intact. This user-selected third-tier change remains active awaiting archival; other new capabilities remain on demand. No payment, release, push, deployment or business-instance operation was performed.

**v0.2.19 second tier implemented (2026-10-29):** [port-upstream-0.2.19-tier2](./openspec/changes/port-upstream-0.2.19-tier2/verification.md) is 23/23 complete, committed on local main as `cccf931efe201671f8cd27284b04e74114f4f894`. S01–S06 cover Grok response-model audits and ordered catalog observations, SQLite single-key semantic CAS/managed ordinary edits/transaction cache boundaries, intact bounded image uploads, channel search dimension preservation, declared Gemini catalogs and existing test defaults, and x/net 0.62.0 with eight measured MVS changes. The user clarified S01: keep migration 227, retain conflict through the usage-write boundary and verify SQLite billing fallback; no conflict column was added. Backend 54-package unit/build, seven-package targeted/network race, SQLite migration/quota/fixed-window/dedup and lint 0 issues passed; frontend 2105 tests passed with 2 existing skips, plus typecheck/lint/build/critical and real-browser save/catalog acceptance. Existing test fixtures were made deterministic for global-map resets, quota/background catalog request counts and independently generated Codex milliseconds; related production gate/fingerprint/captcha code is unchanged. VERSION 1.1.20, Go 1.27.3, migration 227, frontend dependencies and Ent/Wire are unchanged; remaining audit findings are in [implementation evidence](./docs/upstream-sync/implementation-0.2.19-tier2/README.md). Frozen intake and source SHAs retain their content; the single user-confirmed S01 scenario has its original snapshot preserved. This change remains active awaiting archival; Terra remains unimplemented. No payment, release, push, deployment or business-instance operation was performed.

**v0.2.19 first tier implemented (2026-10-29):** [port-upstream-0.2.19-tier1](./openspec/changes/port-upstream-0.2.19-tier1/verification.md) is 26/26 complete, committed on local main as `683354b2cdd2461dc581fa14158be09ef51de29b`. F01–F07 cover Claude Code UTF-16 fingerprints, Codex in-flight cache freshness, compatible cache-hit usage, Gemini file-reference images, malformed Antigravity tool terminals, bounded pagination and administrator guard HTTP 400 errors. Backend 54-package unit/build, eight-package targeted race, SQLite migration/quota/dedup and lint 0 issues passed; OpenSpec strict and frozen-scope checks are recorded in [implementation evidence](./docs/upstream-sync/implementation-0.2.19-tier1/README.md). Final unit checks removed proxy variables only in the test subprocess after a controlled probe traced the existing captcha test failure to SDK NO_PROXY host:port matching; captcha code/tests are unchanged. VERSION 1.1.20, Go 1.27.3, migration 227, dependencies, frontend and generated Ent/Wire are unchanged; frozen intake sources and other proposals retain their hashes. The change remains active awaiting archival. Second tier and Terra remain unimplemented; no payment, release, push, deployment or business-instance data operation was performed.

**Latest release intake (2026-10-28):** [PORTING-0.2.19.md](./docs/upstream-sync/PORTING-0.2.19.md) assesses v0.2.18–v0.2.19 against `49939fb2cf96633688548ca39204d96558e59d79` / VERSION 1.1.20 / Go 1.27.3 / SQLite migration 227: 50 non-merge commits, 20 PRs and 3 direct commits; 276 candidate-file states A0/C203/X61/N12, rechecked reverse-before-forward. First tier 7 clusters and second tier 6 clusters have new unimplemented OpenSpec proposals, prioritizing Claude Code, Codex and Grok. The user selected GPT-6.1 Terra for actual use; it remains a third-tier new capability with a [separate proposal](./openspec/changes/port-upstream-0.2.19-terra/README.md), also unimplemented. Other new capabilities remain on demand; payment and absent-platform/PG-only changes remain excluded. Ten-package baseline targeted tests, model/compatibility neighbor tests, 61 frontend tests, SQLite/isolated defect probes and an isolated x/net 0.62 minimum-graph check passed; these are assessment evidence, not implementation acceptance. Source baselines, historical evidence and 3732 product-file hashes remain unchanged. No code port, version/dependency/schema change, release or deployment was performed; full scope and older leftovers are in the report.

**Selected personal maintenance implemented (2026-10-28):** [port-upstream-0.2.18-personal-maintenance](./openspec/changes/port-upstream-0.2.18-personal-maintenance/verification.md) completes the user-selected SQLite background jobs, operation-log clear-all fix and lightweight scheduling optimizations. UsageCleanupService, AccountExpiryService and ScheduledTestRunnerService start on SQLite under their existing controls; legacy TEXT scans, mixed-format due comparisons, cancellation and shutdown/overlap are verified. Audit cleanup retains fresh TOTP and a synchronous trace; Enter cannot submit twice. PR #8249 uses a bounded request-local lazy probe; #8250 removes signature allocations, measuring 56%–80% lower function time. Backend 54-package unit/build, four-package targeted race, SQLite migration/quota/dedup and lint 0 issues passed; frontend 2099 tests passed with 2 existing skips, plus typecheck/lint/build and critical tests. Implementation is committed on local main as `b554bea554202e3008b99e3faa7a29420467f318`; [code hashes and scope](./docs/upstream-sync/implementation-0.2.18-personal-maintenance/scope-facts.json) retain the tested pre-commit snapshot against `5c01d92c7`. VERSION 1.1.20, Go 1.27.3, migration 227, dependencies and generated Ent/Wire are unchanged. Historical evidence is untouched; no business-instance clearing, release or deployment was performed. The change remains active awaiting archival; other third/fourth-tier and payment work is outside the selection.

**0.2.5 second-tier remaining tasks completed (2026-10-28):** [port-upstream-0.2.5-tier2](./openspec/changes/port-upstream-0.2.5-tier2/verification.md) is 33/33 complete, committed on local main as `b349c7408325a0e4eba4f32268ad7a3820e2ad4f`. The user requested completion of the remaining tasks, including S12b quota cleanup and SQLite migration 227. Quotas now store only configured limits (zero counts), clearing soft-deletes rows, usage/snapshot updates do not recreate missing/deleted rows, and old TEXT window columns are read compatibly. SQLite snapshots use bind-safe atomic batches; default WS factors remain 1.0 after the independent decision, with the numeric 5.0 patch explicitly not adopted. Backend 54-package unit/build, migrations, quota race and lint 0 issues passed; frontend 2095 tests passed with 2 existing skips, plus typecheck/lint/build and critical tests. Ent/Wire, dependencies and VERSION 1.1.20 are unchanged; current maximum migration is 227. This and the completed 0.2.0 tail remain active awaiting archival; there are no remaining unfinished historical active changes. Frozen sources/TSV and original verification snapshots retain their values; no live business-data migration or deployment was performed.

**0.2.0 second-tier tail completed (2026-10-27):** [port-upstream-0.2.0-p0-tail-and-p1](./openspec/changes/port-upstream-0.2.0-p0-tail-and-p1/verification.md) is now 128/128 complete. The user authorized execution of its three remaining tasks: type-stub removal, four-surface real-browser acceptance, and full post-merge gate/SHA writeback. Implementation is committed on local main as `5b3e1a7168a8bd2711d5992e23256c37f14d8911`; backend 54-package unit/build, migration, HTTP/2 race and Go lint 0 issues passed; frontend 2090 tests passed with 2 existing skips, plus typecheck/lint/build and critical tests. Local/CI golangci-lint is pinned to 2.14.0 (x/tools v0.50.0 supports Go 1.27.3 export V5). The change remains active awaiting archival; the historical unfinished active change is now only 0.2.5 tier2. Frozen sources/audits remain unchanged; VERSION 1.1.20, Go 1.27.3 and migration 226 are preserved.

**OpenSpec archive (2026-10-24):** 19 completed changes are under `openspec/changes/archive/`; the historical unfinished changes `port-upstream-0.2.0-p0-tail-and-p1` and `port-upstream-0.2.5-tier2` remain active. The user explicitly confirmed completion of 0.1.180 P0, 0.2.0 tier 1, 0.2.5 tier 1, and 0.2.9 tier 1; their completion records close historical unchecked tasks and preserve original task/verification snapshots. Main specifications contain 42 capabilities and 251 non-payment requirements. See the [archive index](./openspec/changes/archive/README.md) and [main specifications](./openspec/specs/). Frozen source documents and raw evidence retain their original content and historical paths; archive READMEs provide current evidence links.

**User-selected personal features implemented (2026-10-26):** [port-upstream-0.2.18-personal-features](./openspec/changes/port-upstream-0.2.18-personal-features/README.md) implements usage-log output TPS, account filter persistence, account names in ops request details, rule-level account/model cooldown scope, OpenAI API Key Web Search emulation, and configurable minimum output tokens. The user authorized implementation after the initial specifications and later selected restoration of the full `/admin/ops` page, sidebar and menu entry in standard/simple mode. Backend 54-package unit/build, targeted race and real SQLite/miniredis checks passed; frontend 2089 tests passed with 2 existing skips, plus lint/typecheck/build. Implementation is committed as `0e38d06d82557a0a3b372ca9165a24b9b3acbc93`; see [verification and implementation hashes](./openspec/changes/port-upstream-0.2.18-personal-features/verification.md). The user requested release v1.1.20, prepared as a separate VERSION commit with [detailed release notes](./docs/releases/v1.1.20.md) and [readiness evidence](./docs/releases/evidence-v1.1.20/readiness.md). The change remains unarchived; no business-instance deployment was performed. Sources from v0.2.15–v0.2.18 and fork `c36a19b7a` remain frozen; [supplemental assessment and patch sites](./docs/upstream-sync/PORTING-0.2.18-personal-features.md). Implementation verification used VERSION 1.1.19; the release VERSION is 1.1.20. Go 1.27.3, migration 226 and dependencies are unchanged; frozen implementation evidence retains its original values. `gateway.min_output_tokens` defaults to 128 and requires at least 16, applying only to existing compatibility conversions. These six selected third-tier capabilities do not change historical intake or authorize unrelated capabilities/payment work. TPS covers usage logs, excluding ops percentile statistics; other page restrictions remain in place.

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

### Personal-use scope: no payment features

This is a personal-use project. **Do not introduce, enable, extend, or port payment features**, including payment providers (such as EasyPay), checkout, recharge, paid subscription purchase/renewal, refunds, and related commercial UI, configuration, dependencies, or migrations. This applies to upstream bugfixes as well as new features; a small or cleanly applicable patch does not make payment functionality in scope.

- During upstream intake, classify payment-only changes as **第四档 / out of scope for personal use**, not 第一档/第二档 or optional work to carry along.
- For mixed PRs, take only independently useful non-payment changes. Do not import payment code as a prerequisite; defer the dependent item if it cannot be separated.
- Apply this boundary to new proposals and existing unimplemented OpenSpec tasks. An older task listing is not authorization to implement payment functionality.
- Preserve gateway usage recording, token/cost statistics, account quotas, and upstream account subscription metadata needed for personal operation. These are not payment features. This rule does not authorize deleting existing payment code or data as an unrelated cleanup.

**Previous intake and implementation:** [PORTING-0.2.18.md](./docs/upstream-sync/PORTING-0.2.18.md) evaluates v0.2.17–v0.2.18 on 2026-10-23 against `873962ce7` / VERSION 1.1.18: 56 non-merge commits, 35 PRs and 8 direct commits; frozen file states A4/C170/X117/N21. First tier 6 clusters and second tier 10 clusters cover selected behavior from 18 PRs plus one direct fixture commit. Both batches were implemented from HEAD `498432759` and committed together as `2b79ac320`; see [tier-1 verification](openspec/changes/archive/2026-10-24-port-upstream-0.2.18-tier1/verification.md) and [tier-2 verification](openspec/changes/archive/2026-10-24-port-upstream-0.2.18-tier2/verification.md) for acceptance status and evidence. Backend 54-package unit/build, targeted race and real SQLite settlement/dedup/fixed-window checks passed; frontend 2027 tests passed with 2 existing skips, plus lint/typecheck/build. Tier-2 was implemented with three agents on 2026-10-24, covering credential routing, Responses timestamps, Codex retirement/capacity/identity, Anthropic and Gemini interrupted usage, existing Antigravity/Vertex behavior, image download boundaries and targeted dependency maintenance. Go is now 1.27.3; Ent's export-format incompatibility required the measured minimum x/tools v0.50.0 prerequisite and its forced transitive versions. Frontend follow-redirects is 1.16.1; remaining advisories are in the tier-2 dependency report. Preserve SQLite/miniredis/simple mode, migration 226 and fixed quota-window reset behavior. Implementation verification used VERSION 1.1.18; VERSION is now 1.1.19 for the user-requested release, with [release notes](./docs/releases/v1.1.19.md) and separate readiness evidence. Frozen source-baseline and intake evidence remain unchanged. New capabilities remain deferred; payment changes remain out of scope. Implementation is committed; v1.1.19 is prepared as a separate version commit and annotated tag. No business-instance deployment has been performed. Older leftovers remain in §5.

**Previous intake and implementation:** [PORTING-0.2.16.md](./docs/upstream-sync/PORTING-0.2.16.md) evaluates v0.2.13–v0.2.16 against `128c5f7bf` / VERSION 1.1.16: 149 non-merge commits, 89 PRs and 4 direct VERSION commits; frozen file states A0/C367/X168/N40. On 2026-10-14, the first tier's 14 clusters / selected behavior from 40 PRs were implemented at `7c544c03f` after `0fbcab935`; see [verification](openspec/changes/archive/2026-10-24-port-upstream-0.2.16-tier1/verification.md). The second tier's 16 clusters were implemented and verified on 2026-10-18, committed as `847db09e1` (S01–S15) and `5ff6e5bf9` (S16), and fast-forwarded into main; see [tier-2 verification](openspec/changes/archive/2026-10-24-port-upstream-0.2.16-tier2/verification.md). Includes deleted-Key SQLite settlement, Claude/Codex/Grok protocol fixes, existing Gemini/image cooldown fixes, frontend lifecycle/input boundaries, account-test diagnostics and existing API aliases. F10 additionally repairs the existing error-message helper after tests exposed a missing prerequisite. No new Haiku/platform/payment capabilities. Preserve the local fixed quota-window reset fix at `128c5f7bf`. Tier-1 implementation verification used VERSION 1.1.16; migration 226 and dependencies remained unchanged at that stage. VERSION is now 1.1.18 for release preparation; source-baseline and historical intake/verification evidence remain frozen. No deployment was performed. Tier-2 implementation verification used VERSION 1.1.17 and migration 226, with targeted DOMPurify/axios/Vue/source-map-js maintenance. Full dependency audit decreased from 103 to 70 records; production high findings remain the two xlsx advisories. Expired exceptions were removed, not renewed; remaining maintenance is recorded in the tier-2 dependency report and §5.

**Previous intake and implementation:** [PORTING-0.2.12.md](./docs/upstream-sync/PORTING-0.2.12.md) (2026-10-02): first tier 3 clusters and second tier 3 clusters are implemented and verified at `405246789`, based on `5eb58199e`, on `sync/upstream-0.2.12`. Includes Grok CLI 1.0.46 identity, Antigravity client-error sanitization, corrected error-code guidance, atomic email attempt/reset-token handling, and the user-requested priority quick adjustment / API key group sorting. Fork adaptations preserve legacy JSON attempt counts and isolate priority updates across reused rows. Full backend unit/build, targeted race, SQLite dialect/usage dedup/group sorting, frontend typecheck/lint/build passed; frontend 1854 passed / 2 existing skips. Both OpenSpec verification reports are complete; source-baseline and intake evidence remain frozen. Payment promotion/public order verification, TypeSafe and unrelated axios upgrade remain excluded. Implementation verification used VERSION 1.1.15; migration 226 and dependencies remain unchanged. VERSION is now 1.1.16 for release. The local Docker container `sub2api-new-sqlite` was subsequently rebuilt and verified healthy at the user’s request using version 1.1.15 / `975e3c1b3`, with data/config preserved and a rollback backup; historical intake and implementation reports retain their original version and deployment status. Older leftovers remain in §5.

**Previous intake and latest implementation:** [PORTING-0.2.11.md](./docs/upstream-sync/PORTING-0.2.11.md) (2026-09-30): first tier 4 clusters and second tier 6 clusters are implemented and verified at `96f32d018`, based on `f13dee245`. Backend full unit/build, targeted race, SQLite dialect/usage dedup and frontend typecheck/lint/build passed; frontend 1836 passed / 2 existing skips. See both OpenSpec verification reports. The v0.2.9 F04 thinking-disabled prerequisite and only 14 generic usage type lines from `695ebede7` are included. Uses the final #7736 remote/file catalog design. VERSION is now 1.1.15 for release preparation; migration 226, dependencies and SQLite/miniredis/simple mode remain unchanged. Local Docker `sub2api-new-sqlite` and the remote user systemd service were updated and verified healthy at the user’s request using the implementation worktree at version 1.1.14, before release preparation. Historical intake and deployment evidence retain that version. Astra Ultrafast, native reset management, new limits and all other third/fourth-tier items remain excluded.

**Previous intake status:** [PORTING-0.2.9.md](./docs/upstream-sync/PORTING-0.2.9.md) was assessed on 2026-09-28. Its second tier's 48 PRs / 16 clusters **landed at `159d04222`** and are included in v1.1.14; [acceptance evidence](openspec/changes/archive/2026-10-24-port-upstream-0.2.9-tier2/verification.md) records backend unit/build/race, SQLite, frontend 1808 passed / 2 existing skips and browser checks. The user confirmed the first tier (39 PRs / 15 clusters) complete on 2026-10-24, and it is now archived; see its [completion record](./openspec/changes/archive/2026-10-24-port-upstream-0.2.9-tier1/completion-record.md). Earlier pending labels and blank verification slots are historical snapshots, superseded for completion status by that explicit user confirmation; this archival update does not infer a new implementation SHA or merge path. The v0.2.5 main batches already landed at bfbcd79. Frozen intake and verification notes retain their historical baseline; current leftovers are restated in the latest assessment §5.

**Merging upstream:** git history was rewritten (no shared ancestor). Do **not** `git merge upstream/main`. Cherry-pick or port by feature. Process + current checklists: [latest v0.2.19 assessment](./docs/upstream-sync/PORTING-0.2.19.md), [selected personal maintenance supplement](./docs/upstream-sync/PORTING-0.2.18-personal-maintenance.md), [previous v0.2.18 assessment](./docs/upstream-sync/PORTING-0.2.18.md), [docs/upstream-sync/README.md](./docs/upstream-sync/README.md), [docs/upstream-sync/PORTING-0.2.4.md](./docs/upstream-sync/PORTING-0.2.4.md) (previous intake, 2026-09-12: v0.2.1–v0.2.4 including the v0.2.2 tag; 13 第一档 items + 12 第二档 clusters **landed on main and verified** at `a26b6b55d`, including the API Key Responses namespace follow-up fix; release target `v1.1.12`; OpenSpec changes, frozen intake evidence and final implementation report linked there; earlier leftovers rechecked), [docs/upstream-sync/PORTING-0.2.0.md](./docs/upstream-sync/PORTING-0.2.0.md) (0.1.185 + 0.2.0 first/second-tier main batches merged into main at `c9d9bebe8`; DOMPurify maintenance completed in v0.2.16 tier-2), [docs/upstream-sync/PORTING-0.1.184.md](./docs/upstream-sync/PORTING-0.1.184.md) (0.1.184, P0/P1 merged), [docs/upstream-sync/PORTING-0.1.183.md](./docs/upstream-sync/PORTING-0.1.183.md) (0.1.181–0.1.183 bugfixes), [docs/upstream-sync/PORTING-0.1.180.md](./docs/upstream-sync/PORTING-0.1.180.md).

**Go version:** `1.27.3` (from `backend/go.mod`). CI asserts this string; bump go.mod and workflow version checks together.

**Frontend package manager:** **pnpm only** (not npm). Commit `frontend/pnpm-lock.yaml` after dependency changes. pnpm v11 needs `frontend/pnpm-workspace.yaml` `allowBuilds` for `esbuild` / `vue-demi` postinstall.

## Upstream release intake (run this whenever a new upstream release lands)

Triggered by "check upstream", "evaluate the new release", or a new tag showing on <https://github.com/Wei-Shaw/sub2api/releases>. Produce artifacts, not a chat-only answer.

**Default assessment priority: Claude Code, Codex, and Grok bug fixes.**

- Evaluate fixes to these three existing client/provider paths first, including authentication, request headers, protocol conversion, tool calls, streaming/SSE/WebSocket behavior, model mappings, retry/failover, account quotas, and usage/cost recording. Shared gateway fixes count when their effect on these paths is verified.
- Within each tier, prioritize these fixes in the report and proposed landing order over unrelated UI conveniences, general operations enhancements, and new platforms. Keep shared security, data-integrity, and SQLite correctness issues visible according to their actual severity.
- Recheck older pending fixes affecting these paths and include only the necessary, measured prerequisites with their consumers. State which path benefits and how it will be verified.
- Keep the full release inventory and four-tier rules below. This priority does not promote 第三档/第四档 automatically or authorize implementation. New capabilities remain on demand even when related to Claude Code, Codex, or Grok; payment features remain out of scope.

**1. Fix the baseline.** `git fetch upstream --tags --prune`; list `<last-evaluated-tag>..<newest-tag>` non-merge commits. The last round's *upstream main* may already sit inside the next tag, so part of the range can already be ported — run a **reverse** `git apply --check` over the whole candidate set first (`ALREADY` state) before the forward one.

**2. Judge per candidate, four states per file**: `CLEAN` (forward applies) / `ALREADY` (reverse applies = already here) / `CONFLICT` / `NOBASE`. For a feature upstream split into many micro-commits, take the whole PR diff (`git diff <merge>^1 <merge>`), never the individual commits. Then, for every new call the patch introduces, `grep -rn "func .*<symbol>"` — three states never prove it compiles. `CONFLICT` may mean the changed function does not exist here, i.e. **this fork does not have the bug**; verify before filing it as work.

**3. Rank into four tiers** (this is the priority model — keep the tier numbers, downstream artifacts reference them):

Apply the personal-use scope above before ranking: payment-only changes are out of scope even when the defect exists locally and the patch is cheap.

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
golangci-lint run ./...   # CI uses v2.14.0 (x/tools v0.50.0 supports Go 1.27.3 export V5)
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
7. **Disabling a background service is not a SQLite adaptation.** The former `skipSQLiteBackgroundJobs` (`internal/service/wire.go`) skipped whole Start calls and silently disabled features (the earlier `c35a482` scheduler outage is historical evidence). On 2026-10-28, UsageCleanupService, AccountExpiryService and ScheduledTestRunnerService were audited per-SQL, adapted and restored; see [current verification](./openspec/changes/port-upstream-0.2.18-personal-maintenance/verification.md). Keep configuration, cancellation, result-storage and scheduling propagation boundaries. Legacy TEXT scans and mixed-format chronological comparisons need separate verification. The 2026-08-16 SQL audit and `$1` placeholder conclusion remain historical; PG→SQLite reference: [docs/upstream-sync/README.md §5](./docs/upstream-sync/README.md).

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
