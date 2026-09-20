//go:build unit

package admin

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/stretchr/testify/require"
)

func TestAccountRefreshSuccessData(t *testing.T) {
	account := AccountWithConcurrency{Account: &dto.Account{ID: 42}}

	partial := accountRefreshSuccessData(account, "missing_project_id_temporary")
	partialMap, ok := partial.(map[string]any)
	require.True(t, ok)
	require.Equal(t, account, partialMap["account"])
	require.Equal(t, "missing_project_id_temporary", partialMap["warning"])
	require.NotEmpty(t, partialMap["message"])

	require.Equal(t, account, accountRefreshSuccessData(account, ""))
}
