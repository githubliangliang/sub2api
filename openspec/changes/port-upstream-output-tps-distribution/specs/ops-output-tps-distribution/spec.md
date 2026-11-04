## ADDED Requirements

### Requirement: 窗口分布必须基于有效请求的完整耗时速率
系统 SHALL 对保留用量明细中的每条有效记录计算输出 Token ÷ 总耗时，包含首字等待，MUST NOT 重复加入推理 Token 或使用总 Token 吞吐替代。

#### Scenario: 多种请求耗时不同
- **WHEN** 当前窗口包含速率1、10、100、1000 tok/s的有效sync/stream/ws_v2/cyber请求
- **THEN** P5/P10/P50/平均值 SHALL 分别为2.35、3.7、55、277.75，有效样本 SHALL 为4

#### Scenario: 不适用记录
- **WHEN** 明细为图片、Live、unknown、非正输出或无有效总耗时
- **THEN** 该记录 MUST NOT 进入输出 TPS 分布

### Requirement: 所有查询模式必须计算当前筛选的完整窗口
系统 SHALL 在 raw/auto/preagg 下复用[start,end)时间、平台和分组过滤并对全窗口计算连续分位数，MUST NOT 平均小时分位数。

#### Scenario: 包含小时头尾和稳定聚合桶
- **WHEN** 样本跨越窗口头尾及多个完整小时且已有小时聚合
- **THEN** 各模式的输出分布 SHALL 与直接计算完整明细相同

#### Scenario: 平台归属和明细清理
- **WHEN** 账号平台与所属分组平台不同或无分组，或者已清理部分明细
- **THEN** 平台 SHALL 优先分组、无分组回退账号，统计 SHALL 仅包含筛选后的保留明细

### Requirement: 空样本和统计不可用必须保留可识别语义
系统 SHALL 为无样本返回可空统计与零样本数，为自身查询超时返回不可用统计并保留基础概览，MUST 传播调用方取消/截止和非超时数据库错误。

#### Scenario: 无有效样本
- **WHEN** 当前筛选没有有效请求
- **THEN** output_tps SHALL 含四个null与sample_count=0

#### Scenario: 统计超时或调用取消
- **WHEN** 输出速率查询超过自身2秒预算，或者调用方取消/截止
- **THEN** 自身超时 SHALL 保留概览且output_tps=null，调用方取消/截止 SHALL 返回对应错误

### Requirement: 运维页面必须展示并更新窗口输出速率
系统 SHALL 在已有普通和全屏运维页面显示P50、P5、P10、平均值与有效样本数，使用当前筛选与刷新结果，MUST NOT 把缺失统计伪装为零。

#### Scenario: 切换筛选或刷新结果
- **WHEN** 当前概览更新且包含输出分布
- **THEN** 卡片 SHALL 使用最新概览，以一位小数和tok/s显示速率，并展示样本数

#### Scenario: 空样本或缺失字段
- **WHEN** 后端返回空样本或output_tps为null/缺失
- **THEN** 速率 SHALL 显示—，空样本数显示0，不可用样本数显示—
