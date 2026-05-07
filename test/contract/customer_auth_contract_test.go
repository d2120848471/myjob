package contract_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"myjob/internal/app"
	modelruntime "myjob/internal/model/runtime"

	"github.com/stretchr/testify/require"
)

func TestCustomerSMSKeysAreSceneScoped(t *testing.T) {
	require.Equal(t, "customer:sms:register:13800005555", app.CustomerSMSCodeKey("register", "13800005555"))
	require.Equal(t, "customer:sms:forgot_password:13800005555", app.CustomerSMSCodeKey("forgot_password", "13800005555"))
	require.Equal(t, "customer:sms:send_lock:register:13800005555", app.CustomerSMSSendLockKey("register", "13800005555"))
	require.Equal(t, "customer:sms:attempts:register:13800005555", app.CustomerSMSAttemptsKey("register", "13800005555"))
}

func TestCustomerSessionKeysAreSeparateFromAdminSessions(t *testing.T) {
	require.Equal(t, "customer:session:jti-001", app.CustomerSessionKey("jti-001"))
	require.Equal(t, "customer:user:sessions:1001", app.CustomerSessionsKey(1001))
}

func TestCustomerMockSMSCodeCanBeRead(t *testing.T) {
	h := newTestHarness(t)
	cfg, err := h.app.Core().LoadSMSConfig(context.Background())
	require.NoError(t, err)
	err = h.app.Core().Sender().SendCode(context.Background(), "13800006666", "123456", cfg)
	require.NoError(t, err)
	code := h.lastSMSCode(t, "13800006666")
	require.Equal(t, "123456", code)
}

func TestCustomerAuth_RegisterLoginForgotPasswordFlow(t *testing.T) {
	h := newTestHarness(t)

	sendRegister := h.postJSON("/api/customer/auth/sms/send", map[string]any{
		"phone": "13800005555",
		"scene": "register",
	}, "")
	require.Equal(t, 0, sendRegister.Code)
	code := h.lastSMSCode(t, "13800005555")

	register := h.postJSON("/api/customer/auth/register", map[string]any{
		"company_name":         "测试公司",
		"phone":                "13800005555",
		"sms_code":             code,
		"password":             "Abc_123",
		"confirm_password":     "Abc_123",
		"pay_password":         "123456",
		"confirm_pay_password": "123456",
	}, "")
	require.Equal(t, 0, register.Code)
	var registerData struct {
		Token    string `json:"token"`
		Customer struct {
			ID          int64  `json:"id"`
			CompanyName string `json:"company_name"`
			Phone       string `json:"phone"`
		} `json:"customer"`
	}
	require.NoError(t, json.Unmarshal(register.Data, &registerData))
	require.NotEmpty(t, registerData.Token)
	require.Equal(t, "测试公司", registerData.Customer.CompanyName)

	login := h.postJSON("/api/customer/auth/login", map[string]any{
		"phone":    "13800005555",
		"password": "Abc_123",
	}, "")
	require.Equal(t, 0, login.Code)

	sendForgot := h.postJSON("/api/customer/auth/sms/send", map[string]any{
		"phone": "13800005555",
		"scene": "forgot_password",
	}, "")
	require.Equal(t, 0, sendForgot.Code)
	forgotCode := h.lastSMSCode(t, "13800005555")

	reset := h.postJSON("/api/customer/auth/forgot-password", map[string]any{
		"phone":            "13800005555",
		"sms_code":         forgotCode,
		"password":         "New_123",
		"confirm_password": "New_123",
	}, "")
	require.Equal(t, 0, reset.Code)

	oldLogin := h.postJSON("/api/customer/auth/login", map[string]any{
		"phone":    "13800005555",
		"password": "Abc_123",
	}, "")
	require.NotEqual(t, 0, oldLogin.Code)

	newLogin := h.postJSON("/api/customer/auth/login", map[string]any{
		"phone":    "13800005555",
		"password": "New_123",
	}, "")
	require.Equal(t, 0, newLogin.Code)
}

func TestCustomerAuth_RejectsSceneReuseAndInvalidPayPassword(t *testing.T) {
	h := newTestHarness(t)

	send := h.postJSON("/api/customer/auth/sms/send", map[string]any{
		"phone": "13800007777",
		"scene": "register",
	}, "")
	require.Equal(t, 0, send.Code)
	code := h.lastSMSCode(t, "13800007777")

	resetWithRegisterCode := h.postJSON("/api/customer/auth/forgot-password", map[string]any{
		"phone":            "13800007777",
		"sms_code":         code,
		"password":         "New_123",
		"confirm_password": "New_123",
	}, "")
	require.NotEqual(t, 0, resetWithRegisterCode.Code)

	register := h.postJSON("/api/customer/auth/register", map[string]any{
		"company_name":         "支付密码错误公司",
		"phone":                "13800007777",
		"sms_code":             code,
		"password":             "Abc_123",
		"confirm_password":     "Abc_123",
		"pay_password":         "abc123",
		"confirm_pay_password": "abc123",
	}, "")
	require.NotEqual(t, 0, register.Code)
}

func TestCustomerAuthSecurityBoundaries(t *testing.T) {
	h := newTestHarness(t)
	adminToken := h.loginAdmin(t)

	t.Run("rejects cross scene code", func(t *testing.T) {
		customerID := createAdminCustomer(t, h, adminToken, "场景隔离客户", "13800007778")
		require.NotZero(t, customerID)

		payload := modelruntime.CustomerSMSCodePayload{
			Scene: app.CustomerSMSSceneRegister,
			Phone: "13800007778",
			Code:  "111111",
		}
		data, err := json.Marshal(payload)
		require.NoError(t, err)
		err = h.app.Core().RedisSetString(
			context.Background(),
			app.CustomerSMSCodeKey(app.CustomerSMSSceneRegister, "13800007778"),
			string(data),
			time.Minute,
		)
		require.NoError(t, err)

		resetWithRegisterCode := h.postJSON("/api/customer/auth/forgot-password", map[string]any{
			"phone":            "13800007778",
			"sms_code":         "111111",
			"password":         "New_123",
			"confirm_password": "New_123",
		}, "")
		require.NotEqual(t, 0, resetWithRegisterCode.Code)
	})

	t.Run("limits wrong sms attempts", func(t *testing.T) {
		customerID := createAdminCustomer(t, h, adminToken, "验证码次数客户", "13800007779")
		require.NotZero(t, customerID)

		sendForgot := h.postJSON("/api/customer/auth/sms/send", map[string]any{
			"phone": "13800007779",
			"scene": "forgot_password",
		}, "")
		require.Equal(t, 0, sendForgot.Code)
		code := h.lastSMSCode(t, "13800007779")
		wrongCode := "000000"
		if code == wrongCode {
			wrongCode = "999999"
		}

		for i := 0; i < 5; i++ {
			wrong := h.postJSON("/api/customer/auth/forgot-password", map[string]any{
				"phone":            "13800007779",
				"sms_code":         wrongCode,
				"password":         "New_123",
				"confirm_password": "New_123",
			}, "")
			require.NotEqual(t, 0, wrong.Code)
		}

		resetWithCorrectCode := h.postJSON("/api/customer/auth/forgot-password", map[string]any{
			"phone":            "13800007779",
			"sms_code":         code,
			"password":         "New_123",
			"confirm_password": "New_123",
		}, "")
		require.NotEqual(t, 0, resetWithCorrectCode.Code)
	})

	t.Run("token invalidation rules", func(t *testing.T) {
		customerID := createAdminCustomer(t, h, adminToken, "令牌失效客户", "13800007781")

		login := h.postJSON("/api/customer/auth/login", map[string]any{
			"phone":    "13800007781",
			"password": "Abc_123",
		}, "")
		require.Equal(t, 0, login.Code)
		var loginData struct {
			Token string `json:"token"`
		}
		require.NoError(t, json.Unmarshal(login.Data, &loginData))
		require.NotEmpty(t, loginData.Token)

		_, _, err := h.app.Core().AuthenticateCustomerRequest(context.Background(), "Bearer "+loginData.Token)
		require.NoError(t, err)

		resetPayPassword := h.patchJSON(fmt.Sprintf("/api/admin/customers/%d/pay-password", customerID), map[string]any{
			"pay_password":         "654321",
			"confirm_pay_password": "654321",
		}, adminToken)
		require.Equal(t, 0, resetPayPassword.Code)
		_, _, err = h.app.Core().AuthenticateCustomerRequest(context.Background(), "Bearer "+loginData.Token)
		require.NoError(t, err)

		resetPassword := h.patchJSON(fmt.Sprintf("/api/admin/customers/%d/password", customerID), map[string]any{
			"password":         "New_123",
			"confirm_password": "New_123",
		}, adminToken)
		require.Equal(t, 0, resetPassword.Code)
		_, _, err = h.app.Core().AuthenticateCustomerRequest(context.Background(), "Bearer "+loginData.Token)
		require.Error(t, err)

		relogin := h.postJSON("/api/customer/auth/login", map[string]any{
			"phone":    "13800007781",
			"password": "New_123",
		}, "")
		require.Equal(t, 0, relogin.Code)
		var reloginData struct {
			Token string `json:"token"`
		}
		require.NoError(t, json.Unmarshal(relogin.Data, &reloginData))

		disable := h.patchJSON(fmt.Sprintf("/api/admin/customers/%d/status", customerID), map[string]any{"status": 0}, adminToken)
		require.Equal(t, 0, disable.Code)
		_, _, err = h.app.Core().AuthenticateCustomerRequest(context.Background(), "Bearer "+reloginData.Token)
		require.Error(t, err)
	})

	t.Run("cleans sms cache when sender fails", func(t *testing.T) {
		h.app.SetSMSSender(failingSMSSender{})

		send := h.postJSON("/api/customer/auth/sms/send", map[string]any{
			"phone": "13800007780",
			"scene": "register",
		}, "")
		require.NotEqual(t, 0, send.Code)

		exists, err := h.app.Redis().GroupGeneric().Exists(
			context.Background(),
			app.CustomerSMSCodeKey(app.CustomerSMSSceneRegister, "13800007780"),
			app.CustomerSMSSendLockKey(app.CustomerSMSSceneRegister, "13800007780"),
		)
		require.NoError(t, err)
		require.EqualValues(t, 0, exists)
	})
}

func TestCustomerLoginRejectsDisabledAndDeletedAccounts(t *testing.T) {
	h := newTestHarness(t)
	adminToken := h.loginAdmin(t)

	create := h.postJSON("/api/admin/customers", map[string]any{
		"company_name":         "登录禁用客户",
		"phone":                "13800010002",
		"password":             "Abc_123",
		"confirm_password":     "Abc_123",
		"pay_password":         "123456",
		"confirm_pay_password": "123456",
		"status":               1,
	}, adminToken)
	require.Equal(t, 0, create.Code)
	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(create.Data, &data))

	require.Equal(t, 0, h.patchJSON(fmt.Sprintf("/api/admin/customers/%d/status", data.ID), map[string]any{"status": 0}, adminToken).Code)
	disabledLogin := h.postJSON("/api/customer/auth/login", map[string]any{"phone": "13800010002", "password": "Abc_123"}, "")
	require.NotEqual(t, 0, disabledLogin.Code)

	require.Equal(t, 0, h.deleteJSON(fmt.Sprintf("/api/admin/customers/%d", data.ID), adminToken).Code)
	deletedLogin := h.postJSON("/api/customer/auth/login", map[string]any{"phone": "13800010002", "password": "Abc_123"}, "")
	require.NotEqual(t, 0, deletedLogin.Code)
}

func createAdminCustomer(t *testing.T, h *testHarness, adminToken, companyName, phone string) int64 {
	t.Helper()
	create := h.postJSON("/api/admin/customers", map[string]any{
		"company_name":         companyName,
		"phone":                phone,
		"password":             "Abc_123",
		"confirm_password":     "Abc_123",
		"pay_password":         "123456",
		"confirm_pay_password": "123456",
		"status":               1,
	}, adminToken)
	require.Equal(t, 0, create.Code)
	var data struct {
		ID int64 `json:"id"`
	}
	require.NoError(t, json.Unmarshal(create.Data, &data))
	require.NotZero(t, data.ID)
	return data.ID
}
