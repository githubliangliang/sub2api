# 规格文档校验

2026-10-24；仅文档校验，产品未实施。

## OpenSpec

实际工具版本：1.14.3。执行：

```bash
rtk pnpm dlx @fission-ai/openspec validate port-upstream-0.2.18-personal-features --strict --json --no-interactive
rtk pnpm dlx @fission-ai/openspec status --change port-upstream-0.2.18-personal-features --json
```

两命令退出0。严格校验结果：1 item passed，0 failed，issues为空。status确认proposal/specs/design/tasks四类规划文档齐全；其isPlanningComplete/isComplete表示规划文档完成，不表示实施任务完成。

## 文档与范围

- 六个新增capability共21个ADDED Requirements，加一个既有需求MODIFIED delta；合计22个需求、49个Scenario。
- 每个需求具有SHALL/MUST，每个Scenario有WHEN/THEN；结构复核无错误。
- tasks.md的35个任务和verification.md的4个完成条件均未勾选，所有实施证据槽空白。
- README/proposal/source-baseline/source-feature-map/design/tasks/verification与schema元文件齐全；全部已写入文档的相对文件链接可解析。
- `git diff --check`退出0。此变更只新增spec/评估/证据并小范围更新AGENTS与上游索引。
- 检查时HEAD已由用户另一项归档工作推进至`67b1a07270acd79f5d54432b44ae396c3f7e360c`。相对冻结`c36a19b7a1a132e690eaf9ce0d878b301ffffcb4`，backend/frontend/deploy/.github无commit差异，也无产品工作树差异；冻结source-baseline不改写。

相关现有测试的四包通过、前端122项通过见baseline-checks.md，不计入新增功能实施验收。不执行发布、tag、部署或真实搜索调用。

来源文件及机器检查证据的冻结摘要见frozen-files.sha256；其路径均相对仓库根目录。
