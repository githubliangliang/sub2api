package service

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratedCodexIdentityUsesRoutedModel(t *testing.T) {
	for _, tc := range []struct {
		public, target string
		gpt            bool
	}{
		{"grok-4.5", "grok-4.5", false},
		{"deepseek-v4-pro", "deepseek-v4-pro", false},
		{"anthropic/claude-sonnet-5.5", "anthropic/claude-sonnet-5.5", false},
		{"gemini-2.5-pro", "gemini-2.5-pro", false},
		{"gpt-oss-120b", "gpt-oss-120b", false},
		{"vendor/gpt-oss-120b", "vendor/gpt-oss-120b", false},
		{"company-codex-alias", "company-codex-alias", false},
		{"company-model", "company-model", false},
		{"gpt-5.5", "gpt-5.5", true},
		{"vendor/gpt-5.2", "vendor/gpt-5.2", true},
		{"gpt-6.1-sol", "gpt-6.1-sol", true},
		{"gpt-6.2-astra", "gpt-6.2-astra", true},
		{"gpt-5.5", "grok-4.5", false},
		{"company-coder", "gpt-5.1", true},
	} {
		t.Run(tc.public+"_to_"+tc.target, func(t *testing.T) {
			const groupID int64 = 707
			account := Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"api_key": "fixture", "model_mapping": map[string]any{tc.public: tc.target}}}
			svc := &GatewayService{accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{groupID: {account}}}}
			body, err := svc.BuildCodexModelsManifestForGroup(context.Background(), &Group{ID: groupID, Platform: PlatformComposite}, "", []string{tc.public})
			require.NoError(t, err)
			models := decodeCodexManifestModels(t, body)
			require.Len(t, models, 1)
			require.Equal(t, tc.public, models[0]["slug"])
			prompt := generatedCodexPrompt(t, models[0])
			base := openai.CodexBaseInstructionsForModel(tc.target)
			if tc.gpt {
				require.Equal(t, base, prompt)
			} else {
				expected := strings.Replace(base, "You are Codex, a coding agent based on GPT-5.", "You are Codex, a coding agent.", 1)
				expected = strings.Replace(expected, "You are Codex, based on GPT-5.", "You are Codex.", 1)
				require.NotEqual(t, base, expected, "fixture must exercise a bundled identity")
				require.Equal(t, expected, prompt, "only the generated opening identity may change")
			}
			require.NotEmpty(t, models[0]["truncation_policy"])
			require.Contains(t, models[0], "availability_nux")
		})
	}
}

func TestCodexExplicitUpstreamInstructionsRemainIntact(t *testing.T) {
	const explicit = "You are GPT-5.2. Follow operator instructions; compare results with GPT-6. {{personality}}"
	raw, err := json.Marshal(map[string]any{"models": []any{map[string]any{"slug": "company-model", "model_messages": map[string]any{"instructions_template": explicit, "extra": "preserved"}, "unknown": "preserved"}}})
	require.NoError(t, err)
	body, _, err := mergeConfiguredCodexModelsManifest(raw, []string{"company-model"}, nil, false, nil)
	require.NoError(t, err)
	model := decodeCodexManifestModels(t, body)[0]
	require.Equal(t, explicit, generatedCodexPrompt(t, model))
	require.Equal(t, "preserved", model["unknown"])
}

func TestGeneratedCodexIdentityForMergedAndCompletedAliases(t *testing.T) {
	for _, tc := range []struct{ public, target string }{{"gpt-5.5", "grok-4.5"}, {"company-model", "gpt-5.1"}} {
		t.Run(tc.public, func(t *testing.T) {
			account := Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Credentials: map[string]any{"base_url": "https://fixture.example.invalid", "model_mapping": map[string]any{tc.public: tc.target}}}
			svc := &OpenAIGatewayService{accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{707: {account}}}}
			manifest := &CodexModelsManifest{Body: []byte(`{"models":[]}`)}
			err := svc.MergeGroupConfiguredCodexModels(context.Background(), &Group{ID: 707, Platform: PlatformOpenAI}, manifest, nil, "")
			require.NoError(t, err)
			expected := openai.CodexBaseInstructionsForModel(tc.target)
			if tc.target == "grok-4.5" {
				expected = strings.Replace(expected, " based on GPT-5", "", 1)
			}
			models := decodeCodexManifestModels(t, manifest.Body)
			require.Len(t, models, 1)
			assert.Equal(t, expected, generatedCodexPrompt(t, models[0]))
			// Missing upstream instructions are generated using the routed identity;
			// explicit upstream instructions still win in the completion path.
			for _, explicit := range []bool{false, true} {
				entry := map[string]any{"slug": tc.public}
				if explicit {
					entry["model_messages"] = map[string]any{"instructions_template": "You are GPT-5. Operator supplied. {{personality}}"}
				}
				body, marshalErr := json.Marshal(map[string]any{"models": []any{entry}})
				require.NoError(t, marshalErr)
				body, err = completeAPIKeyCodexModelsManifestMetadata(body, true, &account)
				require.NoError(t, err)
				prompt := generatedCodexPrompt(t, decodeCodexManifestModels(t, body)[0])
				if explicit {
					require.Equal(t, "You are GPT-5. Operator supplied. {{personality}}", prompt)
				} else {
					require.Equal(t, expected, prompt)
				}
			}
		})
	}
}

func generatedCodexPrompt(t *testing.T, model map[string]any) string {
	t.Helper()
	messages, ok := model["model_messages"].(map[string]any)
	require.True(t, ok)
	prompt, ok := messages["instructions_template"].(string)
	require.True(t, ok)
	return prompt
}
