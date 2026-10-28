package repository

import "time"

// sqliteTime scans both native DATETIME values and historical TEXT time columns.
// Null values remain distinguishable from a populated time.
type sqliteTime struct {
	Time  time.Time
	Valid bool
}

func (t *sqliteTime) Scan(value any) error {
	*t = sqliteTime{}
	if value == nil {
		return nil
	}
	parsed, err := scanSchedulerOutboxTime(value)
	if err != nil {
		return err
	}
	t.Time = parsed
	t.Valid = true
	return nil
}
