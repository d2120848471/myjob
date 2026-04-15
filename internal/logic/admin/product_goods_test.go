package adminlogic

import (
	"database/sql"
	"testing"
)

type staticSQLResult struct {
	rows int64
}

func (r staticSQLResult) LastInsertId() (int64, error) {
	return 0, nil
}

func (r staticSQLResult) RowsAffected() (int64, error) {
	return r.rows, nil
}

func TestEnsureMutationAffected(t *testing.T) {
	t.Run("命中 1 行时通过", func(t *testing.T) {
		if err := ensureMutationAffected(staticSQLResult{rows: 1}); err != nil {
			t.Fatalf("expected success, got error: %v", err)
		}
	})

	t.Run("命中 0 行时返回未找到", func(t *testing.T) {
		err := ensureMutationAffected(staticSQLResult{rows: 0})
		if err != sql.ErrNoRows {
			t.Fatalf("expected sql.ErrNoRows, got: %v", err)
		}
	})
}

func TestProductGoodsStatusSelectSQL(t *testing.T) {
	t.Run("mysql 需要锁定目标商品行", func(t *testing.T) {
		sqlText := productGoodsStatusSelectSQL("mysql", 2)
		expected := "SELECT id FROM product_goods WHERE is_deleted = 0 AND id IN (?,?) FOR UPDATE"
		if sqlText != expected {
			t.Fatalf("expected %q, got %q", expected, sqlText)
		}
	})

	t.Run("sqlite 不追加 for update", func(t *testing.T) {
		sqlText := productGoodsStatusSelectSQL("sqlite", 1)
		expected := "SELECT id FROM product_goods WHERE is_deleted = 0 AND id IN (?)"
		if sqlText != expected {
			t.Fatalf("expected %q, got %q", expected, sqlText)
		}
	})
}
