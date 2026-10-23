package service

import (
	"context"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"testing"
)

func TestGPT55RetirementModelSupport(t *testing.T) {
	for _, typ := range []string{AccountTypeOAuth, AccountTypeSetupToken, AccountTypeAPIKey} {
		for _, passthrough := range []bool{false, true} {
			t.Run(typ+"/passthrough="+map[bool]string{true: "true", false: "false"}[passthrough], func(t *testing.T) {
				account := &Account{Platform: PlatformOpenAI, Type: typ, Extra: map[string]any{"openai_passthrough": passthrough}}
				for _, model := range []string{"gpt-5.5", "openai/gpt-5.5-high", "gpt-5.5-2026-04-23", "GPT_5.5", "gpt-5.5-openai-compact"} {
					require.Equal(t, typ == AccountTypeAPIKey, account.IsModelSupported(model), model)
				}
				require.True(t, account.IsModelSupported("gpt-6.1-sol"))
			})
		}
	}
	for _, mapping := range []map[string]any{
		{"legacy": "gpt-5.5"}, {"legacy*": "gpt-5.5"},
	} {
		account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"model_mapping": mapping}}
		require.False(t, account.IsModelSupported("legacy"))
	}
	account := &Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"model_mapping": map[string]any{"gpt-5.5": "gpt-6.1-sol"}}}
	require.True(t, account.IsModelSupported("gpt-5.5"), "explicit administrator replacement remains authoritative")
	account.Type = AccountTypeAPIKey
	require.True(t, account.IsModelSupported("gpt-5.5"))
	parent := int64(3)
	shadow := &Account{Platform: PlatformOpenAI, Type: AccountTypeAPIKey, ParentAccountID: &parent}
	require.False(t, shadow.IsModelSupported("gpt-5.5"))
}

func TestGPT55RetirementCatalogs(t *testing.T) {
	oauth := Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth}
	api := Account{ID: 2, Platform: PlatformOpenAI, Type: AccountTypeAPIKey}
	group := &Group{ID: 1, Platform: PlatformOpenAI}
	for _, tc := range []struct {
		name      string
		accounts  []Account
		wantGPT55 bool
	}{
		{"oauth", []Account{oauth}, false}, {"api", []Account{api}, true}, {"mixed", []Account{oauth, api}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{1: tc.accounts}}
			gateway := &GatewayService{accountRepo: repo}
			models := gateway.GetAvailableModels(context.Background(), &group.ID, PlatformOpenAI)
			if models != nil {
				require.Equal(t, tc.wantGPT55, containsModelID(models, "gpt-5.5"))
			}
			body, err := buildCodexModelsManifestForAccounts(PlatformOpenAI, []string{"gpt-5.5", "gpt-6.1-sol"}, tc.accounts, nil, true)
			require.NoError(t, err)
			require.Equal(t, tc.wantGPT55, containsManifestModel(body, "gpt-5.5"))
			// A stale upstream body must be filtered before evaluating its old ETag.
			original := []byte(`{"models":[{"slug":"gpt-5.5","unknown":1},{"slug":"gpt-6.1-sol","unknown":2}],"unknown_top":3}`)
			etag := CodexModelsManifestETag(original)
			manifest := &CodexModelsManifest{Body: original, ETag: etag}
			service := &OpenAIGatewayService{accountRepo: repo}
			require.NoError(t, service.MergeGroupConfiguredCodexModels(context.Background(), group, manifest, etag))
			if !tc.wantGPT55 {
				require.False(t, manifest.NotModified)
				require.NotEqual(t, etag, manifest.ETag)
				require.False(t, containsManifestModel(manifest.Body, "gpt-5.5"))
				require.Equal(t, int64(3), gjson.GetBytes(manifest.Body, "unknown_top").Int())
				require.Equal(t, int64(2), gjson.GetBytes(manifest.Body, "models.0.unknown").Int())
			}
		})
	}
	// All configured entries retired: plain discovery must remain an authoritative empty list.
	oauth.Credentials = map[string]any{"model_mapping": map[string]any{"gpt-5.5": "gpt-5.5"}}
	repo := codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{1: {oauth}}}
	models := (&GatewayService{accountRepo: repo}).GetAvailableModels(context.Background(), &group.ID, PlatformOpenAI)
	require.NotNil(t, models)
	require.Empty(t, models)
}

func containsModelID(models []string, id string) bool {
	for _, model := range models {
		if model == id {
			return true
		}
	}
	return false
}

func containsManifestModel(body []byte, id string) bool {
	for _, m := range gjson.GetBytes(body, "models").Array() {
		if m.Get("slug").String() == id {
			return true
		}
	}
	return false
}

func TestGPT55RetirementSchedulerAndDiagnosis(t *testing.T) {
	ctx := context.Background()
	groupID := int64(1)
	oauth := Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true}
	api := Account{ID: 2, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	for _, typ := range []string{AccountTypeOAuth, AccountTypeSetupToken, AccountTypeAPIKey} {
		account := oauth
		account.Type = typ
		scheduler := &defaultOpenAIAccountScheduler{}
		ok, _ := scheduler.isAccountRequestCompatibleReason(ctx, &account, OpenAIAccountScheduleRequest{RequestedModel: "gpt-5.5"})
		require.Equal(t, typ == AccountTypeAPIKey, ok)
		reason := openAICompatibleAccountEligibilityFailureReasonBeforeProfit(ctx, &account, PlatformOpenAI, "gpt-5.5", false, "")
		require.Equal(t, typ == AccountTypeAPIKey, reason == "")
	}
	oauth.Credentials = map[string]any{"compact_model_mapping": map[string]any{"gpt-6.1-sol": "gpt-5.5"}}
	scheduler := &defaultOpenAIAccountScheduler{}
	ok, reason := scheduler.isAccountRequestCompatibleReason(ctx, &oauth, OpenAIAccountScheduleRequest{RequestedModel: "gpt-6.1-sol", RequireCompact: true})
	require.False(t, ok)
	require.Equal(t, "model_not_supported", reason)
	oauth.Credentials = map[string]any{"compact_model_mapping": map[string]any{"gpt-5.5": "gpt-6.1-sol"}}
	ok, _ = scheduler.isAccountRequestCompatibleReason(ctx, &oauth, OpenAIAccountScheduleRequest{RequestedModel: "gpt-5.5", RequireCompact: true})
	require.True(t, ok, "an explicit compact replacement remains usable")
	for _, accounts := range [][]Account{{oauth}, {oauth, api}} {
		svc := &OpenAIGatewayService{accountRepo: codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{groupID: accounts}}}
		diagnosis := svc.DiagnoseModelAvailabilityForPlatform(ctx, &groupID, "gpt-5.5", PlatformOpenAI)
		require.True(t, diagnosis.HasAccountsInPool)
		require.True(t, diagnosis.HasRetiredSubscriptionModel)
		require.Equal(t, len(accounts) == 2, diagnosis.HasModelSupport)
	}
}

func TestGPT55RetirementEmptyListingSource(t *testing.T) {
	// A configured alias targeting retired GPT-5.5 must not reappear in a generated catalog.
	account := Account{Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"model_mapping": map[string]any{"legacy": "gpt-5.5"}}}
	require.Empty(t, openAIConfiguredCodexModelIDsForGroup([]Account{account}, nil))
	body, err := buildCodexModelsManifestForAccounts(PlatformComposite, []string{"gpt-5.5", "gpt-6.1-sol"}, []Account{account}, nil, true)
	require.NoError(t, err)
	require.False(t, containsManifestModel(body, "gpt-5.5"))
}

func TestGPT55RetirementPublicAliasesAndETag(t *testing.T) {
	for _, tc := range []struct {
		name    string
		mapping map[string]any
		want    []string
	}{
		{"retired alias", map[string]any{"legacy": "gpt-5.5"}, []string{"gpt-6.1-sol"}},
		{"replacement alias", map[string]any{"gpt-5.5": "gpt-6.1-sol", "legacy": "gpt-5.5"}, []string{"gpt-5.5", "gpt-6.1-sol"}},
		{"crossed aliases", map[string]any{"gpt-5.5": "gpt-6.1-sol", "gpt-6.1-sol": "gpt-5.5"}, []string{"legacy", "gpt-5.5"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			account := Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Credentials: map[string]any{"model_mapping": tc.mapping}}
			repo := codexModelsVisibilityAccountRepo{byGroup: map[int64][]Account{1: {account}}}
			svc := &OpenAIGatewayService{accountRepo: repo}
			original := []byte(`{"models":[{"slug":"legacy","extra":1},{"slug":"gpt-5.5","extra":2},{"slug":"gpt-6.1-sol","extra":3}],"extra_top":true}`)
			etag := CodexModelsManifestETag(original)
			response := &CodexModelsManifest{Body: append([]byte(nil), original...), ETag: etag}
			group := &Group{ID: 1, Platform: PlatformOpenAI}
			require.NoError(t, svc.MergeGroupConfiguredCodexModels(context.Background(), group, response, etag))
			require.False(t, response.NotModified)
			require.ElementsMatch(t, tc.want, codexManifestModelSlugs(t, response.Body))
			require.True(t, gjson.GetBytes(response.Body, "extra_top").Bool())
			require.Equal(t, codexModelsManifestBodyETag(response.Body), response.ETag)
			finalETag := response.ETag
			require.NoError(t, svc.MergeGroupConfiguredCodexModels(context.Background(), group, response, finalETag))
			require.True(t, response.NotModified, "remote and downloaded file clients share the final filtered catalog ETag")
			require.Contains(t, string(original), `"slug":"gpt-5.5"`, "shared upstream payload must not be modified")
		})
	}
}
