# 浏览器验收

2026-10-14，agent-browser 独立会话 `sub2api-tier1`，本机 Vite `127.0.0.1:4179`，实际 Vue 组件和项目 CSS，模拟分组数据。未登录或操作业务实例。

- 224px 账号分组列：中文长名高32px、英文无空格长名高64px；文字节点 clientWidth=scrollWidth=188px，clientHeight=scrollHeight，无裁切；页面无水平溢出。
- 点击 +2：完整列出5个分组，长名换行。[展开截图](./groups-expanded.png)。
- 普通 GroupBadge 默认 truncate：96px容器内文字 clientWidth=80px、scrollWidth=395px，text-overflow=ellipsis。[截图](./groups.png)。
- 打开父弹窗后打开子弹窗；第一次 Esc 后只剩父弹窗，body modal-open=true，焦点回到 open-child；第二次 Esc 后弹窗0，body modal-open=false，焦点回到 open-parent。
- 用于挂载组件的临时入口已在验证后删除，Vite 与浏览器会话随后关闭。

组件测试另外覆盖 disabled 选择器关闭与搜索清理、IME Enter/方向键、非 IME 导航、定价显示、开关服务端状态、卸载后迟到请求、时区与设置保存保护。浏览器记录不代表真实上游凭据/在线 provider 测试。
