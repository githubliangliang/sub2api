package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/model"
	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestPassthroughErrorScalarBoundaries(t *testing.T) {
	for _, field := range []string{"type", "code", "param"} {
		for _, tc := range []struct{ name, raw, want string }{
			{"string", `"input"`, "input"},
			{"number", "1234", "1234"},
			{"null", "null", ""},
			{"bool", "true", ""},
			{"object", `{"secret":"private"}`, ""},
			{"array", `["private"]`, ""},
			{"boundary", `"` + strings.Repeat("a", 256) + `"`, strings.Repeat("a", 256)},
			{"oversized", `"` + strings.Repeat("a", 257) + `"`, ""},
			{"utf8 boundary", `"` + strings.Repeat("界", 85) + `"`, strings.Repeat("界", 85)},
			{"utf8 oversized", `"` + strings.Repeat("界", 86) + `"`, ""},
		} {
			t.Run(field+"/"+tc.name, func(t *testing.T) {
				rec, c := newOpenAIPassthroughErrorTestContext()
				body := []byte(`{"error":{"message":"rejected",` + fmt.Sprintf("%q", field) + `:` + tc.raw + `}}`)
				writeOpenAIPassthroughStructuredErrorEnvelope(c, 400, nil, body, nil, "upstream_error", "rejected", "")
				value := gjson.Get(rec.Body.String(), "error."+field)
				if field == "type" && tc.want == "" {
					require.Equal(t, "upstream_error", value.String())
				} else {
					require.Equal(t, tc.want, value.String())
				}
				require.NotContains(t, rec.Body.String(), "private")
			})
		}
	}
	secret := strings.Repeat("s", 300)
	require.Empty(t, openAIPassthroughClientErrorField(&Account{Credentials: map[string]any{"api_key": secret}}, gjson.Parse(`"`+secret+`"`)), "raw oversized identifiers cannot become accepted by redaction")
	for _, raw := range []string{`true`, `false`, `{"token":"private"}`, `["private"]`} {
		for _, outer := range []string{"", `"code":null,`, `"code":" ",`} {
			inner, _ := json.Marshal(`{"error":{"code":` + raw + `,"message":"inner"}}`)
			body := []byte(`{"error":{` + outer + `"message":` + string(inner) + `}}`)
			require.Empty(t, openAIPassthroughClientErrorCode(nil, body), string(body))
		}
	}
	for _, outer := range []string{"", `"code":null,`, `"code":" ",`} {
		rec, c := newOpenAIPassthroughErrorTestContext()
		inner, _ := json.Marshal(`{"error":{"code":"invalid_encrypted_content","message":"Encrypted content could not be verified."}}`)
		forwardOpenAIPassthroughErrorForTest(t, c, 400, `{"error":{`+outer+`"message":`+string(inner)+`}}`, map[string]any{"api_key": "synthetic-api-key"})
		require.Equal(t, "invalid_encrypted_content", gjson.Get(rec.Body.String(), "error.code").String())
	}
	for _, raw := range []string{`true`, `{"code":"invalid_encrypted_content"}`, `["invalid_encrypted_content"]`} {
		rec, c := newOpenAIPassthroughErrorTestContext()
		inner, _ := json.Marshal(`{"error":{"code":"invalid_encrypted_content","message":"private opaque error"}}`)
		forwardOpenAIPassthroughErrorForTest(t, c, 400, `{"error":{"code":`+raw+`,"message":`+string(inner)+`}}`, map[string]any{"api_key": "synthetic-api-key"})
		require.Equal(t, "Upstream request failed", gjson.Get(rec.Body.String(), "error.message").String())
		require.False(t, gjson.Get(rec.Body.String(), "error.code").Exists())
	}
}

func TestPassthroughErrorRulesUseOriginalBodyPlatformAndPriority(t *testing.T) {
	for _, platform := range []string{PlatformOpenAI, PlatformGrok} {
		t.Run(platform, func(t *testing.T) {
			rec, c := newOpenAIPassthroughErrorTestContext()
			rules := &ErrorPassthroughService{}
			status := 422
			wrongStatus := 409
			custom := "selected administrator text"
			rules.setLocalCache([]*model.ErrorPassthroughRule{
				{ID: 1, Enabled: true, Priority: 0, Platforms: []string{PlatformAnthropic}, ErrorCodes: []int{400}, MatchMode: model.MatchModeAny, ResponseCode: &wrongStatus, CustomMessage: &custom},
				{ID: 2, Enabled: true, Priority: 1, Platforms: []string{platform}, ErrorCodes: []int{400}, Keywords: []string{"body_only_code"}, MatchMode: model.MatchModeAll, ResponseCode: &status, CustomMessage: &custom, SkipMonitoring: true},
				{ID: 3, Enabled: true, Priority: 2, Platforms: []string{platform}, ErrorCodes: []int{400}, MatchMode: model.MatchModeAny, ResponseCode: &wrongStatus, CustomMessage: &custom},
			})
			BindErrorPassthroughService(c, rules)
			account := &Account{ID: 1, Platform: platform, Type: AccountTypeAPIKey}
			svc := &OpenAIGatewayService{cfg: &config.Config{}}
			body := []byte(`{"error":{"code":"body_only_code","message":"rejected"}}`)
			err := svc.handleErrorResponsePassthrough(context.Background(), &http.Response{StatusCode: 400, Header: http.Header{}}, c, account, nil, body)
			require.Error(t, err)
			require.Equal(t, status, rec.Code)
			require.Equal(t, custom, gjson.Get(rec.Body.String(), "error.message").String())
			require.False(t, gjson.Get(rec.Body.String(), "error.code").Exists())
			skip, _ := c.Get(OpsSkipPassthroughKey)
			require.Equal(t, true, skip)
			original, _ := c.Get(OpsUpstreamStatusCodeKey)
			require.Equal(t, 400, original)
			require.True(t, IsResponseCommitted(c))
		})
	}
}

type passthroughErrorLocalHTTP struct{ client *http.Client }

func (p passthroughErrorLocalHTTP) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	return p.client.Do(req)
}
func (p passthroughErrorLocalHTTP) DoWithTLS(req *http.Request, proxy string, id int64, n int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	return p.Do(req, proxy, id, n)
}

func TestPassthroughErrorActualHTTPAndCompactLifecycle(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, heartbeat := range []bool{false, true} {
		for _, mappedStatus := range []int{400, 422, 200} {
			for _, structured := range []bool{false, true} {
				t.Run(fmt.Sprintf("heartbeat=%t/status=%d/structured=%t", heartbeat, mappedStatus, structured), func(t *testing.T) {
					upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if r.URL.Path != "/v1/responses/compact" {
							t.Errorf("unexpected upstream path: %s", r.URL.Path)
						}
						w.Header().Set("Content-Type", "application/json")
						w.Header().Set("Server", "synthetic-private")
						w.Header().Set("Set-Cookie", "synthetic-private")
						w.WriteHeader(400)
						_, _ = io.WriteString(w, `{"error":{"type":"invalid_request_error","code":"invalid_encrypted_content","param":"input","message":"Encrypted content could not be verified."},"debug":"synthetic-private"}`)
					}))
					defer upstream.Close()
					result := make(chan error, 1)
					router := gin.New()
					router.POST("/v1/responses/compact", func(c *gin.Context) {
						MarkOpenAICompactClientStream(c)
						stop := StartOpenAICompactSSEKeepalive(c, time.Hour)
						defer stop()
						if heartbeat {
							v, _ := c.Get(openAICompactSSEKeepaliveKey)
							keepalive, ok := v.(*openAICompactSSEKeepalive)
							if !ok || !keepalive.beat() {
								result <- fmt.Errorf("heartbeat failed")
								return
							}
						}
						ObserveOpenAIResponsesSequence(c, 9)
						custom := "explicit administrator message"
						rules := &ErrorPassthroughService{}
						rules.setLocalCache([]*model.ErrorPassthroughRule{{ID: 1, Enabled: true, Priority: 1, Platforms: []string{PlatformOpenAI}, Keywords: []string{"invalid_encrypted_content"}, ErrorCodes: []int{400}, MatchMode: model.MatchModeAll, ResponseCode: &mappedStatus, PassthroughBody: structured, CustomMessage: &custom, SkipMonitoring: true}})
						BindErrorPassthroughService(c, rules)
						cfg := &config.Config{Security: config.SecurityConfig{URLAllowlist: config.URLAllowlistConfig{AllowInsecureHTTP: true}}}
						svc := &OpenAIGatewayService{cfg: cfg, httpUpstream: passthroughErrorLocalHTTP{upstream.Client()}}
						account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Concurrency: 1, Status: StatusActive, Schedulable: true, Extra: map[string]any{"openai_passthrough": true}, Credentials: map[string]any{"api_key": "synthetic-api-key", "base_url": upstream.URL}}
						forwardResult, err := svc.Forward(c.Request.Context(), c, account, []byte(`{"model":"gpt-6-astra","input":"hi"}`))
						if err == nil || forwardResult != nil {
							result <- fmt.Errorf("failed request produced success usage: %#v %v", forwardResult, err)
							return
						}
						if skip, _ := c.Get(OpsSkipPassthroughKey); skip != true {
							result <- fmt.Errorf("original body rule did not suppress monitoring")
							return
						}
						if heartbeat {
							// A second failure attempt must neither advance the sequence nor write bytes.
							n := c.Writer.Size()
							next := responsesSequenceFromContext(c).peek()
							writeOpenAICompactSSEFailureWithCode(c, 400, "other", "duplicate")
							if c.Writer.Size() != n || responsesSequenceFromContext(c).peek() != next {
								result <- fmt.Errorf("duplicate compact failure wrote bytes or consumed sequence")
								return
							}
							streamErr, ok := GetOpsStreamError(c)
							if !ok || streamErr.ErrType != "upstream_error" {
								result <- fmt.Errorf("missing internal stream error")
								return
							}
						}
						result <- nil
					})
					downstream := httptest.NewServer(router)
					defer downstream.Close()
					response, err := downstream.Client().Post(downstream.URL+"/v1/responses/compact", "application/json", strings.NewReader(`{}`))
					require.NoError(t, err)
					body, err := io.ReadAll(response.Body)
					require.NoError(t, err)
					require.NoError(t, response.Body.Close())
					require.NoError(t, <-result)
					require.Empty(t, response.Header.Get("Server"))
					require.Empty(t, response.Header.Get("Set-Cookie"))
					require.NotContains(t, string(body), "synthetic-private")
					if heartbeat {
						require.Equal(t, 200, response.StatusCode)
						require.Contains(t, response.Header.Get("Content-Type"), "text/event-stream")
						events := parseCompactBridgeSSE(t, stripKeepaliveComments(string(body)))
						require.Len(t, events, 1)
						require.Equal(t, "response.failed", events[0][0])
						require.EqualValues(t, 10, gjson.Get(events[0][1], "sequence_number").Int())
						code := "upstream_error"
						message := "explicit administrator message"
						if structured {
							code = "invalid_encrypted_content"
							message = "Encrypted content could not be verified."
						}
						require.Equal(t, code, gjson.Get(events[0][1], "response.error.code").String())
						require.Equal(t, message, gjson.Get(events[0][1], "response.error.message").String())
						require.NotContains(t, string(body), "response.completed")
						require.NotContains(t, string(body), "\"usage\"")
					} else {
						require.Equal(t, mappedStatus, response.StatusCode)
						require.Contains(t, response.Header.Get("Content-Type"), "application/json")
						require.NotContains(t, string(body), "event:")
					}
				})
			}
		}
	}
}

func TestPassthroughCompactCanceledBeforeHeartbeatKeepsJSONError(t *testing.T) {
	rec, c := newOpenAIPassthroughErrorTestContext()
	ctx, cancel := context.WithCancel(c.Request.Context())
	c.Request = c.Request.WithContext(ctx)
	MarkOpenAICompactClientStream(c)
	stop := StartOpenAICompactSSEKeepalive(c, time.Hour)
	cancel()
	stop()
	stop()
	forwardOpenAIPassthroughErrorForTest(t, c, 400, `{"error":{"code":"invalid_encrypted_content","message":"Encrypted content could not be verified."}}`, map[string]any{"api_key": "synthetic-api-key"})
	require.Equal(t, 400, rec.Code)
	require.NotContains(t, rec.Body.String(), "event:")
	require.NotContains(t, rec.Body.String(), "keepalive")
	require.Equal(t, "invalid_encrypted_content", gjson.Get(rec.Body.String(), "error.code").String())
}

func TestPassthroughErrorRawRulePriorityDoesNotInheritMessageOnlySuppression(t *testing.T) {
	rec, c := newOpenAIPassthroughErrorTestContext()
	rules := &ErrorPassthroughService{}
	rules.setLocalCache([]*model.ErrorPassthroughRule{
		{ID: 1, Enabled: true, Priority: 1, Platforms: []string{PlatformOpenAI}, Keywords: []string{"raw_only_code"}, MatchMode: model.MatchModeAny, PassthroughCode: true, PassthroughBody: true},
		{ID: 2, Enabled: true, Priority: 2, Platforms: []string{PlatformOpenAI}, Keywords: []string{"shared rejected message"}, MatchMode: model.MatchModeAny, PassthroughCode: true, PassthroughBody: true, SkipMonitoring: true},
	})
	BindErrorPassthroughService(c, rules)
	account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	svc := &OpenAIGatewayService{cfg: &config.Config{}}
	err := svc.handleErrorResponsePassthrough(context.Background(), &http.Response{StatusCode: 400, Header: http.Header{}}, c, account, nil, []byte(`{"error":{"code":"raw_only_code","message":"shared rejected message"}}`))
	require.Error(t, err)
	require.Equal(t, "raw_only_code", gjson.Get(rec.Body.String(), "error.code").String())
	skip, _ := c.Get(OpsSkipPassthroughKey)
	require.NotEqual(t, true, skip, "higher-priority original-body rule must decide monitoring")
}
