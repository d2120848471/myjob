package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

// GroupService 定义用户组、菜单树与授权保存相关能力。
type GroupService interface {
	List(ctx context.Context, req *adminapi.GroupListReq) (*adminapi.GroupListRes, error)
	Add(ctx context.Context, req *adminapi.GroupCreateReq, actor entity.AdminUser, ip string) (*adminapi.GroupCreateRes, error)
	Edit(ctx context.Context, req *adminapi.GroupUpdateReq, actor entity.AdminUser, ip string) (*adminapi.GroupUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.GroupDeleteReq, actor entity.AdminUser, ip string) (*adminapi.GroupDeleteRes, error)
	Status(ctx context.Context, req *adminapi.GroupStatusReq, actor entity.AdminUser, ip string) (*adminapi.GroupStatusRes, error)
	AuthGet(ctx context.Context, req *adminapi.GroupPermissionsGetReq) (*adminapi.GroupPermissionsGetRes, error)
	AuthSave(ctx context.Context, req *adminapi.GroupPermissionsSaveReq, actor entity.AdminUser, ip string) (*adminapi.GroupPermissionsSaveRes, error)
	MenuTree(ctx context.Context, req *adminapi.MenuTreeReq) (*adminapi.MenuTreeRes, error)
}
