//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// 本文件钉住 extractOpenAISSETerminalEvent 换成 frame 识别之后的三条不变式。
// 这个函数有 3 个产品调用点（handleSSEToJSON / handlePassthroughSSEToJSON /
// 流式读取器旁路），改它的判定口径影响面比看起来大，所以单独立文件覆盖。

func TestExtractOpenAISSETerminalEvent_RecognizesEventLineWithoutJSONType(t *testing.T) {
	// 兼容上游的典型形态：类型只在 `event:` 行，data 里没有 "type"。
	// 旧实现只读 data 的 "type"，这类帧完全落空（terminalOK=false），
	// 非流式路径于是把它当普通 SSE 按 200 收尾。
	body := "event: error\ndata: {\"error\":{\"message\":\"boom\"}}\n\ndata: [DONE]\n"

	terminalType, payload, ok := extractOpenAISSETerminalEvent(body)
	require.True(t, ok)
	require.Equal(t, "error", terminalType)
	require.JSONEq(t, `{"error":{"message":"boom"}}`, string(payload))
}

func TestExtractOpenAISSETerminalEvent_PayloadTypeWinsOverEventLine(t *testing.T) {
	// data 里的 "type" 优先于 `event:` 行（effectiveOpenAISSEEventType 的口径）：
	// 有些上游 `event:` 行写的是通用名，真正的语义在 payload 里。
	body := "event: message\ndata: {\"type\":\"response.failed\",\"error\":{\"message\":\"nope\"}}\n\n"

	terminalType, _, ok := extractOpenAISSETerminalEvent(body)
	require.True(t, ok)
	require.Equal(t, "response.failed", terminalType)
}

func TestExtractOpenAISSETerminalEvent_LastTerminalFrameWins(t *testing.T) {
	// 取最后一个匹配帧：一段体里可能先出 response.failed 再补一个 error 帧，
	// 最后那个才是上游真正的收尾表态。旧实现取第一个。
	body := "event: response.failed\ndata: {\"type\":\"response.failed\",\"error\":{\"message\":\"first\"}}\n\n" +
		"event: error\ndata: {\"type\":\"error\",\"error\":{\"message\":\"last\"}}\n\ndata: [DONE]\n"

	terminalType, payload, ok := extractOpenAISSETerminalEvent(body)
	require.True(t, ok)
	require.Equal(t, "error", terminalType)
	require.Contains(t, string(payload), "last")
	require.NotContains(t, string(payload), "first")
}

func TestExtractOpenAISSETerminalEvent_NonTerminalFramesAndDoneOnly(t *testing.T) {
	// 只有增量事件与 [DONE]：不构成终止失败事件，必须报 false，
	// 否则两条非流式路径会把正常流误判成失败。
	body := "event: response.output_text.delta\ndata: {\"type\":\"response.output_text.delta\",\"delta\":\"hi\"}\n\ndata: [DONE]\n"

	terminalType, payload, ok := extractOpenAISSETerminalEvent(body)
	require.False(t, ok)
	require.Empty(t, terminalType)
	require.Nil(t, payload)
}

func TestExtractOpenAISSETerminalEvent_CompletedStaysTerminal(t *testing.T) {
	// 成功收尾也是终止事件（调用方据 terminalType 再分派），不能因为加了 "error"
	// 就把原有的成功分支挤掉。
	body := "event: response.completed\ndata: {\"type\":\"response.completed\",\"response\":{\"id\":\"resp_1\"}}\n\ndata: [DONE]\n"

	terminalType, payload, ok := extractOpenAISSETerminalEvent(body)
	require.True(t, ok)
	require.Equal(t, "response.completed", terminalType)
	require.Contains(t, string(payload), "resp_1")
}
