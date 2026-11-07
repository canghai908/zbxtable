package model

import (
	"errors"
	"fmt"
	"sync"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"
)

// 配置项内存缓存：config_key -> 原始 config_value（未解密）。
// 系统配置极少变更，缓存命中可避免每次读取都查库（GetConfigValueByKey 全局调用约 80 处）。
// 仅在 UpdateConfig 成功后按 key 失效，保证一致性；未命中（key 不存在）不缓存，
// 以免启动阶段尚未创建的配置被永久记成“缺失”。
var (
	configValueCache   = make(map[string]string)
	configValueCacheMu sync.RWMutex
)

func getCachedConfigRaw(key string) (string, bool) {
	configValueCacheMu.RLock()
	v, ok := configValueCache[key]
	configValueCacheMu.RUnlock()
	return v, ok
}

func setCachedConfigRaw(key, value string) {
	configValueCacheMu.Lock()
	configValueCache[key] = value
	configValueCacheMu.Unlock()
}

// invalidateConfigCache 使指定配置项缓存失效（写入后调用）
func invalidateConfigCache(key string) {
	configValueCacheMu.Lock()
	delete(configValueCache, key)
	configValueCacheMu.Unlock()
}

// TableName alarm
func (t *Config) TableName() string {
	return TableName("config")
}

type Config struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:name;size:255" json:"name"`
	ConfigKey   string    `gorm:"column:config_key;size:255" json:"config_key"`
	ConfigValue string    `gorm:"column:config_value;type:text" json:"config_value"`
	Comment     string    `gorm:"column:comment;size:255" json:"comment"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func isSensitiveConfigKey(key string) bool {
	switch key {
	case "email_secret", "wechat_secret", "deepseek_api_key", "custom_api_key":
		return true
	default:
		return false
	}
}

func encryptConfigValueIfSensitive(key, value, encryptionKey string) (string, error) {
	if !isSensitiveConfigKey(key) || value == "" {
		return value, nil
	}

	encryptedValue, err := utils.EncryptString(value, encryptionKey)
	if err != nil {
		return "", err
	}
	return encryptedValue, nil
}

func decryptConfigValueIfSensitive(key, value, encryptionKey string) (string, error) {
	if !isSensitiveConfigKey(key) || value == "" {
		return value, nil
	}

	decryptedValue, err := utils.DecryptString(value, encryptionKey)
	if err != nil {
		return "", err
	}
	return decryptedValue, nil
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
	if v.ConfigKey == "encryption_key" {
		return errors.New("加密密钥不允许修改，如需更换请联系系统管理员")
	}

	if IsTaskConfigKey(v.ConfigKey) {
		if err := ValidateTaskConfigValue(v.ConfigKey, m.ConfigValue); err != nil {
			return fmt.Errorf("计划任务配置无效: %w", err)
		}
	}

	// 对敏感字段进行加密
	valueToSave, err := encryptConfigValueIfSensitive(v.ConfigKey, m.ConfigValue, GetEncryptionKey())
	if err != nil {
		return errors.New("加密失败: " + err.Error())
	}

	err = DB.Model(&Config{}).Where("id = ?", m.ID).Update("config_value", valueToSave).Error
	if err != nil {
		return err
	}

	// 配置已变更，使缓存失效
	invalidateConfigCache(v.ConfigKey)

	if IsTaskConfigKey(v.ConfigKey) {
		ReloadTaskScheduler()
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
	switch m.ConfigValue {
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

	// 先查内存缓存；未命中则回库并缓存原始值
	raw, ok := getCachedConfigRaw(key)
	if !ok {
		var c Config
		err := DB.Where("config_key = ?", key).First(&c).Error
		if err != nil {
			// 不缓存未命中，避免启动阶段尚未创建的配置被永久记成缺失
			return defaultVal
		}
		raw = c.ConfigValue
		setCachedConfigRaw(key, raw)
	}

	if raw == "" {
		return defaultVal
	}

	// 如果是敏感字段，尝试解密
	if isSensitiveConfigKey(key) {
		decryptedValue, err := decryptConfigValueIfSensitive(key, raw, GetEncryptionKey())
		if err != nil {
			// 解密失败，可能是旧数据未加密，直接返回原值
			logger.Log.Warnf("解密配置 %s 失败，返回原值: %v", key, err)
			return raw
		}
		return decryptedValue
	}

	return raw
}
