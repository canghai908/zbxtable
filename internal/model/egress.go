package model

import (
	"time"
	"zbxtable/pkg/logger"
)

// TableName alarm (保留旧表用于兼容)
func (t *Egress) TableName() string {
	return TableName("egress")
}

type Egress struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	NameOne   string    `gorm:"column:name_one;size:255" json:"name_one"`
	InOne     string    `gorm:"column:in_one;size:255" json:"in_one"`
	OutOne    string    `gorm:"column:out_one;size:255" json:"out_one"`
	NameTwo   string    `gorm:"column:name_two;size:255" json:"name_two"`
	InTwo     string    `gorm:"column:in_two;size:255" json:"in_two"`
	OutTwo    string    `gorm:"column:out_two;size:255" json:"out_two"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	Status    int       `gorm:"column:status" json:"status"`
}

// EgressConfig 新的出口配置表（支持多个出口）
type EgressConfig struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name       string    `gorm:"column:name;size:255;not null" json:"name"`                    // 出口名称
	InstanceID   string    `gorm:"column:instance_id;size:100;not null" json:"instance_id"`          // 实例ID
	HostID     string    `gorm:"column:host_id;size:100;not null" json:"host_id"`              // 主机ID
	InItemID   string    `gorm:"column:in_item_id;size:100;not null" json:"in_item_id"`        // 入流量指标ID
	OutItemID  string    `gorm:"column:out_item_id;size:100;not null" json:"out_item_id"`      // 出流量指标ID
	Status     int       `gorm:"column:status;default:1" json:"status"`                         // 状态：1启用，0禁用
	SortOrder  int       `gorm:"column:sort_order;default:0" json:"sort_order"`                 // 排序
	CreatedAt  time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

func (t *EgressConfig) TableName() string {
	return TableName("egress_config")
}

// GetAllEgressConfigs 获取所有出口配置
func GetAllEgressConfigs() ([]EgressConfig, error) {
	var configs []EgressConfig
	err := DB.Where("status = ?", 1).Order("sort_order ASC, id ASC").Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// GetEgressConfigByID 根据ID获取出口配置
func GetEgressConfigByID(id int64) (*EgressConfig, error) {
	var config EgressConfig
	err := DB.Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// AddEgressConfig 添加出口配置
func AddEgressConfig(config *EgressConfig) error {
	return DB.Create(config).Error
}

// UpdateEgressConfig 更新出口配置
func UpdateEgressConfig(config *EgressConfig) error {
	return DB.Model(&EgressConfig{}).Where("id = ?", config.ID).Updates(map[string]interface{}{
		"name":        config.Name,
		"instance_id": config.InstanceID,
		"host_id":     config.HostID,
		"in_item_id":  config.InItemID,
		"out_item_id": config.OutItemID,
		"status":      config.Status,
		"sort_order":  config.SortOrder,
	}).Error
}

// DeleteEgressConfig 删除出口配置
func DeleteEgressConfig(id int64) error {
	return DB.Delete(&EgressConfig{}, id).Error
}

// 旧的兼容函数
// get id
func GetEgress() (v *Egress, err error) {
	v = &Egress{}
	err = DB.Where("id = ?", 1).First(v).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

// get all
func UpdateEgress(m *Egress) (err error) {
	var v Egress
	err = DB.Where("id = ?", m.ID).First(&v).Error
	if err != nil {
		return err
	}
	m.CreatedAt = v.CreatedAt
	m.Status = 1
	err = DB.Model(&Egress{}).Where("id = ?", m.ID).Updates(map[string]interface{}{
		"name_one":   m.NameOne,
		"in_one":     m.InOne,
		"out_one":    m.OutOne,
		"name_two":   m.NameTwo,
		"in_two":     m.InTwo,
		"out_two":    m.OutTwo,
		"created_at": m.CreatedAt,
		"status":     m.Status,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

// CollectEgressData 采集出口数据并存储到缓存
func CollectEgressData() error {
	configs, err := GetAllEgressConfigs()
	if err != nil {
		logger.Log.Error("获取出口配置失败:", err)
		return err
	}

	var egressDataList []map[string]interface{}

	for _, config := range configs {
		// 获取实例
		inst, err := GetZabbixInstanceAPI(config.InstanceID)
		if err != nil {
			logger.Log.Errorf("获取实例失败 (instance_id=%s): %v", config.InstanceID, err)
			continue
		}

		// 获取入流量最新值
		inHistory, err := GetLastHistoryByItemIDFromInstance(inst, config.InItemID)
		if err != nil {
			logger.Log.Errorf("获取入流量数据失败 (item_id=%s): %v", config.InItemID, err)
			continue
		}

		// 获取出流量最新值
		outHistory, err := GetLastHistoryByItemIDFromInstance(inst, config.OutItemID)
		if err != nil {
			logger.Log.Errorf("获取出流量数据失败 (item_id=%s): %v", config.OutItemID, err)
			continue
		}

		egressData := map[string]interface{}{
			"id":          config.ID,
			"name":        config.Name,
			"instance_id": config.InstanceID,
			"in_value":    inHistory.Value,
			"out_value":   outHistory.Value,
			"in_item_id":  config.InItemID,
			"out_item_id": config.OutItemID,
			"timestamp":   time.Now().Unix(),
		}
		egressDataList = append(egressDataList, egressData)
	}

	// 存储到缓存
	if len(egressDataList) > 0 {
		data, err := json.Marshal(egressDataList)
		if err != nil {
			logger.Log.Error("序列化出口数据失败:", err)
			return err
		}
		err = CacheSet("egress_data", string(data), 0)
		if err != nil {
			logger.Log.Error("存储出口数据到缓存失败:", err)
			return err
		}
		logger.Log.Info("出口数据采集成功，共", len(egressDataList), "个出口")
	}

	return nil
}

// GetLastHistoryByItemIDFromInstance 获取指定指标的最新历史数据
func GetLastHistoryByItemIDFromInstance(inst *APIInstance, itemID string) (*History, error) {
	// 获取指标信息以确定值类型
	itemInfo, err := GetItemByIDFromInstance(inst, itemID)
	if err != nil || len(itemInfo) == 0 {
		return nil, err
	}

	valueType := itemInfo[0].ValueType
	now := time.Now().Unix()
	start := now - 300 // 最近5分钟

	histories, err := GetHistoryByItemIDFromInstance(inst, itemID, valueType, start, now)
	if err != nil {
		return nil, err
	}

	if len(histories) == 0 {
		return &History{Value: "0"}, nil
	}

	// 返回最新的一条数据
	return &histories[len(histories)-1], nil
}
