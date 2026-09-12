package service

import (
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestSyncUpstreamModelCatalogAstraPartialRefreshPreservesKnownCapabilities(t *testing.T) {
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"data":[
			{"id":"gpt-6-astra","supports_search_tool":false,"apply_patch_tool_type":null},
			{"id":"still-listed"},{"id":"gpt-image-2"}
		]}`))},
		{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"openai":{"id":"openai","models":{
			"gpt-6-astra":{"reasoning":true,"reasoning_options":[{"type":"effort","values":["low","medium","high","xhigh","max"]}],"modalities":{"input":["text","image"]},"limit":{"context":1050000,"output":128000}},
			"mapped-only":{"reasoning":false,"modalities":{"input":["text"]},"limit":{"context":64000}},
			"gpt-image-2":{"reasoning":false,"modalities":{"input":["text","image"]},"limit":{"context":0}}
		}}}`))},
	}}
	repo := &upstreamModelMetadataRepoStub{}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream, cfg: upstreamModelSyncTestConfig()}
	account := &Account{ID: 114, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "test", "base_url": "https://api.openai.com/v1",
			"model_mapping": map[string]any{"public-model": "mapped-only"}},
	}
	old := UpstreamModelMetadata{ID: "still-listed", ContextWindow: 256000}
	account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
		"still-listed": old, "removed": {ID: "removed", ContextWindow: 128000},
		"gpt-6-astra": {ID: "gpt-6-astra", CodexToolCapabilities: map[string]json.RawMessage{
			"supports_search_tool": json.RawMessage("true"), "apply_patch_tool_type": json.RawMessage(`"freeform"`),
			"comp_hash": json.RawMessage(`"3000"`),
		}},
	}})

	catalog, err := svc.SyncUpstreamModelCatalog(context.Background(), account)
	require.NoError(t, err)
	require.Equal(t, UpstreamModelMetadataPartialCode, catalog.Warnings[0].Code)
	require.NotContains(t, catalog.Models, "mapped-only", "capability enrichment must not change discovery")
	snapshot := account.GetUpstreamModelMetadataSnapshot()
	require.Equal(t, old, snapshot.Models["still-listed"])
	require.NotContains(t, snapshot.Models, "removed")
	require.NotContains(t, snapshot.Models, "gpt-image-2")
	require.Contains(t, snapshot.Models, "mapped-only")
	astra := snapshot.Models["gpt-6-astra"]
	require.Equal(t, int64(1050000), astra.ContextWindow)
	require.Equal(t, []string{"low", "medium", "high", "xhigh", "max"}, astra.SupportedReasoningLevels)
	require.JSONEq(t, "false", string(astra.CodexToolCapabilities["supports_search_tool"]))
	require.JSONEq(t, "null", string(astra.CodexToolCapabilities["apply_patch_tool_type"]))
	require.JSONEq(t, `"3000"`, string(astra.CodexToolCapabilities["comp_hash"]))
	require.NotNil(t, repo.updates)
}

func TestMatchModelsDevProviderOfficialHostsWithoutAPI(t *testing.T) {
	registry := map[string]modelsDevProvider{
		"openai": {ID: "openai", Models: map[string]modelsDevModel{"gpt-6-astra": {ID: "gpt-6-astra"}}},
		"relay":  {ID: "relay", API: "https://relay.example/v1"},
	}
	for _, baseURL := range []string{"https://api.openai.com/v1", "https://chatgpt.com/backend-api/codex"} {
		provider, ok := matchModelsDevProvider(registry, baseURL)
		require.True(t, ok)
		require.Equal(t, "openai", provider.ID)
	}
	for _, baseURL := range []string{"https://api.openai.com.evil.example/v1", "https://unknown.example/v1"} {
		_, ok := matchModelsDevProvider(registry, baseURL)
		require.False(t, ok)
	}
	provider, ok := matchModelsDevProvider(registry, "https://relay.example/v1")
	require.True(t, ok)
	require.Equal(t, "relay", provider.ID)
	metadata := map[string]UpstreamModelMetadata{"gpt-6-astra": {
		Reasoning: new(bool), InputModalities: []string{"text"}, ContextWindow: 1050000,
	}}
	require.False(t, upstreamCatalogNeedsRegistry(capabilitySyncModelIDs([]string{"gpt-6-astra", "gpt-image-2"}), metadata))
}

func TestSyncUpstreamModelCatalogIgnoresDedicatedMediaModelsForCompleteness(t *testing.T) {
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{"object":"list","data":[
				{"id":"gpt-6-astra","object":"model"},
				{"id":"gpt-image-2","object":"model"}
			]}`)),
		},
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
				"openai": {
					"id": "openai",
					"models": {
						"gpt-6-astra": {
							"id": "gpt-6-astra",
							"name": "GPT-6 Astra",
							"reasoning": true,
							"reasoning_options": [{"type":"effort","values":["low","medium","high","xhigh","max"]}],
							"modalities": {"input":["text","image"],"output":["text"]},
							"limit": {"context":1050000,"output":128000}
						},
						"gpt-image-2": {
							"id": "gpt-image-2",
							"name": "gpt-image-2",
							"reasoning": false,
							"modalities": {"input":["text","image"],"output":["image"]},
							"limit": {"context":0,"output":0}
						}
					}
				}
			}`)),
		},
	}}
	repo := &upstreamModelMetadataRepoStub{}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream, cfg: upstreamModelSyncTestConfig()}

	catalog, err := svc.SyncUpstreamModelCatalog(context.Background(), &Account{
		ID: 113, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":  "sk-test",
			"base_url": "https://api.openai.com/v1",
			"model_mapping": map[string]any{
				"gpt-6-astra": "gpt-6-astra",
				"gpt-image-2": "gpt-image-2",
			},
		},
	})
	require.NoError(t, err)
	require.Empty(t, catalog.Warnings, "media generators must not keep agent capability sync in a failed state")
	require.NotNil(t, repo.updates)

	encoded, err := json.Marshal(repo.updates[UpstreamModelMetadataExtraKey])
	require.NoError(t, err)
	var snapshot UpstreamModelMetadataSnapshot
	require.NoError(t, json.Unmarshal(encoded, &snapshot))
	require.Contains(t, snapshot.Models, "gpt-6-astra")
	require.Equal(t, []string{"low", "medium", "high", "xhigh", "max"}, snapshot.Models["gpt-6-astra"].SupportedReasoningLevels)
	require.NotContains(t, snapshot.Models, "gpt-image-2")
}

func TestMatchModelsDevProviderFallsBackToOpenAIProviderWithoutAPIField(t *testing.T) {
	t.Parallel()

	registry := map[string]modelsDevProvider{
		"openai": {
			ID:   "openai",
			Name: "OpenAI",
			// Official models.dev entry currently omits `api`.
			Models: map[string]modelsDevModel{
				"gpt-6-astra": {ID: "gpt-6-astra", Name: "GPT-6 Astra"},
			},
		},
		"opencode": {
			ID:  "opencode",
			API: "https://opencode.ai/zen/v1",
			Models: map[string]modelsDevModel{
				"x-preview-f-free": {ID: "x-preview-f-free", Name: "Ox"},
			},
		},
	}

	for _, baseURL := range []string{
		"https://api.openai.com",
		"https://api.openai.com/v1",
		"https://chatgpt.com/backend-api/codex",
	} {
		provider, ok := matchModelsDevProvider(registry, baseURL)
		require.True(t, ok, baseURL)
		require.Equal(t, "openai", provider.ID, baseURL)
		require.Contains(t, provider.Models, "gpt-6-astra", baseURL)
	}

	_, ok := matchModelsDevProvider(registry, "https://compatible.example/v1")
	require.False(t, ok, "custom hosts must not inherit the official OpenAI provider by name")

	provider, ok := matchModelsDevProvider(registry, "https://opencode.ai/zen/v1")
	require.True(t, ok)
	require.Equal(t, "opencode", provider.ID)
}

func TestSyncUpstreamModelCatalogEnrichesOfficialOpenAIHostWithoutRegistryAPIField(t *testing.T) {
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{"object":"list","data":[
				{"id":"gpt-6-astra","object":"model"},
				{"id":"gpt-5.6-sol","object":"model"}
			]}`)),
		},
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
				"openai": {
					"id": "openai",
					"name": "OpenAI",
					"models": {
						"gpt-6-astra": {
							"id": "gpt-6-astra",
							"name": "GPT-6 Astra",
							"description": "OpenAI GPT-6 Astra",
							"reasoning": true,
							"reasoning_options": [{"type":"effort","values":["low","medium","high","xhigh","max"]}],
							"modalities": {"input":["text","image","pdf"],"output":["text"]},
							"limit": {"context":1050000,"output":128000}
						},
						"gpt-5.6-sol": {
							"id": "gpt-5.6-sol",
							"name": "GPT-5.6 Sol",
							"reasoning": true,
							"reasoning_options": [{"type":"effort","values":["low","medium","high","xhigh","max"]}],
							"modalities": {"input":["text","image"],"output":["text"]},
							"limit": {"context":1050000,"output":128000}
						}
					}
				}
			}`)),
		},
	}}
	repo := &upstreamModelMetadataRepoStub{}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream, cfg: upstreamModelSyncTestConfig()}

	catalog, err := svc.SyncUpstreamModelCatalog(context.Background(), &Account{
		ID: 110, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":  "sk-test",
			"base_url": "https://api.openai.com/v1",
		},
	})
	require.NoError(t, err)
	require.Empty(t, catalog.Warnings)
	require.Equal(t, []string{"gpt-5.6-sol", "gpt-6-astra"}, catalog.Models)

	astra := catalog.Metadata["gpt-6-astra"]
	require.Equal(t, "GPT-6 Astra", astra.DisplayName)
	require.NotNil(t, astra.Reasoning)
	require.True(t, *astra.Reasoning)
	require.Equal(t, []string{"low", "medium", "high", "xhigh", "max"}, astra.SupportedReasoningLevels)
	require.Equal(t, []string{"text", "image"}, astra.InputModalities)
	require.Equal(t, int64(1_050_000), astra.ContextWindow)

	require.NotNil(t, repo.updates)
	encoded, err := json.Marshal(repo.updates[UpstreamModelMetadataExtraKey])
	require.NoError(t, err)
	var snapshot UpstreamModelMetadataSnapshot
	require.NoError(t, json.Unmarshal(encoded, &snapshot))
	require.Equal(t, "models.dev", snapshot.Source)
	require.Equal(t, astra, snapshot.Models["gpt-6-astra"])
}

func TestSyncUpstreamModelCatalogPersistsCompleteModelsWhenSomeRemainIncomplete(t *testing.T) {
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{"models":[
				{
					"id":"complete-model",
					"display_name":"Complete",
					"reasoning":true,
					"supported_reasoning_levels":[{"effort":"low"},{"effort":"high"}],
					"input_modalities":["text","image"],
					"context_window":128000
				},
				{
					"id":"incomplete-model",
					"display_name":"Incomplete Only"
				}
			]}`)),
		},
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
				"provider": {
					"id": "provider",
					"api": "https://provider.example/v1",
					"models": {}
				}
			}`)),
		},
	}}
	repo := &upstreamModelMetadataRepoStub{}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream, cfg: upstreamModelSyncTestConfig()}

	catalog, err := svc.SyncUpstreamModelCatalog(context.Background(), &Account{
		ID: 111, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"api_key": "key", "base_url": "https://provider.example/v1"},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"complete-model", "incomplete-model"}, catalog.Models)
	require.Equal(t, "upstream_model_metadata_partial", catalog.Warnings[0].Code)
	require.NotNil(t, repo.updates)

	encoded, err := json.Marshal(repo.updates[UpstreamModelMetadataExtraKey])
	require.NoError(t, err)
	var snapshot UpstreamModelMetadataSnapshot
	require.NoError(t, json.Unmarshal(encoded, &snapshot))
	require.Contains(t, snapshot.Models, "complete-model")
	require.NotContains(t, snapshot.Models, "incomplete-model")
	require.Equal(t, []string{"low", "high"}, snapshot.Models["complete-model"].SupportedReasoningLevels)
}

func TestSyncUpstreamModelCatalogEnrichesConfiguredMappingModelsMissingFromUpstreamList(t *testing.T) {
	upstream := &httpUpstreamRecorder{responses: []*http.Response{
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader(`{"object":"list","data":[{"id":"gpt-5.6-sol","object":"model"}]}`)),
		},
		{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body: io.NopCloser(strings.NewReader(`{
				"openai": {
					"id": "openai",
					"models": {
						"gpt-5.6-sol": {
							"id": "gpt-5.6-sol",
							"name": "GPT-5.6 Sol",
							"reasoning": true,
							"reasoning_options": [{"type":"effort","values":["low","medium","high","xhigh","max"]}],
							"modalities": {"input":["text","image"],"output":["text"]},
							"limit": {"context":1050000,"output":128000}
						},
						"gpt-6-astra": {
							"id": "gpt-6-astra",
							"name": "GPT-6 Astra",
							"reasoning": true,
							"reasoning_options": [{"type":"effort","values":["low","medium","high","xhigh","max"]}],
							"modalities": {"input":["text","image"],"output":["text"]},
							"limit": {"context":1050000,"output":128000}
						}
					}
				}
			}`)),
		},
	}}
	repo := &upstreamModelMetadataRepoStub{}
	svc := &AccountTestService{accountRepo: repo, httpUpstream: upstream, cfg: upstreamModelSyncTestConfig()}

	catalog, err := svc.SyncUpstreamModelCatalog(context.Background(), &Account{
		ID: 112, Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{
			"api_key":  "sk-test",
			"base_url": "https://api.openai.com/v1",
			"model_mapping": map[string]any{
				"gpt-6-astra": "gpt-6-astra",
			},
		},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"gpt-5.6-sol"}, catalog.Models, "UI model list stays upstream-only")
	require.Empty(t, catalog.Warnings)
	require.Contains(t, catalog.Metadata, "gpt-6-astra")
	require.Equal(t, []string{"low", "medium", "high", "xhigh", "max"}, catalog.Metadata["gpt-6-astra"].SupportedReasoningLevels)

	encoded, err := json.Marshal(repo.updates[UpstreamModelMetadataExtraKey])
	require.NoError(t, err)
	var snapshot UpstreamModelMetadataSnapshot
	require.NoError(t, json.Unmarshal(encoded, &snapshot))
	require.Contains(t, snapshot.Models, "gpt-5.6-sol")
	require.Contains(t, snapshot.Models, "gpt-6-astra")
}
