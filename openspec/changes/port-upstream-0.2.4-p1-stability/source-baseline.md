# 来源基线（实施期间不可改写）

评估日期：2026-09-12。第二档：单机稳定性与兼容性。

| 基线 | 完整 SHA |
|---|---|
| 本 fork HEAD / v1.1.11 | c9d9bebe87926791bfdd716f59abcba413fee20a |
| 上次上游终点 v0.2.0 | aa236488351eb71e120fc2b6fb32e36b0374c918 |
| v0.2.1 | 578785ee7fb35030b094b69624efe25670a36f5f |
| v0.2.2（tag，无对应 Release 列表项） | 5485f368b29d05adb95a00f71801c7c23d8f48af |
| v0.2.3 | 8fa67d477d6651a744754392a8982ea589c26ae6 |
| v0.2.4（本轮代码终态） | 5de5e2bed035d43591a2e10e51f420ef6a84eb98 |
| 拉取时 upstream/main，仅供观察 | 4726bdd08b6201d426a80529b79be123a4008d20 |

本仓库 VERSION=1.1.11，最大 SQLite 迁移为 226；本 change 不需要新迁移。
本仓库与上游无共同祖先。来源 PR 使用第一父到 merge 的整体 diff，不直接 merge 上游。
详细来源见 [source-feature-map.md](./source-feature-map.md)，逐文件四态见
[files.tsv](../../../docs/upstream-sync/evidence-0.2.4/files.tsv)。

四态固定在 HEAD 的独立临时 index。评估阶段测试在当前工作树执行，包含用户已有的
openai_gateway_forward.go、openai_ws_http_bridge.go 与两个 openai_responses_input_metadata*.go 改动。
这些文件不属于本 change 已完成内容；实施时需要保留并重新跑 metadata 回归。

实施应基于第一档落地后另记的实际起点；不要为更新实际起点而改写本文件。S06 来源 PR 的兑换测试、其它功能断言不是 Astra 的交付范围。

实际实施起点、完成 SHA、测试和取舍填 verification.md；这里保持评估时的来源不变。

