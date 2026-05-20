package handler

import (
	stdjson "encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestAIChatCustomMissingConfig(t *testing.T) {
	originalConfigGetter := aiConfigValueByKey
	originalClient := aiHTTPClient
	defer func() {
		aiConfigValueByKey = originalConfigGetter
		aiHTTPClient = originalClient
	}()

	aiConfigValueByKey = func(key string, defaultVal string) string {
		values := map[string]string{
			"ai_type": "custom",
		}
		if v, ok := values[key]; ok {
			return v
		}
		return defaultVal
	}
	aiHTTPClient = &http.Client{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/ai/chat", strings.NewReader(`{"message":"analyze this alarm"}`))

	AIChat(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := stdjson.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != 500 {
		t.Fatalf("response code = %d, want 500", resp.Code)
	}
	if !strings.Contains(resp.Message, "Custom API Key 未配置") {
		t.Fatalf("response message = %q, want missing Custom API Key error", resp.Message)
	}
}

func TestAIChatCustomStreamsOpenAICompatibleResponse(t *testing.T) {
	originalConfigGetter := aiConfigValueByKey
	originalClient := aiHTTPClient
	defer func() {
		aiConfigValueByKey = originalConfigGetter
		aiHTTPClient = originalClient
	}()

	aiConfigValueByKey = func(key string, defaultVal string) string {
		values := map[string]string{
			"ai_type":               "custom",
			"custom_api_key":        "test-api-key",
			"custom_model":          "openai-compatible-model",
			"custom_base_url":       "https://example.test/v1",
			"alarm_analysis_prompt": "",
		}
		if v, ok := values[key]; ok {
			return v
		}
		return defaultVal
	}

	var capturedBody string
	aiHTTPClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			if req.URL.String() != "https://example.test/v1/chat/completions" {
				return nil, fmt.Errorf("unexpected request url %q", req.URL.String())
			}
			if got := req.Header.Get("Authorization"); got != "Bearer test-api-key" {
				return nil, fmt.Errorf("unexpected Authorization header %q", got)
			}
			if got := req.Header.Get("Content-Type"); got != "application/json" {
				return nil, fmt.Errorf("unexpected Content-Type header %q", got)
			}

			bodyBytes, err := io.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}
			capturedBody = string(bodyBytes)

			respBody := strings.Join([]string{
				"data: {\"choices\":[{\"delta\":{\"content\":\"hello\"},\"finish_reason\":\"\"}]}",
				"data: {\"choices\":[{\"delta\":{\"content\":\" world\"},\"finish_reason\":\"stop\"}]}",
				"data: [DONE]",
				"",
			}, "\n")

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(respBody)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/ai/chat", strings.NewReader(`{"message":"analyze this alarm"}`))

	AIChat(c)

	if body := w.Body.String(); body != "hello world" {
		t.Fatalf("streamed body = %q, want %q", body, "hello world")
	}
	if !strings.Contains(capturedBody, `"model":"openai-compatible-model"`) {
		t.Fatalf("expected request body to contain custom model, got %s", capturedBody)
	}
	if !strings.Contains(capturedBody, `"stream":true`) {
		t.Fatalf("expected request body to enable stream, got %s", capturedBody)
	}
}

func TestAIChatCustomPropagatesUpstreamError(t *testing.T) {
	originalConfigGetter := aiConfigValueByKey
	originalClient := aiHTTPClient
	defer func() {
		aiConfigValueByKey = originalConfigGetter
		aiHTTPClient = originalClient
	}()

	aiConfigValueByKey = func(key string, defaultVal string) string {
		values := map[string]string{
			"ai_type":         "custom",
			"custom_api_key":  "test-api-key",
			"custom_model":    "openai-compatible-model",
			"custom_base_url": "https://example.test/v1",
		}
		if v, ok := values[key]; ok {
			return v
		}
		return defaultVal
	}
	aiHTTPClient = &http.Client{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`{"error":"bad request"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/ai/chat", strings.NewReader(`{"message":"analyze this alarm"}`))

	AIChat(c)

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := stdjson.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code != http.StatusBadRequest {
		t.Fatalf("response code = %d, want %d", resp.Code, http.StatusBadRequest)
	}
	if !strings.Contains(resp.Message, "Custom AI API 错误") {
		t.Fatalf("response message = %q, want upstream error prefix", resp.Message)
	}
}
