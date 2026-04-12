package admincontroller

import (
	"context"

	adminapi "myjob/api"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type SubjectController struct{ svc service.SubjectService }

func NewSubject(svc service.SubjectService) *SubjectController { return &SubjectController{svc: svc} }

func (c *SubjectController) List(ctx context.Context, req *adminapi.SubjectListReq) (res *adminapi.SubjectListRes, err error) {
	return c.svc.List(ctx, req)
}

func (c *SubjectController) Create(ctx context.Context, req *adminapi.SubjectCreateReq) (res *adminapi.SubjectCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *SubjectController) Update(ctx context.Context, req *adminapi.SubjectUpdateReq) (res *adminapi.SubjectUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
