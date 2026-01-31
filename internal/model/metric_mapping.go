package model

import (
	"fmt"
	"strings"
	"time"
	"zbxtable/pkg/logger"
)

// GetMetricConfig 解析指标配置
func (m *MetricMapping) GetMetricConfig() (*MetricConfig, error) {
	var config MetricConfig
	if m.MetricConfig == "" {
		return &config, nil
	}
	err := json.Unmarshal([]byte(m.MetricConfig), &config)
	return &config, err
}

// SetMetricConfig 设置指标配置
func (m *MetricMapping) SetMetricConfig(config *MetricConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}
	m.MetricConfig = string(data)
	return nil
}

// CreateOrUpdateMetricMapping 创建或更新映射配置
func CreateOrUpdateMetricMapping(m *MetricMapping) error {
	var existing MetricMapping
	err := DB.Where("instance_id = ? AND system_type = ?", m.InstanceID, m.SystemType).First(&existing).Error

	if err != nil {
		// 不存在，创建
		m.CreatedAt = time.Now()
		m.UpdatedAt = time.Now()
		return DB.Create(m).Error
	}

	// 存在，更新
	m.ID = existing.ID
	m.CreatedAt = existing.CreatedAt
	m.UpdatedAt = time.Now()
	return DB.Save(m).Error
}

// GetMetricMappingsByInstance 获取实例的所有映射配置
func GetMetricMappingsByInstance(instanceID int) ([]MetricMapping, error) {
	var mappings []MetricMapping
	err := DB.Where("instance_id = ?", instanceID).Find(&mappings).Error
	return mappings, err
}

// GetAllMetricMappings 获取所有映射配置
func GetAllMetricMappings() ([]MetricMapping, error) {
	var mappings []MetricMapping
	err := DB.Find(&mappings).Error
	return mappings, err
}

// GetMetricMappingByID 根据ID获取映射配置
func GetMetricMappingByID(id int64) (*MetricMapping, error) {
	var mapping MetricMapping
	err := DB.First(&mapping, id).Error
	return &mapping, err
}

// DeleteMetricMapping 删除映射配置
func DeleteMetricMapping(id int64) error {
	return DB.Delete(&MetricMapping{}, id).Error
}

// GetAutoInitMappings 获取需要自动初始化的映射配置
func GetAutoInitMappings() ([]MetricMapping, error) {
	var mappings []MetricMapping
	err := DB.Where("auto_init = ?", 1).Find(&mappings).Error
	return mappings, err
}

// GetFailedMappingsForRetry 获取需要重试的失败映射
func GetFailedMappingsForRetry() ([]MetricMapping, error) {
	var mappings []MetricMapping
	err := DB.Where("status = ? AND retry_count < max_retry", 2).Find(&mappings).Error
	return mappings, err
}

// ExecuteMetricMapping 执行指标映射
func ExecuteMetricMapping(mapping *MetricMapping, execType string) error {
	// 创建执行历史记录
	history := &MetricMappingHistory{
		MappingID:  mapping.ID,
		InstanceID: mapping.InstanceID,
		SystemType: mapping.SystemType,
		ExecType:   execType,
		StartTime:  time.Now(),
		Status:     "running",
	}
	DB.Create(history)

	// 获取实例API
	apiInstance, err := GetAPIByInstanceID(mapping.InstanceID)
	if err != nil {
		return updateMappingError(mapping, history, fmt.Errorf("获取实例API失败: %w", err))
	}

	// 解析配置
	config, err := mapping.GetMetricConfig()
	if err != nil {
		return updateMappingError(mapping, history, fmt.Errorf("解析配置失败: %w", err))
	}

	// 执行映射
	groupIDs := splitGroupIDs(mapping.HostGroupIDs)
	affectedHosts, err := executeMapping(apiInstance, config, groupIDs)
	if err != nil {
		return updateMappingError(mapping, history, err)
	}

	// 更新成功状态
	return updateMappingSuccess(mapping, history, affectedHosts)
}

// 辅助函数：更新失败状态
func updateMappingError(mapping *MetricMapping, history *MetricMappingHistory, err error) error {
	now := time.Now()
	duration := int(now.Sub(history.StartTime).Seconds())

	// 更新历史记录
	DB.Model(history).Updates(map[string]interface{}{
		"end_time":      &now,
		"duration":      duration,
		"status":        "failed",
		"error_message": err.Error(),
	})

	// 更新映射配置
	mapping.Status = 2
	mapping.LastInitAt = &now
	mapping.InitError = err.Error()
	mapping.RetryCount++
	DB.Save(mapping)

	logger.Log.Errorf("指标映射执行失败 [ID=%d, Instance=%d]: %v", mapping.ID, mapping.InstanceID, err)
	return err
}

// 辅助函数：更新成功状态
func updateMappingSuccess(mapping *MetricMapping, history *MetricMappingHistory, affectedHosts int) error {
	now := time.Now()
	duration := int(now.Sub(history.StartTime).Seconds())

	// 更新历史记录
	DB.Model(history).Updates(map[string]interface{}{
		"end_time":       &now,
		"duration":       duration,
		"status":         "success",
		"affected_hosts": affectedHosts,
	})

	// 更新映射配置
	mapping.Status = 1
	mapping.LastInitAt = &now
	mapping.LastSuccessAt = &now
	mapping.InitError = ""
	mapping.RetryCount = 0
	DB.Save(mapping)

	logger.Log.Infof("指标映射执行成功 [ID=%d, Instance=%d, Hosts=%d]", mapping.ID, mapping.InstanceID, affectedHosts)
	return nil
}

// 辅助函数：分割主机组ID
func splitGroupIDs(groupIDs string) []string {
	if groupIDs == "" {
		return []string{}
	}
	return strings.Split(groupIDs, ",")
}

// 辅助函数：执行映射
func executeMapping(apiInstance *APIInstance, config *MetricConfig, groupIDs []string) (int, error) {
	// 获取主机列表
	OutputPar := []string{"hostid"}
	rep, err := apiInstance.API.CallWithError("host.get", Params{
		"output":   OutputPar,
		"groupids": groupIDs,
	})
	if err != nil {
		return 0, err
	}

	type hostData struct {
		HostID string `json:"hostid"`
	}
	var hosts []hostData
	resByte, _ := json.Marshal(rep.Result)
	json.Unmarshal(resByte, &hosts)

	if len(hosts) == 0 {
		return 0, fmt.Errorf("未在指定的主机组中找到任何主机")
	}

	// 设置主机类型
	InventoryPara := make(map[string]string)
	InventoryPara["type"] = config.HostType
	_, err = apiInstance.API.CallWithError("host.massupdate", Params{
		"hosts":          hosts,
		"inventory_mode": 1,
		"inventory":      InventoryPara,
	})
	if err != nil {
		return 0, fmt.Errorf("设置主机类型失败: %w", err)
	}

	// 绑定指标
	for fieldName, itemIDs := range config.Metrics {
		if itemIDs == "" {
			continue
		}
		inventoryLink := getInventoryLinkByField(fieldName)
		if inventoryLink == 0 {
			logger.Log.Warnf("未知的字段名: %s", fieldName)
			continue
		}
		err = bindItemsToInventory(apiInstance, itemIDs, inventoryLink)
		if err != nil {
			logger.Log.Errorf("绑定指标失败 [field=%s]: %v", fieldName, err)
			// 继续执行，不中断
		}
	}

	// 绑定ICMP
	if config.PingTemplateID != "" {
		err = ICMPToInventoryWithInstance(config.PingTemplateID, apiInstance)
		if err != nil {
			logger.Log.Errorf("绑定ICMP失败: %v", err)
			// 继续执行，不中断
		}
	}

	return len(hosts), nil
}

// 辅助函数：根据字段名获取inventory link
func getInventoryLinkByField(fieldName string) int {
	mapping := map[string]int{
		"uptime":             UptimeID,
		"cpu_core":           CPUCore,
		"cpu_utilization":    CPUUtilizationID,
		"memory_utilization": MemoryUtilizationID,
		"memory_total":       MemoryTotalID,
		"memory_used":        MemoryUsedID,
		"model":              Model,
	}
	return mapping[fieldName]
}

// 辅助函数：绑定指标到inventory
func bindItemsToInventory(apiInstance *APIInstance, itemIDs string, inventoryLink int) error {
	ids := strings.Split(itemIDs, ",")
	for _, id := range ids {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		_, err := apiInstance.API.CallWithError("item.update", Params{
			"itemid":         id,
			"inventory_link": inventoryLink,
		})
		if err != nil {
			return fmt.Errorf("更新item %s 失败: %w", id, err)
		}
	}
	return nil
}

// GetMappingHistory 获取映射执行历史
func GetMappingHistory(mappingID int64, page, limit int) ([]MetricMappingHistory, int64, error) {
	var history []MetricMappingHistory
	var total int64

	query := DB.Model(&MetricMappingHistory{})
	if mappingID > 0 {
		query = query.Where("mapping_id = ?", mappingID)
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&history).Error

	return history, total, err
}

// GetAllMappingHistory 获取所有执行历史
func GetAllMappingHistory(page, limit int) ([]MetricMappingHistory, int64, error) {
	return GetMappingHistory(0, page, limit)
}
