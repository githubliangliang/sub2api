package service

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// 本文件钉住 32ac921f2 在**原生 /v1/messages 路径**上的效果。
//
// 上游那条修复只改了 applyClaudeCodeOAuthMimicryToBody，而该函数仅被桥接路径
// （ForwardAsResponses / ForwardAsChatCompletions）调用；Claude Code → 原生
// /v1/messages 走的是 Forward 里另一处注入点，也是本 fork 的主用法。少了这条用例，
// 把 gateway_forward.go 里那行 claudeOAuthSystemPromptBlocksForModel 删掉不会有任何
// 测试变红——恰好是 README §4 第 13 条「守卫类单行改动空转」的形状。
//
// 走真实 Forward（而不是只测组合函数），断言的是**实际发往上游的 body**。

// expansionProbe 是通用 CLI 展开块的首句，用来在解析出的 block 文本里判定它在不在。
// 不直接用整个 claudeCodeSystemPromptExpansion 常量做比较，是因为该常量本身会随上游
// 提示词微调而变化，而这个用例要钉的是「展开块在/不在」，不是它的逐字内容。
const expansionProbe = "You are an interactive agent that helps users with software engineering tasks."

func newFableOAuthForwardServiceForTest(upstream *anthropicHTTPUpstreamRecorder) *GatewayService {
	// 复用 gateway_forward_partial_usage_test.go 的装配：settingService 为 nil 时
	// claudeOAuthSystemPromptInjectionSettings 返回 (true, "", "")，即"注入开启 +
	// 用内置默认 blocks"，正是需要覆盖的分支。
	return newForwardPartialUsageServiceForTest(upstream)
}

func forwardOnceForFableTest(t *testing.T, model string) []byte {
	t.Helper()
	gin.SetMode(gin.TestMode)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodPost, "/v1/messages", nil)
	// 非 Claude Code 客户端 + OAuth 账号 ⇒ shouldMimicClaudeCode 为真。
	c.Request.Header.Set("User-Agent", "opencode/1.0")

	body := []byte(`{"model":"` + model + `","system":"Project instructions","messages":[{"role":"user","content":[{"type":"text","text":"hello"}]}]}`)
	parsed, err := ParseGatewayRequest(NewRequestBodyRef(body), PlatformAnthropic)
	require.NoError(t, err)

	upstreamSSE := strings.Join([]string{
		`event: message_start`,
		`data: {"type":"message_start","message":{"id":"msg_1","type":"message","role":"assistant","model":"` + model + `","content":[],"usage":{"input_tokens":11}}}`,
		"",
		`event: message_delta`,
		`data: {"type":"message_delta","delta":{"stop_reason":"end_turn"},"usage":{"output_tokens":3}}`,
		"",
		`event: message_stop`,
		`data: {"type":"message_stop"}`,
		"",
		"",
	}, "\n")

	upstream := &anthropicHTTPUpstreamRecorder{resp: &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
		Body:       io.NopCloser(strings.NewReader(upstreamSSE)),
	}}

	svc := newFableOAuthForwardServiceForTest(upstream)
	_, _ = svc.Forward(context.Background(), c, newAnthropicOAuthAccountForPartialUsageTest(), parsed)

	require.NotEmpty(t, upstream.lastBody, "上游请求体未被记录，说明 Forward 没走到发送阶段")
	return upstream.lastBody
}

func TestGatewayService_Forward_FableOAuthOmitsRefusedExpansion(t *testing.T) {
	sent := forwardOnceForFableTest(t, "claude-fable-5")

	system := gjson.GetBytes(sent, "system").Array()
	require.Len(t, system, 2, "Fable 只保留 billing header + Claude Code 身份两块")
	require.Contains(t, system[0].Get("text").String(), "x-anthropic-billing-header:")
	require.Equal(t, claudeCodeSystemPrompt, system[1].Get("text").String())
	// 断言必须打在**解析后**的 block 文本上：body 是 JSON，换行被转义成 \n，
	// 直接对 string(sent) 做 Contains(claudeCodeSystemPromptExpansion) 永远为假，
	// 那样即使展开块真的还在，用例也照样"通过"。
	for i, block := range system {
		require.NotContains(t, block.Get("text").String(), expansionProbe,
			"system[%d] 仍带通用 CLI 展开块，Fable 上游会回 stop_reason=refusal + 零 output token", i)
	}
	// 客户端原本的 system 指令仍被迁进消息历史，不能连带丢掉。
	require.Contains(t, gjson.GetBytes(sent, "messages.0.content.0.text").String(), "Project instructions")
}

func TestGatewayService_Forward_NonFableOAuthKeepsExpansion(t *testing.T) {
	sent := forwardOnceForFableTest(t, "claude-sonnet-4-5-20250929")

	system := gjson.GetBytes(sent, "system").Array()
	require.Len(t, system, 3, "非 Fable 模型保持 billing header + 身份 + 展开块三块")
	require.Contains(t, system[2].Get("text").String(), expansionProbe,
		"非 Fable 模型必须保持原有 blocks 形态，Fable 分支不得外溢")
}
