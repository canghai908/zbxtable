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
