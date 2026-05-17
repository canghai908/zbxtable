package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// AssetType 资产类型定义
type AssetType struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:name;size:100;not null" json:"name"`
	TypeCode    string    `gorm:"column:type_code;size:50;not null;uniqueIndex" json:"type_code"`
	Icon        string    `gorm:"column:icon;size:100;default:''" json:"icon"`
	Description string    `gorm:"column:description;size:500;default:''" json:"description"`
	MonitorType string    `gorm:"column:monitor_type;size:20;default:'agent'" json:"monitor_type"`
	SortOrder   int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (t *AssetType) TableName() string {
	return TableName("asset_type")
}

// GetAllAssetTypes 获取所有资产类型，按 sort_order 升序排列
func GetAllAssetTypes() ([]AssetType, error) {
	var list []AssetType
	err := DB.Order("sort_order ASC, id ASC").Find(&list).Error
	return list, err
}

// GetAssetTypeByID 按 ID 查找资产类型
func GetAssetTypeByID(id int64) (*AssetType, error) {
	var at AssetType
	err := DB.Where("id = ?", id).First(&at).Error
	if err != nil {
		return nil, err
	}
	return &at, nil
}

// GetAssetTypeByCode 按 type_code 查找资产类型
func GetAssetTypeByCode(code string) (*AssetType, error) {
	var at AssetType
	err := DB.Where("type_code = ?", code).First(&at).Error
	if err != nil {
		return nil, err
	}
	return &at, nil
}

// CreateAssetType 创建资产类型
func CreateAssetType(m *AssetType) error {
	return DB.Create(m).Error
}

// UpdateAssetType 更新资产类型
func UpdateAssetType(m *AssetType) error {
	return DB.Model(&AssetType{}).Where("id = ?", m.ID).Updates(map[string]interface{}{
		"name":         m.Name,
		"type_code":    m.TypeCode,
		"icon":         m.Icon,
		"description":  m.Description,
		"monitor_type": m.MonitorType,
		"sort_order":   m.SortOrder,
		"updated_at":   time.Now(),
	}).Error
}

// DeleteAssetType 删除资产类型，删前校验无 System 记录引用
func DeleteAssetType(id int64) error {
	at, err := GetAssetTypeByID(id)
	if err != nil {
		return err
	}
	var count int64
	DB.Model(&System{}).Where("type_code = ?", at.TypeCode).Count(&count)
	if count > 0 {
		return errors.New("该资产类型已被绑定配置引用，请先删除对应的资产绑定配置")
	}
	return DB.Delete(&AssetType{}, id).Error
}

// GetMonitorType 返回 type_code 对应的监控类型（agent/snmp/ipmi/jmx）
func GetMonitorType(typeCode string) string {
	at, err := GetAssetTypeByCode(typeCode)
	if err != nil {
		// 兜底：HW_ 前缀默认为 snmp
		if len(typeCode) >= 3 && typeCode[:3] == "HW_" {
			return "snmp"
		}
		return "agent"
	}
	if at.MonitorType == "" {
		return "agent"
	}
	return at.MonitorType
}

// IsHardwareType 判断是否为非 Agent 硬件类型（snmp/ipmi 均视为硬件）
func IsHardwareType(typeCode string) bool {
	mt := GetMonitorType(typeCode)
	return mt == "snmp" || mt == "ipmi"
}

// InitDefaultAssetTypes 幂等预置4条默认资产类型
func InitDefaultAssetTypes() error {
	defaults := []AssetType{
		{Name: "Linux", TypeCode: "VM_LIN", Icon: "desktop", MonitorType: "agent", SortOrder: 1},
		{Name: "Windows", TypeCode: "VM_WIN", Icon: "windows", MonitorType: "agent", SortOrder: 2},
		{Name: "网络设备", TypeCode: "HW_NET", Icon: "cluster", MonitorType: "snmp", SortOrder: 3},
		{Name: "物理服务器", TypeCode: "HW_SRV", Icon: "database", MonitorType: "snmp", SortOrder: 4},
		{Name: "光纤交换机", TypeCode: "HW_FIB", Icon: "share-alt", MonitorType: "snmp", SortOrder: 5},
		{Name: "存储设备", TypeCode: "HW_STO", Icon: "hdd", MonitorType: "snmp", SortOrder: 6},
	}
	for _, at := range defaults {
		var existing AssetType
		err := DB.Where("type_code = ?", at.TypeCode).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if createErr := DB.Create(&at).Error; createErr != nil {
				return createErr
			}
		}
	}
	return nil
}

// MigrateAssetTypeNames 统一存量默认资产类型名称（幂等）
func MigrateAssetTypeNames() error {
	renameMap := map[string][]string{
		"VM_LIN": {"Linux", "Linux操作系统", "Linux主机"},
		"VM_WIN": {"Windows", "Windows操作系统", "Windows主机"},
	}

	for typeCode, aliases := range renameMap {
		targetName := aliases[0]
		if err := DB.Model(&AssetType{}).
			Where("type_code = ? AND name IN ?", typeCode, aliases).
			Update("name", targetName).Error; err != nil {
			return err
		}
	}

	return nil
}

// MigrateAssetTypeMonitorType 将存量 asset_type 记录从 is_hardware 回填 monitor_type（幂等）
func MigrateAssetTypeMonitorType() error {
	// is_hardware=true → snmp，is_hardware=false → agent（仅对 monitor_type 为空的记录操作）
	DB.Exec("UPDATE " + TableName("asset_type") + " SET monitor_type = 'snmp' WHERE (monitor_type IS NULL OR monitor_type = '') AND is_hardware = 1")
	DB.Exec("UPDATE " + TableName("asset_type") + " SET monitor_type = 'agent' WHERE (monitor_type IS NULL OR monitor_type = '') AND (is_hardware = 0 OR is_hardware IS NULL)")
	return nil
}

// MigrateSystemTypeCodes 将存量 System 记录按旧 ID 映射回填 type_code（幂等）
func MigrateSystemTypeCodes() error {
	legacyMap := map[int64]string{
		1: "VM_LIN",
		2: "VM_WIN",
		3: "HW_NET",
		4: "HW_SRV",
	}
	for id, code := range legacyMap {
		DB.Model(&System{}).
			Where("id = ? AND (type_code IS NULL OR type_code = '')", id).
			Update("type_code", code)
	}
	return nil
}
