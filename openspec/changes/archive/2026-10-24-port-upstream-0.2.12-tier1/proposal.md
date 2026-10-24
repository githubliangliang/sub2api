# Why

本地Grok CLI身份落后、Antigravity错误出口可暴露账号池身份，自定义错误码说明与现有行为不符。修复这些既有路径可改善正确性。

## What Changes

- F01 Grok CLI身份匹配官方交互版本。
- F02 Antigravity客户端错误隐藏账号池身份。
- F03 自定义错误码说明符合执行语义。

## Impact

只采用已选行为及必要基座。无需schema或依赖升级；保持SQLite、可关闭Redis、simple mode、使用记录去重和自有版本。具体patch sites只放PORTING，基线测试不是新行为验收。
