package model

import "time"

// InspectionTask 自动化巡检任务配置
type InspectionTask struct {
	ID             int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name           string     `gorm:"column:name;size:200;not null" json:"name"`
	ZIDs           string     `gorm:"column:z_ids;size:500" json:"z_ids"`
	HostTypes      string     `gorm:"column:host_types;size:500" json:"host_types"`
	GroupIDs       string     `gorm:"column:group_ids;type:text" json:"group_ids"`
	InspectMode    string     `gorm:"column:inspect_mode;size:30;default:scheduled" json:"inspect_mode"`
	Cycle          string     `gorm:"column:cycle;size:30;default:day" json:"cycle"`
	CronExpr       string     `gorm:"column:cron_expr;size:100" json:"cron_expr"`
	CpuThreshold   float64    `gorm:"column:cpu_threshold;default:80" json:"cpu_threshold"`
	MemThreshold   float64    `gorm:"column:mem_threshold;default:80" json:"mem_threshold"`
	DiskThreshold  float64    `gorm:"column:disk_threshold;default:85" json:"disk_threshold"`
	NotifyChannels string     `gorm:"column:notify_channels;size:300" json:"notify_channels"`
	NotifyUserIDs  string     `gorm:"column:notify_user_ids;size:500" json:"notify_user_ids"`
	NotifyGroupIDs string     `gorm:"column:notify_group_ids;size:500" json:"notify_group_ids"`
	NotifyEmails   string     `gorm:"column:notify_emails;size:500" json:"notify_emails"`
	Status         string     `gorm:"column:status;size:20;default:1" json:"status"`
	ExecStatus     int        `gorm:"column:exec_status;default:0" json:"exec_status"`
	LastScore      int        `gorm:"column:last_score;default:0" json:"last_score"`
	LastGrade      string     `gorm:"column:last_grade;size:40" json:"last_grade"`
	LastReportID   int        `gorm:"column:last_report_id;default:0" json:"last_report_id"`
	LastRunAt      *time.Time `gorm:"column:last_run_at;type:timestamp" json:"last_run_at"`
	LastFinishAt   *time.Time `gorm:"column:last_finish_at;type:timestamp" json:"last_finish_at"`
	Desc           string     `gorm:"column:desc;size:500" json:"desc"`
	CreatedAt      time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName 指定表名
func (t *InspectionTask) TableName() string {
	return TableName("inspection_task")
}

// InspectionReport 巡检报告与健康评分记录
type InspectionReport struct {
	ID               int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TaskID           int        `gorm:"column:task_id;index" json:"task_id"`
	TaskName         string     `gorm:"column:task_name;size:200" json:"task_name"`
	TriggerType      string     `gorm:"column:trigger_type;size:30" json:"trigger_type"`
	Status           int        `gorm:"column:status;default:0" json:"status"`
	HealthScore      int        `gorm:"column:health_score;default:0" json:"health_score"`
	HealthGrade      string     `gorm:"column:health_grade;size:40" json:"health_grade"`
	TotalInstances   int        `gorm:"column:total_instances;default:0" json:"total_instances"`
	TotalHosts       int        `gorm:"column:total_hosts;default:0" json:"total_hosts"`
	AvailableHosts   int        `gorm:"column:available_hosts;default:0" json:"available_hosts"`
	UnavailableHosts int        `gorm:"column:unavailable_hosts;default:0" json:"unavailable_hosts"`
	HighCpuHosts     int        `gorm:"column:high_cpu_hosts;default:0" json:"high_cpu_hosts"`
	HighMemHosts     int        `gorm:"column:high_mem_hosts;default:0" json:"high_mem_hosts"`
	HighDiskHosts    int        `gorm:"column:high_disk_hosts;default:0" json:"high_disk_hosts"`
	ActiveAlarms     int        `gorm:"column:active_alarms;default:0" json:"active_alarms"`
	CriticalAlarms   int        `gorm:"column:critical_alarms;default:0" json:"critical_alarms"`
	AbnormalHosts    int        `gorm:"column:abnormal_hosts;default:0" json:"abnormal_hosts"`
	SummaryJSON      string     `gorm:"column:summary_json;type:text" json:"-"`
	HostsJSON        string     `gorm:"column:hosts_json;type:text" json:"-"`
	Files            string     `gorm:"column:files;size:300" json:"files"`
	ExcelFile        string     `gorm:"column:excel_file;size:300" json:"excel_file"`
	HtmlFile         string     `gorm:"column:html_file;size:300" json:"html_file"`
	PdfFile          string     `gorm:"column:pdf_file;size:300" json:"pdf_file"`
	ErrorMsg         string     `gorm:"column:error_msg;size:500" json:"error_msg"`
	StartTime        time.Time  `gorm:"column:start_time;type:timestamp" json:"start_time"`
	EndTime          *time.Time `gorm:"column:end_time;type:timestamp" json:"end_time"`
	TotalTime        int64      `gorm:"column:total_time;default:0" json:"total_time"`
}

// TableName 指定表名
func (t *InspectionReport) TableName() string {
	return TableName("inspection_report")
}

// InspectionScoreBreakdown 四维健康度得分明细
type InspectionScoreBreakdown struct {
	AvailabilityScore  int `json:"availability_score"`
	AvailabilityMax    int `json:"availability_max"`
	AvailabilityDeduct int `json:"availability_deduct"`
	ResourceScore      int `json:"resource_score"`
	ResourceMax        int `json:"resource_max"`
	ResourceDeduct     int `json:"resource_deduct"`
	AlarmScore         int `json:"alarm_score"`
	AlarmMax           int `json:"alarm_max"`
	AlarmDeduct        int `json:"alarm_deduct"`
	StabilityScore     int `json:"stability_score"`
	StabilityMax       int `json:"stability_max"`
	StabilityDeduct    int `json:"stability_deduct"`
}

// InspectionInstanceStat 按实例维度的巡检统计
type InspectionInstanceStat struct {
	ZID              int    `json:"zid"`
	InstanceName     string `json:"instance_name"`
	TotalHosts       int    `json:"total_hosts"`
	AvailableHosts   int    `json:"available_hosts"`
	UnavailableHosts int    `json:"unavailable_hosts"`
	AbnormalHosts    int    `json:"abnormal_hosts"`
	ActiveAlarms     int    `json:"active_alarms"`
	AvgScore         int    `json:"avg_score"`
	Reachable        bool   `json:"reachable"`
	Error            string `json:"error,omitempty"`
}

// InspectionTypeStat 按资产类型维度的巡检统计
type InspectionTypeStat struct {
	TypeCode       string `json:"type_code"`
	TypeName       string `json:"type_name"`
	TotalHosts     int    `json:"total_hosts"`
	AvailableHosts int    `json:"available_hosts"`
	AbnormalHosts  int    `json:"abnormal_hosts"`
}

// InspectionAlarmItem 巡检捕获的活跃/未恢复告警明细
type InspectionAlarmItem struct {
	ZID          int    `json:"zid"`
	InstanceName string `json:"instance_name"`
	HostID       string `json:"host_id"`
	Host         string `json:"host"`
	Hostname     string `json:"hostname"`
	HostIP       string `json:"host_ip"`
	Level        string `json:"level"`
	SeverityNum  int    `json:"severity_num"`
	Message      string `json:"message"`
	ItemName     string `json:"item_name"`
	OccurTime    string `json:"occur_time"`
	EventID      string `json:"eventid"`
	Acknowledged bool   `json:"acknowledged"`
}

// InspectionHostItem 单台主机巡检明细
type InspectionHostItem struct {
	ZID            int      `json:"zid"`
	InstanceName   string   `json:"instance_name"`
	HostID         string   `json:"host_id"`
	Host           string   `json:"host"`
	Name           string   `json:"name"`
	IP             string   `json:"ip"`
	TypeCode       string   `json:"type_code"`
	TypeName       string   `json:"type_name"`
	OS             string   `json:"os"`
	Available      string   `json:"available"`
	Ping           string   `json:"ping"`
	CpuPct         float64  `json:"cpu_pct"`
	HasCpu         bool     `json:"has_cpu"`
	CpuHigh        bool     `json:"cpu_high"`
	MemPct         float64  `json:"mem_pct"`
	HasMem         bool     `json:"has_mem"`
	MemHigh        bool     `json:"mem_high"`
	DiskPct        float64  `json:"disk_pct"`
	HasDisk        bool     `json:"has_disk"`
	DiskHigh       bool     `json:"disk_high"`
	DiskMount      string   `json:"disk_mount"`
	Uptime         string   `json:"uptime"`
	ActiveAlarms   int      `json:"active_alarms"`
	CriticalAlarms int      `json:"critical_alarms"`
	HostScore      int      `json:"host_score"`
	Status         string   `json:"status"`
	Issues         []string `json:"issues"`
}

// InspectionSummary 巡检汇总结构（序列化至 SummaryJSON）
type InspectionSummary struct {
	ScoreBreakdown  InspectionScoreBreakdown `json:"score_breakdown"`
	InstanceStats   []InspectionInstanceStat `json:"instance_stats"`
	TypeStats       []InspectionTypeStat     `json:"type_stats"`
	Recommendations []string                 `json:"recommendations"`
	TopAlarms       []InspectionAlarmItem    `json:"top_alarms"`
	CpuThreshold    float64                  `json:"cpu_threshold"`
	MemThreshold    float64                  `json:"mem_threshold"`
	DiskThreshold   float64                  `json:"disk_threshold"`
}

// InspectionReportDetail 完整巡检报告详情
type InspectionReportDetail struct {
	InspectionReport
	Summary InspectionSummary    `json:"summary"`
	Hosts   []InspectionHostItem `json:"hosts"`
}

// InspectionOverview 巡检中心概览统计
type InspectionOverview struct {
	TotalTasks     int64              `json:"total_tasks"`
	EnabledTasks   int64              `json:"enabled_tasks"`
	TotalReports   int64              `json:"total_reports"`
	LatestAvgScore int                `json:"latest_avg_score"`
	LatestGrade    string             `json:"latest_grade"`
	GradeCounts    map[string]int     `json:"grade_counts"`
	RecentReports  []InspectionReport `json:"recent_reports"`
}
