package model

import (
	stdjson "encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/logger"
)

type HostReportBulkConfig struct {
	SelectionMode    string          `json:"selection_mode"`
	ZID              interface{}     `json:"zid"`
	HostID           string          `json:"host_id"`
	ItemIDs          []string        `json:"item_ids"`
	GroupIDs         []string        `json:"group_ids"`
	TagFilters       []HostTagFilter `json:"tag_filters"`
	ReferenceHostID  string          `json:"reference_host_id"`
	ReferenceItemIDs []string        `json:"reference_item_ids"`
}

type hostReportTemplateItem struct {
	ItemID string
	Key    string
}

func HasBulkHostReportConfig(payload string) bool {
	if strings.TrimSpace(payload) == "" {
		return false
	}

	var configs []HostReportBulkConfig
	if err := json.Unmarshal([]byte(payload), &configs); err != nil {
		return false
	}

	for _, config := range configs {
		if config.SelectionMode != "" && config.SelectionMode != "single" {
			return true
		}
	}

	return false
}

func ResolveHostReportConfigs(payload string) ([]HostReportConfig, []string, error) {
	var rawConfigs []HostReportBulkConfig
	if err := json.Unmarshal([]byte(payload), &rawConfigs); err != nil {
		return nil, nil, fmt.Errorf("主机配置解析失败: %w", err)
	}

	expandedConfigs := make([]HostReportConfig, 0)
	flatItemIDs := make([]string, 0)
	hostItemsCache := make(map[string]map[string]string)
	referenceItemsCache := make(map[string][]hostReportTemplateItem)

	for _, rawConfig := range rawConfigs {
		selectionMode := rawConfig.SelectionMode
		if selectionMode == "" {
			selectionMode = "single"
		}

		zid := normalizeReportConfigZID(rawConfig.ZID)
		switch selectionMode {
		case "single":
			itemIDs := normalizeStringList(rawConfig.ItemIDs)
			if zid == "" || strings.TrimSpace(rawConfig.HostID) == "" || len(itemIDs) == 0 {
				continue
			}
			expandedConfigs = append(expandedConfigs, HostReportConfig{
				ZID:     zid,
				HostID:  strings.TrimSpace(rawConfig.HostID),
				ItemIDs: itemIDs,
			})
			flatItemIDs = append(flatItemIDs, itemIDs...)
		case "host_group", "tag":
			if zid == "" {
				return nil, nil, fmt.Errorf("批量主机配置缺少实例")
			}
			if strings.TrimSpace(rawConfig.ReferenceHostID) == "" {
				return nil, nil, fmt.Errorf("批量主机配置缺少参考主机")
			}

			inst, err := loadAPIInstanceByReportZID(zid)
			if err != nil {
				return nil, nil, err
			}

			templateItems, err := loadReferenceTemplateItems(zid, rawConfig.ReferenceHostID, rawConfig.ReferenceItemIDs, referenceItemsCache)
			if err != nil {
				return nil, nil, err
			}
			if len(templateItems) == 0 {
				return nil, nil, fmt.Errorf("参考指标不能为空")
			}

			matchedHosts, err := resolveBulkConfigHosts(inst, selectionMode, rawConfig)
			if err != nil {
				return nil, nil, err
			}

			for _, host := range matchedHosts {
				hostID := strings.TrimSpace(host.HostID)
				if hostID == "" {
					continue
				}

				itemMap, err := loadHostItemsMap(zid, hostID, hostItemsCache)
				if err != nil {
					return nil, nil, err
				}

				matchedItemIDs := make([]string, 0, len(templateItems))
				for _, templateItem := range templateItems {
					if targetItemID, ok := itemMap[templateItem.Key]; ok {
						matchedItemIDs = append(matchedItemIDs, targetItemID)
					}
				}

				if len(matchedItemIDs) == 0 {
					continue
				}

				expandedConfigs = append(expandedConfigs, HostReportConfig{
					ZID:     zid,
					HostID:  hostID,
					ItemIDs: matchedItemIDs,
				})
				flatItemIDs = append(flatItemIDs, matchedItemIDs...)
			}
		}
	}

	return expandedConfigs, flatItemIDs, nil
}

func BuildResolvedHostReportConfig(payload string) (string, string, error) {
	expandedConfigs, itemIDs, err := ResolveHostReportConfigs(payload)
	if err != nil {
		return "", "", err
	}
	if len(expandedConfigs) == 0 {
		return "", "", fmt.Errorf("没有可用的主机指标配置")
	}

	hostIDsJSON, err := json.Marshal(expandedConfigs)
	if err != nil {
		return "", "", err
	}
	itemIDsJSON, err := json.Marshal(itemIDs)
	if err != nil {
		return "", "", err
	}

	return string(hostIDsJSON), string(itemIDsJSON), nil
}

func PrepareHostReportExecutionConfig(report *Report) error {
	if report == nil || report.ReportType != "host" || !HasBulkHostReportConfig(report.HostIds) {
		return nil
	}

	resolvedHostIDs := strings.TrimSpace(report.ResolvedHostIds)
	resolvedItemIDs := strings.TrimSpace(report.ItemIds)
	if resolvedHostIDs == "" || resolvedItemIDs == "" {
		var err error
		resolvedHostIDs, resolvedItemIDs, err = BuildResolvedHostReportConfig(report.HostIds)
		if err != nil {
			return err
		}
		if report.ID > 0 {
			if err := UpdateReportResolvedHostConfigByID(report.ID, resolvedHostIDs, resolvedItemIDs); err != nil {
				return err
			}
		}
		report.ResolvedHostIds = resolvedHostIDs
		report.ItemIds = resolvedItemIDs
	}

	report.HostIds = resolvedHostIDs
	report.ItemIds = resolvedItemIDs
	return nil
}

func ProcessHostReportConfigAsync(reportID int, generateRealtime bool) {
	report, err := GetReportsByID(reportID)
	if err != nil {
		logger.Log.Error("读取报表失败:", err)
		return
	}

	startedAt := time.Now()
	report.ExecStatus = strconv.Itoa(Running)
	report.StartAt = &startedAt
	report.EndAt = nil
	if err := UpdateReportExecStatusByID(report); err != nil {
		logger.Log.Error("更新报表执行状态失败:", err)
		return
	}

	resolvedHostIDs, resolvedItemIDs, err := BuildResolvedHostReportConfig(report.HostIds)
	if err != nil {
		logger.Log.Error("解析批量主机报表配置失败:", err)
		markReportAsyncFailed(report, err)
		return
	}

	report.ResolvedHostIds = resolvedHostIDs
	report.ItemIds = resolvedItemIDs
	if err := UpdateReportResolvedHostConfigByID(report.ID, report.ResolvedHostIds, report.ItemIds); err != nil {
		logger.Log.Error("保存批量主机报表执行配置失败:", err)
		markReportAsyncFailed(report, err)
		return
	}

	if !generateRealtime {
		finishedAt := time.Now()
		report.ExecStatus = strconv.Itoa(NoBegin)
		report.EndAt = &finishedAt
		if err := UpdateReportExecStatusByID(report); err != nil {
			logger.Log.Error("更新报表执行状态失败:", err)
		}
		return
	}

	taskReport := *report
	taskReport.HostIds = taskReport.ResolvedHostIds
	if taskReport.ReportMode == "realtime" && len(taskReport.Cycle) == 0 {
		taskReport.Cycle = "realtime"
	}
	if err := TaskHostReport(taskReport); err != nil {
		logger.Log.Error("实时报表生成失败:", err)
		markReportAsyncFailed(report, err)
		return
	}

	finishedAt := time.Now()
	report.ExecStatus = strconv.Itoa(Success)
	report.EndAt = &finishedAt
	if err := UpdateReportExecStatusByID(report); err != nil {
		logger.Log.Error("更新报表执行状态失败:", err)
	}
}

func markReportAsyncFailed(report *Report, err error) {
	if report == nil {
		return
	}
	finishedAt := time.Now()
	report.ExecStatus = strconv.Itoa(Failed)
	report.EndAt = &finishedAt
	if updateErr := UpdateReportExecStatusByID(report); updateErr != nil {
		logger.Log.Error("更新报表失败状态失败:", updateErr)
	}
}

func normalizeReportConfigZID(raw interface{}) string {
	if raw == nil {
		return ""
	}

	switch value := raw.(type) {
	case string:
		return strings.TrimSpace(value)
	case float64:
		return strconv.Itoa(int(value))
	case int:
		return strconv.Itoa(value)
	case int64:
		return strconv.FormatInt(value, 10)
	case stdjson.Number:
		return value.String()
	default:
		return strings.TrimSpace(fmt.Sprint(value))
	}
}

func normalizeStringList(values []string) []string {
	normalized := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		normalized = append(normalized, trimmed)
	}
	return normalized
}

func loadAPIInstanceByReportZID(zid string) (*APIInstance, error) {
	id, err := strconv.Atoi(zid)
	if err != nil {
		return nil, fmt.Errorf("实例ID无效: %s", zid)
	}
	inst, err := GetAPIByZID(id)
	if err != nil {
		return nil, fmt.Errorf("获取实例失败: %w", err)
	}
	return inst, nil
}

func loadReferenceTemplateItems(zid, referenceHostID string, referenceItemIDs []string, cache map[string][]hostReportTemplateItem) ([]hostReportTemplateItem, error) {
	cacheKey := zid + ":" + strings.TrimSpace(referenceHostID)
	if cachedItems, ok := cache[cacheKey]; ok {
		return filterTemplateItemsBySelection(cachedItems, referenceItemIDs), nil
	}

	hostItems, err := loadHostSelectableItems(zid, referenceHostID)
	if err != nil {
		return nil, fmt.Errorf("加载参考主机指标失败: %w", err)
	}

	flattened := make([]hostReportTemplateItem, 0, len(hostItems))
	for _, item := range hostItems {
		flattened = append(flattened, hostReportTemplateItem{
			ItemID: strings.TrimSpace(item.Itemid),
			Key:    strings.TrimSpace(item.Key),
		})
	}
	cache[cacheKey] = flattened
	return filterTemplateItemsBySelection(flattened, referenceItemIDs), nil
}

func filterTemplateItemsBySelection(items []hostReportTemplateItem, selectedIDs []string) []hostReportTemplateItem {
	selectedSet := make(map[string]struct{}, len(selectedIDs))
	for _, itemID := range selectedIDs {
		trimmed := strings.TrimSpace(itemID)
		if trimmed == "" {
			continue
		}
		selectedSet[trimmed] = struct{}{}
	}

	filtered := make([]hostReportTemplateItem, 0, len(selectedSet))
	for _, item := range items {
		if _, ok := selectedSet[item.ItemID]; ok && item.Key != "" {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func resolveBulkConfigHosts(inst *APIInstance, selectionMode string, config HostReportBulkConfig) ([]Hosts, error) {
	switch selectionMode {
	case "host_group":
		groupIDs := normalizeStringList(config.GroupIDs)
		if len(groupIDs) == 0 {
			return nil, fmt.Errorf("主机组不能为空")
		}
		return GetHostsByGroupIDsFromInstance(inst, groupIDs)
	case "tag":
		tagFilters := normalizeHostTagFilters(config.TagFilters)
		if len(tagFilters) == 0 {
			return nil, fmt.Errorf("Tag 条件不能为空")
		}
		return GetHostsByTagsFromInstance(inst, tagFilters)
	default:
		return nil, fmt.Errorf("不支持的主机选择方式: %s", selectionMode)
	}
}

func loadHostItemsMap(zid, hostID string, cache map[string]map[string]string) (map[string]string, error) {
	cacheKey := zid + ":" + hostID
	if itemsMap, ok := cache[cacheKey]; ok {
		return itemsMap, nil
	}

	hostItems, err := loadHostSelectableItems(zid, hostID)
	if err != nil {
		return nil, fmt.Errorf("加载主机指标失败: %w", err)
	}

	itemsMap := make(map[string]string)
	for _, item := range hostItems {
		itemKey := strings.TrimSpace(item.Key)
		itemID := strings.TrimSpace(item.Itemid)
		if itemKey == "" || itemID == "" {
			continue
		}
		itemsMap[itemKey] = itemID
	}
	cache[cacheKey] = itemsMap
	return itemsMap, nil
}

func loadHostSelectableItems(zid, hostID string) ([]Items, error) {
	items, _, err := GetAllItemByHostIDFromInstance(zid, hostID)
	if err != nil {
		return nil, err
	}
	return items, nil
}
