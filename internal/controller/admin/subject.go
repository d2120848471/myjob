package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

// SubjectController 提供主体管理相关 HTTP handler。
type SubjectController struct{ svc service.SubjectService }

// NewSubject 创建 SubjectController。
func NewSubject(svc service.SubjectService) *SubjectController { return &SubjectController{svc: svc} }

// List 返回主体列表。
func (c *SubjectController) List(ctx context.Context, req *adminapi.SubjectListReq) (res *adminapi.SubjectListRes, err error) {
	return c.svc.List(ctx, req)
}

// Create 新增主体，并记录操作人与客户端 IP。
func (c *SubjectController) Create(ctx context.Context, req *adminapi.SubjectCreateReq) (res *adminapi.SubjectCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

// Update 编辑主体信息，并记录操作人与客户端 IP。
func (c *SubjectController) Update(ctx context.Context, req *adminapi.SubjectUpdateReq) (res *adminapi.SubjectUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
