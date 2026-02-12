package model

import (
	"fmt"
	"regexp"
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
	err := DB.Where("zid = ? AND system_type = ?", m.ZID, m.SystemType).First(&existing).Error

	if err != nil {
		m.CreatedAt = time.Now()
		m.UpdatedAt = time.Now()
		return DB.Create(m).Error
	}

	m.ID = existing.ID
	m.CreatedAt = existing.CreatedAt
	m.UpdatedAt = time.Now()
	return DB.Save(m).Error
}

// GetMetricMappingsByInstance 获取实例的所有映射配置
func GetMetricMappingsByInstance(instanceID int) ([]MetricMapping, error) {
	var mappings []MetricMapping
	err := DB.Where("zid = ?", instanceID).Find(&mappings).Error
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
	history := &MetricMappingHistory{
		MappingID:  mapping.ID,
		ZID:        mapping.ZID,
		SystemType: mapping.SystemType,
		ExecType:   execType,
		StartTime:  time.Now(),
		Status:     "running",
	}
	DB.Create(history)

	apiInstance, err := GetAPIByZID(mapping.ZID)
	if err != nil {
		return updateMappingError(mapping, history, fmt.Errorf("获取实例API失败: %w", err))
	}

	config, err := mapping.GetMetricConfig()
	if err != nil {
		return updateMappingError(mapping, history, fmt.Errorf("解析配置失败: %w", err))
	}

	groupIDs := splitGroupIDs(mapping.HostGroupIDs)
	affectedHosts, err := executeMapping(apiInstance, config, groupIDs)
	if err != nil {
		return updateMappingError(mapping, history, err)
	}

	return updateMappingSuccess(mapping, history, affectedHosts)
}

func updateMappingError(mapping *MetricMapping, history *MetricMappingHistory, err error) error {
	now := time.Now()
	duration := int(now.Sub(history.StartTime).Seconds())

	DB.Model(history).Updates(map[string]interface{}{
		"end_time":      &now,
		"duration":      duration,
		"status":        "failed",
		"error_message": err.Error(),
	})

	mapping.Status = 2
	mapping.LastInitAt = &now
	mapping.InitError = err.Error()
	mapping.RetryCount++
	DB.Save(mapping)

	logger.Log.Errorf("指标映射执行失败 [ID=%d, zid=%d]: %v", mapping.ID, mapping.ZID, err)
	return err
}

func updateMappingSuccess(mapping *MetricMapping, history *MetricMappingHistory, affectedHosts int) error {
	now := time.Now()
	duration := int(now.Sub(history.StartTime).Seconds())

	DB.Model(history).Updates(map[string]interface{}{
		"end_time":       &now,
		"duration":       duration,
		"status":         "success",
		"affected_hosts": affectedHosts,
	})

	mapping.Status = 1
	mapping.LastInitAt = &now
	mapping.LastSuccessAt = &now
	mapping.InitError = ""
	mapping.RetryCount = 0
	DB.Save(mapping)

	logger.Log.Infof("指标映射执行成功 [ID=%d, zid=%d, Hosts=%d]", mapping.ID, mapping.ZID, affectedHosts)
	return nil
}

func splitGroupIDs(groupIDs string) []string {
	if groupIDs == "" {
		return []string{}
	}
	return strings.Split(groupIDs, ",")
}

func executeMapping(apiInstance *APIInstance, config *MetricConfig, groupIDs []string) (int, error) {
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
		}
	}

	if config.PingTemplateID != "" {
		err = ICMPToInventoryWithInstance(config.PingTemplateID, apiInstance)
		if err != nil {
			logger.Log.Errorf("绑定ICMP失败: %v", err)
		}
	}

	return len(hosts), nil
}

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

func GetAllMappingHistory(page, limit int) ([]MetricMappingHistory, int64, error) {
	return GetMappingHistory(0, page, limit)
}

func GetMappingRules(enabledOnly bool) ([]MappingRule, error) {
	var rules []MappingRule
	query := DB.Order("priority asc")
	if enabledOnly {
		query = query.Where("is_enabled = ?", 1)
	}
	err := query.Find(&rules).Error
	return rules, err
}

func CreateOrUpdateMappingRule(r *MappingRule) error {
	if r.ID > 0 {
		r.UpdatedAt = time.Now()
		return DB.Save(r).Error
	}
	r.CreatedAt = time.Now()
	r.UpdatedAt = time.Now()
	return DB.Create(r).Error
}

func DeleteMappingRule(id int64) error {
	return DB.Delete(&MappingRule{}, id).Error
}

func InitDefaultMappingRules() {
	defaultRules := []MappingRule{
		{
			RuleName:    "内置-CPU使用率",
			TargetField: "cpu_utilization",
			MatchType:   "regex",
			MatchValue:  `^system\.cpu\.util(\[.*\])?$`,
			Priority:    1,
			IsBuiltin:   1,
			IsEnabled:   1,
		},
		{
			RuleName:    "内置-系统运行时间",
			TargetField: "uptime",
			MatchType:   "key",
			MatchValue:  "system.uptime",
			Priority:    1,
			IsBuiltin:   1,
			IsEnabled:   1,
		},
		{
			RuleName:    "内置-内存总量",
			TargetField: "memory_total",
			MatchType:   "regex",
			MatchValue:  `^vm\.memory\.size\[total\]$`,
			Priority:    1,
			IsBuiltin:   1,
			IsEnabled:   1,
		},
		{
			RuleName:    "内置-内存已用",
			TargetField: "memory_used",
			MatchType:   "regex",
			MatchValue:  `^vm\.memory\.size\[used\]$`,
			Priority:    1,
			IsBuiltin:   1,
			IsEnabled:   1,
		},
	}

	for _, rule := range defaultRules {
		var count int64
		DB.Model(&MappingRule{}).Where("target_field = ? AND is_builtin = 1", rule.TargetField).Count(&count)
		if count == 0 {
			DB.Create(&rule)
		}
	}
}

func ExecuteMappingOnTemplates(apiInstance *APIInstance, rules []MappingRule) (int, error) {
	templateMap := make(map[string]bool)
	for _, rule := range rules {
		if rule.TemplateIDs != "" {
			ids := strings.Split(rule.TemplateIDs, ",")
			for _, id := range ids {
				if id = strings.TrimSpace(id); id != "" {
					templateMap[id] = true
				}
			}
		}
	}

	var targetTemplateIDs []string
	for id := range templateMap {
		targetTemplateIDs = append(targetTemplateIDs, id)
	}

	params := Params{
		"output": []string{"itemid", "name", "key_", "templateid", "inventory_link"},
	}
	if len(targetTemplateIDs) > 0 {
		params["templateids"] = targetTemplateIDs
	} else {
		params["inherited"] = false
		params["templated"] = true
	}

	rep, err := apiInstance.API.CallWithError("item.get", params)
	if err != nil {
		return 0, fmt.Errorf("获取模板指标失败: %w", err)
	}

	type itemData struct {
		ItemID        string `json:"itemid"`
		Name          string `json:"name"`
		Key           string `json:"key_"`
		TemplateID    string `json:"templateid"`
		InventoryLink string `json:"inventory_link"`
	}
	var items []itemData
	resByte, _ := json.Marshal(rep.Result)
	json.Unmarshal(resByte, &items)

	finalBindings := make(map[string]map[string]string)
	affectedTemplates := make(map[string]bool)

	for _, item := range items {
		tID := item.TemplateID
		if tID == "" || tID == "0" {
			continue
		}

		if _, ok := finalBindings[tID]; !ok {
			finalBindings[tID] = make(map[string]string)
		}

		for _, rule := range rules {
			if rule.TemplateIDs != "" && !strings.Contains(rule.TemplateIDs, tID) {
				continue
			}

			if _, ok := finalBindings[tID][rule.TargetField]; ok {
				continue
			}

			isMatch := false
			switch rule.MatchType {
			case "key":
				isMatch = item.Key == rule.MatchValue
			case "name":
				isMatch = strings.Contains(item.Name, rule.MatchValue)
			case "regex":
				reg, err := regexp.Compile(rule.MatchValue)
				if err == nil {
					isMatch = reg.MatchString(item.Key)
				}
			}

			if isMatch {
				finalBindings[tID][rule.TargetField] = item.ItemID
				affectedTemplates[tID] = true
				break
			}
		}
	}

	updateCount := 0
	for tID, fields := range finalBindings {
		for field, itemID := range fields {
			inventoryLink := getInventoryLinkByField(field)
			if inventoryLink == 0 {
				continue
			}

			_, err := apiInstance.API.CallWithError("item.update", Params{
				"itemid":         itemID,
				"inventory_link": inventoryLink,
			})
			if err != nil {
				logger.Log.Errorf("更新模板 [ID=%s] 指标 [Field=%s] 失败: %v", tID, field, err)
			} else {
				updateCount++
			}
		}
	}

	return len(affectedTemplates), nil
}

// GetTemplateItemsDebug 获取模板 Items 用于调试 (包含 inventory_link)
func GetTemplateItemsDebug(apiInstance *APIInstance, templateID string) (any, error) {
	params := Params{
		"output":      []string{"itemid", "name", "key_", "inventory_link"},
		"templateids": templateID,
	}
	rep, err := apiInstance.API.CallWithError("item.get", params)
	if err != nil {
		return nil, err
	}
	return rep.Result, nil
}
