# 来源到行为映射

| ID | PR / 固定merge SHA | 行为 |
|---|---|---|
| S01 模型目录和计费别名 | #7509 `4318a63bd886b1a64c49015979c61bf34eca19ff`；#7466 `1716e991517096eec6ec92d025b62b24f70b63f0`；#7597 `69060621872d80a66eed31324c0767bb82e3a810`；#7568 `11608c51f0dd8b3de7ec9c8e3b849327bd94cec1` | 接入 GPT-6 Sol/Luna、Opus 5.5、Grok 4.7 的现有平台目录/能力/价格别名，并将 GPT≥5 按推理模型桥接 |
| S02 工具 schema 兼容 | #7345 `d5cea617a5795ba1c159f90a20efeae54df138c1`；#7489 `fa3f526785814cfc21072b3215e5ac6fe182a3b0`；#7077 `da25b18db7dea8570de1edafab1837ab9c55e132`；#7393 `6655f4ec0eee4e1313fb949a85c9e3a304cb9061` | 处理根级 union、required:null、tuple/prefixItems，并保留 string const 与 enum 的交集 |
| S03 工具参数和 PDF 保真 | #7272 `51f73840a17d237dc44a1a9d2ce38bfc400e696a`；#7570 `00ae7c38e2886e687316603d40d8b622d92f1bf7`；#7585 `51a10e42f70e76aef8db2af5ef4fcea643baf00e` | arguments.done 等于已发送 delta；仅 block_start 携带参数也保留；内嵌 PDF 转为 document/inlineData |
| S04 流结束、取消与错误协议 | #6473 `e26abaef7d50d99c41b3e814b1d9dd46956438f5`；#7303 `583398d1869f6361146dd09cf6a58ca6a21a61ac`；#7276 `13be6ca27c8fa5b6940773f8ffc79974454eb01b`；#7300 `5ad7cf4fb9653dceb575b12a281fd7c798ec8839`；#7339 `231e72fdde89bd2c4cc172dd1f51ff90e5d0b0a9`；#7097 `b12a187d4f2d61d1b309a1698c680a99b7d019ac`；#7448 `9d9e90960e397e6adcda71e3de43af26a27ca071`；#7609 `79c18ec836c0f700ee41ced949d603c598cf368c` | terminal 完整发送后结束；先取消单次请求再关响应体；心跳不算语义输出；错误只发一次；Gemini 传输失败换号；客户端取消归类 499 |
| S05 Antigravity 裸模型和空流 | #7258 `74431ffad8b9446100a6c1416fe69e908c03ae81`；#7432 `3442a6a53e2555300217268cb763012383850f0a`；#7278 `0c9e83b59cf9f8647434008a7377ad481b291d3f`；#7607 `5a350108c43c3cbb1ec8d31c0c1b16c7a6cedd04` | 各入口将裸 Gemini 名按 thinking 配置映射；识别特定 SDK 的心跳兼容；MALFORMED_FUNCTION_CALL 空流触发 failover |
| S06 调度路由和账号可用性 | #7427 `b0e32fe5be4017e21a89f383fa5952444bca9d15`；#7481 `d8eeb0989b07841467ba2742d692b95f9fe8d4f4`；#7139 `b350e079f868bf8ecba7917d48106a5d5c6c3df6`；#7195 `9efc167482dfe73f3e315b220f38cfd442865827` | 非高级调度遵守 previous_response 归属；热路径用轻量分组读；API Key 未知模型 401 不误禁账号；暂停 OAuth 仍刷新 token |
| S07 账号成本与图片价继承 | #7619 `e4b34446afc993552a414d40be4cfd74f4fa566c`；#7573 `1ff2bd6d95989d7a8a457f7bb3ec22cb0fd2b899` | 账号成本是否收取长上下文溢价取决于账号 gate；渠道图片价未填继承目录，显式 0 才免费 |
| S08 Codex 和 CC Switch 配置 | #7395 `2448a38b55fd8b62a0554a6e87edd8c53ba65a15`；#7322 `ba5737fe14b859b46de99648767c2f147f59a682`；#7622 `f45126c8ec29914a9c266bbe7d0525f50f3cd7f2`；#7628 `ec334fda2caa9b88aa55c41cc89102c76267cc69`；#7549 `66f07efd264086258c25107932348fc6e19316ac` | Codex config 请求正确 /v1；CC Switch保留配置端点（含显式/v1及子路径）、仅去尾斜杠且不自动追加/v1；usage路径恰好一个/v1；Windows catalog用~/ |
| S09 代理回退一致性 | #7475 `fcd07ba0aff605de99640370a23f9293fa5fb9e2`；#7341 `ed2360d7ea752b7632097fbcd34c001c395dbf0d`；#7342 `a079a6596a8085aa5e27bf21798d360c6cce1ad8` | 代理恢复改变网络身份时失效探针；过期扫描写入前重新核对快照；禁用代理不可作为回退目标 |
| S10 Responses 探测未知态 | #7571 `9375e288ba6f16798917193e5abb1f76d1c14161` | 探测模型不存在时不写入endpoint能力标记（未知或已有值都保留），优先选普通GPT文本模型 |
| S11 公开响应模型别名 | #7304 `bbdcfbac0a3c0982a33639fba372d55d25724e2f` | Responses 与 Chat 返回请求方的公开模型别名，同时保留真实上游模型用于记录 |
| S12 现有内容审计输入边界 | #7314 `7403a011775c2543cb0e212e8cf1dd52f1a13cc1`；#7487 `bb1f40e35f1467c30ced2782803c3abd48c47771` | 关键词检查不丢弃客户端 reminder 内文本，尾随 system 不遮蔽当前用户输入 |
| S13 Antigravity 系统身份兼容 | #7256 `aeab57b9b21a726d8e55bcc0909bc49523c1b30b`；#7411 `d6e8b44bcb1247bc4f9730ce93e9ed7558cdb7f8` | 仅在 Antigravity 转换中移除前导 attribution 并中和前导 SDK 身份 |
| S14 共享弹窗与标签收尾 | #7319 `2b12de14eed7fc915147d82f647057bcaa078dec`；#7498 `fc4465f78bf0f1233f3ebd3fc6b4bf2a97b6ad06` | 嵌套弹窗关闭后滚动锁计数正确，模型标签 IME 不提前提交 |
| S15 Codex WS 上下文切换 | #7615 `ff6b85e30b4add5bfda9af98c127d0a336b45012` | window_id 改变时删除旧 previous_response_id 并清空旧续聊推断锚点 |
| S16 Codex 请求边界兼容 | #7038 `02d9901b49787d1eaa0d1553307d733f78f09aa8`；#7426 `680a992fc94a359a6ef2c4ca178fe9681a7130a7` | 映射到 GPT-5.5 时移除不兼容 Lite 标记但保留历史与工具；turn metadata 序列化保留非 ASCII 转义 |
