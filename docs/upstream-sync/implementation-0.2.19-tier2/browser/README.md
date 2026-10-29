# 真实浏览器验收（2026-10-29）

使用 agent-browser 0.39.2 / Chromium、当前 Vite/Vue/Pinia/i18n/CSS 与独立本地 [API 夹具](./fixture-api.py)。业务 API 与上游 provider 均未连接；后端协议、并发和 SQLite 在独立实际服务测试中验证。不是业务实例端到端验收。

渠道页实际关闭 OpenAI 搜索和 Anthropic 平台 section 后提交 false，unknown future 和 unrelated feature 保留；重新开启 OpenAI，再隐藏全局开关，保存后 true 保留。[第一轮 API](./requests-first.json) / 根目录 browser-channel-save-state、browser-channel-hidden-state。提示文案最终采用来源的“对应平台”语义，[最终截图](./channel-openai-search-final.png)；旧截图是修复文案前的过程快照。

实际 /admin/accounts 的主测试组件默认选择 gemini-3.5-flash，普通目录含五个既有模型；Google One 默认保留 2.5-flash 并可显式选择 2.0-flash；合并目录含 25 个既有 Antigravity 项，总计 247 个唯一项。[默认](./gemini-ordinary-default.png) / [普通下拉](./gemini-ordinary-models.png) / [旧模型](./gemini-legacy-explicit.png)。

截图均用 view_image 目视复核，[哈希与说明](./visual-review.json)。真实页面 JS errors 为空；两轮会话/API/Vite 均关闭并核验端口释放，[清理](./cleanup.json)。首轮 pnpm 父进程退出后的 Vite 子进程按本轮特定环境识别后终止；最终启动器使用独立进程组，避免遗留。
