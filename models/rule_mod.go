package models

import (
	"time"
)

type Rule struct {
	ID         int       `orm:"column(id);auto" json:"id"`
	Name       string    `orm:"column(name);size(200);null" json:"name"`
	TenantID   string    `orm:"column(tenant_id);size(100);null" json:"tenant_id"`
	Conditions string    `orm:"column(conditions);type(text);null" json:"conditions"`
	Sweek      string    `orm:"column(s_week);size(200);null" json:"s_week"`
	Stime      string    `orm:"column(s_time);size(200);null" json:"s_time"`
	Etime      string    `orm:"column(e_time);size(200);null" json:"e_time"`
	Channel    string    `orm:"column(channel);size(200);null" json:"channel"`
	UserIds    string    `orm:"column(user_ids);size(200);null" json:"user_ids"`
	GroupIds   string    `orm:"column(group_ids);size(200);null" json:"group_ids"`
	Note       string    `orm:"column(note);size(200);null" json:"note"`
	MType      string    `orm:"column(m_type);size(200);null" json:"m_type"` // 1告警分发 2.默认规则 3.屏蔽规则  4
	Status     string    `orm:"column(status);size(40);null" json:"status"`
	Created    time.Time `orm:"column(created);type(datetime);null;auto_now_add;"  json:"created"`
	Updated    time.Time `orm:"column(updated);type(datetime);null;auto_now"  json:"updated"`
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

//TableName alarm
func (t *Rule) TableName() string {
	return TableName("rule")
}
