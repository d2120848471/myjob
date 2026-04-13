package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	authlib "myjob/internal/library/auth"
	modelruntime "myjob/internal/model/runtime"
)

const (
	systemConfigCacheVersion = 1
	systemConfigRateScale    = 4
)

var (
	ErrUnknownSystemConfigGroup  = errors.New("system config group not found")
	ErrSystemConfigNotConfigured = errors.New("system config not configured")
)

type SystemConfigSpec struct {
	Group        string
	Key          string
	StorageKey   string
	Label        string
	ValueType    string
	Unit         string
	Required     bool
	DefaultValue string
	Validator    func(string) error
}

type systemConfigGroupDefinition struct {
	Label string
	Items []SystemConfigSpec
}

var systemConfigRegistry = map[string]systemConfigGroupDefinition{
	"finance": {
		Label: "财务参数",
		Items: []SystemConfigSpec{
			{
				Group:      "finance",
				Key:        "tax_exclusive_rate",
				StorageKey: "finance_tax_exclusive_rate",
				Label:      "未税->含税税率",
				ValueType:  "decimal",
				Unit:       "%",
				Required:   true,
				Validator:  validateRateValue,
			},
			{
				Group:      "finance",
				Key:        "tax_inclusive_rate",
				StorageKey: "finance_tax_inclusive_rate",
				Label:      "含税->未税税率",
				ValueType:  "decimal",
				Unit:       "%",
				Required:   true,
				Validator:  validateRateValue,
			},
		},
	},
	"integration": {
		Label: "集成参数",
		Items: []SystemConfigSpec{
			{
				Group:      "integration",
				Key:        "robot_webhook_url",
				StorageKey: "integration_robot_webhook_url",
				Label:      "机器人回调地址",
				ValueType:  "url",
				Validator:  validateRobotWebhookURL,
			},
		},
	},
}

func SystemConfigSpecs(group string) ([]SystemConfigSpec, bool) {
	definition, ok := systemConfigRegistry[strings.TrimSpace(group)]
	if !ok {
		return nil, false
	}
	items := make([]SystemConfigSpec, len(definition.Items))
	copy(items, definition.Items)
	return items, true
}

func SystemConfigGroupLabel(group string) string {
	if definition, ok := systemConfigRegistry[strings.TrimSpace(group)]; ok {
		return definition.Label
	}
	return group
}

func LookupSystemConfigSpec(group, key string) (SystemConfigSpec, bool) {
	definition, ok := systemConfigRegistry[strings.TrimSpace(group)]
	if !ok {
		return SystemConfigSpec{}, false
	}
	key = strings.TrimSpace(key)
	for _, item := range definition.Items {
		if item.Key == key {
			return item, true
		}
	}
	return SystemConfigSpec{}, false
}

func (c *Core) LoadSystemConfigGroup(ctx context.Context, group string) (modelruntime.SystemConfigGroupState, error) {
	group = strings.TrimSpace(group)
	definition, ok := systemConfigRegistry[group]
	if !ok {
		return modelruntime.SystemConfigGroupState{}, ErrUnknownSystemConfigGroup
	}
	if cached, err := c.RedisGetString(ctx, authlib.SystemConfigCacheKey(group)); err == nil {
		var state modelruntime.SystemConfigGroupState
		if json.Unmarshal([]byte(cached), &state) == nil && state.Version == systemConfigCacheVersion {
			return state, nil
		}
	}

	rows, err := c.DB().GetCore().GetAll(ctx, `SELECT config_key, config_value, updated_at FROM system_config WHERE config_key LIKE ?`, group+"_%")
	if err != nil {
		return modelruntime.SystemConfigGroupState{}, err
	}

	values := make(map[string]string, len(rows))
	updatedAtByKey := make(map[string]time.Time, len(rows))
	for _, row := range rows {
		storageKey := row["config_key"].String()
		values[storageKey] = strings.TrimSpace(row["config_value"].String())
		if updatedAt, ok := parseConfigUpdatedAt(row["updated_at"].Val()); ok {
			updatedAtByKey[storageKey] = updatedAt
		}
	}

	state := modelruntime.SystemConfigGroupState{
		Version: systemConfigCacheVersion,
		Group:   group,
		Items:   make([]modelruntime.SystemConfigItem, 0, len(definition.Items)),
	}
	for _, spec := range definition.Items {
		item := modelruntime.SystemConfigItem{
			Key:       spec.Key,
			Label:     spec.Label,
			ValueType: spec.ValueType,
			Unit:      spec.Unit,
			Required:  spec.Required,
		}
		if value, exists := values[spec.StorageKey]; exists {
			item.Value = value
			item.Configured = value != ""
			item.UpdatedAt = updatedAtByKey[spec.StorageKey]
		} else if spec.DefaultValue != "" {
			item.Value = spec.DefaultValue
		}
		state.Items = append(state.Items, item)
	}

	data, _ := json.Marshal(state)
	_ = c.RedisSetString(ctx, authlib.SystemConfigCacheKey(group), string(data), 30*time.Minute)
	return state, nil
}

func (c *Core) LoadFinanceTaxConfig(ctx context.Context) (modelruntime.FinanceTaxConfig, error) {
	state, err := c.LoadSystemConfigGroup(ctx, "finance")
	if err != nil {
		return modelruntime.FinanceTaxConfig{}, err
	}

	items := make(map[string]modelruntime.SystemConfigItem, len(state.Items))
	for _, item := range state.Items {
		items[item.Key] = item
	}

	exclusive, ok := items["tax_exclusive_rate"]
	if !ok || !exclusive.Configured {
		return modelruntime.FinanceTaxConfig{}, fmt.Errorf("未税->含税税率未配置: %w", ErrSystemConfigNotConfigured)
	}
	inclusive, ok := items["tax_inclusive_rate"]
	if !ok || !inclusive.Configured {
		return modelruntime.FinanceTaxConfig{}, fmt.Errorf("含税->未税税率未配置: %w", ErrSystemConfigNotConfigured)
	}

	if err := validateRateValue(exclusive.Value); err != nil {
		return modelruntime.FinanceTaxConfig{}, err
	}
	if err := validateRateValue(inclusive.Value); err != nil {
		return modelruntime.FinanceTaxConfig{}, err
	}
	exclusiveScaled, err := parseDecimalToScaled(exclusive.Value, systemConfigRateScale)
	if err != nil {
		return modelruntime.FinanceTaxConfig{}, err
	}
	inclusiveScaled, err := parseDecimalToScaled(inclusive.Value, systemConfigRateScale)
	if err != nil {
		return modelruntime.FinanceTaxConfig{}, err
	}

	return modelruntime.FinanceTaxConfig{
		TaxExclusiveRate:       exclusive.Value,
		TaxExclusiveRateScaled: exclusiveScaled,
		TaxInclusiveRate:       inclusive.Value,
		TaxInclusiveRateScaled: inclusiveScaled,
	}, nil
}

func validateRateValue(value string) error {
	scaled, err := parseDecimalToScaled(value, systemConfigRateScale)
	if err != nil {
		return err
	}
	limit := int64(100) * pow10(systemConfigRateScale)
	if scaled <= 0 || scaled >= limit {
		return errors.New("税率必须大于0且小于100")
	}
	return nil
}

func validateRobotWebhookURL(value string) error {
	if len(value) > 512 {
		return errors.New("机器人回调地址长度不能超过512")
	}
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed.Host == "" {
		return errors.New("机器人回调地址格式错误")
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("机器人回调地址格式错误")
	}
	return nil
}

func parseDecimalToScaled(value string, scale int) (int64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, errors.New("数值不能为空")
	}
	parts := strings.Split(value, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, errors.New("数值格式错误")
	}
	if !digitsOnly(parts[0]) {
		return 0, errors.New("数值格式错误")
	}

	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
		if fraction == "" || !digitsOnly(fraction) {
			return 0, errors.New("数值格式错误")
		}
		if len(fraction) > scale {
			return 0, fmt.Errorf("数值最多支持%d位小数", scale)
		}
	}

	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, errors.New("数值格式错误")
	}

	scaled := whole * pow10(scale)
	if fraction != "" {
		padding := scale - len(fraction)
		if padding > 0 {
			fraction += strings.Repeat("0", padding)
		}
		fractionValue, parseErr := strconv.ParseInt(fraction, 10, 64)
		if parseErr != nil {
			return 0, errors.New("数值格式错误")
		}
		scaled += fractionValue
	}
	return scaled, nil
}

func digitsOnly(value string) bool {
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return value != ""
}

func pow10(scale int) int64 {
	result := int64(1)
	for i := 0; i < scale; i++ {
		result *= 10
	}
	return result
}
