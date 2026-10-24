# 来源与行为映射

| 簇 | PR / merge SHA | 行为 | 范围边界 |
|---|---|---|---|
| S01 Sonnet 5.5多入口兼容 | [#7683](https://github.com/Wei-Shaw/sub2api/pull/7683) `dd6cb410e6db51b5c5662f4edce562b3b68c271b` | Sonnet 5.5目录、定价、effort、thinking签名与工具beta在本地已有入口中一致。 | 使用映射后的模型；保留Opus 5.5和其它模型契约，排除不存在的原生Anthropic平台路径。 |
| S02 GPT-6.1 Sol及账号套餐元数据 | [#7730](https://github.com/Wei-Shaw/sub2api/pull/7730) `327c32218406cef47186736c7473275cd1b39918` | 现有网关路径识别GPT-6.1 Sol并一致执行请求能力、模型目录与价格规则，账号套餐显示保留SKU身份。 | 先完成0.2.9 F04两桥契约；不引入Astra Ultrafast/Fast发送策略、新价格引擎或支付功能。 |
| S03 流式用量输出与计费一致 | [#7380](https://github.com/Wei-Shaw/sub2api/pull/7380) `1f955af353f0e0a3f6137aa26090e85aee1109cc` | 现有Anthropic到Responses/Chat桥在收到真实usage时转发，并将权威prompt总量与缓存桶统一到计费口径。 | 只补695ebede7的14行通用字段及现有消费者；不引入缺失native平台。保留取消drain、成功去重和终态结束。 |
| S04 Antigravity首内容前保活边界 | [#7643](https://github.com/Wei-Shaw/sub2api/pull/7643) `883a53fea0aeb26b1268f582e2676ed2fb21d5c2` | 兼容流等待首内容时15秒发送注释保活，并在2分钟无语义内容时终止；已提交响应后的失败通过单次SSE错误报告。 | ping不算首token或语义输出；HTTP 200提交后不得换号重放。保持既有断开和用量处理。 |
| S05 Composite模型归属与多轮WS路由 | [#7378](https://github.com/Wei-Shaw/sub2api/pull/7378) `ccb027b51506b2f64c16290ac33b185c547d8963` | OpenAI两种调度执行公开alias归属检查，WS按responses/any解析路由并隔离公开模型与出站映射。 | 只支持既有OpenAI/Grok目标，不引入model_allowlist或插件；Wire依赖正确生成。 |
| S06 Codex远程目录与文件兼容 | [#7680](https://github.com/Wei-Shaw/sub2api/pull/7680) `41dcfec4895a868ee8b4a92526d0aecd9afb4418`；[#7736](https://github.com/Wei-Shaw/sub2api/pull/7736) `48b71095887f8f150a26d3f471a344565c31ea58` | 提供Codex 0.156.0+远程目录和旧客户端本地文件选择，采用v0.2.11最终行为。 | 不单独实施#7680隐藏OpenAI目录的中间态；保留旧S08 URL/Windows/CC Switch契约，不恢复缺失平台。 |
