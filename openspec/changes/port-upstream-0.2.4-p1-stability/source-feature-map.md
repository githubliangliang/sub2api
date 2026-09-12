# 来源与行为映射

每行 SHA 是 PR 的 merge commit；统一使用 `git diff <sha>^1 <sha>` 查看完整功能来源。
同一 ID 的多个 PR 共同构成该行为簇，不能把中间态当终态。
行为契约在 specs，patch sites 只在 [PORTING-0.2.4.md](../../../docs/upstream-sync/PORTING-0.2.4.md)。

| ID | 首次所在 tag | 完整来源 SHA | 来源 PR |
|---|---|---|---|
| S07 | v0.2.1 | `78e1aaedb983a3e30d250e15b1093be3d3ca8018` | PR #6535 from feeeei/feat/pricing-custom-file-hot-reload |
| S01 | v0.2.1 | `91cf660377af627c1e48a2f20776982ae6dc3df1` | PR #6397 from feeeei/fix/openai-ws-replay-oom-6382 |
| S10 | v0.2.1 | `15278fe2f9f145eb1e45ad136901c0da83cd0be0` | PR #6580 from OG-Wang/fix/sidebar-collapse-when-active |
| S05 | v0.2.1 | `d7fb9b5d35bf3f19cf68df9c329974ce99f7932a` | PR #6593 from wucm667/fix/issue-6584-chat-tool-discovery |
| S06 | v0.2.1 | `85f431a072b391e704df6cc3a12f7db46b62d1ef` | PR #6620 from wucm667/fix/issue-6615-astra-messages-cache |
| S07 | v0.2.1 | `63dc24b5e8a736279d60c8270b763c3e733c9688` | PR #6626 from wucm667/feat/issue-6624-none-reasoning-mapping |
| S05 | v0.2.1 | `68f099707e2eb9d7c4e9e7e8176c15a588209345` | PR #6553 from wucm667/fix/issue-6547-heartbeat-bootstrap |
| S05 | v0.2.1 | `620eb3fd006918772882dbe57e62702a2cdfd663` | PR #6581 from r266-tech/r266/forward-opencode-session |
| S04 | v0.2.1 | `9517f12cdd34b42c499e96d0a97aef668d14edcc` | PR #6510 from wall-wxk/fix/release-session-slot-on-forward-failure |
| S11 | v0.2.1 | `9b4fcfb897455eebc17812a0af2d365c0c86fc3f` | PR #6606 from luckydududu/feat/claude-cli-version-env-override |
| S06 | v0.2.1 | `0d9e7e1530855c93d4d00fc87406960d4a7fefd5` | PR #6628 from baryon/contrib/openai-gpt6-astra-capability-sync |
| S05 | v0.2.1 | `cc52c93ad7d60c5364c66bbf8c2efacb253a29d7` | PR #6539 from creamtea47/codex/fix-delegation-bootstrap-resume |
| S10 | v0.2.1 | `aafe93b335763e0586f4f30665751c279dc78c73` | PR #6557 from Ritel-T/pr/compact-admin-account-list |
| S06 | v0.2.1 | `ed7c8f2208f478907782d2ad646463941d455379` | PR #6572 from lyy0709/codex/gpt-6-astra |
| S05 | v0.2.2 | `8116d235ef62dc1d88e7a2a9603058e87e368cea` | PR #6629 from superdav42/fix/preserve-function-arguments-done |
| S11 | v0.2.2 | `787a6a33df3c7a7d17563d1e9d61e3d4800e38e7` | PR #6481 from alfadb/fix/claude-cli-version-floor |
| S10 | v0.2.2 | `c05bc4d3ccb4d52b88e7261df58d03ed5f59aa15` | PR #6705 from ouchihao/codex/issue-4827-usage-api-key-filter |
| S06 | v0.2.2 | `561fc1c3e455e405af4d604339e8e9df4f044445` | PR #6718 from lyy0709/codex/astra-ultra-catalog |
| S06 | v0.2.2 | `cbb4b7e53e4bd7e94f09a72325f27250b0360e5c` | PR #6678 from gebdalaoli-arch/codex/stale-astra-input-modalities-v021 |
| S06 | v0.2.2 | `8193b80a59ada85d8cdff4f1e47a9be48fca6292` | PR #6743 from MarushiruDonato/fix/gpt6-astra-instructions |
| S06 | v0.2.2 | `5485f368b29d05adb95a00f71801c7c23d8f48af` | PR #6690 from alfadb/fix/astra-pro-mode |
| S08 | v0.2.4 | `28b807a7353154d09f48b084e757ed866824d6f4` | PR #6816 from Pluviobyte/codex/fix-repeated-proxy-fallback |
| S08 | v0.2.4 | `4892b8f17edcda6a4355ded98c190ac5775982b9` | PR #6810 from Pluviobyte/codex/fix-announcement-proxy-date-range |
| S10 | v0.2.4 | `3ee92b40c5fc34e1254f304219f1da4881b8d00e` | PR #6659 from ouchihao/codex/issue-77-remove-inactive-groups |
| S10 | v0.2.4 | `6378fb0cbaff51c968470ef2fe5271e79d59b9fc` | PR #6798 from savvym/codex/fix-account-menu-overflow |
| S10 | v0.2.4 | `979fe247f9026db677598de170f93d6e1930215e` | PR #6821 from 22zyfeng-collab/fix/antigravity-oauth-plan-type |
| S02 | v0.2.4 | `b6384452347fb6d240fe25dbfca202ca77e6acfb` | PR #6754 from ccemoji/codex/minimal-upstream-fix |
| S12 | v0.2.4 | `1923d1c27d411beb5ad8c1fb875900e06c8ec11c` | PR #6814 from feeeei/fix/go-redis-pool-nil-ctx-panic |
| S08 | v0.2.4 | `9e7039ebca5d4cd668ca1868becbd243d14f3c65` | PR #6811 from Pluviobyte/codex/fix-proxy-partial-update |
| S10 | v0.2.4 | `68773aab9862256a274b5a78e76c73d93847cb60` | PR #6764 from ouchihao/codex/issue-1499-select-failed-token-refresh |
| S10 | v0.2.4 | `ae4cc14b280e91f407b12a141c768d88d6c535ab` | PR #6819 from LeeeeeeM/codex/fix-registration-visibility |
| S08 | v0.2.4 | `5cbb9191a76394d6c3a95ec8a83b2f8e80265802` | PR #6836 from r266-tech/fix-proxy-expiry-json-range |
| S02 | v0.2.4 | `43569bb44c3eea843cec5fcacf0688961cd97f6b` | PR #6434 from swjturay/codex/upstream-client-disconnect-drain |
| S09 | v0.2.4 | `95acbf1f031a291335c60515ee368014139da113` | PR #6424 from cyhhao/fix/bound-ops-system-log-storage |
| S03 | v0.2.4 | `6f0d0ababc64f79f06b6a532b1f0d041b1716d94` | PR #6320 from spongehah/fix/runtime-block-honor-persisted-cooldown |
| S08 | v0.2.4 | `c54897a59dfe3faa819f2a33117aa423353e2a1e` | PR #6815 from Pluviobyte/codex/fix-directed-proxy-backups |

第一档 F01-F13、第二档 S01-S12 的 ID 与 PORTING 一致。
来源 PR 内不相关测试/生成物不自动成为交付要求；所有排除项在实施 verification 中逐一记录。
完整候选档位与理由见 [candidates.tsv](../../../docs/upstream-sync/evidence-0.2.4/candidates.tsv)。

