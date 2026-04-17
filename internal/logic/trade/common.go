package tradelogic

import (
	"time"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/shopspring/decimal"
)

func apiErr(code gcode.Code, message string) error {
	return gerror.NewCode(code, message)
}

const (
	// RouteModeFixedOrder 按 sort asc, id asc 选路。
	RouteModeFixedOrder = "fixed_order"
	// RouteModeLowestCostFirst 按 cost_price asc, sort asc, id asc 选路。
	RouteModeLowestCostFirst = "lowest_cost_first"
	// RouteModeWeightPercent 按 weight 随机抽取（仅保留 weight>0）。
	RouteModeWeightPercent = "weight_percent"
	// RouteModeTimePeriod 仅保留命中 start_time~end_time 的绑定后再排序。
	RouteModeTimePeriod = "time_period"
	// RouteModeRandom 从可用绑定中随机抽取。
	RouteModeRandom = "random"
)

const (
	DockStatusEnabled  = "enabled"
	DockStatusDisabled = "disabled"
)

const (
	AddTypeFixed   = "fixed"
	AddTypePercent = "percent"
)

// CreateTradeOrderInput 表示创建交易订单前的核心输入（open API 或内部调用统一对齐）。
//
// 注意：该输入仅用于交易域内部逻辑与后续写库，不直接作为对外协议。
type CreateTradeOrderInput struct {
	CallerID      int64     // 调用方 ID（对外调用方主数据）
	ClientOrderNo string    // 调用方侧幂等单号
	GoodsCode     string    // 商品编码
	Quantity      int       // 购买数量
	PayloadJSON   string    // 调用方传入的 payload（原样 JSON 字符串）
	RequestIP     string    // 请求 IP（审计用）
	RequestedAt   time.Time // 请求时间（审计/幂等辅助用）
}

// CandidateBinding 表示用于交易选路的“可用绑定”快照。
type CandidateBinding struct {
	ID                int64
	GoodsID           int64
	PlatformAccountID int64
	ProviderCode      string
	ProviderName      string

	SupplierGoodsNo   string
	SupplierGoodsName string

	DockStatus string
	Sort       int
	Weight     int
	StartTime  string
	EndTime    string

	ValidateTemplateID *int64

	SourceCostPrice    decimal.Decimal
	CostPrice          decimal.Decimal
	TaxAdjustDirection string
	TaxAdjustRate      decimal.Decimal
	TaxAdjustAmount    decimal.Decimal

	IsAutoChange bool
	AddType      string
	DefaultPrice decimal.Decimal
}

// FulfillmentPlanItem 表示数量拆单后的单个履约分片规划。
type FulfillmentPlanItem struct {
	FulfillmentNo   string
	AttemptQuantity int
}
