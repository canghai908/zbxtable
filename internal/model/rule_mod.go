package model

import (
	"time"
)

type Rule struct {
	ID         int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ZIDs       string    `gorm:"column:z_ids" json:"z_ids"` // 实例标识符（字符串，支持通配符 "*"）
	Name       string    `gorm:"column:name;size:200" json:"name"`
	Conditions string    `gorm:"column:conditions;type:text" json:"conditions"`
	Sweek      string    `gorm:"column:s_week;size:200" json:"s_week"`
	Stime      string    `gorm:"column:s_time;size:200" json:"s_time"`
	Etime      string    `gorm:"column:e_time;size:200" json:"e_time"`
	Channel    string    `gorm:"column:channel;size:200" json:"channel"` //mail 邮件,wechat 企业微信， wechat_robot 企业微信机器人
	UserIds    string    `gorm:"column:user_ids;size:200" json:"user_ids"`
	GroupIds   string    `gorm:"column:group_ids;size:200" json:"group_ids"`
	Note       string    `gorm:"column:note;size:200" json:"note"`
	MType      string    `gorm:"column:m_type;size:200" json:"m_type"` // 1告警分发 2.默认规则 3.屏蔽规则
	Status     string    `gorm:"column:status;size:40" json:"status"`
	Created    time.Time `gorm:"column:created;autoCreateTime" json:"created"`
	Updated    time.Time `gorm:"column:updated;autoUpdateTime" json:"updated"`
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
