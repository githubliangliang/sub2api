# 浏览器验收

2026-10-18，agent-browser 独立会话 `sub2api-tier2-ac5eb5d836c2`，Vite `127.0.0.1:4179`，真实 Vue 组件、项目 CSS 与定向更新后的依赖。API 使用隔离适配器与模拟数据，未登录或操作业务实例。

- S09：打开 Group A 时挂起列表响应，保持弹窗打开并切换到 Group B。B 行显示 `user2@example.com` 后释放 A 的旧响应，仍只显示 B。编辑倍率为 3，实际保存载荷为 `/admin/groups/2/rate-multipliers: {"entries":[{"user_id":2,"rate_multiplier":3}]}`，保存后弹窗关闭。[截图](./group-rates.png)
- S10：提交用户名 `Submitted`，挂起响应，继续输入 `NextDraft`；释放保存响应后，服务器用户字段为 `Submitted`，输入框仍是 `NextDraft`。[截图](./profile-draft.png)
- S13：已有 OpenAI 时下拉仅列 Claude/Gemini/Antigravity/Grok；添加 Gemini 后初始模型为空。保存后的载荷保持 OpenAI 并包含 `{ "platform":"gemini", "enabled":true, "models":[] }`。[截图](./monitor-platform.png)
- 所见组件无水平溢出、字段遮挡或错误覆盖层；工具栏是临时测试入口，不是产品页面的一部分。
- 临时 HTML/Vue/TS 入口已删除，浏览器会话已关闭，Vite 已停止。首次资料请求模拟路径误设为 `/user/profile`，修正为实际 `/user` 后重新完成了以上有效验收。

浏览器记录覆盖实际渲染与交互；其它异步成功/失败、卸载、分页、输入边界由组件测试覆盖。未进行真实上游凭据或生产业务写入验证。
