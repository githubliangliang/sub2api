//go:build unit

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

	"github.com/Wei-Shaw/sub2api/internal/pkg/apicompat"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestGPT61TerraIdentityAndEffort(t *testing.T) {
	t.Run("public defaults", func(t *testing.T) {
		require.Contains(t, openai.DefaultModelIDs(), "gpt-6.1-terra")
		require.Equal(t, "gpt-5.6-sol", openai.DefaultModelIDs()[0])
		require.Equal(t, "gpt-5.4", openai.DefaultTestModel)
	})
	for _, effort := range []string{"none", "low", "medium", "high", "xhigh", "max"} {
		t.Run(effort, func(t *testing.T) {
			model := "tenant/openai/GPT_6.1_Terra_" + strings.ToUpper(effort)
			require.Equal(t, "gpt-6.1-terra", normalizeCodexModel(model))
			require.Equal(t, "gpt-6.1-terra", normalizeKnownOpenAICodexModel(model))
			require.Equal(t, "gpt-6.1-terra", NormalizeOpenAICompatRequestedModel(model))
			req := &apicompat.AnthropicRequest{Model: model}
			applyOpenAICompatModelNormalization(req)
			require.Equal(t, "gpt-6.1-terra", req.Model)
			require.NotNil(t, req.OutputConfig)
			require.Equal(t, effort, req.OutputConfig.Effort)
		})
		t.Run("conversion-"+effort, func(t *testing.T) {
			require.Equal(t, effort, normalizeOpenAIReasoningEffortForModel(effort, "gpt-6.1-terra"))
			req := &apicompat.AnthropicRequest{Model: "gpt-6.1-terra", OutputConfig: &apicompat.AnthropicOutputConfig{Effort: effort}}
			converted, err := apicompat.AnthropicToResponses(req)
			require.NoError(t, err)
			require.Equal(t, effort, converted.Reasoning.Effort)
			chat, err := apicompat.ChatCompletionsToResponses(&apicompat.ChatCompletionsRequest{
				Model: "gpt-6.1-terra", ReasoningEffort: effort,
				Tools: []apicompat.ChatTool{{Type: "function", Function: &apicompat.ChatFunction{Name: "lookup"}}},
			})
			require.NoError(t, err)
			require.Equal(t, "gpt-6.1-terra", chat.Model)
			require.Equal(t, effort, chat.Reasoning.Effort)
			require.Len(t, chat.Tools, 1)
			require.Equal(t, "lookup", chat.Tools[0].Name)
		})
	}
	for _, model := range []string{"gpt-6.1-terra", "gpt-6.1-terra-openai-compact", "openai/gpt_6.1_terra", " OpenAI/GPT 6.1 TERRA HIGH "} {
		t.Run(model, func(t *testing.T) {
			require.Equal(t, "gpt-6.1-terra", normalizeKnownOpenAICodexModel(model))
			require.Equal(t, "gpt-6.1-terra", NormalizeOpenAICompatRequestedModel(model))
		})
	}
	for _, suffix := range []string{"pro", "ultra", "minimal", "preview", "compact"} {
		t.Run("unsupported-"+suffix, func(t *testing.T) {
			model := "gpt-6.1-terra-" + suffix
			require.Empty(t, normalizeKnownOpenAICodexModel(model))
			require.Equal(t, model, normalizeCodexModel(model))
			require.Equal(t, model, NormalizeOpenAICompatRequestedModel(model))
		})
	}
}

func TestGPT61TerraReasoningOverridesAndSolRejection(t *testing.T) {
	for _, body := range []string{
		`{"model":"gpt-6.1-terra-max","thinking":{"type":"disabled"},"output_config":{"effort":"max"}}`,
		`{"model":"gpt-6.1-terra-high","output_config":{"effort":"none"}}`,
		`{"model":"gpt-6.1-terra-low","output_config":{"effort":"high"}}`,
	} {
		t.Run(body, func(t *testing.T) {
			var req apicompat.AnthropicRequest
			require.NoError(t, json.Unmarshal([]byte(body), &req))
			applyOpenAICompatModelNormalization(&req)
			require.Equal(t, "gpt-6.1-terra", req.Model)
			converted, err := apicompat.AnthropicToResponses(&req)
			require.NoError(t, err)
			want := "none"
			if gjson.Get(body, "output_config.effort").String() == "high" {
				want = "high"
			}
			require.Equal(t, want, converted.Reasoning.Effort)
			require.Equal(t, want, openAICompatAnthropicReasoningEffort(&req, req.Model, converted.Reasoning.Effort))
		})
	}
	defaultReq, err := apicompat.AnthropicToResponses(&apicompat.AnthropicRequest{Model: "gpt-6.1-terra"})
	require.NoError(t, err)
	require.Equal(t, "medium", defaultReq.Reasoning.Effort)
	for _, effort := range []string{"none", "minimal"} {
		require.Error(t, validateGPT61SolCompatRequest([]byte(`{"reasoning":{"effort":"`+effort+`"}}`), "gpt-6.1-sol"))
		require.NoError(t, validateGPT61SolCompatRequest([]byte(`{"reasoning":{"effort":"none"}}`), "gpt-6.1-terra"))
	}
}

func TestGPT61TerraConfiguredManifestDefaultsAndMetadata(t *testing.T) {
	t.Run("default descriptor", func(t *testing.T) {
		descriptor := newConfiguredCodexModelDescriptor("gpt-6.1-terra")
		require.Equal(t, "medium", *descriptor.DefaultReasoningLevel)
		var efforts []string
		for _, level := range descriptor.SupportedReasoningLevels {
			efforts = append(efforts, level.Effort)
		}
		require.Equal(t, []string{"none", "low", "medium", "high", "xhigh", "max"}, efforts)
		require.EqualValues(t, 1_050_000, descriptor.ContextWindow)
		require.EqualValues(t, 1_050_000, descriptor.MaxContextWindow)
		require.Empty(t, descriptor.officialMetadata)
		require.Nil(t, descriptor.MultiAgentReasoningEffort)
	})
	for _, baseURL := range []string{"https://api.openai.com/v1", "https://models.example/v1"} {
		t.Run(baseURL, func(t *testing.T) {
			account := *newCodexModelsAPIKeyTestAccount(baseURL)
			account.Credentials["model_mapping"] = map[string]any{"public": "gpt-6.1-terra"}
			reasoning := true
			account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
				"gpt-6.1-terra": {ID: "gpt-6.1-terra", Reasoning: &reasoning, DefaultReasoningLevel: "high", SupportedReasoningLevels: []string{"high"}, InputModalities: []string{"text"}, ContextWindow: 100_000},
			}})
			service := &GatewayService{accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{1: {account}}}}
			body, err := service.BuildCodexModelsManifestForGroup(context.Background(), &Group{ID: 1, Platform: PlatformOpenAI}, "", []string{"public"})
			require.NoError(t, err)
			models := decodeCodexManifestModels(t, body)
			require.Len(t, models, 1)
			require.Equal(t, "public", models[0]["slug"])
			require.Equal(t, []any{"text"}, models[0]["input_modalities"])
			require.Equal(t, []string{"high"}, effortsFromManifestModel(t, models[0]))
			require.EqualValues(t, 100_000, models[0]["context_window"])
			manifest := &CodexModelsManifest{Body: []byte(`{"models":[{"slug":"gpt-6.1-terra"}]}`)}
			require.NoError(t, (&OpenAIGatewayService{}).CompleteAPIKeyCodexModelsManifestForClient(manifest, &account))
			models = decodeCodexManifestModels(t, manifest.Body)
			require.Equal(t, "high", models[0]["default_reasoning_level"])
			require.Equal(t, []string{"high"}, effortsFromManifestModel(t, models[0]))
			require.Equal(t, []any{"text"}, models[0]["input_modalities"])
			require.EqualValues(t, 100_000, models[0]["context_window"])
		})
	}
}

func TestGPT61TerraDiscoveryPreservesRawCatalogAndCache(t *testing.T) {
	for _, catalog := range []string{
		`{"models":[]}`,
		`{"models":[{"slug":"gpt-6.1-terra","input_modalities":["text"],"default_reasoning_level":"high","supported_reasoning_levels":[{"effort":"high"}],"context_window":100000,"vendor_extension":{"kept":true}}],"vendor_catalog":{"kept":true}}`,
	} {
		for _, baseURL := range []string{"https://api.openai.com/v1", "https://models.example/v1"} {
			t.Run(baseURL+catalog, func(t *testing.T) {
				calls := 0
				gateway := newCodexModelsAPIKeyTestService(&codexModelsHTTPUpstreamStub{do: func(_ *http.Request, _ string, _ int64, _ int) (*http.Response, error) {
					calls++
					return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Etag": []string{`"terra-upstream"`}}, Body: io.NopCloser(strings.NewReader(catalog))}, nil
				}})
				account := newCodexModelsAPIKeyTestAccount(baseURL)
				before, err := gateway.FetchCodexModelsManifest(context.Background(), account, "", "")
				require.NoError(t, err)
				rawBody := append([]byte(nil), before.Body...)
				rawSource := append([]byte(nil), before.upstreamSourceBody...)
				rawETag := before.ETag
				require.NoError(t, gateway.CompleteAPIKeyCodexModelsManifestForClient(before, account))
				models := decodeCodexManifestModels(t, before.Body)
				if catalog == `{"models":[]}` {
					require.Empty(t, models)
				} else {
					require.Len(t, models, 1)
					require.Equal(t, []any{"text"}, models[0]["input_modalities"])
					require.Equal(t, []string{"high"}, effortsFromManifestModel(t, models[0]))
					require.EqualValues(t, 100_000, models[0]["context_window"])
					require.Equal(t, map[string]any{"kept": true}, models[0]["vendor_extension"])
				}
				after, err := gateway.FetchCodexModelsManifest(context.Background(), account, "", "")
				require.NoError(t, err)
				require.Equal(t, rawBody, after.Body)
				require.Equal(t, rawSource, after.upstreamSourceBody)
				require.Equal(t, []byte(catalog), after.upstreamSourceBody)
				require.Equal(t, rawETag, after.ETag)
				require.Equal(t, 1, calls)
				notModified, err := gateway.FetchCodexModelsManifest(context.Background(), account, "", after.ETag)
				require.NoError(t, err)
				require.True(t, notModified.NotModified)
				require.Equal(t, 1, calls)
			})
		}
	}
}

func TestGPT61TerraOnlyChatRejectsToolsBeforeForwarding(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, protocol := range []string{"chat-tools", "chat-functions", "responses", "messages"} {
		for _, effort := range []string{"", "none", "max"} {
			for _, stream := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/stream=%t", protocol, effort, stream), func(t *testing.T) {
					body := fmt.Sprintf(`{"model":"public","messages":[{"role":"user","content":"hi"}],"reasoning_effort":%q,"stream":%t,"tools":[{"type":"function","function":{"name":"lookup","parameters":{"type":"object"}}}]}`, effort, stream)
					if protocol == "chat-functions" {
						body = fmt.Sprintf(`{"model":"public","messages":[{"role":"user","content":"hi"}],"reasoning_effort":%q,"stream":%t,"functions":[{"name":"lookup","parameters":{"type":"object"}}]}`, effort, stream)
					} else if protocol == "responses" {
						body = fmt.Sprintf(`{"model":"public","input":"hi","reasoning":{"effort":%q},"stream":%t,"tools":[{"type":"function","name":"lookup","parameters":{"type":"object"}}]}`, effort, stream)
					} else if protocol == "messages" {
						body = fmt.Sprintf(`{"model":"public","max_tokens":128,"messages":[{"role":"user","content":"hi"}],"output_config":{"effort":%q},"stream":%t,"tools":[{"name":"lookup","input_schema":{"type":"object"}}]}`, effort, stream)
					}
					recorder := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(recorder)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+protocol, strings.NewReader(body))
					account := forceChatResponsesFallbackAccount()
					account.Credentials["model_mapping"] = map[string]any{"public": "openai/GPT_6.1_Terra_max"}
					upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(`{"id":"chat_terra","model":"gpt-6.1-terra","choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":9,"completion_tokens":4}}`))}}
					svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
					var err error
					switch protocol {
					case "responses":
						_, err = svc.Forward(context.Background(), c, account, []byte(body))
					case "messages":
						_, err = svc.ForwardAsAnthropic(context.Background(), c, account, []byte(body), "", "")
					default:
						_, err = svc.ForwardAsChatCompletions(context.Background(), c, account, []byte(body), "", "")
					}
					require.ErrorContains(t, err, "gpt-6.1-terra requires Responses")
					require.Equal(t, http.StatusBadRequest, recorder.Code)
					require.Nil(t, upstream.lastReq)
					require.Equal(t, "invalid_request_error", gjson.Get(recorder.Body.String(), "error.type").String())
					if protocol == "messages" {
						require.Equal(t, "error", gjson.Get(recorder.Body.String(), "type").String())
					}
				})
			}
		}
	}
}

func TestGPT61TerraOnlyChatTextPreservesRawCompatibility(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, protocol := range []string{"chat", "responses", "messages"} {
		for _, stream := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/stream=%t", protocol, stream), func(t *testing.T) {
				body := fmt.Sprintf(`{"model":"public","messages":[{"role":"user","content":"hi"}],"input":"hi","max_tokens":128,"reasoning_effort":"none","reasoning":{"effort":"none"},"output_config":{"effort":"none"},"stream":%t,"vendor_extension":{"kept":true}}`, stream)
				response := `{"id":"chat_terra","model":"gpt-6.1-terra","choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1000,"completion_tokens":10,"prompt_tokens_details":{"cached_tokens":200,"cache_write_tokens":300}}}`
				contentType := "application/json"
				if stream {
					contentType = "text/event-stream"
					response = "data: " + response + "\n\ndata: [DONE]\n\n"
				}
				upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{contentType}}, Body: io.NopCloser(strings.NewReader(response))}}
				svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
				account := forceChatResponsesFallbackAccount()
				account.Credentials["model_mapping"] = map[string]any{"public": "gpt-6.1-terra"}
				recorder := httptest.NewRecorder()
				c, _ := gin.CreateTestContext(recorder)
				c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+protocol, strings.NewReader(body))
				var result *OpenAIForwardResult
				var err error
				switch protocol {
				case "responses":
					result, err = svc.Forward(context.Background(), c, account, []byte(body))
				case "messages":
					result, err = svc.ForwardAsAnthropic(context.Background(), c, account, []byte(body), "", "")
				default:
					result, err = svc.ForwardAsChatCompletions(context.Background(), c, account, []byte(body), "", "")
				}
				require.NoError(t, err)
				require.NotNil(t, result)
				require.Equal(t, "/v1/chat/completions", upstream.lastReq.URL.Path)
				require.Equal(t, "gpt-6.1-terra", gjson.GetBytes(upstream.lastBody, "model").String())
				require.NotNil(t, result.ReasoningEffort)
				require.Equal(t, "none", *result.ReasoningEffort)
				require.Equal(t, 1000, result.Usage.InputTokens)
				require.Equal(t, 10, result.Usage.OutputTokens)
				require.Equal(t, 200, result.Usage.CacheReadInputTokens)
				require.Equal(t, 300, result.Usage.CacheCreationInputTokens)
				if protocol == "chat" {
					require.True(t, gjson.GetBytes(upstream.lastBody, "vendor_extension.kept").Bool())
					require.Contains(t, recorder.Body.String(), `"choices"`)
				}
			})
		}
	}
}

func TestGPT61TerraResponsesToolsPreserveEffortAndUsage(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, protocol := range []string{"chat", "legacy-functions", "messages", "responses"} {
		for _, effort := range []string{"none", "low", "medium", "high", "xhigh", "max"} {
			for _, stream := range []bool{false, true} {
				t.Run(fmt.Sprintf("%s/%s/stream=%t", protocol, effort, stream), func(t *testing.T) {
					body := fmt.Sprintf(`{"model":"public","messages":[{"role":"user","content":"hi"}],"reasoning_effort":%q,"stream":%t,"tools":[{"type":"function","function":{"name":"lookup","parameters":{"type":"object"}}}]}`, effort, stream)
					if protocol == "legacy-functions" {
						body = fmt.Sprintf(`{"model":"public","messages":[{"role":"user","content":"hi"}],"reasoning_effort":%q,"stream":%t,"functions":[{"name":"lookup","parameters":{"type":"object"}}]}`, effort, stream)
					} else if protocol == "messages" {
						body = fmt.Sprintf(`{"model":"public","max_tokens":128,"messages":[{"role":"user","content":"hi"}],"output_config":{"effort":%q},"stream":%t,"tools":[{"name":"lookup","input_schema":{"type":"object"}}]}`, effort, stream)
					} else if protocol == "responses" {
						body = fmt.Sprintf(`{"model":"public","input":"hi","reasoning":{"effort":%q},"stream":%t,"tools":[{"type":"function","name":"lookup","parameters":{"type":"object"}}]}`, effort, stream)
					}
					const response = `{"id":"resp_terra","model":"gpt-6.1-terra","status":"completed","output":[{"id":"fc_terra","type":"function_call","call_id":"call_lookup","name":"lookup","arguments":"{}"}],"usage":{"input_tokens":1000,"output_tokens":10,"input_tokens_details":{"cached_tokens":200,"cache_write_tokens":300}}}`
					responseBody := response
					contentType := "application/json"
					if protocol != "responses" || stream {
						contentType = "text/event-stream"
						responseBody = "data: " + `{"type":"response.output_item.added","output_index":0,"item":{"id":"fc_terra","type":"function_call","call_id":"call_lookup","name":"lookup","arguments":""}}` + "\n\ndata: " + `{"type":"response.function_call_arguments.delta","item_id":"fc_terra","output_index":0,"delta":"{}"}` + "\n\ndata: " + `{"type":"response.output_item.done","output_index":0,"item":{"id":"fc_terra","type":"function_call","call_id":"call_lookup","name":"lookup","arguments":"{}"}}` + "\n\ndata: {\"type\":\"response.completed\",\"response\":" + response + "}\n\ndata: [DONE]\n\n"
					}
					upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{contentType}}, Body: io.NopCloser(strings.NewReader(responseBody))}}
					svc := &OpenAIGatewayService{cfg: rawChatCompletionsTestConfig(), httpUpstream: upstream}
					account := rawChatCompletionsTestAccount()
					account.Credentials["model_mapping"] = map[string]any{"public": "gpt-6.1-terra"}
					recorder := httptest.NewRecorder()
					c, _ := gin.CreateTestContext(recorder)
					c.Request = httptest.NewRequest(http.MethodPost, "/v1/"+protocol, strings.NewReader(body))
					var result *OpenAIForwardResult
					var err error
					switch protocol {
					case "messages":
						result, err = svc.ForwardAsAnthropic(context.Background(), c, account, []byte(body), "", "")
					case "responses":
						result, err = svc.Forward(context.Background(), c, account, []byte(body))
					default:
						result, err = svc.ForwardAsChatCompletions(context.Background(), c, account, []byte(body), "", "")
					}
					require.NoError(t, err)
					require.NotNil(t, result)
					require.Equal(t, "/v1/responses", upstream.lastReq.URL.Path)
					require.Equal(t, "gpt-6.1-terra", gjson.GetBytes(upstream.lastBody, "model").String())
					require.Equal(t, effort, gjson.GetBytes(upstream.lastBody, "reasoning.effort").String())
					require.Equal(t, "lookup", gjson.GetBytes(upstream.lastBody, "tools.0.name").String())
					require.NotNil(t, result.ReasoningEffort)
					require.Equal(t, effort, *result.ReasoningEffort)
					require.Equal(t, 1000, result.Usage.InputTokens)
					require.Equal(t, 200, result.Usage.CacheReadInputTokens)
					require.Equal(t, 300, result.Usage.CacheCreationInputTokens)
					require.Equal(t, 10, result.Usage.OutputTokens)
					require.Equal(t, "gpt-6.1-terra", result.UpstreamResponseModel)
					require.Contains(t, recorder.Body.String(), "lookup")
				})
			}
		}
	}
}

func TestGPT61TerraUsageHTTPAndWSKeepAllExistingBuckets(t *testing.T) {
	response := []byte(`{"model":"gpt-6.1-terra","usage":{"input_tokens":1000,"output_tokens":10,"input_tokens_details":{"cached_tokens":200,"cache_write_tokens":300,"image_tokens":50},"output_tokens_details":{"image_tokens":2}}}`)
	httpUsage, ok := extractOpenAIUsageFromJSONBytes(response)
	require.True(t, ok)
	var wsUsage OpenAIUsage
	parseOpenAIWSResponseUsageFromCompletedEvent(append(append([]byte(`{"type":"response.completed","response":`), response...), '}'), &wsUsage)
	require.Equal(t, httpUsage, wsUsage)
	require.Equal(t, 1000, httpUsage.InputTokens)
	require.Equal(t, 10, httpUsage.OutputTokens)
	require.Equal(t, 200, httpUsage.CacheReadInputTokens)
	require.Equal(t, 300, httpUsage.CacheCreationInputTokens)
	require.Equal(t, 50, httpUsage.ImageInputTokens)
	require.Equal(t, 2, httpUsage.ImageOutputTokens)
}

func TestGPT61TerraLiveModelIDsKeepEmptyDiscoveryEmpty(t *testing.T) {
	for _, catalog := range []string{`{"data":[]}`, `{"data":[{"id":"gpt-6.1-terra"}]}`} {
		t.Run(catalog, func(t *testing.T) {
			upstream := &httpUpstreamRecorder{resp: &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": []string{"application/json"}}, Body: io.NopCloser(strings.NewReader(catalog))}}
			svc := &AccountTestService{httpUpstream: upstream, cfg: upstreamModelSyncTestConfig()}
			models, err := svc.FetchUpstreamSupportedModels(context.Background(), newCodexModelsAPIKeyTestAccount("https://models.example/v1"))
			if catalog == `{"data":[]}` {
				require.ErrorContains(t, err, "Upstream returned no supported models")
				require.Empty(t, models)
			} else {
				require.NoError(t, err)
				require.Equal(t, []string{"gpt-6.1-terra"}, models)
			}
			require.Equal(t, "/v1/models", upstream.lastReq.URL.Path)
		})
	}
}

func TestGPT61TerraNoneFilteringUsesMappedModelAcrossRepresentations(t *testing.T) {
	for _, mapping := range []struct {
		target string
		keep   bool
	}{
		{target: "openai/GPT_6.1_Terra_max", keep: true},
		{target: "custom-text", keep: false},
	} {
		t.Run(mapping.target, func(t *testing.T) {
			account := rawChatCompletionsTestAccount()
			account.Credentials["model_mapping"] = map[string]any{"public": mapping.target}
			body := []byte(`{"model":"public","reasoning":{"effort":"none"},"reasoning_effort":"none"}`)
			filtered, err := filterOpenAIResponsesNoneReasoningEffortForAccount(account, body)
			require.NoError(t, err)
			var object map[string]any
			require.NoError(t, json.Unmarshal(body, &object))
			deleteOpenAIResponsesNoneReasoningEffortFromObject(account, object)
			filteredObject, err := json.Marshal(object)
			require.NoError(t, err)
			for _, representation := range [][]byte{filtered, filteredObject} {
				require.Equal(t, mapping.keep, gjson.GetBytes(representation, "reasoning.effort").Exists())
				require.Equal(t, mapping.keep, gjson.GetBytes(representation, "reasoning_effort").Exists())
			}
		})
	}
}
