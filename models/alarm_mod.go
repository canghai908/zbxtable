package models

import "time"

const (
	NotifySuccess = iota // 0
	NotifyMuted          // 1
	NotifyDefault        // 2
)

//Alarm struct
type Alarm struct {
	ID int `orm:"column(id);auto" json:"id"`
	//v2 add begin
	TenantID      string    `orm:"column(tenant_id);size(255)" json:"tenant_id"`
	HostID        string    `orm:"column(host_id);size(255)" json:"host_id"`
	Hostname      string    `orm:"column(hostname);size(255)" json:"hostname"`
	Host          string    `orm:"column(host);size(200)" json:"host"`
	HostsIP       string    `orm:"column(host_ip);size(200)" json:"host_ip"`
	TriggerID     int64     `orm:"column(trigger_id);size(200)" json:"trigger_id"`
	ItemID        int64     `orm:"column(item_id);size(200)" json:"item_id"`
	ItemName      string    `orm:"column(item_name);type(text)" json:"item_name"`
	ItemValue     string    `orm:"column(item_value);type(text)" json:"item_value"`
	Hgroup        string    `orm:"column(hgroup);size(200)" json:"hgroup"`
	OccurTime     time.Time `orm:"column(occurtime);type(datetime)" json:"occur_time"`
	Level         string    `orm:"column(level);size(200)" json:"level"`
	Message       string    `orm:"column(message);type(text)" json:"message"`
	Hkey          string    `orm:"column(hkey);size(3000)" json:"hkey"`
	Detail        string    `orm:"column(detail);type(text)" json:"detail"`
	EventID       int64     `orm:"column(event_id);size(200)" json:"eventid"`
	EventDuration string    `orm:"column(event_duration);size(50)" json:"event_duration"`
	Status        string    `orm:"column(status);size(200)" json:"status"`
	NotifyStatus  string    `orm:"column(notify_status);size(10)" json:"notify_status"`
}

//ListQueryAlarm query
type ListQueryAlarm struct {
	Host   string   `json:"host"`
	Period []string `json:"period"`
}

//ListExportAlarm struct
type ListExportAlarm struct {
	Begin    string `json:"begin"`
	End      string `json:"end"`
	Hosts    string `json:"hosts"`
	TenantID string `json:"tenant_id"`
	Status   string `json:"status"`
	Level    string `json:"level"`
}

//ListAnalysisAlarm qu
type ListAnalysisAlarm struct {
	Begin    string `json:"begin"`
	End      string `json:"end"`
	TenantID string `json:"tenant_id"`
}

//Pie struct
type Pie struct {
	Value int    `json:"value"`
	Name  string `json:"name"`
}

//AlarmList struct
type AlarmList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []Alarm `json:"items"`
		Total int64   `json:"total"`
	} `json:"data"`
}

//AlarmList struct
type AlarmTendantList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}

//AnalysisList struct
type AnalysisList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Level      []string `json:"level"`
		LevelCount []Pie    `json:"level_count"`
		Host       []string `json:"host"`
		HostCount  []int    `json:"host_count"`
	} `json:"data"`
}
