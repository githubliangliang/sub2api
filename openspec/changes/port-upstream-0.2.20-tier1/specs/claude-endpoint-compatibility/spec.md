## ADDED Requirements

### Requirement: F12 Anthropic兼容 Cloudflare REST URL

现有Anthropic自定义base URL消费者 SHALL 仅在精确Cloudflare账号REST messages路径省略beta查询，其余上游保持原行为。

#### Scenario: 具体边界与可观测结果

- **WHEN** 已验证base URL精确为HTTPS api.cloudflare.com的client/v4/accounts/账号/ai路径
- **THEN** 三条messages调用链省略beta查询；native gateway、相似域名和不匹配路径保持原有beta行为

#### Scenario: 正反例与相邻行为

- **WHEN** 对该消费者执行边界回归和相邻路径验证
- **THEN** 精确HTTPS host/443/账号路径、相似域名/额外path/query/fragment/rawpath/userinfo负例、Cloudflare native gateway保留beta；三条调用链一致。
