package model

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"io"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"
	appLogger "zbxtable/pkg/logger"
)

// Opt-in integration test. A unique schema is created and removed in the test
// database; normal CI runs skip it without requiring a database service.
func TestOnCallPostgresIntegration(t *testing.T) {
	dsn := os.Getenv("ZBXTABLE_ONCALL_TEST_DSN")
	if dsn == "" {
		t.Skip("set ZBXTABLE_ONCALL_TEST_DSN to a PostgreSQL test database URL")
	}
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		t.Fatal("test DSN must be a PostgreSQL URL")
	}
	control, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := control.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	schemaName := fmt.Sprintf("oncall_test_%d", time.Now().UnixNano())
	if err := control.Exec("CREATE SCHEMA " + schemaName).Error; err != nil {
		t.Fatal(err)
	}
	defer control.Exec("DROP SCHEMA " + schemaName + " CASCADE")
	q := u.Query()
	q.Set("search_path", schemaName)
	u.RawQuery = q.Encode()
	db, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	conn, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	oldLog := appLogger.Log
	appLogger.Log = logrus.New()
	appLogger.Log.SetOutput(io.Discard)
	defer func() { appLogger.Log = oldLog }()
	old := DB
	DB = db
	defer func() { DB = old }()
	// Start with an older rule table to exercise additive startup migration.
	if err := db.Exec("CREATE TABLE zbxtable_rule (id bigserial PRIMARY KEY, name text, m_type text, status text)").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("INSERT INTO zbxtable_rule (name,m_type,status) VALUES ('legacy-before-oncall','1','0')").Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&User{}, &UserGroup{}, &Rule{}, &OnCallSchedule{}, &OnCallOverride{}, &OnCallAudit{}, &OnCallDispatchSnapshot{}, &Alarm{}, &ZabbixInstance{}); err != nil {
		t.Fatal(err)
	}
	var legacy Rule
	if err := db.Where("name = ?", "legacy-before-oncall").First(&legacy).Error; err != nil || legacy.ScheduleIDs != nil {
		t.Fatalf("old rule migration lost data: %+v %v", legacy, err)
	}
	users := []User{{ID: 1, Username: "test-admin", Role: "admin", Email: "one@example.test"}, {ID: 2, Username: "two", Wechat: "two"}, {ID: 3, Username: "fallback", Email: "backup@example.test", Wechat: "backup"}}
	if err := db.Create(&users).Error; err != nil {
		t.Fatal(err)
	}
	s := dispatchTestSchedule()
	s.ID = 0
	s.FallbackGroupIDs = []int{}
	if err := SaveOnCallSchedule(&s, "test-admin"); err != nil {
		t.Fatal(err)
	}
	at, _ := time.Parse(time.RFC3339, "2026-11-20T10:00:00+08:00")
	d, err := LoadScheduleDelivery(s.ID, at, "mail")
	if err != nil || !reflect.DeepEqual(d.UserIDs, []int{1}) {
		t.Fatalf("database primary: %+v %v", d, err)
	}
	d, err = LoadScheduleDelivery(s.ID, at, "wechat")
	if err != nil || d.Source != "fallback" || !reflect.DeepEqual(d.UserIDs, []int{3}) {
		t.Fatalf("database fallback: %+v %v", d, err)
	}
	o := OnCallOverride{ScheduleID: s.ID, Start: "2026-11-20T09:30:00+08:00", End: "2026-11-20T10:30:00+08:00", UserIDs: []int{2}, Reason: "temporary cover"}
	if err := AddOnCallOverride(&o, "test-admin"); err != nil {
		t.Fatal(err)
	}
	other := o
	other.ID = 0
	if err := AddOnCallOverride(&other, "test-admin"); err == nil {
		t.Fatal("overlapping override persisted")
	}
	d, err = LoadScheduleDelivery(s.ID, at, "wechat")
	if err != nil || d.Source != "override" || !reflect.DeepEqual(d.UserIDs, []int{2}) {
		t.Fatalf("database override: %+v %v", d, err)
	}
	if err := RemoveOnCallOverride(s.ID, o.ID, "test-admin"); err != nil {
		t.Fatal(err)
	}
	r := Rule{Name: "oncall-test", MType: "1", Channel: "mail", ScheduleIDs: []int{s.ID}, Conditions: `[{"r_type":"host","r_func":"==","r_value":"prod"}]`}
	if _, err := AddRule(&r); err != nil {
		t.Fatal(err)
	}
	if err := DeleteOnCallSchedule(s.ID, "test-admin"); err == nil {
		t.Fatal("deleted referenced schedule")
	}
	loaded, err := GetRuleByID(r.ID)
	if err != nil || !reflect.DeepEqual(loaded.ScheduleIDs, []int{s.ID}) {
		t.Fatalf("rule JSON lost: %+v %v", loaded, err)
	}
	snap, err := SnapshotDispatchRecipients(41, 123, DispatchChannel{Channel: "mail", UserIDs: []int{1}, Schedules: []ScheduleDecision{{ScheduleID: s.ID, UserIDs: []int{1}, Source: "rotation"}}})
	if err != nil || !reflect.DeepEqual(snap.UserIDs, []int{1}) {
		t.Fatalf("snapshot persistence: %+v %v", snap, err)
	}
	snap, err = SnapshotDispatchRecipients(41, 123, DispatchChannel{Channel: "mail", UserIDs: []int{3}, Schedules: []ScheduleDecision{{ScheduleID: s.ID, UserIDs: []int{3}, Source: "fallback"}}})
	if err != nil || !reflect.DeepEqual(snap.UserIDs, []int{1}) {
		t.Fatalf("snapshot changed on retry: %+v %v", snap, err)
	}
	s.Members = []int{2, 1}
	s.Enabled = false
	if err := SaveOnCallSchedule(&s, "test-admin"); err != nil {
		t.Fatal(err)
	}
	check, _, err := LoadOnCallSchedule(s.ID)
	if err != nil || check.Enabled || !reflect.DeepEqual(check.Members, []int{2, 1}) {
		t.Fatalf("schedule update serialization: %+v %v", check, err)
	}
	channels, err := ResolveEventDispatchRecipients([]Rule{r}, nil, at, 41, 123, LoadScheduleDelivery, FindOnCallDispatchSnapshot)
	if err != nil || !reflect.DeepEqual(channels[0].UserIDs, []int{1}) {
		t.Fatalf("retry snapshot after disabling: %+v %v", channels, err)
	}
	r.ScheduleIDs = nil
	if err := UpdateRule(&r, "test-admin"); err != nil {
		t.Fatal(err)
	}
	loaded, _ = GetRuleByID(r.ID)
	if len(loaded.ScheduleIDs) != 1 {
		t.Fatal("legacy update erased schedule")
	}
	r.ScheduleIDs = []int{}
	if err := UpdateRule(&r, "test-admin"); err != nil {
		t.Fatal(err)
	}
	loaded, _ = GetRuleByID(r.ID)
	if len(loaded.ScheduleIDs) != 0 {
		t.Fatal("explicit remove failed")
	}

	// Exercise actual GenAlert with the database and in-memory queue. No worker
	// runs here, so no email/webhook or other external notification is sent.
	now := time.Now()
	runtimeSchedule := dispatchTestSchedule()
	runtimeSchedule.ID = 0
	runtimeSchedule.Name = "runtime ops"
	runtimeSchedule.Rotation = "fixed"
	runtimeSchedule.Start = now.AddDate(0, 0, -1).Format(time.RFC3339)
	runtimeSchedule.Members = []int{1}
	runtimeSchedule.FallbackGroupIDs = []int{}
	if err := SaveOnCallSchedule(&runtimeSchedule, "test-admin"); err != nil {
		t.Fatal(err)
	}
	runtimeRule := Rule{Name: "runtime routing", MType: "1", ZIDs: "41", Status: "0", Channel: "mail,wechat", ScheduleIDs: []int{runtimeSchedule.ID}, Conditions: `[{"r_type":"host","r_func":"==","r_value":"prod"}]`, Stime: "00:00", Etime: "23:59", Sweek: "0,1,2,3,4,5,6"}
	if _, err := AddRule(&runtimeRule); err != nil {
		t.Fatal(err)
	}
	fixedRule := runtimeRule
	fixedRule.ID = 0
	fixedRule.Name = "fixed cc"
	fixedRule.ScheduleIDs = nil
	fixedRule.Channel = "mail"
	fixedRule.UserIds = "3"
	if _, err := AddRule(&fixedRule); err != nil {
		t.Fatal(err)
	}
	queueMu.Lock()
	oldQueues := queueStore
	queueStore = make(map[string]*Queue)
	queueMu.Unlock()
	defer func() { queueMu.Lock(); queueStore = oldQueues; queueMu.Unlock() }()
	makeAlarm := func(eventID int64) Alarm {
		a := Alarm{ZID: 41, Host: "prod", EventID: eventID, OccurTime: now, Status: "1"}
		if err := db.Create(&a).Error; err != nil {
			t.Fatal(err)
		}
		return a
	}
	alarm := makeAlarm(9001)
	GenAlert(&alarm)
	for ch, want := range map[string]string{"mail": "1,3", "wechat": "3"} {
		raw, e := CacheRPop(ch)
		if e != nil || raw == "" {
			t.Fatalf("missing %s queue: %v", ch, e)
		}
		var event Event
		if err := json.Unmarshal([]byte(raw), &event); err != nil {
			t.Fatal(err)
		}
		if event.ToUsers != want {
			t.Fatalf("%s ToUsers %s want %s", ch, event.ToUsers, want)
		}
	}
	var status Alarm
	if err := db.First(&status, alarm.ID).Error; err != nil || status.NotifyStatus != "0" {
		t.Fatalf("notification status %+v %v", status, err)
	}
	mute := Rule{Name: "mute mail", MType: "3", ZIDs: "41", Status: "0", Channel: "mail", Conditions: runtimeRule.Conditions, Stime: now.Add(-time.Hour).Format("2006-01-02 15:04:05"), Etime: now.Add(time.Hour).Format("2006-01-02 15:04:05")}
	if _, err := AddRule(&mute); err != nil {
		t.Fatal(err)
	}
	mutedAlarm := makeAlarm(9002)
	GenAlert(&mutedAlarm)
	if raw, _ := CacheRPop("mail"); raw != "" {
		t.Fatal("muted mail was queued")
	}
	if raw, _ := CacheRPop("wechat"); raw == "" {
		t.Fatal("unmuted channel not queued")
	}
	// A schedule that has no deliverable users or fallback must not report success.
	if err := db.Delete(&mute).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Delete(&fixedRule).Error; err != nil {
		t.Fatal(err)
	}
	runtimeSchedule.FallbackUserIDs = []int{}
	runtimeSchedule.Members = []int{2}
	runtimeSchedule.Enabled = false
	if err := SaveOnCallSchedule(&runtimeSchedule, "test-admin"); err != nil {
		t.Fatal(err)
	}
	gapAlarm := makeAlarm(9003)
	GenAlert(&gapAlarm)
	status = Alarm{}
	if err := db.First(&status, gapAlarm.ID).Error; err != nil {
		t.Fatal(err)
	}
	if status.NotifyStatus != "3" {
		t.Fatalf("uncovered alarm falsely succeeded: %+v", status)
	}
	for _, ch := range []string{"mail", "wechat"} {
		if raw, _ := CacheRPop(ch); raw != "" {
			t.Fatalf("uncovered %s queued", ch)
		}
	}
	// The previous event still uses its recorded users despite this new gap.
	retry := makeAlarm(9001)
	GenAlert(&retry)
	if raw, _ := CacheRPop("mail"); raw == "" {
		t.Fatal("retry lost recorded mail recipients")
	}
	if raw, _ := CacheRPop("wechat"); raw == "" {
		t.Fatal("retry lost recorded wechat recipients")
	}
	if err := DeleteOnCallSchedule(s.ID, "test-admin"); err != nil {
		t.Fatal(err)
	}
	var count int64
	if err := db.Model(&OnCallAudit{}).Where("schedule_id = ?", s.ID).Count(&count).Error; err != nil || count < 5 {
		t.Fatalf("audit lost: %d %v", count, err)
	}
}
