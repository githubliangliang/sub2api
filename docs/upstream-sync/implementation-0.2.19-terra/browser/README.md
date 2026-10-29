# Terra 真实浏览器验收（2026-10-30）

使用 agent-browser 0.39.2 / Chromium、当前 Vite/Vue/Pinia/i18n/CSS，API 为独立本地 [夹具](./fixture-api.py)。账号及密钥均为合成数据；后台 provider、业务实例和业务数据库未连接。后端协议和结算由实际服务/SQLite 用例另外验收。

实际 `/keys` 打开“使用密钥”并选择 OpenCode，生成 JSON 成功解析，Terra context/output 为 1050000/128000，六档为 none/low/medium/high/xhigh/max，Sol 原配置及未指定全局 model 保留。[解析证据](../browser-opencode-config.txt) / [截图](./opencode-terra.png)。Codex 原模板仍为 `model = "gpt-5.5"`，[默认值](../browser-codex-default.txt)。

实际点击“复制”后，在测试页包装原 `navigator.clipboard.writeText`，透传调用并观测其 promise 成功完成；实际写入参数与渲染配置逐字一致，Terra 条目解析正确。[原生写入完成与内容证据](../browser-opencode-copy-proof.txt)。浏览器禁止读取剪贴板的过程输出保留；这不是产品复制失败，也未把读取失败记录成 PASS。

实际 `/admin/accounts` 编辑 API Key 账号，从模型列表选择 Terra，提交后持久化 canonical 身份映射并重开确认。[白名单保存](../browser-terra-whitelist-state.txt) / [重开](../browser-whitelist-reopened-proof.txt) / [截图](./whitelist-terra.png)。切到映射模式，点击 GPT-6.1 Terra 预设，源/目标均为 `gpt-6.1-terra`；实际提交成功，[两次 PUT 的保存结果](../browser-terra-preset-persisted.txt) / [预设截图](./mapping-terra.png) / [API 请求](./requests.json)。

首轮夹具未提供 load_factor，表单保存回传 0 后重开触发既有 min=1 HTML 校验；[定位](../browser-invalid-field-context.txt) 后将该合成账号表单填为 1，完成第二次保存。相关产品表单和校验代码未改。首轮过渡动画遮挡及验证过程输出保留，最终验收采用成功输出。

三张截图均已通过 view_image 目视复核。[页面 JS errors](../browser-final-page-errors.txt) 为空；[独立浏览器关闭](../browser-close.txt)、[API/Vite 停止及端口释放](./cleanup.json) 完成。此处为真实前端交互加本地 API 夹具验收，不作为业务实例端到端可用性证明。
