# 第一档后端实现报告

实施基线50c34ad49；产品状态：工作树未提交。范围F01–F08、F15，13PR。
Go工具链1.27.1（go.mod仍1.27.0）。Windows/Linux archive均来自go.dev，官方SHA256已核验。WSL GCC用于后续race。
WSL无法直连下载，Windows `go mod download` 验证依赖后共享 `.git/gomodcache`，Linux使用本地file GOPROXY；未关闭sumdb或修改go.mod/go.sum。

## 基线与TDD

50c34ad49的backend通过git archive提取到独立/tmp/sub2api-tier1-baseline-50c34ad49；完整 `go test -p 4 -tags=unit ./...`、`go build -p 4 ./...` 和四个相关package的 `go test -list` 均exit0。
来源baseline-unit.txt、baseline-build.txt、baseline-list.txt。

先只移植测试，未改产品。修正了两个上游fixture差异（quotaSnapshot本地叫snapshot；OAuth header测试的API Key路径需要本地cfg初始化）。这些编译/初始化错误不计RED。
最终有效RED：backend-red-valid.txt，四个package均为断言失败，无编译错误或panic。
再移植产品，GREEN：backend-green.txt，四个package均pass，exit0。
相同命令由 backend-selected.sh 固定，用 `bash ./backend-selected.sh <log-name>` 执行，完整选择表达式保留在脚本中。

| 簇 | PR | 有效RED证据/覆盖 |
|---|---|---|
| F01 | #7311 | TestSchedulerMetadataAccountPreservesRPMPolicy：两种RPM策略投影前10、投影后0；同时覆盖未配置不凭空加字段。新增本地roundtrip而非引入缺失上游test file |
| F02 | #7404 | TestApplyOAuthCredentialsPreservesExistingNonAuthCredentials：model_mapping/account_id丢失；新认证替换旧值，password/sso/cookie仍剥离 |
| F03 | #7349/#7362 | TestGetOpenAICodexCanonicalUserAgentOutboundIdentity 非法UA丢失已解析版本；TestAccountUsageService_OpenAIQueriesPreserveRefreshError 查询快照清除refresh error |
| F04 | #7538 | 两条bridge thinking disabled被映射medium/xhigh；Responses实际转发未保持none。默认medium及未禁用max规则随现有测试保留 |
| F05 | #7638/#7617 | OAuth mimic结构化beta丢失；ordinary OpenAI多agent头为空、混合/compact/APIKey分支丢字段。测试覆盖unknown token/drop、重复token、无beta、APIKey原样、compact清理 |
| F06 | #7635/#7569 | 角色项Type为空；final空文本/无message时丢delta正文。覆盖已有terminal文本权威和无delta不造message |
| F07 | #7572 | TestAlphaSearchFallbackRequiresSuccessfulCompletion：最初八种失败/截断情形仍返回成功，审查再补两种缺status反例；正常completion输出与citation反例保持 |
| F08 | #7624 | 未来reset+陈旧快照的threshold/utilization误放行；覆盖reset已过/无reset及规范窗口。修改已有Issue2994 fixture以表达无reset的fail-open，未删除该反例 |
| F15 | #7488/#6920 | 65字节的message/reasoning/function/custom ID保留；Grok四种调度暂停无法读配额。补64字节/未知类型/原始body不变；无认证/过期token/缺代理仍拒绝 |

## 产品适配

- 各PR产品均按对应merge^1 diff最小移植，未整文件覆盖tag终态。
- #7311仅补现有filterSchedulerExtra字段。
- #7617文件级反向ALREADY误匹配passthrough map；真实失败表明ordinary openaiAllowedHeaders缺项，因此手工将openai-beta加入该map并保留源PR strip调用。原始冻结四态不改，验收文档会更正语义判断。
- #7488适配本地分支结构，在三个已识别类型的前缀检查中加len(id)>64；空ID和未识别类型仍不剥离。
- #6920源测试引用quotaSnapshot本地无此方法，使用已有snapshot，不新增测试用产品API。
- #7638新增常量及窄白名单兼容，未知Anthropic beta与policy drop保持原有安全边界。

## 当前结果

所有目标回归已pass。最终门禁结果以同目录README.md及OpenSpec verification.md为准。
产品diff和新增文件见backend-review.diff；未修改依赖、迁移、VERSION、Ent/Wire生成文件或支付代码。

## F07 审查修复：完成事件必须明确成功

审查发现 `response.completed` 携带空 response 或带 output 但缺少 status 时，旧条件仅在 status 存在时拒绝非 completed，仍可能返回成功与可计费结果。为满足冻结规格“completed 且状态成功”，本次在上游原 patch 之外收紧为 `response["status"] == "completed"` 才接受。未知状态、failed、incomplete 和提前 EOF 的拒绝逻辑未放宽。

按 TDD 先新增 `missing_status`（先有 delta，再收到空 response）与 `missing_status_with_output`（终态有 output 但无 status）两条回归；断言错误、nil 计费结果及未写成功响应。成功 fixture `alphaSearchResponsesSSE` 补充明确的 completed 状态，原有 output/citations 断言保留。

- RED 命令：`wsl -d Ubuntu --exec bash /mnt/d/work/work_space/github_code/sub2api/./run-go.sh alpha-status-red /mnt/d/work/work_space/github_code/sub2api/backend go test -p 2 -tags=unit ./internal/service -run AlphaSearch -count=1`
- RED 结果：exit 1，仅上述两个新增用例失败，均为 `An error is expected but got nil`，无编译错误或 panic。日志 `./alpha-status-red.txt`。
- GREEN 命令：`wsl -d Ubuntu --exec bash /mnt/d/work/work_space/github_code/sub2api/./run-go.sh alpha-status-green /mnt/d/work/work_space/github_code/sub2api/backend go test -p 2 -tags=unit ./internal/service -run AlphaSearch -count=1`
- GREEN 结果：全部 AlphaSearch 测试通过，exit 0，package 运行 0.075s。日志 `./alpha-status-green.txt`。

本次仅修改 `openai_alpha_search.go`、`openai_alpha_search_completion_test.go`、`openai_alpha_search_test.go` 及本报告；未 stage、commit、push。完整回归由 root 执行。
