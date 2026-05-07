package customerlogic

import "myjob/internal/app"

// AuthLogic 提供客户注册、登录、短信验证码和忘记密码能力。
type AuthLogic struct{ core *app.Core }

// NewAuthLogic 创建客户认证逻辑。
func NewAuthLogic(core *app.Core) *AuthLogic { return &AuthLogic{core: core} }
