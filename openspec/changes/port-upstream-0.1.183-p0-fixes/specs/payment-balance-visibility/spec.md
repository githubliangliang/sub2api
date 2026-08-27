## ADDED Requirements

### Requirement: 充值完成后用户余额必须立即可见
系统 SHALL 在支付结果页确认订单已履约后主动刷新当前用户余额，MUST NOT 依赖用户手动刷新页面
才能看到到账金额。刷新失败 MUST NOT 阻塞或改写支付结果页对订单状态的展示。

#### Scenario: 支付成功并已履约
- **WHEN** 用户从支付渠道返回，订单状态为已履约
- **THEN** 页面 MUST 触发一次余额刷新
- **THEN** 展示的余额 MUST 包含本次充值金额

#### Scenario: 订单尚未履约
- **WHEN** 支付结果页拿到的订单仍处于待处理状态
- **THEN** MUST NOT 展示已到账的余额
- **THEN** 履约确认后 MUST 再触发余额刷新

#### Scenario: 余额刷新失败
- **WHEN** 余额刷新请求失败
- **THEN** 支付结果页 MUST 仍正确展示订单状态
- **THEN** MUST NOT 抛出未捕获错误或使页面进入空白状态
