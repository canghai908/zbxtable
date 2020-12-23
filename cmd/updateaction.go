package cmd

import (
	"encoding/json"
	"errors"
	"github.com/astaxie/beego/logs"
	"github.com/urfave/cli"
	"strings"
)

type MediaResult struct {
	MediaTypeID string `json:"mediatypeid"`
}
type ActionResult struct {
	ActionID string `json:"actionid"`
}

func GetMediaTypeID(value interface{}) (string, error) {
	p, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	var hb []MediaResult
	err = json.Unmarshal(p, &hb)
	if err != nil {
		return "", err
	}
	if len(hb) < 1 {
		return "", errors.New("MS-Agent Media not found")
	}
	return hb[0].MediaTypeID, nil
}
func GetActionID(value interface{}) (string, error) {
	p, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	var hb []ActionResult
	err = json.Unmarshal(p, &hb)
	if err != nil {
		return "", err
	}
	if len(hb) < 1 {
		return "", errors.New("MS-Agent Action not found")
	}
	return hb[0].ActionID, nil
}

var (
	// Install cli
	Ua = cli.Command{
		Name:        "ua",
		Usage:       "Update action",
		Description: "Update action",
		Action:      UpdateAction,
	}
)

func UpdateAction(c *cli.Context) {
	//login zabbix to get version
	version, err := LoginZabbix()
	if err != nil {
		logs.Error(err)
	}
	if strings.HasPrefix(version, "5") {
		//media get
		MediaParams1 := make(map[string]interface{}, 0)
		userpar := make(map[string]string)
		userpar["name"] = "MS-Agent Media"
		MediaParams1["filter"] = userpar
		MediaParams1["output"] = "mediatypeid"
		ma, err := API.CallWithError("mediatype.get", MediaParams1)
		if err != nil {
			logs.Error(err)
			return
		}
		//zabbix version 5
		MediaParams := make(map[string]interface{}, 0)
		MessageTemplates := make(map[int]interface{}, 1)
		MediaParams["message_templates"] = MessageTemplates
		mediatypeid, err := GetMediaTypeID(ma.Result)
		if err != nil {
			logs.Error(err)
			return
		}
		MediaParams["mediatypeid"] = mediatypeid
		Operations := make(map[string]interface{}, 0)
		Operations["eventsource"] = "0"
		Operations["recovery"] = "0"
		Operations["subject"] = "{TRIGGER.STATUS}"
		Operations["message"] = CreateEventTpl()
		RecoveryOperations := make(map[string]interface{}, 0)
		RecoveryOperations["eventsource"] = "0"
		RecoveryOperations["recovery"] = "1"
		RecoveryOperations["subject"] = "{TRIGGER.STATUS}"
		RecoveryOperations["message"] = CreateEventTpl()
		MessageTemplates[0] = Operations
		MessageTemplates[1] = RecoveryOperations
		_, err = API.CallWithError("mediatype.update", MediaParams)
		if err != nil {
			logs.Error(err)
			return
		}
		logs.Info("Update MS-Agent Media successed")
	} else {
		MediaParams1 := make(map[string]interface{}, 0)
		userpar := make(map[string]string)
		userpar["name"] = "MS-Agent Action"
		MediaParams1["filter"] = userpar
		MediaParams1["output"] = "actionid"
		ma, err := API.CallWithError("action.get", MediaParams1)
		if err != nil {
			logs.Error(err)
			return
		}
		actpara := make(map[string]interface{}, 0)
		actionid, err := GetActionID(ma.Result)
		if err != nil {
			logs.Error(err)
			return
		}
		actpara["actionid"] = actionid
		actpara["def_longdata"] = CreateEventTpl()
		actpara["def_shortdata"] = "{TRIGGER.STATUS}"
		actpara["r_longdata"] = CreateEventTpl()
		actpara["r_shortdata"] = "{TRIGGER.STATUS}"
		//action create
		_, err = API.CallWithError("action.update", actpara)
		if err != nil {
			logs.Error(err)
			return
		}
		logs.Info("Update MS-Agent Action successed")
	}
}
