package model

import (
	"fmt"
	"strconv"
	"strings"
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

// GetCountHost 获取所有实例的主机统计（多实例聚合版本，动态资产类型）
func GetCountHost() (IndexInfo, error) {
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
	for _, at := range assetTypes {
		var total int64
		for _, inst := range instances {
			count, cErr := getCountByTypeFromInstance(inst, at.TypeCode)
			if cErr != nil {
				logger.Log.Errorf("从实例 %s 获取 %s 主机数量失败: %v", inst.Name, at.TypeCode, cErr)
				continue
			}
			total += count
		}
		result[at.TypeCode] = total
	}
	return buildIndexInfoFromCounts(assetTypes, result), nil
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
		// 处理带实例前缀的主机名：tenant_id_hostname
		fullKey := fmt.Sprintf("%v", z.Member)
		hostname := fullKey
		instanceName := ""

		// 尝试分离实例ID和主机名
		parts := strings.SplitN(fullKey, "_", 2)
		if len(parts) == 2 {
			tenantID := parts[0]
			hostname = parts[1] // 使用主机名部分

			// 根据 tenant_id 查询实例名称
			var tenant ZabbixInstance
			err := DB.Where("instance = ?", tenantID).First(&tenant).Error
			if err == nil {
				instanceName = tenant.Name
			} else {
				logger.Log.Errorf("查询租户失败 (tenant_id=%s): %v", tenantID, err)
			}
		}

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
	tree[0].Name = "资产树"
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
