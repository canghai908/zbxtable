package model

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"zbxtable/pkg/logger"
)

type aggregationBucket struct {
	Key              string
	RuleID           int
	RuleName         string
	WindowSec        int
	WindowStart      time.Time
	WindowEnd        time.Time
	TotalCount       int
	SuppressedEvents []Event
}

var (
	aggregationMu      sync.Mutex
	aggregationBuckets = make(map[string]*aggregationBucket)
	lifecycleOnce      sync.Once
)

// ResetAggregationBuckets clears in-memory storm suppression windows (primarily for tests).
func ResetAggregationBuckets() {
	aggregationMu.Lock()
	aggregationBuckets = make(map[string]*aggregationBucket)
	aggregationMu.Unlock()
}

func updateAlarmDispatchState(alarmID int, notifyStatus string, matchedRuleID int) {
	updates := map[string]any{
		"notify_status": notifyStatus,
	}
	if matchedRuleID > 0 {
		updates["matched_rule_id"] = matchedRuleID
	}
	if err := DB.Model(&Alarm{}).Where("id = ?", alarmID).Updates(updates).Error; err != nil {
		logger.Log.Error("update alarm dispatch state error:", err)
	}
}

func isProblemStatus(status string) bool {
	s := strings.TrimSpace(status)
	return s == "1" || s == "故障"
}

func isRecoveryStatus(status string) bool {
	s := strings.TrimSpace(status)
	return s == "0" || s == "恢复"
}

func markRecoveredAlarmsClosed(alarm *Alarm) {
	if alarm == nil || DB == nil || !isRecoveryStatus(alarm.Status) {
		return
	}
	q := DB.Model(&Alarm{}).
		Where("zid = ? AND escalation_step = 0 AND (status = ? OR status = ?)", alarm.ZID, "1", "故障")
	switch {
	case alarm.EventID > 0 && alarm.TriggerID > 0:
		q = q.Where("event_id = ? OR (host = ? AND trigger_id = ?)", alarm.EventID, alarm.Host, alarm.TriggerID)
	case alarm.EventID > 0:
		q = q.Where("event_id = ?", alarm.EventID)
	case alarm.TriggerID > 0:
		q = q.Where("host = ? AND trigger_id = ?", alarm.Host, alarm.TriggerID)
	default:
		return
	}
	if err := q.Update("escalation_step", -1).Error; err != nil {
		logger.Log.Warnf("close recovered alarms for escalation failed: %v", err)
	}
}

func shouldSuppressAndBuffer(rule *Rule, alarm *Alarm, event *Event, now time.Time) bool {
	if rule == nil || rule.SuppressWindow <= 0 {
		return false
	}
	threshold := rule.SuppressThreshold
	if threshold <= 0 {
		threshold = 1
	}
	key := fmt.Sprintf("%d:%d:%s:%s:%s", alarm.ZID, rule.ID, event.Channel, alarm.Host, alarm.Status)

	var expiredToFlush *aggregationBucket

	aggregationMu.Lock()
	bucket, exists := aggregationBuckets[key]
	if exists && !now.Before(bucket.WindowEnd) {
		if len(bucket.SuppressedEvents) > 0 {
			expiredToFlush = bucket
		}
		delete(aggregationBuckets, key)
		exists = false
	}
	if !exists {
		aggregationBuckets[key] = &aggregationBucket{
			Key:         key,
			RuleID:      rule.ID,
			RuleName:    rule.Name,
			WindowSec:   rule.SuppressWindow,
			WindowStart: now,
			WindowEnd:   now.Add(time.Duration(rule.SuppressWindow) * time.Second),
			TotalCount:  1,
		}
		aggregationMu.Unlock()
		if expiredToFlush != nil {
			sendToQueue(buildAggregatedSummaryEvent(expiredToFlush))
		}
		return false
	}

	bucket.TotalCount++
	if bucket.TotalCount <= threshold {
		aggregationMu.Unlock()
		return false
	}

	bucket.SuppressedEvents = append(bucket.SuppressedEvents, *event)
	aggregationMu.Unlock()
	return true
}

func buildAggregatedSummaryEvent(bucket *aggregationBucket) *Event {
	first := bucket.SuppressedEvents[0]
	summary := first
	hostLabel := first.Hostname
	if strings.TrimSpace(hostLabel) == "" {
		hostLabel = first.Host
	}
	summary.Message = fmt.Sprintf("[风暴抑制汇总] %s 在 %ds 内共触发 %d 条告警（已合并 %d 条）",
		hostLabel, bucket.WindowSec, bucket.TotalCount, len(bucket.SuppressedEvents))

	limit := 5
	if len(bucket.SuppressedEvents) < limit {
		limit = len(bucket.SuppressedEvents)
	}
	details := make([]string, 0, limit+1)
	for i := 0; i < limit; i++ {
		ev := bucket.SuppressedEvents[i]
		if strings.TrimSpace(ev.Detail) != "" {
			details = append(details, fmt.Sprintf("%d. %s (%s)", i+1, ev.Message, ev.Detail))
		} else {
			details = append(details, fmt.Sprintf("%d. %s", i+1, ev.Message))
		}
	}
	if len(bucket.SuppressedEvents) > limit {
		details = append(details, fmt.Sprintf("... 共合并 %d 条", len(bucket.SuppressedEvents)))
	}
	summary.Detail = strings.Join(details, "; ")
	summary.OccurTime = time.Now()
	return &summary
}

// FlushExpiredAggregations flushes all expired storm-suppression windows and enqueues summary events.
func FlushExpiredAggregations(now time.Time) int {
	var expired []*aggregationBucket

	aggregationMu.Lock()
	for k, bucket := range aggregationBuckets {
		if !now.Before(bucket.WindowEnd) {
			expired = append(expired, bucket)
			delete(aggregationBuckets, k)
		}
	}
	aggregationMu.Unlock()

	flushed := 0
	for _, bucket := range expired {
		if len(bucket.SuppressedEvents) == 0 {
			continue
		}
		sendToQueue(buildAggregatedSummaryEvent(bucket))
		flushed++
	}
	return flushed
}

// AcknowledgeAlarms marks alarms as acknowledged and optionally syncs acknowledgement to Zabbix.
func AcknowledgeAlarms(ids []int, username, note string, syncZabbix bool) (int64, error) {
	if DB == nil {
		return 0, errors.New("database not initialized")
	}
	seen := make(map[int]bool, len(ids))
	cleanIDs := make([]int, 0, len(ids))
	for _, id := range ids {
		if id > 0 && !seen[id] {
			seen[id] = true
			cleanIDs = append(cleanIDs, id)
		}
	}
	if len(cleanIDs) == 0 {
		return 0, errors.New("请选择有效的告警ID")
	}

	now := time.Now()
	cleanNote := strings.TrimSpace(note)
	res := DB.Model(&Alarm{}).Where("id IN ?", cleanIDs).Updates(map[string]any{
		"ack_status": 1,
		"ack_user":   username,
		"ack_time":   &now,
		"ack_note":   cleanNote,
	})
	if res.Error != nil {
		return 0, res.Error
	}

	if syncZabbix && res.RowsAffected > 0 {
		var alarms []Alarm
		if err := DB.Where("id IN ? AND zid > 0 AND event_id > 0", cleanIDs).Find(&alarms).Error; err == nil {
			for _, a := range alarms {
				inst, err := GetAPIByZID(a.ZID)
				if err != nil || inst == nil || inst.API == nil {
					logger.Log.Warnf("Zabbix API unavailable for ack sync (zid=%d, eventid=%d): %v", a.ZID, a.EventID, err)
					continue
				}
				msg := cleanNote
				if msg == "" {
					msg = fmt.Sprintf("Acknowledged by %s via ZbxTable", username)
				} else {
					msg = fmt.Sprintf("[%s] %s", username, msg)
				}
				if _, err := inst.API.CallWithError("event.acknowledge", Params{
					"eventids": strconv.FormatInt(a.EventID, 10),
					"action":   6,
					"message":  msg,
				}); err != nil {
					logger.Log.Warnf("Zabbix event.acknowledge failed (zid=%d, eventid=%d): %v", a.ZID, a.EventID, err)
				}
			}
		}
	}
	return res.RowsAffected, nil
}

func isAlarmRecovered(alarm Alarm) bool {
	if DB == nil || alarm.ZID <= 0 {
		return false
	}
	q := DB.Model(&Alarm{}).
		Where("zid = ? AND (status = ? OR status = ?) AND occurtime >= ?", alarm.ZID, "0", "恢复", alarm.OccurTime)
	switch {
	case alarm.EventID > 0 && alarm.TriggerID > 0:
		q = q.Where("event_id = ? OR (host = ? AND trigger_id = ?)", alarm.EventID, alarm.Host, alarm.TriggerID)
	case alarm.EventID > 0:
		q = q.Where("event_id = ?", alarm.EventID)
	case alarm.TriggerID > 0:
		q = q.Where("host = ? AND trigger_id = ?", alarm.Host, alarm.TriggerID)
	default:
		return false
	}
	var count int64
	if err := q.Count(&count).Error; err != nil {
		return false
	}
	return count > 0
}

func resolveEscalationChannels(rule Rule) []string {
	raw := strings.TrimSpace(rule.EscalationChannel)
	if raw == "" {
		raw = rule.Channel
	}
	seen := make(map[string]bool)
	channels := make([]string, 0)
	for _, ch := range strings.Split(raw, ",") {
		norm := normalizeChannel(ch)
		if norm == "" || seen[norm] {
			continue
		}
		seen[norm] = true
		channels = append(channels, norm)
	}
	return channels
}

func resolveEscalationRecipientsByChannel(rule Rule, channels []string) map[string][]int {
	result := make(map[string][]int)
	if len(channels) == 0 {
		return result
	}

	// 1. 优先使用规则指定的升级接收人/组
	if strings.TrimSpace(rule.EscalationUserIDs) != "" || strings.TrimSpace(rule.EscalationGroupIDs) != "" {
		gids := map[string]struct{}{}
		for _, gid := range strings.Split(rule.EscalationGroupIDs, ",") {
			if strings.TrimSpace(gid) != "" {
				gids[strings.TrimSpace(gid)] = struct{}{}
			}
		}
		groupMap, err := getBatchUsers(nil, gids)
		if err != nil {
			logger.Log.Warnf("load escalation group users failed: %v", err)
			groupMap = map[string][]string{}
		}
		idStrs := getRuleUserIDs(rule.EscalationUserIDs, rule.EscalationGroupIDs, groupMap)
		var uids []int
		for _, s := range idStrs {
			if id, e := strconv.Atoi(strings.TrimSpace(s)); e == nil && id > 0 {
				uids = append(uids, id)
			}
		}
		if len(uids) == 0 {
			return result
		}
		var users []User
		if err := DB.Select("id", "status", "email", "wechat", "wechat_robot_key", "ding_talk", "feishu_robot_key", "webhook_url").
			Where("id IN ?", uids).Find(&users).Error; err != nil {
			logger.Log.Warnf("load escalation users failed: %v", err)
			return result
		}
		userMap := make(map[int]User, len(users))
		for _, u := range users {
			userMap[u.ID] = u
		}
		for _, ch := range channels {
			seen := map[int]bool{}
			var targets []int
			for _, id := range uids {
				if u, ok := userMap[id]; ok && !seen[id] && onCallUserDeliverable(u, ch) {
					seen[id] = true
					targets = append(targets, id)
				}
			}
			sort.Ints(targets)
			if len(targets) > 0 {
				result[ch] = targets
			}
		}
		return result
	}

	// 2. 未显式配置升级接收人时，回退至规则关联排班的 FallbackUserIDs / FallbackGroupIDs
	if len(rule.ScheduleIDs) == 0 {
		return result
	}
	for _, sid := range rule.ScheduleIDs {
		s, overrides, err := LoadOnCallSchedule(sid)
		if err != nil {
			continue
		}
		users, groups, err := LoadScheduleDeliveryInputs(s, overrides)
		if err != nil {
			continue
		}
		fallbackIDs := append([]int{}, s.FallbackUserIDs...)
		for _, gid := range s.FallbackGroupIDs {
			fallbackIDs = append(fallbackIDs, groups[gid]...)
		}
		for _, ch := range channels {
			seen := map[int]bool{}
			var targets []int
			for _, id := range fallbackIDs {
				if u, ok := users[id]; ok && !seen[id] && onCallUserDeliverable(u, ch) {
					seen[id] = true
					targets = append(targets, id)
				}
			}
			// 若兜底未命中，尝试排班全部成员中该渠道可达的人员
			if len(targets) == 0 {
				for _, id := range s.Members {
					if u, ok := users[id]; ok && !seen[id] && onCallUserDeliverable(u, ch) {
						seen[id] = true
						targets = append(targets, id)
					}
				}
			}
			if len(targets) > 0 {
				existing := map[int]bool{}
				for _, id := range result[ch] {
					existing[id] = true
				}
				for _, id := range targets {
					if !existing[id] {
						result[ch] = append(result[ch], id)
						existing[id] = true
					}
				}
				sort.Ints(result[ch])
			}
		}
	}
	return result
}

// ProcessDueEscalations scans unacknowledged problem alarms past rule escalation_minutes and dispatches escalation alerts.
func ProcessDueEscalations(now time.Time) (int, error) {
	if DB == nil {
		return 0, nil
	}
	var alarms []Alarm
	if err := DB.Where("ack_status = 0 AND escalation_step = 0 AND matched_rule_id > 0 AND (status = ? OR status = ?)", "1", "故障").
		Find(&alarms).Error; err != nil {
		return 0, err
	}
	if len(alarms) == 0 {
		return 0, nil
	}

	escalatedCount := 0
	for _, alarm := range alarms {
		if isAlarmRecovered(alarm) {
			_ = DB.Model(&Alarm{}).Where("id = ? AND escalation_step = 0", alarm.ID).Update("escalation_step", -1).Error
			continue
		}

		var rule Rule
		if err := DB.Where("id = ?", alarm.MatchedRuleID).First(&rule).Error; err != nil {
			continue
		}
		if rule.Status != "0" || rule.EscalationMinutes <= 0 {
			continue
		}
		if now.Sub(alarm.OccurTime) < time.Duration(rule.EscalationMinutes)*time.Minute {
			continue
		}

		channels := resolveEscalationChannels(rule)
		recipientsByChannel := resolveEscalationRecipientsByChannel(rule, channels)
		if len(recipientsByChannel) == 0 {
			continue
		}

		sentForAlarm := 0
		for _, ch := range channels {
			uids := recipientsByChannel[ch]
			if len(uids) == 0 {
				continue
			}
			uList := make([]string, 0, len(uids))
			for _, id := range uids {
				uList = append(uList, strconv.Itoa(id))
			}
			event := &Event{
				ID:            alarm.ID,
				ZID:           alarm.ZID,
				Host:          alarm.Host,
				Hgroup:        alarm.Hgroup,
				ItemName:      alarm.ItemName,
				Hkey:          alarm.Hkey,
				Message:       fmt.Sprintf("[告警升级/超时%d分钟未认领] %s", rule.EscalationMinutes, alarm.Message),
				Level:         alarm.Level,
				OccurTime:     alarm.OccurTime,
				Channel:       ch,
				Rule:          rule.Name + " [升级]",
				RuleType:      rule.MType,
				Hostname:      alarm.Hostname,
				HostsIP:       alarm.HostsIP,
				TriggerID:     alarm.TriggerID,
				ItemID:        alarm.ItemID,
				ItemValue:     alarm.ItemValue,
				Detail:        alarm.Detail,
				Status:        alarm.Status,
				EventID:       alarm.EventID,
				EventDuration: alarm.EventDuration,
				ToUsers:       strings.Join(uList, ","),
			}
			if IsMuted(event) {
				continue
			}
			sendToQueue(event)
			sentForAlarm++
		}

		if sentForAlarm > 0 {
			if err := DB.Model(&Alarm{}).Where("id = ? AND escalation_step = 0", alarm.ID).Updates(map[string]any{
				"escalation_step": 1,
				"escalated_at":    &now,
				"notify_status":   strconv.Itoa(NotifyEscalated),
			}).Error; err == nil {
				escalatedCount++
			}
		}
	}
	return escalatedCount, nil
}

// StartAlertLifecycleWorker launches the background loop that flushes expired aggregation windows and checks alarm escalations.
func StartAlertLifecycleWorker() {
	lifecycleOnce.Do(func() {
		go func() {
			aggTicker := time.NewTicker(2 * time.Second)
			escTicker := time.NewTicker(30 * time.Second)
			defer aggTicker.Stop()
			defer escTicker.Stop()

			for {
				select {
				case now := <-aggTicker.C:
					FlushExpiredAggregations(now)
				case now := <-escTicker.C:
					if _, err := ProcessDueEscalations(now); err != nil {
						logger.Log.Warnf("process due escalations failed: %v", err)
					}
				}
			}
		}()
	})
}
