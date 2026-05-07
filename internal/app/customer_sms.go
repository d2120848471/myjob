package app

import "fmt"

const (
	CustomerSMSSceneRegister       = "register"
	CustomerSMSSceneForgotPassword = "forgot_password"
)

// CustomerSMSCodeKey 返回客户短信验证码 Redis key，按场景和手机号隔离。
func CustomerSMSCodeKey(scene, phone string) string {
	return fmt.Sprintf("customer:sms:%s:%s", scene, phone)
}

// CustomerSMSSendLockKey 返回客户短信发送频控锁 Redis key。
func CustomerSMSSendLockKey(scene, phone string) string {
	return fmt.Sprintf("customer:sms:send_lock:%s:%s", scene, phone)
}

// CustomerSMSAttemptsKey 返回客户短信验证码错误次数 Redis key。
func CustomerSMSAttemptsKey(scene, phone string) string {
	return fmt.Sprintf("customer:sms:attempts:%s:%s", scene, phone)
}
