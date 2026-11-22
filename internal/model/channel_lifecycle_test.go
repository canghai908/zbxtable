package model

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
	"zbxtable/pkg/utils"

	"github.com/sirupsen/logrus"
	appLogger "zbxtable/pkg/logger"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestChannelNormalizationAndDeliverability(t *testing.T) {
	cases := []struct {
		raw       string
		want      string
		supported bool
	}{
		{"mail", "mail", true},
		{"wechat", "wechat", true},
		{"wechatrobot", "wechat_robot", true},
		{"wechat-robot", "wechat_robot", true},
		{"dingtalk", "dingtalk_robot", true},
		{"dingding", "dingtalk_robot", true},
		{"ding_talk", "dingtalk_robot", true},
		{"dingtalk-robot", "dingtalk_robot", true},
		{"feishu", "feishu_robot", true},
		{"lark", "feishu_robot", true},
		{"feishu_robot", "feishu_robot", true},
		{"webhook", "webhook", true},
		{"custom_webhook", "webhook", true},
		{"sms", "sms", false},
	}
	for _, tc := range cases {
		if got := normalizeChannel(tc.raw); got != tc.want {
			t.Fatalf("normalizeChannel(%q) = %q, want %q", tc.raw, got, tc.want)
		}
		if got := IsSupportedChannel(tc.raw); got != tc.supported {
			t.Fatalf("IsSupportedChannel(%q) = %v, want %v", tc.raw, got, tc.supported)
		}
	}

	u := User{
		Status:         0,
		DingTalk:       "dt-token",
		FeishuRobotKey: "fs-key",
		WebhookURL:     "https://hooks.example.test/alert",
	}
	if !onCallUserDeliverable(u, "dingtalk_robot") || !onCallUserDeliverable(u, "feishu_robot") || !onCallUserDeliverable(u, "webhook") {
		t.Fatal("expected enabled user with new channel credentials to be deliverable")
	}
	u.Status = 1
	if onCallUserDeliverable(u, "dingtalk_robot") || onCallUserDeliverable(u, "feishu_robot") || onCallUserDeliverable(u, "webhook") {
		t.Fatal("expected disabled user not to be deliverable")
	}
}

func TestDingTalkSigningAndHTTPDelivery(t *testing.T) {
	now := time.UnixMilli(1700000000123)
	secret := "SECtestsecret123"
	signedURL := BuildDingTalkWebhookURL("token-abc", secret, now)
	parsed, err := url.Parse(signedURL)
	if err != nil {
		t.Fatal(err)
	}
	if parsed.Host != "oapi.dingtalk.com" || parsed.Query().Get("access_token") != "token-abc" {
		t.Fatalf("unexpected DingTalk URL: %s", signedURL)
	}
	if parsed.Query().Get("timestamp") != "1700000000123" {
		t.Fatalf("unexpected timestamp query: %s", parsed.Query().Get("timestamp"))
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte("1700000000123\n" + secret))
	wantSign := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if parsed.Query().Get("sign") != wantSign {
		t.Fatalf("sign mismatch: got %q want %q", parsed.Query().Get("sign"), wantSign)
	}

	oldClient := http.DefaultClient
	defer func() { http.DefaultClient = oldClient }()

	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"msgtype":"text"`) || !strings.Contains(string(body), "CPU high") {
			t.Fatalf("unexpected DingTalk payload: %s", string(body))
		}
		if r.URL.Query().Get("access_token") != "secret-token" || r.URL.Query().Get("sign") == "" {
			t.Fatalf("missing query params in DingTalk request: %s", r.URL.String())
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"errcode":0,"errmsg":"ok"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	if err := sendDingTalkRobotHTTP("CPU high", "secret-token", secret, now); err != nil {
		t.Fatalf("expected DingTalk delivery success, got: %v", err)
	}

	// Ensure transport errors do not leak the token in error message
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("dial tcp failed for https://oapi.dingtalk.com/robot/send?access_token=never-leak-token")
	})}
	if err := sendDingTalkRobotHTTP("CPU high", "never-leak-token", "", now); err == nil || strings.Contains(err.Error(), "never-leak-token") {
		t.Fatalf("expected sanitized error without token, got: %v", err)
	}
}

func TestFeishuSigningAndHTTPDelivery(t *testing.T) {
	now := time.Unix(1700000000, 0)
	secret := "feishu-sec-xyz"
	wantSign := BuildFeishuSign(secret, now.Unix())

	mac := hmac.New(sha256.New, []byte(fmt.Sprintf("%d\n%s", now.Unix(), secret)))
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	if wantSign != expected {
		t.Fatalf("feishu sign mismatch: got %q want %q", wantSign, expected)
	}

	oldClient := http.DefaultClient
	defer func() { http.DefaultClient = oldClient }()

	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://open.feishu.cn/open-apis/bot/v2/hook/fs-token-1" {
			t.Fatalf("unexpected Feishu URL: %s", r.URL.String())
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatal(err)
		}
		if payload["msg_type"] != "text" || payload["timestamp"] != strconv.FormatInt(now.Unix(), 10) || payload["sign"] != wantSign {
			t.Fatalf("unexpected Feishu payload: %s", string(body))
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"code":0,"msg":"success"}`)),
			Header:     make(http.Header),
		}, nil
	})}

	if err := sendFeishuRobotHTTP("Disk full", "fs-token-1", secret, now); err != nil {
		t.Fatalf("expected Feishu delivery success, got: %v", err)
	}

	// Business error rejection
	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"code":19021,"msg":"sign match fail"}`)), Header: make(http.Header)}, nil
	})}
	if err := sendFeishuRobotHTTP("Disk full", "fs-token-1", secret, now); err == nil || !strings.Contains(err.Error(), "19021") {
		t.Fatalf("expected Feishu code error, got: %v", err)
	}
}

func TestWebhookChannelHTTPDelivery(t *testing.T) {
	var gotAuth, gotToken string
	var gotPayload map[string]any

	oldClient := http.DefaultClient
	defer func() { http.DefaultClient = oldClient }()

	http.DefaultClient = &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		gotAuth = r.Header.Get("Authorization")
		gotToken = r.Header.Get("X-Webhook-Token")
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &gotPayload); err != nil {
			t.Fatal(err)
		}
		return &http.Response{
			StatusCode: http.StatusAccepted,
			Body:       io.NopCloser(strings.NewReader(`{"ok":true}`)),
			Header:     make(http.Header),
		}, nil
	})}

	ev := &Event{
		ID:           12,
		EventID:      9001,
		ZID:          1,
		Hostname:     "web-01",
		HostsIP:      "10.0.0.5",
		Level:        "4",
		Status:       "1",
		Message:      "HTTP 500 spike",
		Detail:       "rate=15%",
		OccurTime:    time.Now(),
		Rule:         "prod-webhook",
		InstanceName: "prod-zbx",
	}
	if err := sendWebhookChannelHTTP(ev, "alert text", "https://hooks.example.test/zbx", "wh-secret"); err != nil {
		t.Fatalf("expected webhook 202 success, got: %v", err)
	}
	if gotAuth != "Bearer wh-secret" || gotToken != "wh-secret" {
		t.Fatalf("unexpected headers: auth=%q token=%q", gotAuth, gotToken)
	}
	if gotPayload["hostname"] != "web-01" || gotPayload["level"] != "High" || gotPayload["status"] != "故障" || gotPayload["content"] != "alert text" {
		t.Fatalf("unexpected webhook payload: %+v", gotPayload)
	}
}

func TestStormSuppressionBufferingAndFlush(t *testing.T) {
	oldLog := appLogger.Log
	appLogger.Log = logrus.New()
	appLogger.Log.SetOutput(io.Discard)
	defer func() { appLogger.Log = oldLog }()

	queueMu.Lock()
	oldQueues := queueStore
	queueStore = make(map[string]*Queue)
	queueMu.Unlock()
	defer func() {
		queueMu.Lock()
		queueStore = oldQueues
		queueMu.Unlock()
	}()
	ResetAggregationBuckets()
	defer ResetAggregationBuckets()

	rule := &Rule{
		ID:                7,
		Name:              "storm-rule",
		SuppressWindow:    60,
		SuppressThreshold: 1,
	}
	alarm := &Alarm{ID: 100, ZID: 1, Host: "db-01", Status: "1"}
	t0 := time.Date(2026, 11, 21, 10, 0, 0, 0, time.Local)

	ev1 := &Event{ID: 101, ZID: 1, Host: "db-01", Hostname: "db-01", Channel: "dingtalk_robot", Message: "Disk IO high", Detail: "95%", Status: "1", ToUsers: "1"}
	ev2 := &Event{ID: 102, ZID: 1, Host: "db-01", Hostname: "db-01", Channel: "dingtalk_robot", Message: "CPU load high", Detail: "12.5", Status: "1", ToUsers: "1"}
	ev3 := &Event{ID: 103, ZID: 1, Host: "db-01", Hostname: "db-01", Channel: "dingtalk_robot", Message: "Mem usage high", Detail: "92%", Status: "1", ToUsers: "1"}

	if shouldSuppressAndBuffer(rule, alarm, ev1, t0) {
		t.Fatal("first event within threshold=1 should not be suppressed")
	}
	if !shouldSuppressAndBuffer(rule, alarm, ev2, t0.Add(10*time.Second)) {
		t.Fatal("second event within window should be suppressed")
	}
	if !shouldSuppressAndBuffer(rule, alarm, ev3, t0.Add(20*time.Second)) {
		t.Fatal("third event within window should be suppressed")
	}

	if flushed := FlushExpiredAggregations(t0.Add(30 * time.Second)); flushed != 0 {
		t.Fatalf("expected 0 flushed before window end, got %d", flushed)
	}
	if flushed := FlushExpiredAggregations(t0.Add(61 * time.Second)); flushed != 1 {
		t.Fatalf("expected 1 flushed after window end, got %d", flushed)
	}

	raw, err := CacheRPop("dingtalk_robot")
	if err != nil || raw == "" {
		t.Fatalf("expected aggregated event in dingtalk_robot queue, err=%v", err)
	}
	var summary Event
	if err := json.Unmarshal([]byte(raw), &summary); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(summary.Message, "[风暴抑制汇总]") || !strings.Contains(summary.Message, "共触发 3 条告警（已合并 2 条）") {
		t.Fatalf("unexpected summary message: %s", summary.Message)
	}
	if !strings.Contains(summary.Detail, "CPU load high") || !strings.Contains(summary.Detail, "Mem usage high") {
		t.Fatalf("unexpected summary detail: %s", summary.Detail)
	}
}

func TestUserCredentialRoundTripEncryption(t *testing.T) {
	key := "0123456789abcdef0123456789abcdef"
	enc, err := utils.EncryptString("SEC-secret", key)
	if err != nil {
		t.Fatal(err)
	}
	if got := decryptUserCredential(enc, key); got != "SEC-secret" {
		t.Fatalf("decryptUserCredential = %q, want %q", got, "SEC-secret")
	}
	if got := decryptUserCredential("legacy-plain-value", key); got != "legacy-plain-value" {
		t.Fatalf("fallback decryptUserCredential = %q, want legacy-plain-value", got)
	}
}

func TestAckAndEscalationPostgres(t *testing.T) {
	db := onCallScenarioDB(t)
	users := []User{
		{ID: 1, Username: "oncall-primary", Role: "admin", Email: "primary@example.test", Status: 0},
		{ID: 2, Username: "escalation-lead", Role: "user", DingTalk: "dt-lead-key", FeishuRobotKey: "fs-lead-key", WebhookURL: "https://hook.example.test/lead", Status: 0},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatal(err)
	}
	sched := OnCallSchedule{
		Name:            "esc-sched",
		Timezone:        "Asia/Shanghai",
		Rotation:        "fixed",
		Start:           time.Now().Add(-2 * time.Hour).Format(time.RFC3339),
		Members:         []int{1},
		Weekdays:        []int{0, 1, 2, 3, 4, 5, 6},
		WindowStart:     "00:00",
		WindowEnd:       "00:00",
		FallbackUserIDs: []int{2},
		Enabled:         true,
	}
	if err := SaveOnCallSchedule(&sched, "oncall-primary"); err != nil {
		t.Fatal(err)
	}

	rule := Rule{
		Name:              "esc-rule",
		MType:             "1",
		Status:            "0",
		ZIDs:              "41",
		Channel:           "mail",
		ScheduleIDs:       []int{sched.ID},
		Conditions:        `[{"r_type":"host","r_func":"==","r_value":"prod-app"}]`,
		Stime:             "00:00",
		Etime:             "23:59",
		Sweek:             "0,1,2,3,4,5,6",
		EscalationMinutes: 10,
		EscalationChannel: "dingtalk_robot",
	}
	if _, err := AddRule(&rule); err != nil {
		t.Fatal(err)
	}

	past := time.Now().Add(-15 * time.Minute)
	unacked := Alarm{
		ZID:           41,
		Host:          "prod-app",
		Hostname:      "prod-app",
		HostsIP:       "10.0.0.8",
		TriggerID:     7001,
		Message:       "Service down",
		Level:         "5",
		Status:        "1",
		EventID:       88001,
		OccurTime:     past,
		MatchedRuleID: rule.ID,
	}
	acked := Alarm{
		ZID:           41,
		Host:          "prod-app",
		Hostname:      "prod-app",
		TriggerID:     7002,
		Message:       "Already acked",
		Level:         "4",
		Status:        "1",
		EventID:       88002,
		OccurTime:     past,
		MatchedRuleID: rule.ID,
	}
	if _, err := AddAlarm(&unacked); err != nil {
		t.Fatal(err)
	}
	if _, err := AddAlarm(&acked); err != nil {
		t.Fatal(err)
	}
	if n, err := AcknowledgeAlarms([]int{acked.ID}, "oncall-primary", "handling now", false); err != nil || n != 1 {
		t.Fatalf("AcknowledgeAlarms failed: n=%d err=%v", n, err)
	}

	escalated, err := ProcessDueEscalations(time.Now())
	if err != nil {
		t.Fatalf("ProcessDueEscalations failed: %v", err)
	}
	if escalated != 1 {
		t.Fatalf("expected 1 escalated alarm, got %d", escalated)
	}

	raw, err := CacheRPop("dingtalk_robot")
	if err != nil || raw == "" {
		t.Fatalf("expected escalated event in dingtalk_robot queue, err=%v", err)
	}
	var ev Event
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.ToUsers != "2" || !strings.Contains(ev.Message, "[告警升级/超时10分钟未认领]") {
		t.Fatalf("unexpected escalated event: %+v", ev)
	}

	var refreshed Alarm
	if err := db.First(&refreshed, unacked.ID).Error; err != nil {
		t.Fatal(err)
	}
	if refreshed.EscalationStep != 1 || refreshed.NotifyStatus != strconv.Itoa(NotifyEscalated) {
		t.Fatalf("unexpected alarm escalation state: %+v", refreshed)
	}
}
