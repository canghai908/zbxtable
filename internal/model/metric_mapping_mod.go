package model

import "time"

// MetricMapping 指标映射配置
type MetricMapping struct {
	ID              int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	InstanceID      int        `gorm:"column:instance_id;not null" json:"instance_id"`
	SystemType      string     `gorm:"column:system_type;size:50;not null" json:"system_type"`
	HostGroupIDs    string     `gorm:"column:host_group_ids;type:text" json:"host_group_ids"`
	MetricConfig    string     `gorm:"column:metric_config;type:json" json:"metric_config"`
	
	// 自动化配置
	AutoInit        int        `gorm:"column:auto_init;default:0" json:"auto_init"`
	InitCron        string     `gorm:"column:init_cron;size:50;default:'0 0 2 * * *'" json:"init_cron"`
	InitOnNewHost   int        `gorm:"column:init_on_new_host;default:0" json:"init_on_new_host"`
	
	// 执行状态
	Status          int        `gorm:"column:status;default:0" json:"status"`
	LastInitAt      *time.Time `gorm:"column:last_init_at" json:"last_init_at"`
	LastSuccessAt   *time.Time `gorm:"column:last_success_at" json:"last_success_at"`
	InitError       string     `gorm:"column:init_error;type:text" json:"init_error"`
	RetryCount      int        `gorm:"column:retry_count;default:0" json:"retry_count"`
	MaxRetry        int        `gorm:"column:max_retry;default:3" json:"max_retry"`
	
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
	CreatedBy       string     `gorm:"column:created_by;size:100" json:"created_by"`
}

// MetricConfig 指标配置结构（JSON）
type MetricConfig struct {
	HostType        string            `json:"host_type"`        // VM_LIN/VM_WIN/HW_NET/HW_SRV
	Metrics         map[string]string `json:"metrics"`          // 指标映射：key=字段名，value=itemid列表
	PingTemplateID  string            `json:"ping_template_id"` // ICMP模板ID
}

// MetricMappingHistory 映射执行历史
type MetricMappingHistory struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MappingID      int64      `gorm:"column:mapping_id;not null" json:"mapping_id"`
	InstanceID     int        `gorm:"column:instance_id;not null" json:"instance_id"`
	SystemType     string     `gorm:"column:system_type;size:50;not null" json:"system_type"`
	
	ExecType       string     `gorm:"column:exec_type;size:20;not null" json:"exec_type"`
	StartTime      time.Time  `gorm:"column:start_time;not null" json:"start_time"`
	EndTime        *time.Time `gorm:"column:end_time" json:"end_time"`
	Duration       int        `gorm:"column:duration" json:"duration"`
	
	Status         string     `gorm:"column:status;size:20;not null" json:"status"`
	AffectedHosts  int        `gorm:"column:affected_hosts;default:0" json:"affected_hosts"`
	ErrorMessage   string     `gorm:"column:error_message;type:text" json:"error_message"`
	DetailLog      string     `gorm:"column:detail_log;type:text" json:"detail_log"`
	
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

// MetricMappingList 响应结构
type MetricMappingList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}

func (t *MetricMapping) TableName() string {
	return TableName("metric_mapping")
}

func (t *MetricMappingHistory) TableName() string {
	return TableName("metric_mapping_history")
}
