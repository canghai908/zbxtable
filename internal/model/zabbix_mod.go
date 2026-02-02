package model

import "time"

// ZabbixInstance Zabbix 实例表
// 每个实例对应一个独立的 Zabbix 系统
type ZabbixInstance struct {
	ID           int    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`                  // 实例主键ID
	Instance     string `gorm:"column:instance;size:255;uniqueIndex;not null" json:"instance"` // 实例标识（唯一，可以是中英文数字混合）
	Name         string `gorm:"column:name;size:255;not null" json:"name"`                     // Zabbix 名称
	URL          string `gorm:"column:url;size:512;not null" json:"url"`                       // Zabbix URL
	User         string `gorm:"column:user;size:255" json:"user"`                              // Zabbix 用户名
	Pass         string `gorm:"column:pass;size:255" json:"pass"`                              // Zabbix 密码
	Token        string `gorm:"column:token;size:2048" json:"token"`                           // Zabbix API Token（优先）
	WebhookToken string `gorm:"column:webhook_token;size:2048" json:"webhook_token"`           // Webhook 认证 Token（用于 MS-Agent/Webhook）

	// 状态字段
	Enabled bool `gorm:"column:enabled;default:true" json:"enabled"` // 是否启用

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

func (t *ZabbixInstance) TableName() string {
	return TableName("zabbix")
}
