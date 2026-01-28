package models

import (
	"errors"
	"strings"
	"sync"
	"zbxtable/utils"
)

// last inserted Id on success.
func AddTopoData(m *TopologyData) (id int64, err error) {
	err = DB.Create(m).Error
	if err != nil {
		return 0, err
	}
	return int64(m.ID), err
}

// GetAllTopology t
func GetAllTopoData() (cnt int64, topodata []TopologyData, err error) {
	var topologys []TopologyData
	err = DB.Order("created_at DESC").Find(&topologys).Error
	if err != nil {
		utils.Log.Debug(err)
		return 0, []TopologyData{}, err
	}
	cnt = int64(len(topologys))
	return cnt, topologys, nil
}
func GetTriggerValueByTriggerID(TriggerID string) (value string, err error) {
	tri, err := GetTriggerValue(TriggerID)
	if err != nil || len(tri) == 0 {
		utils.Log.Debug(err)
		return "2", err
	}
	return tri[0].Value, nil
}

func GetFlowByFlowID(FLowID string) (flow string, err error) {
	p, err := GetItemByID(FLowID)
	if err != nil {
		utils.Log.Debug(err)
		return "", err
	}
	if len(p) == 0 {
		return "", errors.New("flow id is null")
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
			return "", err
		}
		flow = utils.FormatTraffic(NetItem[0].Lastvalue) + "/" + utils.FormatTraffic(p[0].Lastvalue)
		return flow, nil
	}
}

// host info
func GetHostInfoByID(hostid string, wg *sync.WaitGroup, info chan string) {
	defer wg.Done()
	p, err := GetHostInfoTopology(hostid)
	if err != nil {
		utils.Log.Debug(err)
	}
	StrP, err := json.Marshal(&p)
	if err != nil {
		utils.Log.Debug(err)
	}
	info <- string(StrP)
	return
}
