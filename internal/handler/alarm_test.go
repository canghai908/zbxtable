package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

func TestAnalysisAlarmSupportsGetQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	original := analysisAlarmFunc
	analysisAlarmFunc = func(start, end time.Time, zid string) ([]string, []model.Pie, []string, []int, error) {
		if zid != "demo-instance" {
			t.Fatalf("expected zid demo-instance, got %q", zid)
		}
		if start.Format("2006-01-02 15:04:05") != "2026-06-01 00:00:00" {
			t.Fatalf("unexpected start time: %s", start.Format("2006-01-02 15:04:05"))
		}
		if end.Format("2006-01-02 15:04:05") != "2026-06-02 00:00:00" {
			t.Fatalf("unexpected end time: %s", end.Format("2006-01-02 15:04:05"))
		}
		return []string{"critical"}, []model.Pie{{Name: "critical", Value: 3}}, []string{"host-a"}, []int{3}, nil
	}
	defer func() {
		analysisAlarmFunc = original
	}()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/alarm/analysis?begin=2026-06-01%2000:00:00&end=2026-06-02%2000:00:00&zid=demo-instance", nil)

	AnalysisAlarm(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Host []string `json:"host"`
		} `json:"data"`
	}
	if err := jsoniter.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != 200 {
		t.Fatalf("expected response code 200, got %d", resp.Code)
	}
	if len(resp.Data.Host) != 1 || resp.Data.Host[0] != "host-a" {
		t.Fatalf("unexpected host list: %#v", resp.Data.Host)
	}
}

func TestAnalysisAlarmStillSupportsPostBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	original := analysisAlarmFunc
	analysisAlarmFunc = func(start, end time.Time, zid string) ([]string, []model.Pie, []string, []int, error) {
		if zid != "legacy-client" {
			t.Fatalf("expected zid legacy-client, got %q", zid)
		}
		return []string{"warning"}, []model.Pie{{Name: "warning", Value: 1}}, []string{"host-b"}, []int{1}, nil
	}
	defer func() {
		analysisAlarmFunc = original
	}()

	body := `{"begin":"2026-06-01 00:00:00","end":"2026-06-02 00:00:00","zid":"legacy-client"}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/api/v1/alarm/analysis", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	AnalysisAlarm(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Level []string `json:"level"`
		} `json:"data"`
	}
	if err := jsoniter.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response failed: %v", err)
	}
	if resp.Code != 200 {
		t.Fatalf("expected response code 200, got %d", resp.Code)
	}
	if len(resp.Data.Level) != 1 || resp.Data.Level[0] != "warning" {
		t.Fatalf("unexpected level list: %#v", resp.Data.Level)
	}
}
