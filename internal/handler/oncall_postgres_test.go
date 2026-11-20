package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
	"zbxtable/internal/middleware"
	"zbxtable/internal/model"
	appLogger "zbxtable/pkg/logger"
)

func TestOnCallAPIWithPostgres(t *testing.T) {
	dsn := os.Getenv("ZBXTABLE_ONCALL_TEST_DSN")
	if dsn == "" {
		t.Skip("set ZBXTABLE_ONCALL_TEST_DSN to a PostgreSQL test database URL")
	}
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		t.Fatal("test DSN must be PostgreSQL URL")
	}
	control, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormlog.Default.LogMode(gormlog.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := control.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	name := fmt.Sprintf("oncall_api_test_%d", time.Now().UnixNano())
	if err := control.Exec("CREATE SCHEMA " + name).Error; err != nil {
		t.Fatal(err)
	}
	defer control.Exec("DROP SCHEMA " + name + " CASCADE")
	q := u.Query()
	q.Set("search_path", name)
	u.RawQuery = q.Encode()
	db, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{Logger: gormlog.Default.LogMode(gormlog.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	conn, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	old := model.DB
	model.DB = db
	defer func() { model.DB = old }()
	if err := db.AutoMigrate(&model.User{}, &model.UserGroup{}, &model.Rule{}, &model.OnCallSchedule{}, &model.OnCallOverride{}, &model.OnCallAudit{}, &model.OnCallDispatchSnapshot{}, &model.ZabbixInstance{}, &model.Alarm{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&[]model.User{{ID: 1, Username: "admin", Role: "admin", Email: "one@example.test"}, {ID: 2, Username: "backup", Role: "user", Email: "backup@example.test"}}).Error; err != nil {
		t.Fatal(err)
	}
	oldLog := appLogger.Log
	appLogger.Log = logrus.New()
	appLogger.Log.SetOutput(io.Discard)
	defer func() { appLogger.Log = oldLog }()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		actor := c.GetHeader("X-Test-Actor")
		if actor == "" {
			actor = "admin"
		}
		c.Set("username", actor)
	})
	group := r.Group("/oncall/schedules", middleware.RequireAdmin())
	group.GET("", ListOnCallSchedules)
	group.POST("", CreateOnCallSchedule)
	group.GET("/:id", GetOnCallSchedule)
	group.PUT("/:id", UpdateOnCallSchedule)
	group.DELETE("/:id", DeleteOnCallSchedule)
	group.POST("/:id/preview", PreviewOnCallSchedule)
	group.POST("/:id/overrides", CreateOnCallOverride)
	group.DELETE("/:id/overrides/:override_id", DeleteOnCallOverride)
	group.GET("/:id/audit", ListOnCallAudit)
	r.POST("/rule", CreateRule)
	r.DELETE("/rule/:id", DeleteRule)
	r.POST("/rule/preview", middleware.RequireAdmin(), PreviewRuleDispatch)
	call := func(method, path, body, actor string, want int) map[string]any {
		t.Helper()
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Test-Actor", actor)
		r.ServeHTTP(w, req)
		var got map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
			t.Fatal(err)
		}
		if got["code"] != float64(want) {
			t.Fatalf("%s %s: %s", method, path, w.Body.String())
		}
		return got
	}
	call("GET", "/oncall/schedules", "", "backup", 403)
	created := call("POST", "/oncall/schedules", `{"name":"API ops","timezone":"Asia/Shanghai","rotation":"daily","start":"2026-11-20T09:00:00+08:00","members":[1,2],"weekdays":[0,1,2,3,4,5,6],"window_start":"00:00","window_end":"00:00","fallback_user_ids":[2],"fallback_group_ids":[],"enabled":true}`, "admin", 200)
	sid := int(created["data"].(map[string]any)["id"].(float64))
	path := fmt.Sprintf("/oncall/schedules/%d", sid)
	call("GET", path, "", "admin", 200)
	preview := call("POST", path+"/preview", `{"at":"2026-11-20T10:00:00+08:00","days":7,"channel":"mail"}`, "admin", 200)
	if len(preview["data"].(map[string]any)["slots"].([]any)) == 0 {
		t.Fatal("preview empty")
	}
	rule := fmt.Sprintf(`{"name":"schedule only","m_type":"1","status":"0","schedule_ids":[%d],"z_ids":"41","channel":"mail","s_week":"0,1,2,3,4,5,6","s_time":"00:00","e_time":"23:59","conditions":"[{\"r_type\":\"host\",\"r_func\":\"==\",\"r_value\":\"prod\"}]"}`, sid)
	call("POST", "/rule", rule, "backup", 403)
	call("POST", "/rule", rule, "admin", 200)
	var saved model.Rule
	if err := db.First(&saved).Error; err != nil {
		t.Fatal(err)
	}
	result := call("POST", "/rule/preview", `{"at":"2026-11-20T10:00:00+08:00","zid":41,"data":{"host":"prod"},"rule":`+rule+`}`, "admin", 200)
	data := result["data"].(map[string]any)
	if data["matched"] != true {
		t.Fatalf("rule preview not matched %+v", data)
	}
	channels := data["channels"].([]any)
	recipients := channels[0].(map[string]any)["user_ids"].([]any)
	if len(recipients) != 1 || recipients[0] != float64(1) {
		t.Fatalf("wrong simulated recipients %+v", data)
	}
	var snapshots int64
	if err := db.Model(&model.OnCallDispatchSnapshot{}).Count(&snapshots).Error; err != nil || snapshots != 0 {
		t.Fatal("preview wrote delivery snapshot")
	}
	overBody := `{"start":"2026-11-20T10:00:00+08:00","end":"2026-11-20T11:00:00+08:00","user_ids":[2],"reason":"cover"}`
	over := call("POST", path+"/overrides", overBody, "admin", 200)
	oid := int(over["data"].(map[string]any)["id"].(float64))
	call("POST", path+"/overrides", overBody, "admin", 422)
	preview = call("POST", path+"/preview", `{"at":"2026-11-20T10:30:00+08:00","days":1,"channel":"mail"}`, "admin", 200)
	if preview["data"].(map[string]any)["current"].(map[string]any)["source"] != "override" {
		t.Fatal("API override not effective")
	}
	call("DELETE", fmt.Sprintf("%s/overrides/%d", path, oid), "", "admin", 200)
	call("DELETE", path, "", "admin", 422)
	call("DELETE", fmt.Sprintf("/rule/%d", saved.ID), "", "admin", 200)
	call("DELETE", path, "", "admin", 200)
	audit := call("GET", path+"/audit", "", "admin", 200)
	if audit["data"].(map[string]any)["total"].(float64) < 4 {
		t.Fatal("audit not preserved after delete")
	}

	t.Run("three_daily_members_roster_and_delivery_are_distinct", func(t *testing.T) {
		members := []model.User{{ID: 11, Username: "daily-one", Email: "one@example.test"}, {ID: 22, Username: "daily-two", Wechat: "two"}, {ID: 33, Username: "daily-three", Wechat: "three"}}
		if err := db.Create(&members).Error; err != nil {
			t.Fatal(err)
		}
		payload := map[string]any{"name": "three daily members", "timezone": "Asia/Shanghai", "rotation": "daily", "start": "2026-11-20T09:00:00+08:00", "members": []int{11, 22}, "weekdays": []int{0, 1, 2, 3, 4, 5, 6}, "window_start": "00:00", "window_end": "00:00", "fallback_user_ids": []int{11}, "fallback_group_ids": []int{}, "enabled": true}
		body, _ := json.Marshal(payload)
		created := call("POST", "/oncall/schedules", string(body), "admin", 200)
		sid := int(created["data"].(map[string]any)["id"].(float64))
		path := fmt.Sprintf("/oncall/schedules/%d", sid)
		// Exercise the actual edit path, not just a newly created three-member row.
		payload["members"] = []int{11, 22, 33}
		body, _ = json.Marshal(payload)
		call("PUT", path, string(body), "admin", 200)
		persisted := call("GET", path, "", "admin", 200)["data"].(map[string]any)["schedule"].(map[string]any)
		if !reflect.DeepEqual(persisted["members"], []any{float64(11), float64(22), float64(33)}) {
			t.Fatalf("members lost on save: %+v", persisted)
		}
		at, _ := time.Parse(time.RFC3339, "2026-11-20T10:00:00+08:00")
		for _, channel := range []string{"", "mail"} {
			request := fmt.Sprintf(`{"at":%q,"days":7,"channel":%q}`, at.Format(time.RFC3339), channel)
			result := call("POST", path+"/preview", request, "admin", 200)["data"].(map[string]any)
			slots := result["slots"].([]any)
			for day := 0; day < 7; day++ {
				instant := at.AddDate(0, 0, day)
				found := false
				for _, entry := range slots {
					slot := entry.(map[string]any)
					start, err := time.Parse(time.RFC3339Nano, slot["start"].(string))
					if err != nil {
						t.Fatal(err)
					}
					end, err := time.Parse(time.RFC3339Nano, slot["end"].(string))
					if err != nil {
						t.Fatal(err)
					}
					if instant.Before(start) || !instant.Before(end) {
						continue
					}
					found = true
					want := []int{11, 22, 33}[day%3]
					source := "rotation"
					if channel == "mail" {
						want = 11
						if day%3 != 0 {
							source = "fallback"
						}
					}
					if !reflect.DeepEqual(slot["user_ids"], []any{float64(want)}) || slot["source"] != source {
						t.Fatalf("channel=%q day=%d slot=%+v", channel, day, slot)
					}
				}
				if !found {
					t.Fatalf("day %d missing from preview", day)
				}
			}
		}
	})

	// Real HTTP requests exercise authentication, rule persistence, webhook input,
	// alarm storage and the same queue used by production consumers.
	t.Run("webhook_to_oncall_dispatch", func(t *testing.T) {
		now := time.Now()
		runtimeBody := fmt.Sprintf(`{"name":"webhook ops","timezone":"Asia/Shanghai","rotation":"fixed","start":%q,"members":[1],"weekdays":[0,1,2,3,4,5,6],"window_start":"00:00","window_end":"00:00","fallback_user_ids":[2],"fallback_group_ids":[],"enabled":true}`, now.Add(-2*time.Hour).Format(time.RFC3339))
		made := call("POST", "/oncall/schedules", runtimeBody, "admin", 200)
		runtimeID := int(made["data"].(map[string]any)["id"].(float64))
		instance := model.ZabbixInstance{ID: 41, Name: "loopback integration", Instance: "test-instance", URL: "http://127.0.0.1/unused", Enabled: true, WebhookToken: "test-webhook-token"}
		if err := db.Create(&instance).Error; err != nil {
			t.Fatal(err)
		}
		draft := model.Rule{Name: "webhook rule", ZIDs: "41", MType: "1", Status: "0", Channel: "mail", ScheduleIDs: []int{runtimeID}, Stime: "00:00", Etime: "23:59", Sweek: "0,1,2,3,4,5,6", Conditions: `{"version":2,"op":"and","children":[{"field":"host","operator":"=","value":"prod-api"},{"op":"or","children":[{"field":"trigger","operator":"like","value":"CPU"},{"field":"trigger","operator":"like","value":"Memory"}]}]}`}
		draftJSON, err := json.Marshal(draft)
		if err != nil {
			t.Fatal(err)
		}
		call("POST", "/rule", string(draftJSON), "admin", 200)
		r.POST("/receive", ReceiveGin)
		r.GET("/alarms", GetAllAlarm)
		server := httptest.NewServer(r)
		defer server.Close()
		body, err := json.Marshal(model.EventTpl{HostHost: "prod-api", Hostname: "api node", HostsIP: "192.0.2.30", EventTime: now.Format("2006-01-02 15:04:05"), Severity: "High", TriggerName: "CPU high", TriggerValue: "1", ItemID: 501, ItemName: "CPU utilization", ItemValue: "95.5", EventID: 41001})
		if err != nil {
			t.Fatal(err)
		}
		encoded := strings.ReplaceAll(string(body), `"`, `¦`)
		send := func(token string) map[string]any {
			t.Helper()
			req, err := http.NewRequest("POST", server.URL+"/receive", strings.NewReader(encoded))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("X-Instance", instance.Instance)
			req.Header.Set("X-Token", token)
			client := server.Client()
			client.Timeout = 5 * time.Second
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			var result map[string]any
			if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
				t.Fatal(err)
			}
			return result
		}
		denied := send("wrong-token")
		if denied["data"].(map[string]any)["id"] != float64(0) {
			t.Fatal("unauthenticated webhook accepted")
		}
		var count int64
		if err := db.Model(&model.Alarm{}).Count(&count).Error; err != nil || count != 0 {
			t.Fatal("invalid token stored an alarm")
		}
		result := send(instance.WebhookToken)
		alarmID := int(result["data"].(map[string]any)["id"].(float64))
		if alarmID <= 0 {
			t.Fatalf("webhook rejected: %+v", result)
		}
		var alarm model.Alarm
		if err := db.First(&alarm, alarmID).Error; err != nil {
			t.Fatal(err)
		}
		if alarm.NotifyStatus != "0" || alarm.ZID != 41 || alarm.ItemValue != "95.5" {
			t.Fatalf("stored alarm: %+v", alarm)
		}
		raw, err := model.CacheRPop("mail")
		if err != nil || raw == "" {
			t.Fatal("webhook did not dispatch")
		}
		var event model.Event
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			t.Fatal(err)
		}
		if event.ID != alarmID || event.ToUsers != "1" || event.InstanceName != instance.Name {
			t.Fatalf("queued event: %+v", event)
		}
		var snap model.OnCallDispatchSnapshot
		if err := db.Where("zid = ? AND event_id = ?", 41, 41001).First(&snap).Error; err != nil || !reflect.DeepEqual(snap.UserIDs, []int{1}) {
			t.Fatalf("webhook snapshot: %+v %v", snap, err)
		}
		// A persisted rule's preview agrees with actual recipients and creates no queue.
		previewBody := fmt.Sprintf(`{"at":%q,"zid":41,"data":{"host":"prod-api","trigger":"CPU high"},"rule":%s}`, now.Format(time.RFC3339), draftJSON)
		simulated := call("POST", "/rule/preview", previewBody, "admin", 200)["data"].(map[string]any)
		actualIDs := simulated["channels"].([]any)[0].(map[string]any)["user_ids"].([]any)
		if len(actualIDs) != 1 || actualIDs[0] != float64(1) {
			t.Fatalf("preview differs from dispatch: %+v", simulated)
		}
		if extra, _ := model.CacheRPop("mail"); extra != "" {
			t.Fatal("preview sent a notification")
		}
		resp, err := server.Client().Get(server.URL + "/alarms?zid=41")
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var listed map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&listed); err != nil {
			t.Fatal(err)
		}
		if listed["data"].(map[string]any)["total"] != float64(1) {
			t.Fatalf("alarm not queryable: %+v", listed)
		}
	})

}
