# 本地浏览器检查

使用agent-browser运行真实Vite/Vue页面；API由独立127.0.0.1:33877本地固定数据桩提供，Vite使用独立33876端口。未登录或修改业务实例、未使用真实凭据、未调用真实搜索服务。浏览器与两项临时服务检查后已关闭。

| 检查 | 实际结果 | 证据 |
|---|---|---|
| 管理员使用日志 | 100输出Token、2000ms总耗时、1500ms首字，显示50.0 tok/s | usage-admin-desktop.png |
| 用户使用日志移动端 | 390×844视口，同一记录显示50.0 tok/s；页面无整体横向溢出 | usage-user-mobile.png |
| 筛选首请求恢复 | localStorage保存openai/error后打开账号页，首个列表请求已带platform=openai/status=error；lite=1、include_scheduler_score=0 | requests.json、account-filters-desktop.png |
| OpenAI账号编辑配置 | 真实表单scope选择为model、选项“仅冷却当前模型/暂停整个账号”；搜索为enabled、选项“默认/开启/关闭” | account-scope-search-desktop.png |

检查语言为中文。截图只证明这些固定数据的实际渲染，其他边界由对应组件/服务/真实SQLite测试验收；不把固定API桩当作后端端到端证据。英文文案与两种运维详情布局由Vitest/i18n检查覆盖。

首次账号页面触发既有新手引导，关闭后继续检查；未修改产品引导逻辑。Vite/vue-tsc监视器最后报告0错误；没有为了截图改动产品配置。
