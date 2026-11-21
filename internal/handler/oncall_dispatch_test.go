package handler

import (
	"github.com/gin-gonic/gin"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRulePreviewRejectsInvalidInputWithoutDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/preview", PreviewRuleDispatch)
	for _, tt := range []struct {
		body   string
		code   int
		reason string
	}{
		{`{`, 400, ""},
		{`{"at":"not-a-time"}`, 422, ""},
		{`{"at":"2026-11-20T10:00:00+08:00","rule":{"m_type":"3"}}`, 422, ""},
		{`{"at":"2026-11-20T10:00:00+08:00","rule":{"m_type":"1","conditions":"[]"}}`, 422, ""},
		{`{"at":"2026-11-20T10:00:00+08:00","zid":1,"rule":{"m_type":"2","status":"1","z_ids":"1"}}`, 200, "规则未启用"},
		{`{"at":"2026-11-20T10:00:00+08:00","zid":1,"rule":{"m_type":"2","status":"0","z_ids":"12"}}`, 200, "测试实例不在规则范围内"},
	} {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/preview", strings.NewReader(tt.body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		var got struct {
			Code int `json:"code"`
			Data struct {
				Matched bool   `json:"matched"`
				Reason  string `json:"reason"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got.Code != tt.code || got.Data.Matched || got.Data.Reason != tt.reason {
			t.Fatalf("%s: %s", tt.body, w.Body.String())
		}
	}
}

func TestScheduleRulesRequireAdminSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/rule", CreateRule)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/rule", strings.NewReader(`{"name":"scheduled","m_type":"2","conditions":"","schedule_ids":[1]}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	var got struct {
		Code int `json:"code"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if got.Code != 401 {
		t.Fatalf("schedule rule bypassed admin: %s", w.Body.String())
	}
	protected := gin.New()
	RegisterOnCallRoutes(protected.Group("/schedules"))
	w = httptest.NewRecorder()
	protected.ServeHTTP(w, httptest.NewRequest("GET", "/schedules", nil))
	_ = json.Unmarshal(w.Body.Bytes(), &got)
	if got.Code != 401 {
		t.Fatalf("schedule list bypassed login: %s", w.Body.String())
	}
}

func TestRuleScheduleIDsPreserveOmissionAndClear(t *testing.T) {
	if ids, err := parseRuleScheduleIDs([]byte(`{}`)); err != nil || ids != nil {
		t.Fatalf("omission not preserved %v %v", ids, err)
	}
	if ids, err := parseRuleScheduleIDs([]byte(`{"schedule_ids":[]}`)); err != nil || ids == nil || len(ids) != 0 {
		t.Fatalf("clear not preserved %v %v", ids, err)
	}
	if _, err := parseRuleScheduleIDs([]byte(`{"schedule_ids":["bad"]}`)); err == nil {
		t.Fatal("invalid schedule IDs accepted")
	}
}
