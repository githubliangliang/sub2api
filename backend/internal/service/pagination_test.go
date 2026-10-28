package service

import (
	"math"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

func TestPaginateAPIKeysLargePage(t *testing.T) {
	t.Parallel()
	for _, size := range []int{0, 1, 2, 3, 20, 21, 1000, 1001} {
		params := pagination.PaginationParams{Page: math.MaxInt, PageSize: size}
		var keys []APIKey
		require.NotPanics(t, func() { keys = paginateAPIKeys([]APIKey{{ID: 1}}, params) }, "page_size=%d", size)
		require.Empty(t, keys, "page_size=%d", size)
	}
}
