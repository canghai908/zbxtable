package models

import (
	"github.com/astaxie/beego/logs"
	"strconv"
)

//GetTriggers get porblems
func GetTriggers() ([]EndTrigger, int64, error) {
	par11 := []string{"hostid", "name"}
	filter := make(map[string]string)
	filter["value"] = "1"
	triggers, err := API.CallWithError("trigger.get", Params{"output": "extend",
		"sortfield":       "lastchange",
		"sortorder":       "DESC",
		"selectHosts":     par11,
		"selectLastEvent": "extend",
		"filter":          filter,
		"maintenance":     false,
		"only_true":       true,
		"monitored":       true})
	if err != nil {
		logs.Debug(err)
		return []EndTrigger{}, 0, err
	}
	hba, err := json.Marshal(triggers.Result)
	if err != nil {
		logs.Debug(err)
		return []EndTrigger{}, 0, err
	}
	var hb []LastTriggers
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Debug(err)
		return []EndTrigger{}, 0, err
	}
	var bs EndTrigger
	var ma []EndTrigger
	for _, v := range hb {
		bs.Acknowledged = v.LastEvent.Acknowledged
		bs.Hostid = v.Hosts[0].Hostid
		bs.Name = v.Hosts[0].Name
		bs.Lastchange = v.Lastchange
		bs.LastEventName = v.LastEvent.Name
		bs.Severity = v.LastEvent.Severity
		bs.Eventid = v.LastEvent.Eventid
		bs.Objectid = v.LastEvent.Objectid
		ma = append(ma, bs)
	}
	return ma, int64(len(ma)), nil
}

//GetTriggerList get porblems
func GetTriggerList(hostid string) ([]TriggerListStr, int64, error) {
	//par11 := []string{"hostid", "name"}
	//filter := make(map[string]string)
	//filter["value"] = "1"
	triggers, err := API.CallWithError("trigger.get", Params{"output": "extend",
		"sortorder":         "DESC",
		"hostids":           hostid,
		"monitored":         true,
		"expandDescription": true})
	if err != nil {
		logs.Debug(err)
		return []TriggerListStr{}, 0, err
	}
	hba, err := json.Marshal(triggers.Result)
	if err != nil {
		logs.Debug(err)
		return []TriggerListStr{}, 0, err
	}
	var hb []TriggerListStr
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Debug(err)
		return []TriggerListStr{}, 0, err
	}
	return hb, int64(len(hb)), nil
}

//GetTriggerList get porblems
func GetTriggerHostCount(hostid string) (int64, error) {
	filter := make(map[string]string)
	filter["value"] = "1"
	triggers, err := API.CallWithError("trigger.get", Params{"output": "extend",
		"hostids":     hostid,
		"filter":      filter,
		"maintenance": false,
		"only_true":   true,
		"countOutput": true,
		"monitored":   true})
	if err != nil {
		logs.Debug(err)
		return 0, err
	}
	CountTrigger := triggers.Result.(string)
	count, err := strconv.ParseInt(CountTrigger, 10, 64)
	if err != nil {
		return 0, err
	}
	return count, nil
}

//GetTriggerList get porblems
func GetTriggerValue(triggerid string) ([]TriggerListStr, error) {
	OutputPar := []string{"value", "status", "state", "description"}
	triggers, err := API.CallWithError("trigger.get", Params{"output": OutputPar,
		"triggerids": triggerid})
	if err != nil {
		logs.Debug(err)
		return []TriggerListStr{}, err
	}
	hba, err := json.Marshal(triggers.Result)
	if err != nil {
		logs.Debug(err)
		return []TriggerListStr{}, err
	}
	var hb []TriggerListStr
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Debug(err)
		return []TriggerListStr{}, err
	}
	return hb, nil
}
