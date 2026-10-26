package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestLoadMinOutputTokensRejectsExplicitInvalidValue(t *testing.T) {
	for _, value := range []any{0, -1, 15, 16.5, "16.5", "999999999999999999999999999999"} {
		t.Run(fmt.Sprint(value), func(t *testing.T) {
			resetViperWithJWTSecret(t)
			viper.Set("gateway.min_output_tokens", value)
			_, err := Load()
			require.ErrorContains(t, err, "gateway.min_output_tokens")
		})
	}
}

func TestLoadMinOutputTokensDefaultsAndConfiguration(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		resetViperWithJWTSecret(t)
		cfg, err := Load()
		require.NoError(t, err)
		require.Equal(t, 128, cfg.Gateway.MinOutputTokens)
	})
	for _, n := range []int{16, 128, 256} {
		t.Run(fmt.Sprintf("environment_%d", n), func(t *testing.T) {
			resetViperWithJWTSecret(t)
			t.Setenv("GATEWAY_MIN_OUTPUT_TOKENS", fmt.Sprint(n))
			cfg, err := Load()
			require.NoError(t, err)
			require.Equal(t, n, cfg.Gateway.MinOutputTokens)
		})
	}
	t.Run("yaml and environment precedence", func(t *testing.T) {
		resetViperWithJWTSecret(t)
		file := filepath.Join(t.TempDir(), "config.yaml")
		require.NoError(t, os.WriteFile(file, []byte("gateway:\n  min_output_tokens: 256\n"), 0600))
		t.Setenv("CONFIG_FILE", file)
		cfg, err := Load()
		require.NoError(t, err)
		require.Equal(t, 256, cfg.Gateway.MinOutputTokens)
		t.Setenv("GATEWAY_MIN_OUTPUT_TOKENS", "16")
		cfg, err = Load()
		require.NoError(t, err)
		require.Equal(t, 16, cfg.Gateway.MinOutputTokens)
	})
	for _, value := range []string{"0", "15", "16.5", "not-a-number", "999999999999999999999999999999"} {
		t.Run("invalid environment "+value, func(t *testing.T) {
			resetViperWithJWTSecret(t)
			t.Setenv("GATEWAY_MIN_OUTPUT_TOKENS", value)
			_, err := Load()
			require.ErrorContains(t, err, "gateway.min_output_tokens")
		})
	}
}
