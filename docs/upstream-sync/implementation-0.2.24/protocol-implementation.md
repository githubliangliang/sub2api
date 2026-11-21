# v0.2.24 协议与窄热点实施记录（2026-11-21）

实际起点为 `d2488b1105f091b6fce4445395fec1835306fcf8`。此记录仅说明 F03/F04/F06/S03/S08/S12/S13 的本地实现与验证；不改变冻结 intake、source-baseline、VERSION1.1.26、Go1.27.4、SQLite230、auth21、依赖或生成工具。未提交、推送、发布、归档、操作业务实例或真实凭据。

## 根因与实现

| 簇 | 真实缺口与最小适配 | 回归与控制 |
|---|---|---|
| F03 | `normalizeCodexPageContext` 未识别 IDE/独立文本。保留本地名字，扩充 ide_context/untrusted_content/proposed_plan；使用 `*string` 区分空串与null/错误类型。输出仍为不可信user文本。 | `openai_page_context_024_test.go`；原 page-context/real call_id/duplicate-member/page snapshot 测试保留。IDE仅非空白，untrusted_content可空，proposed_plan非空但可空白。 |
| F04 | 本地已采集原始HTTP/每轮WS与SQLite230 requested/effective；只缺 input 的 configuration_update。取最后一个可归一化有效effort，忽略空/无effort/非法值；原顶层reasoning/reasoning_effort/output_config/模型后缀兜底保持。 | `openai_protocol_024_test.go`；多update、none/max、null/空/非法值、Unicode转义、本轮省略、字段归属；原WS effort与政策四路径等价控制。最终input标量克隆，避免新路径重新保活body。 |
| F06 | Responses DTO曾丢stop；现用RawMessage保留字符串/数组并在回Chat时回填，Chat→Responses→Chat可保留。真正native Responses的Chat桥清除stop；null按省略。既有OpenAI APIKey桥提升system消息的动态function声明，保留parameters原始数值与顺序。 | `stop_024_test.go`、`openai_buffered_024_test.go`；原Cursor/Grok白名单回归。原生Chat透传路径仍先分流，OAuth不提升动态声明；没有新增Kimi平台或Anthropic-native消费者。 |
| S03 | 原buffered accumulator只有聚合文本/函数，丢hosted done item，并把不同content_index合并。适配冻结474行output-slot helper，在done/added/delta身份与索引之间重建空终态。有效非空终态优先，failed不填充。function object arguments保留；custom input使用原有wrapper。 | `openai_buffered_024_test.go`通过真实loopback HTTP/SSE读取器覆盖hosted/custom/顺序/独立part/终态与Grok search控制；`responses_to_anthropic_tool_input_test.go`验证对象/截断/空/自定义文本。provider-defined hosted字段用独立RawMessage保留，涵盖web/file/computer/code/image/x-search及大整数，不新增ToolSearch消费者。 |
| S08 | metadata归一化曾删audio；静态Sol订阅目录与APIKey能力需要区分。保留audio元数据，在可服务候选上作保守交集；只有原生Responses路由可声明audio。无metadata的OAuth Sol使用已冻结订阅声明，APIKey无证据不声明。 | `openai_codex_audio_024_test.go`：OAuth/APIKey/混合、缺metadata、raw Codex目录/OpenAI list/nested modalities、强制及探测Chat退化；原manifest metadata/mapping/ETag测试保留。 |
| S12 | raw view标量引用请求体；usage UserAgent/IP取input字段地址会保活整个input；审计clone/snapshot标量可能借用body。仅克隆持久标量、取局部字段地址，并在桥接序列化后释放不再需要的Input引用。 | 1MiB请求体修改后view字段不变、requested input effort归属、审计clone/snapshot字段归属；保留请求内raw ranges及异步审计body自己的已有深拷贝，不全面复制body或引入COW。 |

## 失败、适配与实际命令

- `protocol-before.log`：首次新增回归已编译，F03/F04/F06/S12目标实际失败，控制保持；不是最终全量门禁失败。
- `protocol-audio-before.log`：原目录丢audio及Chat错误保留audio目标实际失败；源夹具的 `OpenAIModelsResponse` 已适配为本地 `CodexModelsManifest`。
- `protocol-buffered-before.log`：loopback SSE的hosted item丢失和content_index合并实际失败；不是源夹具编译错误。
- `protocol-after.log` 第一轮曾暴露源 `resolveUpstreamProtocol/APIProtocolResponses` 在本地不存在，改用已有 `shouldForwardOpenAIResponsesViaRawChatCompletions`；另一次读到并行代理尚未完成的QuotaFetcher构造器。中间编译失败的工具输出摘录另记 `protocol-fixture-adaptations.md`；本文件在反复修正中复用了after日志路径，现存after.log是最后一次4包通过的完整输出，不冒充保留了每次原始失败日志。
- `protocol-neighbors.log`：handler/service/apicompat/securityaudit四包相关既有邻域与新增回归通过。
- `protocol-hotpath-candidate.log`：Cursor旧fixture须补stop，Grok native Responses的stop:null须按省略；精准修复后两项旧控制通过。
- `protocol-hotpath-equivalence.log`：候选还在test-only阶段时，本地旧实现与候选差分通过。
- `protocol-hotpath-adopted.log`：产品采纳后差分、真实miniredis account/snapshot/legacy hydration、模型支持/成员/调用者图、JSON转义/freeform/非法JSON与范围刷新通过。`TestFilterThinkingBlocksSignedHistorySkipsDecode`定向也通过。

测试均在backend目录离线执行：

```
GOPROXY=off GOSUMDB=off GOTOOLCHAIN=local \
/home/liang/go/pkg/mod/golang.org/toolchain@v0.0.1-go1.27.4.linux-amd64/bin/go \
test -tags=unit -p 2 <相关包> -run <日志记录的测试前缀> -count=1
```

service最小测试之前两次观察到共享工作区中间编译状态（未定义转义helper、未完成构造器/夹具），已分别替换为本地必要条件及等同本地签名；它们不算产品行为回归。默认lint另发现 `openai_buffered_024_test.go` 使用 unit-only 的 `forceChatResponsesFallbackAccount` helper，已给该文件补与其夹具一致的 `//go:build unit`；属于测试标签适配，不涉及产品逻辑。完整组合unit/race/lint/build与全树哈希由主验证记录负责，本文件不把定向通过写成全量验收。

## S13：逐hunk等价与收益

`hotpath_024_test.go`保留修改前的独立函数作为差分基准，生产函数只有通过等价与成本检查后才采纳。单P基准使用约655KiB合成Claude/Codex请求；scheduler为40组展开对象，不是生产样本。`protocol-hotpath-benchmark.log`为采纳前候选的两次测量：

| hunk | 单P约耗时，前→后 | 分配，前→后 | 决定及边界 |
|---|---|---|---|
| scheduler组投影 #8845 | 0.914–0.922ms→0.0494–0.0496ms | 579090B/22次→21880B/23次 | 采纳只投影序列化副本。GroupIDs、AccountID/GroupID/priority/CreatedAt保留；原Groups/AccountGroups关系图不变；旧展开cache仍可读。 |
| Claude TTL顺序 #8862 | 1.125–1.132ms→0.691–0.692ms | 464B/9次→208B/4次 | 采纳已排除转义后的无1h必要条件；thinking块限制仍执行，存在1h/Unicode转义仍完整处理。 |
| 图片意图 #8866 | 0.983ms→0.643–0.644ms | 0→0 | 仅采纳必要条件门；Unicode转义或非ASCII保守回到完整扫描，已有Grok namespace语义不变。没有采用来源“非ASCII模型可能漏判”的宽松假设。 |
| request根字段扫描 #8868 | 1.182–1.184ms→0.566ms | 候选阶段216B/2次相同 | 采纳单次根遍历；真实产品另克隆持久标量，防止保活body。[1m]改写重新扫描；只在已验证相同字节且body未改写的Responses/Chat handler使用Validated入口；发生effort政策改写则走checked，后续Replace/refresh仍校验。 |
| namespace #8870 | 2.381–2.391ms→1.381–1.399ms | 1991136B/9次→663824B/5次 | 采纳raw view与单次prefix/item/suffix重建；保留本地首次实际修改前的完整JSON校验，不能用来源替换删除本地非法输入保护。 |
| schema #8872 | 0.995–1.015ms→0.520–0.521ms | 663664B/3次→0 | 只给本地sanitizer添加null/regex必要条件；原字段/实例数据/freeform保持。未导入完整来源parser、跳过二次完整校验或JSON string decoder重写；本地结构与来源不相同，没有将整簇机械替换。 |
| handler validated复用 | 见最终基准validated行 | 只保留必要标量克隆 | 同一局部body已通过gjson校验，只有未发生effort政策改写时复用Validated；改写时重新checked。未新增image validated公开入口：图片普通负例已有必要条件门收益，无需再扩API。 |

`protocol-hotpath-final-benchmark.log`保留采纳后的复测；当时组合构建/测试并行，绝对耗时显著波动，图片有一次反向结果，不能用该次当成确定性能承诺。`protocol-hotpath-image-repeat.log`补充4次测量，旧路径1.008–1.035ms，新路径0.650–0.782ms，均0分配。分配下降与语义差分分别独立验证。

短fuzz命令：`GOMAXPROCS=2 ... go test -tags=unit -parallel=1 ./internal/service -run '^$' -fuzz '^FuzzHotpath024Gates$' -fuzztime=5s`，输出在 `protocol-hotpath-fuzz.log`，8个种子及3986次执行通过（27个新增interesting输入）。种子包含Unicode/正则/普通字面量，测试内JSON编码随机字符串并比较完整旧sanitizer与新路径，保留properties ID/id与default实例键空间。

这些函数级单P数据不是1C1G生产实例内存上限、实际业务吞吐或生产收益保证。64MiB body/40MiB SSE/8MiB暂存/120s idle及固定quota/dedup策略未在此范围调整。S01入口与WS挂载由主代理实施，本代理已只读复核各链位于Composite改写之前及首/后轮/session.update检查位置。

## 复核补充：嵌套父字段读取语义

独立review提出需要按父字段seen停止读取重复对象。实际本地gjson小程序证实，`metadata={}`、null、数组、字符串或仅不同大小写子键时，`gjson.Get("metadata.user_id")`会继续寻找后续同名父对象，直到匹配子键；子键显式null/空串则Exists=true并保留首值。因此现有 `!metadataUID.Exists()`（以及thinking/output_config/systemInstruction同样）与旧路径读取一致，按父字段seen停止会破坏等价。已补四个嵌套父字段的空对象/null/数组/显式子null差分案例并通过（`protocol-root-scan-review.log`）；S01仍在外部拒绝这些重复父字段。实际小程序输出存于 `protocol-gjson-parent-semantics.log`，没有据静态假设改产品。

Validated前置条件另按review收紧：只有初次校验后的相同字节且effort政策未改写body时使用；主代理已在两个handler增加checked回退。本记录不把JSON序列化合法性等同于spec要求的同字节证据。
