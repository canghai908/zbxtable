package models

import (
	"github.com/astaxie/beego/logs"
)

//TriggersRes rest
type TriggersRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []EndTrigger `json:"items"`
		Total int64        `json:"total"`
	} `json:"data"`
}

//LastTriggers struct
type LastTriggers struct {
	Comments        string `json:"comments"`
	CorrelationMode string `json:"correlation_mode"`
	CorrelationTag  string `json:"correlation_tag"`
	Description     string `json:"description"`
	Details         string `json:"details"`
	Error           string `json:"error"`
	Expression      string `json:"expression"`
	Flags           string `json:"flags"`
	Hosts           []struct {
		Hostid string `json:"hostid"`
		Name   string `json:"name"`
	} `json:"hosts"`
	LastEvent struct {
		Acknowledged string `json:"acknowledged"`
		Clock        string `json:"clock"`
		Eventid      string `json:"eventid"`
		Name         string `json:"name"`
		Ns           string `json:"ns"`
		Object       string `json:"object"`
		Objectid     string `json:"objectid"`
		Severity     string `json:"severity"`
		Source       string `json:"source"`
		Value        string `json:"value"`
	} `json:"lastEvent"`
	Lastchange         string `json:"lastchange"`
	ManualClose        string `json:"manual_close"`
	Priority           string `json:"priority"`
	RecoveryExpression string `json:"recovery_expression"`
	RecoveryMode       string `json:"recovery_mode"`
	State              string `json:"state"`
	Status             string `json:"status"`
	Templateid         string `json:"templateid"`
	Triggerid          string `json:"triggerid"`
	Type               string `json:"type"`
	URL                string `json:"url"`
	Value              string `json:"value"`
}

//EndTrigger struct
type EndTrigger struct {
	Acknowledged  string `json:"acknowledged"`
	Hostid        string `json:"hostid"`
	Name          string `json:"name"`
	Lastchange    string `json:"lastchange"`
	LastEventName string `json:"lasteventname"`
	Severity      string `json:"severity"`
	Eventid       string `json:"eventid"`
	Objectid      string `json:"objectid"`
}

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
		"maintenance":     "false",
		"only_true":       "true",
		"monitored":       "true"})
	if err != nil {
		logs.Error(err)
		return []EndTrigger{}, 0, err
	}
	hba, err := json.Marshal(triggers.Result)
	if err != nil {
		logs.Error(err)
		return []EndTrigger{}, 0, err
	}
	var hb []LastTriggers
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Error(err)
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
