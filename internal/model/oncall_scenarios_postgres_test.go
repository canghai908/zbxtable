package model

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlog "gorm.io/gorm/logger"
	appLogger "zbxtable/pkg/logger"
)

// Each integration run owns a schema, queue and config cache. Never uses the
// application's configured database or notification credentials.
func onCallScenarioDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("ZBXTABLE_ONCALL_TEST_DSN")
	if dsn == "" {
		t.Skip("set ZBXTABLE_ONCALL_TEST_DSN to a PostgreSQL test database URL")
	}
	u, err := url.Parse(dsn)
	if err != nil || (u.Scheme != "postgres" && u.Scheme != "postgresql") {
		t.Fatal("test DSN must be a PostgreSQL URL")
	}
	control, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormlog.Default.LogMode(gormlog.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	raw, err := control.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { raw.Close() })
	schema := fmt.Sprintf("oncall_scenarios_%d", time.Now().UnixNano())
	if err = control.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if e := control.Exec("DROP SCHEMA " + schema + " CASCADE").Error; e != nil {
			t.Errorf("cleanup schema: %v", e)
		}
	})
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	db, err := gorm.Open(postgres.Open(u.String()), &gorm.Config{Logger: gormlog.Default.LogMode(gormlog.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	conn, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	if err = db.AutoMigrate(&User{}, &UserGroup{}, &Rule{}, &OnCallSchedule{}, &OnCallOverride{}, &OnCallAudit{}, &OnCallDispatchSnapshot{}, &Alarm{}, &ZabbixInstance{}, &Config{}, &EventLog{}); err != nil {
		t.Fatal(err)
	}
	oldDB, oldLog := DB, appLogger.Log
	DB = db
	appLogger.Log = logrus.New()
	appLogger.Log.SetOutput(io.Discard)
	queueMu.Lock()
	oldQueues := queueStore
	queueStore = make(map[string]*Queue)
	queueMu.Unlock()
	configValueCacheMu.Lock()
	oldCache := configValueCache
	configValueCache = make(map[string]string)
	configValueCacheMu.Unlock()
	t.Cleanup(func() {
		DB = oldDB
		appLogger.Log = oldLog
		queueMu.Lock()
		queueStore = oldQueues
		queueMu.Unlock()
		configValueCacheMu.Lock()
		configValueCache = oldCache
		configValueCacheMu.Unlock()
	})
	return db
}

func TestOnCallDispatchScenarioMatrixPostgres(t *testing.T) {
	db := onCallScenarioDB(t)
	users := []User{
		{ID: 1, Username: "primary", Role: "admin", Email: "primary@example.test", Wechat: "primary", WechatRobotKey: "test-primary"},
		{ID: 2, Username: "next-shift", Email: "next@example.test", Wechat: "next"},
		{ID: 3, Username: "cover", Email: "cover@example.test", Wechat: "cover"},
		{ID: 4, Username: "fallback", Email: "fallback@example.test", Wechat: "fallback"},
		{ID: 5, Username: "global", Email: "global@example.test", Wechat: "global"},
		{ID: 6, Username: "missing-address"},
		{ID: 7, Username: "disabled", Status: 1, Email: "disabled@example.test"},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&UserGroup{ID: 10, Name: "ops", Member: "1,3,4"}).Error; err != nil {
		t.Fatal(err)
	}
	type scenario struct {
		name   string
		setup  func(*OnCallSchedule, *Rule, *Alarm) []Rule
		want   map[string]string
		source string
		status string
	}
	fixed := func(uid string) func(*OnCallSchedule, *Rule, *Alarm) []Rule {
		return func(_ *OnCallSchedule, r *Rule, _ *Alarm) []Rule { r.ScheduleIDs = nil; r.UserIds = uid; return nil }
	}
	cases := []scenario{
		{name: "legacy_fixed_user", setup: func(s *OnCallSchedule, r *Rule, a *Alarm) []Rule {
			fixed("3")(s, r, a)
			r.Conditions = `[{"r_type":"host","r_func":"==","r_value":"prod-db"}]`
			return nil
		}, want: map[string]string{"mail": "3"}},
		{name: "nested_AND_OR_CPU", want: map[string]string{"mail": "1"}, source: "rotation"},
		{name: "nested_OR_Memory", setup: func(_ *OnCallSchedule, _ *Rule, a *Alarm) []Rule { a.Message = "Memory high"; return nil }, want: map[string]string{"mail": "1"}, source: "rotation"},
		{name: "AND_miss_instance_default", setup: func(_ *OnCallSchedule, _ *Rule, a *Alarm) []Rule { a.Host = "dev-db"; return nil }, want: map[string]string{"mail": "4"}},
		{name: "OR_miss_instance_default", setup: func(_ *OnCallSchedule, _ *Rule, a *Alarm) []Rule { a.Message = "Disk high"; return nil }, want: map[string]string{"mail": "4"}},
		{name: "other_instance_global_default", setup: func(_ *OnCallSchedule, _ *Rule, a *Alarm) []Rule { a.ZID = 142; return nil }, want: map[string]string{"mail": "5"}},
		{name: "disabled_rule_default", setup: func(_ *OnCallSchedule, r *Rule, _ *Alarm) []Rule { r.Status = "1"; return nil }, want: map[string]string{"mail": "4"}},
		{name: "user_group_schedule_deduplication", setup: func(_ *OnCallSchedule, r *Rule, _ *Alarm) []Rule {
			r.UserIds = "1,3"
			r.GroupIds = "10"
			other := *r
			other.Name = "second matching rule"
			other.UserIds = "1,4"
			return []Rule{other}
		}, want: map[string]string{"mail": "1,3,4"}, source: "rotation"},
		{name: "fixed_parallel_members", setup: func(s *OnCallSchedule, _ *Rule, _ *Alarm) []Rule {
			s.Rotation = "fixed"
			s.Members = []int{1, 2}
			return nil
		}, want: map[string]string{"mail": "1,2"}, source: "rotation"},
		{name: "daily_handoff", setup: func(s *OnCallSchedule, _ *Rule, _ *Alarm) []Rule {
			s.Start = time.Now().Add(-25 * time.Hour).Format(time.RFC3339)
			return nil
		}, want: map[string]string{"mail": "2"}, source: "rotation"},
		{name: "weekly_handoff", setup: func(s *OnCallSchedule, _ *Rule, _ *Alarm) []Rule {
			s.Rotation = "weekly"
			s.Start = time.Now().AddDate(0, 0, -8).Format(time.RFC3339)
			return nil
		}, want: map[string]string{"mail": "2"}, source: "rotation"},
		{name: "active_override", want: map[string]string{"mail": "3"}, source: "override"},
		{name: "expired_override", want: map[string]string{"mail": "1"}, source: "rotation"},
		{name: "future_override", want: map[string]string{"mail": "1"}, source: "rotation"},
		{name: "coverage_gap_fallback", setup: func(s *OnCallSchedule, _ *Rule, _ *Alarm) []Rule {
			s.Weekdays = []int{int(time.Now().In(time.Local).Weekday()+1) % 7}
			return nil
		}, want: map[string]string{"mail": "4"}, source: "fallback"},
		{name: "before_schedule_start_fallback", setup: func(s *OnCallSchedule, _ *Rule, _ *Alarm) []Rule {
			s.Start = time.Now().Add(time.Hour).Format(time.RFC3339)
			return nil
		}, want: map[string]string{"mail": "4"}, source: "fallback"},
		{name: "disabled_schedule_fallback_group", setup: func(s *OnCallSchedule, _ *Rule, _ *Alarm) []Rule {
			s.Enabled = false
			s.FallbackUserIDs = nil
			s.FallbackGroupIDs = []int{10}
			return nil
		}, want: map[string]string{"mail": "1,3,4"}, source: "fallback"},
		{name: "missing_channel_address_fallback", setup: func(s *OnCallSchedule, _ *Rule, _ *Alarm) []Rule { s.Members = []int{6}; return nil }, want: map[string]string{"mail": "4"}, source: "fallback"},
		{name: "disabled_member_fallback", setup: func(s *OnCallSchedule, _ *Rule, _ *Alarm) []Rule { s.Members = []int{7}; return nil }, want: map[string]string{"mail": "4"}, source: "fallback"},
		{name: "uncovered_is_not_success", setup: func(s *OnCallSchedule, _ *Rule, _ *Alarm) []Rule {
			s.Members = []int{6}
			s.FallbackUserIDs = nil
			return nil
		}, want: map[string]string{}, status: "3"},
		{name: "mail_muted_wechat_delivers", setup: func(_ *OnCallSchedule, r *Rule, _ *Alarm) []Rule {
			r.Channel = "mail,wechat"
			return []Rule{{Name: "mail maintenance", MType: "3", ZIDs: "41", Status: "0", Channel: "mail", Conditions: workflowConditions, Stime: time.Now().Add(-time.Hour).Format("2006-01-02 15:04:05"), Etime: time.Now().Add(time.Hour).Format("2006-01-02 15:04:05")}}
		}, want: map[string]string{"wechat": "1"}, source: "rotation"},
		{name: "all_channels_muted", setup: func(_ *OnCallSchedule, _ *Rule, _ *Alarm) []Rule {
			return []Rule{{Name: "all maintenance", MType: "3", ZIDs: "41", Status: "0", Channel: "", Conditions: workflowConditions, Stime: time.Now().Add(-time.Hour).Format("2006-01-02 15:04:05"), Etime: time.Now().Add(time.Hour).Format("2006-01-02 15:04:05")}}
		}, want: map[string]string{}, status: "1"},
		{name: "expired_mute_delivers", setup: func(_ *OnCallSchedule, _ *Rule, _ *Alarm) []Rule {
			return []Rule{{Name: "expired maintenance", MType: "3", ZIDs: "41", Status: "0", Channel: "mail", Conditions: workflowConditions, Stime: time.Now().Add(-2 * time.Hour).Format("2006-01-02 15:04:05"), Etime: time.Now().Add(-time.Hour).Format("2006-01-02 15:04:05")}}
		}, want: map[string]string{"mail": "1"}, source: "rotation"},
	}
	for i, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			queueMu.Lock()
			queueStore = make(map[string]*Queue)
			queueMu.Unlock()
			for _, table := range []any{&Rule{}, &OnCallOverride{}, &OnCallDispatchSnapshot{}} {
				if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error; err != nil {
					t.Fatal(err)
				}
			}
			now := time.Now()
			s := OnCallSchedule{Name: tt.name, Timezone: "Asia/Shanghai", Rotation: "daily", Start: now.Add(-2 * time.Hour).Format(time.RFC3339), Members: []int{1, 2}, Weekdays: []int{0, 1, 2, 3, 4, 5, 6}, WindowStart: "00:00", WindowEnd: "00:00", FallbackUserIDs: []int{4}, Enabled: true}
			r := Rule{Name: tt.name, MType: "1", Status: "0", ZIDs: "41", Channel: "mail", Conditions: workflowConditions, Stime: "00:00", Etime: "23:59", Sweek: "0,1,2,3,4,5,6", ScheduleIDs: []int{-1}}
			a := Alarm{ZID: 41, Host: "prod-db", Hostname: "production database", HostsIP: "192.0.2.10", Hgroup: "Databases", ItemName: "CPU utilization", ItemID: 501, ItemValue: "97", Hkey: "system.cpu.util", Message: "CPU high", Level: "High", Status: "1", EventID: int64(20000 + i), OccurTime: now}
			var extra []Rule
			if tt.setup != nil {
				extra = tt.setup(&s, &r, &a)
			}
			// Keep coverage testing in the schedule's timezone, independent of runner TZ.
			if tt.name == "coverage_gap_fallback" {
				loc, _ := time.LoadLocation(s.Timezone)
				s.Weekdays = []int{(int(now.In(loc).Weekday()) + 1) % 7}
			}
			if err := SaveOnCallSchedule(&s, "primary"); err != nil {
				t.Fatal(err)
			}
			if len(r.ScheduleIDs) > 0 {
				r.ScheduleIDs = []int{s.ID}
			}
			if strings.Contains(tt.name, "override") {
				start, end := now.Add(-time.Hour), now.Add(time.Hour)
				if tt.name == "expired_override" {
					start, end = now.Add(-90*time.Minute), now.Add(-time.Hour)
				}
				if tt.name == "future_override" {
					start, end = now.Add(time.Hour), now.Add(2*time.Hour)
				}
				o := OnCallOverride{ScheduleID: s.ID, Start: start.Format(time.RFC3339), End: end.Format(time.RFC3339), UserIDs: []int{3}, Reason: tt.name}
				if err := AddOnCallOverride(&o, "primary"); err != nil {
					t.Fatal(err)
				}
			}
			defaults := []Rule{{Name: "instance fallback", MType: "2", Status: "0", ZIDs: "41", Channel: "mail", UserIds: "4", Stime: "00:00", Etime: "23:59", Sweek: r.Sweek}, {Name: "global fallback", MType: "2", Status: "0", ZIDs: "*", Channel: "mail", UserIds: "5", Stime: "00:00", Etime: "23:59", Sweek: r.Sweek}}
			for _, row := range append(append([]Rule{r}, extra...), defaults...) {
				row.ID = 0
				if len(row.ScheduleIDs) > 0 {
					row.ScheduleIDs = []int{s.ID}
				}
				if _, err := AddRule(&row); err != nil {
					t.Fatal(err)
				}
			}
			// Exercise the same incoming data adapter used by the webhook, not a mocked matcher.
			tpl := EventTpl{HostsID: "101", HostHost: a.Host, Hostname: a.Hostname, HostsIP: a.HostsIP, HostGroup: a.Hgroup, EventTime: now.Format("2006-01-02 15:04:05"), Severity: a.Level, TriggerName: a.Message, TriggerKey: a.Hkey, TriggerValue: a.Status, ItemID: a.ItemID, ItemName: a.ItemName, ItemValue: a.ItemValue, EventID: a.EventID}
			body, err := json.Marshal(tpl)
			if err != nil {
				t.Fatal(err)
			}
			id, err := MsAdd(a.ZID, []byte(strings.ReplaceAll(string(body), `"`, `¦`)))
			if err != nil {
				t.Fatal(err)
			}
			loaded, err := GetAlarmByID(int(id))
			if err != nil {
				t.Fatal(err)
			}
			wantStatus := tt.status
			if wantStatus == "" {
				wantStatus = "0"
				if strings.HasSuffix(tt.name, "default") {
					wantStatus = "2"
				}
			}
			if loaded.NotifyStatus != wantStatus || loaded.ItemID != 501 || loaded.ItemValue != "97" {
				t.Fatalf("alarm status/data: %+v", loaded)
			}
			got := map[string]string{}
			for _, ch := range []string{"mail", "wechat", "wechat_robot"} {
				raw, err := CacheRPop(ch)
				if err != nil {
					t.Fatal(err)
				}
				if raw == "" {
					continue
				}
				var e Event
				if err := json.Unmarshal([]byte(raw), &e); err != nil {
					t.Fatal(err)
				}
				if e.EventID != a.EventID || e.ItemID != 501 || e.ItemValue != "97" {
					t.Fatalf("queue lost alarm fields: %+v", e)
				}
				got[ch] = e.ToUsers
				if another, _ := CacheRPop(ch); another != "" {
					t.Fatalf("duplicate %s event", ch)
				}
				if tt.source != "" {
					snap, found, err := FindOnCallDispatchSnapshot(a.ZID, a.EventID, ch)
					if err != nil || !found || len(snap.Decisions) == 0 || snap.Decisions[0].Source != tt.source {
						t.Fatalf("snapshot source: %+v %v", snap, err)
					}
				}
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("recipients %v want %v", got, tt.want)
			}
			t.Logf("alarm=%d schedule=%d event=%d recipients=%v notify_status=%s", id, s.ID, a.EventID, got, loaded.NotifyStatus)
		})
	}
}
