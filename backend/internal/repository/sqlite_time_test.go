package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSQLiteTimeScannerLegacyAndNative(t *testing.T) {
	want := time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC)
	for _, value := range []any{want, "2026-10-28 12:00:00", "2026-10-28T20:00:00+08:00", []byte("2026-10-28T12:00:00Z")} {
		var scanned sqliteTime
		require.NoError(t, scanned.Scan(value))
		require.True(t, scanned.Valid)
		require.True(t, want.Equal(scanned.Time))
		require.NoError(t, scanned.Scan(nil))
		require.False(t, scanned.Valid)
		require.True(t, scanned.Time.IsZero())
	}
	var scanned sqliteTime
	require.Error(t, scanned.Scan("invalid"))
	require.Error(t, scanned.Scan(123))
	require.False(t, scanned.Valid)
}
