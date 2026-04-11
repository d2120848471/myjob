package app

import (
	"admin/utility/ipx"
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (a *Application) handleSMSConfigGet(w http.ResponseWriter, r *http.Request, _ principal, _ AdminUser) {
	state, err := a.loadSMSConfigState(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "短信配置读取失败")
		return
	}
	resp := SMSConfigGetResponse{
		AccessKeyMasked:           ipx.MaskAccessKey(state.Config.AccessKey),
		AccessKeySecretMasked:     ipx.MaskSecret(state.Config.AccessKeySecret),
		AccessKeyConfigured:       state.AccessKeyConfigured,
		AccessKeySecretConfigured: state.AccessKeySecretConfigured,
		SignName:                  state.Config.SignName,
		TemplateCode:              state.Config.TemplateCode,
		ExpireMinutes:             state.Config.ExpireMinutes,
		IntervalMinutes:           state.Config.IntervalMinutes,
	}
	if !state.UpdatedAt.IsZero() {
		resp.UpdatedAt = state.UpdatedAt.In(time.Local).Format("2006-01-02 15:04:05")
	}
	writeSuccess(w, resp)
}

func (a *Application) handleSMSConfigSave(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	var req SMSConfigSaveRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, 400, "参数错误")
		return
	}
	state, err := a.loadSMSConfigState(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "短信配置读取失败")
		return
	}
	finalCfg, logDesc, err := mergeSMSConfigSave(state, req)
	if err != nil {
		writeError(w, http.StatusBadRequest, 400, err.Error())
		return
	}
	if err = a.saveSMSConfig(r.Context(), finalCfg); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "短信配置保存失败")
		return
	}
	_ = a.redis.Del(r.Context(), smsConfigCacheKey()).Err()
	a.writeOperation(r.Context(), actor, logDesc, requestIP(r))
	writeSuccess(w, map[string]interface{}{})
}

func mergeSMSConfigSave(current smsConfigState, req SMSConfigSaveRequest) (SMSConfig, string, error) {
	req.AccessKey = strings.TrimSpace(req.AccessKey)
	req.AccessKeySecret = strings.TrimSpace(req.AccessKeySecret)
	req.SignName = strings.TrimSpace(req.SignName)
	req.TemplateCode = strings.TrimSpace(req.TemplateCode)

	if req.SignName == "" || req.TemplateCode == "" {
		return SMSConfig{}, "", &apiError{HTTPStatus: http.StatusBadRequest, Code: 400, Message: "短信配置不能为空"}
	}
	if req.ExpireMinutes < 1 || req.ExpireMinutes > 60 || req.IntervalMinutes < 1 || req.IntervalMinutes > 10 {
		return SMSConfig{}, "", &apiError{HTTPStatus: http.StatusBadRequest, Code: 400, Message: "短信配置范围错误"}
	}

	finalCfg := current.Config
	finalCfg.SignName = req.SignName
	finalCfg.TemplateCode = req.TemplateCode
	finalCfg.ExpireMinutes = req.ExpireMinutes
	finalCfg.IntervalMinutes = req.IntervalMinutes

	// 首次配置必须显式提交两把密钥，避免把“保留旧值”误解释为可用数据。
	if !current.AccessKeyConfigured && req.KeepAccessKey {
		return SMSConfig{}, "", &apiError{HTTPStatus: http.StatusBadRequest, Code: 400, Message: "首次配置必须填写AccessKey"}
	}
	if !current.AccessKeySecretConfigured && req.KeepAccessKeySecret {
		return SMSConfig{}, "", &apiError{HTTPStatus: http.StatusBadRequest, Code: 400, Message: "首次配置必须填写AccessKeySecret"}
	}

	accessKeyChanged := false
	if req.KeepAccessKey {
		if !current.AccessKeyConfigured {
			return SMSConfig{}, "", &apiError{HTTPStatus: http.StatusBadRequest, Code: 400, Message: "当前没有可保留的AccessKey"}
		}
	} else {
		if req.AccessKey == "" {
			return SMSConfig{}, "", &apiError{HTTPStatus: http.StatusBadRequest, Code: 400, Message: "请输入AccessKey"}
		}
		finalCfg.AccessKey = req.AccessKey
		accessKeyChanged = req.AccessKey != current.Config.AccessKey
	}

	accessKeySecretChanged := false
	if req.KeepAccessKeySecret {
		if !current.AccessKeySecretConfigured {
			return SMSConfig{}, "", &apiError{HTTPStatus: http.StatusBadRequest, Code: 400, Message: "当前没有可保留的AccessKeySecret"}
		}
	} else {
		if req.AccessKeySecret == "" {
			return SMSConfig{}, "", &apiError{HTTPStatus: http.StatusBadRequest, Code: 400, Message: "请输入AccessKeySecret"}
		}
		finalCfg.AccessKeySecret = req.AccessKeySecret
		accessKeySecretChanged = req.AccessKeySecret != current.Config.AccessKeySecret
	}

	businessChanged := finalCfg.SignName != current.Config.SignName ||
		finalCfg.TemplateCode != current.Config.TemplateCode ||
		finalCfg.ExpireMinutes != current.Config.ExpireMinutes ||
		finalCfg.IntervalMinutes != current.Config.IntervalMinutes

	return finalCfg, buildSMSConfigLog(accessKeyChanged, accessKeySecretChanged, businessChanged), nil
}

func (a *Application) saveSMSConfig(ctx context.Context, cfg SMSConfig) error {
	tx, err := a.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	values := map[string]string{
		"sms_access_key":        cfg.AccessKey,
		"sms_access_key_secret": cfg.AccessKeySecret,
		"sms_sign_name":         cfg.SignName,
		"sms_template_code":     cfg.TemplateCode,
		"sms_expire_minutes":    strconv.Itoa(cfg.ExpireMinutes),
		"sms_interval_minutes":  strconv.Itoa(cfg.IntervalMinutes),
	}
	for key, value := range values {
		var exists int
		if err = tx.GetContext(ctx, &exists, `SELECT COUNT(*) FROM system_config WHERE config_key = ?`, key); err != nil {
			return err
		}
		if exists == 0 {
			if _, err = tx.ExecContext(ctx, `INSERT INTO system_config (config_key, config_value, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, key, value, smsConfigDescription(key), a.now(), a.now()); err != nil {
				return err
			}
			continue
		}
		if _, err = tx.ExecContext(ctx, `UPDATE system_config SET config_value = ?, updated_at = ? WHERE config_key = ?`, value, a.now(), key); err != nil {
			return err
		}
	}
	return tx.Commit()
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
