package model

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// AssetTypeField 设备类型列表字段配置
type AssetTypeField struct {
	Key      string `json:"key"`      // Hosts struct 对应的 JSON 字段名
	Label    string `json:"label"`    // 列头显示名称
	Render   string `json:"render"`   // 渲染方式: text/progress/status/ping/tag/link
	Width    int    `json:"width"`    // 列宽（0 表示自适应）
	Visible  bool   `json:"visible"`  // 是否在列表中显示
	Sortable bool   `json:"sortable"` // 是否可排序
	Order    int    `json:"order"`    // 列顺序
}

// AssetType 资产类型定义
type AssetType struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name        string    `gorm:"column:name;size:100;not null" json:"name"`
	TypeCode    string    `gorm:"column:type_code;size:50;not null;uniqueIndex" json:"type_code"`
	Icon        string    `gorm:"column:icon;size:100;default:''" json:"icon"`
	Description string    `gorm:"column:description;size:500;default:''" json:"description"`
	MonitorType string    `gorm:"column:monitor_type;size:20;default:'agent'" json:"monitor_type"`
	SortOrder   int       `gorm:"column:sort_order;default:0" json:"sort_order"`
	// MenuGroup 指定该类型挂载的菜单分组：host/net/server/custom（空表示不在菜单显示）
	MenuGroup   string    `gorm:"column:menu_group;size:20;default:''" json:"menu_group"`
	// ListFields 列表字段配置，存储为 JSON 数组（TEXT 列在 MySQL 不能有默认值）
	ListFields  string    `gorm:"column:list_fields;type:text" json:"list_fields"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// ParseListFields 解析 ListFields JSON 字符串为字段配置列表
func (a *AssetType) ParseListFields() []AssetTypeField {
	if a.ListFields == "" {
		return nil
	}
	var fields []AssetTypeField
	if err := json.Unmarshal([]byte(a.ListFields), &fields); err != nil {
		return nil
	}
	return fields
}

// SetListFields 将字段配置列表序列化存入 ListFields
func (a *AssetType) SetListFields(fields []AssetTypeField) error {
	b, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	a.ListFields = string(b)
	return nil
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
		"menu_group":   m.MenuGroup,
		"updated_at":   time.Now(),
	}).Error
}

// GetAssetTypeFields 获取指定类型的列表字段配置
func GetAssetTypeFields(id int64) ([]AssetTypeField, error) {
	at, err := GetAssetTypeByID(id)
	if err != nil {
		return nil, err
	}
	fields := at.ParseListFields()
	if fields == nil {
		fields = DefaultFieldConfig(at.TypeCode, at.MonitorType)
	}
	return fields, nil
}

// UpdateAssetTypeFields 更新指定类型的列表字段配置
func UpdateAssetTypeFields(id int64, fields []AssetTypeField) error {
	b, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	return DB.Model(&AssetType{}).Where("id = ?", id).Updates(map[string]interface{}{
		"list_fields": string(b),
		"updated_at":  time.Now(),
	}).Error
}

// DefaultFieldConfig 根据 typeCode 和 monitorType 返回默认字段配置
func DefaultFieldConfig(typeCode, monitorType string) []AssetTypeField {
	// 通用基础字段（所有类型都有）
	base := []AssetTypeField{
		{Key: "hostid", Label: "主机ID", Render: "text", Visible: false, Order: 0},
		{Key: "name", Label: "设备名称", Render: "text", Visible: true, Order: 1},
		{Key: "instance_name", Label: "所属实例", Render: "tag", Visible: true, Width: 120, Order: 2},
		{Key: "interfaces", Label: "IP地址", Render: "text", Visible: true, Order: 3},
		{Key: "available", Label: "采集状态", Render: "status", Visible: true, Width: 100, Order: 10},
		{Key: "ping", Label: "Ping状态", Render: "ping", Visible: true, Order: 11},
	}

	switch typeCode {
	case "VM_LIN":
		return append(base, []AssetTypeField{
			{Key: "os", Label: "操作系统", Render: "text", Visible: true, Order: 4},
			{Key: "uptime", Label: "运行时长", Render: "text", Visible: true, Order: 5},
			{Key: "cpu_utilization", Label: "CPU使用率", Render: "progress", Visible: true, Order: 6},
			{Key: "memory_utilization", Label: "内存使用率", Render: "progress", Visible: true, Order: 7},
		}...)
	case "VM_WIN":
		return append(base, []AssetTypeField{
			{Key: "os", Label: "操作系统", Render: "text", Visible: true, Order: 4},
			{Key: "uptime", Label: "运行时长", Render: "text", Visible: true, Order: 5},
			{Key: "cpu_utilization", Label: "CPU使用率", Render: "progress", Visible: true, Order: 6},
			{Key: "memory_utilization", Label: "内存使用率", Render: "progress", Visible: true, Order: 7},
		}...)
	case "HW_SRV":
		return append(base, []AssetTypeField{
			{Key: "model", Label: "设备型号", Render: "text", Visible: true, Order: 4},
			{Key: "serial_no", Label: "序列号", Render: "text", Visible: true, Order: 5},
			{Key: "os", Label: "操作系统", Render: "text", Visible: true, Order: 6},
			{Key: "location", Label: "设备位置", Render: "text", Visible: true, Order: 7},
			{Key: "cpu_utilization", Label: "CPU使用率", Render: "progress", Visible: true, Order: 8},
			{Key: "memory_utilization", Label: "内存使用率", Render: "progress", Visible: true, Order: 9},
		}...)
	default:
		// SNMP/IPMI 硬件类型通用默认
		if monitorType == "snmp" || monitorType == "ipmi" {
			return append(base, []AssetTypeField{
				{Key: "model", Label: "设备型号", Render: "text", Visible: true, Order: 4},
				{Key: "serial_no", Label: "序列号", Render: "text", Visible: true, Order: 5},
				{Key: "location", Label: "设备位置", Render: "text", Visible: true, Order: 6},
				{Key: "vendor", Label: "厂商", Render: "text", Visible: true, Order: 7},
			}...)
		}
		// Agent 类型通用默认
		return append(base, []AssetTypeField{
			{Key: "os", Label: "操作系统", Render: "text", Visible: true, Order: 4},
			{Key: "uptime", Label: "运行时长", Render: "text", Visible: true, Order: 5},
			{Key: "cpu_utilization", Label: "CPU使用率", Render: "progress", Visible: true, Order: 6},
			{Key: "memory_utilization", Label: "内存使用率", Render: "progress", Visible: true, Order: 7},
		}...)
	}
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

// InitDefaultAssetTypes 幂等预置默认资产类型
func InitDefaultAssetTypes() error {
	defaults := []AssetType{
		{Name: "Linux", TypeCode: "VM_LIN", Icon: "desktop", MonitorType: "agent", SortOrder: 1, MenuGroup: "host"},
		{Name: "Windows", TypeCode: "VM_WIN", Icon: "windows", MonitorType: "agent", SortOrder: 2, MenuGroup: "host"},
		{Name: "网络设备", TypeCode: "HW_NET", Icon: "cluster", MonitorType: "snmp", SortOrder: 3, MenuGroup: "net"},
		{Name: "物理服务器", TypeCode: "HW_SRV", Icon: "database", MonitorType: "snmp", SortOrder: 4, MenuGroup: "server"},
		{Name: "光纤交换机", TypeCode: "HW_FIB", Icon: "share-alt", MonitorType: "snmp", SortOrder: 5, MenuGroup: "server"},
		{Name: "存储设备", TypeCode: "HW_STO", Icon: "hdd", MonitorType: "snmp", SortOrder: 6, MenuGroup: "server"},
	}
	for _, at := range defaults {
		var existing AssetType
		err := DB.Where("type_code = ?", at.TypeCode).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if createErr := DB.Create(&at).Error; createErr != nil {
				return createErr
			}
		} else if err == nil && existing.MenuGroup == "" {
			// 回填历史记录的 menu_group
			DB.Model(&AssetType{}).Where("type_code = ?", at.TypeCode).
				Update("menu_group", at.MenuGroup)
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
