package model

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type onCallRobotTransport func(*http.Request) (*http.Response, error)

func (f onCallRobotTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestOnCallRobotDeliveryPostgres(t *testing.T) {
	db := onCallScenarioDB(t)
	t.Chdir(filepath.Join("..", ".."))
	user := User{ID: 1, Username: "robot-oncall", Role: "admin", WechatRobotKey: "loopback-only-key"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&Config{ConfigKey: "encryption_key", ConfigValue: "test-only-key-01234567890123456789"}).Error; err != nil {
		t.Fatal(err)
	}
	s := OnCallSchedule{Name: "robot oncall", Timezone: "Asia/Shanghai", Rotation: "fixed", Start: time.Now().Add(-time.Hour).Format(time.RFC3339), Members: []int{1}, Weekdays: []int{0, 1, 2, 3, 4, 5, 6}, WindowStart: "00:00", WindowEnd: "00:00", Enabled: true}
	if err := SaveOnCallSchedule(&s, "robot-oncall"); err != nil {
		t.Fatal(err)
	}
	rule := Rule{Name: "robot oncall dispatch", MType: "1", Status: "0", ZIDs: "41", Channel: "wechat_robot", ScheduleIDs: []int{s.ID}, Conditions: workflowConditions, Stime: "00:00", Etime: "23:59", Sweek: "0,1,2,3,4,5,6"}
	if _, err := AddRule(&rule); err != nil {
		t.Fatal(err)
	}
	type captured struct {
		key     string
		message WechatRobotMessage
	}
	received := make(chan captured, 1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var message WechatRobotMessage
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			t.Error(err)
		}
		if err := json.Unmarshal(body, &message); err != nil {
			t.Error(err)
		}
		if r.Method != "POST" || r.Header.Get("Content-Type") != "application/json" {
			t.Error("unexpected robot request")
		}
		received <- captured{r.URL.Query().Get("key"), message}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"errcode":0,"errmsg":"ok"}`)
	}))
	defer server.Close()
	target, _ := url.Parse(server.URL)
	oldClient := http.DefaultClient
	http.DefaultClient = &http.Client{Timeout: 3 * time.Second, Transport: onCallRobotTransport(func(r *http.Request) (*http.Response, error) {
		if r.URL.Scheme != "https" || r.URL.Host != "qyapi.weixin.qq.com" || r.URL.Path != "/cgi-bin/webhook/send" {
			return nil, errors.New("non-test request refused")
		}
		clone := r.Clone(r.Context())
		copyURL := *r.URL
		clone.URL = &copyURL
		clone.URL.Scheme = target.Scheme
		clone.URL.Host = target.Host
		clone.Host = target.Host
		return server.Client().Transport.RoundTrip(clone)
	})}
	defer func() { http.DefaultClient = oldClient }()
	alarm := Alarm{ZID: 41, Host: "prod-db", Hostname: "production-db", HostsIP: "192.0.2.10", ItemName: "CPU", ItemValue: "97", Detail: "CPU:97", ItemID: 501, Message: "CPU high", Level: "High", Status: "1", EventID: 31001, OccurTime: time.Now()}
	if _, err := AddAlarm(&alarm); err != nil {
		t.Fatal(err)
	}
	GenAlert(&alarm)
	events := PopAllWechatRobot()
	if len(events) != 1 || events[0].ToUsers != "1" {
		t.Fatalf("queued: %+v", events)
	}
	oldWorkers := WechatRobotWorkerChan
	WechatRobotWorkerChan = make(chan int, 1)
	defer func() { WechatRobotWorkerChan = oldWorkers }()
	WechatRobotWorkerChan <- 1
	SendWechatRobot(events[0])
	select {
	case got := <-received:
		if got.key != user.WechatRobotKey || got.message.MsgType != "text" || !strings.Contains(got.message.Text.Content, "CPU high") || !strings.Contains(got.message.Text.Content, "97") {
			t.Fatalf("local receiver content mismatch: %+v", got.message)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("local robot receiver got no request")
	}
	logs, err := GetEventLogByAlarmID(alarm.ID)
	if err != nil || len(logs) != 1 || logs[0].Status != "0" || logs[0].EventID != alarm.EventID {
		t.Fatalf("delivery log: %+v %v", logs, err)
	}
}

func TestOnCallRobotFailureReportingPostgres(t *testing.T) {
	db := onCallScenarioDB(t)
	oldClient := http.DefaultClient
	defer func() { http.DefaultClient = oldClient }()
	for index, tt := range []struct {
		name, body string
		status     int
		network    bool
		wantError  bool
	}{
		{name: "accepted", body: `{"errcode":0,"errmsg":"ok"}`, status: 200},
		{name: "business_rejection", body: `{"errcode":93000,"errmsg":"invalid webhook url"}`, status: 200, wantError: true},
		{name: "invalid_json", body: `not-json`, status: 200, wantError: true},
		{name: "oversized_response", body: strings.Repeat(" ", 64*1024) + `{"errcode":0}`, status: 200, wantError: true},
		{name: "missing_errcode", body: `{}`, status: 200, wantError: true},
		{name: "http_rejection", body: `bad gateway`, status: 502, wantError: true},
		{name: "network_error", network: true, wantError: true},
		{name: "response_read_error", status: 200, wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			http.DefaultClient = &http.Client{Transport: onCallRobotTransport(func(*http.Request) (*http.Response, error) {
				if tt.network {
					return nil, errors.New("loopback network failure")
				}
				if tt.name == "response_read_error" {
					reader, writer := io.Pipe()
					_ = writer.CloseWithError(errors.New("test response read failure"))
					return &http.Response{StatusCode: 200, Body: reader, Header: make(http.Header)}, nil
				}
				return &http.Response{StatusCode: tt.status, Body: io.NopCloser(strings.NewReader(tt.body)), Header: make(http.Header)}, nil
			})}
			event := Event{ID: 100, EventID: int64(32000 + index), Rule: "failure-test"}
			err := SendWechatRobotAlert(User{Username: "test", WechatRobotKey: "never-log-this-key"}, &event, "test content", "never-log-this-key")
			if (err != nil) != tt.wantError {
				t.Fatalf("error=%v wantError=%v", err, tt.wantError)
			}
			if err != nil && strings.Contains(err.Error(), "never-log-this-key") {
				t.Fatal("error leaked webhook key")
			}
			var log EventLog
			var count int64
			if err := db.Model(&EventLog{}).Where("event_id = ?", event.EventID).Count(&count).Error; err != nil || count != 1 {
				t.Fatalf("expected one log for event %d, got %d: %v", event.EventID, count, err)
			}
			if err := db.Where("event_id = ?", event.EventID).First(&log).Error; err != nil {
				t.Fatal(err)
			}
			want := "0"
			if tt.wantError {
				want = "1"
			}
			if log.Status != want || (log.NotifyError != "") != tt.wantError {
				t.Fatalf("delivery log %+v", log)
			}
		})
	}
	t.Run("log_storage_failure", func(t *testing.T) {
		if err := db.Migrator().DropTable(&EventLog{}); err != nil {
			t.Fatal(err)
		}
		http.DefaultClient = &http.Client{Transport: onCallRobotTransport(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"errcode":0}`)), Header: make(http.Header)}, nil
		})}
		if err := SendWechatRobotAlert(User{}, &Event{}, "test", "test-only"); err == nil {
			t.Fatal("log storage failure was swallowed")
		}
	})
}
