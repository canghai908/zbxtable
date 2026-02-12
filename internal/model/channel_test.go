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
