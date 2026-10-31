## ADDED Requirements

### Requirement: 前端 HTML 净化器必须只有一份且是已打补丁的版本
依赖树中 `dompurify` SHALL 解析为**唯一一份** `>= 3.4.14` 的副本。直接依赖与 mermaid 的传递依赖
MUST 去重到同一版本，MUST NOT 同时存在两个不同版本的副本。

依据：本仓库有 7 个 `DOMPurify.sanitize` 调用点，其中 `/legal/:documentId` 是 `requiresAuth: false`
的公开页。净化跑在**访问者**的浏览器里，所以「输入由管理员自填」不等于「受众只有管理员」。
升级前树里有两份副本（直接 `3.3.1` + mermaid 的 `3.3.3`），合计命中 18 条 sanitizer-bypass advisory。

#### Scenario: 依赖树中的副本数
- **WHEN** 检查 lockfile 中的 `dompurify` 解析结果
- **THEN** MUST 只出现一个版本
- **THEN** 该版本 MUST `>= 3.4.14`

#### Scenario: mermaid 的传递依赖
- **WHEN** 检查 `. > @lobehub/icons > @lobehub/ui > mermaid > dompurify` 这条路径
- **THEN** MUST 解析到与直接依赖相同的版本

#### Scenario: 审计结果
- **WHEN** 运行依赖审计
- **THEN** MUST NOT 再报出任何 `dompurify` advisory

### Requirement: 版本约束必须在两处 overrides 中同时声明
覆盖约束 SHALL 同时写入 `frontend/package.json` 的 `pnpm.overrides` 与
`frontend/pnpm-workspace.yaml` 的 `overrides`。两处 MUST 保持一致。

依据：本地 pnpm 11 **不再读** `package.json` 的 `pnpm` 字段（安装时会打印 `"pnpm" field ... ignored`），
只读 workspace 那份；而 CI 是 `pnpm/action-setup@v6` + `version: 9`、`Dockerfile` 是 `pnpm@9`、
`deploy/Dockerfile` 是 `pnpm@9.15.9`，pnpm 9 只读 `package.json` 那份。**只改一处 = 一边去重、
另一边没去重。**

#### Scenario: 本地安装
- **WHEN** 用本机 pnpm 11 安装
- **THEN** 去重 MUST 生效（约束来自 workspace 那份）

#### Scenario: CI 与镜像构建
- **WHEN** 用 pnpm 9 安装
- **THEN** 去重 MUST 生效（约束来自 `package.json` 那份）

#### Scenario: 两处不一致
- **WHEN** 两处 overrides 的 dompurify 约束不同
- **THEN** 视为缺陷（本变更 MUST 保持两处同值）

### Requirement: 三个依赖文件必须在同一个提交内一起改
`frontend/package.json`、`frontend/pnpm-workspace.yaml`、`frontend/pnpm-lock.yaml` SHALL 在同一个
提交内一起变更。lockfile SHALL 由 `pnpm install --lockfile-only` 重新生成，MUST NOT 从上游抄
（上游那份是 pnpm 9 产物，本仓库本地是 pnpm 11）。

依据：3 个 workflow 与 2 个 Dockerfile 共 5 处使用 `pnpm install --frozen-lockfile`
（`security-scan.yml` / `backend-ci.yml` / `release.yml` / `Dockerfile` / `deploy/Dockerfile`），
只改 `package.json` 会让 CI 与镜像构建立刻失败。

#### Scenario: frozen-lockfile 安装
- **WHEN** 在改动后的树上运行 `pnpm install --frozen-lockfile`
- **THEN** MUST 成功（lockfile 与 manifest 一致）

#### Scenario: lockfile churn 范围
- **WHEN** 比对 lockfile 变更
- **THEN** 变更 MUST 只涉及 dompurify 相关条目与 overrides 块
- **THEN** MUST NOT 出现其它包的版本变化

### Requirement: 净化行为不得改变
升级 MUST NOT 改变任何现有调用点的净化输出。特别地，`utils/sanitize.ts` 的
`DOMPurify.sanitize(svg, { USE_PROFILES: { svg: true, svgFilters: true } })` 对既有输入的输出
MUST 与升级前一致。

依据：已在 jsdom 下把 `3.3.1` 与 `3.4.14` 并排跑过该调用，良性 SVG 与带 `onclick` 的 SVG
两种输入的输出**逐字节相同**。

#### Scenario: SVG 图标渲染
- **WHEN** 侧栏自定义 SVG 图标与 SVG 上传预览渲染
- **THEN** 渲染结果 MUST 与升级前一致

#### Scenario: 公告 / 法律文档 / 自定义页 / 模型广场
- **WHEN** 这四类富文本渲染
- **THEN** MUST 与升级前一致（含 `CustomPageView` 的 `ADD_TAGS:['iframe']` 仍然允许 iframe）

### Requirement: 不得把本项当作让审计门禁转绿的手段
本项 SHALL NOT 修改 `.github/audit-exceptions.yml`。CI 门禁是
`pnpm audit --prod --audit-level=high`，而 dompurify 的 18 条 advisory **全部是 low/moderate**
⇒ 它们从来没有进入门禁，升级前后门禁结果 MUST 相同。

依据：门禁当前报出的 prod+high 是 `xlsx` ×2 与 `nanoid` ×1，三条都由现有例外条目覆盖
（`expires_on: 2026-10-06`）。本项与它们无关。

#### Scenario: 门禁结果
- **WHEN** 升级后运行 `pnpm audit --prod --audit-level=high`
- **THEN** 报出的 advisory 集合 MUST 与升级前相同（`xlsx` ×2 + `nanoid` ×1）

#### Scenario: 例外清单
- **WHEN** 检查 `.github/audit-exceptions.yml`
- **THEN** MUST 与升级前逐字节相同（本项不新增、不删除、不续期任何条目）
