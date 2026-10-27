# 9.14 真实浏览器验收

2026-10-27 使用 agent-browser / Chromium 运行真实 Vite/Vue 页面。独立会话为 `p020-tail-20261027-ac5eb5d836c2`；Vite 与固定数据 API 分别使用 `127.0.0.1:34916`、`:34917`。桌面视口为 1280×900，移动视口为 390×844，语言为中文。

页面、组件、Pinia、i18n 和 CSS 均来自当前产品源码。API、用户和 iframe 内容使用[本地夹具](./fixture-api.py)，不将此检查记为业务实例的后端端到端验收。

| 必须覆盖的入口 | 实际检查与结果 | 截图 |
|---|---|---|
| `/legal/:documentId` | 匿名打开 `/legal/qa-policy`；标题、段落、列表、表格、行内代码和图片可读，页面无横向溢出。 | [桌面](./legal-desktop.png)、[移动](./legal-mobile.png) |
| `/custom/:id` | `/custom/qa-guide` 的 Markdown、3 项目录及内嵌 iframe 正常；移动端 iframe 宽 308px，不超过内容宽 356px。`/custom/qa-embed` 的 URL 嵌入模式也正常。 | [桌面](./custom-desktop.png)、[移动](./custom-mobile.png)、[URL 嵌入](./custom-embed-mode.png) |
| 公告弹窗 / 铃铛 | 弹窗富文本可读；标记已读后进入列表、打开第二条详情；两次 Escape 依次关闭详情和列表，恢复 body 滚动。表格、行内代码、图片与链接保留，移动端正文采用组件原有内部滚动。 | [弹窗桌面](./announcement-popup-desktop.png)、[弹窗移动](./announcement-popup-mobile.png)、[列表](./announcement-list-desktop.png)、[详情桌面](./announcement-detail-desktop.png)、[详情移动](./announcement-detail-mobile.png) |
| 侧栏自定义 SVG | 两个图标实际尺寸均为 20×20，保留圆形、路径及青绿色填充；展开、72px 收起侧栏和移动抽屉均可见。`onclick` 等事件属性已移除。 | [展开](./custom-desktop.png)、[收起](./sidebar-collapsed.png)、[移动抽屉](./sidebar-mobile.png) |

当前 fork 页面壳没有挂载公告铃铛/弹窗，因此这两项使用[临时 HTML](./announcement-harness.html)和[临时入口](./announcement-harness.ts)挂载真实组件。该入口使用原组件的真实网络消费者和已读状态逻辑；验收后已删除 frontend 下的两个临时文件。此检查不增加产品菜单或公告入口。

在同一真实 Chromium 中，另加载经过 npm 归档完整性校验的 DOMPurify 3.3.1，并与产品使用的 3.4.17 比较。法律文档、公告弹窗、公告铃铛、自定义 Markdown 和 SVG 的五种调用输出逐字节一致；自定义页 iframe 保留，事件处理属性和 `javascript:` 链接被移除，测试标记未执行。[旧版来源与校验](./baseline-source.json)明确标注该模块仅位于临时目录，没有加入项目依赖树。

全部 12 张最终截图均已用 `view_image` 人工目视复核，配合真实 DOM 和交互断言验收。[视觉记录](./visual-review.json)保存尺寸和 SHA-256，[命令记录](./commands.json)保存实际操作及输出，[API 记录](./requests.json)只记录请求路径及是否存在鉴权头。

初始夹具中的脚本结束标记触发了既有 Vite 配置注入的 HTML 解析错误；有效验收改用事件属性和危险协议输入。第一次等待公告超时确认了页面壳未挂载组件，随后改用上述真实组件入口。首次收起侧栏截图处于动画中，最终截图等待宽度稳定为 72px 后重拍。这些试验记录保留，不计作有效验收。

浏览器会话、Vite 和 API 服务已关闭，两个端口已确认释放，临时 frontend 入口已删除。见[清理记录](./cleanup.json)。
