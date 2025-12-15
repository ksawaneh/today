// Package reports provides daily and weekly report generation for the today app.
package reports

import (
	"encoding/json"
)

// FormatDailyJSON formats a daily report as JSON.
func FormatDailyJSON(report *DailyReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}

// FormatWeeklyJSON formats a weekly report as JSON.
func FormatWeeklyJSON(report *WeeklyReport) ([]byte, error) {
	return json.MarshalIndent(report, "", "  ")
}
