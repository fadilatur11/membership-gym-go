package datetime

import "time"

func NowInTimezone(tz string) time.Time {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.FixedZone("Asia/Jakarta", 7*60*60)
	}
	return time.Now().In(loc)
}

func TodayInTimezone(tz string) time.Time {
	return DateOnly(NowInTimezone(tz))
}

func DateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func DaysRemaining(endDate time.Time, now time.Time) int {
	return int(DateOnly(endDate).Sub(DateOnly(now)).Hours() / 24)
}
