package cmd

import (
	"encoding/json"
	"errors"
	"github.com/astaxie/beego/logs"
	"github.com/urfave/cli/v2"
	"strings"
)

var (
	// Install cli
	Ua = &cli.Command{
		Name:   "ua",
		Usage:  "Update action",
		Action: UpdateAction,
	}
)

//UserGrpInfo st
type UserInfo struct {
	Mediatypes []struct {
		Mediatypeid string `json:"mediatypeid"`
	} `json:"mediatypes"`
	Userid  string `json:"userid"`
	Usrgrps []struct {
		Usrgrpid string `json:"usrgrpid"`
	} `json:"usrgrps"`
}

//GetUserID by username
func GetUserID(Username string) (userinfo UserInfo, err error) {
	Params := make(map[string]interface{}, 0)
	FilterParams := make(map[string]string)
	FilterParams["name"] = Username
	Params["selectUsrgrps"] = "usrgrpid"
	Params["selectMediatypes"] = "mediatypeid"
	Params["filter"] = FilterParams
	Params["output"] = "userid,usrgrps"
	res, err := API.CallWithError("user.get", Params)
	if err != nil {
		return UserInfo{}, err
	}
	p, err := json.Marshal(res.Result)
	if err != nil {
		return UserInfo{}, err
	}
	if len(string(p)) < 3 {
		return UserInfo{}, errors.New("User : " + Username + " not found")
	}
	var hb []UserInfo
	err = json.Unmarshal(p, &hb)
	if err != nil {
		return UserInfo{}, err
	}
	return hb[0], nil
}

//GetMediaTypeID by medianame
func GetMediaTypeID(MediaName string) (string, error) {
	Params := make(map[string]interface{}, 0)
	FilterParams := make(map[string]string)
	FilterParams["name"] = MediaName
	FilterParams["description"] = MediaName
	Params["filter"] = FilterParams
	Params["output"] = "mediatypeid"
	res, err := API.CallWithError("mediatype.get", Params)
	if err != nil {
		return "", err
	}
	p, err := json.Marshal(res.Result)
	if err != nil {
		return "", err
	}
	type MediaResult struct {
		MediaTypeID string `json:"mediatypeid"`
	}

	var hb []MediaResult
	err = json.Unmarshal(p, &hb)
	if err != nil {
		return "", err
	}
	if len(hb) < 1 {
		return "", errors.New("Mediatype :" + MediaName + " not found")
	}
	return hb[0].MediaTypeID, nil
}

//GetActionID by actionname
func GetActionID(ActionName string) (string, error) {
	//action get
	Params := make(map[string]interface{}, 0)
	FilterParams := make(map[string]string)
	FilterParams["name"] = ActionName
	Params["filter"] = FilterParams
	Params["output"] = "actionid"
	res, err := API.CallWithError("action.get", Params)
	if err != nil {
		logs.Error(err)
		return "", err
	}
	p, err := json.Marshal(res.Result)
	if err != nil {
		return "", err
	}
	type ActionResult struct {
		ActionID string `json:"actionid"`
	}
	var hb []ActionResult
	err = json.Unmarshal(p, &hb)
	if err != nil {
		return "", err
	}
	if len(hb) < 1 {
		return "", errors.New("Action :" + ActionName + " not found")
	}
	return hb[0].ActionID, nil
}

func UpdateAction(*cli.Context) error {
	//CheckConfExist
	CheckConfExist()
	//login zabbix to get version
	version, err := CheckZabbixAPI(InitConfig("zabbix_web"), InitConfig("zabbix_user"), InitConfig("zabbix_pass"))
	if err != nil {
		logs.Error(err)
		return err
	}
	//mediatypeid
	mediatypeid, err := GetMediaTypeID(MSMedia)
	if err != nil {
		logs.Error(err)
		return err
	}
	//actionid
	actionid, err := GetActionID(MSAction)
	if err != nil {
		logs.Error(err)
		return err
	}
	//userid get
	userid, err := GetUserID(MSName)
	if err != nil {
		logs.Error(err)
		return err
	}
	ActtionParams := make(map[string]interface{}, 0)
	ActtionParams["actionid"] = actionid
	OperationsParams := make(map[string]interface{}, 0)
	OpeationDetailParams := make(map[string]string, 0)
	if strings.HasPrefix(version, "5") {
		OpeationDetailParams["default_msg"] = "0"
		OpeationDetailParams["subject"] = "{TRIGGER.STATUS}"
		OpeationDetailParams["message"] = CreateEventTpl()
	} else {
		ActtionParams["def_longdata"] = CreateEventTpl()
		ActtionParams["def_shortdata"] = "{TRIGGER.STATUS}"
		ActtionParams["r_longdata"] = CreateEventTpl()
		ActtionParams["r_shortdata"] = "{TRIGGER.STATUS}"
		OpeationDetailParams["default_msg"] = "1"
	}
	logs.Info(mediatypeid)
	OpeationDetailParams["mediatypeid"] = mediatypeid
	usemap := make(map[string]string, 0)
	usemap["userid"] = userid.Userid
	v := make(map[int]interface{})
	v[0] = usemap
	OperationsParams["opmessage_usr"] = v
	OperationsParams["operationtype"] = "0"
	OperationsParams["opmessage"] = OpeationDetailParams
	//two operations
	OperationsMap := make(map[int]interface{}, 0)
	OperationsMap[0] = OperationsParams
	ActtionParams["operations"] = OperationsMap
	RecoverOerationsMap := make(map[int]interface{}, 0)
	RecoverOerationsMap[0] = OperationsParams
	ActtionParams["recovery_operations"] = RecoverOerationsMap
	//action update
	_, err = API.CallWithError("action.update", ActtionParams)
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Info("Update MS-Agent Action successed")
	return nil
}
