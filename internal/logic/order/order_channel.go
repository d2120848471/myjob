package orderlogic

import (
	"context"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"time"

	"myjob/internal/consts"
	"myjob/internal/library/channelpricing"

	"github.com/shopspring/decimal"
)

type orderChannelCandidate struct {
	BindingID           int64  `db:"binding_id"`
	PlatformAccountID   int64  `db:"platform_account_id"`
	PlatformAccountName string `db:"platform_account_name"`
	PlatformSubjectID   int64  `db:"platform_subject_id"`
	PlatformSubjectName string `db:"platform_subject_name"`
	ProviderCode        string `db:"provider_code"`
	SupplierGoodsNo     string `db:"supplier_goods_no"`
	SupplierGoodsName   string `db:"supplier_goods_name"`
	SourceCostPrice     string `db:"source_cost_price"`
	CostPrice           string `db:"cost_price"`
	DefaultSellPrice    string `db:"default_sell_price"`
	IsAutoChange        int    `db:"is_auto_change"`
	AddType             string `db:"add_type"`
	DefaultPrice        string `db:"default_price"`
	Sort                int    `db:"sort"`
	OrderWeight         string `db:"order_weight"`
	OrderTimeStart      string `db:"order_time_start"`
	OrderTimeEnd        string `db:"order_time_end"`
}

func (l *OrderLogic) loadCandidateChannels(ctx context.Context, goodsID int64, attempted map[int64]struct{}) ([]orderChannelCandidate, error) {
	rows := make([]orderChannelCandidate, 0)
	if err := l.core.DB().GetCore().GetScan(ctx, &rows, `
SELECT
    b.id AS binding_id,
    a.id AS platform_account_id,
    a.name AS platform_account_name,
    a.subject_id AS platform_subject_id,
    COALESCE(s.name, '') AS platform_subject_name,
    a.provider_code,
    b.supplier_goods_no,
    b.supplier_goods_name,
    b.cost_price,
    COALESCE(g.default_sell_price, '0.0000') AS default_sell_price,
    b.is_auto_change,
    b.add_type,
    b.default_price,
    b.sort,
    b.order_weight,
    b.order_time_start,
    b.order_time_end
FROM product_goods_channel_binding b
JOIN supplier_platform_account a ON a.id = b.platform_account_id
JOIN product_goods g ON g.id = b.goods_id
LEFT JOIN admin_subject s ON s.id = a.subject_id
WHERE b.goods_id = ?
  AND g.status = 1
  AND g.is_deleted = 0
  AND g.goods_type = ?
  AND g.supply_type = ?
  AND b.is_deleted = 0
  AND b.dock_status = 1
  AND b.supplier_goods_no <> ''
  AND a.is_deleted = 0
  AND a.status = 1
  AND a.provider_code = 'kakayun'
ORDER BY b.sort ASC, b.id ASC
	`, goodsID, openOrderGoodsTypeDirectRecharge, openOrderSupplyTypeChannel); err != nil {
		return nil, apiErr(consts.CodeInternalError, "候选渠道查询失败")
	}
	filtered := make([]orderChannelCandidate, 0, len(rows))
	for _, row := range rows {
		if attempted != nil {
			if _, ok := attempted[row.BindingID]; ok {
				continue
			}
		}
		filtered = append(filtered, row)
	}
	return filtered, nil
}

func (c orderChannelCandidate) pricingRule() channelpricing.Rule {
	return channelpricing.Rule{
		DefaultSellPrice: c.DefaultSellPrice,
		CostPrice:        c.CostPrice,
		IsAutoChange:     c.IsAutoChange,
		AddType:          c.AddType,
		ProfitValue:      c.DefaultPrice,
	}
}

func selectCandidate(candidates []orderChannelCandidate, attempted map[int64]struct{}, strategy string, now time.Time) orderChannelCandidate {
	filtered := make([]orderChannelCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		if attempted != nil {
			if _, ok := attempted[candidate.BindingID]; ok {
				continue
			}
		}
		if strategy == "time_window" && !candidate.matchesTimeWindow(now) {
			continue
		}
		filtered = append(filtered, candidate)
	}
	if len(filtered) == 0 {
		return orderChannelCandidate{}
	}
	switch strategy {
	case "lowest_cost":
		sort.SliceStable(filtered, func(i, j int) bool {
			left, _ := decimal.NewFromString(filtered[i].CostPrice)
			right, _ := decimal.NewFromString(filtered[j].CostPrice)
			if left.Equal(right) {
				return filtered[i].BindingID < filtered[j].BindingID
			}
			return left.LessThan(right)
		})
	case "random":
		return filtered[rand.IntN(len(filtered))]
	case "weighted_percent":
		return selectWeightedCandidate(filtered)
	default:
		sort.SliceStable(filtered, func(i, j int) bool {
			if filtered[i].Sort == filtered[j].Sort {
				return filtered[i].BindingID < filtered[j].BindingID
			}
			return filtered[i].Sort < filtered[j].Sort
		})
	}
	return filtered[0]
}

// selectWeightedCandidate 按库存配置的百分比权重抽样，权重为 0 的渠道不会被选中。
func selectWeightedCandidate(candidates []orderChannelCandidate) orderChannelCandidate {
	weights := make([]int, len(candidates))
	total := 0
	for index, candidate := range candidates {
		weight := orderWeightUnit(candidate.OrderWeight)
		if weight <= 0 {
			continue
		}
		weights[index] = weight
		total += weight
	}
	if total <= 0 {
		return orderChannelCandidate{}
	}
	target := rand.IntN(total)
	current := 0
	for index, weight := range weights {
		if weight <= 0 {
			continue
		}
		current += weight
		if target < current {
			return candidates[index]
		}
	}
	return orderChannelCandidate{}
}

func orderWeightUnit(value string) int {
	weight, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil || !weight.GreaterThan(decimal.Zero) {
		return 0
	}
	return int(weight.Mul(decimal.NewFromInt(10000)).Round(0).IntPart())
}

func (c orderChannelCandidate) matchesTimeWindow(now time.Time) bool {
	start, startOK := parseHHMM(c.OrderTimeStart)
	end, endOK := parseHHMM(c.OrderTimeEnd)
	if strings.TrimSpace(c.OrderTimeStart) == "" || strings.TrimSpace(c.OrderTimeEnd) == "" {
		return true
	}
	if !startOK || !endOK {
		return false
	}
	current := now.Hour()*60 + now.Minute()
	if start <= end {
		return current >= start && current <= end
	}
	return current >= start || current <= end
}

func parseHHMM(value string) (int, bool) {
	parts := strings.Split(strings.TrimSpace(value), ":")
	if len(parts) != 2 {
		return 0, false
	}
	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return 0, false
	}
	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return 0, false
	}
	return hour*60 + minute, true
}
