package model

import (
	"errors"
	"gorm.io/gorm/schema"
	"reflect"
	"sync"
	"testing"
	"time"
)

func dispatchTestSchedule() OnCallSchedule {
	return OnCallSchedule{ID: 1, Name: "ops", Timezone: "Asia/Shanghai", Rotation: "daily", Start: "2026-11-20T09:00:00+08:00", Members: []int{1, 2}, Weekdays: []int{0, 1, 2, 3, 4, 5, 6}, WindowStart: "00:00", WindowEnd: "00:00", FallbackUserIDs: []int{3}, FallbackGroupIDs: []int{8}, Enabled: true}
}
func TestOnCallDeliveryFallbackAndEligibility(t *testing.T) {
	s := dispatchTestSchedule()
	at, _ := time.Parse(time.RFC3339, "2026-11-20T10:00:00+08:00")
	users := map[int]User{1: {ID: 1, Email: "one@example.test", Status: 0}, 2: {ID: 2, Wechat: "two", Status: 0}, 3: {ID: 3, Email: "three@example.test", Wechat: "three", Status: 0}, 4: {ID: 4, Email: "four@example.test", Status: 1}}
	groups := map[int][]int{8: {3, 4}}
	for _, tt := range []struct {
		name, channel string
		ids           []int
		source        string
	}{
		{"primary email", "mail", []int{1}, "rotation"},
		{"primary missing address fallback", "wechat", []int{3}, "fallback"},
		{"no valid primary or fallback", "wechat_robot", []int{}, "uncovered"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d, err := ResolveScheduleDelivery(s, nil, at, tt.channel, users, groups)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(d.UserIDs, tt.ids) || d.Source != tt.source {
				t.Fatalf("got %+v", d)
			}
		})
	}
	u := users[1]
	u.Status = 1
	users[1] = u
	d, err := ResolveScheduleDelivery(s, nil, at, "mail", users, groups)
	if err != nil || d.Source != "fallback" || !reflect.DeepEqual(d.UserIDs, []int{3}) {
		t.Fatalf("disabled primary: %+v %v", d, err)
	}
	override := OnCallOverride{ScheduleID: 1, Start: "2026-11-20T09:30:00+08:00", End: "2026-11-20T11:30:00+08:00", UserIDs: []int{2}, Reason: "cover"}
	d, err = ResolveScheduleDelivery(s, []OnCallOverride{override}, at, "wechat", users, groups)
	if err != nil || d.Source != "override" || !reflect.DeepEqual(d.UserIDs, []int{2}) {
		t.Fatalf("override: %+v %v", d, err)
	}
	s.WindowStart = "20:00"
	s.WindowEnd = "22:00"
	d, err = ResolveScheduleDelivery(s, nil, at, "mail", users, groups)
	if err != nil || d.Source != "fallback" {
		t.Fatalf("gap: %+v %v", d, err)
	}
}

func TestOnCallDispatchMixedRecipientsAndFailureIsolation(t *testing.T) {
	at := time.Now()
	rules := []Rule{{UserIds: "1", GroupIds: "8", Channel: "mail,wechatrobot", ScheduleIDs: []int{10}}, {UserIds: "1,2", Channel: "mail", ScheduleIDs: []int{10}}}
	calls := 0
	loader := func(id int, at time.Time, ch string) (ScheduleDecision, error) {
		calls++
		return ScheduleDecision{ScheduleID: id, UserIDs: []int{2, 3}, Source: "rotation"}, nil
	}
	channels, err := ResolveDispatchRecipients(rules, map[string][]string{"8": {"1", "2"}}, at, loader)
	if err != nil || len(channels) != 2 {
		t.Fatalf("%+v %v", channels, err)
	}
	for _, ch := range channels {
		if !reflect.DeepEqual(ch.UserIDs, []int{1, 2, 3}) {
			t.Fatalf("not deduped %+v", ch)
		}
	}
	if channels[1].Channel != "wechat_robot" || calls == 0 {
		t.Fatal("channel alias not normalized")
	}
	channels, err = ResolveDispatchRecipients(rules, nil, at, func(int, time.Time, string) (ScheduleDecision, error) {
		return ScheduleDecision{}, errors.New("offline")
	})
	if err != nil || len(channels[0].UserIDs) != 2 || channels[0].Schedules[0].Source != "error" {
		t.Fatalf("fixed recipients lost: %+v %v", channels, err)
	}
	legacy, err := ResolveDispatchRecipients([]Rule{{UserIds: "1,2,1", Channel: "mail"}}, nil, at, func(int, time.Time, string) (ScheduleDecision, error) {
		t.Fatal("legacy loaded schedule")
		return ScheduleDecision{}, nil
	})
	if err != nil || !reflect.DeepEqual(legacy[0].UserIDs, []int{1, 2}) {
		t.Fatalf("legacy changed %+v %v", legacy, err)
	}
}

func TestOnCallSnapshotBeforeCurrentShift(t *testing.T) {
	saved := OnCallDispatchSnapshot{UserIDs: []int{5}, Decisions: []ScheduleDecision{{ScheduleID: 1, UserIDs: []int{5}, Source: "rotation"}}}
	channels, err := ResolveEventDispatchRecipients([]Rule{{ScheduleIDs: []int{1}, Channel: "mail", UserIds: "9"}}, nil, time.Now(), 41, 100,
		func(int, time.Time, string) (ScheduleDecision, error) {
			t.Fatal("current shift queried despite snapshot")
			return ScheduleDecision{}, errors.New("gap")
		},
		func(zid int, eventID int64, ch string) (OnCallDispatchSnapshot, bool, error) {
			if zid != 41 || eventID != 100 || ch != "mail" {
				t.Fatal("wrong snapshot key")
			}
			return saved, true, nil
		})
	if err != nil || !reflect.DeepEqual(channels[0].UserIDs, []int{5}) {
		t.Fatalf("retry changed recipient %+v %v", channels, err)
	}
}

func TestOnCallSnapshotSchema(t *testing.T) {
	parsed, err := schema.Parse(&OnCallDispatchSnapshot{}, &sync.Map{}, schema.NamingStrategy{})
	if err != nil {
		t.Fatal(err)
	}
	for field, column := range map[string]string{"ZID": "zid", "EventID": "event_id", "Channel": "channel", "UserIDs": "user_ids", "Decisions": "decisions"} {
		if got := parsed.LookUpField(field).DBName; got != column {
			t.Fatalf("%s DB column %s want %s", field, got, column)
		}
	}
	if parsed.LookUpField("UserIDs").Serializer == nil || parsed.LookUpField("Decisions").Serializer == nil {
		t.Fatal("snapshot JSON serialization missing")
	}
}
