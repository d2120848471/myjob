package sms

import (
	"context"
	"errors"
	"testing"

	modelruntime "myjob/internal/model/runtime"

	dysms "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/stretchr/testify/require"
)

type fakeAliyunSMSClient struct {
	response *dysms.SendSmsResponse
	err      error
}

func (f fakeAliyunSMSClient) SendSms(_ *dysms.SendSmsRequest) (*dysms.SendSmsResponse, error) {
	return f.response, f.err
}

func TestAliyunSMSSender_SendLoginCode(t *testing.T) {
	t.Parallel()

	sender := &AliyunSender{
		newClient: func(cfg modelruntime.SMSConfig) (aliyunSMSAPI, error) {
			require.Equal(t, "ak", cfg.AccessKey)
			require.Equal(t, "sk", cfg.AccessKeySecret)
			return fakeAliyunSMSClient{response: &dysms.SendSmsResponse{Body: &dysms.SendSmsResponseBody{Code: tea.String("OK")}}}, nil
		},
	}

	err := sender.SendLoginCode(context.Background(), "13800001111", "123456", modelruntime.SMSConfig{
		AccessKey:       "ak",
		AccessKeySecret: "sk",
		SignName:        "签名",
		TemplateCode:    "SMS_000001",
	})
	require.NoError(t, err)
}

func TestAliyunSMSSender_SendLoginCode_ErrorCases(t *testing.T) {
	t.Parallel()

	sender := &AliyunSender{newClient: func(modelruntime.SMSConfig) (aliyunSMSAPI, error) {
		return fakeAliyunSMSClient{err: errors.New("network")}, nil
	}}
	err := sender.SendLoginCode(context.Background(), "13800001111", "123456", modelruntime.SMSConfig{SignName: "签名", TemplateCode: "SMS_1"})
	require.ErrorContains(t, err, "network")

	sender = &AliyunSender{newClient: func(modelruntime.SMSConfig) (aliyunSMSAPI, error) {
		return fakeAliyunSMSClient{response: &dysms.SendSmsResponse{Body: &dysms.SendSmsResponseBody{Code: tea.String("Invalid"), Message: tea.String("bad")}}}, nil
	}}
	err = sender.SendLoginCode(context.Background(), "13800001111", "123456", modelruntime.SMSConfig{SignName: "签名", TemplateCode: "SMS_1"})
	require.ErrorContains(t, err, "Invalid")
}
