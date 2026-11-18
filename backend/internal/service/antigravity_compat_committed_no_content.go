package service

import (
	"context"
	"errors"
	"log/slog"
)

type antigravityCompatCommittedNoContentError struct{ msg string }

func (e *antigravityCompatCommittedNoContentError) Error() string { return e.msg }

func newAntigravityCompatCommittedNoContentError(msg string) error {
	return &antigravityCompatCommittedNoContentError{msg: msg}
}

func IsAntigravityCompatCommittedNoContent(err error) bool {
	var target *antigravityCompatCommittedNoContentError
	return errors.As(err, &target)
}

func (s *GatewayService) ReleaseStickySessionAfterCommittedNoContent(ctx context.Context, groupID *int64, sessionHash string, accountID int64, err error) bool {
	if s == nil || s.cache == nil || sessionHash == "" || accountID <= 0 || !IsAntigravityCompatCommittedNoContent(err) {
		return false
	}
	cache, ok := s.cache.(StickySessionAccountCompareDeleter)
	if !ok {
		return false
	}
	removed, deleteErr := cache.DeleteSessionAccountIDIfMatch(ctx, derefGroupID(groupID), sessionHash, accountID)
	if deleteErr != nil {
		slog.Warn("antigravity_compat_sticky_release_failed", "group_id", derefGroupID(groupID), "account_id", accountID, "error", deleteErr)
		return false
	}
	return removed
}
