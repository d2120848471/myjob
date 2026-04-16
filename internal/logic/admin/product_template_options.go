package adminlogic

import (
	"context"

	adminapi "myjob/api"
	"myjob/internal/model/entity"
)

const productTemplateTypeLocal = "local"

var productTemplateValidateTypes = []entity.ProductTemplateValidateTypeItem{
	{ID: 1, Title: "手机号"},
	{ID: 2, Title: "QQ号"},
	{ID: 3, Title: "手机号或者QQ号"},
	{ID: 4, Title: "邮箱"},
	{ID: 5, Title: "网址"},
	{ID: 6, Title: "纯数字"},
	{ID: 7, Title: "微信号"},
	{ID: 8, Title: "手机号或者微信号"},
	{ID: 9, Title: "QQ号或者微信号"},
	{ID: 10, Title: "手机号或者QQ号或微信号"},
	{ID: 11, Title: "禁止填写手机号"},
	{ID: 12, Title: "禁止填写邮箱"},
}

var productTemplateValidateTypeTitles = func() map[int]string {
	items := make(map[int]string, len(productTemplateValidateTypes))
	for _, item := range productTemplateValidateTypes {
		items[item.ID] = item.Title
	}
	return items
}()

// ValidateTypes 返回商品模板“校验类型”枚举列表。
func (l *ProductTemplateLogic) ValidateTypes(ctx context.Context, req *adminapi.ProductTemplateValidateTypeListReq) (*adminapi.ProductTemplateValidateTypeListRes, error) {
	items := make([]entity.ProductTemplateValidateTypeItem, 0, len(productTemplateValidateTypes))
	items = append(items, productTemplateValidateTypes...)
	return &adminapi.ProductTemplateValidateTypeListRes{List: items}, nil
}

func productTemplateTypeLabel(value string) string {
	if value == productTemplateTypeLocal {
		return "本地模板"
	}
	return value
}

func productTemplateSharedLabel(value int) string {
	if value == 1 {
		return "共享"
	}
	return "不共享"
}
