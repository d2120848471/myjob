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

	"github.com/gogf/gf/v2/database/gdb"
)

type SystemConfigLogic struct{ core *app.Core }

type systemConfigMutation struct {
	Spec  app.SystemConfigSpec
	Value string
}

func (l *SystemConfigLogic) Get(ctx context.Context, req *adminapi.SettingsSystemGetReq) (*adminapi.SettingsSystemGetRes, error) {
	state, err := l.core.LoadSystemConfigGroup(ctx, req.Group)
	if err != nil {
		if errors.Is(err, app.ErrUnknownSystemConfigGroup) {
			return nil, apiErr(consts.CodeBadRequest, "系统参数分组不存在")
		}
		return nil, apiErr(consts.CodeInternalError, "系统参数读取失败")
	}

	resp := &adminapi.SettingsSystemGetRes{
		Group: state.Group,
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
	return resp, nil
}

func (l *SystemConfigLogic) Save(ctx context.Context, req *adminapi.SettingsSystemSaveReq, actor app.AdminUser, ip string) (*adminapi.SettingsSystemSaveRes, error) {
	group := strings.TrimSpace(req.Group)
	if group == "" {
		return nil, apiErr(consts.CodeBadRequest, "系统参数分组不能为空")
	}
	specs, ok := app.SystemConfigSpecs(group)
	if !ok {
		return nil, apiErr(consts.CodeBadRequest, "系统参数分组不存在")
	}
	if len(req.Items) == 0 {
		return nil, apiErr(consts.CodeBadRequest, "系统参数不能为空")
	}

	// 先整批校验，再统一落库，保证同一组参数要么全成功要么全失败。
	required := make(map[string]app.SystemConfigSpec, len(specs))
	for _, spec := range specs {
		if spec.Required {
			required[spec.Key] = spec
		}
	}

	mutations := make([]systemConfigMutation, 0, len(req.Items))
	seen := make(map[string]struct{}, len(req.Items))
	for _, item := range req.Items {
		key := strings.TrimSpace(item.Key)
		if key == "" {
			return nil, apiErr(consts.CodeBadRequest, "系统参数键不能为空")
		}
		if _, exists := seen[key]; exists {
			return nil, apiErr(consts.CodeBadRequest, "系统参数键重复")
		}
		spec, exists := app.LookupSystemConfigSpec(group, key)
		if !exists {
			return nil, apiErr(consts.CodeBadRequest, "系统参数项不存在")
		}

		value := strings.TrimSpace(item.Value)
		if spec.Required && value == "" {
			return nil, apiErr(consts.CodeBadRequest, spec.Label+"不能为空")
		}
		if value != "" && spec.Validator != nil {
			if err := spec.Validator(value); err != nil {
				return nil, apiErr(consts.CodeBadRequest, err.Error())
			}
		}

		mutations = append(mutations, systemConfigMutation{Spec: spec, Value: value})
		seen[key] = struct{}{}
	}

	for key, spec := range required {
		if _, exists := seen[key]; !exists {
			return nil, apiErr(consts.CodeBadRequest, spec.Label+"不能为空")
		}
	}
	if err := saveSystemConfigMutations(ctx, l.core, mutations); err != nil {
		return nil, apiErr(consts.CodeInternalError, "系统参数保存失败")
	}
	_, _ = l.core.Redis().GroupGeneric().Del(ctx, authlib.SystemConfigCacheKey(group))
	l.core.WriteOperation(ctx, actor, buildSystemConfigLog(group, mutations), ip)
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

func buildSystemConfigLog(group string, items []systemConfigMutation) string {
	labels := make([]string, 0, len(items))
	for _, item := range items {
		labels = append(labels, item.Spec.Label)
	}
	return "更新" + app.SystemConfigGroupLabel(group) + "：" + strings.Join(labels, "、")
}
