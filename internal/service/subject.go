package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// SubjectService 定义主体管理相关能力。
type SubjectService interface {
	List(ctx context.Context, req *adminapi.SubjectListReq) (*adminapi.SubjectListRes, error)
	Add(ctx context.Context, req *adminapi.SubjectCreateReq, actor entity.AdminUser, ip string) (*adminapi.SubjectCreateRes, error)
	Edit(ctx context.Context, req *adminapi.SubjectUpdateReq, actor entity.AdminUser, ip string) (*adminapi.SubjectUpdateRes, error)
}
