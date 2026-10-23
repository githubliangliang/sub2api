# 来源与行为映射

| 簇 | 来源 | 行为 |
|---|---|---|
| F01 Codex 工作区导入身份隔离 | [#8088](https://github.com/Wei-Shaw/sub2api/pull/8088) | 以 workspace/account 与 user 的组合区分完整 OAuth 身份；同一工作区重复导入更新原账号，不覆盖其它工作区。 |
| F02 Codex 前导元数据不阻止重试 | [#8139](https://github.com/Wei-Shaw/sub2api/pull/8139) | 排队和审查元数据保持首个语义输出之前的状态，后续容量错误仍可安全重试，首输出超时仍有效。 |
| F03 Grok Fast 用量成本价卡 | [#8170](https://github.com/Wei-Shaw/sub2api/pull/8170) | 既有 Grok 模型映射/透传请求按 Fast 独立价卡计算；短上下文 input/cache/output 每百万为 4/1/12，达到 200k 后为 6/1.5/18（来源补丁值）。 |
| F04 Cloudflare 1010 不处罚账号 | [#8255](https://github.com/Wei-Shaw/sub2api/pull/8255) | 识别明确的 Cloudflare 1010 或 browser_signature_banned，所有既有平台跳过账号处罚，保留请求错误和换号处理。 |
| F05 Gemini 缓存命中用量非负 | [#8177](https://github.com/Wei-Shaw/sub2api/pull/8177) | 仅缓存计数或缓存大于 prompt 时非缓存 input_tokens 下限为0，同时保留缓存、输出和图片用量。 |
| F06 Gemini 工具 Schema 常量和数据边界 | [#8188](https://github.com/Wei-Shaw/sub2api/pull/8188) | 字符串const转为单值enum，与已有enum求交；移除不支持的非字符串const和注解，区分Schema节点、属性名和数据值。 |

混合PR只摘本表行为；其余第三/四档保持暂缓/排除。patch sites仅在PORTING，规格不规定源文件逐行复制。
