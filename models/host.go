package models

import (
	"context"
	"errors"
	"github.com/astaxie/beego/logs"
	"math"
	"strconv"
	"strings"
	"zbxtable/utils"
)

// HostsList func
func HostsList(HostType, page, limit, hosts, model, ip, available string) ([]Hosts, int64, error) {
	SelectInterfacesPar := []string{"ip", "port", "available", "error"}
	SearchInventoryInventoryPar := make(map[string]string)
	SearchInventoryInventoryPar["type"] = HostType
	rep, err := API.CallWithError("host.get", Params{
		"output":           "extend",
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
			d.MemoryUsed = v.Inventory.SoftwareAppD
			d.MemoryTotal = v.Inventory.SoftwareAppC
			d.Uptime = v.Inventory.SoftwareAppE
			d.DateHwInstall = v.Inventory.DateHwInstall
			d.DateHwExpiry = v.Inventory.DateHwExpiry
			d.MAC = v.Inventory.MacaddressA
			d.ResourceID = v.Inventory.SerialnoB
			d.SerialNo = v.Inventory.SerialnoA
			d.Location = v.Inventory.Location
			d.Department = v.Inventory.SiteCity
			d.Vendor = v.Inventory.Vendor
			d.Department = v.Inventory.SiteCity
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
			d.SerialNo = v.Inventory.SerialnoA
			d.Location = v.Inventory.Location
			d.Department = v.Inventory.SiteCity
			d.Vendor = v.Inventory.Vendor
			d.Department = v.Inventory.SiteCity
			d.Error = v.Error
			if HostType == "HW_NET" || HostType == "HW_SRV" {
				d.Available = v.SnmpAvailable
				d.Error = v.SnmpError
				d.SerialNo = v.Inventory.SerialnoA
				d.Location = v.Inventory.LocationLon
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

// get net host by name
func GetNetHostByName(name string) ([]Hosts, error) {
	var ctx = context.Background()
	val, err := RDB.Get(ctx, "HW_NET_OVERVIEW").Result()
	if err != nil {
		return []Hosts{}, nil
	}
	var list []Hosts
	err = json.Unmarshal([]byte(val), &list)
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

// host get
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
	d.Department = hb[0].Inventory.SiteCity
	return d, nil
}

// host get
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
		logs.Error(err)
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

// GetInterfaceItem 网络设备接口获取
// func GetInterfaceItem(hostid string) ([]InterfaceData, error) {
func GetInterfaceData(hostid string) ([]InterfaceData, error) {
	//zabbix 5.4以后版本处理
	if ZBX_V {
		ItemsOutput := []string{"itemid", "tags", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock", "valuemapid"}
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
			"sortfield":  "name",
			"tags":       Par})
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
		var TagsList []InterfaceData
		var Tags InterfaceData
		//page
		var ItemList []interface{}
		for _, v := range ts {
			for _, vv := range v.Tags {
				if vv.Tag == "interface" && strings.Contains(v.Name, vv.Value) {
					switch {
					case strings.Contains(v.Name, "Bits received"):
						v.Name = vv.Value
						ItemList = append(ItemList, v)
					}
				}
			}
		}
		page := len(ts) / len(ItemList)
		//遍历
		for k, v := range ts {
			//遍历tags
			for _, vv := range v.Tags {
				//查找tags为interface
				if vv.Tag == "interface" && strings.Contains(v.Name, vv.Value) {
					Tags.Name = vv.Value
					var id int64
					id, err := strconv.ParseInt(v.Itemid, 10, 64)
					if err != nil {
						id = 0
					}
					Tags.ID = id
					switch {
					case strings.Contains(v.Name, "Bits received"):
						Tags.BitsReceived = utils.InterfaceTrafficeStrTofloat64(v.Lastvalue)
						Tags.Lastclock = v.Lastclock
					case strings.Contains(v.Name, "Inbound packets discarded"):
						Tags.InDiscarded = utils.InterfaceStrToInt64((v.Lastvalue))
					case strings.Contains(v.Name, "Inbound packets with errors"):
						Tags.InErrors = utils.InterfaceStrToInt64((v.Lastvalue))
					case strings.Contains(v.Name, "Bits sent"):
						Tags.BitsSent = utils.InterfaceTrafficeStrTofloat64(v.Lastvalue)
					case strings.Contains(v.Name, "Outbound packets discarded"):
						Tags.OutDiscarded = utils.InterfaceStrToInt64((v.Lastvalue))
					case strings.Contains(v.Name, "Outbound packets with errors"):
						Tags.OutErrors = utils.InterfaceStrToInt64((v.Lastvalue))
					case strings.Contains(v.Name, "Speed"):
						Tags.Speed = v.Lastvalue
					case strings.Contains(v.Name, "Operational status"):
						if v.ValuemapID != "0" {
							p, _ := GetValueMapByID(v.ValuemapID, v.Lastvalue)
							Tags.OperationalStatus = p + "(" + v.Lastvalue + ")"
						} else {
							Tags.OperationalStatus = v.Lastvalue
						}
					}
					if (k+1)%page == 0 {
						TagsList = append(TagsList, Tags)
					}
				}
			}
		}
		return TagsList, nil
	}
	//旧版本处理
	selectItemsPar := []string{"itemid", "value_type", "name", "key_", "delay", "units", "lastvalue", "lastclock", "valuemapid"}
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
	//
	var list InterfaceDataList
	var data InterfaceData
	for _, v := range ApplicationRes {
		data.ID = utils.InterfaceStrToInt64(v.Applicationid)
		data.Name = strings.Replace(v.Name, "Interface ", "", -1)
		data.InDiscarded = utils.InterfaceStrToInt64(v.Items[0].Lastvalue)
		data.InErrors = utils.InterfaceStrToInt64(v.Items[1].Lastvalue)
		data.BitsReceived = utils.InterfaceTrafficeStrTofloat64(v.Items[2].Lastvalue)
		data.OutDiscarded = utils.InterfaceStrToInt64(v.Items[3].Lastvalue)
		data.OutErrors = utils.InterfaceStrToInt64(v.Items[4].Lastvalue)
		data.BitsSent = utils.InterfaceTrafficeStrTofloat64(v.Items[5].Lastvalue)
		data.Speed = v.Items[6].Lastvalue
		//valuemap 映射
		if v.Items[7].ValuemapID != "0" {
			p, _ := GetValueMapByID(v.Items[7].ValuemapID, v.Items[7].Lastvalue)
			data.OperationalStatus = p + "(" + v.Items[7].Lastvalue + ")"
		} else {
			data.OperationalStatus = v.Items[7].Lastvalue
		}
		data.Lastclock = v.Items[2].Lastclock
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

// HostsList func
func GetHostsList(HostType string) ([]Hosts, int64, error) {
	//获取主机列表
	//OutputPar := []string{"hostid", "host", "available", "status", "name", "error"}
	//SelectInventoryPar := []string{"model", "chassis", "contact"}
	SelectInterfacesPar := []string{"ip", "port", "available", "error"}
	SearchInventoryInventoryPar := make(map[string]string)
	SearchInventoryInventoryPar["type"] = HostType
	rep, err := API.CallWithError("host.get", Params{
		"output":           "extend",
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
			d.MemoryUsed = v.Inventory.SoftwareAppD
			d.MemoryTotal = v.Inventory.SoftwareAppC
			d.Uptime = v.Inventory.SoftwareAppE
			if HostType == "HW_NET" || HostType == "HW_SRV" {
				if len(v.Interfaces) != 0 {
					d.SerialNo = v.Inventory.SerialnoA
					d.Location = v.Inventory.Location
					d.Department = v.Inventory.SiteCity
				}
			}
			dt = append(dt, d)
		}

	} else {
		for _, v := range hb {
			var count int64
			var err error
			count, err = GetTriggerHostCount(v.Hostid)
			if err != nil {
				logs.Error(err)
				count = 0
			}
			if len(v.Interfaces) != 0 {
				d.Interfaces = v.Interfaces[0].IP
				d.Available = v.Interfaces[0].Available
				d.Error = v.Interfaces[0].Error
			}
			d.HostID = v.Hostid
			d.Host = v.Host
			d.Name = v.Name
			d.Status = v.Status
			//物理服务器可用性为ipmi
			d.Model = v.Inventory.Model
			d.OS = v.Inventory.Os
			d.NumberOfCores = v.Inventory.Software
			d.CPUUtilization = v.Inventory.SoftwareAppA
			d.MemoryUtilization = v.Inventory.SoftwareAppB
			d.MemoryUsed = v.Inventory.SoftwareAppD
			d.MemoryTotal = v.Inventory.SoftwareAppC
			d.Uptime = v.Inventory.SoftwareAppE
			d.Alarm = strconv.FormatInt(count, 10)
			if HostType == "HW_NET" || HostType == "HW_SRV" {
				d.Available = v.SnmpAvailable
				d.Error = v.SnmpError
				d.SerialNo = v.Inventory.SerialnoA
				d.Location = v.Inventory.LocationLon
				d.Department = v.Inventory.SiteCity
			}
			dt = append(dt, d)
		}
	}
	return dt, int64(len(dt)), err
}

// GetLinFileSystemItem linux文件系统
func GetLinFilesSystemData(hostid string) ([]LinFilesSystemData, error) {
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
		var TagsList []LinFilesSystemData
		var Tags LinFilesSystemData
		var count, rootcount int64
		count = 0
		rootcount = 0
		//
		for _, v := range ts {
			for _, vv := range v.Tags {
				if vv.Tag == "filesystem" || strings.Contains(vv.Value, "Filesystem") {
					//root /
					if vv.Value == "/" {
						Tags.Name = vv.Value
						switch {
						case v.Key == "vfs.fs.size["+vv.Value+",used]":
							Tags.UsedSpace = utils.InterfaceStrToInt64(v.Lastvalue)
							rootcount++
						case v.Key == "vfs.fs.inode["+vv.Value+",pfree]":
							Tags.InodesPUsed = utils.Float64Round2(float64(100) - utils.DecFloat64Round2(v.Lastvalue))
							Tags.Lastclock = v.Lastclock
							rootcount++
						case v.Key == "vfs.fs.size["+vv.Value+",pused]":
							Tags.SpaceUtilization = utils.DecFloat64Round2(v.Lastvalue)
							rootcount++
						case v.Key == "vfs.fs.size["+vv.Value+",total]":
							Tags.TotalSpace = utils.InterfaceStrToInt64(v.Lastvalue)
							rootcount++
						}
						if rootcount%4 == 0 {
							TagsList = append(TagsList, Tags)

						}
					} else {
						//if strings.Contains(v.Name, vv.Value) && vv.Value != "/" {
						Tags.Name = vv.Value
						switch {
						case v.Key == "vfs.fs.size["+vv.Value+",used]":
							Tags.UsedSpace = utils.InterfaceStrToInt64(v.Lastvalue)
							count++
						case v.Key == "vfs.fs.inode["+vv.Value+",pfree]":
							Tags.InodesPUsed = utils.Float64Round2(float64(100) - utils.DecFloat64Round2(v.Lastvalue))
							Tags.Lastclock = v.Lastclock
							count++
						case v.Key == "vfs.fs.size["+vv.Value+",pused]":
							Tags.SpaceUtilization = utils.DecFloat64Round2(v.Lastvalue)
							count++
						case v.Key == "vfs.fs.size["+vv.Value+",total]":
							Tags.TotalSpace = utils.InterfaceStrToInt64(v.Lastvalue)
							count++
						}
						if count%4 == 0 {
							TagsList = append(TagsList, Tags)
						} //}
					}
				}
			}
		}
		return TagsList, nil
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

// GetLinFileSystemItem 网络设备接口获取
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
		var TagsList []WinFilesSystemData
		var Tags WinFilesSystemData
		var count int64
		count = 0
		for _, v := range ts {
			for _, vv := range v.Tags {
				if vv.Tag == "filesystem" {
					if strings.Contains(v.Name, vv.Value) {
						Tags.Name = vv.Value
						switch {
						case v.Key == "vfs.fs.size["+vv.Value+",pused]":
							Tags.SpaceUtilization = utils.DecFloat64Round2(v.Lastvalue)
							Tags.Lastclock = v.Lastclock
							count++
						case v.Key == "vfs.fs.size["+vv.Value+",total]":
							Tags.TotalSpace = utils.InterfaceStrToInt64(v.Lastvalue)
							count++
						case v.Key == "vfs.fs.size["+vv.Value+",used]":
							Tags.UsedSpace = utils.InterfaceStrToInt64(v.Lastvalue)
							count++
						}
						if count%3 == 0 {
							TagsList = append(TagsList, Tags)
						}
					}
				}
			}
		}
		return TagsList, nil
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
