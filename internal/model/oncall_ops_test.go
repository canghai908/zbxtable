package model

import (
	"reflect"
	"testing"
	"time"
)

func TestOnCallOverviewDailyAndPersonalShift(t *testing.T) {
	s := testSchedule()
	s.ID = 1
	at := mustScheduleTime(t, "2026-11-21T10:00:00+08:00")
	row := BuildOnCallOverview(s, nil, 33, nil, at)
	if row.Error != "" || !reflect.DeepEqual(row.Current.UserIDs, []int{22}) || row.MyRole != "member" {
		t.Fatalf("current: %+v", row)
	}
	if row.MyNext == nil || row.MyNext.Start != "2026-11-22T09:00:00+08:00" || row.MyNext.End != "2026-11-23T09:00:00+08:00" {
		t.Fatalf("next shift split at midnight: %+v", row.MyNext)
	}
	first, _ := time.Parse(time.RFC3339, row.Today[0].Start)
	last, _ := time.Parse(time.RFC3339, row.Today[len(row.Today)-1].End)
	if first.Format(time.RFC3339) != "2026-11-21T00:00:00+08:00" || last.Format(time.RFC3339) != "2026-11-22T00:00:00+08:00" {
		t.Fatalf("today bounds %s %s", first, last)
	}
	ongoing := BuildOnCallOverview(s, nil, 22, nil, at)
	if ongoing.MyNext == nil || ongoing.MyNext.Start != at.Format(time.RFC3339Nano) || ongoing.MyNext.End != "2026-11-22T09:00:00+08:00" {
		t.Fatalf("ongoing: %+v", ongoing.MyNext)
	}
}

func TestOnCallOverviewCivilDayLengths(t *testing.T) {
	for _, tt := range []struct {
		name, zone, at string
		hours          int
	}{
		{"spring", "America/New_York", "2026-03-08T12:00:00-04:00", 23},
		{"fall", "America/New_York", "2026-11-01T12:00:00-05:00", 25},
		{"midnight jump", "America/Sao_Paulo", "2018-11-04T12:00:00-02:00", 23},
	} {
		t.Run(tt.name, func(t *testing.T) {
			at := mustScheduleTime(t, tt.at)
			s := testSchedule()
			s.Timezone = tt.zone
			s.Rotation = "fixed"
			s.Start = at.AddDate(0, 0, -2).Format(time.RFC3339)
			row := BuildOnCallOverview(s, nil, 11, nil, at)
			if row.Error != "" || len(row.Today) == 0 {
				t.Fatalf("today %+v", row)
			}
			first, _ := time.Parse(time.RFC3339, row.Today[0].Start)
			last, _ := time.Parse(time.RFC3339, row.Today[len(row.Today)-1].End)
			if last.Sub(first) != time.Duration(tt.hours)*time.Hour {
				t.Fatalf("today duration %s want %dh", last.Sub(first), tt.hours)
			}
		})
	}
}

func TestOnCallOverviewOverridesGapsAndInvalidIsolation(t *testing.T) {
	s := testSchedule()
	s.ID = 1
	at := mustScheduleTime(t, "2026-11-20T10:00:00+08:00")
	override := OnCallOverride{ScheduleID: 1, Start: "2026-11-20T11:00:00+08:00", End: "2026-11-20T12:00:00+08:00", UserIDs: []int{99}, Reason: "cover"}
	row := BuildOnCallOverview(s, []OnCallOverride{override}, 99, nil, at)
	if row.MyRole != "override" || row.MyNext == nil || row.MyNext.Source != "override" {
		t.Fatalf("override: %+v", row)
	}
	s.WindowStart = "22:00"
	s.WindowEnd = "06:00"
	row = BuildOnCallOverview(s, nil, 11, nil, at)
	if row.Current.Source != "gap" || row.MyNext == nil || row.MyNext.Start != "2026-11-20T22:00:00+08:00" {
		t.Fatalf("overnight: %+v", row)
	}
	s.Enabled = false
	row = BuildOnCallOverview(s, nil, 11, nil, at)
	if row.Current.Source != "disabled" || row.MyNext != nil {
		t.Fatalf("disabled: %+v", row)
	}
	s.Timezone = "bad/timezone"
	if bad := BuildOnCallOverview(s, nil, 11, nil, at); bad.Error == "" {
		t.Fatal("invalid row not marked")
	}
}

func TestOnCallMembershipUsesUpcomingCoverAndFallback(t *testing.T) {
	s := testSchedule()
	s.ID = 1
	s.FallbackUserIDs = []int{55}
	s.FallbackGroupIDs = []int{8}
	at := mustScheduleTime(t, "2026-11-20T10:00:00+08:00")
	overrides := []OnCallOverride{{ScheduleID: 1, Start: "2026-11-19T09:00:00+08:00", End: "2026-11-20T10:00:00+08:00", UserIDs: []int{77}, Reason: "ended"}}
	for id, want := range map[int]string{11: "member", 55: "fallback", 66: "fallback", 77: "", 88: ""} {
		if role := OnCallViewerRole(s, overrides, id, map[int][]int{8: {66}}, at); role != want {
			t.Fatalf("user=%d role=%s want=%s", id, role, want)
		}
	}
}

func TestOnCallPersonalHorizonUsesScheduleCivilDaysAcrossDST(t *testing.T) {
	s := testSchedule()
	s.ID = 1
	s.Timezone = "America/New_York"
	at := mustScheduleTime(t, "2026-10-15T12:00:00Z")
	// The final local day contains one extra hour after the autumn transition.
	start := at.Add(31*24*time.Hour + 30*time.Minute)
	override := OnCallOverride{ScheduleID: 1, Start: start.Format(time.RFC3339), End: start.Add(10 * time.Minute).Format(time.RFC3339), UserIDs: []int{99}, Reason: "late window cover"}
	if role := OnCallViewerRole(s, []OnCallOverride{override}, 99, nil, at); role != "override" {
		t.Fatalf("lost personal cover in local-day horizon: %s", role)
	}
}

func TestOnCallMidnightDailyHandoff(t *testing.T) {
	for _, zone := range []string{"Asia/Shanghai", "UTC", "America/New_York"} {
		t.Run(zone, func(t *testing.T) {
			loc, err := time.LoadLocation(zone)
			if err != nil {
				t.Fatal(err)
			}
			s := testSchedule()
			s.Timezone = zone
			anchor := time.Date(2026, 11, 1, 0, 0, 0, 0, loc)
			s.Start = anchor.Format(time.RFC3339)
			for day := 0; day < 7; day++ {
				start := anchor.AddDate(0, 0, day)
				end := anchor.AddDate(0, 0, day+1)
				want := []int{s.Members[day%len(s.Members)]}
				for _, at := range []time.Time{start, start.Add(time.Nanosecond), end.Add(-time.Nanosecond)} {
					got, err := ResolveOnCall(s, nil, at)
					if err != nil || !reflect.DeepEqual(got.UserIDs, want) {
						t.Fatalf("%s: got %+v err %v, want %v", at, got, err, want)
					}
				}
				row := BuildOnCallOverview(s, nil, want[0], nil, start.Add(12*time.Hour))
				wantSlots := 1
				// A DST transition is displayed as an explicit timezone boundary,
				// but both parts must retain the same person without a gap.
				if zone == "America/New_York" && day == 0 {
					wantSlots = 2
				}
				if row.Error != "" || len(row.Today) != wantSlots {
					t.Fatalf("day %d should have one complete shift: %+v", day, row)
				}
				if row.Today[0].Start != start.Format(time.RFC3339Nano) || row.Today[len(row.Today)-1].End != end.Format(time.RFC3339Nano) {
					t.Fatalf("day %d has partial coverage: %+v", day, row.Today)
				}
				for i, slot := range row.Today {
					if !reflect.DeepEqual(slot.UserIDs, want) || (i > 0 && row.Today[i-1].End != slot.Start) {
						t.Fatalf("day %d has a gap or wrong member: %+v", day, row.Today)
					}
				}
			}
			slots, err := PreviewOnCallSlots(s, nil, anchor, 7)
			wantSlots := 7
			if zone == "America/New_York" {
				wantSlots = 8
			}
			if err != nil || len(slots) != wantSlots {
				t.Fatalf("seven-day preview: %+v, err %v", slots, err)
			}
			for i, slot := range slots {
				at := mustScheduleTime(t, slot.Start).In(loc)
				day := int(civilDay(at) - civilDay(anchor))
				if !reflect.DeepEqual(slot.UserIDs, []int{s.Members[day%len(s.Members)]}) || (i > 0 && slots[i-1].End != slot.Start) {
					t.Fatalf("preview gap, overlap or incorrect member at %d: %+v", i, slots)
				}
			}
		})
	}
}
