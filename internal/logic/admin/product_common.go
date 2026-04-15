package adminlogic

import (
	"context"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
)

func normalizeSortAction(action string) (string, error) {
	action = strings.TrimSpace(strings.ToLower(action))
	switch action {
	case "top", "up", "down", "bottom":
		return action, nil
	default:
		return "", fmt.Errorf("排序动作错误")
	}
}

func moveIDByAction(ids []int64, targetID int64, action string) []int64 {
	index := indexOfID(ids, targetID)
	if index < 0 {
		return append([]int64(nil), ids...)
	}
	newIndex := index
	switch action {
	case "top":
		newIndex = 0
	case "up":
		if index > 0 {
			newIndex = index - 1
		}
	case "down":
		if index < len(ids)-1 {
			newIndex = index + 1
		}
	case "bottom":
		newIndex = len(ids) - 1
	}
	ordered := append([]int64(nil), ids...)
	if newIndex == index {
		return ordered
	}
	ordered = append(ordered[:index], ordered[index+1:]...)
	if newIndex >= len(ordered) {
		ordered = append(ordered, targetID)
		return ordered
	}
	ordered = append(ordered[:newIndex], append([]int64{targetID}, ordered[newIndex:]...)...)
	return ordered
}

func uniqueInt64s(ids []int64) ([]int64, error) {
	return uniquePositiveInt64s(ids, "品牌ID")
}

func uniquePositiveInt64s(ids []int64, fieldName string) ([]int64, error) {
	seen := make(map[int64]struct{}, len(ids))
	result := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id <= 0 {
			return nil, fmt.Errorf("%s必须是正整数", fieldName)
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result, nil
}

func indexOfID(ids []int64, targetID int64) int {
	for index, id := range ids {
		if id == targetID {
			return index
		}
	}
	return -1
}

func sqlPlaceholders(count int) string {
	if count <= 0 {
		return ""
	}
	return strings.TrimSuffix(strings.Repeat("?,", count), ",")
}

func countActiveGoodsReference(ctx context.Context, db gdb.DB, column string, id int64) (int, error) {
	value, err := db.GetCore().GetValue(ctx, `SELECT COUNT(*) FROM product_goods WHERE is_deleted = 0 AND `+column+` = ?`, id)
	if err != nil {
		return 0, err
	}
	return value.Int(), nil
}

func hasActiveGoodsReferences(ctx context.Context, db gdb.DB, column string, ids []int64) (bool, error) {
	if len(ids) == 0 {
		return false, nil
	}
	rows := make([]struct {
		ID int64 `db:"id"`
	}, 0)
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	if err := db.GetCore().GetScan(ctx, &rows, `SELECT DISTINCT `+column+` AS id FROM product_goods WHERE is_deleted = 0 AND `+column+` IN (`+sqlPlaceholders(len(ids))+`) LIMIT 1`, args...); err != nil {
		return false, err
	}
	return len(rows) > 0, nil
}
