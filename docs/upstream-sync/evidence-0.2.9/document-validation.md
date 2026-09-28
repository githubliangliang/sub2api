# 文档验收

2026-09-28，产品基线 bfbcd79，评估文档静态检查通过。

- 215条非merge提交集合与Git一致；170候选唯一，覆盖161PR和9直接提交。
- 867条候选/文件对无重复，逐候选文件数、四态总数匹配；分档覆盖每个候选一次。
- 第一档39PR/15簇、第二档48PR/16簇；来源映射、spec、未勾选tasks对应一致。
- 两份change各8份必需文档，37个相对链接可解析，无行尾空白。
- 后端产品、前端产品、部署和工作流无Git差异；旧冻结evidence/source文档未改。
- 来源测试清单：{'EXISTS': 61, 'MISSING': 53}；只作源码存在性清单，缺失项需新增/适配，不能记作测试通过。
- 关键符号核对见base-symbols.tsv；不是Go类型检查，不保证所有新增调用可编译。
- 独立读者复核已修正：具体WHEN/THEN、#7367非负整数、#7446单独减号占位、#7622保留显式/v1、#7571不写已有能力、错误的hunk邻近函数定位。

OpenSpec官方 strict 校验（CLI 1.4.0）已执行：

```text
openspec validate port-upstream-0.2.9-tier1 --strict --no-interactive
Change 'port-upstream-0.2.9-tier1' is valid

openspec validate port-upstream-0.2.9-tier2 --strict --no-interactive
Change 'port-upstream-0.2.9-tier2' is valid
```

两次命令退出码均为0。独立读者随后复核修订版，确认原先指出的问题已解决，限定复核未发现重大事实错误。
此为文档质量检查；产品代码未实施，产品测试结果及缺口见baseline-tests.md。
