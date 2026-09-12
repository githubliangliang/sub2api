package handler

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

type channelSchedulingUpstream struct {
	service.HTTPUpstream
	response string
	body     []byte
}

type channelSchedulingAccountRepo struct {
	openAIWSUsageHandlerAccountRepoStub
}

func (r *channelSchedulingAccountRepo) ListModelAvailabilityCandidates(context.Context, *int64, []string, bool) ([]service.Account, error) {
	return []service.Account{r.account}, nil
}

func (u *channelSchedulingUpstream) Do(req *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
	var err error
	u.body, err = io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(u.response))}, nil
}

// Exercise the real handlers and scheduler: the account has only the channel's
// target in its capability mapping, so using the public alias rejects it.
func TestOpenAIHTTP_ChannelMappedTargetSelectsAccountWithoutRequestedAlias(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		path     string
		body     string
		response string
	}{
		{"/v1/responses", `{"model":"public-alias","input":"hello","stream":false}`, `{"id":"resp_mapped","object":"response","model":"gpt-5.6-sol","status":"completed","output":[],"usage":{"input_tokens":2,"output_tokens":1}}`},
		{"/v1/chat/completions", `{"model":"public-alias","messages":[{"role":"user","content":"hello"}],"stream":false}`, `{"id":"chatcmpl_mapped","object":"chat.completion","model":"gpt-5.6-sol","choices":[{"index":0,"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":2,"completion_tokens":1,"total_tokens":3}}`},
		{"/v1/embeddings", `{"model":"public-alias","input":"hello"}`, `{"object":"list","model":"gpt-5.6-sol","data":[{"object":"embedding","index":0,"embedding":[0.1,0.2]}],"usage":{"prompt_tokens":2,"total_tokens":2}}`},
	} {
		t.Run(tc.path, func(t *testing.T) {
			groupID := int64(4201)
			account := service.Account{ID: 9901, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Status: service.StatusActive, Schedulable: true, Concurrency: 1,
				Credentials: map[string]any{"api_key": "sk-test", "base_url": "https://api.example.test", "model_mapping": map[string]any{"gpt-5.6-sol": "gpt-5.6-sol"}},
				Extra:       map[string]any{"openai_responses_mode": "force_responses", "openai_passthrough": false},
			}
			if tc.path == "/v1/chat/completions" {
				account.Extra["openai_responses_mode"] = "force_chat_completions"
			}
			channelSvc := service.NewChannelService(&openAIWSUsageHandlerChannelRepoStub{
				channels:       []service.Channel{{ID: 7701, Status: service.StatusActive, GroupIDs: []int64{groupID}, ModelMapping: map[string]map[string]string{service.PlatformOpenAI: {"public-alias": "gpt-5.6-sol"}}, BillingModelSource: service.BillingModelSourceChannelMapped}},
				groupPlatforms: map[int64]string{groupID: service.PlatformOpenAI},
			}, nil, nil, nil)
			cfg := &config.Config{RunMode: config.RunModeSimple}
			cfg.Default.RateMultiplier = 1
			cfg.Security.URLAllowlist.Enabled = false
			cfg.Gateway.Scheduling.LoadBatchEnabled = true
			billingCache := service.NewBillingCacheService(nil, nil, nil, nil, nil, nil, cfg, nil)
			t.Cleanup(billingCache.Stop)
			upstream := &channelSchedulingUpstream{response: tc.response}
			usageRepo := &openAIWSUsageHandlerUsageLogRepoStub{created: make(chan *service.UsageLog, 1)}
			concurrency := service.NewConcurrencyService(&concurrencyCacheMock{acquireAccountSlotFn: func(context.Context, int64, int, string) (bool, error) { return true, nil }})
			accountRepo := &channelSchedulingAccountRepo{openAIWSUsageHandlerAccountRepoStub{account: account}}
			gateway := service.NewOpenAIGatewayService(accountRepo, usageRepo, nil, nil, nil, nil, nil, cfg, nil, concurrency, service.NewBillingService(cfg, nil), nil, billingCache, upstream, &service.DeferredService{}, nil, nil, nil, channelSvc, nil, nil, nil)
			h := NewOpenAIGatewayHandler(gateway, concurrency, billingCache, service.NewAPIKeyService(nil, nil, nil, nil, nil, nil, cfg), nil, nil, nil, nil, cfg)
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, tc.path, strings.NewReader(tc.body))
			c.Set(string(middleware.ContextKeyAPIKey), &service.APIKey{ID: 1801, GroupID: &groupID, User: &service.User{ID: 1701, Status: service.StatusActive}, Group: &service.Group{ID: groupID, Platform: service.PlatformOpenAI, Status: service.StatusActive}})
			c.Set(string(middleware.ContextKeyUser), middleware.AuthSubject{UserID: 1701})
			switch tc.path {
			case "/v1/responses":
				h.Responses(c)
			case "/v1/chat/completions":
				h.ChatCompletions(c)
			case "/v1/embeddings":
				h.Embeddings(c)
			}
			require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
			require.Equal(t, "gpt-5.6-sol", gjson.GetBytes(upstream.body, "model").String())
			select {
			case usage := <-usageRepo.created:
				require.Equal(t, "public-alias", usage.RequestedModel)
				if tc.path != "/v1/responses" {
					require.NotNil(t, usage.UpstreamModel)
					require.Equal(t, "gpt-5.6-sol", *usage.UpstreamModel)
				}
				require.Positive(t, usage.TotalCost)
				require.NotNil(t, usage.ModelMappingChain)
				require.Equal(t, "public-alias→gpt-5.6-sol", *usage.ModelMappingChain)
			case <-time.After(3 * time.Second):
				t.Fatal("usage not recorded")
			}
		})
	}
}

func TestOpenAIChannelForwardModelForScheduler(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mapping service.ChannelMappingResult
		want    string
	}{
		{"mapped", service.ChannelMappingResult{Mapped: true, MappedModel: "  target  "}, "target"},
		{"unmapped", service.ChannelMappingResult{MappedModel: "ignored"}, "public"},
		{"empty_target", service.ChannelMappingResult{Mapped: true, MappedModel: " \t "}, "public"},
	} {
		t.Run(tc.name, func(t *testing.T) { require.Equal(t, tc.want, openAIChannelForwardModel(tc.mapping, "public")) })
	}
}
