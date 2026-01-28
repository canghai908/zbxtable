package models

import "time"

// ZabbixInstance 多 Zabbix 配置
type ZabbixInstance struct {
	ID              int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Name            string     `json:"name" gorm:"type:varchar(255);not null"`
	WebURL          string     `json:"web_url" gorm:"type:varchar(512);not null"` // e.g. http://zabbix.example.com
	User            string     `json:"user" gorm:"type:varchar(255)"`
	Pass            string     `json:"pass" gorm:"type:varchar(255)"`
	Token           string     `json:"token" gorm:"type:varchar(512)"`
	Enabled         bool       `json:"enabled" gorm:"default:true;index"`
	IsActive        bool       `json:"is_active" gorm:"default:false;index"`
	Version         string     `json:"version" gorm:"type:varchar(64)"`
	LastTestOk      bool       `json:"last_test_ok" gorm:"default:false;index"`
	LastTestMessage string     `json:"last_test_message" gorm:"type:varchar(1024)"`
	LastTestAt      *time.Time `json:"last_test_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (t *ZabbixInstance) TableName() string {
	return TableName("zabbix_instance")
}
