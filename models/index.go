package models

import (
	"fmt"
	"strconv"

	"github.com/astaxie/beego/logs"
)

func getCountByType(hostType string) (int64, error) {
	inventoryParams := make(map[string]string)
	inventoryParams["type"] = hostType
	hostCount, err := API.CallWithError("host.get", Params{
		"output":          "extend",
		"searchInventory": inventoryParams,
		"countOutput":     true})
	if err != nil {
		return 0, err
	}
	count, _ := strconv.ParseInt(hostCount.Result.(string), 10, 64)
	return count, nil
}

func GetCountHost() (IndexInfo, error) {
	d := IndexInfo{}
	var err error
	//d.Hosts, err = getCountByType("host.get")
	//d.Items, err = getCountByType("item.get")
	//d.Problems, err = getCountByType("problem.get")
	//d.Triggers, err = getCountByType("trigger.get")
	d.LinCount, err = getCountByType("VM_LIN")
	if err != nil {
		return d, err
	}
	d.WinCount, err = getCountByType("VM_WIN")
	if err != nil {
		return d, err
	}
	d.SRVCount, err = getCountByType("HW_SRV")
	if err != nil {
		return d, err
	}
	d.NETCount, err = getCountByType("HW_NET")
	if err != nil {
		return IndexInfo{}, err
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
	ret, err := CacheZRevRangeWithScores(MetType1+"_"+MetType2, 0, top_n)
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
	listmap := make(map[string][]Hosts)
	for _, v := range list {
		var ArrayOne []Hosts
		p, err := CacheGet(v + "_OVERVIEW")
		if err != nil || p == "" {
			logs.Error(err)
			continue
		}
		err = json.Unmarshal([]byte(p), &ArrayOne)
		if err != nil {
			logs.Error(err)
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

func GetZbxSession() (string, error) {
	p, err := CacheGet("zbx_session")
	if err != nil || p == "" {
		logs.Error(err)
		return "", err
	}
	return p, nil
}
