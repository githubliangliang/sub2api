package repository

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

// 本文件是 0.1.183 §4.1（上游 49752060 + b20f29d1）的 SQLite 回归。
//
// 为什么必须真跑 SQL：channel_monitor_v2_repo_test.go 里既有的两条相关用例都是
// **字符串检查**（strings.ToLower(channelMonitorV2ClassifyErrorsSQL) 里 grep 关键字），
// 语法错误一条都抓不到。而这次改动往 CTE 里加了两个 LEFT JOIN 与一个多层 CASE，
// 正是最容易写出「Go 侧编译通过、SQLite 上语法失败」的形状——上游 b20f29d1 修的就是
// 49752060 写出的 NULLIF 少一个参数的硬 SQL 错误。CLAUDE.md 的 SQLite 约束也点名这一类。

func TestChannelMonitorV2ClassifyErrorsRunsOnSQLiteAndResolvesCompositePlatform(t *testing.T) {
	db := openChannelMonitorSQLite(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `
	CREATE TABLE ops_error_logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		request_id TEXT, created_at DATETIME NOT NULL,
		platform TEXT, group_id INTEGER, account_id INTEGER, user_id INTEGER,
		requested_model TEXT, model TEXT,
		error_type TEXT, error_owner TEXT, error_source TEXT,
		error_message TEXT, upstream_error_message TEXT, upstream_error_detail TEXT,
		error_body TEXT, upstream_errors TEXT,
		status_code INTEGER, upstream_status_code INTEGER,
		is_count_tokens BOOLEAN NOT NULL DEFAULT 0
	);
	CREATE TABLE groups (id INTEGER PRIMARY KEY, platform TEXT);
	CREATE TABLE accounts (id INTEGER PRIMARY KEY, platform TEXT);
	`)
	require.NoError(t, err)

	// composite 分组 + openai 账号：错误行自己记的 platform 是 'composite'。
	_, err = db.ExecContext(ctx, `INSERT INTO groups (id, platform) VALUES (7, 'composite'), (8, 'openai')`)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, `INSERT INTO accounts (id, platform) VALUES (70, 'openai')`)
	require.NoError(t, err)

	now := time.Now().UTC().Truncate(time.Minute)
	insert := func(requestID, platform string, groupID, accountID int64) {
		t.Helper()
		_, err := db.ExecContext(ctx, `INSERT INTO ops_error_logs
			(request_id, created_at, platform, group_id, account_id, user_id,
			 requested_model, model, error_type, error_owner, error_source,
			 error_message, upstream_error_message, upstream_error_detail, error_body,
			 upstream_errors, status_code, upstream_status_code, is_count_tokens)
			VALUES ($1,$2,$3,$4,$5,1,'gpt-5','gpt-5','upstream_error','provider','upstream',
			        'boom','boom','','','[]',500,500,0)`,
			requestID, now.Add(-time.Minute), platform, groupID, accountID)
		require.NoError(t, err)
	}
	insert("req-composite", "composite", 7, 70)
	insert("req-plain", "openai", 8, 70)

	// 关键断言之一：这条 SQL 必须在真 SQLite 上跑通（字符串检查覆盖不到）。
	_, err = db.ExecContext(ctx, channelMonitorV2ClassifyErrorsSQL,
		now.Add(-time.Hour), now.Add(time.Hour))
	require.NoError(t, err, "channelMonitorV2ClassifyErrorsSQL 必须在 SQLite 上可执行")

	// 临时表按 `SELECT *, ... FROM dedup` 投影，没有 request_id 列；用 group_id 区分两行。
	rows, err := db.QueryContext(ctx,
		`SELECT group_id, platform FROM channel_monitor_v2_classified_errors ORDER BY group_id`)
	require.NoError(t, err)
	defer func() { _ = rows.Close() }()

	got := map[int64]string{}
	for rows.Next() {
		var groupID int64
		var platform string
		require.NoError(t, rows.Scan(&groupID, &platform))
		got[groupID] = platform
	}
	require.NoError(t, rows.Err())
	require.Len(t, got, 2, "两行错误都应进入分类表")

	// group 7 是 composite：必须解析到账号的真实平台，否则这行错误会被监控 v2 的
	// 每一个查询按「composite 不是已启用 config platform」过滤掉、面板上看不到。
	require.Equal(t, "openai", got[7], "composite 分组应解析为账号真实平台")
	// group 8 非 composite，保持原样。
	require.Equal(t, "openai", got[8])
}
