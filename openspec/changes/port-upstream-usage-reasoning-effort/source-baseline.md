# 冻结来源

| 项目 | 2026-11-19 提案基线 |
|---|---|
| fork | v1.1.26，提交 `9c073cb30bff9dfd9364a059d4e0929b14d8b8ff`，tree `77b99e4296bf3309970eb2230cf6affe162c9619` |
| 本地数据与工具链 | SQLite 最大迁移 229，auth snapshot 21，Go 1.27.4；usage_logs 已有 `reasoning_effort` |
| 历史来源 | [PR #6188](https://github.com/Wei-Shaw/sub2api/pull/6188)，最早包含它的正式标签 v0.1.184，指向 `e98ef32eb29aecd30d1def615912ec4dc93173f3` |
| 完整 PR | 第一父提交 `d4754c211a1b5ad71b80734e7d4daec827c54f4f` → 合并提交 `7b693ae4295e20329f18ff451b29a38879cb4705`，41 个文件，+1020/-106 |
| 对照内容 | `git diff --no-color --no-ext-diff --binary d4754c211a1b5ad71b80734e7d4daec827c54f4f 7b693ae4295e20329f18ff451b29a38879cb4705` 的 SHA-256 为 `8a7b48f7341f7f17f212a49aa6f7c9cefcfcbf7a01162174790f2f244497f90d` |

后续 [v0.2.23 遗留评估](../../../docs/upstream-sync/PORTING-0.2.23.md) 仍将 requested effort 持久化列为按需。选取上述完整 PR 的行为，按当前 fork 的 SQLite、HTTP/WS 和分组规则适配。实施证据另记入 verification。
