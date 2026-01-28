package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/utils"

	"github.com/Knetic/govaluate"
)

// alert gen by rules
func GenAlert(alarm *Alarm) bool {
	var rules []Rule
	query := DB.Model(&Rule{}).
		Where("tenant_id LIKE ?", "%"+alarm.TenantID+"%").
		Where("m_type = ?", "1").
		Where("status = ?", "0")
	err := query.Select("id", "name", "conditions", "tenant_id", "note", "s_week",
		"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created").
		Find(&rules).Error
	if err != nil {
		return true
	}
	count := 0
	if len(rules) == 0 {
		count = 1
	}
	//遍历rules
	for _, v := range rules {
		//rfunc
		if !MeetConditions(alarm, &v) {
			count++
			continue
		}
		//time and week
		if isNoneAlarm(alarm.OccurTime, &v) {
			count++
			continue
		}
		//event
		event := &Event{
			ID:            alarm.ID,
			TenantID:      alarm.TenantID,
			HostID:        alarm.HostID,
			Hostname:      alarm.Hostname,
			Host:          alarm.Host,
			HostsIP:       alarm.HostsIP,
			TriggerID:     alarm.TriggerID,
			ItemID:        alarm.ItemID,
			ItemName:      alarm.ItemName,
			ItemValue:     alarm.ItemValue,
			Hgroup:        alarm.Hgroup,
			OccurTime:     alarm.OccurTime,
			Level:         alarm.Level,
			Message:       alarm.Message,
			Hkey:          alarm.Hkey,
			Detail:        alarm.Detail,
			Status:        alarm.Status,
			EventID:       alarm.EventID,
			EventDuration: alarm.EventDuration,
			Rule:          v.Name,
			RuleType:      strconv.Itoa(RuleCust),
			Channel:       v.Channel,
			UserIds:       v.UserIds,
			GroupIds:      v.GroupIds,
		}
		//push event to redis
		sendEvent(event)
	}
	//no rule select
	if count > 0 {
		//update alarm notify status to NotifyMuted
		ala := Alarm{ID: alarm.ID, NotifyStatus: strconv.Itoa(NotifyDefault)}
		_, err := UpdateAlarmStatus(&ala)
		if err != nil {
			utils.Log.Error(err)
			return false
		}
		//select default rule
		// 优先查找匹配当前租户的默认规则
		var rules []Rule
		fmt.Println("CCC", alarm.TenantID)
		query := DB.Model(&Rule{}).
			Where("tenant_id LIKE ?", "%"+alarm.TenantID+"%").
			Where("m_type = ?", "1").
			Where("status = ?", "0").Debug()
		err = query.Select("id", "name", "conditions", "tenant_id", "note", "s_week",
			"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created").
			Find(&rules).Error
		if err != nil {
			utils.Log.Error(err)
			return false
		}

		// 如果没有找到匹配租户的默认规则，则查找全局默认规则（tenant_id 为空或 "*"）
		if len(rules) == 0 {
			query = DB.Model(&Rule{}).
				Where("m_type = ?", "2").
				Where("status = ?", "0").
				Where("(tenant_id = ? OR tenant_id = ? OR tenant_id IS NULL)", "", "*")
			err = query.Select("id", "name", "conditions", "tenant_id", "note", "s_week",
				"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created").
				Find(&rules).Error
			if err != nil {
				utils.Log.Error(err)
				return false
			}
		}

		//default rule disable
		if len(rules) == 0 {
			utils.Log.Errorf("default rule is null for tenant: %s", alarm.TenantID)
			return false
		}
		//event
		event := &Event{
			ID:            alarm.ID,
			TenantID:      alarm.TenantID,
			HostID:        alarm.HostID,
			Hostname:      alarm.Hostname,
			Host:          alarm.Host,
			HostsIP:       alarm.HostsIP,
			TriggerID:     alarm.TriggerID,
			ItemID:        alarm.ItemID,
			ItemName:      alarm.ItemName,
			ItemValue:     alarm.ItemValue,
			Hgroup:        alarm.Hgroup,
			OccurTime:     alarm.OccurTime,
			Level:         alarm.Level,
			Message:       alarm.Message,
			Hkey:          alarm.Hkey,
			Detail:        alarm.Detail,
			Status:        alarm.Status,
			EventID:       alarm.EventID,
			EventDuration: alarm.EventDuration,
			Rule:          rules[0].Name,
			RuleType:      strconv.Itoa(RuleDefault),
			Channel:       rules[0].Channel,
			UserIds:       rules[0].UserIds,
			GroupIds:      rules[0].GroupIds,
		}
		//push event to redis
		sendEvent(event)
	}
	return true
}

// GetEventUser 查找事件用户信息
func GetEventUser(groupIds, userIds string) (list []string, err error) {
	//用户组信息,判断用户组是否为空
	var guidList []string
	if groupIds != "" {
		var gList []UserGroup
		//组分隔
		gids := strings.Split(groupIds, ",")
		err = DB.Where("id IN ?", gids).Find(&gList).Error
		if err != nil {
			return
		}
		//get group userid
		if len(gList) != 0 {
			for _, v := range gList {
				ids := strings.Split(v.Member, ",")
				guidList = append(guidList, ids...)
			}
		}
	}
	var ids []string
	uid := strings.Split(userIds, ",")
	if len(guidList) != 0 {
		//添加用户组
		ids = utils.UniqueArr(utils.MergeArr(guidList, uid))
	} else {
		//添加用户
		ids = uid
	}
	//get all userids unique
	if len(ids) != 0 {
		var plist []Manager
		err = DB.Where("id IN ?", ids).Select("id", "username",
			"email", "wechat", "wechat_robot_key", "phone", "ding_talk").Find(&plist).Error
		if err != nil {
			return []string{}, err
		}
		//get ids
		var eList []string
		for _, v := range plist {
			eList = append(eList, strconv.Itoa(v.ID))
		}
		return eList, nil
	}
	return []string{}, err
}
func sendEvent(event *Event) {
	if len(event.Channel) == 0 {
		return
	}
	channel := strings.Split(event.Channel, ",")
	//遍历channel
	for _, v := range channel {
		event.Channel = v
		//event mute
		var alarm Alarm
		if IsMuted(event) {
			//update alarm notify status to NotifyMuted
			alarm = Alarm{ID: event.ID, NotifyStatus: strconv.Itoa(NotifyMuted)}
			_, err := UpdateAlarmStatus(&alarm)
			if err != nil {
				utils.Log.Error(err)
			}
			continue
		}
		//GetEventUser
		toUsers, err := GetEventUser(event.GroupIds, event.UserIds)
		if err != nil {
			utils.Log.Error(err)
			return
		}
		if len(toUsers) == 0 {
			return
		}
		userList := strings.Join(toUsers, ",")
		event.ToUsers = userList
		p, _ := json.Marshal(event)
		err = CacheLPush(v, p)
		if err != nil {
			utils.Log.Error(err)
			return
		}
		//update alarm status
		var notifyStatus string
		if event.RuleType == strconv.Itoa(RuleDefault) {
			notifyStatus = strconv.Itoa(NotifyDefault)
		} else {
			notifyStatus = strconv.Itoa(NotifySuccess)
		}
		alarm = Alarm{ID: event.ID, NotifyStatus: notifyStatus}
		_, err = UpdateAlarmStatus(&alarm)
		if err != nil {
			utils.Log.Error(err)
			return
		}
	}
}

// mut
func IsMuted(event *Event) bool {
	var rules []Rule
	query := DB.Model(&Rule{}).
		Where("tenant_id LIKE ?", "%"+event.TenantID+"%").
		Where("m_type = ?", "3").
		Where("status = ?", "0")
	err := query.Select("id", "name", "conditions", "tenant_id", "note", "s_week",
		"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created").
		Find(&rules).Error
	if err != nil {
		return false
	}
	if len(rules) == 0 {
		return false
	}
	//mute rule
	for _, v := range rules {
		//not match
		if !MeetEventConditions(event, &v) {
			continue
		}
		//not match
		if IsMuteTime(event, &v) {
			continue
		}
		//not match
		if isMuteChannel(event, &v) {
			continue
		}
		//not muted
		return true
	}
	return false
}

// IsMuteTime not
func IsMuteTime(event *Event, rule *Rule) bool {
	stime, _ := utils.ParTime(rule.Stime)
	etime, _ := utils.ParTime(rule.Etime)
	triggerTime := event.OccurTime.Unix()
	//stime
	if stime.Unix() <= etime.Unix() {
		if triggerTime < stime.Unix() || triggerTime > etime.Unix() {
			return true
		}
	} else {
		if triggerTime < stime.Unix() && triggerTime > etime.Unix() {
			return true
		}
	}
	return false

}

func MeetEventConditions(event *Event, rule *Rule) bool {
	//var value string
	var conds []Conditions
	err := json.Unmarshal([]byte(rule.Conditions), &conds)
	if err != nil {
		return false
	}
	if len(conds) == 0 {
		return false
	}
	count := 0
	for _, v := range conds {
		var val string
		switch v.RType {
		case "host":
			val = event.Host
		case "group":
			val = event.Hgroup
		case "item":
			val = event.ItemName
		case "key":
			val = event.Hkey
		case "trigger":
			val = event.Message
		case "severity":
			val = event.Level
		}
		expression, err := govaluate.NewEvaluableExpression("'" + val + "'" + v.RFunc + "'" + v.Rvalue + "'")
		if err != nil {
			//return true and event not send！！
			utils.Log.Error(err)
			continue
		}
		result, err := expression.Evaluate(nil)
		if err != nil {
			//return true and event not send！！
			utils.Log.Error(err)
			continue
		}
		if result.(bool) {
			count++
		}
	}
	return count == len(conds)
}

func isMuteChannel(event *Event, rule *Rule) bool {
	return !strings.Contains(rule.Channel, event.Channel)
}

// no alarm
func isNoneAlarm(occurtime time.Time, rule *Rule) bool {
	triggerTime := occurtime.Format("15:04")
	triggerWeek := strconv.Itoa(int(occurtime.Weekday()))
	//stime
	if rule.Stime <= rule.Etime {
		if triggerTime < rule.Stime || triggerTime > rule.Etime {
			return true
		}
	} else {
		if triggerTime < rule.Stime && triggerTime > rule.Etime {
			return true
		}
	}
	//sweek
	return !strings.Contains(rule.Sweek, triggerWeek)
}

func MeetConditions(alarm *Alarm, rule *Rule) bool {
	//var value string
	var conds []Conditions
	err := json.Unmarshal([]byte(rule.Conditions), &conds)
	if err != nil {
		return false
	}
	if len(conds) == 0 {
		return false
	}
	count := 0
	for _, v := range conds {
		var val string
		switch v.RType {
		case "host":
			val = alarm.Host
		case "group":
			val = alarm.Hgroup
		case "item":
			val = alarm.ItemName
		case "key":
			val = alarm.Hkey
		case "trigger":
			val = alarm.Message
		case "severity":
			val = alarm.Level
		}
		expression, err := govaluate.NewEvaluableExpression("'" + val + "'" + v.RFunc + "'" + v.Rvalue + "'")
		if err != nil {
			//return true and event not send！！
			utils.Log.Error(err)
			continue
		}
		result, err := expression.Evaluate(nil)
		if err != nil {
			//return true and event not send！！
			utils.Log.Error(err)
			continue
		}
		if result.(bool) {
			count++
		}
	}
	return count == len(conds)
}
