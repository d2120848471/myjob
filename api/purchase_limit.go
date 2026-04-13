package api

import "github.com/gogf/gf/v2/frame/g"

type PurchaseLimitStrategyListReq struct {
	g.Meta   `path:"/purchase-limit-strategies" method:"get" tags:"商品购买数量限制策略" summary:"策略列表" security:"BearerAuth" dc:"分页查询商品购买数量限制策略"`
	Page     int    `json:"page" dc:"页码"`
	PageSize int    `json:"page_size" dc:"每页条数"`
	Keyword  string `json:"keyword" dc:"关键词"`
}

type PurchaseLimitStrategyListRes struct {
	List       []PurchaseLimitStrategyListItem `json:"list" dc:"策略列表"`
	Pagination PaginationRes                   `json:"pagination" dc:"分页信息"`
}

type PurchaseLimitStrategyCreateReq struct {
	g.Meta     `path:"/purchase-limit-strategies" method:"post" tags:"商品购买数量限制策略" summary:"新增策略" security:"BearerAuth" dc:"新增商品购买数量限制策略"`
	Name       string `json:"name" dc:"策略名称"`
	LimitType  int    `json:"limit_type" dc:"限制类型"`
	PeriodType int    `json:"period_type" dc:"周期类型"`
	Period     int    `json:"period" dc:"限制周期"`
	LimitNums  int    `json:"limit_nums" dc:"限制数量"`
	LimitTimes int    `json:"limit_times" dc:"限制笔数"`
}

type PurchaseLimitStrategyCreateRes struct {
	ID int64 `json:"id" dc:"策略ID"`
}

type PurchaseLimitStrategyUpdateReq struct {
	g.Meta     `path:"/purchase-limit-strategies/{id}" method:"put" tags:"商品购买数量限制策略" summary:"编辑策略" security:"BearerAuth" dc:"编辑商品购买数量限制策略"`
	ID         int64  `json:"id" in:"path" v:"required#策略ID不能为空" dc:"策略ID"`
	Name       string `json:"name" dc:"策略名称"`
	LimitType  int    `json:"limit_type" dc:"限制类型"`
	PeriodType int    `json:"period_type" dc:"周期类型"`
	Period     int    `json:"period" dc:"限制周期"`
	LimitNums  int    `json:"limit_nums" dc:"限制数量"`
	LimitTimes int    `json:"limit_times" dc:"限制笔数"`
}

type PurchaseLimitStrategyUpdateRes struct{}

type PurchaseLimitStrategyDeleteReq struct {
	g.Meta `path:"/purchase-limit-strategies/{id}" method:"delete" tags:"商品购买数量限制策略" summary:"删除策略" security:"BearerAuth" dc:"删除商品购买数量限制策略"`
	ID     int64 `json:"id" in:"path" v:"required#策略ID不能为空" dc:"策略ID"`
}

type PurchaseLimitStrategyDeleteRes struct{}

type PurchaseLimitStrategyStatusReq struct {
	g.Meta `path:"/purchase-limit-strategies/{id}/status" method:"patch" tags:"商品购买数量限制策略" summary:"切换状态" security:"BearerAuth" dc:"切换商品购买数量限制策略状态"`
	ID     int64 `json:"id" in:"path" v:"required#策略ID不能为空" dc:"策略ID"`
	Status int   `json:"status" dc:"状态"`
}

type PurchaseLimitStrategyStatusRes struct{}

type PurchaseLimitStrategyEnumsReq struct {
	g.Meta `path:"/purchase-limit-strategies/enums" method:"get" tags:"商品购买数量限制策略" summary:"读取枚举" security:"BearerAuth" dc:"读取商品购买数量限制策略枚举"`
}

type PurchaseLimitStrategyEnumsRes struct {
	LimitTypes  []PurchaseLimitEnumItem `json:"limit_types" dc:"限制类型枚举"`
	PeriodTypes []PurchaseLimitEnumItem `json:"period_types" dc:"周期类型枚举"`
}
