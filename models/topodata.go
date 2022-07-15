package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"strings"
	"sync"
	"zbxtable/utils"
)

// last inserted Id on success.
func AddTopoData(m *TopologyData) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	if err != nil {
		return 0, err
	}
	return id, err
}

// GetAllTopology t
func GetAllTopoData() (cnt int64, topodata []TopologyData, err error) {
	o := orm.NewOrm()
	var topologys []TopologyData
	al := new(TopologyData)
	//count topology
	_, err = o.QueryTable(al).OrderBy("-created_at").All(&topologys)
	if err != nil {
		logs.Debug(err)
		return 0, []TopologyData{}, err
	}
	cnt = int64(len(topologys))
	return cnt, topologys, nil
}

func GetTriggerValueByTriggerID(TriggerID string, wg *sync.WaitGroup, trigger chan string) {
	defer wg.Done()
	tri, err := GetTriggerValue(TriggerID)
	if err != nil || len(tri) == 0 {
		logs.Debug(err)
		trigger <- "2"
		return
	} else {
		trigger <- tri[0].Value
	}
}
func GetFlowByFlowID(FLowID string, wg *sync.WaitGroup, flow chan string) {
	defer wg.Done()
	p, err := GetItemByID(FLowID)
	if err != nil {
		logs.Debug(err)
		flow <- ""
		return
	}
	if len(p) == 0 {
		flow <- ""
		return
	} else {
		//获取数据，判断是否为流量接口
		var NewItemKey string
		if strings.Contains(p[0].Name, "Bits sent") {
			NewkeyIN := strings.Replace(p[0].Key, "net.if.out", "net.if.in", -1)
			NewItemKey = strings.Replace(NewkeyIN, "Out", "In", -1)
		}
		if strings.Contains(p[0].Name, "Bits received") {
			NewkeyOut := strings.Replace(p[0].Key, "net.if.in", "net.if.out", -1)
			NewItemKey = strings.Replace(NewkeyOut, "In", "Out", -1)
		}
		NetItem, err := GetItemByKey(p[0].Hostid, NewItemKey)
		if err != nil {
			flow <- ""
		} else {
			flow <- utils.FormatTraffic(NetItem[0].Lastvalue) + "/" + utils.FormatTraffic(p[0].Lastvalue)
		}
	}
}

//host info
func GetHostInfoByID(hostid string, wg *sync.WaitGroup, info chan string) {
	defer wg.Done()
	p, err := GetHostInfoTopology(hostid)
	if err != nil {
		logs.Debug(err)
	}
	StrP, err := json.Marshal(&p)
	if err != nil {
		logs.Debug(err)
	}
	info <- string(StrP)
}
