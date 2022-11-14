package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

const (
	EventSuccess = iota // 0
	EventFailed         // 1
	//Notify               // 2
)

//AlarmList struct
type EventLogRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []EventLog `json:"items"`
		Total int64      `json:"total"`
	} `json:"data"`
}

type EventLog struct {
	ID            int       `orm:"column(id);auto" json:"id"`
	AlarmID       int64     `orm:"column(alarm_id);size(100);null" json:"alarm_id"`
	EventID       int64     `orm:"column(event_id);size(100);null" json:"eventid"`
	Rule          string    `orm:"column(rule);size(100);null" json:"rule"`
	Channel       string    `orm:"column(channel);size(100);null" json:"channel"`
	User          string    `orm:"column(user);size(100);null" json:"user"`
	Account       string    `orm:"column(account);size(100);null" json:"account"`
	NotifyTime    time.Time `orm:"column(notify_time);type(datetime);null" json:"notify_time"`
	NotifyContent string    `orm:"column(notify_content);type(text);null" json:"notify_content"`
	NotifyError   string    `orm:"column(notify_error);type(text);null" json:"notify_error"`
	Status        string    `orm:"column(status);size(10);null" json:"status"`
}

//TableName alarm
func (t *EventLog) TableName() string {
	return TableName("event_log")
}

type EventTpl struct {
	HostsID       string `json:"host_id"`
	HostHost      string `json:"host_host"`
	Hostname      string `json:"hostname"`
	HostsIP       string `json:"host_ip"`
	HostGroup     string `json:"host_group"`
	EventTime     string `json:"event_time"`
	Severity      string `json:"severity"`
	TriggerID     int64  `json:"trigger_id"`
	TriggerName   string `json:"trigger_name"`
	TriggerKey    string `json:"trigger_key"`
	TriggerValue  string `json:"trigger_value"`
	ItemID        int64  `json:"item_id"`
	ItemName      string `json:"item_name"`
	ItemValue     string `json:"item_value"`
	EventID       int64  `json:"event_id"`
	EventDuration string `json:"event_duration"`
}

type Event struct {
	ID            int       `orm:"column(id);auto" json:"id"`
	TenantID      string    `orm:"column(tenant_id);size(255)" json:"tenant_id"`
	HostID        string    `orm:"column(host_id);size(255)" json:"host_id"`
	Hostname      string    `orm:"column(hostname);size(255)" json:"hostname"`
	Host          string    `orm:"column(host);size(200)" json:"host"`
	HostsIP       string    `orm:"column(host_ip);size(200)" json:"host_ip"`
	TriggerID     int64     `orm:"column(trigger_id);size(200)" json:"trigger_id"`
	ItemID        int64     `orm:"column(item_id);size(200)" json:"item_id"`
	ItemName      string    `orm:"column(item_name);size(3000)" json:"item_name"`
	ItemValue     string    `orm:"column(item_value);size(3000)" json:"item_value"`
	Hgroup        string    `orm:"column(hgroup);size(200)" json:"hgroup"`
	OccurTime     time.Time `orm:"column(occur_time);type(datetime)" json:"occur_time"`
	Level         string    `orm:"column(level);size(200)" json:"level"`
	Message       string    `orm:"column(message);size(3000)" json:"message"`
	Hkey          string    `orm:"column(hkey);size(3000)" json:"hkey"`
	Detail        string    `orm:"column(detail);size(3000)" json:"detail"`
	Status        string    `orm:"column(status);size(200)" json:"status"`
	EventID       int64     `orm:"column(event_id);size(200)" json:"eventid"`
	EventDuration string    `orm:"column(event_duration);size(50)" json:"event_duration"`
	Rule          string    `orm:"column(rule);size(200)" json:"rule"`
	RuleType      string    `orm:"column(rule_type);size(10)" json:"rule_type"`
	Channel       string    `orm:"column(channel);size(255)" json:"channel"`
	ToUsers       string    `orm:"column(to_users);size(255)" json:"to_users"`
	UserIds       string    `orm:"column(user_ids);size(255)" json:"user_ids"`
	GroupIds      string    `orm:"column(group_ids);size(255)" json:"group_ids"`
}

// AddAlarm insert a new Alarm into database and returns
// last inserted Id on success.
func AddEventLog(m *EventLog) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetAlarmByID retrieves Alarm by Id. Returns error if
// Id doesn't exist
func GetEventLogByAlarmID(id int) (v []EventLog, err error) {
	o := orm.NewOrm()
	var eventLog EventLog
	var thisEvents []EventLog
	_, err = o.QueryTable(eventLog).Filter("alarm_id", id).
		//OrderBy("-occurtime").
		All(&thisEvents)
	if err != nil {
		return []EventLog{}, err
	}
	return thisEvents, nil
}
