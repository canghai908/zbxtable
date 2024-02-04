package models

import "time"

// system
type System struct {
	ID                  int64     `orm:"column(id);auto" json:"id"`
	Name                string    `orm:"column(name);size(255)" json:"name"`
	GroupID             string    `orm:"column(group_id);size(255)" json:"group_id"`
	CPUUtilizationID    string    `orm:"column(cpu_utilization_id);size(200)" json:"cpu_utilization_id"`
	MemoryUtilizationID string    `orm:"column(memory_utilization_id);size(200)" json:"memory_utilization_id"`
	MemoryUsedID        string    `orm:"column(memory_used_id);size(200)" json:"memory_used_id"`
	MemoryTotalID       string    `orm:"column(memory_total_id);size(200)" json:"memory_total_id"`
	UptimeID            string    `orm:"column(uptime_id);size(200)" json:"uptime_id"`
	CPUCore             string    `orm:"column(cpu_core);size(200)" json:"cpu_core"`
	Model               string    `orm:"model(cpu_core);size(200)" json:"model"`
	PingTemplateID      string    `orm:"column(ping_template_id);size(200)" json:"ping_template_id"`
	Ping                string    `orm:"column(ping);size(200)" json:"ping"`
	PingLoss            string    `orm:"column(ping_loss);size(200)" json:"ping_loss"`
	PingSec             string    `orm:"column(ping_sec);size(200)" json:"ping_sec"`
	InitedAt            time.Time `orm:"column(inited_at);type(datetime);null" json:"inited_at"`
	CreatedAt           time.Time `orm:"column(created_at);type(datetime);auto_now_add" json:"created_at"`
	UpdatedAt           time.Time `orm:"column(updated_at);type(datetime);auto_now" json:"updated_at"`
	Status              int       `orm:"column(status);size(200)" json:"status"`
}

// SystemList struct
type SystemList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}
