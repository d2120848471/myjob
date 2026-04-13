package api

import (
	"myjob/internal/model/dto/admin"
	"myjob/internal/model/entity"
	modelruntime "myjob/internal/model/runtime"
)

type PaginationRes = admin.Pagination

type LoginUser = modelruntime.LoginUser

type GroupListItem = entity.GroupListItem

type UserListItem = entity.UserListItem

type SubjectItem = entity.AdminSubject

type BrandListItem = entity.BrandListItem

type IndustryListItem = entity.IndustryListItem

type IndustryBrandRelationItem = entity.IndustryBrandRelationItem

type BrandSelectorItem = entity.BrandSelectorItem

type ProductTemplateListItem = entity.ProductTemplateListItem

type ProductTemplateValidateTypeItem = entity.ProductTemplateValidateTypeItem

type PurchaseLimitStrategyListItem = entity.PurchaseLimitStrategyListItem

type PurchaseLimitEnumItem = entity.PurchaseLimitEnumItem

type MenuItem = entity.AdminMenu

type OperationLogItem = entity.OperationLog

type LoginLogItem = entity.LoginLog
