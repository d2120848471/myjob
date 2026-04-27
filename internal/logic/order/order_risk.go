package orderlogic

import (
	"context"
	"errors"
	"strings"
	"time"

	adminapi "myjob/api"
	runtimeapp "myjob/internal/app"

	"github.com/gogf/gf/v2/database/gdb"
)

const (
	rechargeRiskTokenSnapshotMaxRunes = 64
	externalOrderLastReceiptMaxRunes  = 512
)

var (
	errOrderNoRetriesExhausted    = errors.New("订单号生成重试失败")
	errRechargeRiskRuleNoLongerOn = errors.New("充值风控规则已失效")
)

type rechargeRiskMatch struct {
	RuleID       int64  `db:"rule_id"`
	GoodsKeyword string `db:"goods_keyword"`
	Reason       string `db:"reason"`
}

// matchRechargeRisk 使用账号精确匹配、商品名关键词包含匹配，避免误拦截同账号下其他商品。
func (l *OrderLogic) matchRechargeRisk(ctx context.Context, account, goodsName string) (rechargeRiskMatch, bool, error) {
	rows := make([]rechargeRiskMatch, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT id AS rule_id, goods_keyword, reason
FROM recharge_risk_rule
WHERE account = ? AND status = 1 AND is_deleted = 0
ORDER BY id ASC
`, account); err != nil {
		return rechargeRiskMatch{}, false, err
	}
	for _, row := range rows {
		if strings.TrimSpace(row.GoodsKeyword) != "" && strings.Contains(goodsName, row.GoodsKeyword) {
			return row, true, nil
		}
	}
	return rechargeRiskMatch{}, false, nil
}

// createRiskFailedOpenOrder 原子复核规则并写入失败订单、拦截流水和命中次数，确保本地拦截结果可追溯。
func (l *OrderLogic) createRiskFailedOpenOrder(ctx context.Context, req *adminapi.OpenOrderCreateReq, goods openOrderGoods, account, unitPrice, orderAmount string, match rechargeRiskMatch, now time.Time) (string, bool, error) {
	orderNo := ""
	for attempt := 0; attempt < maxOpenOrderCreateAttempts; attempt++ {
		orderNo = l.nextOrderNo()
		err := l.core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
			// 命中查询发生在事务外，写入前必须再次确认规则仍启用，避免管理员刚停用/删除后仍被拦截。
			updateResult, err := tx.Exec(`UPDATE recharge_risk_rule SET hit_count = hit_count + 1, updated_at = ? WHERE id = ? AND status = 1 AND is_deleted = 0`, now, match.RuleID)
			if err != nil {
				return err
			}
			affected, err := updateResult.RowsAffected()
			if err != nil {
				return err
			}
			if affected == 0 {
				return errRechargeRiskRuleNoLongerOn
			}
			result, err := tx.Exec(`
INSERT INTO external_order (
    order_no, goods_id, goods_code, goods_name, goods_type, supply_type, subject_id, subject_name,
    has_tax, account, quantity, unit_price, order_amount, cost_amount, profit_amount,
    status, attempt_count, last_receipt, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '0.0000', ?, ?, 0, ?, ?, ?)
`, orderNo, goods.ID, goods.GoodsCode, goods.Name, goods.GoodsType, goods.SupplyType, nullableInt64Arg(goods.SubjectID), goods.SubjectName, goods.HasTax,
				account, req.Quantity, unitPrice, orderAmount, orderAmount, OrderStatusFailed, riskOrderReceipt(match.Reason), now, now)
			if err != nil {
				return err
			}
			orderID, err := result.LastInsertId()
			if err != nil {
				return err
			}
			if _, err = tx.Exec(`
INSERT INTO recharge_risk_record (
    rule_id, order_id, order_no, account, goods_id, goods_code, goods_name,
    matched_keyword, reason, request_token_masked, intercepted_at, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, match.RuleID, orderID, orderNo, account, goods.ID, goods.GoodsCode, goods.Name, match.GoodsKeyword, match.Reason, riskRecordTokenSnapshot(req.Token), now, now); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			if errors.Is(err, errRechargeRiskRuleNoLongerOn) {
				return "", false, nil
			}
			if isOrderNoUniqueConflict(err) {
				continue
			}
			return "", false, err
		}
		return orderNo, true, nil
	}
	return "", false, errOrderNoRetriesExhausted
}

func riskOrderReceipt(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "充值账号命中风控"
	}
	return limitRunes("充值账号命中风控："+reason, externalOrderLastReceiptMaxRunes)
}

func riskRecordTokenSnapshot(token string) string {
	token = strings.TrimSpace(token)
	if len([]rune(token)) <= rechargeRiskTokenSnapshotMaxRunes {
		return runtimeapp.MaskSecret(token)
	}
	runes := []rune(token)
	prefixLen := 4
	suffixLen := 4
	maskLen := rechargeRiskTokenSnapshotMaxRunes - prefixLen - suffixLen
	return string(runes[:prefixLen]) + strings.Repeat("*", maskLen) + string(runes[len(runes)-suffixLen:])
}

func limitRunes(value string, maxRunes int) string {
	if maxRunes <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	return string(runes[:maxRunes])
}
