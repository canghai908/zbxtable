package model

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"
)

// HostsListMultiInstance 多实例主机列表查询（聚合所有启用的实例）
func HostsListMultiInstance(HostType, page, limit, hosts, model, ip, available string) ([]Hosts, int64, error) {
	// 获取所有启用的 API 实例
	instances, err := GetAllEnabledAPIInstances()
	if err != nil {
		logger.Log.Errorf("获取启用的实例失败: %v", err)
		return []Hosts{}, 0, err
	}

	// 并发查询所有实例
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allHosts []Hosts

	for _, inst := range instances {
		wg.Add(1)
		go func(instance *APIInstance) {
			defer wg.Done()

			// 查询该实例的主机列表
			hosts, err := queryHostsFromInstance(instance, HostType)
			if err != nil {
				logger.Log.Errorf("查询实例 %s 的主机失败: %v", instance.Name, err)
				return
			}

			// 为每个主机添加实例信息
			mu.Lock()
			for i := range hosts {
				hosts[i].ZID = instance.ZID
				hosts[i].InstanceName = instance.Name
			}
			allHosts = append(allHosts, hosts...)
			mu.Unlock()
		}(inst)
	}

	wg.Wait()

	// 过滤数据
	var filteredHosts []Hosts
	for _, h := range allHosts {
		if hosts != "" && !strings.Contains(h.Name, hosts) {
			continue
		}
		if model != "" && !strings.Contains(h.Model, model) {
			continue
		}
		if ip != "" && !strings.Contains(h.Interfaces, ip) {
			continue
		}
		if available != "" && !strings.Contains(h.Available, available) {
			continue
		}
		filteredHosts = append(filteredHosts, h)
	}

	// 分页处理
	return paginateHosts(filteredHosts, page, limit)
}

// queryHostsFromInstance 从单个实例查询主机列表
func queryHostsFromInstance(inst *APIInstance, HostType string) ([]Hosts, error) {
	SelectInterfacesPar := []string{"ip", "port", "available", "error"}
	SearchInventoryInventoryPar := make(map[string]string)
	SearchInventoryInventoryPar["type"] = HostType
	filterPar := make(map[string]string)
	filterPar["status"] = "0"

	rep, err := inst.API.CallWithError("host.get", Params{
		"output":           "extend",
		"filter":           filterPar,
		"searchInventory":  SearchInventoryInventoryPar,
		"selectInventory":  "extend",
		"selectInterfaces": SelectInterfacesPar})
	if err != nil {
		return nil, err
	}

	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return nil, err
	}

	var hb ListHosts
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return nil, err
	}

	var hosts []Hosts
	for _, v := range hb {
		var d Hosts
		d.HostID = v.Hostid
		d.Host = v.Host
		d.Name = v.Name
		if len(v.Interfaces) != 0 {
			d.Interfaces = v.Interfaces[0].IP
			d.Available = v.Interfaces[0].Available
			d.Error = v.Interfaces[0].Error
		}
		d.Status = v.Status
		d.Model = v.Inventory.Model
		d.OS = v.Inventory.Os
		d.NumberOfCores = v.Inventory.Software
		d.CPUUtilization = v.Inventory.SoftwareAppA
		d.MemoryUtilization = v.Inventory.SoftwareAppB
		d.MemoryTotal = v.Inventory.SoftwareAppC
		d.MemoryUsed = v.Inventory.SoftwareAppD
		d.Uptime = v.Inventory.SoftwareAppE
		d.DateHwInstall = v.Inventory.DateHwInstall
		d.DateHwExpiry = v.Inventory.DateHwExpiry
		d.MAC = v.Inventory.MacaddressA
		d.ResourceID = v.Inventory.SerialnoB
		d.Vendor = v.Inventory.Vendor
		d.Ping = v.Inventory.Poc1Name
		d.PingLoss = v.Inventory.Poc1Email
		d.PingSec = v.Inventory.Poc1PhoneA

		if HostType == "HW_NET" || HostType == "HW_SRV" {
			if len(v.Interfaces) != 0 {
				d.SerialNo = v.Inventory.SerialnoA
				d.Location = v.Inventory.Location
				d.Department = v.Inventory.SiteCity
			}
		}

		// 处理旧版本的特殊字段
		if !inst.IsV54OrLater {
			if HostType == "HW_NET" || HostType == "HW_SRV" {
				d.Available = v.SnmpAvailable
				d.Error = v.SnmpError
			}
		}

		hosts = append(hosts, d)
	}

	return hosts, nil
}

// paginateHosts 分页处理
func paginateHosts(hosts []Hosts, page, limit string) ([]Hosts, int64, error) {
	IntPage, err := strconv.Atoi(page)
	if err != nil {
		IntPage = 1
	}
	IntLimit, err := strconv.Atoi(limit)
	if err != nil {
		IntLimit = 10
	}

	total := int64(len(hosts))
	if total == 0 {
		return hosts, 0, nil
	}

	// 计算分页
	totalpages := int(math.Ceil(float64(total) / float64(IntLimit)))
	if IntPage > totalpages {
		IntPage = totalpages
	}
	if IntPage <= 0 {
		IntPage = 1
	}

	begin := (IntPage - 1) * IntLimit
	end := IntPage * IntLimit
	if end > int(total) {
		end = int(total)
	}

	return hosts[begin:end], total, nil
}

// HostsList func (保留原函数用于单实例兼容)
func HostsList(HostType, page, limit, hosts, model, ip, available string) ([]Hosts, int64, error) {
	SelectInterfacesPar := []string{"ip", "port", "available", "error"}
	SearchInventoryInventoryPar := make(map[string]string)
	SearchInventoryInventoryPar["type"] = HostType
	filterPar := make(map[string]string)
	filterPar["status"] = "0"
	//old version
	rep, err := API.CallWithError("host.get", Params{
		"output":           "extend",
		"filter":           filterPar,
		"searchInventory":  SearchInventoryInventoryPar,
		"selectInventory":  "extend",
		"selectInterfaces": SelectInterfacesPar})
	if err != nil {
		return []Hosts{}, 0, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []Hosts{}, 0, err
	}
	var hb ListHosts
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []Hosts{}, 0, err
	}
	var dt []Hosts
	var d Hosts
	//new version
	if ZBX_V {
		for _, v := range hb {
			d.HostID = v.Hostid
			d.Host = v.Host
			d.Name = v.Name
			if len(v.Interfaces) != 0 {
				d.Interfaces = v.Interfaces[0].IP
				d.Available = v.Interfaces[0].Available
				d.Error = v.Interfaces[0].Error
			}
			d.Status = v.Status
			//物理服务器可用性为ipmi
			d.Model = v.Inventory.Model
			d.OS = v.Inventory.Os
			d.NumberOfCores = v.Inventory.Software
			d.CPUUtilization = v.Inventory.SoftwareAppA
			d.MemoryUtilization = v.Inventory.SoftwareAppB
			d.MemoryTotal = v.Inventory.SoftwareAppC
			d.MemoryUsed = v.Inventory.SoftwareAppD
			d.Uptime = v.Inventory.SoftwareAppE
			d.DateHwInstall = v.Inventory.DateHwInstall
			d.DateHwExpiry = v.Inventory.DateHwExpiry
			d.MAC = v.Inventory.MacaddressA
			d.ResourceID = v.Inventory.SerialnoB
			//d.SerialNo = v.Inventory.SerialnoA
			//d.Location = v.Inventory.Location
			//d.Department = v.Inventory.SiteCity
			d.Vendor = v.Inventory.Vendor
			d.Ping = v.Inventory.Poc1Name
			d.PingLoss = v.Inventory.Poc1Email
			d.PingSec = v.Inventory.Poc1PhoneA
			if HostType == "HW_NET" || HostType == "HW_SRV" {
				if len(v.Interfaces) != 0 {
					d.SerialNo = v.Inventory.SerialnoA
					d.Location = v.Inventory.Location
					d.Department = v.Inventory.SiteCity
				}
			}
			if hosts != "" && strings.Contains(d.Name, hosts) {
				dt = append(dt, d)
			} else if model != "" && strings.Contains(d.Model, model) {
				dt = append(dt, d)
			} else if ip != "" && strings.Contains(d.Interfaces, ip) {
				dt = append(dt, d)
			} else if available != "" && strings.Contains(d.Available, available) {
				dt = append(dt, d)
			} else if (hosts == "") && (model == "") && (ip == "") && (available == "") {
				dt = append(dt, d)
			}
		}
	} else {
		//老版本
		for _, v := range hb {
			d.HostID = v.Hostid
			d.Host = v.Host
			d.Name = v.Name
			if len(v.Interfaces) != 0 {
				d.Interfaces = v.Interfaces[0].IP
				d.Available = v.Interfaces[0].Available
			}
			d.Status = v.Status
			d.Error = v.Error
			//物理服务器可用性为ipmi
			d.Model = v.Inventory.Model
			d.OS = v.Inventory.Os
			d.NumberOfCores = v.Inventory.Software
			d.CPUUtilization = v.Inventory.SoftwareAppA
			d.MemoryUtilization = v.Inventory.SoftwareAppB
			d.MemoryUsed = v.Inventory.SoftwareAppD
			d.MemoryTotal = v.Inventory.SoftwareAppC
			d.Uptime = v.Inventory.SoftwareAppE
			d.DateHwInstall = v.Inventory.DateHwInstall
			d.DateHwExpiry = v.Inventory.DateHwExpiry
			d.MAC = v.Inventory.MacaddressA
			d.ResourceID = v.Inventory.SerialnoB
			//d.SerialNo = v.Inventory.SerialnoA
			d.Location = v.Inventory.Location
			//d.Department = v.Inventory.SiteCity
			d.Vendor = v.Inventory.Vendor
			d.Error = v.Error
			d.Ping = v.Inventory.Poc1Name
			d.PingLoss = v.Inventory.Poc1Email
			d.PingSec = v.Inventory.Poc1PhoneA
			if HostType == "HW_NET" || HostType == "HW_SRV" {
				d.Available = v.SnmpAvailable
				d.Error = v.SnmpError
				d.SerialNo = v.Inventory.SerialnoA
				d.Location = v.Inventory.Location
				d.Department = v.Inventory.SiteCity
			}
			if hosts != "" && strings.Contains(d.Name, hosts) {
				dt = append(dt, d)
			} else if model != "" && strings.Contains(d.Model, model) {
				dt = append(dt, d)
			} else if ip != "" && strings.Contains(d.Interfaces, ip) {
				dt = append(dt, d)
			} else if available != "" && strings.Contains(d.Available, available) {
				dt = append(dt, d)
			} else if (hosts == "") && (model == "") && (ip == "") && (available == "") {
				dt = append(dt, d)
			}
		}

	}
	IntPage, err := strconv.Atoi(page)
	if err != nil {
		IntPage = 1
	}
	IntLimit, err := strconv.Atoi(limit)
	if err != nil {
		IntLimit = 10
	}
	//如果dt为空直接返回
	if len(dt) == 0 {
		return dt, int64(len(dt)), err
	}
	//分页
	nums := len(dt)

	//end int
	//page总数
	totalpages := int(math.Ceil(float64(nums) / float64(IntLimit)))
	if IntPage >= totalpages {
		IntPage = totalpages
	}
	if IntPage <= 0 {
		IntPage = 1
	}
	//结束页数据
	var end int
	//begin 开始页数据

	begin := (IntPage - 1) * IntLimit
	if IntPage == totalpages {
		end = nums
	}
	if IntPage < totalpages {
		end = IntPage * IntLimit
	} else {
		end = nums
	}
	//根据开始和结束返回数据列表
	var newthostlist []Hosts
	for i := begin; i < end; i++ {
		newthostlist = append(newthostlist, dt[i])
	}
	return newthostlist, int64(len(dt)), err

}

// HostsListFromInstance 从指定实例获取主机列表
func HostsListFromInstance(instanceID, HostType, page, limit, hosts, model, ip, available string) ([]Hosts, int64, error) {
	// 获取实例API
	inst, err := GetZabbixInstanceAPI(instanceID)
	if err != nil {
		return []Hosts{}, 0, fmt.Errorf("获取实例API失败: %v", err)
	}

	// 查询该实例的主机列表
	allHosts, err := queryHostsFromInstance(inst, HostType)
	if err != nil {
		return []Hosts{}, 0, fmt.Errorf("查询实例主机失败: %v", err)
	}

	// 过滤数据
	var filteredHosts []Hosts
	for _, h := range allHosts {
		if hosts != "" && !strings.Contains(h.Name, hosts) {
			continue
		}
		if model != "" && !strings.Contains(h.Model, model) {
			continue
		}
		if ip != "" && !strings.Contains(h.Interfaces, ip) {
			continue
		}
		if available != "" && !strings.Contains(h.Available, available) {
			continue
		}
		filteredHosts = append(filteredHosts, h)
	}

	// 分页处理
	return paginateHosts(filteredHosts, page, limit)
}

// GetNetHostByName get net host by name
func GetNetHostByName(name string) ([]Hosts, error) {
	val, err := CacheGet("HW_NET_OVERVIEW")
	if err != nil || val == "" {
		return []Hosts{}, nil
	}
	var list []Hosts
	err = json.Unmarshal([]byte(val), &list)
	if err != nil {
		return []Hosts{}, nil
	}
	var newlist []Hosts
	if name == "" {
		return list, nil
	}
	for _, v := range list {
		if strings.ContainsAny(strings.ToLower(v.Name), strings.ToLower(name)) {
			newlist = append(newlist, v)
		}
	}
	return newlist, nil
}

// SearchHostFromInstance 从指定实例搜索主机
func SearchHostFromInstance(inst *APIInstance, name string) ([]Hosts, error) {
	filterPar := make(map[string]string)
	filterPar["status"] = "0"

	// 如果提供了名称，添加搜索条件
	searchPar := make(map[string]interface{})
	if name != "" {
		searchPar["name"] = name
	}

	SelectInterfacesPar := []string{"ip", "port", "available", "error"}

	params := Params{
		"output":           "extend",
		"filter":           filterPar,
		"selectInventory":  "extend",
		"selectInterfaces": SelectInterfacesPar,
	}

	if name != "" {
		params["search"] = searchPar
		params["searchWildcardsEnabled"] = true
	}

	rep, err := inst.API.CallWithError("host.get", params)
	if err != nil {
		return []Hosts{}, err
	}

	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return []Hosts{}, err
	}

	var hb ListHosts
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return []Hosts{}, err
	}

	var hosts []Hosts
	for _, v := range hb {
		var d Hosts
		d.HostID = v.Hostid
		d.Host = v.Host
		d.Name = v.Name
		if len(v.Interfaces) != 0 {
			d.Interfaces = v.Interfaces[0].IP
			d.Available = v.Interfaces[0].Available
			d.Error = v.Interfaces[0].Error
		}
		d.Status = v.Status
		d.ZID = inst.ZID
		d.InstanceName = inst.Name
		hosts = append(hosts, d)
	}

	return hosts, nil
}

// GetHostFromInstance 从指定实例获取主机详情
func GetHostFromInstance(inst *APIInstance, hostid string) (Hosts, error) {
	OutputPar := []string{"hostid", "host", "available", "status", "name", "error"}
	SelectInterfacesPar := []string{"ip", "port"}
	rep, err := inst.API.CallWithError("host.get", Params{
		"output":           OutputPar,
		"hostids":          hostid,
		"selectInventory":  "extend",
		"selectInterfaces": SelectInterfacesPar})
	if err != nil {
		return Hosts{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return Hosts{}, err
	}
	var hb ListHosts
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return Hosts{}, err
	}
	if len(hb) == 0 {
		return Hosts{}, errors.New("主机不存在")
	}

	var d Hosts
	d.HostID = hb[0].Hostid
	d.Host = hb[0].Host
	d.Name = hb[0].Name
	if len(hb[0].Interfaces) > 0 {
		d.Interfaces = hb[0].Interfaces[0].IP
	}
	d.Status = hb[0].Status
	d.Available = hb[0].Available
	d.Error = hb[0].Error
	d.NumberOfCores = hb[0].Inventory.Software
	d.CPUUtilization = hb[0].Inventory.SoftwareAppA
	d.MemoryUtilization = hb[0].Inventory.SoftwareAppB
	d.MemoryUsed = hb[0].Inventory.SoftwareAppD
	d.MemoryTotal = hb[0].Inventory.SoftwareAppC
	d.Uptime = hb[0].Inventory.SoftwareAppE
	d.OS = hb[0].Inventory.Os
	d.SystemName = hb[0].Inventory.Name
	d.SerialNo = hb[0].Inventory.SerialnoA
	d.Model = hb[0].Inventory.Model
	d.Location = hb[0].Inventory.Location
	d.DateHwExpiry = hb[0].Inventory.DateHwExpiry
	d.DateHwInstall = hb[0].Inventory.DateHwInstall
	d.Vendor = hb[0].Inventory.Vendor
	d.ResourceID = hb[0].Inventory.SerialnoB
	d.MAC = hb[0].Inventory.MacaddressA
	d.Department = hb[0].Inventory.SiteCity
	d.Ping = hb[0].Inventory.Poc1Name
	d.PingLoss = hb[0].Inventory.Poc1Email
	d.PingSec = hb[0].Inventory.Poc1PhoneA
	// 添加实例信息
	d.ZID = inst.ZID
	d.InstanceName = inst.Name
	return d, nil
}

// host get (保留用于兼容)
func GetHost(hostid string) (Hosts, error) {
	OutputPar := []string{"hostid", "host", "available", "status", "name", "error"}
	//SelectInventoryPar := []string{"model", "chassis", "contact", "asset_tag", "location", "hardware"}
	SelectInterfacesPar := []string{"ip", "port"}
	rep, err := API.CallWithError("host.get", Params{
		"output":           OutputPar,
		"hostids":          hostid,
		"selectInventory":  "extend",
		"selectInterfaces": SelectInterfacesPar})
	if err != nil {
		return Hosts{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return Hosts{}, err
	}
	var hb ListHosts
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return Hosts{}, err
	}
	if len(hb) == 0 {
		return Hosts{}, errors.New("主机不存在")
	}
	var d Hosts
	d.HostID = hb[0].Hostid
	d.Host = hb[0].Host
	d.Name = hb[0].Name
	if len(hb[0].Interfaces) > 0 {
		d.Interfaces = hb[0].Interfaces[0].IP
	}
	d.Status = hb[0].Status
	d.Available = hb[0].Available
	d.Error = hb[0].Error
	d.NumberOfCores = hb[0].Inventory.Software
	d.CPUUtilization = hb[0].Inventory.SoftwareAppA
	d.MemoryUtilization = hb[0].Inventory.SoftwareAppB
	d.MemoryUsed = hb[0].Inventory.SoftwareAppD
	d.MemoryTotal = hb[0].Inventory.SoftwareAppC
	d.Uptime = hb[0].Inventory.SoftwareAppE
	d.OS = hb[0].Inventory.Os
	d.SystemName = hb[0].Inventory.Name
	d.SerialNo = hb[0].Inventory.SerialnoA
	d.Model = hb[0].Inventory.Model
	d.Location = hb[0].Inventory.Location
	d.DateHwExpiry = hb[0].Inventory.DateHwExpiry
	d.DateHwInstall = hb[0].Inventory.DateHwInstall
	d.Vendor = hb[0].Inventory.Vendor
	d.ResourceID = hb[0].Inventory.SerialnoB
	d.MAC = hb[0].Inventory.MacaddressA
	d.Department = hb[0].Inventory.SiteCity
	d.Ping = hb[0].Inventory.Poc1Name
	d.PingLoss = hb[0].Inventory.Poc1Email
	d.PingSec = hb[0].Inventory.Poc1PhoneA
	return d, nil
}

// GetHostInfoTopology host
func GetHostInfoTopology(hostid string) (Hosts, error) {
	//获取基本信息
	OutputPar := []string{"hostid", "host", "available", "status", "name", "error"}
	//SelectInventoryPar := []string{"model", "chassis", "contact", "asset_tag", "location", "hardware"}
	SelectInterfacesPar := []string{"ip", "port"}
	rep, err := API.CallWithError("host.get", Params{
		"output":           OutputPar,
		"hostids":          hostid,
		"selectInventory":  "extend",
		"selectInterfaces": SelectInterfacesPar})
	if err != nil {
		return Hosts{}, err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return Hosts{}, err
	}
	var hb ListHosts
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return Hosts{}, err
	}
	//record not found
	if len(hb) == 0 {
		return Hosts{}, errors.New("hosts not found")
	}
	count, err := GetTriggerHostCount(hostid)
	if err != nil {
		logger.Log.Error(err)
	}

	var d Hosts
	d.HostID = hb[0].Hostid
	d.Host = hb[0].Host
	d.Name = hb[0].Name
	d.Interfaces = hb[0].Interfaces[0].IP
	d.Status = hb[0].Status
	d.Available = hb[0].Available
	d.Error = hb[0].Error
	d.NumberOfCores = hb[0].Inventory.Software
	d.CPUUtilization = hb[0].Inventory.SoftwareAppA
	d.MemoryUtilization = hb[0].Inventory.SoftwareAppB
	d.MemoryUsed = hb[0].Inventory.SoftwareAppD
	d.MemoryTotal = hb[0].Inventory.SoftwareAppC
	d.Uptime = hb[0].Inventory.SoftwareAppE
	d.OS = hb[0].Inventory.Os
	d.SystemName = hb[0].Inventory.Name
	d.SerialNo = hb[0].Inventory.SerialnoA
	d.Model = hb[0].Inventory.Model
	d.Location = hb[0].Inventory.Location
	d.DateHwExpiry = hb[0].Inventory.DateHwExpiry
	d.DateHwInstall = hb[0].Inventory.DateHwInstall
	d.Vendor = hb[0].Inventory.Vendor
	d.ResourceID = hb[0].Inventory.SerialnoB
	d.MAC = hb[0].Inventory.MacaddressA
	d.Alarm = strconv.FormatInt(count, 10)
	d.Ping = hb[0].Inventory.Poc1Name
	d.PingLoss = hb[0].Inventory.Poc1Email
	d.PingSec = hb[0].Inventory.Poc1PhoneA
	return d, nil
}

// GetMonItem 获取主机cpu、内存、磁盘、网卡流量itemid
func GetMonItem(hostid string) (MonItemList, error) {
	selectItemsPar := []string{"itemid", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
	Key2Par := []string{"CPU", "Memory", "Filesystem ", "Interface "}
	Search2Par := make(map[string][]string)
	Search2Par["name"] = Key2Par
	rep1, err := API.CallWithError("application.get", Params{
		"output":      "extend",
		"hostids":     hostid,
		"searchByAny": true,
		"search":      Search2Par,
		"selectItems": selectItemsPar,
		"sortfield":   "name"})
	if err != nil {
		return MonItemList{}, err
	}
	ApplicationResByte, err := json.Marshal(rep1.Result)
	if err != nil {
		return MonItemList{}, err
	}
	var ApplicationRes MonItemList
	err = json.Unmarshal(ApplicationResByte, &ApplicationRes)
	if err != nil {
		return MonItemList{}, err
	}
	return ApplicationRes, nil
}

// GetInterfaceData 网络设备接口获取
func GetInterfaceData(hostid string) ([]InterfaceData, error) {
	//zabbix 5.4以后版本处理
	if ZBX_V {
		ItemsOutput := []string{"itemid", "tags", "value_type", "snmp_oid", "name", "key_", "delay", "units", "lastvalue", "lastclock", "valuemapid"}
		selectTags := []string{"tag", "value"}
		Search2Par := make(map[string]string, 1)
		Search2Par["tag"] = "interface"
		Search2Par["value"] = ""
		Par := make(map[int]interface{})
		Par[0] = Search2Par
		rep1, err := API.CallWithError("item.get", Params{
			"output":     ItemsOutput,
			"hostids":    hostid,
			"selectTags": selectTags,
			//itemid排序
			"sortfield": "itemid",
			"tags":      Par})
		if err != nil {
			return InterfaceDataList{}, err
		}
		ApplicationResByte, err := json.Marshal(rep1.Result)
		if err != nil {
			return InterfaceDataList{}, err
		}
		var ts []MonIts
		err = json.Unmarshal(ApplicationResByte, &ts)
		if err != nil {
			return InterfaceDataList{}, err
		}
		//遍历整个数据
		rowData := make([]InterfaceData, 0)
		for _, v := range ts {
			//遍历Tags列表
			for _, vv := range v.Tags {
				//查找tags为interface
				if vv.Tag == "interface" && strings.Contains(v.Name, vv.Value) {
					//fmt.Println(k, vv.Value, v.Name)
					var existingRow *InterfaceData
					for i := range rowData {
						if rowData[i].Name == vv.Value {
							existingRow = &rowData[i]
							break
						}
					}
					if existingRow == nil {
						rowData = append(rowData, InterfaceData{
							Name: vv.Value,
						})
						existingRow = &rowData[len(rowData)-1]
					}
					switch {
					//收流量
					case strings.Contains(v.Name, "Bits received"):
						p := strings.Split(v.SNMPOid, ".")
						existingRow.Index = p[len(p)-1]
						existingRow.Lastclock = v.Lastclock
						existingRow.BitsReceived = v.Lastvalue
						existingRow.BitsReceivedItemId = v.Itemid
						existingRow.BitsReceivedValueType = v.ValueType
					case strings.Contains(v.Name, "Bits sent"):
						existingRow.BitsSent = v.Lastvalue
						existingRow.BitsSentItemId = v.Itemid
						existingRow.BitsSentValueType = v.ValueType
					case strings.Contains(v.Name, "Inbound packets discarded"):
						existingRow.InDiscarded = v.Lastvalue
						existingRow.InDiscardedItemId = v.Itemid
						existingRow.InDiscardedValueType = v.ValueType
					case strings.Contains(v.Name, "Inbound packets with errors"):
						existingRow.InErrors = v.Lastvalue
						existingRow.InErrorsItemId = v.Itemid
						existingRow.InErrorsValueType = v.ValueType
					case strings.Contains(v.Name, "Outbound packets discarded"):
						existingRow.OutDiscarded = v.Lastvalue
						existingRow.OutDiscardedItemId = v.Itemid
						existingRow.OutDiscardedValueType = v.ValueType
					case strings.Contains(v.Name, "Outbound packets with errors"):
						existingRow.OutErrors = v.Lastvalue
						existingRow.OutErrorsItemId = v.Itemid
						existingRow.OutErrorsValueType = v.ValueType
					case strings.Contains(v.Name, "Speed"):
						existingRow.Speed = v.Lastvalue
					case strings.Contains(v.Name, "Operational status"):
						var OperationalStatus string
						if v.ValuemapID != "0" {
							p, _ := GetValueMapByID(v.ValuemapID, v.Lastvalue)
							OperationalStatus = p + "(" + v.Lastvalue + ")"
						} else {
							OperationalStatus = v.Lastvalue
						}
						existingRow.OperationalStatus = OperationalStatus
						existingRow.OperationalStatusItemId = v.Itemid
						existingRow.OperationalStatusValueType = v.ValueType
					}
				}
			}
		}
		return rowData, nil
	}

	//旧版本处理
	//获取网卡应用所有地址
	selectItemsPar := []string{"itemid", "value_type", "name", "key_", "delay", "units",
		"lastvalue", "lastclock", "snmp_oid", "valuemapid"}
	Key2Par := []string{"Interface "}
	Search2Par := make(map[string][]string)
	Search2Par["name"] = Key2Par
	rep1, err := API.CallWithError("application.get", Params{
		"output":      "extend",
		"hostids":     hostid,
		"searchByAny": true,
		"search":      Search2Par,
		"selectItems": selectItemsPar,
		"sortfield":   "name"})
	if err != nil {
		return InterfaceDataList{}, err
	}
	ApplicationResByte, err := json.Marshal(rep1.Result)
	if err != nil {
		return InterfaceDataList{}, err
	}
	var ApplicationRes MonItemList
	err = json.Unmarshal(ApplicationResByte, &ApplicationRes)
	if err != nil {
		return InterfaceDataList{}, err
	}

	var list InterfaceDataList
	var data InterfaceData
	//遍历应用内所有接口
	for _, v := range ApplicationRes {
		//过滤interface {#IFNAME}接口
		if len(v.Items) < 7 {
			continue
		}
		data.Index = v.Applicationid
		data.Name = strings.Replace(v.Name, "Interface ", "", -1)
		for _, vv := range v.Items {
			//接口对应
			//收流量
			if strings.Contains(vv.Name, "Bits received") {
				data.BitsReceived = vv.Lastvalue
				data.Lastclock = vv.Lastclock
				index := strings.Split(vv.SNMPOid, ".")
				data.Index = index[len(index)-1]
				data.BitsReceivedItemId = vv.Itemid
				data.BitsReceivedValueType = vv.ValueType
			}
			//发流量
			if strings.Contains(vv.Name, "Bits sent") {
				data.BitsSent = vv.Lastvalue
				data.BitsSentItemId = vv.Itemid
				data.BitsSentValueType = vv.ValueType

			}
			//Inbound packets with errors
			if strings.Contains(vv.Name, "Inbound packets with errors") {
				data.InErrors = vv.Lastvalue
				data.InErrorsItemId = vv.Itemid
				data.InErrorsValueType = vv.ValueType
			}
			//Outbound packets with errors
			if strings.Contains(vv.Name, "Outbound packets with errors") {
				data.OutErrors = vv.Lastvalue
				data.OutErrorsItemId = vv.Itemid
				data.OutErrorsValueType = vv.ValueType
			}
			//Outbound packets discarded
			if strings.Contains(vv.Name, "Outbound packets discarded") {
				data.OutDiscarded = vv.Lastvalue
				data.OutDiscardedItemId = vv.Itemid
				data.OutDiscardedValueType = vv.ValueType
			}
			//Inbound packets discarded
			if strings.Contains(vv.Name, "Inbound packets discarded") {
				data.InDiscarded = vv.Lastvalue
				data.InDiscardedItemId = vv.Itemid
				data.InDiscardedValueType = vv.ValueType
			}
			//Speed
			if strings.Contains(vv.Name, "Speed") {
				data.Speed = vv.Lastvalue
			}
			//Operational status
			if strings.Contains(vv.Name, "Operational status") {
				if vv.ValuemapID != "0" {
					p, _ := GetValueMapByID(vv.ValuemapID, vv.Lastvalue)
					data.OperationalStatus = p + "(" + vv.Lastvalue + ")"
				} else {
					data.OperationalStatus = vv.Lastvalue
				}
				data.OperationalStatusItemId = vv.Itemid
				data.OperationalStatusValueType = vv.ValueType
			}
		}
		list = append(list, data)
	}
	return list, nil
}

// UpdateHost 主机信息更新
func UpdateHost(Host *Hosts) (MonItemList, error) {
	InventoryPar := make(map[string]string)
	InventoryPar["location"] = Host.Location
	InventoryPar["date_hw_expiry"] = Host.DateHwExpiry
	InventoryPar["date_hw_install"] = Host.DateHwInstall
	InventoryPar["serialno_b"] = Host.ResourceID
	InventoryPar["vendor"] = Host.Vendor
	InventoryPar["macaddress_a"] = Host.MAC
	InventoryPar["site_city"] = Host.Department
	_, err := API.CallWithError("host.update", Params{
		"hostid":    Host.HostID,
		"inventory": InventoryPar})
	if err != nil {
		return MonItemList{}, err
	}
	return MonItemList{}, nil
}

// GetLinFilesSystemData linux文件系统数据获取
func GetLinFilesSystemData(hostid string) ([]LinFilesSystemData, error) {
	//新版本
	if ZBX_V {
		ItemsOutput := []string{"itemid", "tags", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
		selectTags := []string{"tag", "value"}
		Search2Par := make(map[string]string, 1)
		Search2Par["tag"] = "filesystem"
		Search2Par["value"] = ""
		Par := make(map[int]interface{})
		Par[0] = Search2Par
		rep1, err := API.CallWithError("item.get", Params{
			"output":     ItemsOutput,
			"hostids":    hostid,
			"selectTags": selectTags,
			"sortfield":  "name",
			"tags":       Par})
		if err != nil {
			return []LinFilesSystemData{}, err
		}
		ApplicationResByte, err := json.Marshal(rep1.Result)
		if err != nil {
			return []LinFilesSystemData{}, err
		}
		var ts []MonIts
		err = json.Unmarshal(ApplicationResByte, &ts)
		if err != nil {
			return []LinFilesSystemData{}, err
		}
		var linFilesystemData []LinFilesSystemData
		filesystemData := make(map[string]map[string]interface{})
		//遍历文件系统数据
		for _, v := range ts {
			for _, vv := range v.Tags {
				if vv.Tag == "filesystem" || strings.Contains(vv.Value, "Filesystem") {
					fsData, ok := filesystemData[vv.Value]
					if !ok {
						fsData = make(map[string]interface{})
						filesystemData[vv.Value] = fsData
					}
					switch v.Key {
					case "vfs.fs.size[" + vv.Value + ",used]", "vfs.fs.dependent.size[" + vv.Value + ",used]":
						fsData["UsedSpace"] = utils.InterfaceStrToInt64(v.Lastvalue)
					case "vfs.fs.inode[" + vv.Value + ",pfree]", "vfs.fs.dependent.inode[" + vv.Value + ",pfree]":
						fsData["InodesPUsed"] = utils.Float64Round2(float64(100) - utils.DecFloat64Round2(v.Lastvalue))
						fsData["Lastclock"] = v.Lastclock
					case "vfs.fs.size[" + vv.Value + ",pused]", "vfs.fs.dependent.size[" + vv.Value + ",pused]":
						fsData["SpaceUtilization"] = utils.DecFloat64Round2(v.Lastvalue)
					case "vfs.fs.size[" + vv.Value + ",total]", "vfs.fs.dependent.size[" + vv.Value + ",total]":
						fsData["TotalSpace"] = utils.InterfaceStrToInt64(v.Lastvalue)
					}
				}
			}
		}
		// 将 map 转换为 LinFilesSystemData 切片
		for name, data := range filesystemData {
			fsData := LinFilesSystemData{
				Name: name,
			}
			if usedSpace, ok := data["UsedSpace"].(int64); ok {
				fsData.UsedSpace = usedSpace
			}
			if inodesPUsed, ok := data["InodesPUsed"].(float64); ok {
				fsData.InodesPUsed = inodesPUsed
			}
			if spaceUtilization, ok := data["SpaceUtilization"].(float64); ok {
				fsData.SpaceUtilization = spaceUtilization
			}
			if totalSpace, ok := data["TotalSpace"].(int64); ok {
				fsData.TotalSpace = totalSpace
			}
			if lastclock, ok := data["Lastclock"].(string); ok {
				fsData.Lastclock = lastclock
			}
			linFilesystemData = append(linFilesystemData, fsData)
		}
		return linFilesystemData, nil
	}
	//5.4以下版本处理
	selectItemsPar := []string{"itemid", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
	Key2Par := []string{"Filesystem "}
	Search2Par := make(map[string][]string)
	Search2Par["name"] = Key2Par
	rep1, err := API.CallWithError("application.get", Params{
		"output":      "extend",
		"hostids":     hostid,
		"searchByAny": true,
		"search":      Search2Par,
		"selectItems": selectItemsPar,
		"sortfield":   "name"})
	if err != nil {
		return []LinFilesSystemData{}, err
	}
	ApplicationResByte, err := json.Marshal(rep1.Result)
	if err != nil {
		return []LinFilesSystemData{}, err
	}
	var ApplicationRes MonItemList
	err = json.Unmarshal(ApplicationResByte, &ApplicationRes)
	if err != nil {
		return []LinFilesSystemData{}, err
	}
	var list []LinFilesSystemData
	var data LinFilesSystemData
	for _, v := range ApplicationRes {
		data.ID = utils.InterfaceStrToInt64(v.Applicationid)
		data.Name = strings.Replace(v.Name, "Filesystem ", "", -1)
		data.InodesPUsed = utils.DecFloat64Round2(v.Items[0].Lastvalue)
		data.SpaceUtilization = utils.DecFloat64Round2(v.Items[1].Lastvalue)
		data.TotalSpace = utils.InterfaceStrToInt64(v.Items[2].Lastvalue)
		data.UsedSpace = utils.InterfaceStrToInt64(v.Items[3].Lastvalue)
		data.Lastclock = v.Items[2].Lastclock
		list = append(list, data)
	}
	return list, nil
}

// GetWinFilesSystemData windows文件系统获取
func GetWinFilesSystemData(hostid string) ([]WinFilesSystemData, error) {
	//大于等于5.4版本处理
	if ZBX_V {
		ItemsOutput := []string{"itemid", "tags", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
		selectTags := []string{"tag", "value"}
		Search2Par := make(map[string]string, 1)
		Search2Par["tag"] = "filesystem"
		//Search2Par["value"] = ""
		Par := make(map[int]interface{})
		Par[0] = Search2Par
		rep1, err := API.CallWithError("item.get", Params{
			"output":     ItemsOutput,
			"hostids":    hostid,
			"selectTags": selectTags,
			"sortfield":  "name",
			"tags":       Par})
		if err != nil {
			return []WinFilesSystemData{}, err
		}
		ApplicationResByte, err := json.Marshal(rep1.Result)
		if err != nil {
			return []WinFilesSystemData{}, err
		}
		var ts []MonIts
		err = json.Unmarshal(ApplicationResByte, &ts)
		if err != nil {
			return []WinFilesSystemData{}, err
		}
		var fileList []WinFilesSystemData
		filesystemData := make(map[string]map[string]interface{})
		for _, v := range ts {
			for _, vv := range v.Tags {
				if vv.Tag == "filesystem" {
					if strings.Contains(v.Name, vv.Value) {
						fsData, ok := filesystemData[vv.Value]
						if !ok {
							fsData = make(map[string]interface{})
							filesystemData[vv.Value] = fsData
						}
						switch v.Key {
						case "vfs.fs.size[" + vv.Value + ",pused]", "vfs.fs.dependent.size[" + vv.Value + ",pused]":
							fsData["SpaceUtilization"] = utils.DecFloat64Round2(v.Lastvalue)
							fsData["Lastclock"] = v.Lastclock
						case "vfs.fs.size[" + vv.Value + ",total]", "vfs.fs.dependent.size[" + vv.Value + ",total]":
							fsData["TotalSpace"] = utils.InterfaceStrToInt64(v.Lastvalue)
						case "vfs.fs.size[" + vv.Value + ",used]", "vfs.fs.dependent.size[" + vv.Value + ",used]":
							fsData["UsedSpace"] = utils.InterfaceStrToInt64(v.Lastvalue)
						}
					}
				}
			}
		}
		// 将 map 转换为 LinFilesSystemData 切片
		for name, data := range filesystemData {
			fsData := WinFilesSystemData{
				Name: name,
			}
			if usedSpace, ok := data["UsedSpace"].(int64); ok {
				fsData.UsedSpace = usedSpace
			}
			if spaceUtilization, ok := data["SpaceUtilization"].(float64); ok {
				fsData.SpaceUtilization = spaceUtilization
			}
			if totalSpace, ok := data["TotalSpace"].(int64); ok {
				fsData.TotalSpace = totalSpace
			}
			if lastclock, ok := data["Lastclock"].(string); ok {
				fsData.Lastclock = lastclock
			}
			fileList = append(fileList, fsData)
		}
		return fileList, nil
	}
	//5.4以下版本处理
	selectItemsPar := []string{"itemid", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
	Key2Par := []string{"Filesystem "}
	Search2Par := make(map[string][]string)
	Search2Par["name"] = Key2Par
	rep1, err := API.CallWithError("application.get", Params{
		"output":      "extend",
		"hostids":     hostid,
		"searchByAny": true,
		"search":      Search2Par,
		"selectItems": selectItemsPar,
		"sortfield":   "name"})
	if err != nil {
		return []WinFilesSystemData{}, err
	}
	ApplicationResByte, err := json.Marshal(rep1.Result)
	if err != nil {
		return []WinFilesSystemData{}, err
	}
	var ApplicationRes MonItemList
	err = json.Unmarshal(ApplicationResByte, &ApplicationRes)
	if err != nil {
		return []WinFilesSystemData{}, err
	}
	var list []WinFilesSystemData
	var data WinFilesSystemData
	for _, v := range ApplicationRes {
		data.ID = utils.InterfaceStrToInt64(v.Applicationid)
		data.Name = strings.Replace(v.Name, "Filesystem ", "", -1)
		data.SpaceUtilization = utils.DecFloat64Round2(v.Items[0].Lastvalue)
		data.TotalSpace = utils.InterfaceStrToInt64(v.Items[1].Lastvalue)
		data.UsedSpace = utils.InterfaceStrToInt64(v.Items[2].Lastvalue)
		data.Lastclock = v.Items[2].Lastclock
		list = append(list, data)
	}
	return list, nil
}

func GetMonWinData(hostid string) (mon MonWinData, err error) {
	filesystem, err := GetWinFilesSystemData(hostid)
	if err != nil {
		return MonWinData{}, err
	}
	interfaces, err := GetInterfaceData(hostid)
	if err != nil {
		return MonWinData{}, err
	}
	var mo MonWinData
	mo.FileSystem = filesystem
	mo.FileSystemTotal = int64(len(filesystem))
	mo.Interfaces = interfaces
	mo.InterfacesTotal = int64(len(interfaces))
	return mo, nil
}

func GetMonLinData(hostid string) (mon MonLinData, err error) {
	filesystem, err := GetLinFilesSystemData(hostid)
	if err != nil {
		return MonLinData{}, err
	}
	interfaces, err := GetInterfaceData(hostid)
	if err != nil {
		return MonLinData{}, err
	}
	var mo MonLinData
	mo.FileSystem = filesystem
	mo.FileSystemTotal = int64(len(filesystem))
	mo.Interfaces = interfaces
	mo.InterfacesTotal = int64(len(interfaces))
	return mo, nil
}

type PNGData struct {
	Name string `json:"name"`
	Png  string `json:"png"`
}
type GraphReq struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

// FindInstanceByHostID 根据 hostid 查找所属的实例
func FindInstanceByHostID(hostID string) (int, error) {
	// 获取所有启用的实例
	instances, err := GetAllEnabledAPIInstances()
	if err != nil {
		return 0, err
	}

	// 遍历所有实例，查找包含该主机的实例
	for _, inst := range instances {
		// 尝试从该实例查询主机
		rep, err := inst.API.CallWithError("host.get", Params{
			"output":  []string{"hostid"},
			"hostids": hostID,
		})
		if err != nil {
			continue
		}

		hba, err := json.Marshal(rep.Result)
		if err != nil {
			continue
		}

		var hosts []map[string]interface{}
		err = json.Unmarshal(hba, &hosts)
		if err != nil {
			continue
		}

		// 如果找到了主机，返回该实例ID
		if len(hosts) > 0 {
			return inst.ZID, nil
		}
	}

	return 0, errors.New("未找到主机所属的实例")
}

// GetGraphDataFromInstance 从指定实例查看主机的图形数据
func GetGraphDataFromInstance(inst *APIInstance, hostId, start, end string) ([]PNGData, error) {
	var pngData []PNGData
	selectItemsPar := []string{"graphid", "name"}

	p, err := inst.API.CallWithError("graph.get", Params{
		"output":      selectItemsPar,
		"hostids":     hostId,
		"searchByAny": true,
		"sortfield":   "graphid"})
	if err != nil {
		return pngData, err
	}

	st, err := json.Marshal(p.Result)
	if err != nil {
		return pngData, err
	}

	var hba []GraphData
	err = json.Unmarshal(st, &hba)
	if err != nil {
		return pngData, err
	}

	var wg sync.WaitGroup
	pngDataChan := make(chan PNGData, len(hba))

	for _, v := range hba {
		wg.Add(1)
		go func(v GraphData) {
			defer wg.Done()
			// 使用该实例的 JAR 获取图形
			png, _ := GetPNGGraphFromInstance(inst, v.GraphId, start, end)
			pngDataChan <- PNGData{
				Name: v.Name,
				Png:  png,
			}
		}(v)
	}

	go func() {
		wg.Wait()
		close(pngDataChan)
	}()

	for data := range pngDataChan {
		pngData = append(pngData, data)
	}

	return pngData, nil
}

// GetMonItemFromInstance 从指定实例获取主机监控指标
func GetMonItemFromInstance(inst *APIInstance, hostid string) (MonItemList, error) {
	selectItemsPar := []string{"itemid", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
	Key2Par := []string{"CPU", "Memory", "Filesystem ", "Interface "}
	Search2Par := make(map[string][]string)
	Search2Par["name"] = Key2Par
	rep1, err := inst.API.CallWithError("application.get", Params{
		"output":      "extend",
		"hostids":     hostid,
		"searchByAny": true,
		"search":      Search2Par,
		"selectItems": selectItemsPar,
		"sortfield":   "name"})
	if err != nil {
		return MonItemList{}, err
	}
	ApplicationResByte, err := json.Marshal(rep1.Result)
	if err != nil {
		return MonItemList{}, err
	}
	var ApplicationRes MonItemList
	err = json.Unmarshal(ApplicationResByte, &ApplicationRes)
	if err != nil {
		return MonItemList{}, err
	}
	return ApplicationRes, nil
}

// GetInterfaceDataFromInstance 从指定实例获取网络设备接口数据
func GetInterfaceDataFromInstance(inst *APIInstance, hostid string) ([]InterfaceData, error) {
	// 使用实例的版本信息判断
	if inst.IsV54OrLater {
		return getInterfaceDataV54(inst, hostid)
	}
	return getInterfaceDataLegacy(inst, hostid)
}

// getInterfaceDataV54 获取接口数据（5.4及以上版本）
func getInterfaceDataV54(inst *APIInstance, hostid string) ([]InterfaceData, error) {
	ItemsOutput := []string{"itemid", "tags", "value_type", "snmp_oid", "name", "key_", "delay", "units", "lastvalue", "lastclock", "valuemapid"}
	selectTags := []string{"tag", "value"}
	Search2Par := make(map[string]string, 1)
	Search2Par["tag"] = "interface"
	Search2Par["value"] = ""
	Par := make(map[int]interface{})
	Par[0] = Search2Par
	rep1, err := inst.API.CallWithError("item.get", Params{
		"output":     ItemsOutput,
		"hostids":    hostid,
		"selectTags": selectTags,
		"sortfield":  "itemid",
		"tags":       Par})
	if err != nil {
		return []InterfaceData{}, err
	}
	ApplicationResByte, err := json.Marshal(rep1.Result)
	if err != nil {
		return []InterfaceData{}, err
	}
	var ts []MonIts
	err = json.Unmarshal(ApplicationResByte, &ts)
	if err != nil {
		return []InterfaceData{}, err
	}

	rowData := make([]InterfaceData, 0)
	for _, v := range ts {
		for _, vv := range v.Tags {
			if vv.Tag == "interface" && strings.Contains(v.Name, vv.Value) {
				var existingRow *InterfaceData
				for i := range rowData {
					if rowData[i].Name == vv.Value {
						existingRow = &rowData[i]
						break
					}
				}
				if existingRow == nil {
					rowData = append(rowData, InterfaceData{
						Name: vv.Value,
					})
					existingRow = &rowData[len(rowData)-1]
				}
				switch {
				case strings.Contains(v.Name, "Bits received"):
					p := strings.Split(v.SNMPOid, ".")
					existingRow.Index = p[len(p)-1]
					existingRow.Lastclock = v.Lastclock
					existingRow.BitsReceived = v.Lastvalue
					existingRow.BitsReceivedItemId = v.Itemid
					existingRow.BitsReceivedValueType = v.ValueType
				case strings.Contains(v.Name, "Bits sent"):
					existingRow.BitsSent = v.Lastvalue
					existingRow.BitsSentItemId = v.Itemid
					existingRow.BitsSentValueType = v.ValueType
				case strings.Contains(v.Name, "Inbound packets discarded"):
					existingRow.InDiscarded = v.Lastvalue
					existingRow.InDiscardedItemId = v.Itemid
					existingRow.InDiscardedValueType = v.ValueType
				case strings.Contains(v.Name, "Inbound packets with errors"):
					existingRow.InErrors = v.Lastvalue
					existingRow.InErrorsItemId = v.Itemid
					existingRow.InErrorsValueType = v.ValueType
				case strings.Contains(v.Name, "Outbound packets discarded"):
					existingRow.OutDiscarded = v.Lastvalue
					existingRow.OutDiscardedItemId = v.Itemid
					existingRow.OutDiscardedValueType = v.ValueType
				case strings.Contains(v.Name, "Outbound packets with errors"):
					existingRow.OutErrors = v.Lastvalue
					existingRow.OutErrorsItemId = v.Itemid
					existingRow.OutErrorsValueType = v.ValueType
				case strings.Contains(v.Name, "Speed"):
					existingRow.Speed = v.Lastvalue
				case strings.Contains(v.Name, "Operational status"):
					var OperationalStatus string
					if v.ValuemapID != "0" {
						p, _ := GetValueMapByIDFromInstance(inst, v.ValuemapID, v.Lastvalue)
						OperationalStatus = p + "(" + v.Lastvalue + ")"
					} else {
						OperationalStatus = v.Lastvalue
					}
					existingRow.OperationalStatus = OperationalStatus
					existingRow.OperationalStatusItemId = v.Itemid
					existingRow.OperationalStatusValueType = v.ValueType
				}
			}
		}
	}
	return rowData, nil
}

// getInterfaceDataLegacy 获取接口数据（5.4以下版本）
func getInterfaceDataLegacy(inst *APIInstance, hostid string) ([]InterfaceData, error) {
	selectItemsPar := []string{"itemid", "value_type", "name", "key_", "delay", "units",
		"lastvalue", "lastclock", "snmp_oid", "valuemapid"}
	Key2Par := []string{"Interface "}
	Search2Par := make(map[string][]string)
	Search2Par["name"] = Key2Par
	rep1, err := inst.API.CallWithError("application.get", Params{
		"output":      "extend",
		"hostids":     hostid,
		"searchByAny": true,
		"search":      Search2Par,
		"selectItems": selectItemsPar,
		"sortfield":   "name"})
	if err != nil {
		return []InterfaceData{}, err
	}
	ApplicationResByte, err := json.Marshal(rep1.Result)
	if err != nil {
		return []InterfaceData{}, err
	}
	var ApplicationRes MonItemList
	err = json.Unmarshal(ApplicationResByte, &ApplicationRes)
	if err != nil {
		return []InterfaceData{}, err
	}

	var list []InterfaceData
	var data InterfaceData
	for _, v := range ApplicationRes {
		if len(v.Items) < 7 {
			continue
		}
		data.Index = v.Applicationid
		data.Name = strings.Replace(v.Name, "Interface ", "", -1)
		for _, vv := range v.Items {
			if strings.Contains(vv.Name, "Bits received") {
				data.BitsReceived = vv.Lastvalue
				data.Lastclock = vv.Lastclock
				index := strings.Split(vv.SNMPOid, ".")
				data.Index = index[len(index)-1]
				data.BitsReceivedItemId = vv.Itemid
				data.BitsReceivedValueType = vv.ValueType
			}
			if strings.Contains(vv.Name, "Bits sent") {
				data.BitsSent = vv.Lastvalue
				data.BitsSentItemId = vv.Itemid
				data.BitsSentValueType = vv.ValueType
			}
			if strings.Contains(vv.Name, "Inbound packets with errors") {
				data.InErrors = vv.Lastvalue
				data.InErrorsItemId = vv.Itemid
				data.InErrorsValueType = vv.ValueType
			}
			if strings.Contains(vv.Name, "Outbound packets with errors") {
				data.OutErrors = vv.Lastvalue
				data.OutErrorsItemId = vv.Itemid
				data.OutErrorsValueType = vv.ValueType
			}
			if strings.Contains(vv.Name, "Outbound packets discarded") {
				data.OutDiscarded = vv.Lastvalue
				data.OutDiscardedItemId = vv.Itemid
				data.OutDiscardedValueType = vv.ValueType
			}
			if strings.Contains(vv.Name, "Inbound packets discarded") {
				data.InDiscarded = vv.Lastvalue
				data.InDiscardedItemId = vv.Itemid
				data.InDiscardedValueType = vv.ValueType
			}
			if strings.Contains(vv.Name, "Speed") {
				data.Speed = vv.Lastvalue
			}
			if strings.Contains(vv.Name, "Operational status") {
				if vv.ValuemapID != "0" {
					p, _ := GetValueMapByIDFromInstance(inst, vv.ValuemapID, vv.Lastvalue)
					data.OperationalStatus = p + "(" + vv.Lastvalue + ")"
				} else {
					data.OperationalStatus = vv.Lastvalue
				}
				data.OperationalStatusItemId = vv.Itemid
				data.OperationalStatusValueType = vv.ValueType
			}
		}
		list = append(list, data)
	}
	return list, nil
}

// GetMonWinDataFromInstance 从指定实例获取Windows监控数据
func GetMonWinDataFromInstance(inst *APIInstance, hostid string) (MonWinData, error) {
	filesystem, err := GetWinFilesSystemDataFromInstance(inst, hostid)
	if err != nil {
		return MonWinData{}, err
	}
	interfaces, err := GetInterfaceDataFromInstance(inst, hostid)
	if err != nil {
		return MonWinData{}, err
	}
	var mo MonWinData
	mo.FileSystem = filesystem
	mo.FileSystemTotal = int64(len(filesystem))
	mo.Interfaces = interfaces
	mo.InterfacesTotal = int64(len(interfaces))
	return mo, nil
}

// GetMonLinDataFromInstance 从指定实例获取Linux监控数据
func GetMonLinDataFromInstance(inst *APIInstance, hostid string) (MonLinData, error) {
	filesystem, err := GetLinFilesSystemDataFromInstance(inst, hostid)
	if err != nil {
		return MonLinData{}, err
	}
	interfaces, err := GetInterfaceDataFromInstance(inst, hostid)
	if err != nil {
		return MonLinData{}, err
	}
	var mo MonLinData
	mo.FileSystem = filesystem
	mo.FileSystemTotal = int64(len(filesystem))
	mo.Interfaces = interfaces
	mo.InterfacesTotal = int64(len(interfaces))
	return mo, nil
}

// GetWinFilesSystemDataFromInstance 从指定实例获取Windows文件系统数据
func GetWinFilesSystemDataFromInstance(inst *APIInstance, hostid string) ([]WinFilesSystemData, error) {
	if inst.IsV54OrLater {
		return getWinFilesSystemDataV54(inst, hostid)
	}
	return getWinFilesSystemDataLegacy(inst, hostid)
}

// getWinFilesSystemDataV54 获取Windows文件系统数据（5.4及以上版本）
func getWinFilesSystemDataV54(inst *APIInstance, hostid string) ([]WinFilesSystemData, error) {
	ItemsOutput := []string{"itemid", "tags", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
	selectTags := []string{"tag", "value"}
	Search2Par := make(map[string]string, 1)
	Search2Par["tag"] = "filesystem"
	Par := make(map[int]interface{})
	Par[0] = Search2Par
	rep1, err := inst.API.CallWithError("item.get", Params{
		"output":     ItemsOutput,
		"hostids":    hostid,
		"selectTags": selectTags,
		"sortfield":  "name",
		"tags":       Par})
	if err != nil {
		return []WinFilesSystemData{}, err
	}
	ApplicationResByte, err := json.Marshal(rep1.Result)
	if err != nil {
		return []WinFilesSystemData{}, err
	}
	var ts []MonIts
	err = json.Unmarshal(ApplicationResByte, &ts)
	if err != nil {
		return []WinFilesSystemData{}, err
	}
	var fileList []WinFilesSystemData
	filesystemData := make(map[string]map[string]interface{})
	for _, v := range ts {
		for _, vv := range v.Tags {
			if vv.Tag == "filesystem" {
				if strings.Contains(v.Name, vv.Value) {
					fsData, ok := filesystemData[vv.Value]
					if !ok {
						fsData = make(map[string]interface{})
						filesystemData[vv.Value] = fsData
					}
					switch v.Key {
					case "vfs.fs.size[" + vv.Value + ",pused]", "vfs.fs.dependent.size[" + vv.Value + ",pused]":
						fsData["SpaceUtilization"] = utils.DecFloat64Round2(v.Lastvalue)
						fsData["Lastclock"] = v.Lastclock
					case "vfs.fs.size[" + vv.Value + ",total]", "vfs.fs.dependent.size[" + vv.Value + ",total]":
						fsData["TotalSpace"] = utils.InterfaceStrToInt64(v.Lastvalue)
					case "vfs.fs.size[" + vv.Value + ",used]", "vfs.fs.dependent.size[" + vv.Value + ",used]":
						fsData["UsedSpace"] = utils.InterfaceStrToInt64(v.Lastvalue)
					}
				}
			}
		}
	}
	for name, data := range filesystemData {
		fsData := WinFilesSystemData{
			Name: name,
		}
		if usedSpace, ok := data["UsedSpace"].(int64); ok {
			fsData.UsedSpace = usedSpace
		}
		if spaceUtilization, ok := data["SpaceUtilization"].(float64); ok {
			fsData.SpaceUtilization = spaceUtilization
		}
		if totalSpace, ok := data["TotalSpace"].(int64); ok {
			fsData.TotalSpace = totalSpace
		}
		if lastclock, ok := data["Lastclock"].(string); ok {
			fsData.Lastclock = lastclock
		}
		fileList = append(fileList, fsData)
	}
	return fileList, nil
}

// getWinFilesSystemDataLegacy 获取Windows文件系统数据（5.4以下版本）
func getWinFilesSystemDataLegacy(inst *APIInstance, hostid string) ([]WinFilesSystemData, error) {
	selectItemsPar := []string{"itemid", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
	Key2Par := []string{"Filesystem "}
	Search2Par := make(map[string][]string)
	Search2Par["name"] = Key2Par
	rep1, err := inst.API.CallWithError("application.get", Params{
		"output":      "extend",
		"hostids":     hostid,
		"searchByAny": true,
		"search":      Search2Par,
		"selectItems": selectItemsPar,
		"sortfield":   "name"})
	if err != nil {
		return []WinFilesSystemData{}, err
	}
	ApplicationResByte, err := json.Marshal(rep1.Result)
	if err != nil {
		return []WinFilesSystemData{}, err
	}
	var ApplicationRes MonItemList
	err = json.Unmarshal(ApplicationResByte, &ApplicationRes)
	if err != nil {
		return []WinFilesSystemData{}, err
	}
	var list []WinFilesSystemData
	var data WinFilesSystemData
	for _, v := range ApplicationRes {
		data.ID = utils.InterfaceStrToInt64(v.Applicationid)
		data.Name = strings.Replace(v.Name, "Filesystem ", "", -1)
		data.SpaceUtilization = utils.DecFloat64Round2(v.Items[0].Lastvalue)
		data.TotalSpace = utils.InterfaceStrToInt64(v.Items[1].Lastvalue)
		data.UsedSpace = utils.InterfaceStrToInt64(v.Items[2].Lastvalue)
		data.Lastclock = v.Items[2].Lastclock
		list = append(list, data)
	}
	return list, nil
}

// GetLinFilesSystemDataFromInstance 从指定实例获取Linux文件系统数据
func GetLinFilesSystemDataFromInstance(inst *APIInstance, hostid string) ([]LinFilesSystemData, error) {
	if inst.IsV54OrLater {
		return getLinFilesSystemDataV54(inst, hostid)
	}
	return getLinFilesSystemDataLegacy(inst, hostid)
}

// getLinFilesSystemDataV54 获取Linux文件系统数据（5.4及以上版本）
func getLinFilesSystemDataV54(inst *APIInstance, hostid string) ([]LinFilesSystemData, error) {
	ItemsOutput := []string{"itemid", "tags", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
	selectTags := []string{"tag", "value"}
	Search2Par := make(map[string]string, 1)
	Search2Par["tag"] = "filesystem"
	Search2Par["value"] = ""
	Par := make(map[int]interface{})
	Par[0] = Search2Par
	rep1, err := inst.API.CallWithError("item.get", Params{
		"output":     ItemsOutput,
		"hostids":    hostid,
		"selectTags": selectTags,
		"sortfield":  "name",
		"tags":       Par})
	if err != nil {
		return []LinFilesSystemData{}, err
	}
	ApplicationResByte, err := json.Marshal(rep1.Result)
	if err != nil {
		return []LinFilesSystemData{}, err
	}
	var ts []MonIts
	err = json.Unmarshal(ApplicationResByte, &ts)
	if err != nil {
		return []LinFilesSystemData{}, err
	}
	var linFilesystemData []LinFilesSystemData
	filesystemData := make(map[string]map[string]interface{})
	for _, v := range ts {
		for _, vv := range v.Tags {
			if vv.Tag == "filesystem" || strings.Contains(vv.Value, "Filesystem") {
				fsData, ok := filesystemData[vv.Value]
				if !ok {
					fsData = make(map[string]interface{})
					filesystemData[vv.Value] = fsData
				}
				switch v.Key {
				case "vfs.fs.size[" + vv.Value + ",used]", "vfs.fs.dependent.size[" + vv.Value + ",used]":
					fsData["UsedSpace"] = utils.InterfaceStrToInt64(v.Lastvalue)
				case "vfs.fs.inode[" + vv.Value + ",pfree]", "vfs.fs.dependent.inode[" + vv.Value + ",pfree]":
					fsData["InodesPUsed"] = utils.Float64Round2(float64(100) - utils.DecFloat64Round2(v.Lastvalue))
					fsData["Lastclock"] = v.Lastclock
				case "vfs.fs.size[" + vv.Value + ",pused]", "vfs.fs.dependent.size[" + vv.Value + ",pused]":
					fsData["SpaceUtilization"] = utils.DecFloat64Round2(v.Lastvalue)
				case "vfs.fs.size[" + vv.Value + ",total]", "vfs.fs.dependent.size[" + vv.Value + ",total]":
					fsData["TotalSpace"] = utils.InterfaceStrToInt64(v.Lastvalue)
				}
			}
		}
	}
	for name, data := range filesystemData {
		fsData := LinFilesSystemData{
			Name: name,
		}
		if usedSpace, ok := data["UsedSpace"].(int64); ok {
			fsData.UsedSpace = usedSpace
		}
		if inodesPUsed, ok := data["InodesPUsed"].(float64); ok {
			fsData.InodesPUsed = inodesPUsed
		}
		if spaceUtilization, ok := data["SpaceUtilization"].(float64); ok {
			fsData.SpaceUtilization = spaceUtilization
		}
		if totalSpace, ok := data["TotalSpace"].(int64); ok {
			fsData.TotalSpace = totalSpace
		}
		if lastclock, ok := data["Lastclock"].(string); ok {
			fsData.Lastclock = lastclock
		}
		linFilesystemData = append(linFilesystemData, fsData)
	}
	return linFilesystemData, nil
}

// getLinFilesSystemDataLegacy 获取Linux文件系统数据（5.4以下版本）
func getLinFilesSystemDataLegacy(inst *APIInstance, hostid string) ([]LinFilesSystemData, error) {
	selectItemsPar := []string{"itemid", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock"}
	Key2Par := []string{"Filesystem "}
	Search2Par := make(map[string][]string)
	Search2Par["name"] = Key2Par
	rep1, err := inst.API.CallWithError("application.get", Params{
		"output":      "extend",
		"hostids":     hostid,
		"searchByAny": true,
		"search":      Search2Par,
		"selectItems": selectItemsPar,
		"sortfield":   "name"})
	if err != nil {
		return []LinFilesSystemData{}, err
	}
	ApplicationResByte, err := json.Marshal(rep1.Result)
	if err != nil {
		return []LinFilesSystemData{}, err
	}
	var ApplicationRes MonItemList
	err = json.Unmarshal(ApplicationResByte, &ApplicationRes)
	if err != nil {
		return []LinFilesSystemData{}, err
	}
	var list []LinFilesSystemData
	var data LinFilesSystemData
	for _, v := range ApplicationRes {
		data.ID = utils.InterfaceStrToInt64(v.Applicationid)
		data.Name = strings.Replace(v.Name, "Filesystem ", "", -1)
		data.InodesPUsed = utils.DecFloat64Round2(v.Items[0].Lastvalue)
		data.SpaceUtilization = utils.DecFloat64Round2(v.Items[1].Lastvalue)
		data.TotalSpace = utils.InterfaceStrToInt64(v.Items[2].Lastvalue)
		data.UsedSpace = utils.InterfaceStrToInt64(v.Items[3].Lastvalue)
		data.Lastclock = v.Items[2].Lastclock
		list = append(list, data)
	}
	return list, nil
}

// GetValueMapByIDFromInstance 从指定实例获取ValueMap
func GetValueMapByIDFromInstance(inst *APIInstance, valuemapid, value string) (string, error) {
	rep, err := inst.API.CallWithError("valuemap.get", Params{
		"output":         "extend",
		"valuemapids":    valuemapid,
		"selectMappings": "extend"})
	if err != nil {
		return "", err
	}
	hba, err := json.Marshal(rep.Result)
	if err != nil {
		return "", err
	}
	var hb []ValueMap
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		return "", err
	}
	if len(hb) == 0 {
		return value, nil
	}
	for _, v := range hb[0].Mappings {
		if v.Value == value {
			return v.Newvalue, nil
		}
	}
	return value, nil
}
