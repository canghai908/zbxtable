package model

import (
	"errors"
	"time"
)

// TableName alarm
func (t *Config) TableName() string {
	return TableName("config")
}

type Config struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;size:255" json:"name"`
	Key       string    `gorm:"column:key;size:255" json:"key"`
	Value     string    `gorm:"column:value;size:255" json:"value"`
	Comment   string    `gorm:"column:comment;size:255" json:"comment"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// GetConfigList 获取系统配置
func GetConfigList() ([]Config, error) {
	var v []Config
	err := DB.Order("id").Find(&v).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

// GetConfigOne 获取某一个配置
func GetConfigOne(id int64) (v *Config, err error) {
	v = &Config{}
	err = DB.Where("id = ?", id).First(v).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

// UpdateConfig 更新系统配置
func UpdateConfig(m *Config) (err error) {
	var v Config
	//数据面板独立配置
	err = DB.Where("id = ?", m.ID).First(&v).Error
	if err != nil {
		return err
	}
	
	// 禁止修改加密密钥
	if v.Key == "encryption_key" {
		return errors.New("加密密钥不允许修改，如需更换请联系系统管理员")
	}
	
	err = DB.Model(&Config{}).Where("id = ?", m.ID).Update("value", m.Value).Error
	if err != nil {
		return err
	}
	//dash处理
	if m.ID == 1 {
		err = updateZbxDash(m)
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}

// 数据面板控制
func updateZbxDash(m *Config) (err error) {
	var menu Menu
	err = DB.Where("router = ?", "dash").First(&menu).Error
	if err != nil {
		return err
	}
	switch m.Value {
	case "0":
		menu.IsAvailable = true
	case "1":
		menu.IsAvailable = false
	}
	err = DB.Model(&Menu{}).Where("router = ?", "dash").Update("is_available", menu.IsAvailable).Error
	if err != nil {
		return err
	}
	return nil
}

// GetConfigValueByKey 根据 key 获取配置值，如果不存在或出错则返回默认值
func GetConfigValueByKey(key string, defaultVal string) string {
	var c Config
	err := DB.Where("`key` = ?", key).First(&c).Error
	if err == nil && c.Value != "" {
		return c.Value
	}
	return defaultVal
}
