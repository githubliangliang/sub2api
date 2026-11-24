package service

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

// isOpenAIInvalidEncryptedContentError 判断上游 400 错误是否为 Responses
// reasoning/compaction 密文校验失败（invalid_encrypted_content）。
// 该错误属于确定性请求失败，客户端需据此识别密文失效或验签失败。
func isOpenAIInvalidEncryptedContentError(statusCode int, upstreamMsg string, body []byte) bool {
	if statusCode != http.StatusBadRequest {
		return false
	}
	code := strings.ToLower(strings.TrimSpace(openAIPassthroughClientErrorCode(nil, body)))
	if code == "invalid_encrypted_content" {
		return true
	}
	if code == "thinking_signature_invalid" {
		lowerMsg := strings.ToLower(strings.TrimSpace(upstreamMsg))
		return strings.Contains(lowerMsg, "encrypted content") &&
			(strings.Contains(lowerMsg, "could not be verified") || strings.Contains(lowerMsg, "could not be decrypted"))
	}
	return false
}

// openAIPassthroughClientErrorFieldMaxBytes 限制透出的 error.type/code/param
// 单字段长度。三者都是短标识符，超长只可能是上游在夹带其它内容。
const openAIPassthroughClientErrorFieldMaxBytes = 256

// openAIPassthroughClientErrorField 提取允许写给客户端的标量错误字段。
// gjson 对对象/数组的 String() 返回整段原始 JSON，直接透出等于把上游任意
// 扩展字段塞进 type/code/param，因此只接受字符串与数字（部分上游的 code 为
// 数字），并在脱敏后做长度上限校验，超限即丢弃。
func openAIPassthroughClientErrorField(account *Account, value gjson.Result) string {
	if value.Type != gjson.String && value.Type != gjson.Number {
		return ""
	}
	raw := value.String()
	if len(raw) > openAIPassthroughClientErrorFieldMaxBytes {
		return ""
	}
	field := sanitizeUpstreamErrorMessageForAccount(raw, account)
	if len(field) > openAIPassthroughClientErrorFieldMaxBytes {
		return ""
	}
	return field
}

// openAIPassthroughClientErrorCode 返回允许写给客户端的 error.code。
// error.code 缺失、为 null 或为空白字符串时，沿用 extractUpstreamErrorCode 对
// "message 内嵌 JSON"形态的兼容提取（与错误分类看到的 code 保持一致）；
// error.code 是对象、数组或布尔值时不回退，直接视为无 code。
func openAIPassthroughClientErrorCode(account *Account, body []byte) string {
	code := gjson.GetBytes(body, "error.code")
	blank := !code.Exists() || code.Type == gjson.Null ||
		(code.Type == gjson.String && strings.TrimSpace(code.Str) == "")
	if !blank {
		return openAIPassthroughClientErrorField(account, code)
	}
	inner := strings.TrimSpace(gjson.GetBytes(body, "error.message").String())
	if !strings.HasPrefix(inner, "{") {
		return ""
	}
	if code := gjson.Get(inner, "error.code"); code.Exists() {
		return openAIPassthroughClientErrorField(account, code)
	}
	if lastBrace := strings.LastIndex(inner, "}"); lastBrace >= 0 {
		return openAIPassthroughClientErrorField(account, gjson.Get(inner[:lastBrace+1], "error.code"))
	}
	return ""
}

// writeOpenAIPassthroughStructuredErrorEnvelope 在净化后的 OpenAI 错误信封内
// 仅保留标准的 type/code/param/message 四个字段（不透传任意扩展字段与敏感响应头），
// 同时兼容 compact SSE 桥接与响应头净化策略。
func writeOpenAIPassthroughStructuredErrorEnvelope(
	c *gin.Context,
	downstreamStatus int,
	upstreamHeaders http.Header,
	body []byte,
	account *Account,
	fallbackType string,
	upstreamMsg string,
	fallbackMessage string,
) {
	if c == nil {
		return
	}
	errType := openAIPassthroughClientErrorField(account, gjson.GetBytes(body, "error.type"))
	if errType == "" {
		errType = fallbackType
	}
	if errType == "" {
		errType = "upstream_error"
	}
	errorPayload := gin.H{"type": errType}
	code := openAIPassthroughClientErrorCode(account, body)
	if code != "" {
		errorPayload["code"] = code
	}
	if param := openAIPassthroughClientErrorField(account, gjson.GetBytes(body, "error.param")); param != "" {
		errorPayload["param"] = param
	}
	message := sanitizeUpstreamErrorMessageForAccount(upstreamMsg, account)
	if message == "" {
		message = fallbackMessage
	}
	if message == "" {
		message = "Upstream request failed"
	}
	errorPayload["message"] = message

	// compact 心跳已把响应头提交为 200 时，错误只能以 response.failed 终止事件
	// 回传。通用桥接会把 code 固定为 upstream_error，这里直接带上结构化 code，
	// 否则本函数要保留的 error.code 恰好在 compact 流上丢失。
	if openAICompactClientWantsStream(c) && StopOpenAICompactSSEKeepaliveCommitted(c) {
		writeOpenAICompactSSEFailureWithCode(c, downstreamStatus, code, message)
		return
	}

	payload, _ := json.Marshal(gin.H{"error": errorPayload})
	writeOpenAIPassthroughErrorHeaders(c.Writer.Header(), upstreamHeaders)
	c.Data(downstreamStatus, "application/json; charset=utf-8", payload)
}
