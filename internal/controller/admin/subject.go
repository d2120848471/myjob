package admincontroller

import (
	"context"

	v1 "myjob/api/admin/v1"
	authctx "myjob/internal/library/auth"
	"myjob/internal/service"
)

type SubjectController struct{ svc service.SubjectService }

func NewSubject(svc service.SubjectService) *SubjectController { return &SubjectController{svc: svc} }

func (c *SubjectController) List(ctx context.Context, req *v1.SubjectListReq) (res *v1.SubjectListRes, err error) {
	return c.svc.List(ctx, req)
}

func (c *SubjectController) Create(ctx context.Context, req *v1.SubjectCreateReq) (res *v1.SubjectCreateRes, err error) {
	return c.svc.Add(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}

func (c *SubjectController) Update(ctx context.Context, req *v1.SubjectUpdateReq) (res *v1.SubjectUpdateRes, err error) {
	return c.svc.Edit(ctx, req, authctx.MustUserFromCtx(ctx), clientIP(ctx))
}
