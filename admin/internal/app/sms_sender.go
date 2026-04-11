package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysms "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	"github.com/alibabacloud-go/tea/tea"
)

const (
	smsProviderMock    = "mock"
	smsProviderAliyun  = "aliyun"
	aliyunSMSRegionID  = "cn-hangzhou"
	aliyunSMSEndpoint  = "dysmsapi.aliyuncs.com"
	aliyunSMSOKCode    = "OK"
	aliyunSMSTplVarKey = "code"
)

func newSMSSender(cfg RuntimeSMSConfig) SMSSender {
	switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
	case "", smsProviderMock:
		return newMockSMSSender()
	case smsProviderAliyun:
		return newAliyunSMSSender()
	default:
		return newMockSMSSender()
	}
}

type aliyunSMSAPI interface {
	SendSms(request *dysms.SendSmsRequest) (*dysms.SendSmsResponse, error)
}

type aliyunSMSClientFactory func(cfg SMSConfig) (aliyunSMSAPI, error)

type AliyunSMSSender struct {
	newClient aliyunSMSClientFactory
}

func newAliyunSMSSender() *AliyunSMSSender {
	return &AliyunSMSSender{newClient: newAliyunSMSClient}
}

// SendLoginCode 使用已保存的阿里云配置发送登录验证码。
func (s *AliyunSMSSender) SendLoginCode(_ context.Context, phone, code string, cfg SMSConfig) error {
	phone = strings.TrimSpace(phone)
	code = strings.TrimSpace(code)
	signName := strings.TrimSpace(cfg.SignName)
	templateCode := strings.TrimSpace(cfg.TemplateCode)
	if phone == "" {
		return errors.New("短信手机号不能为空")
	}
	if !smsCodeRegexp.MatchString(code) {
		return errors.New("短信验证码格式错误")
	}
	if signName == "" || templateCode == "" {
		return errors.New("阿里云短信签名和模板编号不能为空")
	}

	clientFactory := s.newClient
	if clientFactory == nil {
		clientFactory = newAliyunSMSClient
	}
	client, err := clientFactory(cfg)
	if err != nil {
		return err
	}

	templateParam, err := buildAliyunLoginTemplateParam(code)
	if err != nil {
		return err
	}
	resp, err := client.SendSms(&dysms.SendSmsRequest{
		PhoneNumbers:  tea.String(phone),
		SignName:      tea.String(signName),
		TemplateCode:  tea.String(templateCode),
		TemplateParam: tea.String(templateParam),
	})
	if err != nil {
		return fmt.Errorf("阿里云短信发送失败: %w", err)
	}
	if resp == nil || resp.Body == nil {
		return errors.New("阿里云短信响应为空")
	}
	if tea.StringValue(resp.Body.Code) != aliyunSMSOKCode {
		return fmt.Errorf(
			"阿里云短信发送失败: code=%s message=%s",
			tea.StringValue(resp.Body.Code),
			tea.StringValue(resp.Body.Message),
		)
	}
	return nil
}

func newAliyunSMSClient(cfg SMSConfig) (aliyunSMSAPI, error) {
	accessKey := strings.TrimSpace(cfg.AccessKey)
	accessKeySecret := strings.TrimSpace(cfg.AccessKeySecret)
	if accessKey == "" {
		return nil, errors.New("阿里云 AccessKey 不能为空")
	}
	if accessKeySecret == "" {
		return nil, errors.New("阿里云 AccessKey Secret 不能为空")
	}

	client, err := dysms.NewClient(&openapi.Config{
		AccessKeyId:     tea.String(accessKey),
		AccessKeySecret: tea.String(accessKeySecret),
		RegionId:        tea.String(aliyunSMSRegionID),
		Endpoint:        tea.String(aliyunSMSEndpoint),
	})
	if err != nil {
		return nil, fmt.Errorf("初始化阿里云短信客户端失败: %w", err)
	}
	return client, nil
}

func buildAliyunLoginTemplateParam(code string) (string, error) {
	data, err := json.Marshal(map[string]string{aliyunSMSTplVarKey: code})
	if err != nil {
		return "", fmt.Errorf("构造阿里云短信模板参数失败: %w", err)
	}
	return string(data), nil
}
