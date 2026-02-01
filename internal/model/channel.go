package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"

	"github.com/Knetic/govaluate"
)

// 根据规则生成告警
func GenAlert(alarm *Alarm) bool {
	// 查询匹配当前实例的普通规则（m_type = 1）
	fmt.Println("ADDDDDDDDDD")
	var rules []Rule
	var err error
	fmt.Println(alarm.ZID)
	if alarm.ZID > 0 {
		// 使用 FIND_IN_SET 或 LIKE 来匹配实例ID（存储的是数字ID，用逗号分隔）
		query := DB.Model(&Rule{}).
			Where("m_type = ?", "1").
			Where("status = ?", "0").
			Where("(FIND_IN_SET(?, z_ids) > 0 OR z_ids LIKE ?)", strconv.Itoa(alarm.ZID), "%"+strconv.Itoa(alarm.ZID)+"%")
		err = query.Select("id", "name", "conditions", "z_ids", "note", "s_week",
			"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created").
			Find(&rules).Error
	} else {
		// 如果没有实例ID，查询所有普通规则
		query := DB.Model(&Rule{}).
			Where("m_type = ?", "1").
			Where("status = ?", "0")
		err = query.Select("id", "name", "conditions", "z_ids", "note", "s_week",
			"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created").
			Find(&rules).Error
	}
	fmt.Println(rules)
	if err != nil {
		logger.Log.Error("查询普通规则失败:", err)
		return true
	}

	//找不到实例匹配的规则，走默认规则
	if len(rules) == 0 {
		//更新规则，告警走默认规则
		ala := Alarm{ID: alarm.ID, NotifyStatus: strconv.Itoa(NotifyDefault)}
		_, err := UpdateAlarmStatus(&ala)
		if err != nil {
			logger.Log.Error(err)
		}
		//select default rule
		var rules []Rule

		// 优先查找匹配当前实例ID的默认规则
		if alarm.ZID > 0 {
			// 使用 FIND_IN_SET 或者精确匹配来查找包含当前实例ID的规则
			query := DB.Model(&Rule{}).
				Where("m_type = ?", "2").
				Where("status = ?", "0").
				Where("(FIND_IN_SET(?, z_ids) > 0 OR z_ids LIKE ?)", strconv.Itoa(alarm.ZID), "%"+strconv.Itoa(alarm.ZID)+"%")
			err = query.Select("id", "name", "conditions", "z_ids", "note", "s_week",
				"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created").
				Find(&rules).Error
			if err != nil {
				logger.Log.Error(err)
				return false
			}
		}

		// 如果没有找到匹配实例的默认规则，则查找全局默认规则（z_ids 为 "*"）
		if len(rules) == 0 {
			query := DB.Model(&Rule{}).
				Where("m_type = ?", "2").
				Where("status = ?", "0").
				Where("z_ids = ?", "*")
			err = query.Select("id", "name", "conditions", "z_ids", "note", "s_week",
				"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created").
				Find(&rules).Error
			if err != nil {
				logger.Log.Error(err)
				return false
			}
		}
		fmt.Println("Found rule", rules)

		//default rule disable
		if len(rules) == 0 {
			logger.Log.Errorf("default rule is null for zid: %d", alarm.ZID)
			return false
		}
		//event
		event := &Event{
			ID:            alarm.ID,
			ZID:           alarm.ZID,
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
	//遍历rules
	count := 0
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
			ZID:           alarm.ZID,
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
	fmt.Println("aaa", event)
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
				logger.Log.Error(err)
			}
			continue
		}
		//GetEventUser
		toUsers, err := GetEventUser(event.GroupIds, event.UserIds)
		if err != nil {
			logger.Log.Error(err)
			return
		}
		if len(toUsers) == 0 {
			return
		}
		userList := strings.Join(toUsers, ",")
		event.ToUsers = userList
		p, _ := json.Marshal(event)
		fmt.Println(string(p))
		err = CacheLPush(v, p)
		if err != nil {
			logger.Log.Error(err)
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
			logger.Log.Error(err)
			return
		}
	}
}

// mut
func IsMuted(event *Event) bool {
	// 查询匹配当前实例的屏蔽规则（m_type = 3）
	var rules []Rule
	var err error

	if event.ZID > 0 {
		// 使用 FIND_IN_SET 或 LIKE 来匹配实例ID
		query := DB.Model(&Rule{}).
			Where("m_type = ?", "3").
			Where("status = ?", "0").
			Where("(FIND_IN_SET(?, z_ids) > 0 OR z_ids LIKE ?)", strconv.Itoa(event.ZID), "%"+strconv.Itoa(event.ZID)+"%")
		err = query.Select("id", "name", "conditions", "z_ids", "note", "s_week",
			"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created").
			Find(&rules).Error
	} else {
		// 如果没有实例ID，查询所有屏蔽规则
		query := DB.Model(&Rule{}).
			Where("m_type = ?", "3").
			Where("status = ?", "0")
		err = query.Select("id", "name", "conditions", "z_ids", "note", "s_week",
			"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created").
			Find(&rules).Error
	}

	if err != nil {
		logger.Log.Error("查询屏蔽规则失败:", err)
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
			logger.Log.Error(err)
			continue
		}
		result, err := expression.Evaluate(nil)
		if err != nil {
			//return true and event not send！！
			logger.Log.Error(err)
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
			logger.Log.Error(err)
			continue
		}
		result, err := expression.Evaluate(nil)
		if err != nil {
			//return true and event not send！！
			logger.Log.Error(err)
			continue
		}
		if result.(bool) {
			count++
		}
	}
	return count == len(conds)
}
