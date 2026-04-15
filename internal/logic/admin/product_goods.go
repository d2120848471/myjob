package adminlogic

import "myjob/internal/app"

const (
	productGoodsTypeCardSecret     = "card_secret"
	productGoodsTypeDirectRecharge = "direct_recharge"
	productGoodsSupplyTypeChannel  = "channel"
)

var productGoodsTypeLabels = map[string]string{
	productGoodsTypeCardSecret:     "卡密",
	productGoodsTypeDirectRecharge: "直充",
}

// ProductGoodsLogic 提供商品管理相关业务能力。
type ProductGoodsLogic struct{ core *app.Core }
