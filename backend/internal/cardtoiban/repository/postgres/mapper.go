package postgres

import "time"

func nullableDurationMilliseconds(duration *time.Duration) any {
	if duration == nil {
		return nil
	}

	return duration.Milliseconds()
}
