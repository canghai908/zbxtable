package models

import (
	"context"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
)

func GetRouter(username string) ([]RouterRes, error) {
	m, err := GetManagerByName(username)
	if err != nil {
		return []RouterRes{}, err
	}
	if m.Role == "admin" {
		prouter := []RChildren{
			{Router: "dashboard", Path: "dashboard", Name: "工作台", Icon: "dashboard", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "首页", Router: "workplace", Path: "workplace", Icon: "home"},
					{Name: "资产管理", Router: "inventory", Path: "inventory", Icon: "calendar"},
					{Name: "状态总览", Router: "overview", Path: "overview", Icon: "appstore"},
				}},
			{Router: "host", Path: "host", Name: "主机应用", Icon: "hdd", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "Linux主机", Router: "linux", Path: "linux", Icon: "container"},
					{Name: "Linux主机详情", Router: "linDetail", Path: "lindetail", Meta: Meta{Highlight: "/host", Invisible: true}},
					{Name: "Windows主机", Router: "windows", Path: "windows", Icon: "windows"},
					{Name: "Windows主机详情", Router: "winDetail", Path: "windetail", Meta: Meta{Highlight: "/host", Invisible: true}},
				}},
			{Router: "net", Path: "net", Name: "网络管理", Icon: "cloud", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "网络设备", Router: "netList", Path: "list", Icon: "chrome"},
					{Name: "设备详情", Router: "netDetail", Path: "detail", Meta: Meta{Highlight: "/net", Invisible: true}},
				}},
			{Router: "server", Path: "server", Name: "硬件管理", Icon: "database", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "物理服务器", Router: "srvList", Path: "list", Icon: "mobile"},
					{Name: "设备详情", Router: "srvDetail", Path: "detail", Meta: Meta{Highlight: "/server", Invisible: true}},
				}},
			{Router: "alarm", Path: "alarm", Name: "告警管理", Icon: "alert", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "告警分析", Router: "alarmAnalysis", Path: "analysis", Icon: "hourglass"},
					{Name: "告警查询", Router: "alarmList", Path: "list", Icon: "eye"},
					{Name: "告警分发", Router: "alarmRule", Path: "rule", Icon: "message",
						Authority: Authority{Role: "admin", Permission: "['add','edit','delete','update']"}},
					{Name: "规则添加", Router: "alarmRuleAdd", Path: "rule-add", Meta: Meta{Highlight: "/alarm", Invisible: true}},
					{Name: "规则编辑", Router: "alarmRuleEdit", Path: "rule-edit", Meta: Meta{Highlight: "/alarm", Invisible: true}},
					{Name: "屏蔽规则", Router: "alarmMutes", Path: "mutes", Icon: "stop"},
				}},
			{Router: "topology", Path: "topology", Name: "拓扑管理", Icon: "picture", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "拓扑维护", Router: "topologyList", Path: "list", Icon: "environment"},
					{Name: "拓扑编辑", Router: "topologyDetail", Path: "detail", Meta: Meta{Highlight: "/topology", Invisible: true}},
					{Name: "拓扑展示", Router: "topologyShow", Path: "show", Meta: Meta{Highlight: "/topology", Invisible: true}},
				}},
			{Router: "report", Path: "report", Name: "报表管理", Icon: "file", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "流量报表", Router: "reportTraffic", Path: "traffic", Icon: "file-excel"},
					{Name: "报表编辑", Router: "reportTrafficEdit", Path: "edit", Meta: Meta{Highlight: "/report", Invisible: true}},
					{Name: "报表添加", Router: "reportTrafficAdd", Path: "add", Meta: Meta{Highlight: "/report", Invisible: true}},
				}},
			{Router: "system", Path: "system", Name: "系统管理", Icon: "setting", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "用户管理", Router: "systemUsers", Path: "users", Icon: "meh",
						Authority: Authority{Role: "admin", Permission: "['add','edit','delete','update']"}},
					{Name: "组织管理", Router: "systemGroups", Path: "groups", Icon: "smile",
						Authority: Authority{Role: "admin", Permission: "['add','edit','delete','update']"}},
					{Name: "指标映射", Router: "sysInit", Path: "init", Icon: "interaction"},
					{Name: "映射编辑", Router: "sysInitEdit", Path: "init-edit", Meta: Meta{Highlight: "/system", Invisible: true}},
					{Name: "出口配置", Router: "systemBandwidth", Path: "bandwidth", Icon: "api"},
				}},
		}
		tree := make([]RouterRes, 1)
		tree[0].Router = "root"
		tree[0].Children = prouter
		return tree, nil
	}
	if m.Role == "user" {
		prouter := []RChildren{
			{Router: "dashboard", Path: "dashboard", Name: "工作台", Icon: "dashboard", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "首页", Router: "workplace", Path: "workplace", Icon: "home"},
					{Name: "资产管理", Router: "inventory", Path: "inventory", Icon: "calendar"},
					{Name: "状态总览", Router: "overview", Path: "overview", Icon: "appstore"},
				}},
			{Router: "host", Path: "host", Name: "主机应用", Icon: "hdd", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "Linux主机", Router: "linux", Path: "linux", Icon: "container"},
					{Name: "Linux主机详情", Router: "linDetail", Path: "lindetail", Meta: Meta{Highlight: "/host", Invisible: true}},
					{Name: "Windows主机", Router: "windows", Path: "windows", Icon: "windows"},
					{Name: "Windows主机详情", Router: "winDetail", Path: "windetail", Meta: Meta{Highlight: "/host", Invisible: true}},
				}},
			{Router: "net", Path: "net", Name: "网络管理", Icon: "cloud", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "网络设备", Router: "netList", Path: "list", Icon: "chrome"},
					{Name: "设备详情", Router: "netDetail", Path: "detail", Meta: Meta{Highlight: "/net", Invisible: true}},
				}},
			{Router: "server", Path: "server", Name: "硬件管理", Icon: "database", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "物理服务器", Router: "srvList", Path: "list", Icon: "mobile"},
					{Name: "设备详情", Router: "srvDetail", Path: "detail", Meta: Meta{Highlight: "/server", Invisible: true}},
				}},
			{Router: "alarm", Path: "alarm", Name: "告警管理", Icon: "alert", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "告警分析", Router: "alarmAnalysis", Path: "analysis", Icon: "hourglass"},
					{Name: "告警查询", Router: "alarmList", Path: "list", Icon: "eye"},
					{Name: "告警分发", Router: "alarmRule", Path: "rule", Icon: "message"},
					{Name: "规则添加", Router: "alarmRuleAdd", Path: "rule-add", Meta: Meta{Highlight: "/alarm", Invisible: true}},
					{Name: "规则编辑", Router: "alarmRuleEdit", Path: "rule-edit", Meta: Meta{Highlight: "/alarm", Invisible: true}},
					{Name: "屏蔽规则", Router: "alarmMutes", Path: "mutes", Icon: "stop"},
				}},
			{Router: "topology", Path: "topology", Name: "拓扑管理", Icon: "picture", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "拓扑维护", Router: "topologyList", Path: "list", Icon: "environment"},
					{Name: "拓扑编辑", Router: "topologyDetail", Path: "detail", Meta: Meta{Highlight: "/topology", Invisible: true}},
					{Name: "拓扑展示", Router: "topologyShow", Path: "show", Meta: Meta{Highlight: "/topology", Invisible: true}},
				}},
			{Router: "report", Path: "report", Name: "报表管理", Icon: "file", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "流量报表", Router: "reportTraffic", Path: "list", Icon: "file-excel"},
					{Name: "报表编辑", Router: "reportTrafficEdit", Path: "edit", Meta: Meta{Highlight: "/report", Invisible: true}},
					{Name: "报表添加", Router: "reportTrafficAdd", Path: "add", Meta: Meta{Highlight: "/report", Invisible: true}},
				}},
			{Router: "system", Path: "system", Name: "系统管理", Icon: "setting", Meta: Meta{Page: Page{CacheAble: false}},
				Children: []TRouterChildren{
					{Name: "用户管理", Router: "systemUsers", Path: "users", Icon: "meh", Authority: Authority{Role: "user",
						Permission: "['add','edit','delete','update']"}},
				}},
		}
		tree := make([]RouterRes, 1)
		tree[0].Router = "root"
		tree[0].Children = prouter
		return tree, nil
	}
	return []RouterRes{}, nil
}

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
