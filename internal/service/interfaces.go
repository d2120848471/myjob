package service

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
)

type AuthService interface {
	Login(ctx context.Context, req *adminapi.AuthLoginReq, ip string) (*adminapi.AuthLoginRes, error)
	LoginSMSSend(ctx context.Context, req *adminapi.AuthSMSSendReq) (*adminapi.AuthSMSSendRes, error)
	LoginSMSVerify(ctx context.Context, req *adminapi.AuthSMSVerifyReq) (*adminapi.AuthSMSVerifyRes, error)
	Me(ctx context.Context, principal modelruntime.Principal, user entity.AdminUser) (*adminapi.AuthMeRes, error)
	Logout(ctx context.Context, principal modelruntime.Principal, user entity.AdminUser) (*adminapi.AuthSessionDeleteRes, error)
}

type UserService interface {
	List(ctx context.Context, req *adminapi.UserListReq) (*adminapi.UserListRes, error)
	Trash(ctx context.Context, req *adminapi.UserTrashReq) (*adminapi.UserTrashRes, error)
	Add(ctx context.Context, req *adminapi.UserCreateReq, actor entity.AdminUser, ip string) (*adminapi.UserCreateRes, error)
	Edit(ctx context.Context, req *adminapi.UserUpdateReq, actor entity.AdminUser, ip string) (*adminapi.UserUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.UserDeleteReq, actor entity.AdminUser, ip string) (*adminapi.UserDeleteRes, error)
	Restore(ctx context.Context, req *adminapi.UserRestoreReq, actor entity.AdminUser, ip string) (*adminapi.UserRestoreRes, error)
	Status(ctx context.Context, req *adminapi.UserStatusReq, actor entity.AdminUser, ip string) (*adminapi.UserStatusRes, error)
	Notify(ctx context.Context, req *adminapi.UserNotifyReq, actor entity.AdminUser, ip string) (*adminapi.UserNotifyRes, error)
	SetBusiness(ctx context.Context, req *adminapi.UserBusinessAssignReq, actor entity.AdminUser, ip string) (*adminapi.UserBusinessAssignRes, error)
	CancelBusiness(ctx context.Context, req *adminapi.UserBusinessCancelReq, actor entity.AdminUser, ip string) (*adminapi.UserBusinessCancelRes, error)
}

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

type SubjectService interface {
	List(ctx context.Context, req *adminapi.SubjectListReq) (*adminapi.SubjectListRes, error)
	Add(ctx context.Context, req *adminapi.SubjectCreateReq, actor entity.AdminUser, ip string) (*adminapi.SubjectCreateRes, error)
	Edit(ctx context.Context, req *adminapi.SubjectUpdateReq, actor entity.AdminUser, ip string) (*adminapi.SubjectUpdateRes, error)
}

type BrandService interface {
	List(ctx context.Context, req *adminapi.BrandListReq) (*adminapi.BrandListRes, error)
	Children(ctx context.Context, req *adminapi.BrandChildrenReq) (*adminapi.BrandChildrenRes, error)
	Add(ctx context.Context, req *adminapi.BrandCreateReq, actor entity.AdminUser, ip string) (*adminapi.BrandCreateRes, error)
	Edit(ctx context.Context, req *adminapi.BrandUpdateReq, actor entity.AdminUser, ip string) (*adminapi.BrandUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.BrandDeleteReq, actor entity.AdminUser, ip string) (*adminapi.BrandDeleteRes, error)
	Sort(ctx context.Context, req *adminapi.BrandSortReq, actor entity.AdminUser, ip string) (*adminapi.BrandSortRes, error)
	Visibility(ctx context.Context, req *adminapi.BrandVisibilityReq, actor entity.AdminUser, ip string) (*adminapi.BrandVisibilityRes, error)
	Upload(ctx context.Context, req *adminapi.BrandUploadReq, actor entity.AdminUser, ip string) (*adminapi.BrandUploadRes, error)
}

type IndustryService interface {
	List(ctx context.Context, req *adminapi.IndustryListReq) (*adminapi.IndustryListRes, error)
	Add(ctx context.Context, req *adminapi.IndustryCreateReq, actor entity.AdminUser, ip string) (*adminapi.IndustryCreateRes, error)
	Edit(ctx context.Context, req *adminapi.IndustryUpdateReq, actor entity.AdminUser, ip string) (*adminapi.IndustryUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.IndustryDeleteReq, actor entity.AdminUser, ip string) (*adminapi.IndustryDeleteRes, error)
	Sort(ctx context.Context, req *adminapi.IndustrySortReq, actor entity.AdminUser, ip string) (*adminapi.IndustrySortRes, error)
	BrandSelector(ctx context.Context, req *adminapi.IndustryBrandSelectorReq) (*adminapi.IndustryBrandSelectorRes, error)
	BrandList(ctx context.Context, req *adminapi.IndustryBrandListReq) (*adminapi.IndustryBrandListRes, error)
	BrandAdd(ctx context.Context, req *adminapi.IndustryBrandAddReq, actor entity.AdminUser, ip string) (*adminapi.IndustryBrandAddRes, error)
	BrandDelete(ctx context.Context, req *adminapi.IndustryBrandDeleteReq, actor entity.AdminUser, ip string) (*adminapi.IndustryBrandDeleteRes, error)
	BrandSort(ctx context.Context, req *adminapi.IndustryBrandSortReq, actor entity.AdminUser, ip string) (*adminapi.IndustryBrandSortRes, error)
}

type ProductTemplateService interface {
	List(ctx context.Context, req *adminapi.ProductTemplateListReq) (*adminapi.ProductTemplateListRes, error)
	Add(ctx context.Context, req *adminapi.ProductTemplateCreateReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateCreateRes, error)
	Edit(ctx context.Context, req *adminapi.ProductTemplateUpdateReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.ProductTemplateDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateDeleteRes, error)
	BatchDelete(ctx context.Context, req *adminapi.ProductTemplateBatchDeleteReq, actor entity.AdminUser, ip string) (*adminapi.ProductTemplateBatchDeleteRes, error)
	ValidateTypes(ctx context.Context, req *adminapi.ProductTemplateValidateTypeListReq) (*adminapi.ProductTemplateValidateTypeListRes, error)
}

type PurchaseLimitService interface {
	List(ctx context.Context, req *adminapi.PurchaseLimitStrategyListReq) (*adminapi.PurchaseLimitStrategyListRes, error)
	Add(ctx context.Context, req *adminapi.PurchaseLimitStrategyCreateReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyCreateRes, error)
	Edit(ctx context.Context, req *adminapi.PurchaseLimitStrategyUpdateReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyUpdateRes, error)
	Delete(ctx context.Context, req *adminapi.PurchaseLimitStrategyDeleteReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyDeleteRes, error)
	Status(ctx context.Context, req *adminapi.PurchaseLimitStrategyStatusReq, actor entity.AdminUser, ip string) (*adminapi.PurchaseLimitStrategyStatusRes, error)
	Enums(ctx context.Context, req *adminapi.PurchaseLimitStrategyEnumsReq) (*adminapi.PurchaseLimitStrategyEnumsRes, error)
}

type SMSConfigService interface {
	Get(ctx context.Context, req *adminapi.SettingsSMSGetReq) (*adminapi.SettingsSMSGetRes, error)
	Save(ctx context.Context, req *adminapi.SettingsSMSSaveReq, actor entity.AdminUser, ip string) (*adminapi.SettingsSMSSaveRes, error)
}

type SystemConfigService interface {
	Get(ctx context.Context, req *adminapi.SettingsSystemGetReq) (*adminapi.SettingsSystemGetRes, error)
	Save(ctx context.Context, req *adminapi.SettingsSystemSaveReq, actor entity.AdminUser, ip string) (*adminapi.SettingsSystemSaveRes, error)
}

type AuditLogService interface {
	OperationList(ctx context.Context, req *adminapi.OperationLogListReq) (*adminapi.OperationLogListRes, error)
	LoginList(ctx context.Context, req *adminapi.LoginLogListReq) (*adminapi.LoginLogListRes, error)
}
