package sqlutil

import "strings"

func LikeSearch(keyword string) string {
	return "%" + strings.ToLower(strings.TrimSpace(keyword)) + "%"
}

func BuildOrderBy(allowed map[string]string, requested string, fallback string) string {
	if column, ok := allowed[requested]; ok {
		return column
	}
	return fallback
}
