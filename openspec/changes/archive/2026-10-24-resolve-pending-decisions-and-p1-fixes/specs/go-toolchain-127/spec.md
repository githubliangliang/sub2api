## ADDED Requirements

### Requirement: Go 版本声明与全部断言点必须一致
系统 SHALL 让 `backend/go.mod` 的 Go 版本与 CI 中**每一处**硬断言、
全部容器构建基础镜像、以及文档中的版本字样保持一致。
遗漏任一处会让 CI 在版本校验步骤直接失败，或让镜像与本地构建用不同的工具链。

#### Scenario: 升级 Go 版本
- **WHEN** `backend/go.mod` 的 Go 版本被改为新版本
- **THEN** CI 中所有 `go version | grep -q` 形式的断言 MUST 全部同步（当前共 5 处，
  分布在 3 个 workflow 中）
- **THEN** 全部 3 个 Dockerfile 的 golang 基础镜像 MUST 同步
- **THEN** README / README_CN / README_JA / DEV_GUIDE / CLAUDE.md 中的版本字样 MUST 同步

#### Scenario: 文档对断言数量的描述
- **WHEN** 开发文档描述需要同步的断言位置
- **THEN** 描述 MUST 与实际断言行数一致
- **THEN** MUST NOT 沿用过时的计数

#### Scenario: 只改了 go.mod
- **WHEN** 只升级 `go.mod` 而未同步断言
- **THEN** CI MUST 在版本校验步骤失败
- **THEN** 这是预期的保护行为，不是可绕过的噪音

### Requirement: lint 工具链版本必须与 Go 版本兼容
系统 SHALL 让 golangci-lint 的固定版本能够构建并分析目标 Go 版本的模块。
旧版 golangci-lint 会拒绝目标版本更高的 `go.mod`，因此升级 Go MUST 同时升级 lint 版本。
新版本带出的规则 MUST 逐条处理，MUST NOT 通过整体降级规则集或关闭 linter 来消除。

#### Scenario: 升级 Go 但不升 lint
- **WHEN** Go 升级后 lint 版本未变
- **THEN** lint 步骤 MUST 失败
- **THEN** MUST 通过升级 lint 版本解决

#### Scenario: 新规则报出的问题
- **WHEN** 新版 lint 报出污点分析、废弃 API 或误报类问题
- **THEN** MUST 按具体问题修改代码，或对确认的误报加定点抑制注释
- **THEN** MUST NOT 从启用列表中移除该 linter

### Requirement: 生成代码必须由生成器产出，不得手改
新 Go 版本的默认 JSON 引擎会改变 ORM 生成代码中 raw JSON 字段的类型表示。
系统 MUST 通过重新运行生成器产出这些文件并提交结果，MUST NOT 手工编辑生成物。

#### Scenario: 升级后重新生成
- **WHEN** Go 版本升级带来生成代码的类型变化
- **THEN** MUST 运行 ORM 代码生成并提交生成结果
- **THEN** 生成后再次运行生成器 MUST 不产生新的差异

#### Scenario: 手改生成物
- **WHEN** 有人直接编辑生成的文件
- **THEN** 下一次运行生成器 MUST 覆盖这些修改
- **THEN** 该做法 MUST 被视为不满足本要求

### Requirement: HTTP/2 keepalive 断言必须匹配新的配置路径
新 Go 版本下 HTTP/2 的配置方式发生变化：传输层通过协议注册开启 HTTP/2 而不再写旧的协议映射字段，
keepalive 相关超时映射到新的 HTTP/2 配置结构。
系统 SHALL 更新相应断言，使其校验实际生效的配置路径。

#### Scenario: keepalive 配置生效性
- **WHEN** 为上游传输配置读空闲超时与 ping 超时
- **THEN** 断言 MUST 校验新的 HTTP/2 配置结构上的对应字段
- **THEN** MUST NOT 继续断言旧的协议映射字段非空

#### Scenario: keepalive 行为本身
- **WHEN** 上游连接长时间空闲
- **THEN** keepalive 探测 MUST 仍按配置的超时触发
- **THEN** 行为 MUST 与升级前一致
