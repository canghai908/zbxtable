package model

import "time"

// system
type System struct {
	ID                  int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	InstanceID          int       `gorm:"column:instance_id;index" json:"instance_id"`           // 实例ID
	Name                string    `gorm:"column:name;size:255" json:"name"`
	GroupID             string    `gorm:"column:group_id;size:255" json:"group_id"`
	CPUUtilizationID    string    `gorm:"column:cpu_utilization_id;size:200" json:"cpu_utilization_id"`
	MemoryUtilizationID string    `gorm:"column:memory_utilization_id;size:200" json:"memory_utilization_id"`
	MemoryUsedID        string    `gorm:"column:memory_used_id;size:200" json:"memory_used_id"`
	MemoryTotalID       string    `gorm:"column:memory_total_id;size:200" json:"memory_total_id"`
	UptimeID            string    `gorm:"column:uptime_id;size:200" json:"uptime_id"`
	CPUCore             string    `gorm:"column:cpu_core;size:200" json:"cpu_core"`
	Model               string    `gorm:"column:model;size:200" json:"model"`
	PingTemplateID      string    `gorm:"column:ping_template_id;size:200" json:"ping_template_id"`
	Ping                string    `gorm:"column:ping;size:200" json:"ping"`
	PingLoss            string    `gorm:"column:ping_loss;size:200" json:"ping_loss"`
	PingSec             string    `gorm:"column:ping_sec;size:200" json:"ping_sec"`
	InitedAt            *time.Time `gorm:"column:inited_at;type:datetime" json:"inited_at"`
	CreatedAt           time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	Status              int       `gorm:"column:status" json:"status"`
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
