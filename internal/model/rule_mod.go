package model

import (
	"time"
)

type Rule struct {
	ScheduleIDs        []int     `gorm:"column:schedule_ids;serializer:json;type:text" json:"schedule_ids"`
	ID                 int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ZIDs               string    `gorm:"column:z_ids;size:200" json:"z_ids"` // 实例标识符（字符串，支持通配符 "*"，多个用逗号分隔）
	Name               string    `gorm:"column:name;size:200" json:"name"`
	Conditions         string    `gorm:"column:conditions;type:text" json:"conditions"`
	Sweek              string    `gorm:"column:s_week;size:200" json:"s_week"`
	Stime              string    `gorm:"column:s_time;size:200" json:"s_time"`
	Etime              string    `gorm:"column:e_time;size:200" json:"e_time"`
	Channel            string    `gorm:"column:channel;size:200" json:"channel"` //mail 邮件,wechat 企业微信， wechat_robot 企业微信机器人
	UserIds            string    `gorm:"column:user_ids;size:200" json:"user_ids"`
	GroupIds           string    `gorm:"column:group_ids;size:200" json:"group_ids"`
	Note               string    `gorm:"column:note;size:200" json:"note"`
	MType              string    `gorm:"column:m_type;size:200" json:"m_type"` // 1告警分发 2.默认规则 3.屏蔽规则
	Status             string    `gorm:"column:status;size:40" json:"status"`
	SuppressWindow     int       `gorm:"column:suppress_window;default:0" json:"suppress_window"`          // 聚合抑制时间窗口（秒），<=0 表示不抑制
	SuppressThreshold  int       `gorm:"column:suppress_threshold;default:1" json:"suppress_threshold"`    // 窗口内立即放行条数阈值
	EscalationMinutes  int       `gorm:"column:escalation_minutes;default:0" json:"escalation_minutes"`    // 超时未认领自动升级分钟数，<=0 表示不升级
	EscalationUserIDs  string    `gorm:"column:escalation_user_ids;size:200" json:"escalation_user_ids"`   // 升级接收用户IDs
	EscalationGroupIDs string    `gorm:"column:escalation_group_ids;size:200" json:"escalation_group_ids"` // 升级接收用户组IDs
	EscalationChannel  string    `gorm:"column:escalation_channel;size:200" json:"escalation_channel"`     // 升级通知渠道
	Created            time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	Updated            time.Time `gorm:"column:updated;autoUpdateTime" json:"updated"`
}
type Conditions struct {
	RType  string `orm:"column(r_type);size(200);null" json:"r_type"`
	RFunc  string `orm:"column(r_func);size(200);null" json:"r_func"`
	Rvalue string `orm:"column(r_value);size(200);null" json:"r_value"`
}
type RuleResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}

// TableName alarm
func (t *Rule) TableName() string {
	return TableName("rule")
}
