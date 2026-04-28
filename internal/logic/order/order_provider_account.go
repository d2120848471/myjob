package orderlogic

import (
	"encoding/json"
	"strings"

	supplierprovider "myjob/internal/library/supplierplatform/provider"
	"myjob/internal/model/entity"
)

func supplierAccountConfigFromEntity(account entity.SupplierPlatformAccount) supplierprovider.AccountConfig {
	return supplierprovider.AccountConfig{
		ProviderCode: account.ProviderCode,
		Domain:       account.Domain,
		BackupDomain: account.BackupDomain,
		TokenID:      account.TokenID,
		SecretKey:    account.SecretKey,
		// 扩展配置只影响特定 provider 的兼容参数；历史脏值解析失败时保持默认行为。
		ExtraConfig: parseSupplierAccountExtraConfig(account.ExtraConfig),
	}
}

func parseSupplierAccountExtraConfig(raw string) map[string]any {
	result := map[string]any{}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return result
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return map[string]any{}
	}
	return result
}
