package app

import (
	"strings"
	"time"
)

// ParsePagination 将分页参数归一化到可用范围内。
//
// - page <= 0 视为 1
// - pageSize <= 0 视为 20
// - pageSize 最大为 100
func ParsePagination(page, pageSize int) (int, int) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return page, pageSize
}

// AppendTimeRangeFilters 根据 startTime/endTime 追加 created_at 的筛选条件与参数。
//
// 入参时间格式要求为 "2006-01-02 15:04:05"（与 ParseQueryTime 一致）。
func AppendTimeRangeFilters(startTime, endTime string, conditions *[]string, args *[]any) error {
	if strings.TrimSpace(startTime) != "" {
		parsed, err := ParseQueryTime(startTime)
		if err != nil {
			return err
		}
		*conditions = append(*conditions, "created_at >= ?")
		*args = append(*args, parsed)
	}
	if strings.TrimSpace(endTime) != "" {
		parsed, err := ParseQueryTime(endTime)
		if err != nil {
			return err
		}
		*conditions = append(*conditions, "created_at <= ?")
		*args = append(*args, parsed)
	}
	return nil
}

// ParseQueryTime 解析查询参数中的时间字符串（本地时区）。
func ParseQueryTime(value string) (time.Time, error) {
	return time.ParseInLocation("2006-01-02 15:04:05", strings.TrimSpace(value), time.Local)
}

// ContainsString 判断切片 items 是否包含目标字符串。
func ContainsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}
