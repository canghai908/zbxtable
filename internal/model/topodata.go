package model

import (
	"errors"
	"strings"
	"sync"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"
)

func derivePeerTrafficItemKey(name, key string) (string, error) {
	switch {
	case strings.Contains(key, "net.if.out"):
		newKeyIn := strings.Replace(key, "net.if.out", "net.if.in", 1)
		return strings.ReplaceAll(newKeyIn, "Out", "In"), nil
	case strings.Contains(key, "net.if.in"):
		newKeyOut := strings.Replace(key, "net.if.in", "net.if.out", 1)
		return strings.ReplaceAll(newKeyOut, "In", "Out"), nil
	case strings.Contains(name, "Bits sent"):
		return strings.Replace(key, "Out", "In", 1), nil
	case strings.Contains(name, "Bits received"):
		return strings.Replace(key, "In", "Out", 1), nil
	default:
		return "", errors.New("unsupported traffic item key")
	}
}

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
		logger.Log.Debug(err)
		return 0, []TopologyData{}, err
	}
	cnt = int64(len(topologys))
	return cnt, topologys, nil
}
func GetTriggerValueByTriggerID(TriggerID string) (value string, err error) {
	tri, err := GetTriggerValue(TriggerID)
	if err != nil || len(tri) == 0 {
		logger.Log.Debug(err)
		return "2", err
	}
	return tri[0].Value, nil
}

// GetTriggerValueByTriggerIDFromInstance 从指定实例获取触发器状态
func GetTriggerValueByTriggerIDFromInstance(zid int, TriggerID string) (value string, err error) {
	if zid == 0 {
		// 如果没有指定实例，使用全局API
		return GetTriggerValueByTriggerID(TriggerID)
	}

	inst, err := GetAPIByZID(zid)
	if err != nil {
		logger.Log.Errorf("获取实例API失败: %v", err)
		return "2", err
	}

	OutputPar := []string{"value", "status", "state", "description"}
	triggers, err := inst.API.CallWithError("trigger.get", Params{"output": OutputPar,
		"triggerids": TriggerID})
	if err != nil {
		logger.Log.Debug(err)
		return "2", err
	}
	hba, err := json.Marshal(triggers.Result)
	if err != nil {
		logger.Log.Debug(err)
		return "2", err
	}
	var hb []TriggerListStr
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logger.Log.Debug(err)
		return "2", err
	}
	if len(hb) == 0 {
		return "2", errors.New("trigger not found")
	}
	return hb[0].Value, nil
}

func GetFlowByFlowID(FLowID string) (flow string, err error) {
	p, err := GetItemByID(FLowID)
	if err != nil {
		logger.Log.Debug(err)
		return "", err
	}
	if len(p) == 0 {
		return "", errors.New("flow id is null")
	} else {
		newItemKey, err := derivePeerTrafficItemKey(p[0].Name, p[0].Key)
		if err != nil {
			return "", err
		}
		netItem, err := GetItemByKey(p[0].Hostid, newItemKey)
		if err != nil {
			return "", err
		}
		if len(netItem) == 0 {
			return "", errors.New("net item not found")
		}
		flow = utils.FormatTraffic(netItem[0].Lastvalue) + "/" + utils.FormatTraffic(p[0].Lastvalue)
		return flow, nil
	}
}

// GetFlowByFlowIDFromInstance 从指定实例获取流量数据
func GetFlowByFlowIDFromInstance(zid int, FLowID string) (flow string, err error) {
	if zid == 0 {
		// 如果没有指定实例，使用全局API
		return GetFlowByFlowID(FLowID)
	}

	inst, err := GetAPIByZID(zid)
	if err != nil {
		logger.Log.Errorf("获取实例API失败: %v", err)
		return "", err
	}

	OutputPar := []string{"itemid", "name", "key_", "lastvalue", "hostid"}
	items, err := inst.API.CallWithError("item.get", Params{"output": OutputPar,
		"itemids": FLowID})
	if err != nil {
		logger.Log.Debug(err)
		return "", err
	}
	hba, err := json.Marshal(items.Result)
	if err != nil {
		logger.Log.Debug(err)
		return "", err
	}
	var p []Item
	err = json.Unmarshal(hba, &p)
	if err != nil {
		logger.Log.Debug(err)
		return "", err
	}
	if len(p) == 0 {
		return "", errors.New("flow id is null")
	}

	newItemKey, err := derivePeerTrafficItemKey(p[0].Name, p[0].Key)
	if err != nil {
		return "", err
	}

	netItems, err := inst.API.CallWithError("item.get", Params{
		"output":  OutputPar,
		"hostids": p[0].Hostid,
		"filter":  map[string]string{"key_": newItemKey},
	})
	if err != nil {
		logger.Log.Debug(err)
		return "", err
	}
	netItem, err := unmarshalItemsResult(netItems.Result)
	if err != nil {
		logger.Log.Debug(err)
		return "", err
	}
	if len(netItem) == 0 {
		netItems, err = inst.API.CallWithError("item.get", Params{
			"output":  OutputPar,
			"hostids": p[0].Hostid,
			"search":  Params{"key_": newItemKey},
		})
		if err != nil {
			logger.Log.Debug(err)
			return "", err
		}
		netItem, err = unmarshalItemsResult(netItems.Result)
		if err != nil {
			logger.Log.Debug(err)
			return "", err
		}
	}
	if len(netItem) == 0 {
		return "", errors.New("net item not found")
	}

	flow = utils.FormatTraffic(netItem[0].Lastvalue) + "/" + utils.FormatTraffic(p[0].Lastvalue)
	return flow, nil
}

// host info
func GetHostInfoByID(hostid string, wg *sync.WaitGroup, info chan string) {
	defer wg.Done()
	p, err := GetHostInfoTopology(hostid)
	if err != nil {
		logger.Log.Debug(err)
	}
	StrP, err := json.Marshal(&p)
	if err != nil {
		logger.Log.Debug(err)
	}
	info <- string(StrP)

}
