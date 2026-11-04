package repository

import (
	"context"
	"fmt"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

func (r *opsRepository) queryOutputTPS(ctx context.Context, filter *service.OpsDashboardFilter) (*service.OpsOutputTPS, error) {
	join, where, args, _ := buildUsageWhere(filter, filter.StartTime.UTC(), filter.EndTime.UTC(), 1)
	query := `
WITH rates AS (
  SELECT ul.output_tokens * 1000.0 / ul.duration_ms AS output_tps
  FROM usage_logs ul
  ` + join + `
  ` + where + `
    AND ul.output_tokens > 0 AND ul.duration_ms > 0
    AND ul.image_count = 0 AND ul.image_output_tokens = 0
    AND COALESCE(ul.billing_mode, '') <> 'image'
    ` + fmt.Sprintf("AND ul.request_type IN (%d, %d, %d, %d)",
		service.RequestTypeSync, service.RequestTypeStream, service.RequestTypeWSV2, service.RequestTypeCyberBlocked) + `
), ranked AS (
  SELECT output_tps, ROW_NUMBER() OVER (ORDER BY output_tps) AS rn, COUNT(*) OVER () AS cnt
  FROM rates
), bounds AS (
  SELECT
    MAX(CASE WHEN rn = 1 + (cnt - 1) * 5 / 100 THEN output_tps END) AS p5_lower,
    MAX(CASE WHEN rn = 1 + ((cnt - 1) * 5 + 99) / 100 THEN output_tps END) AS p5_upper,
    MAX(CASE WHEN rn = 1 + (cnt - 1) * 10 / 100 THEN output_tps END) AS p10_lower,
    MAX(CASE WHEN rn = 1 + ((cnt - 1) * 10 + 99) / 100 THEN output_tps END) AS p10_upper,
    MAX(CASE WHEN rn = 1 + (cnt - 1) * 50 / 100 THEN output_tps END) AS p50_lower,
    MAX(CASE WHEN rn = 1 + ((cnt - 1) * 50 + 99) / 100 THEN output_tps END) AS p50_upper,
    AVG(output_tps) AS avg,
    COUNT(*) AS sample_count
  FROM ranked
)
SELECT
  p5_lower + (p5_upper - p5_lower) * ((sample_count - 1) * 5 % 100) / 100.0,
  p10_lower + (p10_upper - p10_lower) * ((sample_count - 1) * 10 % 100) / 100.0,
  p50_lower + (p50_upper - p50_lower) * ((sample_count - 1) * 50 % 100) / 100.0,
  avg, sample_count
FROM bounds`
	// Interpolate adjacent ranks at (n-1)*p, retaining fractional values and
	// empty-set NULLs without loading the individual samples into Go.
	var stats service.OpsOutputTPS
	if err := r.db.QueryRowContext(ctx, query, args...).Scan(&stats.P5, &stats.P10, &stats.P50, &stats.Avg, &stats.SampleCount); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		return nil, err
	}
	return &stats, nil
}
