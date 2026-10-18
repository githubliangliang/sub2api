package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type moderationScopeSettingRepo struct {
	*contentModerationTestSettingRepo
	called bool
}

func (r *moderationScopeSettingRepo) SetContentModerationConfig(context.Context, string, []int64) error {
	r.called = true
	return ErrGroupNotFound
}

func TestContentModerationUpdateConfigRejectsGroupDeletedDuringSave(t *testing.T) {
	repo := &moderationScopeSettingRepo{contentModerationTestSettingRepo: &contentModerationTestSettingRepo{values: map[string]string{}}}
	svc := NewContentModerationService(repo, nil, nil, nil, nil, nil, nil, nil)
	_, err := svc.UpdateConfig(context.Background(), UpdateContentModerationConfigInput{})
	require.ErrorContains(t, err, "审计分组不存在")
	require.True(t, repo.called)
	require.Empty(t, repo.values, "failed scope validation must not publish a new config")
}
