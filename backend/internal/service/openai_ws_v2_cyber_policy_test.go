//go:build unit

package service

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// 本文件覆盖 markOpenAIWSV2PassthroughCyberPolicy（上游 f4e3eb1c5 的产品改动）。
//
// 上游那条提交自带的两个用例本仓库落不了：handler 侧的
// openai_ws_v2_passthrough_cyber_test.go 依赖未移植的 cyber session 改造
// （CyberSessionExplicitBlockKey / TranscriptBlockKeys 那套，上游把
// openai_cyber_session_block.go 从 99 行改到 163 行、换了整套 key 方案）；
// service 侧的 lifecycle 用例依赖另一处未移植文件里的 openAIStream403AccountRepo helper。
// 按 PORTING-0.1.184.md 第 12 条不留「用例在、基座不在」的状态，改为就地覆盖这个函数本身。

func newCyberPolicyTestContext(t *testing.T) *gin.Context {
	t.Helper()
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(nil)
	c.Request, _ = http.NewRequest(http.MethodPost, "/", nil)
	return c
}

func TestMarkOpenAIWSV2PassthroughCyberPolicy_MarksAndReportsHit(t *testing.T) {
	c := newCyberPolicyTestContext(t)
	// 与 detectOpenAICyberPolicy 认的形态一致的风控终止事件。
	payload := []byte(`{"type":"response.failed","response":{"status":"failed","error":{"code":"cyber_policy","message":"blocked by high-risk cyber policy"},"usage":{"input_tokens":7,"output_tokens":0}}}`)

	require.True(t, markOpenAIWSV2PassthroughCyberPolicy(c, payload))

	mark := GetOpsCyberPolicy(c)
	require.NotNil(t, mark, "命中后必须写下 ops 标记")
	require.Equal(t, http.StatusOK, mark.UpstreamStatus, "流式风控走 HTTP 200")
	require.Contains(t, mark.Message, "cyber")
}

func TestMarkOpenAIWSV2PassthroughCyberPolicy_IgnoresOrdinaryFailure(t *testing.T) {
	c := newCyberPolicyTestContext(t)
	// 普通上游失败不能被当成风控——否则会跳过账号副作用（限流/摘号）。
	payload := []byte(`{"type":"response.failed","response":{"status":"failed","error":{"code":"server_error","message":"upstream exploded"}}}`)

	require.False(t, markOpenAIWSV2PassthroughCyberPolicy(c, payload))
	require.Nil(t, GetOpsCyberPolicy(c))
}
