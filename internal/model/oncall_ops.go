package model

import (
	"reflect"
	"time"
)

// OnCallOverview contains roster assignments, not channel-filtered recipients.
// The next personal interval is bounded by LookupUntil and can be ongoing.
type OnCallOverview struct {
	Schedule    OnCallSchedule   `json:"schedule"`
	Current     OnCallResolution `json:"current"`
	Today       []OnCallSlot     `json:"today"`
	MyNext      *OnCallSlot      `json:"my_next"`
	LookupUntil string           `json:"lookup_until"`
	MyRole      string           `json:"my_role"`
	Error       string           `json:"error,omitempty"`
}

func onCallContains(ids []int, id int) bool {
	for _, candidate := range ids {
		if candidate == id {
			return true
		}
	}
	return false
}

// OnCallViewerRole excludes expired cover assignments from the personal view.
func OnCallViewerRole(s OnCallSchedule, overrides []OnCallOverride, userID int, groups map[int][]int, at time.Time) string {
	if onCallContains(s.Members, userID) {
		return "member"
	}
	local := at
	if loc, err := time.LoadLocation(s.Timezone); err == nil {
		local = at.In(loc)
	}
	until := local.AddDate(0, 0, 31)
	for _, o := range overrides {
		if o.ScheduleID != 0 && o.ScheduleID != s.ID {
			continue
		}
		start, e1 := time.Parse(time.RFC3339, o.Start)
		end, e2 := time.Parse(time.RFC3339, o.End)
		if e1 == nil && e2 == nil && end.After(at) && start.Before(until) && onCallContains(o.UserIDs, userID) {
			return "override"
		}
	}
	if onCallContains(s.FallbackUserIDs, userID) {
		return "fallback"
	}
	for _, id := range s.FallbackGroupIDs {
		if onCallContains(groups[id], userID) {
			return "fallback"
		}
	}
	return ""
}

// Civil boundaries are located by date, rather than assuming midnight exists.
// This also handles timezones whose DST jump skips 00:00.
func onCallDayBoundary(at time.Time, target int64) time.Time {
	low, high := at.Add(-48*time.Hour).Unix(), at.Add(48*time.Hour).Unix()
	for high-low > 1 {
		mid := low + (high-low)/2
		if civilDay(time.Unix(mid, 0).In(at.Location())) < target {
			low = mid
		} else {
			high = mid
		}
	}
	return time.Unix(high, 0).In(at.Location())
}

func BuildOnCallOverview(s OnCallSchedule, overrides []OnCallOverride, userID int, groups map[int][]int, at time.Time) OnCallOverview {
	row := OnCallOverview{Schedule: s, Current: OnCallResolution{UserIDs: []int{}, Source: "error"}, Today: []OnCallSlot{}, MyRole: OnCallViewerRole(s, overrides, userID, groups, at)}
	loc, err := time.LoadLocation(s.Timezone)
	if err != nil {
		row.Error = "排班时区无效"
		return row
	}
	local := at.In(loc)
	row.LookupUntil = local.AddDate(0, 0, 31).Format(time.RFC3339Nano)
	row.Current, err = ResolveOnCall(s, overrides, at)
	if err != nil {
		row.Error = err.Error()
		return row
	}
	first := onCallDayBoundary(local, civilDay(local))
	last := onCallDayBoundary(local, civilDay(local)+1)
	slots, err := PreviewOnCallSlots(s, overrides, first, 2)
	if err != nil {
		row.Error = err.Error()
		return row
	}
	for _, slot := range slots {
		start, _ := time.Parse(time.RFC3339, slot.Start)
		end, _ := time.Parse(time.RFC3339, slot.End)
		if !start.Before(last) {
			continue
		}
		if end.After(last) {
			slot.End = last.Format(time.RFC3339Nano)
		}
		row.Today = append(row.Today, slot)
	}
	future, err := PreviewOnCallSlots(s, overrides, at, 31)
	if err != nil {
		row.Error = err.Error()
		return row
	}
	for _, slot := range future {
		if row.MyNext == nil {
			if onCallContains(slot.UserIDs, userID) {
				copy := slot
				row.MyNext = &copy
			}
			continue
		}
		if row.MyNext.End == slot.Start && row.MyNext.Source == slot.Source && row.MyNext.Reason == slot.Reason && reflect.DeepEqual(row.MyNext.UserIDs, slot.UserIDs) {
			row.MyNext.End = slot.End
		} else {
			break
		}
	}
	return row
}
