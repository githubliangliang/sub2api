## ADDED Requirements

### Requirement: Chat Completions 的 file content part 必须转换为 Responses input_file
`/v1/chat/completions` → Responses 转换层 SHALL 处理 `type: "file"` 的 content part，
把它转换成 Responses API 的 `input_file` 部件，并透传 `filename` / `file_data` / `file_id`。
该转换层 MUST NOT 静默丢弃任何携带可用载荷的 content part。

依据：丢弃是静默的——请求返回 200、模型正常作答，只是 prompt 里没有那份文件，
调用方唯一能观察到的异常是 `prompt_tokens` 偏低。

#### Scenario: 带 file_data 的 PDF 附件
- **WHEN** content part 是 `{"type":"file","file":{"filename":"document.pdf","file_data":"data:application/pdf;base64,..."}}`
- **THEN** 转换结果 MUST 含一个 `type: "input_file"` 的部件
- **THEN** 该部件的 `filename` 与 `file_data` MUST 与输入一致
- **THEN** 该部件的 `file_id` MUST 为空

#### Scenario: 带 file_id 的引用
- **WHEN** content part 是 `{"type":"file","file":{"file_id":"file-abc123"}}`
- **THEN** 转换结果 MUST 含一个 `type: "input_file"` 且 `file_id` 为 `file-abc123` 的部件
- **THEN** 该部件的 `file_data` MUST 为空

#### Scenario: 文本与文件混排
- **WHEN** 同一条消息里既有 `text` 又有 `file` 部件
- **THEN** 两者 MUST 都出现在转换结果里，且 MUST 保持原有顺序

### Requirement: 无载荷的 file part 必须被跳过
既没有 `file_data` 又没有 `file_id` 的 `file` content part SHALL 被跳过，
与既有的「空 image URL 被跳过」保持一致。理由是空的 `input_file` 部件会让上游直接 400，
把一个「附件没带上」的客户端错误升级成整个请求失败。

#### Scenario: 只有 filename 的空 file part
- **WHEN** content part 是 `{"type":"file","file":{"filename":"empty.pdf"}}`
- **THEN** 转换结果 MUST NOT 包含对应的 `input_file` 部件
- **THEN** 同一条消息里的其它部件 MUST 不受影响

#### Scenario: 跳过后消息内容为空
- **WHEN** 跳过空 file part 之后该消息没有任何可用部件
- **THEN** 转换结果的 `content` MUST NOT 是 JSON `null`（沿用既有的空字符串回退）
