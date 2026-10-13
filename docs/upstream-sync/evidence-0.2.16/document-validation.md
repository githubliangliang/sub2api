# 文档一致性校验

2026-10-13，基线128c5f7bf。结果：通过。

- 149条非merge、93候选（89PR/4直接）、575逐文件记录与统计一致；全部候选有分档。
- 第一档14簇、第二档16簇，69个本轮PR选定行为；混合来源按子项重复记录，不重复宣称独立PR。
- 两份OpenSpec各8文件；30条Requirement、60个Scenario；所有实施checkbox未勾选，verification实施证据为空。
- 两批`pnpm dlx @fission-ai/openspec validate <change> --strict`通过，原输出见openspec-tier1.txt和openspec-tier2.txt。
- 新文档相对链接有效，来源SHA与冻结baseline一致；git diff --check及新Markdown尾空格检查通过。
- git HEAD未变；状态仅AGENTS/同步索引、新评估证据及两份OpenSpec。产品代码、依赖、迁移、VERSION和旧评估证据未改，无发布部署。

文档校验通过不代表实现完成：embed原基线两处fixture失败、audit103条告警仍保持原状；详情见baseline-tests.md。

提交前格式检查补充：TSV统一为LF换行并引用字段，解析后的12082行（含表头）保持一致；三份控制台日志仅规范空白。已补查暂存区全部58个文件的git diff --cached --check。
