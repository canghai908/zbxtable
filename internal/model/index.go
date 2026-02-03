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

// GetCountHost 获取所有实例的主机统计（多实例聚合版本）
func GetCountHost() (IndexInfo, error) {
	// 获取所有启用的实例
	instances, err := GetAllEnabledAPIInstances()
	if err != nil {
		logger.Log.Errorf("获取启用的实例失败: %v", err)
		return IndexInfo{}, err
	}

	d := IndexInfo{}

	// 聚合所有实例的数据
	for _, inst := range instances {
		linCount, err := getCountByTypeFromInstance(inst, "VM_LIN")
		if err != nil {
			logger.Log.Errorf("从实例 %s 获取 Linux 主机数量失败: %v", inst.Name, err)
		} else {
			d.LinCount += linCount
		}

		winCount, err := getCountByTypeFromInstance(inst, "VM_WIN")
		if err != nil {
			logger.Log.Errorf("从实例 %s 获取 Windows 主机数量失败: %v", inst.Name, err)
		} else {
			d.WinCount += winCount
		}

		srvCount, err := getCountByTypeFromInstance(inst, "HW_SRV")
		if err != nil {
			logger.Log.Errorf("从实例 %s 获取服务器数量失败: %v", inst.Name, err)
		} else {
			d.SRVCount += srvCount
		}

		netCount, err := getCountByTypeFromInstance(inst, "HW_NET")
		if err != nil {
			logger.Log.Errorf("从实例 %s 获取网络设备数量失败: %v", inst.Name, err)
		} else {
			d.NETCount += netCount
		}
	}

	return d, nil
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
	ret, err := CacheZRevRangeWithScores(MetType1+"_"+MetType2, 0, top_n)
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

		p1 := TopList{
			Hostname:     hostname,
			Score:        z.Score,
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
	TwoTree := []TwoChildren{
		{10, "Linux操作系统"},
		{11, "Windows操作系统"},
		{12, "网络设备"},
		{13, "物理服务器"},
	}
	tree := make([]Treeinventory, 1)
	tree[0].ID = 0
	tree[0].Name = "资产树"
	tree[0].TwoChildren = TwoTree

	return tree, nil

}
func GetOverviewData() (OverviewList, error) {
	var list = []string{"VM_LIN", "VM_WIN", "HW_NET", "HW_SRV"}
	//var one OverviewList
	//var datalist []OverviewList
	listmap := make(map[string][]Hosts)
	for _, v := range list {
		var ArrayOne []Hosts
		p, err := CacheGet(v + "_OVERVIEW")
		if err != nil || p == "" {
			logger.Log.Error(err)
			continue
		}
		err = json.Unmarshal([]byte(p), &ArrayOne)
		if err != nil {
			logger.Log.Error(err)
			continue
		}
		listmap[v] = ArrayOne
	}
	var newList OverviewList
	newList.Lin = listmap["VM_LIN"]
	newList.Win = listmap["VM_WIN"]
	newList.NET = listmap["HW_NET"]
	newList.SRV = listmap["HW_SRV"]
	return newList, nil
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
