## ADDED Requirements

### Requirement: 内存指标必须整组取自同一来源，不得混用容器与宿主机
系统 SHALL 让 `used` / `total` / `percent` 三个内存读数来自同一来源：
仅当 cgroup **同时**给出当前用量与**具体**上限（`memory.max` 不是 `max`）时整组采用 cgroup；
否则整组回落到宿主机指标。MUST NOT 出现容器 `used` 搭配宿主机 `total` 的组合——
那会把读数压成极小的百分比（例如 60 MB / 23 GB ≈ 0.3%），等于内存监控失效。

#### Scenario: 容器设置了内存上限
- **WHEN** cgroup 同时提供当前用量与具体上限
- **THEN** 三个读数 MUST 全部来自 cgroup
- **THEN** 百分比 MUST 等于容器用量除以容器上限

#### Scenario: 容器未设置内存上限
- **WHEN** cgroup 提供当前用量，但上限为「无限制」
- **THEN** 三个读数 MUST 全部来自宿主机
- **THEN** MUST NOT 用容器用量除以宿主机总量

#### Scenario: 完全没有 cgroup 数据
- **WHEN** 读取 cgroup 内存失败
- **THEN** 三个读数 MUST 全部来自宿主机

#### Scenario: 宿主机指标也不可用
- **WHEN** cgroup 数据不完整且宿主机指标读取失败
- **THEN** 三个读数 MUST 全部为「不可用」
- **THEN** MUST NOT 输出部分填充、彼此不自洽的读数

### Requirement: CPU 指标保持既有的逐项回落语义
系统 MUST 保持 CPU 采集的既有行为：优先 cgroup，cgroup 不可用时回落宿主机。
CPU 与内存的来源选择 MUST 相互独立，内存的整组约束 MUST NOT 改变 CPU 读数。

#### Scenario: cgroup CPU 可用、cgroup 内存不完整
- **WHEN** cgroup 提供 CPU 使用率，但内存缺少具体上限
- **THEN** CPU MUST 仍来自 cgroup
- **THEN** 内存 MUST 整组来自宿主机
