# port-upstream-0.1.180-p0-fixes

移植上游 v0.1.180 清单里的 **19 项 P0**（21 个 commit）：池模式同账号重试在两条 compat 路径上
完全失效、ops 面板混用 cgroup 与宿主机内存、OAuth 账号被空 `openai_capabilities` 静默排除出调度
这三个「静默失效」项，加上流式输出保真、上游模型目录、Ollama Cloud 兼容、Antigravity 端点、
Composite 视频端点、日志量控制与五项管理台可用性。不引入新功能、不新增迁移。

**清单里的 2 项前端依赖安全项（`4a1da2950` dompurify / `b410c3913` nanoid 审计例外）
已决定推迟**，理由与重新评估的触发条件见 `design.md` 决策 1。

阅读顺序：`proposal.md` → `source-baseline.md` → `source-feature-map.md` → `design.md` →
十一个 `specs/*/spec.md` → `tasks.md` → `verification.md`。

逐条的 patch site 与上游 diff 说明在 [`docs/upstream-sync/PORTING-0.1.180.md`](../../../docs/upstream-sync/PORTING-0.1.180.md) §5。
**本 change 复核时发现该文档漏了两处会导致编译失败或线上 401 的细节**，见 `source-baseline.md` §4。

与 [`port-upstream-0.1.183-p0-fixes`](../port-upstream-0.1.183-p0-fixes/) 的关系：两份 P0
**无文件冲突，可并行**；唯一要协调的是 0.1.180 §6.3 的调度诊断项（不在本 change 内）。
