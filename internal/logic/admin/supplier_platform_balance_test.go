package adminlogic

import (
	"errors"
	"net/http"
	"strings"
	"testing"
)

func TestShouldRetrySupplierCandidate(t *testing.T) {
	t.Run("https html fallback to http", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "https://api.julangvip.com/api/recharge/user/amount/detail", nil)
		response := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		}
		if !shouldRetrySupplierCandidate(request, response, []byte("<!DOCTYPE html><html><body>spa</body></html>"), errors.New("响应解析失败")) {
			t.Fatalf("expected html response on https to retry with next candidate")
		}
	})

	t.Run("https json business failure should stop", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "https://api.julangvip.com/api/recharge/user/amount/detail", nil)
		response := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"application/json;charset=utf-8"}},
		}
		if shouldRetrySupplierCandidate(request, response, []byte(`{"code":1002,"message":"IP不在IP白名单内"}`), errors.New("IP不在IP白名单内")) {
			t.Fatalf("expected json business failure to stop instead of retry")
		}
	})

	t.Run("http html should retry", func(t *testing.T) {
		request, _ := http.NewRequest(http.MethodGet, "http://api.julangvip.com/api/recharge/user/amount/detail", nil)
		response := &http.Response{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/html"}},
		}
		body := []byte(strings.Repeat("x", 32))
		// HTTP-first 顺序下，HTTP 收到 HTML（CDN 默认页）也应继续尝试下一候选。
		if !shouldRetrySupplierCandidate(request, response, body, errors.New("响应解析失败")) {
			t.Fatalf("expected html response on http to retry with next candidate")
		}
	})
}

func TestHTTPClientForProvider(t *testing.T) {
	logic := &SupplierPlatformLogic{
		httpClient: &http.Client{
			Timeout:   10,
			Transport: http.DefaultTransport.(*http.Transport).Clone(),
		},
	}

	kakaClient := logic.httpClientForProvider("kakayun")
	if kakaClient == logic.httpClient {
		t.Fatalf("expected kakayun to use a derived client")
	}
	kakaTransport, ok := kakaClient.Transport.(*http.Transport)
	if !ok {
		t.Fatalf("expected kakayun transport clone")
	}
	if !kakaTransport.DisableCompression {
		t.Fatalf("expected kakayun transport to disable compression")
	}

	otherClient := logic.httpClientForProvider("xingquanyi")
	if otherClient != logic.httpClient {
		t.Fatalf("expected non-kakayun providers to reuse default client")
	}
}
