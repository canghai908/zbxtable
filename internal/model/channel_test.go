package model

import (
	"testing"
	"time"
)

func TestCELMatcher_Match(t *testing.T) {
	matcher := GetCELMatcher()
	data := map[string]any{
		"host":     "server-01",
		"severity": "Average",
		"item":     "CPU load",
	}

	tests := []struct {
		name  string
		conds []Conditions
		want  bool
	}{
		{
			name: "exact match",
			conds: []Conditions{
				{RType: "host", RFunc: "=", Rvalue: "server-01"},
			},
			want: true,
		},
		{
			name: "not equal match",
			conds: []Conditions{
				{RType: "severity", RFunc: "!=", Rvalue: "High"},
			},
			want: true,
		},
		{
			name: "contains match (like)",
			conds: []Conditions{
				{RType: "item", RFunc: "like", Rvalue: "CPU"},
			},
			want: true,
		},
		{
			name: "regex match (=~)",
			conds: []Conditions{
				{RType: "item", RFunc: "=~", Rvalue: "CPU.*"},
			},
			want: true,
		},
		{
			name: "multiple conditions match",
			conds: []Conditions{
				{RType: "host", RFunc: "=", Rvalue: "server-01"},
				{RType: "severity", RFunc: "=", Rvalue: "Average"},
			},
			want: true,
		},
		{
			name: "mismatch",
			conds: []Conditions{
				{RType: "host", RFunc: "=", Rvalue: "server-02"},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matcher.Match(tt.conds, data); got != tt.want {
				t.Errorf("CELMatcher.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchRule(t *testing.T) {
	matcher := GetCELMatcher()
	data := map[string]any{
		"host":     "server-01",
		"severity": "Average",
	}

	tests := []struct {
		name string
		rule Rule
		want bool
	}{
		{
			name: "default rule with empty conditions matches all",
			rule: Rule{MType: "2", Conditions: ""},
			want: true,
		},
		{
			name: "custom rule with empty conditions does not match",
			rule: Rule{MType: "1", Conditions: ""},
			want: false,
		},
		{
			name: "default rule with invalid conditions does not match",
			rule: Rule{MType: "2", Conditions: "invalid-json"},
			want: false,
		},
		{
			name: "default rule with valid conditions still evaluates expression",
			rule: Rule{MType: "2", Conditions: `[{"r_type":"host","r_func":"=","r_value":"server-01"}]`},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchRule(&tt.rule, data, matcher); got != tt.want {
				t.Fatalf("matchRule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchDispatchRules(t *testing.T) {
	matcher := GetCELMatcher()
	matchData := map[string]any{
		"host":     "server-01",
		"group":    "core",
		"item":     "CPU load",
		"key":      "system.cpu.load",
		"trigger":  "cpu high",
		"severity": "Average",
	}
	now := time.Date(2026, 2, 10, 10, 0, 0, 0, time.Local)

	t.Run("falls back to default rule when custom rules do not match", func(t *testing.T) {
		rules, ruleType := matchDispatchRules([]dispatchRuleSet{
			{
				rules: []Rule{
					{
						ID:         1,
						MType:      "1",
						Conditions: `[{"r_type":"host","r_func":"=","r_value":"server-02"}]`,
						Stime:      "00:00",
						Etime:      "23:59",
						Sweek:      "0,1,2,3,4,5,6",
					},
				},
				ruleType: "1",
			},
			{
				rules: []Rule{
					{
						ID:         2,
						MType:      "2",
						Conditions: "",
						Stime:      "00:00",
						Etime:      "23:59",
						Sweek:      "0,1,2,3,4,5,6",
					},
				},
				ruleType: "2",
			},
		}, matchData, now, matcher)

		if ruleType != "2" {
			t.Fatalf("matchDispatchRules() ruleType = %s, want 2", ruleType)
		}
		if len(rules) != 1 || rules[0].ID != 2 {
			t.Fatalf("matchDispatchRules() matched rules = %+v, want default rule", rules)
		}
	})

	t.Run("keeps custom rule priority when custom rule matches", func(t *testing.T) {
		rules, ruleType := matchDispatchRules([]dispatchRuleSet{
			{
				rules: []Rule{
					{
						ID:         3,
						MType:      "1",
						Conditions: `[{"r_type":"host","r_func":"=","r_value":"server-01"}]`,
						Stime:      "00:00",
						Etime:      "23:59",
						Sweek:      "0,1,2,3,4,5,6",
					},
				},
				ruleType: "1",
			},
			{
				rules: []Rule{
					{
						ID:         4,
						MType:      "2",
						Conditions: "",
						Stime:      "00:00",
						Etime:      "23:59",
						Sweek:      "0,1,2,3,4,5,6",
					},
				},
				ruleType: "2",
			},
		}, matchData, now, matcher)

		if ruleType != "1" {
			t.Fatalf("matchDispatchRules() ruleType = %s, want 1", ruleType)
		}
		if len(rules) != 1 || rules[0].ID != 3 {
			t.Fatalf("matchDispatchRules() matched rules = %+v, want custom rule", rules)
		}
	})

	t.Run("does not include default rule when custom rule also matches", func(t *testing.T) {
		rules, ruleType := matchDispatchRules([]dispatchRuleSet{
			{
				rules: []Rule{
					{
						ID:         5,
						MType:      "1",
						Conditions: `[{"r_type":"host","r_func":"=","r_value":"server-01"}]`,
						Stime:      "00:00",
						Etime:      "23:59",
						Sweek:      "0,1,2,3,4,5,6",
					},
				},
				ruleType: "1",
			},
			{
				rules: []Rule{
					{
						ID:         6,
						MType:      "2",
						Conditions: "",
						Stime:      "00:00",
						Etime:      "23:59",
						Sweek:      "0,1,2,3,4,5,6",
					},
				},
				ruleType: "2",
			},
		}, matchData, now, matcher)

		if ruleType != "1" {
			t.Fatalf("matchDispatchRules() ruleType = %s, want 1", ruleType)
		}
		if len(rules) != 1 {
			t.Fatalf("matchDispatchRules() matched rule count = %d, want 1", len(rules))
		}
		if rules[0].ID != 5 {
			t.Fatalf("matchDispatchRules() matched rules = %+v, want only custom rule", rules)
		}
	})

	t.Run("falls back to global default when instance rules do not match", func(t *testing.T) {
		rules, ruleType := matchDispatchRules([]dispatchRuleSet{
			{
				rules: []Rule{
					{
						ID:         7,
						MType:      "1",
						Conditions: `[{"r_type":"host","r_func":"=","r_value":"server-02"}]`,
						Stime:      "00:00",
						Etime:      "23:59",
						Sweek:      "0,1,2,3,4,5,6",
					},
				},
				ruleType: "1",
			},
			{
				rules: []Rule{
					{
						ID:         8,
						MType:      "2",
						Conditions: `[{"r_type":"host","r_func":"=","r_value":"server-03"}]`,
						Stime:      "00:00",
						Etime:      "23:59",
						Sweek:      "0,1,2,3,4,5,6",
					},
				},
				ruleType: "2",
			},
			{
				rules: []Rule{
					{
						ID:         9,
						MType:      "2",
						Conditions: "",
						Stime:      "00:00",
						Etime:      "23:59",
						Sweek:      "0,1,2,3,4,5,6",
					},
				},
				ruleType: "2",
			},
		}, matchData, now, matcher)

		if ruleType != "2" {
			t.Fatalf("matchDispatchRules() ruleType = %s, want 2", ruleType)
		}
		if len(rules) != 1 || rules[0].ID != 9 {
			t.Fatalf("matchDispatchRules() matched rules = %+v, want global default rule", rules)
		}
	})
}

func TestIsNoneAlarm(t *testing.T) {
	now := time.Date(2026, 2, 10, 10, 0, 0, 0, time.Local) // 周二 10:00

	tests := []struct {
		name string
		rule Rule
		want bool // true 表示“不在生效时间”，即不告警
	}{
		{
			name: "inside time and week",
			rule: Rule{Stime: "09:00", Etime: "18:00", Sweek: "1,2,3,4,5"},
			want: false,
		},
		{
			name: "outside time",
			rule: Rule{Stime: "11:00", Etime: "18:00", Sweek: "1,2,3,4,5"},
			want: true,
		},
		{
			name: "outside week",
			rule: Rule{Stime: "09:00", Etime: "18:00", Sweek: "1,3,4,5"},
			want: true,
		},
		{
			name: "cross day match",
			rule: Rule{Stime: "22:00", Etime: "11:00", Sweek: "2"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNoneAlarm(now, &tt.rule); got != tt.want {
				t.Errorf("isNoneAlarm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetRuleUserIDs(t *testing.T) {
	groupMap := map[string][]string{
		"1": {"101", "102"},
		"2": {"102", "103"},
	}

	tests := []struct {
		name    string
		uidsStr string
		gidsStr string
		wantLen int
	}{
		{
			name:    "only uids",
			uidsStr: "201,202",
			gidsStr: "",
			wantLen: 2,
		},
		{
			name:    "only gids",
			uidsStr: "",
			gidsStr: "1",
			wantLen: 2,
		},
		{
			name:    "uids and gids with overlap",
			uidsStr: "101,201",
			gidsStr: "1,2", // 101,102 + 102,103 -> 101,102,103
			wantLen: 4,     // 101, 102, 103, 201
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getRuleUserIDs(tt.uidsStr, tt.gidsStr, groupMap)
			if len(got) != tt.wantLen {
				t.Errorf("getRuleUserIDs() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}
