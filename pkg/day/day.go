package day

import (
	"time"

	"membership-gym/pkg/datetime"
)

func DayOfWeekName(day int) string {
	names := map[int]string{
		1: "Monday",
		2: "Tuesday",
		3: "Wednesday",
		4: "Thursday",
		5: "Friday",
		6: "Saturday",
		7: "Sunday",
	}
	return names[day]
}

func ValidateDayOfWeek(day int) bool {
	return day >= 1 && day <= 7
}

func GetTodayDayOfWeekByTimezone(timezone string) int {
	weekday := datetime.NowInTimezone(timezone).Weekday()
	if weekday == time.Sunday {
		return 7
	}
	return int(weekday)
}
