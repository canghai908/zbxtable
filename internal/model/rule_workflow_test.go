package model

import (
	"reflect"
	"sort"
	"testing"
	"time"
)

const workflowConditions = `{"version":2,"op":"and","children":[{"field":"host","operator":"like","value":"prod-"},{"op":"or","children":[{"field":"trigger","operator":"like","value":"CPU"},{"field":"trigger","operator":"like","value":"Memory"}]}]}`

func TestDispatchRuleWorkflow(t *testing.T) {
	loc := time.FixedZone("Asia/Shanghai", 8*60*60)
	at := time.Date(2026, 11, 20, 10, 0, 0, 0, loc)
	custom := Rule{ID: 1, MType: "1", Conditions: workflowConditions, Stime: "09:00", Etime: "18:00", Sweek: "1,2,3,4,5"}
	fallback := Rule{ID: 9, MType: "2", Stime: "00:00", Etime: "23:59", Sweek: "0,1,2,3,4,5,6"}
	for _, tt := range []struct {
		name, host, message string
		at                  time.Time
		ids                 []int
	}{
		{"CPU routes", "prod-db", "CPU high", at, []int{1, 2}},
		{"Memory routes via OR", "prod-db", "Memory high", at, []int{1, 2}},
		{"AND host mismatch falls back", "dev-db", "CPU high", at, []int{9}},
		{"OR both miss falls back", "prod-db", "Disk high", at, []int{9}},
		{"outside window falls back", "prod-db", "CPU high", at.Add(10 * time.Hour), []int{9}},
		{"weekend falls back", "prod-db", "CPU high", at.AddDate(0, 0, 1), []int{9}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			second := custom
			second.ID = 2
			got, kind := matchDispatchRules([]dispatchRuleSet{{rules: []Rule{custom, second}, ruleType: "1"}, {rules: []Rule{fallback}, ruleType: "2"}}, map[string]any{"host": tt.host, "trigger": tt.message}, tt.at, GetCELMatcher())
			ids := []int{}
			for _, r := range got {
				ids = append(ids, r.ID)
			}
			if !reflect.DeepEqual(ids, tt.ids) {
				t.Fatalf("got %v want %v", ids, tt.ids)
			}
			if (kind == "2") != (tt.ids[0] == 9) {
				t.Fatalf("wrong rule type %s", kind)
			}
		})
	}
}

func TestDispatchOvernightAndBoundaries(t *testing.T) {
	r := Rule{MType: "1", Conditions: `[{"r_type":"host","r_func":"==","r_value":"prod"}]`, Stime: "22:00", Etime: "06:00", Sweek: "0,1,2,3,4,5,6"}
	for _, tt := range []struct {
		hour, minute int
		blocked      bool
	}{{21, 59, true}, {22, 0, false}, {23, 59, false}, {0, 0, false}, {6, 0, false}, {6, 1, true}, {12, 0, true}} {
		at := time.Date(2026, 11, 20, tt.hour, tt.minute, 0, 0, time.UTC)
		if got := isNoneAlarm(at, &r); got != tt.blocked {
			t.Errorf("%02d:%02d blocked=%v want %v", tt.hour, tt.minute, got, tt.blocked)
		}
	}
}

func TestMuteRuleConditionTimeChannel(t *testing.T) {
	loc := time.Local
	at := time.Date(2026, 11, 20, 10, 0, 0, 0, loc)
	r := Rule{MType: "3", Conditions: workflowConditions, Stime: "2026-11-20 09:00:00", Etime: "2026-11-20 11:00:00", Channel: "wechat_robot,mail"}
	for _, tt := range []struct {
		name, host, message, channel string
		at                           time.Time
		want                         bool
	}{
		{"CPU muted", "prod-db", "CPU high", "mail", at, true},
		{"Memory muted", "prod-db", "Memory high", "mail", at, true},
		{"host mismatch", "dev-db", "CPU high", "mail", at, false},
		{"condition mismatch", "prod-db", "Disk high", "mail", at, false},
		{"other channel sends", "prod-db", "CPU high", "sms", at, false},
		{"robot alias", "prod-db", "CPU high", "wechatrobot", at, true},
		{"robot hyphen alias", "prod-db", "CPU high", "wechat-robot", at, true},
		{"before start", "prod-db", "CPU high", "mail", at.Add(-2 * time.Hour), false},
		{"expired", "prod-db", "CPU high", "mail", at.Add(2 * time.Hour), false},
		{"start inclusive", "prod-db", "CPU high", "mail", at.Add(-time.Hour), true},
		{"end inclusive", "prod-db", "CPU high", "mail", at.Add(time.Hour), true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e := Event{Host: tt.host, Message: tt.message, Channel: tt.channel, OccurTime: tt.at}
			// Same three predicates used after IsMuted loads active instance rules.
			got := MeetEventConditions(&e, &r) && !IsMuteTime(&e, &r) && !isMuteChannel(&e, &r)
			if got != tt.want {
				t.Fatalf("muted=%v want %v", got, tt.want)
			}
		})
	}
	r.Channel = ""
	if isMuteChannel(&Event{Channel: "mail"}, &r) {
		t.Fatal("empty channels should match every channel")
	}
	r.Conditions = "[]"
	if MeetEventConditions(&Event{}, &r) {
		t.Fatal("empty mute must not match all")
	}
}

func TestRuleRecipientsDeduplicate(t *testing.T) {
	users := getRuleUserIDs("1,2,1", "10,11", map[string][]string{"10": {"2", "3"}, "11": {"1", "4"}})
	sort.Strings(users)
	if !reflect.DeepEqual(users, []string{"1", "2", "3", "4"}) {
		t.Fatalf("recipients %v", users)
	}
}
