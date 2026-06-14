package model

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"zbxtable/pkg/logger"
)

// getCountByTypeFromInstance 从指定实例获取主机数量
func getCountByTypeFromInstance(inst *APIInstance, hostType string) (int64, error) {
	inventoryParams := make(map[string]string)
	inventoryParams["type"] = hostType
	hostCount, err := inst.API.CallWithError("host.get", Params{
		"output":          "extend",
		"searchInventory": inventoryParams,
		"countOutput":     true})
	if err != nil {
		return 0, err
	}
	count, _ := strconv.ParseInt(hostCount.Result.(string), 10, 64)
	return count, nil
}

func buildIndexInfoFromCounts(assetTypes []AssetType, counts map[string]int64) IndexInfo {
	info := IndexInfo{
		HostCounts:      make(map[string]int64, len(counts)),
		AssetTypeCounts: make([]AssetTypeCount, 0, len(assetTypes)),
	}

	for _, at := range assetTypes {
		count := counts[at.TypeCode]
		info.HostCounts[at.TypeCode] = count
		info.AssetTypeCounts = append(info.AssetTypeCounts, AssetTypeCount{
			TypeCode: at.TypeCode,
			Name:     at.Name,
			Icon:     at.Icon,
			Count:    count,
		})
		info.TotalCount += count

		switch at.TypeCode {
		case "VM_LIN":
			info.LinCount = count
		case "VM_WIN":
			info.WinCount = count
		case "HW_SRV":
			info.SrvCount = count
		case "HW_NET":
			info.NetCount = count
		}
	}

	return info
}

// hostCountCacheKey 主机统计缓存键
const hostCountCacheKey = "index_host_count"

// hostCountCacheTTL 主机统计缓存有效期（10 分钟，后台每 5 分钟刷新一次，TTL > 刷新间隔确保不会因过期而让用户等待）
const hostCountCacheTTL = 10 * time.Minute

// GetCountHost 获取所有实例的主机统计（多实例聚合版本，动态资产类型）。
// 结果缓存 60s：该统计需对 N 个资产类型 × M 个实例做 N×M 次 Zabbix 远程调用，
// 不缓存会导致每次打开页面都串行等待多次远程往返（约 2s）。
func GetCountHost() (IndexInfo, error) {
	// 命中缓存直接返回
	if cached, _ := CacheGet(hostCountCacheKey); cached != "" {
		var info IndexInfo
		if err := json.Unmarshal([]byte(cached), &info); err == nil {
			return info, nil
		}
	}

	info, err := computeCountHost()
	if err != nil {
		return IndexInfo{}, err
	}
	if b, mErr := json.Marshal(info); mErr == nil {
		_ = CacheSet(hostCountCacheKey, string(b), hostCountCacheTTL)
	}
	return info, nil
}

// SyncHostCountCache 主动刷新主机统计缓存（由定时任务每 5 分钟调用一次）。
// 直接调用 computeCountHost 绕过缓存读取，确保缓存始终是最新数据。
func SyncHostCountCache() error {
	info, err := computeCountHost()
	if err != nil {
		logger.Log.Errorf("刷新主机统计缓存失败: %v", err)
		return err
	}
	if b, mErr := json.Marshal(info); mErr == nil {
		_ = CacheSet(hostCountCacheKey, string(b), hostCountCacheTTL)
		logger.Log.Info("主机统计缓存刷新成功")
	}
	return nil
}

// WarmHostCountCache 启动时异步预热主机统计缓存，首次打开资产树页面即命中缓存
func WarmHostCountCache() {
	go func() {
		if err := SyncHostCountCache(); err != nil {
			logger.Log.Warnf("启动预热主机统计缓存失败（非致命）: %v", err)
		}
	}()
}

// computeCountHost 实际计算主机统计（并行查询各实例，减少串行等待）
func computeCountHost() (IndexInfo, error) {
	instances, err := GetAllEnabledAPIInstances()
	if err != nil {
		logger.Log.Errorf("获取启用的实例失败: %v", err)
		return IndexInfo{}, err
	}

	assetTypes, err := GetAllAssetTypes()
	if err != nil {
		logger.Log.Errorf("获取资产类型失败: %v", err)
		return IndexInfo{}, err
	}

	result := make(map[string]int64, len(assetTypes))
	var mu sync.Mutex
	var wg sync.WaitGroup
	// 并行：每个 (类型, 实例) 组合一个 goroutine，避免 N×M 次串行远程往返
	for _, at := range assetTypes {
		for _, inst := range instances {
			wg.Add(1)
			go func(typeCode string, instance *APIInstance) {
				defer wg.Done()
				count, cErr := getCountByTypeFromInstance(instance, typeCode)
				if cErr != nil {
					logger.Log.Errorf("从实例 %s 获取 %s 主机数量失败: %v", instance.Name, typeCode, cErr)
					return
				}
				mu.Lock()
				result[typeCode] += count
				mu.Unlock()
			}(at.TypeCode, inst)
		}
	}
	wg.Wait()
	return buildIndexInfoFromCounts(assetTypes, result), nil
}

func parseTopHostDisplay(fullKey string) (hostname string, instanceName string) {
	hostname = fullKey

	parts := strings.SplitN(fullKey, "_", 3)
	if len(parts) >= 2 {
		tenantID := parts[0]
		if len(parts) == 3 {
			hostname = parts[2]
		} else {
			// 兼容旧缓存格式：tenant_host
			hostname = parts[1]
		}

		var tenant ZabbixInstance
		err := DB.Where("instance = ?", tenantID).First(&tenant).Error
		if err == nil {
			instanceName = tenant.Name
		} else {
			logger.Log.Errorf("查询租户失败 (tenant_id=%s): %v", tenantID, err)
		}
	}

	return hostname, instanceName
}

// GetTopList top数据获取
func GetTopList(host_type, metrics_type, top_num string) (info []TopList, err error) {
	var MetType1, MetType2 string
	switch host_type {
	case "VM_WIN":
		MetType1 = "WIN"
	case "VM_LIN":
		MetType1 = "LIN"
	default:
		MetType1 = "WIN"
	}
	switch metrics_type {
	case "CPU":
		MetType2 = "CPU"
	case "MEM":
		MetType2 = "MEM"
	default:
		MetType2 = "CPU"
	}
	var top_n int64
	p, err := strconv.ParseInt(top_num, 10, 64)
	if err != nil {
		logger.Log.Error(err)
		top_n = 5
	} else {
		top_n = p
	}
	// A2：返回当前排序维度的 TopN，但每条同时补齐 cpu/mem 两项
	// 注意：CacheZRevRangeWithScores 的 stop 是包含式，所以 stop=top_n-1
	stop := top_n - 1
	if stop < 0 {
		stop = 0
	}
	ret, err := CacheZRevRangeWithScores(MetType1+"_"+MetType2, 0, stop)
	if err != nil {
		return []TopList{}, err
	}
	var p2 []TopList
	for _, z := range ret {
		fullKey := fmt.Sprintf("%v", z.Member)
		hostname, instanceName := parseTopHostDisplay(fullKey)

		// 从两个 ZSet 补齐 cpu/mem
		cpuScore, _ := CacheZScore(MetType1+"_CPU", fullKey)
		memScore, _ := CacheZScore(MetType1+"_MEM", fullKey)

		// 当前排序维度的 score 仍保持兼容（用于前端展示/排序）
		score := z.Score
		p1 := TopList{
			Hostname:     hostname,
			Score:        score,
			CPU:          cpuScore,
			MEM:          memScore,
			InstanceName: instanceName,
		}
		p2 = append(p2, p1)
	}
	return p2, nil
}
func GetInventory() ([]Treeinventory, error) {
	//var list = []string{"VM_LIN", "VM_WIN", "HW_NET", "HW_SRV"}
	//var listmap map[string][]TreeChildren
	//listmap = make(map[string][]TreeChildren)
	//var ctx = context.Background()
	//for _, v := range list {
	//	var ArrayOne []TreeChildren
	//	p, err := RDB.Get(ctx, v+"_INVENTORY").Result()
	//	if err != nil {
	//		listmap[v] = ArrayOne
	//		logs.Error(err)
	//		continue
	//	}
	//	err = json.Unmarshal([]byte(p), &ArrayOne)
	//	if err != nil {
	//		listmap[v] = ArrayOne
	//		logs.Error(err)
	//	}
	//	listmap[v] = ArrayOne
	//}
	//tree
	//TwoTree := []TwoChildren{
	//	{10, "Linux操作系统", listmap["VM_LIN"]},
	//	{11, "Windows操作系统", listmap["VM_WIN"]},
	//	{12, "网络设备", listmap["HW_NET"]},
	//	{13, "物理服务器", listmap["HW_SRV"]},
	//}
	assetTypes, err := GetAllAssetTypes()
	if err != nil {
		logger.Log.Errorf("获取资产类型失败: %v", err)
		return nil, err
	}
	TwoTree := make([]TwoChildren, 0, len(assetTypes))
	for i, at := range assetTypes {
		TwoTree = append(TwoTree, TwoChildren{
			ID:       int64(10 + i),
			Name:     at.Name,
			TypeCode: at.TypeCode,
			Icon:     at.Icon,
		})
	}
	tree := make([]Treeinventory, 1)
	tree[0].ID = 0
	tree[0].Name = "设备树"
	tree[0].TwoChildren = TwoTree

	return tree, nil

}
func GetOverviewData() (OverviewList, error) {
	assetTypes, err := GetAllAssetTypes()
	if err != nil {
		logger.Log.Errorf("获取资产类型失败: %v", err)
		return OverviewList{}, err
	}

	result := make(OverviewList)
	for _, at := range assetTypes {
		var hosts []Hosts
		p, cErr := CacheGet(at.TypeCode + "_OVERVIEW")
		if cErr != nil || p == "" {
			result[at.TypeCode] = []Hosts{}
			continue
		}
		if uErr := json.Unmarshal([]byte(p), &hosts); uErr != nil {
			logger.Log.Error(uErr)
			result[at.TypeCode] = []Hosts{}
			continue
		}
		result[at.TypeCode] = hosts
	}
	return result, nil
}
func GetEgressData() (EgressList, error) {
	p, err := CacheGet("Egress")
	if err != nil || p == "" {
		logger.Log.Error(err)
		return EgressList{}, err
	}
	var data EgressList
	err = json.Unmarshal([]byte(p), &data)
	if err != nil {
		logger.Log.Error(err)
		return EgressList{}, err
	}
	return data, nil
}

func GetZbxSession() (string, error) {
	p, err := CacheGet("zbx_session")
	if err != nil || p == "" {
		logger.Log.Error(err)
		return "", err
	}
	return p, nil
}
