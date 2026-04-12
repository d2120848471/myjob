package api

import "github.com/gogf/gf/v2/frame/g"

type OperationLogListReq struct {
	g.Meta    `path:"/logs/operations" method:"get" tags:"日志" summary:"操作日志列表" security:"BearerAuth" dc:"分页查询操作日志"`
	Page      int    `json:"page" dc:"页码"`
	PageSize  int    `json:"page_size" dc:"每页条数"`
	AdminID   string `json:"admin_id" dc:"管理员ID"`
	Keyword   string `json:"keyword" dc:"关键字"`
	StartTime string `json:"start_time" dc:"开始时间"`
	EndTime   string `json:"end_time" dc:"结束时间"`
}

type OperationLogListRes struct {
	List       []OperationLogItem `json:"list" dc:"操作日志列表"`
	Pagination PaginationRes      `json:"pagination" dc:"分页信息"`
}

type LoginLogListReq struct {
	g.Meta    `path:"/logs/logins" method:"get" tags:"日志" summary:"登录日志列表" security:"BearerAuth" dc:"分页查询登录日志"`
	Page      int    `json:"page" dc:"页码"`
	PageSize  int    `json:"page_size" dc:"每页条数"`
	AdminID   string `json:"admin_id" dc:"管理员ID"`
	StartTime string `json:"start_time" dc:"开始时间"`
	EndTime   string `json:"end_time" dc:"结束时间"`
}

type LoginLogListRes struct {
	List       []LoginLogItem `json:"list" dc:"登录日志列表"`
	Pagination PaginationRes  `json:"pagination" dc:"分页信息"`
}
