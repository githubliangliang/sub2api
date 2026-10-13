# 来源与行为映射

| 簇 | 来源 | 行为 |
|---|---|---|
| S01 | [#7893](https://github.com/Wei-Shaw/sub2api/pull/7893) | 首次安装管理员校验统一 |
| S02 | [#7787](https://github.com/Wei-Shaw/sub2api/pull/7787)、[#7788](https://github.com/Wei-Shaw/sub2api/pull/7788) | 旧版函数调用与拒答转换完整性 |
| S03 | [#7827](https://github.com/Wei-Shaw/sub2api/pull/7827) | 显式工具变更 beta 保留及旧前置 |
| S04 | [#7779](https://github.com/Wei-Shaw/sub2api/pull/7779)、[#7792](https://github.com/Wei-Shaw/sub2api/pull/7792) | WebSocket 逐轮分组价格与图片用量 |
| S05 | [#7703](https://github.com/Wei-Shaw/sub2api/pull/7703)、[#7810](https://github.com/Wei-Shaw/sub2api/pull/7810) | 受约束的上游兼容重试 |
| S06 | [#7939](https://github.com/Wei-Shaw/sub2api/pull/7939) | OAuth 搜索历史回放声明 |
| S07 | [#7793](https://github.com/Wei-Shaw/sub2api/pull/7793) | 组合路由公开模型目录 |
| S08 | [#7711](https://github.com/Wei-Shaw/sub2api/pull/7711) | 删除分组同步清理风控范围 |
| S09 | [#7904](https://github.com/Wei-Shaw/sub2api/pull/7904)、[#7900](https://github.com/Wei-Shaw/sub2api/pull/7900)、[#8002](https://github.com/Wei-Shaw/sub2api/pull/8002)、[#7972](https://github.com/Wei-Shaw/sub2api/pull/7972) | 分组倍率/RPM 编辑异步隔离 |
| S10 | [#7903](https://github.com/Wei-Shaw/sub2api/pull/7903)、[#7818](https://github.com/Wei-Shaw/sub2api/pull/7818)、[#7833](https://github.com/Wei-Shaw/sub2api/pull/7833)、[#7841](https://github.com/Wei-Shaw/sub2api/pull/7841)、[#7840](https://github.com/Wei-Shaw/sub2api/pull/7840)、[#7732](https://github.com/Wei-Shaw/sub2api/pull/7732)、[#7765](https://github.com/Wei-Shaw/sub2api/pull/7765)、[#7839](https://github.com/Wei-Shaw/sub2api/pull/7839)、[#7724](https://github.com/Wei-Shaw/sub2api/pull/7724)、[#7722](https://github.com/Wei-Shaw/sub2api/pull/7722) | 异步页面、草稿、确认与弹窗隔离 |
| S11 | [#8036](https://github.com/Wei-Shaw/sub2api/pull/8036) | GPT-6.1 Sol 上游订阅长上下文统计 |
| S12 | [#8006](https://github.com/Wei-Shaw/sub2api/pull/8006) | Gemini 既有路径的错误、额度和脱敏修复 |
| S13 | [#7738](https://github.com/Wei-Shaw/sub2api/pull/7738) | 监控配置补充已有平台入口 |
| S14 | [#7910](https://github.com/Wei-Shaw/sub2api/pull/7910) | Responses 无 namespace 时避免重建 |
| S15 | [#7854](https://github.com/Wei-Shaw/sub2api/pull/7854) | Chat 流心跳保活 |
| S16 | [#7893](https://github.com/Wei-Shaw/sub2api/pull/7893) | 前端依赖安全维护与旧待办归并 |

所有来源按选定行为裁剪；实际patch sites只在PORTING。混合PR的其它部分维持第三/四档。
S03追加旧#7638；S16接管旧#7780 axios子项与0.2.0阶段9 DOMPurify，来源SHA见source-baseline。不得回写旧报告为已完成。
