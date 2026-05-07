package sms

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	modelruntime "myjob/internal/model/runtime"

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

var smsCodeRegexp = regexp.MustCompile(`^\d{6}$`)

// Sender 抽象短信发送能力，业务层只依赖该接口，不感知具体供应商实现。
type Sender interface {
	SendLoginCode(ctx context.Context, phone, code string, cfg modelruntime.SMSConfig) error
	SendCode(ctx context.Context, phone, code string, cfg modelruntime.SMSConfig) error
}

// MockSender 是用于本地/测试的短信发送器实现：不发送真实短信，仅记录验证码。
type MockSender struct {
	mu    sync.RWMutex
	codes map[string]string
}

// NewSender 根据 provider 创建短信发送器实现。
//
// 不支持的 provider 会回退到 mock，避免运行时直接崩溃。
func NewSender(provider string) Sender {
	// Provider 选择统一收口在这里，业务层只依赖 Sender 抽象，不感知具体供应商。
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", smsProviderMock:
		return NewMockSender()
	case smsProviderAliyun:
		return newAliyunSender()
	default:
		return NewMockSender()
	}
}

// NewMockSender 创建一个 mock 短信发送器。
func NewMockSender() *MockSender {
	return &MockSender{codes: make(map[string]string)}
}

// SendLoginCode 记录登录验证码（mock 实现，不产生外部请求）。
func (m *MockSender) SendLoginCode(_ context.Context, phone, code string, _ modelruntime.SMSConfig) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[phone] = code
	return nil
}

// SendCode 记录通用验证码（mock 实现，不产生外部请求）。
func (m *MockSender) SendCode(ctx context.Context, phone, code string, cfg modelruntime.SMSConfig) error {
	return m.SendLoginCode(ctx, phone, code, cfg)
}

// LastCode 返回 mock 发送器给指定手机号记录的最近一次验证码。
func (m *MockSender) LastCode(phone string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	code, ok := m.codes[phone]
	if !ok {
		return "", errors.New("sms code not found")
	}
	return code, nil
}

type aliyunSMSAPI interface {
	SendSms(request *dysms.SendSmsRequest) (*dysms.SendSmsResponse, error)
}

type aliyunSMSClientFactory func(cfg modelruntime.SMSConfig) (aliyunSMSAPI, error)

// AliyunSender 是基于阿里云短信服务的 Sender 实现。
type AliyunSender struct {
	newClient aliyunSMSClientFactory
}

func newAliyunSender() *AliyunSender {
	return &AliyunSender{newClient: newAliyunSMSClient}
}

// SendLoginCode 使用已保存的阿里云配置发送登录验证码。
func (s *AliyunSender) SendLoginCode(_ context.Context, phone, code string, cfg modelruntime.SMSConfig) error {
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

// SendCode 使用已保存的阿里云配置发送通用验证码。
func (s *AliyunSender) SendCode(ctx context.Context, phone, code string, cfg modelruntime.SMSConfig) error {
	return s.SendLoginCode(ctx, phone, code, cfg)
}

func newAliyunSMSClient(cfg modelruntime.SMSConfig) (aliyunSMSAPI, error) {
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
