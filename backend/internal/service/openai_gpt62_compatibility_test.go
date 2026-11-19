//go:build unit

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGPT62CodexMetadataAndDeclaredTiers(t *testing.T) {
	for _, model := range []struct {
		id                string
		context, priority int
	}{{"gpt-6.2-sol", 500_000, 1}, {"gpt-6.2-astra", 272_000, 2}} {
		t.Run(model.id, func(t *testing.T) {
			descriptor := newConfiguredCodexModelDescriptor(model.id)
			require.EqualValues(t, model.context, descriptor.ContextWindow)
			require.EqualValues(t, 872_000, descriptor.MaxContextWindow)
			require.EqualValues(t, model.priority, descriptor.Priority)
			require.Equal(t, "low", *descriptor.DefaultReasoningLevel)
			require.Equal(t, "xhigh", *descriptor.MultiAgentReasoningEffort)
			require.Equal(t, "v2", descriptor.MultiAgentVersion)
			require.NotEmpty(t, openai.CodexBaseInstructionsForModel(model.id))
			var levels []string
			for _, level := range descriptor.SupportedReasoningLevels {
				levels = append(levels, level.Effort)
			}
			require.Equal(t, []string{"low", "medium", "high", "xhigh", "max", "ultra"}, levels)
			api := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
			caps := accountCodexToolCapabilities(api, model.id)
			require.True(t, gjson.GetBytes(caps["service_tiers"], `#(id=="ultrafast")`).Exists())
			require.JSONEq(t, `false`, string(caps["use_responses_lite"]))
			for _, suffix := range []string{"", "-max", "-openai-compact"} {
				body := []byte(fmt.Sprintf(`{"models":[{"slug":%q,"use_responses_lite":true,"provider_field":"keep"}],"provider_root":7}`, model.id+suffix))
				adjusted, err := adjustAPIKeyCodexModelsManifest(body, api)
				require.NoError(t, err)
				require.False(t, gjson.GetBytes(adjusted, "models.0.use_responses_lite").Bool())
				require.Equal(t, "keep", gjson.GetBytes(adjusted, "models.0.provider_field").String())
				require.EqualValues(t, 7, gjson.GetBytes(adjusted, "provider_root").Int())
			}
			// Saved account capabilities retain their fields and authority.
			for _, kind := range []string{AccountTypeAPIKey, AccountTypeOAuth} {
				account := &Account{Platform: PlatformOpenAI, Type: kind}
				tiers := json.RawMessage(`[{"id":"priority","name":"Fast","provider_field":"keep"}]`)
				account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
					model.id: {ID: model.id, CodexToolCapabilities: map[string]json.RawMessage{"service_tiers": tiers}},
				}})
				require.JSONEq(t, string(tiers), string(accountCodexToolCapabilities(account, model.id)["service_tiers"]))
			}
		})
	}
}

func TestGPT62EffortAndSamplingCompatibility(t *testing.T) {
	for _, model := range []string{"gpt-6.2-sol", "gpt-6.2-astra"} {
		for _, effort := range []string{"low", "medium", "high", "xhigh", "max"} {
			sampling := 0.7
			req := &apicompat.AnthropicRequest{Model: model + "-" + effort, Temperature: &sampling, TopP: &sampling}
			applyOpenAICompatModelNormalization(req)
			converted, err := apicompat.AnthropicToResponses(req)
			require.NoError(t, err)
			require.Equal(t, model, converted.Model)
			require.Equal(t, effort, converted.Reasoning.Effort)
			require.Nil(t, converted.Temperature)
			require.Nil(t, converted.TopP)
		}
		body, changed, err := normalizeGPT6ResponsesSampling([]byte(`{"temperature":0.7,"top_p":0.9,"logprobs":true,"top_logprobs":3,"include":["reasoning.encrypted_content","message.output_text.logprobs"]}`), model)
		require.NoError(t, err)
		require.True(t, changed)
		require.JSONEq(t, `{"include":["reasoning.encrypted_content"]}`, string(body))
		for _, effort := range []string{"none", "minimal"} {
			_, err := apicompat.ChatCompletionsToResponses(&apicompat.ChatCompletionsRequest{Model: model, ReasoningEffort: effort})
			require.ErrorContains(t, err, "does not support reasoning effort")
		}
	}
}

func TestGPT62MappedCompatibilityHTTPRejectsDisabledReasoning(t *testing.T) {
	for _, model := range []string{"gpt-6.2-sol", "gpt-6.2-astra"} {
		for _, protocol := range []string{"chat", "responses", "messages"} {
			for _, effort := range []string{"none", "minimal"} {
				t.Run(model+"/"+protocol+"/"+effort, func(t *testing.T) {
					body := fmt.Sprintf(`{"model":"public","messages":[{"role":"user","content":"hi"}],"input":"hi","max_tokens":100,"reasoning_effort":%q,"reasoning":{"effort":%q},"output_config":{"effort":%q}}`, effort, effort, effort)
					recorder := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(recorder)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+protocol, strings.NewReader(body))
					account := &Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"public": model}}}
					svc := &OpenAIGatewayService{cfg: &config.Config{}}
					var err error
					switch protocol {
					case "chat":
						_, err = svc.forwardAsRawChatCompletions(context.Background(), c, account, []byte(body), "")
					case "responses":
						_, err = svc.forwardResponsesViaRawChatCompletions(context.Background(), c, account, []byte(body), time.Now())
					case "messages":
						_, err = svc.forwardAnthropicViaRawChatCompletions(context.Background(), c, account, []byte(body), "")
					}
					require.ErrorContains(t, err, "does not support reasoning effort")
					require.Equal(t, http.StatusBadRequest, recorder.Code)
				})
			}
		}
	}
}
