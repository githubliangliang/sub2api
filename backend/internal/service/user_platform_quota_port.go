package service

import (
	"context"
	"errors"
	"time"
)

// ErrUserPlatformQuotaNotFound service 层 sentinel：quota 记录不存在。
// adapter 将 repository.ErrUserPlatformQuotaNotFound 包装为此错误，
// handler 只需引用 service 包，无需直接依赖 repository 包。
var ErrUserPlatformQuotaNotFound = errors.New("user platform quota not found")

// ErrUserPlatformQuotaFKViolation service 层 sentinel：批量 snapshot UPSERT 时存在
// user_id 不在 users 表的记录（外键违反）。adapter 负责将 repository 层同名 sentinel 包装为此错误。
var ErrUserPlatformQuotaFKViolation = errors.New("user platform quota snapshot FK violation")

// UserPlatformQuotaSnapshot 是 service 层 flusher 向 DB 写入快照时使用的传输结构。
// 字段语义与 repository.UserPlatformQuotaSnapshot 完全对应，由 adapter 负责转换。
type UserPlatformQuotaSnapshot struct {
	UserID             int64
	Platform           string
	DailyUsageUSD      float64
	WeeklyUsageUSD     float64
	MonthlyUsageUSD    float64
	DailyWindowStart   time.Time
	WeeklyWindowStart  time.Time
	MonthlyWindowStart time.Time
}

// UserPlatformQuotaRecord service 层传输结构体（与 repository 层解耦）。
type UserPlatformQuotaRecord struct {
	UserID          int64
	Platform        string
	DailyLimitUSD   *float64
	WeeklyLimitUSD  *float64
	MonthlyLimitUSD *float64
	DailyUsageUSD   float64
	WeeklyUsageUSD  float64
	MonthlyUsageUSD float64
	// 窗口起始时间（可选，用于未来 reset 校验）
	DailyWindowStart   *time.Time
	WeeklyWindowStart  *time.Time
	MonthlyWindowStart *time.Time
}

// HasAnyLimit includes explicit zero limits; an empty record is unlimited.
func (r UserPlatformQuotaRecord) HasAnyLimit() bool {
	return r.DailyLimitUSD != nil || r.WeeklyLimitUSD != nil || r.MonthlyLimitUSD != nil
}

// UserPlatformQuotaRepository 定义 service 层所需的 user × platform quota 数据访问端口。
// repository 包的 userPlatformQuotaRepository 实现此接口。
type UserPlatformQuotaRepository interface {
	// GetByUserPlatform 查询单条配额记录，未找到时返回 (nil, nil)。
	GetByUserPlatform(ctx context.Context, userID int64, platform string) (*UserPlatformQuotaRecord, error)
	// BulkInsertInitial skips unlimited records and only fills NULL limits on conflicts.
	BulkInsertInitial(ctx context.Context, records []UserPlatformQuotaRecord) error
	// IncrementUsageWithReset only updates configured active rows; missing rows are skipped.
	IncrementUsageWithReset(ctx context.Context, userID int64, platform string, cost float64, now time.Time) error
	// ListByUser 查询用户的所有平台配额记录。
	ListByUser(ctx context.Context, userID int64) ([]UserPlatformQuotaRecord, error)
	// UpsertForUser 全量替换该用户所有平台限额配置（事务内）：
	//   1. Soft-delete omitted platforms and inputs with all three limits NULL.
	//   2. Update configured active rows, preserving usage/windows; otherwise create a fresh row.
	// records 为空时仅执行步骤 1。
	UpsertForUser(ctx context.Context, userID int64, records []UserPlatformQuotaRecord) error
	// ResetExpiredWindow 重置指定窗口（"daily"|"weekly"|"monthly"）的用量与起始时间。
	// 未命中活跃记录时返回（service-side wrapper of repository.ErrUserPlatformQuotaNotFound）。
	ResetExpiredWindow(ctx context.Context, userID int64, platform string, window string, newStart time.Time) error
	// BatchSnapshotUsage atomically updates existing configured active rows; missing/deleted rows are skipped.
	BatchSnapshotUsage(ctx context.Context, snapshots []UserPlatformQuotaSnapshot, now time.Time) error
}
