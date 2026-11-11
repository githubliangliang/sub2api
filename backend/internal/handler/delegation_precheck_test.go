package handler

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDelegationPrecheckMatchesExistingNormalizer(t *testing.T) {
	for _, body := range []string{
		`{"input":[{"type":"message","role":"user","content":"plain"}]}`,
		`{"input":[{"type":"function_call_output","namespace":"codex_tui","name":"create_thread","output":"ordinary result"}]}`,
		`{"input":[{"type":"function_call_output","name":"\u0063reate_thread","output":"ordinary"}]}`,
		`{"input":[],"input":[{"type":"function_call_output","name":"send_message_to_thread"}]}`,
		`{"input":[`,
	} {
		original := []byte(body)
		before := bytes.Clone(original)
		want, changed := normalizeCodexCallOutputBootstrap(original, isCodexDelegationCandidate, true)
		got, actualChanged := normalizeCodexDelegationBootstrap(original)
		require.Equal(t, changed, actualChanged, body)
		require.Equal(t, want, got, body)
		require.Equal(t, before, original)
	}
}

func BenchmarkDelegationPrecheck(b *testing.B) {
	body := []byte(`{"input":[{"type":"message","role":"user","content":"` + strings.Repeat("ordinary text ", 32000) + `"}]}`)
	b.Run("Decoded", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = normalizeCodexCallOutputBootstrap(body, isCodexDelegationCandidate, true)
		}
	})
	b.Run("Precheck", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_, _ = normalizeCodexDelegationBootstrap(body)
		}
	})
}
