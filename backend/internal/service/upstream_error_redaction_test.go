package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReplaceCaseInsensitiveOriginalBytes(t *testing.T) {
	for _, prefix := range []string{"", "Ⱥ ", strings.Repeat("Ⱥ", 10), "K ", strings.Repeat("K", 5), "\xff ", strings.Repeat("\xff", 10)} {
		t.Run(prefix, func(t *testing.T) {
			message := prefix + "H.Example and h.example trailing text"
			require.Equal(t, prefix+"[upstream] and [upstream] trailing text", replaceCaseInsensitive(message, "h.example", "[upstream]"))
		})
	}
	require.Equal(t, "hxexample", replaceCaseInsensitive("hxexample", "h.example", "[upstream]"))
	require.Equal(t, "[$1]", replaceCaseInsensitive("[a+b]", "a+b", "$1"))
	require.Empty(t, replaceCaseInsensitive("text \xff", "\xff", "[upstream]"))
	require.Empty(t, replaceCaseInsensitive("", "host", "[upstream]"))
	require.Equal(t, "no match", replaceCaseInsensitive("no match", "host", "[upstream]"))
	require.Equal(t, "text", replaceCaseInsensitive("text", "", "[upstream]"))
}

func TestSanitizeUpstreamErrorMessageForAccount(t *testing.T) {
	account := &Account{Credentials: map[string]any{
		"api_key": "sk-synthetic", "access_token": "sk-synthetic-access",
		"refresh_token": "synthetic-refresh", "base_url": "https://Relay.Example:8443/v1",
	}}
	message := " Context window at RELAY.EXAMPLE:8443 for sk-synthetic-access, sk-synthetic, synthetic-refresh?key=query-secret&access_token=query-access "
	require.Equal(t, "Context window at [upstream]:8443 for [redacted], [redacted], [redacted]?key=***&access_token=***", sanitizeUpstreamErrorMessageForAccount(message, account))
	for _, tc := range []struct {
		name, baseURL, message, want string
	}{
		{"IPv6", "https://[2001:db8::ABCD]:8443/v1", "limit at [2001:DB8::abcd]:8443", "limit at [[upstream]]:8443"},
		{"literal regexp host", "https://a+b.example/v1", "a+b.example axb.example", "[upstream] axb.example"},
		{"no base URL", "", "context window", "context window"},
		{"no hostname", "/v1", "context window", "context window"},
		{"malformed URL", "https://%zz/", "context window", ""},
		{"malformed hostname UTF8", "https://\xff/", "context window \xff", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, sanitizeUpstreamErrorMessageForAccount(tc.message, &Account{Credentials: map[string]any{"base_url": tc.baseURL}}))
		})
	}
	require.Equal(t, "message?refresh_token=***", sanitizeUpstreamErrorMessageForAccount(" message?refresh_token=query ", nil))
	require.Equal(t, "safe", sanitizeUpstreamErrorMessageForAccount("safe", &Account{}))
	require.Empty(t, sanitizeUpstreamErrorMessageForAccount(" ", account))
	require.Equal(t, "[redacted]", sanitizeUpstreamErrorMessageForAccount("abc", &Account{Credentials: map[string]any{"api_key": "abc"}}))
}
