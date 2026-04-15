package adminlogic

import (
	"database/sql"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"
)

func nullableInt64(value sql.NullInt64) int64 {
	if !value.Valid {
		return 0
	}
	return value.Int64
}

func nullableInt64Pointer(value sql.NullInt64) *int64 {
	if !value.Valid {
		return nil
	}
	id := value.Int64
	return &id
}

func nullableString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nullableMoney(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return formatMoney(value.String)
}

func productGoodsRecordString(row gdb.Record, key string) string {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return ""
	}
	return value.String()
}

func productGoodsRecordInt64(row gdb.Record, key string) int64 {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return 0
	}
	return value.Int64()
}

func productGoodsRecordMoney(row gdb.Record, key string) string {
	value := productGoodsRecordString(row, key)
	if value == "" {
		return ""
	}
	return formatMoney(value)
}

func productGoodsRecordNullString(row gdb.Record, key string) sql.NullString {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return sql.NullString{}
	}
	return sql.NullString{String: value.String(), Valid: true}
}

func productGoodsRecordNullInt64(row gdb.Record, key string) sql.NullInt64 {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: value.Int64(), Valid: true}
}

func nullableStringArg(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableMoneyArg(value string) any {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return value
}

func nullableInt64Arg(value *int64) any {
	if value == nil {
		return nil
	}
	return *value
}

func formatAppTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}
