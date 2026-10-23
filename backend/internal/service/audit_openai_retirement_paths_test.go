package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/openai"
	"github.com/stretchr/testify/require"
)

func TestAuditRetirementCatalogUsesConfiguredPoolAcrossCooldown(t *testing.T) {
	groupID := int64(41)
	cooldown := time.Now().Add(time.Hour)
	oauth := Account{ID: 1, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true}
	api := Account{ID: 2, Platform: PlatformOpenAI, Type: AccountTypeAPIKey, Status: StatusActive, Schedulable: true}
	for _, tc := range []struct {
		name      string
		visible   []Account
		catalog   []Account
		wantGPT55 bool
	}{
		{"oauth cooldown", nil, []Account{oauth}, false},
		{"api cooldown", nil, []Account{api}, true},
		{"api cooldown with visible oauth", []Account{oauth}, []Account{oauth, api}, true},
		{"oauth cooldown with visible api", []Account{api}, []Account{oauth, api}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			catalog := append([]Account(nil), tc.catalog...)
			for i := range catalog {
				catalog[i].TempUnschedulableUntil = &cooldown
			}
			repo := splitCodexModelsAccountRepo{
				schedulable: map[int64][]Account{groupID: tc.visible},
				catalog:     map[int64][]Account{groupID: catalog},
			}
			models := (&GatewayService{accountRepo: repo}).GetAvailableModels(context.Background(), &groupID, PlatformOpenAI)
			if !tc.wantGPT55 {
				require.NotNil(t, models, "a cooldown must not resurrect the unfiltered default catalog")
			}
			if models == nil {
				models = openai.DefaultModelIDs()
			}
			require.Equal(t, tc.wantGPT55, containsModelID(models, "gpt-5.5"))
		})
	}
}

func TestAuditPreviousResponseRetirementUsesActualCompactTarget(t *testing.T) {
	for _, tc := range []struct {
		name, typ, model string
		mapping, compact map[string]any
		requireCompact   bool
		want             bool
	}{
		{name: "oauth retired", typ: AccountTypeOAuth, model: "gpt-5.5"},
		{name: "setup retired", typ: AccountTypeSetupToken, model: "gpt-5.5"},
		{name: "apikey remains supported", typ: AccountTypeAPIKey, model: "gpt-5.5", want: true},
		{name: "alias targets retired", typ: AccountTypeOAuth, model: "legacy", mapping: map[string]any{"legacy": "gpt-5.5"}},
		{name: "ordinary explicit replacement", typ: AccountTypeOAuth, model: "gpt-5.5", mapping: map[string]any{"gpt-5.5": "gpt-6.1-sol"}, want: true},
		{name: "compact targets retired", typ: AccountTypeOAuth, model: "gpt-6.1-sol", compact: map[string]any{"gpt-6.1-sol": "gpt-5.5"}, requireCompact: true},
		{name: "compact explicit replacement", typ: AccountTypeOAuth, model: "gpt-5.5", compact: map[string]any{"gpt-5.5": "gpt-6.1-sol"}, requireCompact: true, want: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			groupID := int64(42)
			account := Account{ID: 2, Platform: PlatformOpenAI, Type: tc.typ, Status: StatusActive, Schedulable: true,
				Credentials: map[string]any{"model_mapping": tc.mapping, "compact_model_mapping": tc.compact},
				Extra:       map[string]any{"openai_apikey_responses_websockets_v2_enabled": true, "openai_oauth_responses_websockets_v2_enabled": true, "openai_compact_supported": true},
			}
			store := NewOpenAIWSStateStore(&stubGatewayCache{})
			svc := &OpenAIGatewayService{cfg: newOpenAIWSV2TestConfig(), accountRepo: stubOpenAIAccountRepo{accounts: []Account{account}}, openaiWSStateStore: store}
			require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_before_retirement", account.ID, time.Hour))
			id, selected, _, _ := svc.resolveAccountByPreviousResponseIDForCapability(ctx, &groupID, "resp_before_retirement", tc.model, nil, "", tc.requireCompact)
			require.Equal(t, tc.want, id == account.ID)
			require.Equal(t, tc.want, selected != nil)
			bound, err := store.GetResponseAccount(ctx, groupID, "resp_before_retirement")
			require.NoError(t, err)
			require.Equal(t, account.ID, bound, "unsupported targets must not destroy an otherwise valid response binding")
		})
	}
}

func TestAuditPreviousResponseRetirementRechecksLatestCompactMapping(t *testing.T) {
	ctx := context.Background()
	groupID := int64(43)
	cached := Account{ID: 3, Platform: PlatformOpenAI, Type: AccountTypeOAuth, Status: StatusActive, Schedulable: true,
		Extra: map[string]any{"openai_oauth_responses_websockets_v2_enabled": true, "openai_compact_supported": true},
	}
	latest := cached
	latest.Credentials = map[string]any{"compact_model_mapping": map[string]any{"gpt-6.1-sol": "gpt-5.5"}}
	store := NewOpenAIWSStateStore(&stubGatewayCache{})
	svc := &OpenAIGatewayService{
		cfg:                newOpenAIWSV2TestConfig(),
		accountRepo:        stubOpenAIAccountRepo{accounts: []Account{latest}},
		openaiWSStateStore: store,
		schedulerSnapshot:  &SchedulerSnapshotService{cache: &openAISnapshotCacheStub{accountsByID: map[int64]*Account{cached.ID: &cached}}},
	}
	require.NoError(t, store.BindResponseAccount(ctx, groupID, "resp_stale_mapping", cached.ID, time.Hour))
	id, selected, _, _ := svc.resolveAccountByPreviousResponseIDForCapability(ctx, &groupID, "resp_stale_mapping", "gpt-6.1-sol", nil, "", true)
	require.Zero(t, id, "a stale scheduler snapshot must not bypass the latest compact retirement target")
	require.Nil(t, selected)
	bound, err := store.GetResponseAccount(ctx, groupID, "resp_stale_mapping")
	require.NoError(t, err)
	require.Equal(t, cached.ID, bound)
}
