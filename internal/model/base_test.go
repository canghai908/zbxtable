package model

import (
	"strings"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

type baseJSONTestStruct struct {
	HTML string `json:"html"`
}

type baseJSONOrderStruct struct {
	B int `json:"b"`
	A int `json:"a"`
}

type baseJSONRawMessageStruct struct {
	Raw jsoniter.RawMessage `json:"raw"`
}

func TestJSONIteratorConfig_EscapeHTMLDisabled(t *testing.T) {
	in := baseJSONTestStruct{HTML: "<b>ok</b>"}

	b, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	out := string(b)
	if strings.Contains(out, "\\u003c") || strings.Contains(out, "\\u003e") {
		t.Fatalf("expected EscapeHTML disabled, got escaped json: %s", out)
	}
	if !strings.Contains(out, "<b>ok</b>") {
		t.Fatalf("expected raw html in json, got: %s", out)
	}
}

func TestJSONIteratorConfig_SortMapKeysEnabled(t *testing.T) {
	m := map[string]int{"b": 1, "a": 2}

	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	out := string(b)
	// SortMapKeys=true 期望输出 a 在 b 前面
	if !(strings.Contains(out, "\"a\"") && strings.Contains(out, "\"b\"")) {
		t.Fatalf("unexpected json output: %s", out)
	}
	if strings.Index(out, "\"a\"") > strings.Index(out, "\"b\"") {
		t.Fatalf("expected sorted keys (a before b), got: %s", out)
	}
}

func TestJSONIteratorConfig_ValidateJsonRawMessage(t *testing.T) {
	// 验证无效 JSON 的解析
	var data baseJSONRawMessageStruct
	invalidJSON := []byte(`{"raw": {invalid}}`)

	err := json.Unmarshal(invalidJSON, &data)
	if err == nil {
		t.Fatalf("expected unmarshal error for invalid JSON when ValidateJsonRawMessage=true")
	}

	// 验证有效 JSON 的解析
	validJSON := []byte(`{"raw": {"k":1}}`)
	err = json.Unmarshal(validJSON, &data)
	if err != nil {
		t.Fatalf("expected unmarshal success for valid JSON, got err: %v", err)
	}
}
