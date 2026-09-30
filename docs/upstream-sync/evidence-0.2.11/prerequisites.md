# 跨轮次前置（冻结）

本地基线 `f13dee245790eb2be058935d0b1aa6265da99474`，目标v0.2.11。

- S02：旧0.2.9 F04 / [#7538](https://github.com/Wei-Shaw/sub2api/pull/7538)，merge SHA `b5298fdde8e9c5ba4503aca9bac34c1f87774273`，7文件/+90/-30（旧评估基线均CLEAN）。本地anthropicReasoningEffort缺失，目标helper10行；先完成旧两桥契约并复用，不能宣称第一档全合。
- S03：`695ebede70e0bed4c8fd4c87b5a426448a08ea4c`，原5文件+369/-10。只采用types.go的14行通用定义；mergeAnthropicUsage及两个normalize helper按v0.2.11终态适配本地桥。旧CN原生平台、其它透传消费者不引入。这是旧第四档通用子集明确提升第二档。
- #7730：Astra Ultrafast子提交`e6d191a83f0b37df3cb5b183e9cfd42e47b47a1d`仍第三档。其service-tier能力/价格路径不进入S02；保留本地priority构造。
