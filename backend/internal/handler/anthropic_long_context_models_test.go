package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestAnthropicMetadataLongContextModelsHTTP(t *testing.T) {
	for _, tc := range []struct {
		name   string
		window int64
		mapped string
		routed bool
		want   bool
	}{
		{"threshold", 900000, "upstream", false, true},
		{"below threshold", 899999, "upstream", false, false},
		{"mapped metadata only", 1000000, "short", false, true},
		{"excluded route", 1000000, "upstream", true, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			acc := service.Account{ID: 8, Platform: service.PlatformAnthropic, Type: service.AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"public": tc.mapped, "public[1M]": tc.mapped, "plain": "upstream"}}}
			acc.SetUpstreamModelMetadataSnapshot(service.UpstreamModelMetadataSnapshot{Models: map[string]service.UpstreamModelMetadata{"upstream": {ContextWindow: tc.window}, "short": {ContextWindow: 200000}, "public": {ContextWindow: 1000000}}})
			group := &service.Group{ID: 1, Platform: service.PlatformAnthropic}
			if tc.routed {
				group.ModelRoutingEnabled = true
				group.ModelRouting = map[string][]int64{"plain": {99}}
			}
			h := newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {acc}}})
			// Check the helper separately from serialization to preserve an existing suffix.
			expanded := h.gatewayService.ExpandAnthropicLongContextModels(context.Background(), group, []string{"public", "public[1M]", "plain"})
			require.NotContains(t, expanded, "public[1M][1m]")
			require.Equal(t, tc.want, strings.Contains(strings.Join(expanded, ","), "plain[1m]"))
			gin.SetMode(gin.TestMode)
			r := gin.New()
			r.GET("/v1/models", func(c *gin.Context) {
				id := int64(1)
				c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{GroupID: &id, Group: group})
				h.Models(c)
			})
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/v1/models", nil))
			require.Equal(t, 200, w.Code)
			require.Equal(t, tc.want, strings.Contains(w.Body.String(), `"id":"plain[1m]"`))
		})
	}
}

func TestAnthropicLongContextIntersectionAndMappedMetadata(t *testing.T) {
	makeAccount := func(id int64, window int64) service.Account {
		a := service.Account{ID: id, Platform: service.PlatformAnthropic, Type: service.AccountTypeAPIKey, Credentials: map[string]any{"model_mapping": map[string]any{"public": "upstream"}}}
		if window > 0 {
			a.SetUpstreamModelMetadataSnapshot(service.UpstreamModelMetadataSnapshot{Models: map[string]service.UpstreamModelMetadata{"upstream": {ContextWindow: window}}})
		}
		return a
	}
	for _, tc := range []struct {
		name     string
		accounts []service.Account
		route    []int64
		want     bool
	}{
		{"all long", []service.Account{makeAccount(1, 1000000), makeAccount(2, 900000)}, nil, true},
		{"short candidate", []service.Account{makeAccount(1, 1000000), makeAccount(2, 200000)}, nil, false},
		{"unknown candidate", []service.Account{makeAccount(1, 1000000), makeAccount(2, 0)}, nil, false},
		{"routed to long", []service.Account{makeAccount(1, 1000000), makeAccount(2, 200000)}, []int64{1}, true},
		{"no candidate", nil, nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			h := newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: tc.accounts}})
			group := &service.Group{ID: 1, Platform: service.PlatformAnthropic, ModelRoutingEnabled: len(tc.route) > 0, ModelRouting: map[string][]int64{"public": tc.route}}
			got := h.gatewayService.ExpandAnthropicLongContextModels(context.Background(), group, []string{"public"})
			require.Equal(t, tc.want, strings.Contains(strings.Join(got, ","), "public[1m]"))
		})
	}
	a := makeAccount(1, 200000)
	a.SetUpstreamModelMetadataSnapshot(service.UpstreamModelMetadataSnapshot{Models: map[string]service.UpstreamModelMetadata{"upstream": {ContextWindow: 200000}, "public": {ContextWindow: 1000000}}})
	h := newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{byGroup: map[int64][]service.Account{1: {a}}})
	got := h.gatewayService.ExpandAnthropicLongContextModels(context.Background(), &service.Group{ID: 1}, []string{"public"})
	require.Equal(t, []string{"public"}, got, "public-name metadata must not override the mapped short target")
}
