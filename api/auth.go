package api

import (
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/frame/g"
)

// LoginUser 是登录后返回给前端的用户视图，包含当前登录态所需的基础信息。
type LoginUser = modelruntime.LoginUser

// AuthLoginReq 用于后台账号密码登录。
type AuthLoginReq struct {
	g.Meta   `path:"/auth/login" method:"post" tags:"认证" summary:"账号密码登录" dc:"后台账号密码登录"`
	Username string `json:"username" v:"required#请输入用户名" dc:"用户名"`
	Password string `json:"password" v:"required#请输入密码" dc:"密码"`
}

// AuthLoginRes 返回登录结果。
//
// 当 NeedSMSVerify=true 时，必须使用 LoginToken 继续完成短信二次验证后才会签发正式 Token。
type AuthLoginRes struct {
	NeedSMSVerify bool       `json:"need_sms_verify" dc:"是否需要短信二次验证"`
	LoginToken    string     `json:"login_token,omitempty" dc:"短信验证临时凭证"`
	Phone         string     `json:"phone,omitempty" dc:"脱敏手机号"`
	Reason        string     `json:"reason,omitempty" dc:"触发短信验证原因"`
	Token         string     `json:"token,omitempty" dc:"登录令牌"`
	User          *LoginUser `json:"user,omitempty" dc:"当前登录用户"`
	Permissions   []string   `json:"permissions,omitempty" dc:"权限码列表"`
}

// AuthSMSSendReq 用于在需要短信二验时发送登录验证码。
type AuthSMSSendReq struct {
	g.Meta     `path:"/auth/sms/send" method:"post" tags:"认证" summary:"发送登录验证码" dc:"登录短信二验发送验证码"`
	LoginToken string `json:"login_token" v:"required#login_token不能为空" dc:"登录临时凭证"`
}

// AuthSMSSendRes 表示验证码发送成功（返回体为空）。
type AuthSMSSendRes struct{}

// AuthSMSVerifyReq 用于校验短信验证码并换取正式登录 Token。
type AuthSMSVerifyReq struct {
	g.Meta     `path:"/auth/sms/verify" method:"post" tags:"认证" summary:"校验登录验证码" dc:"短信验证码登录二验"`
	LoginToken string `json:"login_token" v:"required#login_token不能为空" dc:"登录临时凭证"`
	SMSCode    string `json:"sms_code" v:"required#sms_code不能为空" dc:"短信验证码"`
}

// AuthSMSVerifyRes 返回正式登录 Token、当前用户与权限码列表。
type AuthSMSVerifyRes struct {
	Token       string    `json:"token" dc:"登录令牌"`
	User        LoginUser `json:"user" dc:"当前登录用户"`
	Permissions []string  `json:"permissions" dc:"权限码列表"`
}

// AuthMeReq 用于读取当前登录用户信息与权限码列表。
type AuthMeReq struct {
	g.Meta `path:"/auth/me" method:"get" tags:"认证" summary:"获取当前登录信息" security:"BearerAuth" dc:"获取当前登录用户和权限信息"`
}

// AuthMeRes 返回当前登录用户信息与权限码列表。
type AuthMeRes struct {
	User        LoginUser `json:"user" dc:"当前登录用户"`
	Permissions []string  `json:"permissions" dc:"权限码列表"`
}

// AuthSessionDeleteReq 用于退出当前登录会话（注销）。
type AuthSessionDeleteReq struct {
	g.Meta `path:"/auth/session" method:"delete" tags:"认证" summary:"退出登录" security:"BearerAuth" dc:"退出当前登录会话"`
}

// AuthSessionDeleteRes 表示退出登录成功（返回体为空）。
type AuthSessionDeleteRes struct{}
