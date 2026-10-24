# Why

本地验证码计数与重置token消费存在并发时序问题；用户另已明确需要账号优先级快捷调整和密钥分组排序。需缓存接口与前后端行为联合验证。

## What Changes

- S01 验证码尝试上限与重置令牌单次消费。
- S02 账号优先级快捷调整。
- S03 API密钥按分组名稳定排序。

## Impact

只采用已选行为及必要基座。无需schema或依赖升级；保持SQLite、可关闭Redis、simple mode、使用记录去重和自有版本。具体patch sites只放PORTING，基线测试不是新行为验收。
