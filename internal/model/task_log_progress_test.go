package model

import "testing"

func TestTaskLogProgressEncodingAndParsing(t *testing.T) {
	result := buildTaskLogResult(Running, 42, "正在生成 HTML 报表")
	progress, detail := parseTaskLogProgressResult(result)

	if progress != 42 {
		t.Fatalf("expected progress 42, got %d", progress)
	}
	if detail != "正在生成 HTML 报表" {
		t.Fatalf("expected detail to round-trip, got %q", detail)
	}
}

func TestTaskLogProgressParsingFallback(t *testing.T) {
	progress, detail := parseTaskLogProgressResult("执行成功")

	if progress != 0 {
		t.Fatalf("expected progress 0 for plain text result, got %d", progress)
	}
	if detail != "执行成功" {
		t.Fatalf("expected plain result to remain unchanged, got %q", detail)
	}
}
