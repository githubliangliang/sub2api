# 新调用与消费者复核

基线128c5f7bf。完整词法清单9947条，含测试/SQL/类型转换误命中；`hunk-base-inventory.tsv`记录957个生产/文档hunk对应位置。**该清单不是编译验收**，源代码未应用。

| 范围 | 复核与落地要求 |
|---|---|
| F01 | 新增errors.Is使用已有标准库和ErrAPIKeyNotFound；SQLite probe验证既有Apply消费者。忽略Key错误不能吞掉其它数据库错误。 |
| F02/F03/S02 | 新converter/legacy ID方法/refusal累积/MarshalJSON随pkg消费者一起引入；两个CN native缓冲函数不存在，丢弃对应生产与测试hunk。json/fmt/strings均stdlib；同名MarshalJSON命中不能代替接收者类型检查。 |
| F04/S03 | TTL helpers复用collectCacheControlPaths/gjson/sjson；IsSonnet55本地已存在。旧#7638补两个beta常量及窄白名单。有效drop集优先，count_tokens已有不同策略不被覆盖。 |
| F05/F13 | x/net/html本地go.mod已有；SSO do的新headers参数为variadic，旧调用可用，首跳头限制保留。日志helper的logger.FromContext/IntoContext、logredact.RedactText、uuid.NewString、zap字段均已有依赖；只记录脱敏文本。 |
| S04 | APIKeyService.GetByKey与WithOpenAITurnPricingContext已存在；begin/set/forTurn/refresh/withGroup及relay Usage字段需连消费者一起加入，互斥及逐轮归属必须race验证。 |
| S05/S06/S15 | 完整错误体读取用io标准库及已有read-limit函数；所有WithStartTime参数调用一并改。web-search helper264行使用已有marshalOpenAIUpstreamJSON/firstNonEmptyString与gjson/sjson。兼容options同时接map/raw/WS/compact边界；心跳Flush会提交字节，需失败出口测试。 |
| S07 | GetCompositeRouteModels/ListExactPublicModels为新增方法，repo.ListByGroup已有；FilterForListing来自缺失ModelAllowlist，不能照搬。TypeSafe过滤子链不引入；保留本地CustomModelsListEnabled路径。 |
| S08 | lockLiveGroups未命中；上游30行基于PG ANY/pq.Array/FOR SHARE，与SQLite架构不符。SQL内JSONB/to_jsonb均排除，改现有SQLite写事务及JSON读改写，不补同名空实现。 |
| S11 | accountStatsLongContextPricingEnabled未命中，但源函数只有3行return gate==nil||*gate，本地已经内联；不复制整个account_stats_pricing重构。CostInput字段、新豁免helper与本地两条成本路径一起改。 |
| S12 | 新官方URL判定16行仅net/url/strings+host常量，独立抽取；缓存不存在matcher19行和固定响应1常量随ForwardNative消费者。第三方429调用已有SetRateLimited/apply429FallbackRateLimit；拆parseGeminiRetryDelayResetTime时保持官方日限分支。ops四helper随handler登记/日志消费者加入；不补binding/CRUD/lease/inflight新调用。 |
| 前端 | 新增Vue/watch/Promise/DOM API均已有；extractApiErrorMessage/formatDateTimeLocalInput存在。#7738的@/constants/platforms缺失，仅复用本地5平台值；#8006 ModelWhitelistSelector只取catch，不保留来源platformCatalog导入；#7960新组件与其它新平台导入均排除。 |
| S16 | 无新产品调用；pnpm11读取workspace overrides，锁文件通过定向解析生成；不照搬package.json独有overrides或整份源lock。 |

选定小PR中没有本地func/同PR定义的词法项（外部/stdlib方法、类型与SQL需人工区分）：

- PR7779: `WithValue`
- PR7792: `Exists, GetBytes`
- PR7785: `Marshal, TrimSpace, Unmarshal`
- PR7787: `Sprintf`
- PR7869: `Marshal, anthropicContentBlock, claudeContentItem`
- PR7844: `Exists, GetBytes, SetBytes`
- PR7827: `Join`
- PR7910: `DeleteBytes, Exists, Grow, IsObject`
- PR7854: `Fprint`
- PR7887: `EqualFold, NewReader, NewTokenizer, Token`
- PR7837: `TrimSpace`
- PR7878: `GetBool, Int64, NewString, Warn, WithContext, import`
- PR7822: `HasPrefix`
- PR7793: `FilterForListing, TrimSpace`
- PR7703: `Contains`
- PR7939: `Array, Contains, Exists, GetBytes, HasPrefix, IsArray, Join, SetBytes, SetRawBytes, Sprintf, ToLower, TrimSpace, const, import, var`
- PR7711: `COALESCE, NOW, jsonb_agg, jsonb_array_elements, jsonb_build_array, jsonb_set, jsonb_typeof, lockLiveGroups, to_jsonb`
- PR8036: `TrimSpace, accountStatsLongContextPricingEnabled`
- PR8046: `Contains, HasPrefix`
- PR7810: `Atoi, Cut, FindStringSubmatch, ForEach, GetBytes, HasPrefix, HasSuffix, Info, Int64, IsObject, LimitReader, MatchString, MustCompile, ParseBytes, ReadAll, ToLower, TrimSpace, an`

上表已分别处理真实缺失的FilterForListing、lockLiveGroups、accountStatsLongContextPricingEnabled；其它列出的生产调用属于stdlib、既有gjson/sjson/Gin/zap/uuid/html方法，或类型转换/词法误命中。大混合PR只对已选子集作上述调用链承诺；不为排除范围补基座。
