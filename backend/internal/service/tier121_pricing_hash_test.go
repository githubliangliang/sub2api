package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type tier121PricingRemote struct {
	body             []byte
	hash             string
	downloads        int
	bodyErr, hashErr error
}

func (r *tier121PricingRemote) FetchPricingJSON(context.Context, string) ([]byte, error) {
	r.downloads++
	return r.body, r.bodyErr
}
func (r *tier121PricingRemote) FetchHashText(context.Context, string) (string, error) {
	return r.hash, r.hashErr
}
func tier121PricingService(t *testing.T, r *tier121PricingRemote) *PricingService {
	t.Helper()
	return NewPricingService(&config.Config{Pricing: config.PricingConfig{DataDir: t.TempDir(), RemoteURL: "https://example.com/pricing.json", HashURL: "https://example.com/pricing.sha256"}}, r)
}
func TestTier121PricingHash(t *testing.T) {
	old := []byte(`{"model-test":{"input_cost_per_token":0.000001,"output_cost_per_token":0.000002}}`)
	next := []byte(`{"model-test":{"input_cost_per_token":0.000003,"output_cost_per_token":0.000006}}`)
	r := &tier121PricingRemote{body: old, hash: fmt.Sprintf("%x", sha256.Sum256(next))}
	s := tier121PricingService(t, r)
	require.NoError(t, s.ForceUpdate())
	require.Equal(t, fmt.Sprintf("%x", sha256.Sum256(old)), s.localHash)
	saved, err := os.ReadFile(s.getHashFilePath())
	require.NoError(t, err)
	require.Equal(t, s.localHash+"\n", string(saved))
	r.body = next
	require.NoError(t, s.syncWithRemote())
	require.Equal(t, 2, r.downloads)
	require.InDelta(t, 3e-6, s.GetModelPricing("model-test").InputCostPerToken, 1e-12)
	require.NoError(t, s.syncWithRemote())
	require.Equal(t, 2, r.downloads)
	saved, err = os.ReadFile(s.getPricingFilePath())
	require.NoError(t, err)
	require.Equal(t, next, saved)
	// Hash casing does not change the identity of the downloaded bytes.
	r.hash = strings.ToUpper(r.hash)
	require.NoError(t, s.syncWithRemote())
	require.Equal(t, 2, r.downloads)
}
func TestTier121PricingFailureDoesNotPublish(t *testing.T) {
	body := []byte(`{"model-test":{"input_cost_per_token":0.000001,"output_cost_per_token":0.000002}}`)
	for _, mode := range []string{"invalid-json", "network", "missing-hash", "hash-network", "uppercase"} {
		t.Run(mode, func(t *testing.T) {
			r := &tier121PricingRemote{body: body, hash: fmt.Sprintf("%x", sha256.Sum256(body))}
			s := tier121PricingService(t, r)
			require.NoError(t, s.ForceUpdate())
			before := s.localHash
			switch mode {
			case "invalid-json":
				r.body = []byte(`{"broken":`)
			case "network":
				r.bodyErr = errors.New("synthetic transport failure")
			case "missing-hash":
				r.hash = ""
			case "hash-network":
				r.hashErr = errors.New("synthetic hash failure")
			case "uppercase":
				r.hash = strings.ToUpper(r.hash)
			}
			err := s.ForceUpdate()
			if mode == "invalid-json" || mode == "network" {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.True(t, strings.EqualFold(before, s.localHash))
			require.InDelta(t, 1e-6, s.GetModelPricing("model-test").InputCostPerToken, 1e-12)
			saved, err := os.ReadFile(s.getPricingFilePath())
			require.NoError(t, err)
			require.Equal(t, body, saved)
		})
	}
}
