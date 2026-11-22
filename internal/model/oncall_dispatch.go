package model

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// OnCallDispatchSnapshot fixes recipients for a Zabbix event across retries and
// handovers. It contains IDs/explanations only, never notification credentials.
type OnCallDispatchSnapshot struct {
	ID        int                `gorm:"primaryKey" json:"id"`
	ZID       int                `gorm:"column:zid;uniqueIndex:oncall_event_channel" json:"zid"`
	EventID   int64              `gorm:"uniqueIndex:oncall_event_channel" json:"eventid"`
	Channel   string             `gorm:"size:40;uniqueIndex:oncall_event_channel" json:"channel"`
	UserIDs   []int              `gorm:"serializer:json;type:text" json:"user_ids"`
	Decisions []ScheduleDecision `gorm:"serializer:json;type:text" json:"schedules"`
	Created   time.Time          `gorm:"autoCreateTime" json:"created"`
}

func (*OnCallDispatchSnapshot) TableName() string { return TableName("oncall_dispatch_snapshot") }

type ScheduleDecision struct {
	ScheduleID int    `json:"schedule_id"`
	UserIDs    []int  `json:"user_ids"`
	Source     string `json:"source"`
	Reason     string `json:"reason"`
}
type DispatchChannel struct {
	Channel   string             `json:"channel"`
	UserIDs   []int              `json:"user_ids"`
	Schedules []ScheduleDecision `json:"schedules"`
	Muted     bool               `json:"muted"`
}

func serializedScheduleIDs(ids []int) string {
	if ids == nil {
		ids = []int{}
	}
	b, _ := json.Marshal(ids)
	return string(b)
}

func onCallUserDeliverable(u User, channel string) bool {
	if u.Status != 0 {
		return false
	}
	switch normalizeChannel(channel) {
	case "mail":
		return strings.TrimSpace(u.Email) != ""
	case "wechat":
		return strings.TrimSpace(u.Wechat) != ""
	case "wechat_robot":
		return strings.TrimSpace(u.WechatRobotKey) != ""
	case "dingtalk_robot":
		return strings.TrimSpace(u.DingTalk) != ""
	case "feishu_robot":
		return strings.TrimSpace(u.FeishuRobotKey) != ""
	case "webhook":
		return strings.TrimSpace(u.WebhookURL) != ""
	}
	return false
}

// ResolveScheduleDelivery is pure; previews and delivery use identical candidate
// and fallback selection. A gap does not silently become a successful delivery.
func ResolveScheduleDelivery(s OnCallSchedule, overrides []OnCallOverride, at time.Time, channel string, users map[int]User, groups map[int][]int) (ScheduleDecision, error) {
	raw, err := ResolveOnCall(s, overrides, at)
	d := ScheduleDecision{ScheduleID: s.ID, UserIDs: []int{}}
	if err != nil {
		return d, err
	}
	d.Source = raw.Source
	d.Reason = raw.Reason
	eligible := func(ids []int) []int {
		seen := map[int]bool{}
		out := []int{}
		for _, id := range ids {
			if u, ok := users[id]; ok && !seen[id] && onCallUserDeliverable(u, channel) {
				out = append(out, id)
				seen[id] = true
			}
		}
		sort.Ints(out)
		return out
	}
	d.UserIDs = eligible(raw.UserIDs)
	if len(d.UserIDs) > 0 {
		return d, nil
	}
	fallback := append([]int{}, s.FallbackUserIDs...)
	for _, g := range s.FallbackGroupIDs {
		fallback = append(fallback, groups[g]...)
	}
	d.UserIDs = eligible(fallback)
	if len(d.UserIDs) > 0 {
		d.Source = "fallback"
		d.Reason = "当前无可用值班人员，使用渠道可达的备用接收人（" + raw.Reason + "）"
	} else {
		d.Source = "uncovered"
		d.Reason = "当前值班人与兜底均无可用渠道地址（" + raw.Reason + "）"
	}
	return d, nil
}

func LoadScheduleDelivery(id int, at time.Time, channel string) (ScheduleDecision, error) {
	s, overrides, err := LoadOnCallSchedule(id)
	if err != nil {
		return ScheduleDecision{ScheduleID: id}, err
	}
	users, groups, err := LoadScheduleDeliveryInputs(s, overrides)
	if err != nil {
		return ScheduleDecision{ScheduleID: id}, err
	}
	return ResolveScheduleDelivery(s, overrides, at, channel, users, groups)
}

func LoadScheduleDeliveryInputs(s OnCallSchedule, overrides []OnCallOverride) (map[int]User, map[int][]int, error) {
	ids := append(append([]int{}, s.Members...), s.FallbackUserIDs...)
	for _, o := range overrides {
		ids = append(ids, o.UserIDs...)
	}
	groups := map[int][]int{}
	if len(s.FallbackGroupIDs) > 0 {
		var gs []UserGroup
		if err := DB.Where("id IN ?", s.FallbackGroupIDs).Find(&gs).Error; err != nil {
			return nil, nil, err
		}
		for _, g := range gs {
			for _, text := range strings.Split(g.Member, ",") {
				id, e := strconv.Atoi(strings.TrimSpace(text))
				if e == nil && id > 0 {
					groups[g.ID] = append(groups[g.ID], id)
					ids = append(ids, id)
				}
			}
		}
	}
	users := map[int]User{}
	if len(ids) > 0 {
		var rows []User
		if err := DB.Select("id", "status", "email", "wechat", "wechat_robot_key", "ding_talk", "feishu_robot_key", "webhook_url").Where("id IN ?", ids).Find(&rows).Error; err != nil {
			return nil, nil, err
		}
		for _, u := range rows {
			users[u.ID] = u
		}
	}
	return users, groups, nil
}

// ResolveDispatchRecipients preserves fixed users/groups and adds on-call users.
func ResolveDispatchRecipients(rules []Rule, groupMap map[string][]string, at time.Time, load func(int, time.Time, string) (ScheduleDecision, error)) ([]DispatchChannel, error) {
	channels := map[string]*DispatchChannel{}
	for _, r := range rules {
		for _, ch := range strings.Split(r.Channel, ",") {
			ch = normalizeChannel(ch)
			if ch == "" {
				continue
			}
			c := channels[ch]
			if c == nil {
				c = &DispatchChannel{Channel: ch, UserIDs: []int{}, Schedules: []ScheduleDecision{}}
				channels[ch] = c
			}
			for _, s := range getRuleUserIDs(r.UserIds, r.GroupIds, groupMap) {
				id, e := strconv.Atoi(strings.TrimSpace(s))
				if e == nil && id > 0 {
					c.UserIDs = append(c.UserIDs, id)
				}
			}
			for _, id := range r.ScheduleIDs {
				d, err := load(id, at, ch)
				if err != nil {
					d = ScheduleDecision{ScheduleID: id, UserIDs: []int{}, Source: "error", Reason: fmt.Sprintf("排班读取失败: %v", err)}
				}
				c.Schedules = append(c.Schedules, d)
				c.UserIDs = append(c.UserIDs, d.UserIDs...)
			}
		}
	}
	out := []DispatchChannel{}
	for _, c := range channels {
		sort.Ints(c.UserIDs)
		unique := []int{}
		for _, id := range c.UserIDs {
			if len(unique) == 0 || unique[len(unique)-1] != id {
				unique = append(unique, id)
			}
		}
		c.UserIDs = unique
		out = append(out, *c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Channel < out[j].Channel })
	return out, nil
}

func SnapshotDispatchRecipients(zid int, eventID int64, c DispatchChannel) (DispatchChannel, error) {
	if eventID <= 0 || len(c.Schedules) == 0 {
		return c, nil
	}
	snap := OnCallDispatchSnapshot{ZID: zid, EventID: eventID, Channel: c.Channel, UserIDs: c.UserIDs, Decisions: c.Schedules}
	if err := DB.Clauses(clause.OnConflict{DoNothing: true}).Create(&snap).Error; err != nil {
		return c, err
	}
	var saved OnCallDispatchSnapshot
	if err := DB.Where("zid = ? AND event_id = ? AND channel = ?", zid, eventID, c.Channel).First(&saved).Error; err != nil {
		return c, err
	}
	c.UserIDs = saved.UserIDs
	c.Schedules = saved.Decisions
	return c, nil
}

func DispatchRuleTimeMatches(at time.Time, r Rule) bool { return !isNoneAlarm(at, &r) }
func LoadDispatchGroups(rules []Rule) (map[string][]string, error) {
	ids := map[string]struct{}{}
	for _, r := range rules {
		for _, id := range strings.Split(r.GroupIds, ",") {
			if id != "" {
				ids[id] = struct{}{}
			}
		}
	}
	return getBatchUsers(nil, ids)
}

// Existing snapshots take precedence over current shift availability.
func FindOnCallDispatchSnapshot(zid int, eventID int64, channel string) (OnCallDispatchSnapshot, bool, error) {
	var s OnCallDispatchSnapshot
	if eventID <= 0 {
		return s, false, nil
	}
	err := DB.Where("zid = ? AND event_id = ? AND channel = ?", zid, eventID, channel).First(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s, false, nil
	}
	return s, err == nil, err
}

func ResolveEventDispatchRecipients(rules []Rule, groupMap map[string][]string, at time.Time, zid int, eventID int64,
	load func(int, time.Time, string) (ScheduleDecision, error), find func(int, int64, string) (OnCallDispatchSnapshot, bool, error)) ([]DispatchChannel, error) {
	snapshots := map[string]OnCallDispatchSnapshot{}
	checked := map[string]bool{}
	lookup := func(ch string) (OnCallDispatchSnapshot, bool, error) {
		if checked[ch] {
			s, ok := snapshots[ch]
			return s, ok, nil
		}
		s, ok, err := find(zid, eventID, ch)
		if err != nil {
			return s, false, err
		}
		checked[ch] = true
		if ok {
			snapshots[ch] = s
		}
		return s, ok, nil
	}
	channels, err := ResolveDispatchRecipients(rules, groupMap, at, func(id int, t time.Time, ch string) (ScheduleDecision, error) {
		s, ok, err := lookup(ch)
		if err != nil {
			return ScheduleDecision{}, err
		}
		if ok {
			return ScheduleDecision{ScheduleID: id, UserIDs: s.UserIDs, Source: "snapshot", Reason: "沿用该事件首次生成的接收人"}, nil
		}
		return load(id, t, ch)
	})
	if err != nil {
		return nil, err
	}
	for i := range channels {
		s, ok, err := lookup(channels[i].Channel)
		if err != nil {
			return nil, err
		}
		if ok {
			channels[i].UserIDs = s.UserIDs
			channels[i].Schedules = s.Decisions
		}
	}
	return channels, nil
}
