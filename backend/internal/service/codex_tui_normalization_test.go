package service

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCodexTUIOutboundNormalizedToCLI(t *testing.T) {
	for _, ua := range []string{"", "codex-tui/0.100.0 (Linux 6.8.0; x86_64) xterm-256color (codex-tui; 0.100.0)"} {
		identity := resolveCodexOutboundIdentity(ua)
		require.Equal(t, "codex_cli_rs", identity.originator)
		require.True(t, strings.HasPrefix(identity.userAgent, "codex_cli_rs/"+identity.version))
		require.Equal(t, CodexCanonicalClientVersion(), identity.version)
		require.NotContains(t, identity.userAgent, "(codex-tui;")
		if ua != "" {
			require.Contains(t, identity.userAgent, "(Linux 6.8.0; x86_64) xterm-256color")
		}
	}
}
