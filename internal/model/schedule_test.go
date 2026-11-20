package model

import (
	"reflect"
	"testing"
	"time"
)

func testSchedule() OnCallSchedule {
	return OnCallSchedule{ID: 1, Name: "生产运维", Timezone: "Asia/Shanghai", Rotation: "daily", Start: "2026-11-20T09:00:00+08:00", Members: []int{11, 22, 33}, Weekdays: []int{0, 1, 2, 3, 4, 5, 6}, WindowStart: "00:00", WindowEnd: "00:00", Enabled: true}
}
func mustScheduleTime(t *testing.T, s string) time.Time {
	t.Helper()
	at, err := time.Parse(time.RFC3339, s)
	if err != nil {
		t.Fatal(err)
	}
	return at
}
func TestOnCallRotations(t *testing.T) {
	s := testSchedule()
	cases := []struct {
		name, rotation, at string
		want               []int
		source             string
	}{
		{"before anchor", "daily", "2026-11-20T08:59:59+08:00", []int{}, "gap"},
		{"anchor inclusive", "daily", "2026-11-20T09:00:00+08:00", []int{11}, "rotation"},
		{"before handoff", "daily", "2026-11-21T08:59:59+08:00", []int{11}, "rotation"},
		{"handoff", "daily", "2026-11-21T09:00:00+08:00", []int{22}, "rotation"},
		{"wrap round", "daily", "2026-11-23T09:00:00+08:00", []int{11}, "rotation"},
		{"weekly before", "weekly", "2026-11-27T08:59:59+08:00", []int{11}, "rotation"},
		{"weekly boundary", "weekly", "2026-11-27T09:00:00+08:00", []int{22}, "rotation"},
		{"fixed parallel", "fixed", "2026-12-27T09:00:00+08:00", []int{11, 22, 33}, "rotation"},
		{"same instant timezone", "daily", "2026-11-21T01:00:00Z", []int{22}, "rotation"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s.Rotation = tt.rotation
			r, e := ResolveOnCall(s, nil, mustScheduleTime(t, tt.at))
			if e != nil {
				t.Fatal(e)
			}
			if !reflect.DeepEqual(r.UserIDs, tt.want) || r.Source != tt.source {
				t.Fatalf("got %+v want %v %s", r, tt.want, tt.source)
			}
		})
	}
}
func TestOnCallCoverageAndOverrides(t *testing.T) {
	s := testSchedule()
	s.Start = "2026-11-01T09:00:00+08:00"
	s.Rotation = "fixed"
	s.Weekdays = []int{5}
	s.WindowStart = "22:00"
	s.WindowEnd = "06:00"
	for _, tt := range []struct {
		at      string
		covered bool
	}{{"2026-11-20T21:59:59+08:00", false}, {"2026-11-20T22:00:00+08:00", true}, {"2026-11-21T05:59:59+08:00", true}, {"2026-11-21T06:00:00+08:00", false}, {"2026-11-21T22:00:00+08:00", false}} {
		r, e := ResolveOnCall(s, nil, mustScheduleTime(t, tt.at))
		if e != nil {
			t.Fatal(e)
		}
		if (len(r.UserIDs) > 0) != tt.covered {
			t.Errorf("%s %+v", tt.at, r)
		}
	}
	o := OnCallOverride{ScheduleID: 1, Start: "2026-11-21T12:00:00+08:00", End: "2026-11-21T12:00:30+08:00", UserIDs: []int{44, 55}, Reason: "临时维护"}
	for _, tt := range []struct {
		at     string
		source string
	}{{"2026-11-21T11:59:59+08:00", "gap"}, {"2026-11-21T12:00:00+08:00", "override"}, {"2026-11-21T12:00:29+08:00", "override"}, {"2026-11-21T12:00:30+08:00", "gap"}} {
		r, e := ResolveOnCall(s, []OnCallOverride{o}, mustScheduleTime(t, tt.at))
		if e != nil {
			t.Fatal(e)
		}
		if r.Source != tt.source {
			t.Fatalf("%s %+v", tt.at, r)
		}
	}
	if _, e := ResolveOnCall(s, []OnCallOverride{o, o}, mustScheduleTime(t, o.Start)); e == nil {
		t.Fatal("overlap accepted")
	}
	s.Enabled = false
	r, e := ResolveOnCall(s, []OnCallOverride{o}, mustScheduleTime(t, o.Start))
	if e != nil || r.Source != "disabled" || len(r.UserIDs) != 0 {
		t.Fatalf("disabled %+v %v", r, e)
	}
}
func TestOnCallValidation(t *testing.T) {
	for _, mutate := range []func(*OnCallSchedule){
		func(s *OnCallSchedule) { s.Name = "" }, func(s *OnCallSchedule) { s.Timezone = "bad/zone" }, func(s *OnCallSchedule) { s.Timezone = "Local" }, func(s *OnCallSchedule) { s.Rotation = "hourly" }, func(s *OnCallSchedule) { s.Start = "2026-11-20 09:00" }, func(s *OnCallSchedule) { s.Members = []int{1, 1} }, func(s *OnCallSchedule) { s.Members = nil }, func(s *OnCallSchedule) { s.Weekdays = []int{7} }, func(s *OnCallSchedule) { s.Weekdays = []int{1, 1} }, func(s *OnCallSchedule) { s.WindowStart = "24:00" }, func(s *OnCallSchedule) { s.FallbackUserIDs = []int{0} }, func(s *OnCallSchedule) { s.FallbackGroupIDs = []int{1, 1} },
	} {
		s := testSchedule()
		mutate(&s)
		if ValidateOnCallSchedule(s) == nil {
			t.Fatalf("accepted %+v", s)
		}
	}
	for _, o := range []OnCallOverride{{Start: "bad"}, {Start: "2026-11-21T09:00:00Z", End: "2026-11-21T08:00:00Z"}, {Start: "2026-11-21T09:00:00Z", End: "2026-11-21T10:00:00Z", UserIDs: []int{1}, Reason: ""}} {
		if ValidateOnCallOverride(o) == nil {
			t.Fatalf("accepted %+v", o)
		}
	}
}
func TestOnCallDSTAndPreview(t *testing.T) {
	s := testSchedule()
	s.Timezone = "America/New_York"
	s.Start = "2026-03-07T09:00:00-05:00"
	for _, tt := range []struct {
		at   string
		want int
	}{{"2026-03-08T08:59:59-04:00", 11}, {"2026-03-08T09:00:00-04:00", 22}, {"2026-03-09T09:00:00-04:00", 33}} {
		r, e := ResolveOnCall(s, nil, mustScheduleTime(t, tt.at))
		if e != nil || len(r.UserIDs) != 1 || r.UserIDs[0] != tt.want {
			t.Fatalf("DST %s %+v %v", tt.at, r, e)
		}
	}
	s.Start = "2026-10-31T09:00:00-04:00"
	r, e := ResolveOnCall(s, nil, mustScheduleTime(t, "2026-11-01T09:00:00-05:00"))
	if e != nil || r.UserIDs[0] != 22 {
		t.Fatalf("fall DST %+v %v", r, e)
	}
	s = testSchedule()
	from := mustScheduleTime(t, "2026-11-21T10:00:00+08:00")
	o := OnCallOverride{ScheduleID: 1, Start: "2026-11-21T10:01:00+08:00", End: "2026-11-21T10:01:20+08:00", UserIDs: []int{44}, Reason: "20秒代班"}
	slots, e := PreviewOnCallSlots(s, []OnCallOverride{o}, from, 7)
	if e != nil {
		t.Fatal(e)
	}
	found := false
	for i, slot := range slots {
		if i > 0 && slot.Start != slots[i-1].End {
			t.Fatal("preview has gap")
		}
		a := mustScheduleTime(t, slot.Start)
		b := mustScheduleTime(t, slot.End)
		if !b.After(a) {
			t.Fatal("empty slot")
		}
		r, e := ResolveOnCall(s, []OnCallOverride{o}, a)
		if e != nil || !reflect.DeepEqual(r.UserIDs, slot.UserIDs) || r.Source != slot.Source {
			t.Fatal("preview differs from resolver")
		}
		if slot.Source == "override" {
			found = true
			if b.Sub(a) != 20*time.Second {
				t.Fatal("short override missed")
			}
		}
	}
	if !found {
		t.Fatal("override not previewed")
	}
	for _, days := range []int{0, 32} {
		if _, e := PreviewOnCallSlots(s, nil, from, days); e == nil {
			t.Fatal("invalid preview days")
		}
	}
}

func TestOnCallPreviewDSTEveryInterval(t *testing.T) {
	cases := []struct{ name, start, from, ws, we, rotation string }{
		{"fall repeated coverage clock", "2026-10-30T00:00:00-04:00", "2026-11-01T00:00:00-04:00", "01:30", "02:30", "fixed"},
		{"fall narrow repeated coverage window", "2026-10-30T00:00:00-04:00", "2026-11-01T00:00:00-04:00", "01:10", "01:40", "fixed"},
		{"fall ambiguous daily handoff", "2026-10-30T01:30:00-04:00", "2026-11-01T00:00:00-04:00", "00:00", "00:00", "daily"},
		{"spring nonexistent daily handoff", "2026-03-06T02:30:00-05:00", "2026-03-08T00:00:00-05:00", "00:00", "00:00", "daily"},
		{"spring nonexistent window start", "2026-03-06T00:00:00-05:00", "2026-03-08T00:00:00-05:00", "02:30", "03:30", "fixed"},
		{"spring window ends in gap", "2026-03-06T00:00:00-05:00", "2026-03-08T00:00:00-05:00", "01:30", "02:30", "fixed"},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			s := testSchedule()
			s.Timezone = "America/New_York"
			s.Start = tt.start
			s.WindowStart = tt.ws
			s.WindowEnd = tt.we
			s.Rotation = tt.rotation
			from := mustScheduleTime(t, tt.from)
			slots, err := PreviewOnCallSlots(s, nil, from, 1)
			if err != nil {
				t.Fatal(err)
			}
			// Each interval must be constant, including its final nanosecond. Sampling
			// additional minute instants guards against a missed transition inside it.
			for _, slot := range slots {
				a := mustScheduleTime(t, slot.Start)
				b := mustScheduleTime(t, slot.End)
				checks := []time.Time{a, b.Add(-time.Nanosecond)}
				for at := a.Add(time.Minute); at.Before(b); at = at.Add(time.Minute) {
					checks = append(checks, at)
				}
				for _, at := range checks {
					r, err := ResolveOnCall(s, nil, at)
					if err != nil {
						t.Fatal(err)
					}
					if r.Source != slot.Source || !reflect.DeepEqual(r.UserIDs, slot.UserIDs) {
						t.Fatalf("nonconstant interval %s -> %s: at %s got %+v want %+v", slot.Start, slot.End, at.Format(time.RFC3339Nano), r, slot)
					}
				}
			}
			if tt.name == "fall repeated coverage clock" {
				gotGap := false
				gotSecondStart := false
				for _, slot := range slots {
					if slot.Start == "2026-11-01T01:00:00-05:00" && slot.Source == "gap" {
						gotGap = true
					}
					if slot.Start == "2026-11-01T01:30:00-05:00" && slot.Source == "rotation" {
						gotSecondStart = true
					}
				}
				if !gotGap || !gotSecondStart {
					t.Fatalf("missing repeated-hour boundaries: %+v", slots)
				}
			}
		})
	}
}

func TestOnCallPreviewFractionalBoundaries(t *testing.T) {
	s := testSchedule()
	s.Start = "2026-11-20T09:00:00.5+08:00"
	slots, err := PreviewOnCallSlots(s, nil, mustScheduleTime(t, "2026-11-21T08:59:59+08:00"), 1)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, slot := range slots {
		if slot.Start == "2026-11-21T09:00:00.5+08:00" {
			found = true
			if !reflect.DeepEqual(slot.UserIDs, []int{22}) {
				t.Fatalf("wrong handoff %+v", slot)
			}
		}
	}
	if !found {
		t.Fatal("fractional handoff lost precision")
	}
}
