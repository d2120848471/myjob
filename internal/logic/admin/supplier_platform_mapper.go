package adminlogic

import (
	"database/sql"
	"strings"

	"myjob/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/shopspring/decimal"
)

func supplierPlatformAccountFromRecord(row gdb.Record) entity.SupplierPlatformAccount {
	return entity.SupplierPlatformAccount{
		ID:                 row["id"].Int64(),
		Name:               row["name"].String(),
		ProviderCode:       row["provider_code"].String(),
		ProviderName:       row["provider_name"].String(),
		TypeID:             row["type_id"].Int(),
		SubjectID:          row["subject_id"].Int64(),
		HasTax:             row["has_tax"].Int(),
		Status:             row["status"].Int(),
		Domain:             row["domain"].String(),
		BackupDomain:       row["backup_domain"].String(),
		TokenID:            row["token_id"].String(),
		SecretKey:          row["secret_key"].String(),
		ExtraConfig:        row["extra_config"].String(),
		ThresholdAmount:    formatMoney(row["threshold_amount"].String()),
		Sort:               row["sort"].Int(),
		CrowdName:          row["crowd_name"].String(),
		LastBalance:        nullableStringFromRecord(row, "last_balance"),
		LastBalanceStatus:  row["last_balance_status"].Int(),
		LastBalanceMessage: row["last_balance_message"].String(),
		LastBalanceAt:      nullableTimeFromRecord(row, "last_balance_at"),
		LastBalanceTraceID: row["last_balance_trace_id"].String(),
		IsDeleted:          row["is_deleted"].Int(),
		DeletedAt:          nullableTimeFromRecord(row, "deleted_at"),
		CreatedAt:          parseRecordTime(row, "created_at"),
		UpdatedAt:          parseRecordTime(row, "updated_at"),
	}
}

func supplierPlatformListItemFromRecord(row gdb.Record) entity.SupplierPlatformListItem {
	lastBalance := formatMoney(nullableStringFromRecord(row, "last_balance").String)
	threshold := formatMoney(row["threshold_amount"].String())
	connectStatus := row["last_balance_status"].Int()
	return entity.SupplierPlatformListItem{
		ID:                 row["id"].Int64(),
		Name:               row["name"].String(),
		Domain:             row["domain"].String(),
		BackupDomain:       row["backup_domain"].String(),
		ProviderCode:       row["provider_code"].String(),
		ProviderName:       row["provider_name"].String(),
		TypeID:             row["type_id"].Int(),
		TypeName:           row["type_name"].String(),
		SubjectID:          row["subject_id"].Int64(),
		SubjectName:        row["subject_name"].String(),
		HasTax:             row["has_tax"].Int(),
		Status:             row["status"].Int(),
		LastBalance:        lastBalance,
		ThresholdAmount:    threshold,
		BalanceWarning:     balanceWarning(lastBalance, threshold),
		ConnectStatus:      connectStatus,
		ConnectStatusText:  connectStatusText(connectStatus),
		LastBalanceMessage: row["last_balance_message"].String(),
		LastBalanceAt:      formatNullableTime(nullableTimeFromRecord(row, "last_balance_at")),
		Sort:               row["sort"].Int(),
		CrowdName:          row["crowd_name"].String(),
	}
}

func nullableStringFromRecord(row gdb.Record, key string) sql.NullString {
	value, ok := row[key]
	if !ok || value == nil || value.IsNil() {
		return sql.NullString{}
	}
	return sql.NullString{String: strings.TrimSpace(value.String()), Valid: true}
}

func connectStatusText(status int) string {
	switch status {
	case 1:
		return "正常"
	case 2:
		return "异常"
	default:
		return "未验证"
	}
}

func balanceWarning(lastBalance, threshold string) int {
	if strings.TrimSpace(lastBalance) == "" {
		return 0
	}
	thresholdAmount, err := decimal.NewFromString(strings.TrimSpace(threshold))
	if err != nil || thresholdAmount.LessThanOrEqual(decimal.Zero) {
		return 0
	}
	lastAmount, err := decimal.NewFromString(strings.TrimSpace(lastBalance))
	if err != nil {
		return 0
	}
	if lastAmount.LessThan(thresholdAmount) {
		return 1
	}
	return 0
}
