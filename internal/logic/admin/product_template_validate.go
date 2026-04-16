package adminlogic

import (
	"fmt"
	"strings"
)

type normalizedProductTemplateInput struct {
	Title        string
	Type         string
	IsShared     int
	AccountName  string
	ValidateType int
}

func normalizeProductTemplateInput(title, templateType string, isShared int, accountName string, validateType int) (normalizedProductTemplateInput, error) {
	title = strings.TrimSpace(title)
	accountName = strings.TrimSpace(accountName)
	if title == "" {
		return normalizedProductTemplateInput{}, fmt.Errorf("模板名称不能为空")
	}
	if accountName == "" {
		return normalizedProductTemplateInput{}, fmt.Errorf("充值账号名称不能为空")
	}
	if isShared != 0 && isShared != 1 {
		return normalizedProductTemplateInput{}, fmt.Errorf("共享状态错误")
	}
	normalizedType, err := normalizeProductTemplateType(templateType, true)
	if err != nil {
		return normalizedProductTemplateInput{}, err
	}
	if _, ok := productTemplateValidateTypeTitles[validateType]; !ok {
		return normalizedProductTemplateInput{}, fmt.Errorf("验证方式错误")
	}
	return normalizedProductTemplateInput{
		Title:        title,
		Type:         normalizedType,
		IsShared:     isShared,
		AccountName:  accountName,
		ValidateType: validateType,
	}, nil
}

func normalizeProductTemplateType(value string, defaultLocal bool) (string, error) {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		if defaultLocal {
			return productTemplateTypeLocal, nil
		}
		return "", nil
	}
	if value != productTemplateTypeLocal {
		return "", fmt.Errorf("模板类型错误")
	}
	return value, nil
}

func normalizeProductTemplateSharedFilter(value string) (int, bool, error) {
	value = strings.TrimSpace(value)
	switch value {
	case "":
		return 0, false, nil
	case "0":
		return 0, true, nil
	case "1":
		return 1, true, nil
	default:
		return 0, false, fmt.Errorf("共享状态错误")
	}
}
