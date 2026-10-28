package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

var modelMappingSignatureBenchmarkResult uint64

func TestModelMappingSignatureZeroAllocations(t *testing.T) {
	mapping := map[string]any{"claude-alias": "claude-model", "codex-alias": "gpt-model", "grok-alias": "grok-model"}
	allocations := testing.AllocsPerRun(100, func() {
		modelMappingSignatureBenchmarkResult = modelMappingSignature(mapping)
	})
	require.Zero(t, allocations)
}

func TestModelMappingSignatureContentChanges(t *testing.T) {
	first := map[string]any{"a": "one", "b": "two"}
	second := map[string]any{"b": "two", "a": "one"}
	require.Equal(t, modelMappingSignature(first), modelMappingSignature(second))
	first["a"], first["b"] = first["b"], first["a"]
	require.NotEqual(t, modelMappingSignature(first), modelMappingSignature(second))
	first["a"] = 1
	previous := modelMappingSignature(first)
	first["a"] = false
	require.Equal(t, previous, modelMappingSignature(first), "ignored non-string values share one marker")
	first["a"] = ""
	require.NotEqual(t, previous, modelMappingSignature(first))
	require.Zero(t, modelMappingSignature(nil))
}

func TestModelMappingCacheInvalidatesSameSizeMutation(t *testing.T) {
	account := &Account{Platform: PlatformOpenAI, Credentials: map[string]any{
		"model_mapping": map[string]any{"alias": "first"},
	}}
	require.Equal(t, "first", account.GetModelMapping()["alias"])
	mapping, ok := account.Credentials["model_mapping"].(map[string]any)
	require.True(t, ok)
	mapping["alias"] = "second"
	require.Equal(t, "second", account.GetModelMapping()["alias"])
}

func BenchmarkModelMappingSignature(b *testing.B) {
	for _, size := range []int{4, 32, 128} {
		b.Run(fmt.Sprintf("mappings_%d", size), func(b *testing.B) {
			mapping := make(map[string]any, size)
			for i := range size {
				mapping[fmt.Sprintf("model-%03d", i)] = fmt.Sprintf("upstream-%03d", i)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				modelMappingSignatureBenchmarkResult = modelMappingSignature(mapping)
			}
		})
	}
}
