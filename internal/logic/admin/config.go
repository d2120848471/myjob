package adminlogic

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	configv1 "myjob/api/admin/config/v1"
	"myjob/internal/kernel"
	authlib "myjob/internal/library/auth"
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/database/gdb"
)

type SMSConfigLogic struct{ core *kernel.Core }

func (l *SMSConfigLogic) Get(ctx context.Context) (configv1.SMSConfigGetRes, *modelruntime.APIError) {
	state, err := l.core.LoadSMSConfigState(ctx)
	if err != nil {
		return configv1.SMSConfigGetRes{}, apiErr(http.StatusInternalServerError, 500, "短信配置读取失败")
	}
	resp := configv1.SMSConfigGetRes{AccessKeyMasked: kernel.MaskAccessKey(state.Config.AccessKey), AccessKeySecretMasked: kernel.MaskSecret(state.Config.AccessKeySecret), AccessKeyConfigured: state.AccessKeyConfigured, AccessKeySecretConfigured: state.AccessKeySecretConfigured, SignName: state.Config.SignName, TemplateCode: state.Config.TemplateCode, ExpireMinutes: state.Config.ExpireMinutes, IntervalMinutes: state.Config.IntervalMinutes}
	if !state.UpdatedAt.IsZero() {
		resp.UpdatedAt = state.UpdatedAt.In(time.Local).Format("2006-01-02 15:04:05")
	}
	return resp, nil
}

func (l *SMSConfigLogic) Save(ctx context.Context, req configv1.SMSConfigSaveReq, actor kernel.AdminUser, ip string) (map[string]any, *modelruntime.APIError) {
	state, err := l.core.LoadSMSConfigState(ctx)
	if err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "短信配置读取失败")
	}
	finalCfg, logDesc, apiErrValue := mergeSMSConfigSave(state, req)
	if apiErrValue != nil {
		return nil, apiErrValue
	}
	if err = saveSMSConfig(ctx, l.core, finalCfg); err != nil {
		return nil, apiErr(http.StatusInternalServerError, 500, "短信配置保存失败")
	}
	_ = l.core.Redis().Del(ctx, authlib.SMSConfigCacheKey()).Err()
	l.core.WriteOperation(ctx, actor, logDesc, ip)
	return map[string]any{}, nil
}

func mergeSMSConfigSave(current modelruntime.SMSConfigState, req configv1.SMSConfigSaveReq) (modelruntime.SMSConfig, string, *modelruntime.APIError) {
	req.AccessKey = strings.TrimSpace(req.AccessKey)
	req.AccessKeySecret = strings.TrimSpace(req.AccessKeySecret)
	req.SignName = strings.TrimSpace(req.SignName)
	req.TemplateCode = strings.TrimSpace(req.TemplateCode)
	if req.SignName == "" || req.TemplateCode == "" {
		return modelruntime.SMSConfig{}, "", apiErr(http.StatusBadRequest, 400, "短信配置不能为空")
	}
	if req.ExpireMinutes < 1 || req.ExpireMinutes > 60 || req.IntervalMinutes < 1 || req.IntervalMinutes > 10 {
		return modelruntime.SMSConfig{}, "", apiErr(http.StatusBadRequest, 400, "短信配置范围错误")
	}
	finalCfg := current.Config
	finalCfg.SignName = req.SignName
	finalCfg.TemplateCode = req.TemplateCode
	finalCfg.ExpireMinutes = req.ExpireMinutes
	finalCfg.IntervalMinutes = req.IntervalMinutes
	if !current.AccessKeyConfigured && req.KeepAccessKey {
		return modelruntime.SMSConfig{}, "", apiErr(http.StatusBadRequest, 400, "首次配置必须填写AccessKey")
	}
	if !current.AccessKeySecretConfigured && req.KeepAccessKeySecret {
		return modelruntime.SMSConfig{}, "", apiErr(http.StatusBadRequest, 400, "首次配置必须填写AccessKeySecret")
	}
	accessKeyChanged := false
	if req.KeepAccessKey {
		if !current.AccessKeyConfigured {
			return modelruntime.SMSConfig{}, "", apiErr(http.StatusBadRequest, 400, "当前没有可保留的AccessKey")
		}
	} else {
		if req.AccessKey == "" {
			return modelruntime.SMSConfig{}, "", apiErr(http.StatusBadRequest, 400, "请输入AccessKey")
		}
		finalCfg.AccessKey = req.AccessKey
		accessKeyChanged = req.AccessKey != current.Config.AccessKey
	}
	accessKeySecretChanged := false
	if req.KeepAccessKeySecret {
		if !current.AccessKeySecretConfigured {
			return modelruntime.SMSConfig{}, "", apiErr(http.StatusBadRequest, 400, "当前没有可保留的AccessKeySecret")
		}
	} else {
		if req.AccessKeySecret == "" {
			return modelruntime.SMSConfig{}, "", apiErr(http.StatusBadRequest, 400, "请输入AccessKeySecret")
		}
		finalCfg.AccessKeySecret = req.AccessKeySecret
		accessKeySecretChanged = req.AccessKeySecret != current.Config.AccessKeySecret
	}
	businessChanged := finalCfg.SignName != current.Config.SignName || finalCfg.TemplateCode != current.Config.TemplateCode || finalCfg.ExpireMinutes != current.Config.ExpireMinutes || finalCfg.IntervalMinutes != current.Config.IntervalMinutes
	return finalCfg, buildSMSConfigLog(accessKeyChanged, accessKeySecretChanged, businessChanged), nil
}

func saveSMSConfig(ctx context.Context, core *kernel.Core, cfg modelruntime.SMSConfig) error {
	return core.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		values := map[string]string{"sms_access_key": cfg.AccessKey, "sms_access_key_secret": cfg.AccessKeySecret, "sms_sign_name": cfg.SignName, "sms_template_code": cfg.TemplateCode, "sms_expire_minutes": strconv.Itoa(cfg.ExpireMinutes), "sms_interval_minutes": strconv.Itoa(cfg.IntervalMinutes)}
		for key, value := range values {
			exists, err := tx.GetValue(`SELECT COUNT(*) FROM system_config WHERE config_key = ?`, key)
			if err != nil {
				return err
			}
			if exists.Int() == 0 {
				if _, err = tx.Exec(`INSERT INTO system_config (config_key, config_value, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, key, value, smsConfigDescription(key), core.Now(), core.Now()); err != nil {
					return err
				}
				continue
			}
			if _, err = tx.Exec(`UPDATE system_config SET config_value = ?, updated_at = ? WHERE config_key = ?`, value, core.Now(), key); err != nil {
				return err
			}
		}
		return nil
	})
}

func buildSMSConfigLog(accessKeyChanged, accessKeySecretChanged, businessChanged bool) string {
	parts := []string{"更新短信配置"}
	if accessKeyChanged {
		parts = append(parts, "更换AccessKey")
	}
	if accessKeySecretChanged {
		parts = append(parts, "更换AccessKeySecret")
	}
	if businessChanged {
		parts = append(parts, "更新签名模板和时效")
	}
	return strings.Join(parts, "；")
}

func smsConfigDescription(key string) string {
	switch key {
	case "sms_access_key":
		return "阿里云 AccessKey"
	case "sms_access_key_secret":
		return "阿里云 AccessKey Secret"
	case "sms_sign_name":
		return "短信签名"
	case "sms_template_code":
		return "短信模板编号"
	case "sms_expire_minutes":
		return "验证码有效期"
	case "sms_interval_minutes":
		return "验证码发送间隔"
	default:
		return key
	}
}
