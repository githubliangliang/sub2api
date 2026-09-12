# 来源与行为映射

每行 SHA 是 PR 的 merge commit；统一使用 `git diff <sha>^1 <sha>` 查看完整功能来源。
同一 ID 的多个 PR 共同构成该行为簇，不能把中间态当终态。
行为契约在 specs，patch sites 只在 [PORTING-0.2.4.md](../../../docs/upstream-sync/PORTING-0.2.4.md)。

| ID | 首次所在 tag | 完整来源 SHA | 来源 PR |
|---|---|---|---|
| F03 | v0.2.1 | `d8b9a83f8cda66a821b3dcadf8ecbdd2c6c65b90` | PR #6492 from feeeei/main |
| F12 | v0.2.1 | `570bd084c23429077dec0bede49157f8e4191ea7` | PR #6594 from wucm667/fix/issue-6592-preserve-zero-system-metrics |
| F04 | v0.2.1 | `8f1d6af3ed1aae36d2b95cfdbb827ba535787562` | PR #6531 from wucm667/fix/issue-6524-continuation-unavailable-recovery |
| F05 | v0.2.1 | `7b271bbee985e8d62498947811e5886fd93cbc92` | PR #6529 from wucm667/fix/issue-6521-preserve-passthrough-reasoning |
| F02 | v0.2.1 | `83094abf21c5f5752c3e770bc8c70704737a35ec` | PR #6536 from wucm667/fix/issue-6526-mapped-model-scheduling |
| F07 | v0.2.1 | `578785ee7fb35030b094b69624efe25670a36f5f` | PR #6640 from cyhhao/fix/claude-billing-fingerprint |
| F06 | v0.2.2 | `e274de45b784fc39bac880a57ec61c4b1e361cd3` | PR #6702 from a1gaoshanlaiyige/fix/message-cache-json-escaping |
| F10 | v0.2.2 | `c0420e2b8a45598c39f957dd7e5ce6c88ac1c814` | PR #6484 from aofee/codex/fix-fable-credits-required-scope |
| F09 | v0.2.2 | `2cc1e7ef900962735992c42e351430abe3a0603e` | PR #6677 from Lynricsy/fix/claude-thinking-block-binding |
| F11 | v0.2.2 | `f7c48ba59a81e8a2dee644866e3ea1cae9f96f2c` | PR #6674 from wuji-labs/fix/antigravity-toolconfig-always-present |
| F08 | v0.2.4 | `4e9b01fd59b6621c4e02e5ab44dc93ce52c8534d` | PR #6838 from atogumo/fix/claude-code-probe-any-model |
| F01 | v0.2.4 | `e2fd418a964206c374586740025bade1d5493a07` | PR #6372 from clansty/fix/openai-oauth-429-fallback-disabled |
| F13 | v0.1.185（旧项） | `e21b849a926f5f30683ac9a421589f1cf37a4c51` | API Key instructions；本轮重判 |

第一档 F01-F13、第二档 S01-S12 的 ID 与 PORTING 一致。
来源 PR 内不相关测试/生成物不自动成为交付要求；所有排除项在实施 verification 中逐一记录。
完整候选档位与理由见 [candidates.tsv](../../../docs/upstream-sync/evidence-0.2.4/candidates.tsv)。

