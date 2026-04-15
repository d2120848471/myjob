package adminlogic

import (
	"database/sql"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
)

func nullableTimeFromRecord(row gdb.Record, key string) sql.NullTime {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return sql.NullTime{}
	}
	parsed, ok := appTimeFromValue(value.Val())
	if !ok {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: parsed, Valid: true}
}

func parseRecordTime(row gdb.Record, key string) time.Time {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return time.Time{}
	}
	parsed, ok := appTimeFromValue(value.Val())
	if !ok {
		return time.Time{}
	}
	return parsed
}

func appTimeFromValue(raw any) (time.Time, bool) {
	switch value := raw.(type) {
	case time.Time:
		if value.IsZero() {
			return time.Time{}, false
		}
		return value, true
	case string:
		return parseRecordTimeString(value)
	case []byte:
		return parseRecordTimeString(string(value))
	case interface{ String() string }:
		return parseRecordTimeString(value.String())
	default:
		return time.Time{}, false
	}
}

func parseRecordTimeString(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05.999999999-07:00", "2006-01-02 15:04:05 -0700 MST", "2006-01-02 15:04:05.999999999", "2006-01-02 15:04:05"}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, value, time.Local); err == nil {
			return parsed, true
		}
	}
	return time.Time{}, false
}

func formatNullableTime(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return value.Time.Format("2006-01-02 15:04:05")
}
