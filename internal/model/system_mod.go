package model

import "time"

// System 资产绑定配置
type System struct {
	ID                  int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ZID                 int    `gorm:"column:zid;index" json:"zid"`
	Name                string `gorm:"column:name;size:255" json:"name"`
	TypeCode            string `gorm:"column:type_code;size:50;default:''" json:"type_code"`
	GroupID             string `gorm:"column:group_id;size:255" json:"group_id"`
	CPUUtilizationID    string `gorm:"column:cpu_utilization_id;size:200" json:"cpu_utilization_id"`
	MemoryUtilizationID string `gorm:"column:memory_utilization_id;size:200" json:"memory_utilization_id"`
	MemoryUsedID        string `gorm:"column:memory_used_id;size:200" json:"memory_used_id"`
	MemoryTotalID       string `gorm:"column:memory_total_id;size:200" json:"memory_total_id"`
	UptimeID            string `gorm:"column:uptime_id;size:200" json:"uptime_id"`
	CPUCore             string `gorm:"column:cpu_core;size:200" json:"cpu_core"`
	Model               string `gorm:"column:model;size:200" json:"model"`
	PingTemplateID      string `gorm:"column:ping_template_id;size:200" json:"ping_template_id"`
	Ping                string `gorm:"column:ping;size:200" json:"ping"`
	PingLoss            string `gorm:"column:ping_loss;size:200" json:"ping_loss"`
	PingSec             string `gorm:"column:ping_sec;size:200" json:"ping_sec"`

	// 自动化配置
	AutoInit      int    `gorm:"column:auto_init;default:0" json:"auto_init"`
	InitCron      string `gorm:"column:init_cron;size:50;default:'0 0 2 * * *'" json:"init_cron"`
	InitOnNewHost int    `gorm:"column:init_on_new_host;default:0" json:"init_on_new_host"`
	MaxRetry      int    `gorm:"column:max_retry;default:3" json:"max_retry"`

	// 执行状态
	Status        int        `gorm:"column:status;default:0" json:"status"` // 0未初始化 1成功 2失败
	InitedAt      *time.Time `gorm:"column:inited_at;type:timestamp" json:"inited_at"`
	LastSuccessAt *time.Time `gorm:"column:last_success_at" json:"last_success_at"`
	InitError     string     `gorm:"column:init_error;type:text" json:"init_error"`
	RetryCount    int        `gorm:"column:retry_count;default:0" json:"retry_count"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// SystemHistory 资产绑定初始化执行历史
type SystemHistory struct {
	ID       int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	SystemID int64  `gorm:"column:system_id;not null;index" json:"system_id"`
	ZID      int    `gorm:"column:zid;not null" json:"zid"`
	TypeCode string `gorm:"column:type_code;size:50" json:"type_code"`

	ExecType  string     `gorm:"column:exec_type;size:20;not null" json:"exec_type"` // manual/auto/retry
	StartTime time.Time  `gorm:"column:start_time;not null" json:"start_time"`
	EndTime   *time.Time `gorm:"column:end_time" json:"end_time"`
	Duration  int        `gorm:"column:duration;default:0" json:"duration"`

	Status        string `gorm:"column:status;size:20;not null" json:"status"` // running/success/failed
	AffectedHosts int    `gorm:"column:affected_hosts;default:0" json:"affected_hosts"`
	ErrorMessage  string `gorm:"column:error_message;type:text" json:"error_message"`

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (t *SystemHistory) TableName() string { return TableName("system_history") }

// SystemList struct
type SystemList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}
