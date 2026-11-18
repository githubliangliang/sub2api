//go:build unit

package repository

import (
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestPricingStartupHTTPHashDoesNotBlockAndStopCancelsBody(t *testing.T) {
	for _, failure := range []bool{false, true} {
		t.Run(fmt.Sprintf("hash_failure=%t", failure), func(t *testing.T) {
			started, ended := make(chan struct{}), make(chan struct{})
			var once sync.Once
			var downloads atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/prices" {
					downloads.Add(1)
					return
				}
				once.Do(func() { close(started) })
				defer close(ended)
				if failure {
					w.WriteHeader(http.StatusBadGateway)
					return
				}
				w.WriteHeader(http.StatusOK)
				_, _ = io.WriteString(w, "partial hash ")
				w.(http.Flusher).Flush()
				<-r.Context().Done()
			}))
			t.Cleanup(server.Close)
			cfg := &config.Config{Pricing: config.PricingConfig{DataDir: t.TempDir(), RemoteURL: server.URL + "/prices", HashURL: server.URL + "/hash"}}
			cfg.Security.URLAllowlist.AllowInsecureHTTP = true
			body := []byte(`{"claude-sonnet-4-6":{"input_cost_per_token":0,"output_cost_per_token":0.000003}}`)
			require.NoError(t, os.WriteFile(filepath.Join(cfg.Pricing.DataDir, "model_pricing.json"), body, 0600))
			require.NoError(t, os.WriteFile(filepath.Join(cfg.Pricing.DataDir, "model_pricing.sha256"), []byte(fmt.Sprintf("%x", sha256.Sum256(body))), 0600))
			svc := service.NewPricingService(cfg, NewPricingRemoteClient("", false))
			t.Cleanup(svc.Stop)
			before := time.Now()
			require.NoError(t, svc.Initialize())
			require.Less(t, time.Since(before), 250*time.Millisecond)
			require.Zero(t, svc.GetModelPricing("claude-sonnet-4-6").InputCostPerToken)
			select {
			case <-started:
			case <-time.After(time.Second):
				t.Fatal("startup did not reach the real HTTP hash request")
			}
			svc.Stop()
			select {
			case <-ended:
			case <-time.After(time.Second):
				t.Fatal("Stop did not cancel the remote HTTP response body")
			}
			require.Zero(t, downloads.Load())
			saved, err := os.ReadFile(filepath.Join(cfg.Pricing.DataDir, "model_pricing.json"))
			require.NoError(t, err)
			require.Equal(t, body, saved)
		})
	}
}
