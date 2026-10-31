# gateway-dependency-maintenance Specification

## Purpose

约束已有 HTML 解析路径采用经过最小依赖图测量的安全维护版本，验证强制传递依赖及网络相关回归，保持 Go 工具链、SQLite 和既有消费者边界。

## Requirements

### Requirement: Existing HTML parsing uses the measured safe dependency graph
The system SHALL adopt golang.org/x/net 0.62.0 with the measured minimum required module changes to remove GO-2026-6860's quadratic text coalescing, while preserving the fork's Go version, schema and unrelated dependencies.

#### Scenario: Many small text runs are parsed
- **WHEN** an isolated HTML workload doubles adjacent short text runs separated by ignored tags
- **THEN** allocation growth follows the repaired near-linear behavior rather than the old quadratic pattern

#### Scenario: Module resolution requires other versions
- **WHEN** the minimal x/net upgrade forces transitive module changes
- **THEN** the actual graph delta is recorded and only required maintenance is adopted instead of replacing the whole upstream module file

#### Scenario: Build tools and existing network consumers run
- **WHEN** backend build/unit/lint, relevant network/race checks and generation-tool compatibility are verified
- **THEN** they pass on the selected graph and Go 1.27.3
- **AND** a new audit distinguishes the addressed finding from other unresolved findings
