package service

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestResolveBillingServiceTier(t *testing.T) {
	tests := []struct {
		name       string
		requested  string
		observed   string
		billing    string
		downgraded bool
	}{
		{name: "openai priority served as default", requested: "priority", observed: "default", billing: "default", downgraded: true},
		{name: "anthropic fast served as standard", requested: "fast", observed: "standard", billing: "standard", downgraded: true},
		{name: "ultrafast served as priority", requested: "ultrafast", observed: "priority", billing: "priority", downgraded: true},
		{name: "ultrafast served as default", requested: "ultrafast", observed: "default", billing: "default", downgraded: true},
		{name: "ultrafast honoured", requested: "ultrafast", observed: "ultrafast", billing: "ultrafast"},
		{name: "ultrafast missing declaration", requested: "ultrafast", billing: "ultrafast"},
		{name: "response cannot upgrade to ultrafast", requested: "priority", observed: "ultrafast", billing: "priority"},
		{name: "priority honoured", requested: "priority", observed: "priority", billing: "priority"},
		{name: "no declaration keeps request", requested: "priority", observed: "", billing: "priority"},
		{name: "no request no declaration", requested: "", observed: "", billing: ""},
		{name: "response never raises the tier", requested: "", observed: "priority", billing: ""},
		{name: "flex never raised to default", requested: "flex", observed: "default", billing: "flex"},
		{name: "default echoed for untiered request", requested: "", observed: "default", billing: ""},
		{name: "unknown response tier ignored", requested: "priority", observed: "turbo", billing: "priority"},
		{name: "case and whitespace normalised", requested: " Priority ", observed: "DEFAULT", billing: "default", downgraded: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveBillingServiceTier(tt.requested, tt.observed)
			require.Equal(t, tt.billing, got.Billing)
			require.Equal(t, tt.downgraded, got.Downgraded)
		})
	}
}
