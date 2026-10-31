# Design

## Context

用户需要Terra，fork已经有OpenAI/Codex兼容、规范化、reasoning、metadata与价格策略。上游新模型落在这些既有路径，不能为它导入不相关平台或整套新目录/计价系统。

## Decisions

1. 使用同一canonical模型身份贯穿规范化、映射、effort、manifest和usage。只开放source明确的六档effort；Terra允许none、默认medium，max不退化为xhigh，显式thinking禁用优先。Sol与Astra既有规则保持。
2. 工具请求需要Responses；对only-Chat账号在上游调用之前返回对应协议400。无工具文本继续遵循既有raw兼容规则，不扩任何工具新endpoint。
3. 默认描述可给Terra六档effort和缺省context/output，但账号明确metadata优先，不给text-only/high-only上游硬补图片或六档。上游列表和raw缓存不被就地改写。
4. 使用冻结Terra资源和两层fallback，既有pricing优先顺序与显式0保留。长上下文沿当前token桶合计及严格大于272000规则；只扩这个模型的现有policy识别，保留fixed quota和WS默认1.0。
5. 保持默认模型不变，在已有白名单/映射预设/客户端模板加入Terra。OpenCode是既有客户端配置，不能借此加入缺失的OpenCode账号平台。
6. 新source测试中依赖的上游目录API在fork结构不同，按本地真实模型测试/manifest消费者重写测试；不以测试CLEAN为导入整个目录系统的理由。
7. 来源created占位值与真实release时间分开，价格/能力值标明冻结来源，实际账号能力不由公共条目推断。

## Verification

新模型不是现存缺陷修复，但每项新增行为仍须先有失败回归。覆盖六档effort及反例、三个only-Chat工具出口、metadata优先/空目录不增项、HTTP/SSE/WS usage、价格资源/fallback/显式0/threshold边界，以及前端生成配置解析。

组合backend full unit/build/lint、相关race、SQLite迁移227/quota/dedup与完整frontend门禁通过后填写verification。baseline输出不作实施证据，实际起点与SHA另记。无新schema/依赖/迁移/全局默认，也无支付、发布或部署步骤。

