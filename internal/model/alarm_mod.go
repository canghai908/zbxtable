package model

import "time"

const (
	NotifySuccess     = iota // 0
	NotifyMuted              // 1
	NotifyDefault            // 2
	NotifyNoRecipient        // 3
	NotifySuppressed         // 4: 已聚合抑制
	NotifyEscalated          // 5: 已升级通知
)

// Alarm struct
type Alarm struct {
	ID             int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ZID            int        `gorm:"column:zid;index" json:"zid"` // 实例ID（数字主键）
	HostID         string     `gorm:"column:host_id;size:255" json:"host_id"`
	Hostname       string     `gorm:"column:hostname;size:255" json:"hostname"`
	Host           string     `gorm:"column:host;size:200" json:"host"`
	HostsIP        string     `gorm:"column:host_ip;size:200" json:"host_ip"`
	TriggerID      int64      `gorm:"column:trigger_id" json:"trigger_id"`
	ItemID         int64      `gorm:"column:item_id" json:"item_id"`
	ItemName       string     `gorm:"column:item_name;type:text" json:"item_name"`
	ItemValue      string     `gorm:"column:item_value;type:text" json:"item_value"`
	Hgroup         string     `gorm:"column:hgroup;size:200" json:"hgroup"`
	OccurTime      time.Time  `gorm:"column:occurtime;type:timestamp" json:"occur_time"`
	Level          string     `gorm:"column:level;size:200" json:"level"`
	Message        string     `gorm:"column:message;type:text" json:"message"`
	Hkey           string     `gorm:"column:hkey;size:3000" json:"hkey"`
	Detail         string     `gorm:"column:detail;type:text" json:"detail"`
	EventID        int64      `gorm:"column:event_id" json:"eventid"`
	EventDuration  string     `gorm:"column:event_duration;size:50" json:"event_duration"`
	Status         string     `gorm:"column:status;size:200" json:"status"`
	NotifyStatus   string     `gorm:"column:notify_status;size:10" json:"notify_status"`
	AckStatus      int        `gorm:"column:ack_status;default:0" json:"ack_status"`
	AckUser        string     `gorm:"column:ack_user;size:100" json:"ack_user"`
	AckTime        *time.Time `gorm:"column:ack_time;type:timestamp" json:"ack_time"`
	AckNote        string     `gorm:"column:ack_note;size:500" json:"ack_note"`
	MatchedRuleID  int        `gorm:"column:matched_rule_id;default:0;index" json:"matched_rule_id"`
	EscalationStep int        `gorm:"column:escalation_step;default:0" json:"escalation_step"`
	EscalatedAt    *time.Time `gorm:"column:escalated_at;type:timestamp" json:"escalated_at"`
}

// ListQueryAlarm query
type ListQueryAlarm struct {
	Host   string   `json:"host"`
	Period []string `json:"period"`
}

// ListExportAlarm struct
type ListExportAlarm struct {
	Begin  string `json:"begin"`
	End    string `json:"end"`
	Hosts  string `json:"hosts"`
	ZID    string `json:"zid"`
	Status string `json:"status"`
	Level  string `json:"level"`
	HostIP string `json:"host_ip"`
}

// ListAnalysisAlarm qu
type ListAnalysisAlarm struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
	ZID   string `json:"zid"`
}

// Pie struct
type Pie struct {
	Value int    `json:"value"`
	Name  string `json:"name"`
}
