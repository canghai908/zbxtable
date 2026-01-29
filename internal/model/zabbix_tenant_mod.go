package models

import "time"

// ZabbixTenantBinding 将 Zabbix "租户"(TenantID) 与具体 Zabbix 连接实例、Token 绑定
// - 同一 tenant_id 只允许绑定一条记录（简单模型，满足"租户->连接+token"）
// - enabled=false 时该租户暂停接入（Receive 校验失败）
type ZabbixTenantBinding struct {
	ID               int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID         string    `gorm:"column:tenant_id;size:255;uniqueIndex;not null" json:"tenant_id"`
	ZabbixInstanceID int       `gorm:"column:zabbix_instance_id;index;not null" json:"zabbix_instance_id"`
	Token            string    `gorm:"column:token;size:2048" json:"token"`
	Enabled          bool      `gorm:"column:enabled;default:true" json:"enabled"`
	NotifyMethod     string    `gorm:"column:notify_method;size:20;default:msagent" json:"notify_method"` // 通知方式：msagent 或 webhook
	MSAgentInstalled bool      `gorm:"column:ms_agent_installed;default:false" json:"ms_agent_installed"` // MS-Agent 是否已安装
	MSAgentVersion   string    `gorm:"column:ms_agent_version;size:64" json:"ms_agent_version"`           // MS-Agent 版本
	WebhookInstalled bool      `gorm:"column:webhook_installed;default:false" json:"webhook_installed"`   // Webhook 是否已安装
	WebhookURL       string    `gorm:"column:webhook_url;size:512" json:"webhook_url"`                    // Webhook URL
	CreatedAt        time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (t *ZabbixTenantBinding) TableName() string {
	return TableName("zabbix_tenant_binding")
}
