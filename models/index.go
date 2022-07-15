package models

import (
	"context"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
)

//GetCountHost func
func GetCountHost() (info IndexInfo, err error) {
	hosts, err := API.CallWithError("host.get", Params{"output": "extend",
		"countOutput": true})
	if err != nil {
		return IndexInfo{}, err
	}
	items, err := API.CallWithError("item.get", Params{"output": "extend",
		"countOutput": true})
	if err != nil {
		return IndexInfo{}, err
	}
	problems, err := API.CallWithError("problem.get", Params{"output": "extend",
		"countOutput": true,
	})
	if err != nil {
		return IndexInfo{}, err
	}
	triggers, err := API.CallWithError("trigger.get", Params{"output": "extend",
		"countOutput": true})
	if err != nil {
		return IndexInfo{}, err
	}
	//Linux主机
	LinuxInventoryPar := make(map[string]string)
	LinuxInventoryPar["type"] = "VM_LIN"
	LinuxCount, err := API.CallWithError("host.get", Params{
		"output":          "extend",
		"searchInventory": LinuxInventoryPar,
		"countOutput":     true})
	if err != nil {
		return IndexInfo{}, err
	}
	//Windows主机
	WinInventoryPar := make(map[string]string)
	WinInventoryPar["type"] = "VM_WIN"
	WinCount, err := API.CallWithError("host.get", Params{
		"output":          "extend",
		"searchInventory": WinInventoryPar,
		"countOutput":     true})
	if err != nil {
		return IndexInfo{}, err
	}
	//物理服务器
	SrvInventoryPar := make(map[string]string)
	SrvInventoryPar["type"] = "HW_SRV"
	SrvCount, err := API.CallWithError("host.get", Params{
		"output":          "extend",
		"searchInventory": SrvInventoryPar,
		"countOutput":     true})
	if err != nil {
		return IndexInfo{}, err
	}
	//网络设备
	NetInventoryPar := make(map[string]string)
	NetInventoryPar["type"] = "HW_NET"
	NetCount, err := API.CallWithError("host.get", Params{
		"output":          "extend",
		"searchInventory": NetInventoryPar,
		"countOutput":     true})
	if err != nil {
		return IndexInfo{}, err
	}
	d := IndexInfo{}
	hostsint64, _ := strconv.ParseInt(hosts.Result.(string), 10, 64)
	d.Hosts = hostsint64
	itemsint64, _ := strconv.ParseInt(items.Result.(string), 10, 64)
	d.Items = itemsint64
	problemsint64, _ := strconv.ParseInt(problems.Result.(string), 10, 64)
	d.Problems = problemsint64
	triggersint64, _ := strconv.ParseInt(triggers.Result.(string), 10, 64)
	d.Triggers = triggersint64
	linint64, _ := strconv.ParseInt(LinuxCount.Result.(string), 10, 64)
	d.LinCount = linint64
	winint64, _ := strconv.ParseInt(WinCount.Result.(string), 10, 64)
	d.WinCount = winint64
	netint64, _ := strconv.ParseInt(NetCount.Result.(string), 10, 64)
	d.NETCount = netint64
	srvint64, _ := strconv.ParseInt(SrvCount.Result.(string), 10, 64)
	d.SRVCount = srvint64
	return d, nil
}

//GetTopList top数据获取
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
		MetType1 = "CPU"
	}
	var top_n int64
	p, err := strconv.ParseInt(top_num, 10, 64)
	if err != nil {
		logs.Error(err)
		top_n = 5
	} else {
		top_n = p
	}
	var ctx = context.Background()
	ret, err := RDB.ZRevRangeWithScores(ctx, MetType1+"_"+MetType2, 0, top_n).Result()
	if err != nil {
		return []TopList{}, err

	}
	var p1 TopList
	var p2 []TopList
	for _, z := range ret {
		p1.Hostname = fmt.Sprintf("%v", z.Member)
		p1.Score = z.Score
		p2 = append(p2, p1)
	}
	return p2, nil
}
func GetInventory() ([]Treeinventory, error) {
	var list = []string{"VM_LIN", "VM_WIN", "HW_NET", "HW_SRV"}
	var listmap map[string][]TreeChildren
	listmap = make(map[string][]TreeChildren)
	var ctx = context.Background()
	for _, v := range list {
		var ArrayOne []TreeChildren
		p, err := RDB.Get(ctx, v+"_INVENTORY").Result()
		if err != nil {
			listmap[v] = ArrayOne
			logs.Error(err)
			continue
		}
		err = json.Unmarshal([]byte(p), &ArrayOne)
		if err != nil {
			listmap[v] = ArrayOne
			logs.Error(err)
		}
		listmap[v] = ArrayOne
	}
	//tree
	TwoTree := []TwoChildren{
		{10, "Linux操作系统", listmap["VM_LIN"]},
		{11, "Windows操作系统", listmap["VM_WIN"]},
		{12, "网络设备", listmap["HW_NET"]},
		{13, "物理服务器", listmap["HW_SRV"]},
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
	var listmap map[string][]Hosts
	listmap = make(map[string][]Hosts)
	var ctx = context.Background()
	for _, v := range list {
		var ArrayOne []Hosts
		p, err := RDB.Get(ctx, v+"_OVERVIEW").Result()
		if err != nil {
			return OverviewList{}, err
			logs.Error(err)
			continue
		}
		err = json.Unmarshal([]byte(p), &ArrayOne)
		if err != nil {
			logs.Error(err)
			return OverviewList{}, err
		}
		listmap[v] = ArrayOne
	}
	var Newlist OverviewList
	Newlist.Lin = listmap["VM_LIN"]
	Newlist.Win = listmap["VM_WIN"]
	Newlist.NET = listmap["HW_NET"]
	Newlist.SRV = listmap["HW_SRV"]
	return Newlist, nil
}
func GetEgressData() (EgressList, error) {
	var ctx = context.Background()
	p, err := RDB.Get(ctx, "Egress").Result()
	if err != nil {
		logs.Error(err)
		return EgressList{}, err
	}
	var data EgressList
	err = json.Unmarshal([]byte(p), &data)
	if err != nil {
		logs.Error(err)
		return EgressList{}, err
	}
	return data, nil
}
