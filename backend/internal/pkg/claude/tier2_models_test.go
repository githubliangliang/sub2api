package claude

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTier2DottedOpusEffort(t *testing.T) {
	require.Equal(t, []string{"low", "medium", "high", "xhigh", "max"}, EffortLevelsForModel("anthropic/claude-opus-5.5"))
}
