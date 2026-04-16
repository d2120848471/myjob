package api

import (
	"myjob/internal/model/dto/admin"
)

// PaginationRes 是通用分页返回结构的别名，用于列表类接口保持统一的响应形状。
type PaginationRes = admin.Pagination

// 注意：api/common.go 只保留真正跨业务域复用的通用协议别名（例如分页）。
// 明显只服务单个业务域的 Req/Res/Item/Enum 必须放回对应领域协议文件，避免该文件继续吸附无关职责。
