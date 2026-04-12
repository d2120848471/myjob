INSERT INTO system_config (config_key, config_value, description, created_at, updated_at) VALUES
('sms_access_key', '', '阿里云 AccessKey', NOW(), NOW()),
('sms_access_key_secret', '', '阿里云 AccessKey Secret', NOW(), NOW()),
('sms_sign_name', '玖权益', '短信签名', NOW(), NOW()),
('sms_template_code', 'SMS_000001', '短信模板编号', NOW(), NOW()),
('sms_expire_minutes', '30', '验证码有效期', NOW(), NOW()),
('sms_interval_minutes', '1', '验证码发送间隔', NOW(), NOW())
ON DUPLICATE KEY UPDATE config_value = VALUES(config_value), updated_at = VALUES(updated_at);
