package model

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// OnCallSchedule is reusable by dispatch rules. Rotation follows calendar days
// in Timezone, rather than elapsed 24-hour periods, including across DST.
type OnCallSchedule struct {
	ID               int       `gorm:"primaryKey;autoIncrement" json:"id"`
	Name             string    `gorm:"size:200;uniqueIndex" json:"name"`
	Timezone         string    `gorm:"size:100" json:"timezone"`
	Rotation         string    `gorm:"size:20" json:"rotation"`
	Start            string    `gorm:"size:64" json:"start"`
	Members          []int     `gorm:"serializer:json;type:text" json:"members"`
	Weekdays         []int     `gorm:"serializer:json;type:text" json:"weekdays"`
	WindowStart      string    `gorm:"size:5" json:"window_start"`
	WindowEnd        string    `gorm:"size:5" json:"window_end"`
	FallbackUserIDs  []int     `gorm:"serializer:json;type:text" json:"fallback_user_ids"`
	FallbackGroupIDs []int     `gorm:"serializer:json;type:text" json:"fallback_group_ids"`
	Enabled          bool      `json:"enabled"`
	Created          time.Time `gorm:"autoCreateTime" json:"created"`
	Updated          time.Time `gorm:"autoUpdateTime" json:"updated"`
}

func (*OnCallSchedule) TableName() string { return TableName("oncall_schedule") }

type OnCallOverride struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	ScheduleID int       `gorm:"index" json:"schedule_id"`
	Start      string    `gorm:"size:64" json:"start"`
	End        string    `gorm:"size:64" json:"end"`
	UserIDs    []int     `gorm:"serializer:json;type:text" json:"user_ids"`
	Reason     string    `gorm:"size:1000" json:"reason"`
	CreatedBy  string    `gorm:"size:255" json:"created_by"`
	Created    time.Time `gorm:"autoCreateTime" json:"created"`
}

func (*OnCallOverride) TableName() string { return TableName("oncall_override") }

type OnCallAudit struct {
	ID         int       `gorm:"primaryKey;autoIncrement" json:"id"`
	ScheduleID int       `gorm:"index" json:"schedule_id"`
	Actor      string    `gorm:"size:255" json:"actor"`
	Action     string    `gorm:"size:40" json:"action"`
	Detail     string    `gorm:"type:text" json:"detail"`
	Created    time.Time `gorm:"autoCreateTime" json:"created"`
}

func (*OnCallAudit) TableName() string { return TableName("oncall_audit") }

type OnCallResolution struct {
	UserIDs     []int  `json:"user_ids"`
	Source      string `json:"source"`
	Reason      string `json:"reason"`
	NextHandoff string `json:"next_handoff,omitempty"`
}
type OnCallSlot struct {
	Start   string `json:"start"`
	End     string `json:"end"`
	UserIDs []int  `json:"user_ids"`
	Source  string `json:"source"`
	Reason  string `json:"reason"`
}

func validatePositiveIDs(ids []int, label string, required bool) error {
	if required && len(ids) == 0 {
		return fmt.Errorf("%s不能为空", label)
	}
	if len(ids) > 200 {
		return fmt.Errorf("%s最多支持200项", label)
	}
	seen := map[int]bool{}
	for _, id := range ids {
		if id <= 0 || seen[id] {
			return fmt.Errorf("%s包含无效或重复ID", label)
		}
		seen[id] = true
	}
	return nil
}
func scheduleClock(raw string) (int, error) {
	if len(raw) != 5 {
		return 0, errors.New("时间必须为 HH:mm")
	}
	t, err := time.Parse("15:04", raw)
	if err != nil {
		return 0, errors.New("时间必须为 HH:mm")
	}
	return t.Hour()*60 + t.Minute(), nil
}
func ValidateOnCallSchedule(s OnCallSchedule) error {
	if strings.TrimSpace(s.Name) == "" || len(s.Name) > 200 {
		return errors.New("排班名称不能为空且最长200字节")
	}
	if _, err := time.LoadLocation(s.Timezone); err != nil || s.Timezone == "" || s.Timezone == "Local" {
		return errors.New("请使用有效IANA时区，例如 Asia/Shanghai")
	}
	if s.Rotation != "fixed" && s.Rotation != "daily" && s.Rotation != "weekly" {
		return errors.New("轮班方式必须为 fixed/daily/weekly")
	}
	if _, err := time.Parse(time.RFC3339, s.Start); err != nil {
		return errors.New("开始时间必须包含时区，使用 RFC3339")
	}
	if err := validatePositiveIDs(s.Members, "值班成员", true); err != nil {
		return err
	}
	if len(s.Weekdays) == 0 || len(s.Weekdays) > 7 {
		return errors.New("至少选择一个值班星期")
	}
	seen := map[int]bool{}
	for _, d := range s.Weekdays {
		if d < 0 || d > 6 || seen[d] {
			return errors.New("值班星期无效或重复")
		}
		seen[d] = true
	}
	if _, err := scheduleClock(s.WindowStart); err != nil {
		return err
	}
	if _, err := scheduleClock(s.WindowEnd); err != nil {
		return err
	}
	if err := validatePositiveIDs(s.FallbackUserIDs, "兜底用户", false); err != nil {
		return err
	}
	return validatePositiveIDs(s.FallbackGroupIDs, "兜底用户组", false)
}
func ValidateOnCallOverride(o OnCallOverride) error {
	a, e := time.Parse(time.RFC3339, o.Start)
	if e != nil {
		return errors.New("代班开始时间必须为 RFC3339")
	}
	b, e := time.Parse(time.RFC3339, o.End)
	if e != nil || !b.After(a) {
		return errors.New("代班结束时间必须晚于开始时间")
	}
	if strings.TrimSpace(o.Reason) == "" || len(o.Reason) > 1000 {
		return errors.New("代班原因不能为空且最长1000字节")
	}
	return validatePositiveIDs(o.UserIDs, "代班用户", true)
}
func civilDay(t time.Time) int64 {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC).Unix() / 86400
}
func clockOnDay(t time.Time, minutes int) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, minutes/60, minutes%60, 0, 0, t.Location())
}

func ResolveOnCall(s OnCallSchedule, overrides []OnCallOverride, at time.Time) (OnCallResolution, error) {
	result := OnCallResolution{UserIDs: []int{}, Source: "gap"}
	if err := ValidateOnCallSchedule(s); err != nil {
		return result, err
	}
	if !s.Enabled {
		result.Source = "disabled"
		result.Reason = "排班已禁用"
		return result, nil
	}
	anchor, _ := time.Parse(time.RFC3339, s.Start)
	loc, _ := time.LoadLocation(s.Timezone)
	local := at.In(loc)
	anchor = anchor.In(loc)
	if at.Before(anchor) {
		result.Reason = "排班尚未开始"
		return result, nil
	}
	// A temporary assignment replaces normal coverage, including an uncovered
	// weekday/window. Overlapping assignments are rejected on write and on read.
	active := []OnCallOverride{}
	for _, o := range overrides {
		if o.ScheduleID != 0 && o.ScheduleID != s.ID {
			continue
		}
		if err := ValidateOnCallOverride(o); err != nil {
			return result, err
		}
		a, _ := time.Parse(time.RFC3339, o.Start)
		b, _ := time.Parse(time.RFC3339, o.End)
		if !at.Before(a) && at.Before(b) {
			active = append(active, o)
		}
	}
	if len(active) > 1 {
		return result, errors.New("代班时间重叠，请修复排班")
	}
	if len(active) == 1 {
		o := active[0]
		return OnCallResolution{UserIDs: append([]int{}, o.UserIDs...), Source: "override", Reason: o.Reason, NextHandoff: o.End}, nil
	}
	ws, _ := scheduleClock(s.WindowStart)
	we, _ := scheduleClock(s.WindowEnd)
	minute := local.Hour()*60 + local.Minute()
	coverageDay := local
	covered := true
	if ws < we {
		covered = minute >= ws && minute < we
	} else if ws > we {
		covered = minute >= ws || minute < we
		if minute < we {
			coverageDay = local.AddDate(0, 0, -1)
		}
	}
	if !covered {
		result.Reason = "当前时间不在值班时段"
		return result, nil
	}
	dayOK := false
	for _, d := range s.Weekdays {
		if d == int(coverageDay.Weekday()) {
			dayOK = true
		}
	}
	if !dayOK {
		result.Reason = "当前星期不在值班范围"
		return result, nil
	}
	idx := 0
	if s.Rotation != "fixed" {
		shiftDay := local
		boundary := clockOnDay(local, anchor.Hour()*60+anchor.Minute())
		// Preserve anchor seconds as well as minutes.
		boundary = boundary.Add(time.Duration(anchor.Second())*time.Second + time.Duration(anchor.Nanosecond()))
		if local.Before(boundary) {
			shiftDay = local.AddDate(0, 0, -1)
		}
		days := civilDay(shiftDay) - civilDay(anchor)
		if days < 0 {
			result.Reason = "排班尚未开始"
			return result, nil
		}
		period := int64(1)
		if s.Rotation == "weekly" {
			period = 7
		}
		idx = int((days / period) % int64(len(s.Members)))
		nextDays := period - days%period
		next := clockOnDay(shiftDay.AddDate(0, 0, int(nextDays)), anchor.Hour()*60+anchor.Minute()).Add(time.Duration(anchor.Second())*time.Second + time.Duration(anchor.Nanosecond()))
		result.NextHandoff = next.Format(time.RFC3339Nano)
	}
	result.Source = "rotation"
	result.Reason = "按排班轮换"
	result.UserIDs = []int{s.Members[idx]}
	if s.Rotation == "fixed" {
		result.UserIDs = append([]int{}, s.Members...)
	}
	return result, nil
}

// PreviewOnCallSlots splits at exact civil-time, coverage and override
// boundaries. No minute sampling is used, so short overrides are visible.
func PreviewOnCallSlots(s OnCallSchedule, overrides []OnCallOverride, from time.Time, days int) ([]OnCallSlot, error) {
	if days < 1 || days > 31 {
		return nil, errors.New("预览天数必须为1至31天")
	}
	if err := ValidateOnCallSchedule(s); err != nil {
		return nil, err
	}
	loc, _ := time.LoadLocation(s.Timezone)
	from = from.In(loc)
	end := from.AddDate(0, 0, days)
	bounds := []time.Time{from, end}
	add := func(t time.Time) {
		if t.After(from) && t.Before(end) {
			bounds = append(bounds, t)
		}
	}
	anchor, _ := time.Parse(time.RFC3339, s.Start)
	anchor = anchor.In(loc)
	add(anchor)
	ws, _ := scheduleClock(s.WindowStart)
	we, _ := scheduleClock(s.WindowEnd)
	// Wall-clock windows can change at a timezone transition even if no
	// configured clock is crossed (01:45 -> 01:00 during a rollback). A
	// repeated clock also has two distinct instants, both of which must split
	// the preview. Discover the offsets on each exact timezone segment.
	offsets := map[int]bool{}
	zoneUntil := end.AddDate(0, 0, 2)
	for cursor := from.AddDate(0, 0, -2); cursor.Before(zoneUntil); {
		_, offset := cursor.Zone()
		offsets[offset] = true
		_, next := cursor.ZoneBounds()
		if next.IsZero() || !next.Before(zoneUntil) {
			break
		}
		add(next)
		if !next.After(cursor) {
			break
		}
		cursor = next.In(loc)
	}
	addClock := func(day time.Time, minute, second, nano int) {
		// Keep Go's normalized time for a nonexistent clock, since the
		// resolver uses this boundary for its calendar-day rotation.
		add(clockOnDay(day, minute).Add(time.Duration(second)*time.Second + time.Duration(nano)))
		y, m, d := day.Date()
		wall := time.Date(y, m, d, minute/60, minute%60, second, nano, time.UTC)
		for offset := range offsets {
			candidate := wall.Add(-time.Duration(offset) * time.Second).In(loc)
			cy, cm, cd := candidate.Date()
			if cy == y && cm == m && cd == d && candidate.Hour() == minute/60 && candidate.Minute() == minute%60 && candidate.Second() == second && candidate.Nanosecond() == nano {
				add(candidate)
			}
		}
	}
	for d := -1; d <= days+1; d++ {
		day := from.AddDate(0, 0, d)
		addClock(day, 0, 0, 0)
		addClock(day, ws, 0, 0)
		addClock(day, we, 0, 0)
		addClock(day, anchor.Hour()*60+anchor.Minute(), anchor.Second(), anchor.Nanosecond())
	}
	for _, o := range overrides {
		a, e := time.Parse(time.RFC3339, o.Start)
		if e != nil {
			return nil, e
		}
		b, e := time.Parse(time.RFC3339, o.End)
		if e != nil {
			return nil, e
		}
		add(a)
		add(b)
	}
	sort.Slice(bounds, func(i, j int) bool { return bounds[i].Before(bounds[j]) })
	slots := []OnCallSlot{}
	for i := 0; i < len(bounds)-1; i++ {
		if bounds[i].Equal(bounds[i+1]) {
			continue
		}
		r, e := ResolveOnCall(s, overrides, bounds[i])
		if e != nil {
			return nil, e
		}
		slot := OnCallSlot{Start: bounds[i].Format(time.RFC3339Nano), End: bounds[i+1].Format(time.RFC3339Nano), UserIDs: r.UserIDs, Source: r.Source, Reason: r.Reason}
		slots = append(slots, slot)
	}
	return slots, nil
}
