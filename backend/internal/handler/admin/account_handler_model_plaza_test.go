package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupAccountModelPlazaRouter(adminSvc service.AdminService, pricingSvc *service.PricingService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewAccountHandler(adminSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler.SetPricingService(pricingSvc)
	router.GET("/api/v1/admin/model-plaza", handler.ListAccountModelPlaza)
	return router
}

func TestListAccountModelPlaza_AggregatesActiveAccounts(t *testing.T) {
	svc := newStubAdminService()
	svc.accounts = []service.Account{
		{
			ID:       1,
			Name:     "yj",
			Platform: service.PlatformGrok,
			Type:     service.AccountTypeAPIKey,
			Status:   service.StatusActive,
			Credentials: map[string]any{
				"model_mapping": map[string]any{"grok-4.5": "grok-4.5"},
			},
		},
		{
			ID:       2,
			Name:     "shu26",
			Platform: service.PlatformOpenAI,
			Type:     service.AccountTypeAPIKey,
			Status:   service.StatusActive,
			Credentials: map[string]any{
				"model_mapping": map[string]any{
					"gpt-5.4": "gpt-5.4",
					"gpt-5.5": "gpt-5.5",
				},
			},
		},
		{
			ID:       3,
			Name:     "tabitoken",
			Platform: service.PlatformAnthropic,
			Type:     service.AccountTypeAPIKey,
			Status:   service.StatusError,
			Credentials: map[string]any{
				"model_mapping": map[string]any{"claude-opus-5": "claude-opus-5"},
			},
		},
	}

	router := setupAccountModelPlazaRouter(svc, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/model-plaza", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Code int `json:"code"`
		Data struct {
			Items []service.AccountModelPlazaItem `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, 0, body.Code)
	require.Equal(t, []service.AccountModelPlazaItem{
		{ClientID: "grok-4.5", Platform: service.PlatformGrok, UpstreamIDs: []string{"grok-4.5"}, AccountCount: 1},
		{ClientID: "gpt-5.4", Platform: service.PlatformOpenAI, UpstreamIDs: []string{"gpt-5.4"}, AccountCount: 1},
		{ClientID: "gpt-5.5", Platform: service.PlatformOpenAI, UpstreamIDs: []string{"gpt-5.5"}, AccountCount: 1},
	}, body.Data.Items)
}

func TestListAccountModelPlaza_EmptyWhenNoActiveAccounts(t *testing.T) {
	svc := newStubAdminService()
	svc.accounts = []service.Account{
		{ID: 3, Platform: service.PlatformAnthropic, Status: service.StatusError},
	}

	router := setupAccountModelPlazaRouter(svc, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/model-plaza", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Data struct {
			Items []service.AccountModelPlazaItem `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Empty(t, body.Data.Items)
}

func TestListAccountModelPlaza_ReferencePricing(t *testing.T) {
	priceDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(priceDir, "model_pricing.json"), []byte(`{
		"catalog-model": {
			"input_cost_per_token": 0.0000025,
			"output_cost_per_token": 0.000015,
			"cache_creation_input_token_cost": 0.000003125,
			"cache_creation_input_token_cost_above_1hr": 0.000005,
			"cache_read_input_token_cost": 0.00000025
		},
		"priced-alias": {"input_cost_per_token": 0.000007, "output_cost_per_token": 0.000011},
		"free-model": {"input_cost_per_token": 0, "output_cost_per_token": 0, "cache_creation_input_token_cost": 0},
		"image-only": {"mode": "image_generation", "output_cost_per_image": 0.04}
	}`), 0600))
	pricingSvc := service.NewPricingService(&config.Config{
		Pricing: config.PricingConfig{DataDir: priceDir},
	}, nil)
	require.NoError(t, pricingSvc.Initialize())
	t.Cleanup(pricingSvc.Stop)

	svc := newStubAdminService()
	svc.accounts = []service.Account{
		{
			ID: 1, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Status: service.StatusActive,
			Credentials: map[string]any{"model_mapping": map[string]any{
				"catalog-model": "catalog-model",
				"alias-review":  "catalog-model",
				"priced-alias":  "catalog-model",
				"free-model":    "free-model",
				"custom-opus":   "unknown-upstream",
				"image-only":    "image-only",
				"shared-alias":  "catalog-model",
			}},
		},
		{
			ID: 2, Platform: service.PlatformOpenAI, Type: service.AccountTypeAPIKey, Status: service.StatusActive,
			Credentials: map[string]any{"model_mapping": map[string]any{"shared-alias": "other-model"}},
		},
	}
	router := setupAccountModelPlazaRouter(svc, pricingSvc)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/admin/model-plaza", nil))
	require.Equal(t, http.StatusOK, rec.Code)

	var body struct {
		Data struct {
			Items []service.AccountModelPlazaItem `json:"items"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	byID := make(map[string]*service.AccountModelPlazaPricing)
	for _, item := range body.Data.Items {
		byID[item.ClientID] = item.Pricing
	}
	require.Len(t, byID, 7)
	price := byID["catalog-model"]
	require.NotNil(t, price)
	require.Equal(t, "catalog-model", price.Model)
	require.Equal(t, 2.5e-6, price.InputPrice)
	require.Equal(t, 15e-6, price.OutputPrice)
	require.Equal(t, 3.125e-6, *price.CacheWritePrice)
	require.Equal(t, 5e-6, *price.CacheWrite1hPrice)
	require.Equal(t, 0.25e-6, *price.CacheReadPrice)
	require.Equal(t, price, byID["alias-review"], "an unpriced alias uses its sole upstream model")
	require.Equal(t, 7e-6, byID["priced-alias"].InputPrice, "the client ID's own price takes precedence")
	require.Equal(t, "priced-alias", byID["priced-alias"].Model)
	require.NotNil(t, byID["free-model"])
	require.Zero(t, byID["free-model"].InputPrice)
	require.Zero(t, byID["free-model"].OutputPrice)
	require.NotNil(t, byID["free-model"].CacheWritePrice)
	require.Zero(t, *byID["free-model"].CacheWritePrice)
	require.Nil(t, byID["custom-opus"], "a familiar name fragment must not produce a guessed price")
	require.Nil(t, byID["image-only"], "per-image prices must not be displayed as free token prices")
	require.Nil(t, byID["shared-alias"], "multiple upstream models have no unique reference price")
}
