package model

import (
	"strconv"
	"time"
	"zbxtable/pkg/logger"
)

// TableName Egress
func (t *Egress) TableName() string {
	return TableName("egress")
}

// EgressConfig 新的出口配置表（支持多个出口）
type Egress struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ZID       int       `gorm:"column:zid" json:"zid"`                                   // 实例ID
	Name      string    `gorm:"column:name;size:255;not null" json:"name"`               // 出口名称
	HostID    string    `gorm:"column:host_id;size:100;not null" json:"host_id"`         // 主机ID
	InItemID  string    `gorm:"column:in_item_id;size:100;not null" json:"in_item_id"`   // 入流量指标ID
	OutItemID string    `gorm:"column:out_item_id;size:100;not null" json:"out_item_id"` // 出流量指标ID
	Status    int       `gorm:"column:status;default:1" json:"status"`                   // 状态：1启用，0禁用
	SortOrder int       `gorm:"column:sort_order;default:0" json:"sort_order"`           // 排序
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// GetAllEgressConfigs 获取所有出口配置
func GetAllEgress() ([]Egress, error) {
	var configs []Egress
	err := DB.Where("status = ?", 1).Order("sort_order ASC, id ASC").Find(&configs).Error
	if err != nil {
		return nil, err
	}
	return configs, nil
}

// GetEgressConfigByID 根据ID获取出口配置
func GetEgressConfigByID(id int64) (*Egress, error) {
	var config Egress
	err := DB.Where("id = ?", id).First(&config).Error
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// AddEgressConfig 添加出口配置
func AddEgress(engress *Egress) error {
	return DB.Create(engress).Error
}

// UpdateEgressConfig 更新出口配置
func UpdateEgress(engress *Egress) error {
	return DB.Model(&Egress{}).Where("id = ?", engress.ID).Updates(map[string]interface{}{
		"name":        engress.Name,
		"zid":         engress.ZID,
		"host_id":     engress.HostID,
		"in_item_id":  engress.InItemID,
		"out_item_id": engress.OutItemID,
		"status":      engress.Status,
		"sort_order":  engress.SortOrder,
	}).Error
}

// DeleteEgressConfig 删除出口配置
func DeleteEgressConfig(id int64) error {
	return DB.Delete(&Egress{}, id).Error
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

// CollectEgressData 采集出口数据并存储到缓存
func CollectEgressData() error {
	configs, err := GetAllEgress()
	if err != nil {
		logger.Log.Error("获取出口配置失败:", err)
		return err
	}
	var egressDataList []map[string]interface{}
	for _, config := range configs {
		zidStr := strconv.Itoa(config.ZID)
		// 获取实例
		inst, err := GetZabbixInstanceAPI(zidStr)
		if err != nil {
			logger.Log.Errorf("获取实例失败 (zid=%s): %v", config.ZID, err)
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
			"zid":         config.ZID,
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
