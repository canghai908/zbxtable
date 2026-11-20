package handler

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOnCallHandlersValidateBeforeDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/schedules", CreateOnCallSchedule)
	r.PUT("/schedules/:id", UpdateOnCallSchedule)
	r.GET("/schedules/:id", GetOnCallSchedule)
	r.DELETE("/schedules/:id", DeleteOnCallSchedule)
	r.POST("/schedules/:id/preview", PreviewOnCallSchedule)
	r.POST("/schedules/:id/overrides", CreateOnCallOverride)
	r.DELETE("/schedules/:id/overrides/:override_id", DeleteOnCallOverride)
	for _, tt := range []struct {
		method, path, body string
		code               int
	}{
		{"POST", "/schedules", `{`, 400},
		{"POST", "/schedules", `{"name":"test","timezone":"bad/zone"}`, 422},
		{"PUT", "/schedules/1", `{"name":"test","timezone":"Asia/Shanghai","rotation":"hourly"}`, 422},
		{"PUT", "/schedules/0", `{}`, 400},
		{"GET", "/schedules/not-id", "", 400},
		{"DELETE", "/schedules/-1", "", 400},
		{"POST", "/schedules/1/preview", `{"at":"bad"}`, 422},
		{"POST", "/schedules/1/preview", `{"at":"2026-11-20T10:00:00+08:00","days":32}`, 422},
		{"POST", "/schedules/1/preview", `{"at":"2026-11-20T10:00:00+08:00","days":7,"channel":"sms"}`, 422},
		{"POST", "/schedules/1/overrides", `{"start":"bad"}`, 422},
		{"POST", "/schedules/1/overrides", `{"start":"2026-11-20T10:00:00+08:00","end":"2026-11-20T09:00:00+08:00","user_ids":[1],"reason":"replacement"}`, 422},
		{"DELETE", "/schedules/1/overrides/0", "", 400},
	} {
		t.Run(tt.method+tt.path+tt.body, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			var got struct {
				Code int `json:"code"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.Code != tt.code {
				t.Fatalf("got %s want code %d", w.Body.String(), tt.code)
			}
		})
	}
}
