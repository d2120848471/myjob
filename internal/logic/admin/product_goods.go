package adminlogic

import (
	"net/http"
	"time"

	"myjob/internal/app"
)

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
type ProductGoodsLogic struct {
	core               *app.Core
	httpClient         *http.Client
	productPushBaseURL string
}

// NewProductGoodsLogic 创建商品管理业务逻辑，并初始化供应商同步使用的 HTTP 客户端。
func NewProductGoodsLogic(core *app.Core) *ProductGoodsLogic {
	return &ProductGoodsLogic{
		core:       core,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}
