package orderlogic

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"strings"

	adminapi "myjob/api"
	"myjob/internal/consts"

	"github.com/shopspring/decimal"
)

const (
	openOrderGoodsTypeDirectRecharge = "direct_recharge"
	openOrderSupplyTypeChannel       = "channel"
	maxOpenOrderCreateAttempts       = 3
)

type openOrderGoods struct {
	ID               int64         `db:"id"`
	GoodsCode        string        `db:"goods_code"`
	Name             string        `db:"name"`
	GoodsType        string        `db:"goods_type"`
	SupplyType       string        `db:"supply_type"`
	SubjectID        sql.NullInt64 `db:"subject_id"`
	SubjectName      string        `db:"subject_name"`
	HasTax           int           `db:"has_tax"`
	DefaultSellPrice string        `db:"default_sell_price"`
	MinPurchaseQty   int           `db:"min_purchase_qty"`
	MaxPurchaseQty   int           `db:"max_purchase_qty"`
}

// CreateOpenOrder 校验开放 token、商品和候选云发卡渠道后创建待提交订单。
func (l *OrderLogic) CreateOpenOrder(ctx context.Context, req *adminapi.OpenOrderCreateReq) (*adminapi.OpenOrderCreateRes, error) {
	if strings.TrimSpace(req.Token) != strings.TrimSpace(l.core.Config().OpenOrder.Token) {
		return nil, unauthorizedErr()
	}
	goods, err := l.loadOpenOrderGoods(ctx, strings.TrimSpace(req.GoodsID))
	if err != nil {
		return nil, err
	}
	account := strings.TrimSpace(req.Account)
	if account == "" {
		return nil, apiErr(consts.CodeBadRequest, "充值账号不能为空")
	}
	if req.Quantity <= 0 {
		return nil, apiErr(consts.CodeBadRequest, "购买数量必须大于0")
	}
	if req.Quantity < goods.MinPurchaseQty || req.Quantity > goods.MaxPurchaseQty {
		return nil, apiErr(consts.CodeBadRequest, "购买数量不在允许范围内")
	}
	candidates, err := l.loadCandidateChannels(ctx, goods.ID, nil)
	if err != nil {
		return nil, err
	}
	if len(candidates) == 0 {
		return nil, apiErr(consts.CodeBadRequest, "暂无可用云发卡渠道")
	}
	now := l.core.Now()
	unitPrice, err := normalizeOrderMoney(goods.DefaultSellPrice)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "商品售价格式错误")
	}
	orderAmount, err := multiplyOrderMoney(unitPrice, req.Quantity)
	if err != nil {
		return nil, apiErr(consts.CodeBadRequest, "订单金额计算失败")
	}
	orderNo := ""
	for attempt := 0; attempt < maxOpenOrderCreateAttempts; attempt++ {
		orderNo = l.nextOrderNo()
		if _, err = l.core.DB().Exec(ctx, `
INSERT INTO external_order (
    order_no, goods_id, goods_code, goods_name, goods_type, supply_type, subject_id, subject_name,
    has_tax, account, quantity, unit_price, order_amount, cost_amount, profit_amount,
    status, attempt_count, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, '0.0000', ?, ?, 0, ?, ?)
`, orderNo, goods.ID, goods.GoodsCode, goods.Name, goods.GoodsType, goods.SupplyType, nullableInt64Arg(goods.SubjectID), goods.SubjectName, goods.HasTax, account, req.Quantity, unitPrice, orderAmount, orderAmount, OrderStatusPendingSubmit, now, now); err != nil {
			if isOrderNoUniqueConflict(err) {
				continue
			}
			return nil, apiErr(consts.CodeInternalError, "订单创建失败")
		}
		err = nil
		break
	}
	if err != nil {
		return nil, apiErr(consts.CodeInternalError, "订单创建失败")
	}
	l.TriggerSubmit(orderNo)
	return &adminapi.OpenOrderCreateRes{
		OrderNo:    orderNo,
		StatusCode: OrderStatusPendingSubmit,
		StatusText: orderStatusText(OrderStatusPendingSubmit),
		CreatedAt:  formatAppTime(now),
	}, nil
}

func (l *OrderLogic) nextOrderNo() string {
	if l.orderNoGenerator != nil {
		return l.orderNoGenerator()
	}
	return l.generateOrderNo()
}

func (l *OrderLogic) loadOpenOrderGoods(ctx context.Context, goodsCode string) (openOrderGoods, error) {
	if goodsCode == "" {
		return openOrderGoods{}, apiErr(consts.CodeBadRequest, "商品ID不能为空")
	}
	row := openOrderGoods{}
	if err := l.core.DB().GetCore().GetScan(ctx, &row, `
SELECT
    g.id,
    g.goods_code,
    g.name,
    g.goods_type,
    g.supply_type,
    g.subject_id,
    COALESCE(s.name, '') AS subject_name,
    g.has_tax,
    COALESCE(g.default_sell_price, '0.0000') AS default_sell_price,
    g.min_purchase_qty,
    g.max_purchase_qty
FROM product_goods g
LEFT JOIN admin_subject s ON s.id = g.subject_id
WHERE g.goods_code = ? AND g.status = 1 AND g.is_deleted = 0
`, goodsCode); err != nil {
		return openOrderGoods{}, apiErr(consts.CodeBadRequest, "商品不存在")
	}
	if row.GoodsType != openOrderGoodsTypeDirectRecharge {
		return openOrderGoods{}, apiErr(consts.CodeBadRequest, "仅支持直充商品")
	}
	if row.SupplyType != openOrderSupplyTypeChannel {
		return openOrderGoods{}, apiErr(consts.CodeBadRequest, "仅支持渠道供货商品")
	}
	return row, nil
}

func (l *OrderLogic) generateOrderNo() string {
	randomNumber := int64(0)
	if value, err := rand.Int(rand.Reader, big.NewInt(1000000)); err == nil {
		randomNumber = value.Int64()
	}
	return fmt.Sprintf("O%s%06d", l.core.Now().Format("20060102150405"), randomNumber)
}

func nullableInt64Arg(value sql.NullInt64) any {
	if !value.Valid {
		return nil
	}
	return value.Int64
}

func isOrderNoUniqueConflict(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "unique")
}

func normalizeOrderMoney(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		value = "0"
	}
	amount, err := decimal.NewFromString(value)
	if err != nil {
		return "", err
	}
	return amount.Round(4).StringFixed(4), nil
}

func multiplyOrderMoney(unitPrice string, quantity int) (string, error) {
	amount, err := decimal.NewFromString(unitPrice)
	if err != nil {
		return "", err
	}
	return amount.Mul(decimal.NewFromInt(int64(quantity))).Round(4).StringFixed(4), nil
}
