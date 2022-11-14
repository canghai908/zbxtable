package models

import (
	"context"
	"errors"
	"github.com/Knetic/govaluate"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"strconv"
	"strings"
	"time"
	"zbxtable/utils"
)

//alert gen by rules
func GenAlert(alarm *Alarm) bool {
	o := orm.NewOrm()
	var rules []Rule
	al := new(Rule)
	//cond
	cond := orm.NewCondition()
	//tenant
	cond = cond.And("tenant_id__icontains", alarm.TenantID)
	//alert rule
	cond = cond.And("m_type", "1")
	//enable rule
	cond = cond.And("status", "0")
	_, err := o.QueryTable(al).SetCond(cond).
		All(&rules, "id", "name", "conditions", "tenant_id", "note", "s_week",
			"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created")
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
			logs.Error(err)
			return false
		}
		//select default rule
		o := orm.NewOrm()
		var rules []Rule
		al := new(Rule)
		//cond
		cond := orm.NewCondition()
		//tenant
		cond = cond.And("tenant_id__icontains", alarm.TenantID)
		//alert rule
		cond = cond.And("m_type", "2")
		//enable rule
		cond = cond.And("status", "0")
		_, err = o.QueryTable(al).SetCond(cond).
			All(&rules, "id", "name", "conditions", "tenant_id", "note", "s_week",
				"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created")
		if err != nil {
			logs.Error(err)
			return false
		}
		//default rule disable
		if len(rules) == 0 {
			logs.Error(errors.New("default rule is null"))
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

func GetEventUser(groupIds, userIds, channel string) (list []string, err error) {
	o := orm.NewOrm()
	//get grouids
	var group UserGroup
	var gList []UserGroup
	gids := strings.Split(groupIds, ",")
	_, err = o.QueryTable(group).Filter("id__in", gids).All(&gList)
	if err != nil {
		return []string{}, err
	}
	//get group userid
	var guidList []string
	if len(gList) != 0 {
		for _, v := range gList {
			ids := strings.Split(v.Member, ",")
			for _, vv := range ids {
				guidList = append(guidList, vv)
			}
		}
	}
	uid := strings.Split(userIds, ",")
	//get all userids unique
	ids := utils.UniqueArr(utils.MergeArr(guidList, uid))
	if len(ids) != 0 {

		var user Manager
		var plist []Manager
		_, err = o.QueryTable(user).Filter("id__in", ids).All(&plist, "id", "username",
			"email", "wechat", "phone", "ding_talk")
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
	var ctx = context.Background()
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
				logs.Error(err)
			}
			continue
		}
		//popuser
		toUsers, err := GetEventUser(event.GroupIds, event.UserIds, v)
		if err != nil {
			logs.Error(err)
			return
		}
		if len(toUsers) == 0 {
			return
		}
		userList := strings.Join(toUsers, ",")
		event.ToUsers = userList
		p, _ := json.Marshal(event)
		err = RDB.LPush(ctx, v, p).Err()
		if err != nil {
			logs.Error(err)
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
			logs.Error(err)
			return
		}
	}
}

//mut
func IsMuted(event *Event) bool {
	o := orm.NewOrm()
	var rules []Rule
	al := new(Rule)
	//cond
	cond := orm.NewCondition()
	//tenant
	cond = cond.And("tenant_id__icontains", event.TenantID)
	//mute rule
	cond = cond.And("m_type", "3")
	//enable mute rule
	cond = cond.And("status", "0")
	//屏蔽规则
	_, err := o.QueryTable(al).SetCond(cond).
		All(&rules, "id", "name", "conditions", "tenant_id", "note", "s_week",
			"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created")
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

//IsMuteTime not
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
			logs.Error(err)
			continue
		}
		result, err := expression.Evaluate(nil)
		if err != nil {
			//return true and event not send！！
			logs.Error(err)
			continue
		}
		if result.(bool) {
			count++
		}
	}
	if count == len(conds) {
		return true
	}
	return false
}

func isMuteChannel(event *Event, rule *Rule) bool {
	if !strings.Contains(rule.Channel, event.Channel) {
		return true
	}
	return false
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
			logs.Error(err)
			continue
		}
		result, err := expression.Evaluate(nil)
		if err != nil {
			//return true and event not send！！
			logs.Error(err)
			continue
		}
		if result.(bool) {
			count++
		}
	}
	if count == len(conds) {
		return true
	}
	return false
}
