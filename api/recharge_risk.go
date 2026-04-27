package api

import "github.com/gogf/gf/v2/frame/g"

// RechargeRiskRuleListReq 用于分页查询充值账号风控规则。
type RechargeRiskRuleListReq struct {
	g.Meta       `path:"/recharge-risks/rules" method:"get" tags:"充值风控" summary:"风控规则列表" security:"BearerAuth" dc:"分页查询充值账号风控规则"`
	Page         int    `json:"page" dc:"页码"`
	PageSize     int    `json:"page_size" dc:"每页条数"`
	Account      string `json:"account" dc:"充值账号"`
	GoodsKeyword string `json:"goods_keyword" dc:"商品名关键词"`
	Status       string `json:"status" dc:"状态：1启用，0停用，空或-1表示全部"`
}

// RechargeRiskRuleListRes 返回风控规则列表与分页信息。
type RechargeRiskRuleListRes struct {
	List       []RechargeRiskRuleItem `json:"list" dc:"规则列表"`
	Pagination PaginationRes          `json:"pagination" dc:"分页信息"`
}

// RechargeRiskRuleItem 是风控规则列表展示项。
type RechargeRiskRuleItem struct {
	ID            int64  `json:"id" dc:"规则ID"`
	Account       string `json:"account" dc:"充值账号"`
	GoodsKeyword  string `json:"goods_keyword" dc:"商品名关键词"`
	Reason        string `json:"reason" dc:"风控原因"`
	Status        int    `json:"status" dc:"状态"`
	StatusText    string `json:"status_text" dc:"状态文案"`
	HitCount      int    `json:"hit_count" dc:"已拦截次数"`
	CreatedByName string `json:"created_by_name" dc:"创建人"`
	UpdatedByName string `json:"updated_by_name" dc:"更新人"`
	CreatedAt     string `json:"created_at" dc:"创建时间"`
	UpdatedAt     string `json:"updated_at" dc:"更新时间"`
}

// RechargeRiskRuleCreateReq 用于新增充值账号风控规则。
type RechargeRiskRuleCreateReq struct {
	g.Meta       `path:"/recharge-risks/rules" method:"post" tags:"充值风控" summary:"新增风控规则" security:"BearerAuth" dc:"新增充值账号风控规则"`
	Account      string `json:"account" dc:"充值账号"`
	GoodsKeyword string `json:"goods_keyword" dc:"商品名关键词"`
	Reason       string `json:"reason" dc:"风控原因"`
	Status       int    `json:"status" dc:"状态：1启用，0停用"`
}

// RechargeRiskRuleCreateRes 返回新增后的规则 ID。
type RechargeRiskRuleCreateRes struct {
	ID int64 `json:"id" dc:"规则ID"`
}

// RechargeRiskRuleUpdateReq 用于编辑充值账号风控规则。
type RechargeRiskRuleUpdateReq struct {
	g.Meta       `path:"/recharge-risks/rules/{id}" method:"put" tags:"充值风控" summary:"编辑风控规则" security:"BearerAuth" dc:"编辑充值账号风控规则"`
	ID           int64  `json:"id" in:"path" v:"required#规则ID不能为空" dc:"规则ID"`
	Account      string `json:"account" dc:"充值账号"`
	GoodsKeyword string `json:"goods_keyword" dc:"商品名关键词"`
	Reason       string `json:"reason" dc:"风控原因"`
	Status       int    `json:"status" dc:"状态：1启用，0停用"`
}

// RechargeRiskRuleUpdateRes 表示风控规则编辑成功。
type RechargeRiskRuleUpdateRes struct{}

// RechargeRiskRuleStatusReq 用于启用或停用充值账号风控规则。
type RechargeRiskRuleStatusReq struct {
	g.Meta `path:"/recharge-risks/rules/{id}/status" method:"patch" tags:"充值风控" summary:"修改风控规则状态" security:"BearerAuth" dc:"启用或停用充值账号风控规则"`
	ID     int64 `json:"id" in:"path" v:"required#规则ID不能为空" dc:"规则ID"`
	Status int   `json:"status" dc:"状态：1启用，0停用"`
}

// RechargeRiskRuleStatusRes 表示风控规则状态修改成功。
type RechargeRiskRuleStatusRes struct{}

// RechargeRiskRuleDeleteReq 用于软删除充值账号风控规则。
type RechargeRiskRuleDeleteReq struct {
	g.Meta `path:"/recharge-risks/rules/{id}" method:"delete" tags:"充值风控" summary:"删除风控规则" security:"BearerAuth" dc:"软删除充值账号风控规则"`
	ID     int64 `json:"id" in:"path" v:"required#规则ID不能为空" dc:"规则ID"`
}

// RechargeRiskRuleDeleteRes 表示风控规则删除成功。
type RechargeRiskRuleDeleteRes struct{}

// RechargeRiskRecordListReq 用于分页查询充值账号风控拦截记录。
type RechargeRiskRecordListReq struct {
	g.Meta       `path:"/recharge-risks/records" method:"get" tags:"充值风控" summary:"风控记录列表" security:"BearerAuth" dc:"分页查询充值账号风控拦截记录"`
	Page         int    `json:"page" dc:"页码"`
	PageSize     int    `json:"page_size" dc:"每页条数"`
	Account      string `json:"account" dc:"充值账号"`
	GoodsKeyword string `json:"goods_keyword" dc:"商品名关键词"`
	StartTime    string `json:"start_time" dc:"拦截开始时间"`
	EndTime      string `json:"end_time" dc:"拦截结束时间"`
}

// RechargeRiskRecordListRes 返回风控拦截记录列表与分页信息。
type RechargeRiskRecordListRes struct {
	List       []RechargeRiskRecordItem `json:"list" dc:"拦截记录列表"`
	Pagination PaginationRes            `json:"pagination" dc:"分页信息"`
}

// RechargeRiskRecordItem 是风控拦截记录列表展示项。
type RechargeRiskRecordItem struct {
	ID             int64  `json:"id" dc:"记录ID"`
	RuleID         int64  `json:"rule_id" dc:"规则ID"`
	OrderNo        string `json:"order_no" dc:"订单号"`
	Account        string `json:"account" dc:"充值账号"`
	MatchedKeyword string `json:"matched_keyword" dc:"命中关键词"`
	GoodsCode      string `json:"goods_code" dc:"商品编码"`
	GoodsName      string `json:"goods_name" dc:"商品名称"`
	Reason         string `json:"reason" dc:"风控原因"`
	InterceptedAt  string `json:"intercepted_at" dc:"拦截时间"`
}
