package api

import (
	modelruntime "myjob/internal/model/runtime"

	"github.com/gogf/gf/v2/frame/g"
)

// CustomerLoginUser 是客户登录后返回给前端的基础账号视图。
type CustomerLoginUser = modelruntime.CustomerLoginUser

// CustomerAuthSMSSendReq 用于客户注册或找回密码时发送短信验证码。
type CustomerAuthSMSSendReq struct {
	g.Meta `path:"/auth/sms/send" method:"post" tags:"客户认证" summary:"发送客户验证码" dc:"客户注册或找回密码发送验证码"`
	Phone  string `json:"phone" v:"required#请输入手机号" dc:"手机号"`
	Scene  string `json:"scene" v:"required#验证码场景不能为空" dc:"验证码场景：register/forgot_password"`
}

// CustomerAuthSMSSendRes 表示客户验证码发送成功（返回体为空）。
type CustomerAuthSMSSendRes struct{}

// CustomerRegisterReq 用于客户手机号验证码注册。
type CustomerRegisterReq struct {
	g.Meta             `path:"/auth/register" method:"post" tags:"客户认证" summary:"客户注册" dc:"客户手机号验证码注册"`
	CompanyName        string `json:"company_name" v:"required#请输入店铺或公司名称" dc:"店铺或公司名称"`
	Phone              string `json:"phone" v:"required#请输入手机号" dc:"手机号"`
	SMSCode            string `json:"sms_code" v:"required#请输入验证码" dc:"短信验证码"`
	Password           string `json:"password" v:"required#请输入登录密码" dc:"登录密码"`
	ConfirmPassword    string `json:"confirm_password" v:"required#请确认登录密码" dc:"确认登录密码"`
	PayPassword        string `json:"pay_password" v:"required#请输入支付密码" dc:"支付密码"`
	ConfirmPayPassword string `json:"confirm_pay_password" v:"required#请确认支付密码" dc:"确认支付密码"`
}

// CustomerRegisterRes 返回注册成功后的客户 token 和基础信息。
type CustomerRegisterRes struct {
	Token    string            `json:"token" dc:"客户登录令牌"`
	Customer CustomerLoginUser `json:"customer" dc:"当前客户"`
}

// CustomerLoginReq 用于客户手机号和登录密码登录。
type CustomerLoginReq struct {
	g.Meta   `path:"/auth/login" method:"post" tags:"客户认证" summary:"客户登录" dc:"客户手机号密码登录"`
	Phone    string `json:"phone" v:"required#请输入手机号" dc:"手机号"`
	Password string `json:"password" v:"required#请输入登录密码" dc:"登录密码"`
}

// CustomerLoginRes 返回客户登录 token 和基础信息。
type CustomerLoginRes struct {
	Token    string            `json:"token" dc:"客户登录令牌"`
	Customer CustomerLoginUser `json:"customer" dc:"当前客户"`
}

// CustomerForgotPasswordReq 用于客户通过短信验证码重置登录密码。
type CustomerForgotPasswordReq struct {
	g.Meta          `path:"/auth/forgot-password" method:"post" tags:"客户认证" summary:"忘记密码" dc:"客户通过短信验证码重置登录密码"`
	Phone           string `json:"phone" v:"required#请输入手机号" dc:"手机号"`
	SMSCode         string `json:"sms_code" v:"required#请输入验证码" dc:"短信验证码"`
	Password        string `json:"password" v:"required#请输入新密码" dc:"新登录密码"`
	ConfirmPassword string `json:"confirm_password" v:"required#请确认新密码" dc:"确认新登录密码"`
}

// CustomerForgotPasswordRes 表示客户登录密码重置成功（返回体为空）。
type CustomerForgotPasswordRes struct{}
