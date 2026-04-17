package tradelogic

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/mail"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"myjob/internal/app"
	"myjob/internal/consts"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/shopspring/decimal"
)

type sqlRunner interface {
	Exec(sql string, args ...any) (sql.Result, error)
	GetOne(sql string, args ...any) (gdb.Record, error)
	GetScan(pointer any, sql string, args ...any) error
	GetValue(sql string, args ...any) (gdb.Value, error)
}

type ctxDBRunner struct {
	ctx context.Context
	db  gdb.DB
}

func (r ctxDBRunner) Exec(sql string, args ...any) (sql.Result, error) {
	return r.db.Exec(r.ctx, sql, args...)
}

func (r ctxDBRunner) GetOne(sql string, args ...any) (gdb.Record, error) {
	return r.db.GetOne(r.ctx, sql, args...)
}

func (r ctxDBRunner) GetScan(pointer any, sql string, args ...any) error {
	return r.db.GetScan(r.ctx, pointer, sql, args...)
}

func (r ctxDBRunner) GetValue(sql string, args ...any) (gdb.Value, error) {
	return r.db.GetValue(r.ctx, sql, args...)
}

// GoodsSnapshot 表示交易域内部使用的商品快照。
type GoodsSnapshot struct {
	ID                int64
	GoodsCode         string
	GoodsName         string
	SupplyType        string
	HasTax            int
	SubjectID         int64
	SubjectName       string
	DefaultSellPrice  decimal.Decimal
	MinPurchaseQty    int
	MaxPurchaseQty    int
	ProductTemplateID *int64
}

// GoodsChannelConfigSnapshot 表示交易域内部使用的商品级渠道配置快照。
type GoodsChannelConfigSnapshot struct {
	GoodsID               int64
	RouteMode             string
	SmartReplenishEnabled bool
	AttemptTimeoutEnabled bool
	AttemptTimeoutMinutes int
	AllowLoss             bool
	MaxLossAmount         *decimal.Decimal
	SyncCostEnabled       bool
	SyncGoodsNameEnabled  bool
	IsBundle              bool
}

// CandidateBuildOutput 表示候选绑定构建结果。
type CandidateBuildOutput struct {
	Goods      GoodsSnapshot
	Config     GoodsChannelConfigSnapshot
	Candidates []CandidateBinding
}

// BuildCandidateBindings 按“候选绑定生成”规则，读取商品/配置/绑定并做必要过滤与校验。
//
// 注意：
// - 本函数不接真实上游，不做建单。
// - 失败时返回业务错误（带 gcode），由上层统一包装成 {code,message,data}。
func BuildCandidateBindings(ctx context.Context, core *app.Core, goodsCode string, quantity int, payloadJSON string) (CandidateBuildOutput, error) {
	return BuildCandidateBindingsWithDB(ctx, core.DB(), core.Now(), goodsCode, quantity, payloadJSON)
}

// BuildCandidateBindingsWithDB 与 BuildCandidateBindings 等价，但允许上层显式传入事务 tx，
// 确保“幂等判断 + 选路前读”在同一事务内完成。
func BuildCandidateBindingsWithDB(ctx context.Context, db gdb.DB, now time.Time, goodsCode string, quantity int, payloadJSON string) (CandidateBuildOutput, error) {
	return buildCandidateBindings(ctx, ctxDBRunner{ctx: ctx, db: db}, now, goodsCode, quantity, payloadJSON)
}

func buildCandidateBindings(ctx context.Context, runner sqlRunner, now time.Time, goodsCode string, quantity int, payloadJSON string) (CandidateBuildOutput, error) {
	goods, err := loadActiveChannelGoods(ctx, runner, goodsCode)
	if err != nil {
		return CandidateBuildOutput{}, err
	}
	if goods.SupplyType != "channel" {
		return CandidateBuildOutput{}, apiErr(consts.CodeBadRequest, "商品供货方式必须为渠道")
	}
	if quantity < goods.MinPurchaseQty || quantity > goods.MaxPurchaseQty {
		return CandidateBuildOutput{}, apiErr(consts.CodeBadRequest, "购买数量不符合限制")
	}

	payloadValue, err := extractPayloadPrimaryValue(payloadJSON)
	if err != nil {
		return CandidateBuildOutput{}, apiErr(consts.CodeBadRequest, err.Error())
	}
	if goods.ProductTemplateID != nil {
		validateType, err := loadTemplateValidateType(ctx, runner, *goods.ProductTemplateID)
		if err != nil {
			return CandidateBuildOutput{}, err
		}
		if err := validateTemplateValue(payloadValue, validateType); err != nil {
			return CandidateBuildOutput{}, apiErr(consts.CodeBadRequest, "payload不符合商品模板")
		}
	}

	cfg, err := loadGoodsChannelConfigSnapshot(ctx, runner, now, goods.ID)
	if err != nil {
		return CandidateBuildOutput{}, err
	}

	candidates, err := loadCandidateBindings(ctx, runner, goods.SubjectID, goods.ID, payloadValue)
	if err != nil {
		return CandidateBuildOutput{}, err
	}
	if len(candidates) == 0 {
		return CandidateBuildOutput{}, apiErr(consts.CodeBadRequest, "无可用绑定")
	}

	return CandidateBuildOutput{Goods: goods, Config: cfg, Candidates: candidates}, nil
}

type goodsRow struct {
	ID                int64          `db:"id"`
	GoodsCode         string         `db:"goods_code"`
	GoodsName         string         `db:"goods_name"`
	SupplyType        string         `db:"supply_type"`
	HasTax            int            `db:"has_tax"`
	SubjectID         sql.NullInt64  `db:"subject_id"`
	DefaultSellPrice  sql.NullString `db:"default_sell_price"`
	MinPurchaseQty    int            `db:"min_purchase_qty"`
	MaxPurchaseQty    int            `db:"max_purchase_qty"`
	Status            int            `db:"status"`
	IsDeleted         int            `db:"is_deleted"`
	ProductTemplateID sql.NullInt64  `db:"product_template_id"`
}

func loadActiveChannelGoods(ctx context.Context, runner sqlRunner, goodsCode string) (GoodsSnapshot, error) {
	goodsCode = strings.TrimSpace(goodsCode)
	if goodsCode == "" {
		return GoodsSnapshot{}, apiErr(consts.CodeBadRequest, "goods_code不能为空")
	}

	record, err := runner.GetOne(`
SELECT
    p.id,
    p.goods_code,
    p.name AS goods_name,
    p.supply_type,
    p.has_tax,
    p.subject_id,
    COALESCE(sub.name, '') AS subject_name,
    p.default_sell_price,
    p.min_purchase_qty,
    p.max_purchase_qty,
    p.status,
    p.is_deleted,
    p.product_template_id
FROM product_goods p
LEFT JOIN admin_subject sub ON sub.id = p.subject_id
WHERE p.goods_code = ? AND p.is_deleted = 0
`, goodsCode)
	if err != nil {
		return GoodsSnapshot{}, apiErr(consts.CodeInternalError, "读取商品失败")
	}
	if record == nil || len(record) == 0 {
		return GoodsSnapshot{}, apiErr(consts.CodeBadRequest, "商品不存在")
	}
	if record["status"].Int() != 1 {
		return GoodsSnapshot{}, apiErr(consts.CodeBadRequest, "商品不可售")
	}

	subjectValue := record["subject_id"]
	if subjectValue == nil || subjectValue.IsNil() || subjectValue.Int64() <= 0 {
		return GoodsSnapshot{}, apiErr(consts.CodeBadRequest, "商品主体未配置")
	}

	defaultSellRaw := strings.TrimSpace(record["default_sell_price"].String())
	if defaultSellRaw == "" {
		return GoodsSnapshot{}, apiErr(consts.CodeBadRequest, "商品默认售价未配置")
	}
	defaultSell, err := decimal.NewFromString(defaultSellRaw)
	if err != nil {
		return GoodsSnapshot{}, apiErr(consts.CodeBadRequest, "商品默认售价格式错误")
	}

	var templateID *int64
	if value := record["product_template_id"]; value != nil && !value.IsNil() && value.Int64() > 0 {
		id := value.Int64()
		templateID = &id
	}

	return GoodsSnapshot{
		ID:                record["id"].Int64(),
		GoodsCode:         record["goods_code"].String(),
		GoodsName:         record["goods_name"].String(),
		SupplyType:        record["supply_type"].String(),
		HasTax:            record["has_tax"].Int(),
		SubjectID:         subjectValue.Int64(),
		SubjectName:       strings.TrimSpace(record["subject_name"].String()),
		DefaultSellPrice:  Round4(defaultSell),
		MinPurchaseQty:    record["min_purchase_qty"].Int(),
		MaxPurchaseQty:    record["max_purchase_qty"].Int(),
		ProductTemplateID: templateID,
	}, nil
}

func loadGoodsChannelConfigSnapshot(ctx context.Context, runner sqlRunner, now time.Time, goodsID int64) (GoodsChannelConfigSnapshot, error) {
	if goodsID <= 0 {
		return GoodsChannelConfigSnapshot{}, apiErr(consts.CodeBadRequest, "goods_id错误")
	}
	if err := ensureGoodsChannelConfigRow(ctx, runner, now, goodsID); err != nil {
		return GoodsChannelConfigSnapshot{}, apiErr(consts.CodeInternalError, "读取渠道配置失败")
	}
	record, err := runner.GetOne(`
SELECT
    goods_id,
    smart_replenish_enabled,
    attempt_timeout_enabled,
    attempt_timeout_minutes,
    route_mode,
    allow_loss,
    max_loss_amount,
    sync_cost_enabled,
    sync_goods_name_enabled,
    is_bundle
FROM product_goods_channel_config
WHERE goods_id = ?
`, goodsID)
	if err != nil {
		return GoodsChannelConfigSnapshot{}, apiErr(consts.CodeInternalError, "读取渠道配置失败")
	}
	if record == nil || len(record) == 0 {
		return GoodsChannelConfigSnapshot{}, apiErr(consts.CodeInternalError, "读取渠道配置失败")
	}

	var maxLoss *decimal.Decimal
	raw := strings.TrimSpace(record["max_loss_amount"].String())
	if raw != "" && record["max_loss_amount"] != nil && !record["max_loss_amount"].IsNil() {
		amount, err := decimal.NewFromString(raw)
		if err != nil || amount.IsNegative() {
			return GoodsChannelConfigSnapshot{}, apiErr(consts.CodeBadRequest, "max_loss_amount格式错误")
		}
		rounded := Round4(amount)
		maxLoss = &rounded
	}

	return GoodsChannelConfigSnapshot{
		GoodsID:               record["goods_id"].Int64(),
		RouteMode:             strings.TrimSpace(record["route_mode"].String()),
		SmartReplenishEnabled: record["smart_replenish_enabled"].Int() != 0,
		AttemptTimeoutEnabled: record["attempt_timeout_enabled"].Int() != 0,
		AttemptTimeoutMinutes: record["attempt_timeout_minutes"].Int(),
		AllowLoss:             record["allow_loss"].Int() != 0,
		MaxLossAmount:         maxLoss,
		SyncCostEnabled:       record["sync_cost_enabled"].Int() != 0,
		SyncGoodsNameEnabled:  record["sync_goods_name_enabled"].Int() != 0,
		IsBundle:              record["is_bundle"].Int() != 0,
	}, nil
}

func ensureGoodsChannelConfigRow(ctx context.Context, runner sqlRunner, now time.Time, goodsID int64) error {
	exists, err := runner.GetValue(`SELECT COUNT(*) FROM product_goods_channel_config WHERE goods_id = ?`, goodsID)
	if err != nil {
		return err
	}
	if exists.Int() > 0 {
		return nil
	}
	_, err = runner.Exec(`INSERT INTO product_goods_channel_config (goods_id, created_at, updated_at) VALUES (?, ?, ?)`, goodsID, now, now)
	return err
}

type bindingRow struct {
	ID                 int64         `db:"id"`
	GoodsID            int64         `db:"goods_id"`
	PlatformAccountID  int64         `db:"platform_account_id"`
	SupplierGoodsNo    string        `db:"supplier_goods_no"`
	SupplierGoodsName  string        `db:"supplier_goods_name"`
	SourceCostPrice    string        `db:"source_cost_price"`
	CostPrice          string        `db:"cost_price"`
	TaxAdjustDirection string        `db:"tax_adjust_direction"`
	TaxAdjustRate      string        `db:"tax_adjust_rate"`
	TaxAdjustAmount    string        `db:"tax_adjust_amount"`
	DockStatus         string        `db:"dock_status"`
	Sort               int           `db:"sort"`
	Weight             int           `db:"weight"`
	StartTime          string        `db:"start_time"`
	EndTime            string        `db:"end_time"`
	ValidateTemplateID sql.NullInt64 `db:"validate_template_id"`
	IsAutoChange       int           `db:"is_auto_change"`
	AddType            string        `db:"add_type"`
	DefaultPrice       string        `db:"default_price"`
	ProviderCode       string        `db:"provider_code"`
	ProviderName       string        `db:"provider_name"`
}

func loadCandidateBindings(ctx context.Context, runner sqlRunner, goodsSubjectID int64, goodsID int64, payloadValue string) ([]CandidateBinding, error) {
	if goodsID <= 0 {
		return nil, apiErr(consts.CodeBadRequest, "goods_id错误")
	}
	rows := make([]bindingRow, 0)
	if err := runner.GetScan(&rows, `
SELECT
    b.id,
    b.goods_id,
    b.platform_account_id,
    b.supplier_goods_no,
    b.supplier_goods_name,
    b.source_cost_price,
    b.cost_price,
    b.tax_adjust_direction,
    b.tax_adjust_rate,
    b.tax_adjust_amount,
    b.dock_status,
    b.sort,
    b.weight,
    b.start_time,
    b.end_time,
    b.validate_template_id,
    b.is_auto_change,
    b.add_type,
    b.default_price,
    a.provider_code,
    a.provider_name
FROM product_goods_channel_binding b
JOIN supplier_platform_account a ON a.id = b.platform_account_id AND a.is_deleted = 0
JOIN supplier_platform_type st ON st.id = a.type_id AND st.status = 1
WHERE b.goods_id = ? AND b.is_deleted = 0 AND b.dock_status = 'enabled' AND a.subject_id = ?
`, goodsID, goodsSubjectID); err != nil {
		return nil, apiErr(consts.CodeInternalError, "读取绑定失败")
	}
	if len(rows) == 0 {
		return nil, nil
	}

	result := make([]CandidateBinding, 0, len(rows))
	for _, row := range rows {
		sourceCost, err := ParseMoney(row.SourceCostPrice)
		if err != nil || sourceCost.IsNegative() {
			continue
		}
		cost, err := ParseMoney(row.CostPrice)
		if err != nil || cost.IsNegative() {
			continue
		}

		var validateTemplateID *int64
		if row.ValidateTemplateID.Valid && row.ValidateTemplateID.Int64 > 0 {
			id := row.ValidateTemplateID.Int64
			validateTemplateID = &id
			validateType, err := loadTemplateValidateType(ctx, runner, id)
			if err != nil {
				return nil, err
			}
			if err := validateTemplateValue(payloadValue, validateType); err != nil {
				continue
			}
		}

		taxRate := decimal.Zero
		if strings.TrimSpace(row.TaxAdjustRate) != "" {
			if value, err := decimal.NewFromString(strings.TrimSpace(row.TaxAdjustRate)); err == nil {
				taxRate = Round4(value)
			}
		}
		taxAmount := decimal.Zero
		if strings.TrimSpace(row.TaxAdjustAmount) != "" {
			if value, err := decimal.NewFromString(strings.TrimSpace(row.TaxAdjustAmount)); err == nil {
				taxAmount = Round4(value)
			}
		}

		defaultPrice := decimal.Zero
		if strings.TrimSpace(row.DefaultPrice) != "" {
			if value, err := decimal.NewFromString(strings.TrimSpace(row.DefaultPrice)); err == nil {
				defaultPrice = Round4(value)
			} else {
				// default_price 解析失败时直接过滤，避免锁价出错。
				continue
			}
		}

		result = append(result, CandidateBinding{
			ID:                 row.ID,
			GoodsID:            row.GoodsID,
			PlatformAccountID:  row.PlatformAccountID,
			ProviderCode:       strings.TrimSpace(row.ProviderCode),
			ProviderName:       strings.TrimSpace(row.ProviderName),
			SupplierGoodsNo:    strings.TrimSpace(row.SupplierGoodsNo),
			SupplierGoodsName:  strings.TrimSpace(row.SupplierGoodsName),
			DockStatus:         strings.TrimSpace(row.DockStatus),
			Sort:               row.Sort,
			Weight:             row.Weight,
			StartTime:          strings.TrimSpace(row.StartTime),
			EndTime:            strings.TrimSpace(row.EndTime),
			ValidateTemplateID: validateTemplateID,
			SourceCostPrice:    Round4(sourceCost),
			CostPrice:          Round4(cost),
			TaxAdjustDirection: strings.TrimSpace(row.TaxAdjustDirection),
			TaxAdjustRate:      taxRate,
			TaxAdjustAmount:    taxAmount,
			IsAutoChange:       row.IsAutoChange != 0,
			AddType:            strings.TrimSpace(row.AddType),
			DefaultPrice:       defaultPrice,
		})
	}

	return result, nil
}

func extractPayloadPrimaryValue(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", fmt.Errorf("payload不能为空")
	}
	payload := make(map[string]any)
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return "", fmt.Errorf("payload格式错误")
	}

	type kv struct {
		Key   string
		Value string
	}
	candidates := make([]kv, 0, len(payload))
	for key, value := range payload {
		v := strings.TrimSpace(fmt.Sprint(value))
		if v == "" || v == "<nil>" {
			continue
		}
		candidates = append(candidates, kv{Key: strings.TrimSpace(key), Value: v})
	}
	if len(candidates) == 0 {
		return "", fmt.Errorf("payload不能为空")
	}

	priority := []string{"mobile", "qq", "wechat", "email", "url", "account", "value"}
	for _, key := range priority {
		for _, item := range candidates {
			if item.Key == key {
				return item.Value, nil
			}
		}
	}

	if len(candidates) == 1 {
		return candidates[0].Value, nil
	}

	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Key < candidates[j].Key })
	return candidates[0].Value, nil
}

func loadTemplateValidateType(ctx context.Context, runner sqlRunner, templateID int64) (int, error) {
	value, err := runner.GetValue(`SELECT validate_type FROM product_template WHERE id = ?`, templateID)
	if err != nil {
		return 0, apiErr(consts.CodeInternalError, "读取模板失败")
	}
	if value.IsNil() {
		return 0, apiErr(consts.CodeBadRequest, "商品模板不存在")
	}
	return value.Int(), nil
}

var (
	qqRegexp     = regexp.MustCompile(`^[1-9][0-9]{4,11}$`)
	wechatRegexp = regexp.MustCompile(`^[a-zA-Z][-_a-zA-Z0-9]{5,19}$`)
	digitsRegexp = regexp.MustCompile(`^[0-9]+$`)
)

func validateTemplateValue(value string, validateType int) error {
	value = strings.TrimSpace(value)
	if value == "" {
		return fmt.Errorf("empty")
	}

	isPhone := app.PhoneRegexp().MatchString(value)
	isQQ := qqRegexp.MatchString(value)
	isWeChat := wechatRegexp.MatchString(value)
	isDigits := digitsRegexp.MatchString(value)
	isEmail := isEmail(value)
	isURL := isURL(value)

	switch validateType {
	case 1:
		if !isPhone {
			return fmt.Errorf("invalid mobile")
		}
	case 2:
		if !isQQ {
			return fmt.Errorf("invalid qq")
		}
	case 3:
		if !isPhone && !isQQ {
			return fmt.Errorf("invalid mobile/qq")
		}
	case 4:
		if !isEmail {
			return fmt.Errorf("invalid email")
		}
	case 5:
		if !isURL {
			return fmt.Errorf("invalid url")
		}
	case 6:
		if !isDigits {
			return fmt.Errorf("invalid digits")
		}
	case 7:
		if !isWeChat {
			return fmt.Errorf("invalid wechat")
		}
	case 8:
		if !isPhone && !isWeChat {
			return fmt.Errorf("invalid mobile/wechat")
		}
	case 9:
		if !isQQ && !isWeChat {
			return fmt.Errorf("invalid qq/wechat")
		}
	case 10:
		if !isPhone && !isQQ && !isWeChat {
			return fmt.Errorf("invalid mobile/qq/wechat")
		}
	case 11:
		if isPhone {
			return fmt.Errorf("phone forbidden")
		}
	case 12:
		if isEmail {
			return fmt.Errorf("email forbidden")
		}
	default:
		return fmt.Errorf("unknown validate_type")
	}
	return nil
}

func isEmail(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || !strings.Contains(value, "@") {
		return false
	}
	_, err := mail.ParseAddress(value)
	return err == nil
}

func isURL(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}
	return parsed.Scheme != "" && parsed.Host != ""
}
