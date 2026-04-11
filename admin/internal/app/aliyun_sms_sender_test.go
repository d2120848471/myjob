package app

import (
	"context"
	"testing"

	dysms "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/stretchr/testify/require"
)

type fakeAliyunSMSClient struct {
	req  *dysms.SendSmsRequest
	resp *dysms.SendSmsResponse
	err  error
}

func (f *fakeAliyunSMSClient) SendSms(req *dysms.SendSmsRequest) (*dysms.SendSmsResponse, error) {
	f.req = req
	return f.resp, f.err
}

func TestAliyunSMSSender_SendLoginCode_SendsExpectedRequest(t *testing.T) {
	t.Parallel()

	fake := &fakeAliyunSMSClient{
		resp: &dysms.SendSmsResponse{
			Body: &dysms.SendSmsResponseBody{
				Code:    tea.String("OK"),
				Message: tea.String("OK"),
				BizId:   tea.String("biz-1"),
			},
		},
	}
	sender := &AliyunSMSSender{
		newClient: func(cfg SMSConfig) (aliyunSMSAPI, error) {
			require.Equal(t, "test-ak", cfg.AccessKey)
			require.Equal(t, "test-sk", cfg.AccessKeySecret)
			return fake, nil
		},
	}

	err := sender.SendLoginCode(context.Background(), "13800001111", "123456", SMSConfig{
		AccessKey:       "test-ak",
		AccessKeySecret: "test-sk",
		SignName:        "玖权益",
		TemplateCode:    "SMS_308586082",
		ExpireMinutes:   30,
		IntervalMinutes: 1,
	})
	require.NoError(t, err)
	require.NotNil(t, fake.req)
	require.Equal(t, "13800001111", tea.StringValue(fake.req.PhoneNumbers))
	require.Equal(t, "玖权益", tea.StringValue(fake.req.SignName))
	require.Equal(t, "SMS_308586082", tea.StringValue(fake.req.TemplateCode))
	require.JSONEq(t, `{"code":"123456"}`, tea.StringValue(fake.req.TemplateParam))
}

func TestAliyunSMSSender_SendLoginCode_ReturnsErrorWhenAliyunRejects(t *testing.T) {
	t.Parallel()

	sender := &AliyunSMSSender{
		newClient: func(SMSConfig) (aliyunSMSAPI, error) {
			return &fakeAliyunSMSClient{
				resp: &dysms.SendSmsResponse{
					Body: &dysms.SendSmsResponseBody{
						Code:    tea.String("isv.BUSINESS_LIMIT_CONTROL"),
						Message: tea.String("触发业务流控限制"),
					},
				},
			}, nil
		},
	}

	err := sender.SendLoginCode(context.Background(), "13800001111", "123456", SMSConfig{
		AccessKey:       "test-ak",
		AccessKeySecret: "test-sk",
		SignName:        "玖权益",
		TemplateCode:    "SMS_308586082",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "isv.BUSINESS_LIMIT_CONTROL")
}
