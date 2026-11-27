package admin

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/repository"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupAccountModelPlazaRouter(adminSvc service.AdminService, pricingSvc *service.PricingService, settings ...*service.SettingService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	handler := NewAccountHandler(adminSvc, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	handler.SetPricingService(pricingSvc)
	if len(settings) > 0 {
		handler.SetModelPlazaSettingService(settings[0])
	}
	router.GET("/api/v1/admin/model-plaza", handler.ListAccountModelPlaza)
	router.PUT("/api/v1/admin/model-plaza/pricing", handler.UpdateModelPlazaReferencePricing)
	return router
}

func TestAccountModelPlaza_ManualPricingSQLite(t *testing.T) {
	db, err := sql.Open("sqlite", "file:"+filepath.Join(t.TempDir(), "plaza.db")+"?_time_format=sqlite")
	require.NoError(t, err)
	db.SetMaxOpenConns(1)
	_, err = db.Exec(`CREATE TABLE settings (
		id INTEGER PRIMARY KEY AUTOINCREMENT, key TEXT UNIQUE NOT NULL,
		value TEXT NOT NULL, updated_at DATETIME NOT NULL
	)`)
	require.NoError(t, err)
	client := dbent.NewClient(dbent.Driver(entsql.OpenDB(dialect.SQLite, db)))
	t.Cleanup(func() { _ = client.Close() })
	settings := service.NewSettingService(repository.NewSettingRepository(client), &config.Config{})
	adminSvc := newStubAdminService()
	adminSvc.accounts = nil
	for i, platform := range []string{service.PlatformOpenAI, service.PlatformGrok} {
		adminSvc.accounts = append(adminSvc.accounts, service.Account{
			ID: int64(i + 1), Platform: platform, Type: service.AccountTypeAPIKey, Status: service.StatusActive,
			Credentials: map[string]any{"model_mapping": map[string]any{"custom/model:未定价": "unknown"}},
		})
	}
	router := setupAccountModelPlazaRouter(adminSvc, nil, settings)
	put := func(body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/admin/model-plaza/pricing", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		return rec
	}
	list := func(r *gin.Engine) []service.AccountModelPlazaItem {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/admin/model-plaza", nil))
		require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
		var body struct {
			Data struct {
				Items []service.AccountModelPlazaItem `json:"items"`
			} `json:"data"`
		}
		require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
		return body.Data.Items
	}

	require.Nil(t, list(router)[0].Pricing)
	saved := put(`{"platform":"openai","client_id":"custom/model:未定价","pricing":{"input_price":0,"output_price":0.0000125,"cache_read_price":0,"cache_write_1h_price":0.000003}}`)
	require.Equal(t, http.StatusOK, saved.Code, saved.Body.String())
	// A fresh service/handler must read the stored price rather than process-local state.
	reopened := setupAccountModelPlazaRouter(adminSvc, nil, service.NewSettingService(repository.NewSettingRepository(client), &config.Config{}))
	items := list(reopened)
	require.Len(t, items, 2)
	require.Nil(t, items[0].Pricing, "the same model on another platform must remain unpriced")
	price := items[1].Pricing
	require.NotNil(t, price)
	require.True(t, price.Manual)
	require.Equal(t, "custom/model:未定价", price.Model)
	require.Zero(t, price.InputPrice)
	require.Equal(t, 0.0000125, price.OutputPrice)
	require.NotNil(t, price.CacheReadPrice)
	require.Zero(t, *price.CacheReadPrice)
	require.Nil(t, price.CacheWritePrice)
	require.Equal(t, 0.000003, *price.CacheWrite1hPrice)

	for _, pricing := range []string{
		`{}`, `{"input_price":0}`, `{"input_price":null,"output_price":1}`,
		`{"input_price":-1,"output_price":1}`, `{"input_price":1,"output_price":-1}`,
		`{"input_price":1,"output_price":1,"cache_read_price":-1}`,
		`{"input_price":1,"output_price":1,"cache_write_price":-1}`,
		`{"input_price":1,"output_price":1,"cache_write_1h_price":-1}`,
		`{"input_price":1e999,"output_price":1}`,
	} {
		rejected := put(`{"platform":"openai","client_id":"custom/model:未定价","pricing":` + pricing + `}`)
		require.Equal(t, http.StatusBadRequest, rejected.Code, rejected.Body.String())
	}
	require.Equal(t, price, list(router)[1].Pricing, "invalid requests must retain the saved price")
	missing := put(`{"platform":"openai","client_id":"custom/model:未定价"}`)
	require.Equal(t, http.StatusBadRequest, missing.Code, missing.Body.String())
	require.Equal(t, price, list(router)[1].Pricing, "omitted pricing must not clear the saved price")
	updated := put(`{"platform":"openai","client_id":"custom/model:未定价","pricing":{"input_price":0.000001,"output_price":0}}`)
	require.Equal(t, http.StatusOK, updated.Code, updated.Body.String())
	require.Equal(t, 0.000001, list(router)[1].Pricing.InputPrice)
	otherPlatform := put(`{"platform":"grok","client_id":"custom/model:未定价","pricing":{"input_price":0.000002,"output_price":0.000003}}`)
	require.Equal(t, http.StatusOK, otherPlatform.Code, otherPlatform.Body.String())
	require.Equal(t, 0.000001, list(router)[1].Pricing.InputPrice, "saving another platform must retain this price")
	cleared := put(`{"platform":"openai","client_id":"custom/model:未定价","pricing":null}`)
	require.Equal(t, http.StatusOK, cleared.Code, cleared.Body.String())
	require.Nil(t, list(reopened)[1].Pricing)
	require.Equal(t, 0.000002, list(reopened)[0].Pricing.InputPrice, "clearing one platform must retain the other price")
	var count int
	require.NoError(t, db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM settings`).Scan(&count))
	require.Equal(t, 1, count)
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
