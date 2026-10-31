# 来源与行为映射

每行 SHA 是 PR 的 merge commit；统一使用 `git diff <sha>^1 <sha>` 查看完整功能来源。
同一簇的多个 PR 共同构成该行为，不能把中间态当终态。
行为契约在 specs，patch sites 只在 [PORTING-0.2.5.md](../../../docs/upstream-sync/PORTING-0.2.5.md)。

| 簇 | 行为 | 来源 PR | 完整来源 SHA |
|---|---|---|---|
| S01 | Responses 正文恢复与 system/developer 角色归一 | #6925 | `3e4b6a09b2d7b6a1d118f76aef5fa2f759dec912` |
| | | #7094 | `d4439129a7011b994e9fd3021bab9651565eb60c` |
| S02 | OpenAI WS 连接池常驻读循环与容量系数 | #6965 | `8f9a9a255b0b392fcb5659aa4906eef4d58067b5` |
| | | #7064 | `bdb42e22f81fcb633ff0a060961211dd2bcb515b` |
| | | #7043 | `2a4f3d1a7108f5acf5cd80ccaa7bd83ed6361e82` |
| S03 | Antigravity 内置工具与客户端工具混用 | #6689 | `150c52b8de331a985038c5972f993f9623105bf9` |
| S04 | Antigravity 部分刷新告警回包 | #7076 | `dd2e6b36801c7bef5607e0580e5b19263fbbf54c` |
| S05 | DeepSeek 模型名校验 | #6977 | `86f93c28ee34cc74b629dafb748bd5ac5ca8c5ea` |
| S06 | Grok 媒体并发槽与视频归属 | #6661 | `14029e50aa04e5275de60726037eb9b29ab874b7` |
| S07 | apicompat chat 桥 agent_message | #6964 | `310f8b7fa27c44b12724a3a0f7d9858456379f6d` |
| S08 | Claude cache_control 保留与中途 output-config beta | #6943 | `4adcef4ff4185e802ec4c1d10a016de807c5dd80` |
| | | #6995 | `eaa4083e23a94f99bb1442e5c70560b957e3d957` |
| S09 | Codex 配额窗口、调度指标与心跳信封 | #7074 | `30ed40a56a5f4b5ab7b8dd3d685353db3a531c84` |
| | | #7072 | `2061508ad858203aae473af90919e174f14bdbf5` |
| | | #6869 | `43f9383d40da666588efae1aceac90b41d89483c` |
| S10 | 大图正文分配与隐私客户端指纹 | #6783 | `88011a6a2e7bf221df7b1950c3c0e16a22028002` |
| | | #7010 | `3a070ec1bf8ce953e228bc54ddc0dfe915f5e193` |
| S11 | 管理端与用户端前端正确性 | #7026 | `0a378f3432c4177b5149dd5cf22d14e66d92e92e` |
| | | #7112 | `e9169901d2ef15d3cf49b6ebc4f75cd7de292137` |
| | | #6654 | `a444551a6d042f9f0ada68544b653a45b9757532` |
| | | #6916 | `0be30886a5a60afe25a0c5e4a17d33326eeb6024` |
| | | #7023 | `f2b51e3b24e075cf81c294cab667441ffadad842` |
| | | #7025 | `329641a86fc6c15a4efa31ba5f90eaf1ed848cfa` |
| | | #7054 | `f0dd497780acff99b715acb2dfbd52474f0f5778` |
| | | #7055 | `4ff3e6dfb92cb3cb4e3ac2528e5f7f22d64d4449` |
| | | #7111 | `f7e959ae5f9fccb97ee12ffebd0dc3036c32c987` |
| S12 | 平台限额清理与 TTFT 运维详情 | #6954 | `2dff7af0f915bbae0cb7871767574a82022e44f3` |
| | | #6971 | `7145484fde802a656e652e6f8e0e14f567437e95` |

共 12 簇 / 28 个候选。簇 ID 与 [PORTING-0.2.5.md](../../../docs/upstream-sync/PORTING-0.2.5.md) 一致。

来源 PR 内不相关测试与生成物不自动成为交付要求，所有排除项在 verification.md 逐一记录。已知需排除：

- **S02 #7064** 同 PR 夹带 `openaiWsMode.ts` 与 i18n 的 WS 模式文案改写（上游 #6772 系列的 UI 收尾），
  与容量系数无关，可单独判断是否要。
- **S11 #6916** 的 `SubscriptionsView.userUsageLink.spec.ts` 与 S11 #7024（第一档）改同一个 spec 文件，
  两批实施顺序要固定，避免互相覆盖。
- **S12 #6954** 的 `BatchSnapshotUsage` 重写含 PG 专用写法，须按 SQLite 重写而非照抄；
  其 SQL 迁移在本仓库应排为 `227`，不继承上游编号。
- **S09 #6869** 的 `automation_bootstrap_test.go` 为 CONFLICT，按本仓库现有 heartbeat 测试形态重写断言。

完整候选档位与理由见 [candidates.tsv](../../../docs/upstream-sync/evidence-0.2.5/candidates.tsv)。
