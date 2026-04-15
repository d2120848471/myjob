package adminlogic

import (
	"context"
	"errors"
	"strings"
	"time"

	adminapi "myjob/api"
	"myjob/internal/app"
	"myjob/internal/consts"
	authlib "myjob/internal/library/auth"
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/database/gdb"
)

// SystemConfigLogic 提供系统参数配置管理相关业务能力。
type SystemConfigLogic struct{ core *app.Core }

type systemConfigMutation struct {
	Group string
	Spec  app.SystemConfigSpec
	Value string
}

// Get 读取系统参数配置：
// - group 为空：返回所有分组
// - group 非空：返回指定分组
func (l *SystemConfigLogic) Get(ctx context.Context, req *adminapi.SettingsSystemGetReq) (*adminapi.SettingsSystemGetRes, error) {
	group := strings.TrimSpace(req.Group)
	if group == "" {
		states, err := l.core.LoadAllSystemConfigGroups(ctx)
		if err != nil {
			return nil, apiErr(consts.CodeInternalError, "系统参数读取失败")
		}
		resp := &adminapi.SettingsSystemGetRes{
			Groups: make([]adminapi.SettingsSystemGroup, 0, len(states)),
		}
		for _, state := range states {
			resp.Groups = append(resp.Groups, toSystemGroupResponse(state))
		}
		return resp, nil
	}

	state, err := l.core.LoadSystemConfigGroup(ctx, group)
	if err != nil {
		if errors.Is(err, app.ErrUnknownSystemConfigGroup) {
			return nil, apiErr(consts.CodeBadRequest, "系统参数分组不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "系统参数读取失败")
	}

	groupResp := toSystemGroupResponse(state)
	return &adminapi.SettingsSystemGetRes{
		Group:  groupResp.Group,
		Label:  groupResp.Label,
		Items:  groupResp.Items,
		Groups: []adminapi.SettingsSystemGroup{groupResp},
	}, nil
}

// Save 保存系统参数配置，完成落库、清理缓存并写入操作日志。
func (l *SystemConfigLogic) Save(ctx context.Context, req *adminapi.SettingsSystemSaveReq, actor app.AdminUser, ip string) (*adminapi.SettingsSystemSaveRes, error) {
	mutations, groups, err := normalizeSystemSaveRequest(req)
	if err != nil {
		return nil, err
	}
	if err := saveSystemConfigMutations(ctx, l.core, mutations); err != nil {
		return nil, apiErr(consts.CodeInternalError, "系统参数保存失败")
	}
	// 保存后按分组删除缓存，确保读取走最新配置。
	cacheKeys := make([]string, 0, len(groups))
	for _, group := range groups {
		cacheKeys = append(cacheKeys, authlib.SystemConfigCacheKey(group))
	}
	if len(cacheKeys) > 0 {
		_, _ = l.core.Redis().GroupGeneric().Del(ctx, cacheKeys...)
	}
	l.core.WriteOperation(ctx, actor, buildSystemConfigLog(groups, mutations), ip)
	return &adminapi.SettingsSystemSaveRes{}, nil
}

func saveSystemConfigMutations(ctx context.Context, core *app.Core, items []systemConfigMutation) error {
	return core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		now := core.Now()
		for _, item := range items {
			exists, err := tx.GetValue(`SELECT COUNT(*) FROM system_config WHERE config_key = ?`, item.Spec.StorageKey)
			if err != nil {
				return err
			}
			if exists.Int() == 0 {
				if _, err = tx.Exec(`INSERT INTO system_config (config_key, config_value, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, item.Spec.StorageKey, item.Value, item.Spec.Label, now, now); err != nil {
					return err
				}
				continue
			}
			if _, err = tx.Exec(`UPDATE system_config SET config_value = ?, description = ?, updated_at = ? WHERE config_key = ?`, item.Value, item.Spec.Label, now, item.Spec.StorageKey); err != nil {
				return err
			}
		}
		return nil
	})
}

func buildSystemConfigLog(groups []string, items []systemConfigMutation) string {
	groupLabels := make([]string, 0, len(groups))
	for _, group := range groups {
		groupLabels = append(groupLabels, app.SystemConfigGroupLabel(group))
	}
	labels := make([]string, 0, len(items))
	for _, item := range items {
		labels = append(labels, item.Spec.Label)
	}
	return "更新系统参数(" + strings.Join(groupLabels, "、") + ")：" + strings.Join(labels, "、")
}

func normalizeSystemSaveRequest(req *adminapi.SettingsSystemSaveReq) ([]systemConfigMutation, []string, error) {
	saveGroups := req.Groups
	if len(saveGroups) == 0 {
		group := strings.TrimSpace(req.Group)
		if group == "" {
			return nil, nil, apiErr(consts.CodeBadRequest, "系统参数分组不能为空")
		}
		saveGroups = []adminapi.SettingsSystemSaveGroup{{
			Group: group,
			Items: req.Items,
		}}
	}

	mutations := make([]systemConfigMutation, 0)
	normalizedGroups := make([]string, 0, len(saveGroups))
	seenGroups := make(map[string]struct{}, len(saveGroups))
	for _, groupReq := range saveGroups {
		groupMutations, normalizedGroup, err := normalizeSystemSaveGroup(groupReq)
		if err != nil {
			return nil, nil, err
		}
		if _, exists := seenGroups[normalizedGroup]; exists {
			return nil, nil, apiErr(consts.CodeBadRequest, "系统参数分组重复")
		}
		mutations = append(mutations, groupMutations...)
		normalizedGroups = append(normalizedGroups, normalizedGroup)
		seenGroups[normalizedGroup] = struct{}{}
	}
	return mutations, normalizedGroups, nil
}

func normalizeSystemSaveGroup(groupReq adminapi.SettingsSystemSaveGroup) ([]systemConfigMutation, string, error) {
	group := strings.TrimSpace(groupReq.Group)
	if group == "" {
		return nil, "", apiErr(consts.CodeBadRequest, "系统参数分组不能为空")
	}
	specs, ok := app.SystemConfigSpecs(group)
	if !ok {
		return nil, "", apiErr(consts.CodeBadRequest, "系统参数分组不存在")
	}
	if len(groupReq.Items) == 0 {
		return nil, "", apiErr(consts.CodeBadRequest, "系统参数不能为空")
	}

	required := make(map[string]app.SystemConfigSpec, len(specs))
	for _, spec := range specs {
		if spec.Required {
			required[spec.Key] = spec
		}
	}

	mutations := make([]systemConfigMutation, 0, len(groupReq.Items))
	seen := make(map[string]struct{}, len(groupReq.Items))
	for _, item := range groupReq.Items {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			return nil, "", apiErr(consts.CodeBadRequest, "系统参数键不能为空")
		}
		if _, exists := seen[key]; exists {
			return nil, "", apiErr(consts.CodeBadRequest, "系统参数键重复")
		}
		spec, exists := app.LookupSystemConfigSpec(group, key)
		if !exists {
			return nil, "", apiErr(consts.CodeBadRequest, "系统参数项不存在")
		}
		value := strings.TrimSpace(item.Value)
		if spec.Required && value == "" {
			return nil, "", apiErr(consts.CodeBadRequest, spec.Label+"不能为空")
		}
		if value != "" && spec.Validator != nil {
			if err := spec.Validator(value); err != nil {
				return nil, "", apiErr(consts.CodeBadRequest, err.Error())
			}
		}
		mutations = append(mutations, systemConfigMutation{
			Group: group,
			Spec:  spec,
			Value: value,
		})
		seen[key] = struct{}{}
	}

	for key, spec := range required {
		if _, exists := seen[key]; !exists {
			return nil, "", apiErr(consts.CodeBadRequest, spec.Label+"不能为空")
		}
	}
	return mutations, group, nil
}

func toSystemGroupResponse(state modelruntime.SystemConfigGroupState) adminapi.SettingsSystemGroup {
	resp := adminapi.SettingsSystemGroup{
		Group: state.Group,
		Label: state.Label,
		Items: make([]adminapi.SettingsSystemItem, 0, len(state.Items)),
	}
	for _, item := range state.Items {
		apiItem := adminapi.SettingsSystemItem{
			Key:        item.Key,
			Label:      item.Label,
			Value:      item.Value,
			ValueType:  item.ValueType,
			Unit:       item.Unit,
			Required:   item.Required,
			Configured: item.Configured,
		}
		if !item.UpdatedAt.IsZero() {
			apiItem.UpdatedAt = item.UpdatedAt.In(time.Local).Format("2006-01-02 15:04:05")
		}
		resp.Items = append(resp.Items, apiItem)
	}
	return resp
}
