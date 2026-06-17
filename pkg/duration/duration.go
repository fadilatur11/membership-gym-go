package duration

import (
	"fmt"
	"time"
)

func CalculateDurationSeconds(startedAt time.Time, endedAt time.Time) int {
	if endedAt.Before(startedAt) {
		return 0
	}
	return int(endedAt.Sub(startedAt).Seconds())
}

func SecondsToHuman(seconds int) string {
	if seconds <= 0 {
		return "0 menit"
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	if hours > 0 && minutes > 0 {
		return fmt.Sprintf("%d jam %d menit", hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%d jam", hours)
	}
	return fmt.Sprintf("%d menit", minutes)
}
