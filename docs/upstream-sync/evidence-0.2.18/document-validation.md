# 文档验证（2026-10-23）

- 新增Markdown相对链接全部可解析。
- 43候选与decisions逐一对应；35PR，8直接提交；候选非merge计数总和56，初始采集脚本另验证集合严格相等。
- 312文件四态A4/C170/X117/N21与汇总相符；旧#5925补充证据独立，不混入统计。
- 两份OpenSpec各8个Markdown文件；16条Requirement、32个Scenario；全部实施checkbox未勾选，verification证据槽为空。
- 两份OpenSpec官方CLI --strict校验通过。
- 现有协议/定向/相邻/认证/SQLite/路由测试输出通过；overlay探针的首次错误假设单独保留，最终探针复现旧缺陷通过。
- git diff --check通过；tracked变动只有AGENTS.md与上游同步README，新增文件仅本轮报告、证据和两份OpenSpec；HEAD保持873962ce7。
- VERSION1.1.18、迁移226、源码和依赖锁文件未改。无commit/tag/release/deploy。

测试通过范围以baseline-tests.md为准；不表示拟实施代码已通过全量验证。

提交前补检：将新增TSV统一为LF并使用CSV全字段引用，逐行解析结果保持一致；归一化两份日志的空白和design尾部空行。以暂存区git diff --cached --check覆盖全部新增文件。
