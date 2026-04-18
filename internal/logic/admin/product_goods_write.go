package adminlogic

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"time"

	adminapi "myjob/api"
	"myjob/internal/consts"
	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
)

// Add 新增商品，并写入操作日志。
func (l *ProductGoodsLogic) Add(ctx context.Context, req *adminapi.ProductGoodsCreateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsCreateRes, error) {
	normalized, err := l.normalizeProductGoodsInput(ctx, req.BrandID, req.Name, req.GoodsType, req.SupplyType, req.IsExport, req.IsDouyin, req.HasTax, req.SubjectID, req.ExceptionNotify, req.ProductTemplateID, req.PurchaseLimitStrategyID, req.PurchaseNotice, req.TerminalPriceLimit, req.BalanceLimit, req.DefaultSellPrice, req.MinPurchaseQty, req.MaxPurchaseQty, req.Status, nil)
	if err != nil {
		return nil, err
	}

	createdID := int64(0)
	createdCode := ""
	now := l.core.Now()
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		tempGoodsCode := temporaryProductGoodsCode(now)
		result, txErr := tx.Exec(`
INSERT INTO product_goods (
    goods_code, brand_id, name, goods_type, supply_type, is_export, is_douyin, has_tax, subject_id, exception_notify,
    product_template_id, purchase_limit_strategy_id, purchase_notice, terminal_price_limit, balance_limit,
    default_sell_price, min_purchase_qty, max_purchase_qty, status, is_deleted, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?)
`, tempGoodsCode, normalized.BrandID, normalized.Name, normalized.GoodsType, normalized.SupplyType, normalized.IsExport, normalized.IsDouyin, normalized.HasTax, nullableInt64Arg(normalized.SubjectID), normalized.ExceptionNotify, nullableInt64Arg(normalized.ProductTemplateID), nullableInt64Arg(normalized.PurchaseLimitStrategyID), nullableStringArg(normalized.PurchaseNotice), nullableMoneyArg(normalized.TerminalPriceLimit), normalized.BalanceLimit, nullableMoneyArg(normalized.DefaultSellPrice), normalized.MinPurchaseQty, normalized.MaxPurchaseQty, normalized.Status, now, now)
		if txErr != nil {
			return txErr
		}
		id, _ := result.LastInsertId()
		goodsCode := buildGoodsCode(id)
		if _, txErr = tx.Exec(`UPDATE product_goods SET goods_code = ?, updated_at = ? WHERE id = ?`, goodsCode, now, id); txErr != nil {
			return txErr
		}
		if txErr = adjustBrandGoodsCountTx(tx, normalized.BrandID, 1, now); txErr != nil {
			return txErr
		}
		createdID = id
		createdCode = goodsCode
		return nil
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品新增失败")
	}

	l.core.WriteOperation(ctx, actor, fmt.Sprintf("新增商品：%s", normalized.Name), ip)
	return &adminapi.ProductGoodsCreateRes{ID: createdID, GoodsCode: createdCode}, nil
}

// Edit 编辑商品信息，并写入操作日志。
func (l *ProductGoodsLogic) Edit(ctx context.Context, req *adminapi.ProductGoodsUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsUpdateRes, error) {
	current, err := l.getActiveProduct(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品不存在")
	}

	currentStrategyID := nullableInt64Pointer(current.PurchaseLimitStrategyID)
	normalized, err := l.normalizeProductGoodsInput(ctx, req.BrandID, req.Name, req.GoodsType, req.SupplyType, req.IsExport, req.IsDouyin, req.HasTax, req.SubjectID, req.ExceptionNotify, req.ProductTemplateID, req.PurchaseLimitStrategyID, req.PurchaseNotice, req.TerminalPriceLimit, req.BalanceLimit, req.DefaultSellPrice, req.MinPurchaseQty, req.MaxPurchaseQty, req.Status, currentStrategyID)
	if err != nil {
		return nil, err
	}

	now := l.core.Now()
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 先确认商品仍是未删除状态，再处理品牌计数，避免并发软删把 goods_count 改漂。
		result, txErr := tx.Exec(`
UPDATE product_goods
SET brand_id = ?, name = ?, goods_type = ?, supply_type = ?, is_export = ?, is_douyin = ?, has_tax = ?, subject_id = ?, exception_notify = ?,
    product_template_id = ?, purchase_limit_strategy_id = ?, purchase_notice = ?, terminal_price_limit = ?, balance_limit = ?,
    default_sell_price = ?, min_purchase_qty = ?, max_purchase_qty = ?, status = ?, updated_at = ?
WHERE id = ? AND is_deleted = 0
`, normalized.BrandID, normalized.Name, normalized.GoodsType, normalized.SupplyType, normalized.IsExport, normalized.IsDouyin, normalized.HasTax, nullableInt64Arg(normalized.SubjectID), normalized.ExceptionNotify, nullableInt64Arg(normalized.ProductTemplateID), nullableInt64Arg(normalized.PurchaseLimitStrategyID), nullableStringArg(normalized.PurchaseNotice), nullableMoneyArg(normalized.TerminalPriceLimit), normalized.BalanceLimit, nullableMoneyArg(normalized.DefaultSellPrice), normalized.MinPurchaseQty, normalized.MaxPurchaseQty, normalized.Status, now, req.ID)
		if txErr != nil {
			return txErr
		}
		if txErr = ensureMutationAffected(result); txErr != nil {
			return txErr
		}
		if normalized.BrandID != current.BrandID {
			if txErr = adjustBrandGoodsCountTx(tx, current.BrandID, -1, now); txErr != nil {
				return txErr
			}
			if txErr = adjustBrandGoodsCountTx(tx, normalized.BrandID, 1, now); txErr != nil {
				return txErr
			}
		}
		return nil
	}); err != nil {
		if err == sql.ErrNoRows {
			return nil, apiErr(consts.CodeBadRequest, "商品不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "商品编辑失败")
	}

	l.core.WriteOperation(ctx, actor, fmt.Sprintf("编辑商品：%d -> %s", req.ID, normalized.Name), ip)
	return &adminapi.ProductGoodsUpdateRes{}, nil
}

// Delete 软删除商品（is_deleted=1），并在事务内同步品牌商品计数。
func (l *ProductGoodsLogic) Delete(ctx context.Context, req *adminapi.ProductGoodsDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsDeleteRes, error) {
	current, err := l.getActiveProduct(ctx, req.ID)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品不存在")
	}

	now := l.core.Now()
	if err = l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// 删除先校验命中行数，避免重复删除时继续扣减品牌计数。
		result, txErr := tx.Exec(`UPDATE product_goods SET is_deleted = 1, deleted_at = ?, updated_at = ? WHERE id = ? AND is_deleted = 0`, now, now, req.ID)
		if txErr != nil {
			return txErr
		}
		if txErr = ensureMutationAffected(result); txErr != nil {
			return txErr
		}
		return adjustBrandGoodsCountTx(tx, current.BrandID, -1, now)
	}); err != nil {
		if err == sql.ErrNoRows {
			return nil, apiErr(consts.CodeBadRequest, "商品不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "商品删除失败")
	}

	l.core.WriteOperation(ctx, actor, fmt.Sprintf("删除商品：%d -> %s", req.ID, current.Name), ip)
	return &adminapi.ProductGoodsDeleteRes{}, nil
}

// Status 批量切换商品启用/禁用状态，并返回失败项原因列表。
func (l *ProductGoodsLogic) Status(ctx context.Context, req *adminapi.ProductGoodsStatusReq, actor entity.AdminUser, ip string) (*adminapi.ProductGoodsStatusRes, error) {
	if len(req.IDs) == 0 {
		return nil, apiErr(consts.CodeBadRequest, "请至少选择一个商品")
	}
	ids, err := uniquePositiveInt64s(req.IDs, "商品ID")
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, err.Error())
	}
	if req.Status != consts.StatusEnabled && req.Status != consts.StatusDisabled {
		return nil, apiErr(consts.CodeBadRequest, "状态错误")
	}

	successIDs := make([]int64, 0, len(ids))
	failed := make([]adminapi.ProductGoodsStatusFailedItem, 0)
	now := l.core.Now()
	if err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		existingIDs, txErr := loadActiveProductIDSetTx(tx, ids, l.core.Config().Database.Driver)
		if txErr != nil {
			return txErr
		}
		for _, id := range ids {
			if _, ok := existingIDs[id]; ok {
				successIDs = append(successIDs, id)
				continue
			}
			failed = append(failed, adminapi.ProductGoodsStatusFailedItem{
				ID:     id,
				Reason: "商品不存在",
			})
		}
		if len(successIDs) == 0 {
			return nil
		}

		args := make([]any, 0, len(successIDs)+2)
		args = append(args, req.Status, now)
		for _, id := range successIDs {
			args = append(args, id)
		}
		// 事务内先锁定、再更新，避免并发软删把未命中的商品误算进 success_ids。
		if _, txErr = tx.Exec(`
UPDATE product_goods
SET status = ?, updated_at = ?
WHERE is_deleted = 0 AND id IN (`+sqlPlaceholders(len(successIDs))+`)
`, args...); txErr != nil {
			return txErr
		}
		return nil
	}); err != nil {
		return nil, apiErr(consts.CodeInternalError, "商品状态更新失败")
	}

	l.core.WriteOperation(ctx, actor, fmt.Sprintf("批量修改商品状态：status=%d, total=%d, success=%d, failed=%d", req.Status, len(ids), len(successIDs), len(failed)), ip)
	return &adminapi.ProductGoodsStatusRes{
		SuccessIDs:   successIDs,
		SuccessCount: len(successIDs),
		FailedCount:  len(failed),
		Failed:       failed,
	}, nil
}

func loadActiveProductIDSetTx(tx gdb.TX, ids []int64, driver string) (map[int64]struct{}, error) {
	rows := make([]struct {
		ID int64 `db:"id"`
	}, 0, len(ids))
	args := make([]any, 0, len(ids))
	for _, id := range ids {
		args = append(args, id)
	}
	if err := tx.GetScan(&rows, productGoodsStatusSelectSQL(driver, len(ids)), args...); err != nil {
		return nil, err
	}
	result := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		result[row.ID] = struct{}{}
	}
	return result, nil
}

func productGoodsStatusSelectSQL(driver string, idCount int) string {
	query := `SELECT id FROM product_goods WHERE is_deleted = 0 AND id IN (` + sqlPlaceholders(idCount) + `)`
	if strings.EqualFold(strings.TrimSpace(driver), "mysql") {
		return query + ` FOR UPDATE`
	}
	return query
}

func buildGoodsCode(id int64) string {
	return strconv.FormatInt(id, 10)
}

func temporaryProductGoodsCode(now time.Time) string {
	return fmt.Sprintf("TMP%d", now.UnixNano())
}

func adjustBrandGoodsCountTx(tx gdb.TX, brandID int64, delta int, now time.Time) error {
	if delta == 0 {
		return nil
	}
	if _, err := tx.Exec(`
UPDATE product_brand
SET goods_count = CASE WHEN goods_count + ? < 0 THEN 0 ELSE goods_count + ? END,
    updated_at = ?
WHERE id = ?
`, delta, delta, now, brandID); err != nil {
		return err
	}
	return nil
}

func ensureMutationAffected(result sql.Result) error {
	if result == nil {
		return sql.ErrNoRows
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}
