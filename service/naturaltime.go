package service

import (
	"fmt"
	"math"
	"time"
)

func NatualTime(base, value time.Time) string {
	if value.Before(base) {
		return pastNaturalTime(base, value)
	} else {
		return futureNaturalTime(base, value)
	}
}

func futureNaturalTime(base, value time.Time) string {
	dur := value.Sub(base)
	if dur.Seconds() <= 60 {
		return "in a few seconds"
	}
	if dur.Minutes() < 5 {
		return "in a few minutes"
	}
	if dur.Minutes() < 60 {
		return fmt.Sprintf("in %.0f minutes", dur.Minutes())
	}
	if dur.Hours() < 24 {
		return fmt.Sprintf("in %.0f hours", dur.Hours())
	}
	days := math.Floor(dur.Hours() / 24)
	if days == 1 {
		return "tomorrow"
	}
	if days == 2 {
		return "day after tomorrow"
	}
	if days < 30 {
		return fmt.Sprintf("in %.0f days", days)
	}
	months := math.Floor(days / 30)
	if months == 1 {
		return "next month"
	}
	if months < 12 {
		return fmt.Sprintf("in %.0f months", months)
	}

	years := math.Floor(months / 12)
	if years == 1 {
		return "next year"
	}

	return fmt.Sprintf("in %.0f years", years)

}
func pastNaturalTime(base, value time.Time) string {
	dur := base.Sub(value)
	if dur.Seconds() <= 60 {
		return "a few seconds ago"
	}
	if dur.Minutes() < 5 {
		return "a few minutes ago"
	}
	if dur.Minutes() < 60 {
		return fmt.Sprintf("%.0f minutes ago", dur.Minutes())
	}

	days := math.Floor(dur.Hours() / 24)
	startBase := time.Date(base.Year(), base.Month(), base.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := startBase.Add(-24 * time.Hour)
	dayBeforeYesterday := yesterday.Add(-24 * time.Hour)

	//fmt.Println(value, days, startBase, yesterday, dayBeforeYesterday)

	if value.After(startBase) {
		return fmt.Sprintf("%.0f hours ago", dur.Hours())
	}
	if value.After(yesterday) {
		return "yesterday"
	}
	if value.After(dayBeforeYesterday) {
		return "day before yesterday"
	}
	if days < 30 {
		return fmt.Sprintf("%.0f days ago", days)
	}
	months := math.Floor(days / 30)
	if months == 1 {
		return "last month"
	}
	if months < 12 {
		return fmt.Sprintf("%.0f months ago", months)
	}

	years := math.Floor(months / 12)
	if years == 1 {
		return "last year"
	}

	return fmt.Sprintf("%.0f years ago", years)
}
