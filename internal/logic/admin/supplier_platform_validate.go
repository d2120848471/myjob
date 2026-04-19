package adminlogic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	supplierprovider "myjob/internal/library/supplierplatform/provider"

	"github.com/shopspring/decimal"
)

type normalizedSupplierPlatformInput struct {
	Name            string
	ProviderCode    string
	ProviderName    string
	TypeID          int
	SubjectID       int64
	HasTax          int
	Status          int
	Domain          string
	BackupDomain    string
	TokenID         string
	SecretKey       string
	ThresholdAmount string
	Sort            int
	CrowdName       string
}

func (l *SupplierPlatformLogic) normalizeSupplierPlatformInput(ctx context.Context, name, domain, backupDomain string, typeID int, subjectID int64, hasTax int, status *int, defaultStatus int, tokenID, secretKey, thresholdAmount string, sortValue int, crowdName string) (normalizedSupplierPlatformInput, error) {
	name = strings.TrimSpace(name)
	domain = strings.TrimSpace(domain)
	backupDomain = strings.TrimSpace(backupDomain)
	tokenID = strings.TrimSpace(tokenID)
	secretKey = strings.TrimSpace(secretKey)
	crowdName = strings.TrimSpace(crowdName)
	if name == "" {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("平台名称不能为空")
	}
	if domain == "" {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("主域名不能为空")
	}
	if backupDomain == "" {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("备用域名不能为空")
	}
	if typeID <= 0 {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("平台类型不能为空")
	}
	if subjectID <= 0 {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("主体不能为空")
	}
	if hasTax != 0 && hasTax != 1 {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("含税值错误")
	}
	normalizedStatus := defaultStatus
	if status != nil {
		normalizedStatus = *status
	}
	if err := validateBooleanFlag(normalizedStatus, "平台状态"); err != nil {
		return normalizedSupplierPlatformInput{}, err
	}
	if tokenID == "" {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("平台账号ID不能为空")
	}
	if secretKey == "" {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("平台密钥不能为空")
	}
	if len(name) > 128 || len(tokenID) > 128 || len(secretKey) > 255 || len(crowdName) > 128 {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("平台参数长度超出限制")
	}
	if sortValue < 0 {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("排序值不能小于0")
	}
	if err := validateSupplierDomain(domain); err != nil {
		return normalizedSupplierPlatformInput{}, err
	}
	if err := validateSupplierDomain(backupDomain); err != nil {
		return normalizedSupplierPlatformInput{}, err
	}
	domain = normalizeSupplierDomain(domain)
	backupDomain = normalizeSupplierDomain(backupDomain)
	if err := l.ensureSupplierPlatformTypeExists(ctx, typeID); err != nil {
		return normalizedSupplierPlatformInput{}, err
	}
	if err := l.ensureSubjectExists(ctx, subjectID); err != nil {
		return normalizedSupplierPlatformInput{}, err
	}
	amount, err := decimal.NewFromString(strings.TrimSpace(thresholdAmount))
	if err != nil || amount.IsNegative() {
		return normalizedSupplierPlatformInput{}, fmt.Errorf("余额阈值错误")
	}
	providerCode, providerName, err := supplierprovider.ResolveProvider(typeID, domain, backupDomain, name)
	if err != nil {
		return normalizedSupplierPlatformInput{}, err
	}
	return normalizedSupplierPlatformInput{
		Name:            name,
		ProviderCode:    providerCode,
		ProviderName:    providerName,
		TypeID:          typeID,
		SubjectID:       subjectID,
		HasTax:          hasTax,
		Status:          normalizedStatus,
		Domain:          domain,
		BackupDomain:    backupDomain,
		TokenID:         tokenID,
		SecretKey:       secretKey,
		ThresholdAmount: amount.StringFixed(4),
		Sort:            sortValue,
		CrowdName:       crowdName,
	}, nil
}

func (l *SupplierPlatformLogic) ensureSupplierPlatformTypeExists(ctx context.Context, typeID int) error {
	count, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM supplier_platform_type WHERE id = ? AND status = 1`, typeID)
	if err != nil {
		return fmt.Errorf("平台类型查询失败")
	}
	if count.Int() == 0 {
		return fmt.Errorf("平台类型不存在")
	}
	return nil
}

func (l *SupplierPlatformLogic) ensureSubjectExists(ctx context.Context, subjectID int64) error {
	count, err := l.core.DB().GetCore().GetValue(ctx, `SELECT COUNT(*) FROM admin_subject WHERE id = ?`, subjectID)
	if err != nil {
		return fmt.Errorf("主体查询失败")
	}
	if count.Int() == 0 {
		return fmt.Errorf("主体不存在")
	}
	return nil
}

func normalizeSupplierDomain(value string) string {
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimPrefix(value, "https://"), "http://"), "/"))
}

func validateSupplierDomain(value string) error {
	lower := strings.ToLower(strings.TrimSpace(value))
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return fmt.Errorf("域名不能带协议头")
	}
	if strings.Contains(value, "/") {
		return fmt.Errorf("域名不能包含路径")
	}
	return nil
}

func normalizeSupplierTaxFilter(value string) (int, bool, error) {
	value = strings.TrimSpace(value)
	switch value {
	case "", "-1":
		return 0, false, nil
	case "0":
		return 0, true, nil
	case "1":
		return 1, true, nil
	default:
		return 0, false, fmt.Errorf("含税筛选值错误")
	}
}

func normalizeSupplierConnectStatus(value string) (int, bool, error) {
	value = strings.TrimSpace(value)
	switch value {
	case "", "-1":
		return 0, false, nil
	case "0":
		return 0, true, nil
	case "1":
		return 1, true, nil
	case "2":
		return 2, true, nil
	default:
		return 0, false, fmt.Errorf("对接状态筛选值错误")
	}
}

func parseExtraConfig(raw string) (map[string]any, error) {
	result := map[string]any{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return result, nil
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return nil, err
	}
	return result, nil
}
