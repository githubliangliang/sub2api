//go:build unit

package service

import (
	"context"
	"errors"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestAdminService_CreateUser_SnapshotsDefaultPlatformQuotas(t *testing.T) {
	users := &userRepoStub{nextID: 40}
	quotas := &userPlatformQuotaRepoStub{}
	settings := NewSettingService(&settingRepoStub{values: map[string]string{
		SettingKeyDefaultPlatformQuotas: `{
			"openai": {"daily": 100, "weekly": 700, "monthly": 3000},
			"anthropic": {"daily": 100},
			"gemini": {"daily": 500},
			"antigravity": {"daily": 500},
			"grok": {"daily": 0},
			"openai_unused": {},
			"unused": null
		}`,
		SettingKeyAuthSourceDefaultEmailGrantOnSignup: "true",
		SettingKeyAuthSourcePlatformQuotas("email"):   `{"openai":{"daily":3}}`,
	}}, &config.Config{})
	svc := NewAdminService(users, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, settings, nil, nil, nil, nil, nil, nil, nil, nil, quotas)

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "quota-defaults@example.test",
		Password: "strong-pass",
	})

	require.NoError(t, err)
	require.Equal(t, int64(40), user.ID)
	require.Len(t, quotas.bulkInsertCalls, 1, "admin-created users must receive the configured platform quotas")
	records := make(map[string]UserPlatformQuotaRecord)
	for _, record := range quotas.bulkInsertCalls[0] {
		require.Equal(t, user.ID, record.UserID)
		records[record.Platform] = record
	}
	require.Len(t, records, 5, "unconfigured platforms must not receive quota rows")
	require.Equal(t, 100.0, *records["openai"].DailyLimitUSD, "admin creation uses system defaults, not email signup overrides")
	require.Equal(t, 700.0, *records["openai"].WeeklyLimitUSD)
	require.Equal(t, 3000.0, *records["openai"].MonthlyLimitUSD)
	require.Equal(t, 100.0, *records["anthropic"].DailyLimitUSD)
	require.Equal(t, 500.0, *records["gemini"].DailyLimitUSD)
	require.Equal(t, 500.0, *records["antigravity"].DailyLimitUSD)
	require.NotNil(t, records["grok"].DailyLimitUSD, "zero is an explicit limit")
	require.Zero(t, *records["grok"].DailyLimitUSD)
	require.Nil(t, records["anthropic"].WeeklyLimitUSD)
	require.Nil(t, records["anthropic"].MonthlyLimitUSD)
}

func TestAdminService_CreateUser_UnlimitedDefaultsDoNotCreateQuotaRows(t *testing.T) {
	quotas := &userPlatformQuotaRepoStub{}
	svc := &adminServiceImpl{
		userRepo: &userRepoStub{nextID: 41},
		settingService: NewSettingService(&settingRepoStub{values: map[string]string{
			SettingKeyDefaultPlatformQuotas: `{"openai":{"daily":null,"weekly":null,"monthly":null}}`,
		}}, &config.Config{}),
		userPlatformQuotaRepo: quotas,
	}

	_, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "unlimited@example.test",
		Password: "strong-pass",
	})

	require.NoError(t, err)
	require.Empty(t, quotas.bulkInsertCalls)
}

func TestAdminService_CreateUser_AdminSkipsDefaultPlatformQuotas(t *testing.T) {
	quotas := &userPlatformQuotaRepoStub{}
	svc := &adminServiceImpl{
		userRepo: &userRepoStub{nextID: 42},
		settingService: NewSettingService(&settingRepoStub{values: map[string]string{
			SettingKeyDefaultPlatformQuotas: `{"openai":{"daily":100}}`,
		}}, &config.Config{}),
		userPlatformQuotaRepo: quotas,
	}

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "admin@example.test",
		Password: "strong-pass",
		Role:     RoleAdmin,
	})

	require.NoError(t, err)
	require.Equal(t, RoleAdmin, user.Role)
	require.Empty(t, quotas.bulkInsertCalls)
}

func TestAdminService_CreateUser_QuotaWriteFailureDoesNotFailCreation(t *testing.T) {
	users := &userRepoStub{nextID: 43}
	quotas := &userPlatformQuotaRepoStub{bulkInsertErr: errors.New("quota insert failed")}
	svc := &adminServiceImpl{
		userRepo: users,
		settingService: NewSettingService(&settingRepoStub{values: map[string]string{
			SettingKeyDefaultPlatformQuotas: `{"openai":{"daily":100}}`,
		}}, &config.Config{}),
		userPlatformQuotaRepo: quotas,
	}

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "quota-failure@example.test",
		Password: "strong-pass",
	})

	require.NoError(t, err)
	require.Equal(t, int64(43), user.ID)
	require.Len(t, users.created, 1)
	require.Len(t, quotas.bulkInsertCalls, 1)
}
