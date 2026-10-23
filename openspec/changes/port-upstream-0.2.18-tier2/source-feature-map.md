# 来源与行为映射

| 簇 | 来源 | 行为 |
|---|---|---|
| S01 混合分组 OAuth 凭据协议隔离 | [#8106](https://github.com/Wei-Shaw/sub2api/pull/8106) | 混合归属含只能走 OpenAI 网关的 OAuth 且模型兜底会选该族时拒绝歧义；Anthropic 请求构造在发送前阻止外平台 OAuth；count_tokens 给出可回退的404。 |
| S02 Responses 时间戳与失败终态 | [#8140](https://github.com/Wei-Shaw/sub2api/pull/8140)、[`f2e5b55f3`](https://github.com/Wei-Shaw/sub2api/commit/f2e5b55f3e8069df1c02a9a27af537b5fdefa64e) | 非透传 Responses JSON/SSE/WS桥缺失或null的created_at补齐；同一流共享首次确定时间；合成失败带created_at和sequence_number。 |
| S03 Codex 已退役订阅模型调度一致性 | [#8121](https://github.com/Wei-Shaw/sub2api/pull/8121) | 在实际映射目标上判定订阅账号退役模型；普通/粘性/compact调度和客户端目录一致过滤；不支持时返回model_not_found及迁移提示。 |
| S04 Codex 容量错误恢复与出站身份 | [#8305](https://github.com/Wei-Shaw/sub2api/pull/8305) | TUI出站UA与originator成对归一为CLI并保留生效版本/系统指纹；空语义片段不阻止安全换号；容量错误尊重有效上游重试提示且最多8秒。 |
| S05 Anthropic 兼容流中断结算与终态 | [#8210](https://github.com/Wei-Shaw/sub2api/pull/8210) | 客户端断开后停止下游写入、有界读完上游并收集usage；上游异常终止返回对应协议错误而非成功；有已观测用量的失败也记录一次，输出前可重试错误不重复结算。 |
| S06 既有 Antigravity 协议与别名修复 | [#8093](https://github.com/Wei-Shaw/sub2api/pull/8093) | 仅修已有模型和路径：保留工具base64图片/文档、Gemini函数调用及签名；放行已解析到Antigravity的Gemini路由；已有Claude5.5档位别名用基准价格；明确的不同目标映射优先，自映射可选精确已有档位；账号测试按映射目标选协议并剔除通配符候选；兼容已有上游套餐细分ID。 |
| S07 Gemini 错误策略单次裁决与 Vertex 认证 | [#8060](https://github.com/Wei-Shaw/sub2api/pull/8060)、[#8076](https://github.com/Wei-Shaw/sub2api/pull/8076) | 同一上游响应只裁决一次错误策略；Vertex无quota/error明细的共享容量429只做请求内重试；确切SA不存在时标记认证失效并换号。 |
| S08 Gemini 客户端断开保留已观测用量 | [#8168](https://github.com/Wei-Shaw/sub2api/pull/8168) | 确认由客户端断开且已观测到可计费用量时返回部分结果并结算；记录ClientDisconnect与是否收到finishReason；图片只按已观测结果处理。 |
| S09 现有图片转存下载边界 | [#8223](https://github.com/Wei-Shaw/sub2api/pull/8223) | 默认下载客户端逐跳拒绝非公网目的地；直连在socket层校验实际IP；非图片响应失败，不转存为图片。 |
| S10 工具链与重定向依赖维护 | [#8267](https://github.com/Wei-Shaw/sub2api/pull/8267)、[#8275](https://github.com/Wei-Shaw/sub2api/pull/8275) | Go升级目标1.27.3并同步go.mod/CI/Docker版本断言；只升级follow-redirects到修复版本，双份pnpm overrides保持一致并提交锁文件。 |

混合PR只摘本表行为；其余第三/四档保持暂缓/排除。patch sites仅在PORTING，规格不规定源文件逐行复制。
