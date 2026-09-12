package service

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAstraDefaultCatalogDescribesUltraAndContext(t *testing.T) {
	body, err := BuildCodexModelsManifest([]string{"gpt-6", "gpt-6-astra"})
	require.NoError(t, err)
	models := decodeCodexManifestModels(t, body)
	require.Len(t, models, 2)
	for _, model := range models {
		require.Equal(t, float64(1050000), model["context_window"])
		require.Equal(t, []string{"low", "medium", "high", "xhigh", "max", "ultra"}, effortsFromManifestModel(t, model))
		require.Equal(t, "xhigh", model["multi_agent_reasoning_effort"])
		require.Equal(t, "v2", model["multi_agent_version"])
		messages := model["model_messages"].(map[string]any)
		require.Contains(t, messages["instructions_template"], "You are Codex, an agent based on GPT-6.")
	}
}

func TestAstraUltraCatalogPreservesWorkflowMetadata(t *testing.T) {
	// Ultra is a Codex workflow. Its inference effort must survive catalog sync,
	// aliases and group generation instead of silently falling back to max.
	_, metadata, err := extractUpstreamModelCatalog([]byte(`{"models":[{
		"slug":"gpt-6-astra","supported_reasoning_levels":[{"effort":"high"},{"effort":"ultra"}],
		"multi_agent_reasoning_effort":"high","multi_agent_version":"v2"
	}]}`), false)
	require.NoError(t, err)
	account := Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{
		"base_url": "https://relay.example/v1", "model_mapping": map[string]any{"public-astra": "gpt-6-astra"},
	}}
	account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: metadata})
	body, err := buildCodexModelsManifestForAccounts(PlatformOpenAI, []string{"public-astra"}, []Account{account}, nil, true)
	require.NoError(t, err)
	model := decodeCodexManifestModels(t, body)[0]
	require.Equal(t, "high", model["multi_agent_reasoning_effort"])
	require.Equal(t, "v2", model["multi_agent_version"])
	require.Equal(t, []string{"high", "ultra"}, effortsFromManifestModel(t, model))

	peer := Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: account.Credentials}
	body, err = buildCodexModelsManifestForAccounts(PlatformOpenAI, []string{"public-astra"}, []Account{account, peer}, nil, true)
	require.NoError(t, err)
	model = decodeCodexManifestModels(t, body)[0]
	require.Nil(t, model["multi_agent_reasoning_effort"], "do not advertise one account's override for all peers")
	require.Nil(t, model["multi_agent_version"])
}

func TestAstraUltraCatalogPreservesExplicitWorkflowOverrides(t *testing.T) {
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"base_url": "https://relay.example/v1"}}
	for _, fields := range []string{
		`"multi_agent_reasoning_effort":"high","multi_agent_version":"v1"`,
		`"multi_agent_reasoning_effort":null,"multi_agent_version":null`,
	} {
		for _, source := range []string{
			`{"models":[{"slug":"gpt-6-astra",` + fields + `}]}`,
			`{"data":[{"id":"gpt-6-astra",` + fields + `}]}`,
		} {
			converted := convertOpenAIModelListToCodexManifestForAccount([]byte(source), account)
			body, err := completeAPIKeyCodexModelsManifestMetadata(converted, true, account)
			require.NoError(t, err)
			model := decodeCodexManifestModels(t, body)[0]
			var expected map[string]any
			require.NoError(t, json.Unmarshal([]byte(`{`+fields+`}`), &expected))
			for key, value := range expected {
				require.Contains(t, model, key)
				require.Equal(t, value, model[key])
			}
		}
	}
}

func TestAstraCodexToolCapabilitiesUseAccountScopeAndSharedDeclarations(t *testing.T) {
	newAccount := func(baseURL string) Account {
		return Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{
			"base_url": baseURL, "model_mapping": map[string]any{"public-astra": "gpt-6-astra"},
		}}
	}
	official := newAccount("https://api.openai.com/v1")
	custom := newAccount("https://relay.example/v1")
	bridge := newAccount("https://bridge.example/v1")
	bridge.Extra = map[string]any{"openai_responses_supported": false}
	for _, tt := range []struct {
		name     string
		accounts []Account
		search   bool
	}{
		{"official fallback", []Account{official}, true},
		{"custom host has no guessed capability", []Account{custom}, false},
		{"missing peer capability", []Account{official, custom}, false},
		{"implemented chat bridge", []Account{bridge}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			body, err := buildCodexModelsManifestForAccounts(PlatformOpenAI, []string{"public-astra"}, tt.accounts, nil, true)
			require.NoError(t, err)
			model := decodeCodexManifestModels(t, body)[0]
			require.Equal(t, "public-astra", model["slug"])
			require.Equal(t, tt.search, model["supports_search_tool"])
			require.Equal(t, false, model["use_responses_lite"])
			if tt.name == "official fallback" {
				require.Equal(t, "freeform", model["apply_patch_tool_type"])
				require.Equal(t, "3000", model["comp_hash"])
			}
		})
	}
	custom.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
		"gpt-6-astra": {CodexToolCapabilities: map[string]json.RawMessage{
			"supports_search_tool": json.RawMessage("true"), "apply_patch_tool_type": json.RawMessage("null"),
			"comp_hash": json.RawMessage(`"3000"`), "tool_mode": json.RawMessage("null"), "use_responses_lite": json.RawMessage("false"),
		}},
	}})
	body, err := buildCodexModelsManifestForAccounts(PlatformOpenAI, []string{"public-astra"}, []Account{official, custom}, nil, true)
	require.NoError(t, err)
	model := decodeCodexManifestModels(t, body)[0]
	require.Equal(t, true, model["supports_search_tool"])
	require.Nil(t, model["apply_patch_tool_type"], "conflicting tool protocols must not be advertised")
	require.Equal(t, "3000", model["comp_hash"])
}

func TestAstraCodexToolCapabilitiesPreserveLiveNullAndFalse(t *testing.T) {
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey,
		Credentials: map[string]any{"base_url": "https://api.openai.com/v1"}}
	account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
		"gpt-6-astra": {CodexToolCapabilities: map[string]json.RawMessage{
			"supports_search_tool": json.RawMessage("true"), "apply_patch_tool_type": json.RawMessage(`"freeform"`),
			"comp_hash": json.RawMessage(`"3000"`), "tool_mode": json.RawMessage(`"code_mode_only"`),
		}},
	}})
	for _, source := range []string{
		`{"models":[{"slug":"gpt-6-astra","supports_search_tool":false,"apply_patch_tool_type":null,"tool_mode":null}]}`,
		`{"data":[{"id":"gpt-6-astra","supports_search_tool":false,"apply_patch_tool_type":null,"tool_mode":null}]}`,
	} {
		converted := convertOpenAIModelListToCodexManifestForAccount([]byte(source), account)
		body, err := applySyncedAPIKeyCodexModelMetadata(converted, account, source[2:6] == "data")
		require.NoError(t, err)
		body, err = completeAPIKeyCodexModelsManifestMetadata(body, true, account)
		require.NoError(t, err)
		model := decodeCodexManifestModels(t, body)[0]
		require.Equal(t, false, model["supports_search_tool"])
		require.Nil(t, model["apply_patch_tool_type"])
		require.Nil(t, model["tool_mode"])
		require.Equal(t, "3000", model["comp_hash"])
	}
	fields := make(map[string]json.RawMessage)
	applyCodexToolCapabilities(fields, map[string]json.RawMessage{
		"supports_search_tool": json.RawMessage(`"true"`), "comp_hash": json.RawMessage(`{"unexpected":true}`),
		"model_messages": json.RawMessage(`"do not persist prompts"`),
	}, true)
	require.Empty(t, fields)
}

func TestAstraCodexToolCapabilitiesFollowAPIKeyAlias(t *testing.T) {
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{
		"base_url": "https://relay.example/v1", "model_mapping": map[string]any{"my-astra": "gpt-6-astra"},
	}}
	account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
		"gpt-6-astra": {CodexToolCapabilities: map[string]json.RawMessage{
			"supports_search_tool": json.RawMessage("true"), "apply_patch_tool_type": json.RawMessage(`"freeform"`),
			"use_responses_lite": json.RawMessage("false"),
		}},
	}})
	for _, source := range []string{`{"data":[{"id":"my-astra"}]}`, `{"models":[{"slug":"my-astra"}]}`} {
		converted := convertOpenAIModelListToCodexManifestForAccount([]byte(source), account)
		body, err := completeAPIKeyCodexModelsManifestMetadata(converted, true, account)
		require.NoError(t, err)
		models := decodeCodexManifestModels(t, body)
		require.Len(t, models, 1)
		model := models[0]
		require.Equal(t, "my-astra", model["slug"])
		require.Equal(t, "my-astra", model["display_name"])
		require.Equal(t, true, model["supports_search_tool"])
		require.Equal(t, "freeform", model["apply_patch_tool_type"])
		require.Equal(t, false, model["use_responses_lite"])
	}
}

func TestAstraCodexToolCapabilitiesKeepAPIKeyResponsesLiteGuard(t *testing.T) {
	account := Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{
		"base_url": "https://relay.example/v1", "model_mapping": map[string]any{"my-astra": "gpt-6-astra"},
	}}
	account.SetUpstreamModelMetadataSnapshot(UpstreamModelMetadataSnapshot{Models: map[string]UpstreamModelMetadata{
		"gpt-6-astra": {CodexToolCapabilities: map[string]json.RawMessage{"use_responses_lite": json.RawMessage("true")}},
	}})
	body, err := buildCodexModelsManifestForAccounts(PlatformOpenAI, []string{"my-astra"}, []Account{account}, nil, true)
	require.NoError(t, err)
	require.Equal(t, false, decodeCodexManifestModels(t, body)[0]["use_responses_lite"])
	body, err = adjustAPIKeyCodexModelsManifest([]byte(`{"models":[{"slug":"my-astra","use_responses_lite":true}]}`), &account)
	require.NoError(t, err)
	require.Equal(t, false, decodeCodexManifestModels(t, body)[0]["use_responses_lite"])
}
