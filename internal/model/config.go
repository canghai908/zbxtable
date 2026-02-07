package model

import (
	"errors"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"
)

// TableName alarm
func (t *Config) TableName() string {
	return TableName("config")
}

type Config struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"column:name;size:255" json:"name"`
	Key       string    `gorm:"column:key;size:255" json:"key"`
	Value     string    `gorm:"column:value;type:text" json:"value"`
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

	// 对敏感字段进行加密
	valueToSave := m.Value
	sensitiveKeys := []string{
		"email_secret",     // SMTP 密码/授权码
		"wechat_secret",    // 企业微信 Secret
		"deepseek_api_key", // Deepseek API Key
	}

	// 检查是否是敏感字段
	isSensitive := false
	for _, key := range sensitiveKeys {
		if v.Key == key {
			isSensitive = true
			break
		}
	}

	// 如果是敏感字段且值不为空，进行加密
	if isSensitive && valueToSave != "" {
		encryptionKey := GetEncryptionKey()
		encryptedValue, err := utils.EncryptString(valueToSave, encryptionKey)
		if err != nil {
			return errors.New("加密失败: " + err.Error())
		}
		valueToSave = encryptedValue
	}

	err = DB.Model(&Config{}).Where("id = ?", m.ID).Update("value", valueToSave).Error
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
// 对于敏感字段会自动解密
func GetConfigValueByKey(key string, defaultVal string) string {
	// 如果数据库未初始化，直接返回默认值
	if DB == nil {
		return defaultVal
	}

	var c Config
	err := DB.Where("`key` = ?", key).First(&c).Error
	if err != nil || c.Value == "" {
		return defaultVal
	}

	// 定义需要解密的敏感字段
	sensitiveKeys := []string{
		"email_secret",     // SMTP 密码/授权码
		"wechat_secret",    // 企业微信 Secret
		"deepseek_api_key", // Deepseek API Key
	}

	// 检查是否是敏感字段
	isSensitive := false
	for _, skey := range sensitiveKeys {
		if key == skey {
			isSensitive = true
			break
		}
	}

	// 如果是敏感字段，尝试解密
	if isSensitive && c.Value != "" {
		encryptionKey := GetEncryptionKey()
		decryptedValue, err := utils.DecryptString(c.Value, encryptionKey)
		if err != nil {
			// 解密失败，可能是旧数据未加密，直接返回原值
			logger.Log.Warnf("解密配置 %s 失败，返回原值: %v", key, err)
			return c.Value
		}
		return decryptedValue
	}

	return c.Value
}
