package app

import (
	"admin/utility/ipx"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (a *Application) handleSMSConfigGet(w http.ResponseWriter, r *http.Request, _ principal, _ AdminUser) {
	cfg, err := a.loadSMSConfig(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "短信配置读取失败")
		return
	}
	writeSuccess(w, map[string]interface{}{
		"access_key":        ipx.MaskSecret(cfg.AccessKey),
		"access_key_secret": ipx.MaskSecret(cfg.AccessKeySecret),
		"sign_name":         cfg.SignName,
		"template_code":     cfg.TemplateCode,
		"expire_minutes":    cfg.ExpireMinutes,
		"interval_minutes":  cfg.IntervalMinutes,
	})
}

func (a *Application) handleSMSConfigSave(w http.ResponseWriter, r *http.Request, _ principal, actor AdminUser) {
	var req SMSConfig
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, 400, "参数错误")
		return
	}
	if strings.TrimSpace(req.AccessKey) == "" || strings.TrimSpace(req.AccessKeySecret) == "" || strings.TrimSpace(req.SignName) == "" || strings.TrimSpace(req.TemplateCode) == "" {
		writeError(w, http.StatusBadRequest, 400, "短信配置不能为空")
		return
	}
	if req.ExpireMinutes < 1 || req.ExpireMinutes > 60 || req.IntervalMinutes < 1 || req.IntervalMinutes > 10 {
		writeError(w, http.StatusBadRequest, 400, "短信配置范围错误")
		return
	}
	tx, err := a.db.BeginTxx(r.Context(), nil)
	if err != nil {
		writeError(w, http.StatusInternalServerError, 500, "事务开启失败")
		return
	}
	defer tx.Rollback()
	values := map[string]string{
		"sms_access_key":        req.AccessKey,
		"sms_access_key_secret": req.AccessKeySecret,
		"sms_sign_name":         req.SignName,
		"sms_template_code":     req.TemplateCode,
		"sms_expire_minutes":    strconv.Itoa(req.ExpireMinutes),
		"sms_interval_minutes":  strconv.Itoa(req.IntervalMinutes),
	}
	for key, value := range values {
		var exists int
		if err = tx.GetContext(r.Context(), &exists, `SELECT COUNT(*) FROM system_config WHERE config_key = ?`, key); err != nil {
			writeError(w, http.StatusInternalServerError, 500, "短信配置保存失败")
			return
		}
		if exists == 0 {
			if _, err = tx.ExecContext(r.Context(), `INSERT INTO system_config (config_key, config_value, description, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`, key, value, key, a.now(), a.now()); err != nil {
				writeError(w, http.StatusInternalServerError, 500, "短信配置保存失败")
				return
			}
		} else {
			if _, err = tx.ExecContext(r.Context(), `UPDATE system_config SET config_value = ?, updated_at = ? WHERE config_key = ?`, value, a.now(), key); err != nil {
				writeError(w, http.StatusInternalServerError, 500, "短信配置保存失败")
				return
			}
		}
	}
	if err = tx.Commit(); err != nil {
		writeError(w, http.StatusInternalServerError, 500, "短信配置保存失败")
		return
	}
	_ = a.redis.Del(r.Context(), smsConfigCacheKey()).Err()
	a.writeOperation(r.Context(), actor, fmt.Sprintf("更新短信配置：签名=%s，模板=%s，有效期=%d，间隔=%d", req.SignName, req.TemplateCode, req.ExpireMinutes, req.IntervalMinutes), requestIP(r))
	writeSuccess(w, map[string]interface{}{})
}
