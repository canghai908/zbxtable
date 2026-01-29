package models

import "time"

// ZabbixTenant 合并后的 Zabbix 租户表（原 ZabbixInstance + ZabbixTenantBinding）
// 一个租户对应一个 Zabbix 实例
type ZabbixTenant struct {
	ID          int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	TenantID    string `gorm:"column:tenant_id;size:255;uniqueIndex;not null" json:"tenant_id"` // 租户ID（唯一）
	Name        string `gorm:"column:name;size:255;not null" json:"name"`                       // Zabbix 名称
	WebURL      string `gorm:"column:web_url;size:512;not null" json:"web_url"`                 // Zabbix Web URL
	User        string `gorm:"column:user;size:255" json:"user"`                                // Zabbix 用户名
	Pass        string `gorm:"column:pass;size:255" json:"pass"`                                // Zabbix 密码
	ZabbixToken string `gorm:"column:zabbix_token;size:2048" json:"zabbix_token"`               // Zabbix API Token（优先）
	Token       string `gorm:"column:token;size:2048" json:"token"`                             // 租户认证 Token（用于 MS-Agent/Webhook）

	// 状态字段
	Enabled  bool `gorm:"column:enabled;default:true" json:"enabled"`      // 是否启用
	IsActive bool `gorm:"column:is_active;default:false" json:"is_active"` // 是否为当前激活的 Zabbix

	// 连接测试字段
	Version         string     `gorm:"column:version;size:64" json:"version"`                      // Zabbix 版本
	LastTestOk      bool       `gorm:"column:last_test_ok;default:false" json:"last_test_ok"`      // 最后一次测试是否成功
	LastTestMessage string     `gorm:"column:last_test_message;size:512" json:"last_test_message"` // 最后一次测试消息
	LastTestAt      *time.Time `gorm:"column:last_test_at" json:"last_test_at"`                    // 最后一次测试时间

	// 通知配置
	NotifyMethod     string `gorm:"column:notify_method;size:20;default:webhook" json:"notify_method"` // 通知方式：msagent 或 webhook
	MSAgentInstalled bool   `gorm:"column:ms_agent_installed;default:false" json:"ms_agent_installed"` // MS-Agent 是否已安装
	MSAgentVersion   string `gorm:"column:ms_agent_version;size:64" json:"ms_agent_version"`           // MS-Agent 版本
	WebhookInstalled bool   `gorm:"column:webhook_installed;default:false" json:"webhook_installed"`   // Webhook 是否已安装
	WebhookURL       string `gorm:"column:webhook_url;size:512" json:"webhook_url"`                    // Webhook URL

	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (t *ZabbixTenant) TableName() string {
	return TableName("zabbix")
}
