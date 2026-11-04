package repository

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func BenchmarkOpsOutputTPS(b *testing.B) {
	for _, count := range []int{1000, 10000, 100000} {
		b.Run(fmt.Sprint(count), func(b *testing.B) {
			db, err := sql.Open("sqlite", filepath.Join(b.TempDir(), "ops.db")+"?_time_format=sqlite")
			if err != nil {
				b.Fatal(err)
			}
			b.Cleanup(func() { _ = db.Close() })
			db.SetMaxOpenConns(1)
			if err := ApplyMigrations(context.Background(), db); err != nil {
				b.Fatal(err)
			}
			start := time.Date(2026, 1, 17, 0, 0, 0, 0, time.UTC)
			_, err = db.Exec(`WITH RECURSIVE n(i) AS (SELECT 1 UNION ALL SELECT i+1 FROM n WHERE i < $1)
INSERT INTO usage_logs (user_id,api_key_id,account_id,model,output_tokens,duration_ms,request_type,created_at)
SELECT 1,1,1,'gpt-test',1+i%1000,1000+(i%400)*10,$2,$3 FROM n`, count, service.RequestTypeStream, start)
			if err != nil {
				b.Fatal(err)
			}
			repo := &opsRepository{db: db}
			filter := &service.OpsDashboardFilter{StartTime: start, EndTime: start.Add(time.Hour)}
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				ctx, cancel := context.WithTimeout(context.Background(), opsRawLatencyQueryTimeout)
				stats, err := repo.queryOutputTPS(ctx, filter)
				cancel()
				if err != nil {
					b.Fatal(err)
				}
				if stats.SampleCount != int64(count) {
					b.Fatalf("sample_count=%d, want %d", stats.SampleCount, count)
				}
			}
		})
	}
}
