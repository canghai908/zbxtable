package cmd

import (
	"encoding/json"
	"errors"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego/logs"
	zabbix "github.com/canghai908/zabbix-go"
	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
)

// tpl t
type EventTpl struct {
	HostsID       string `json:"host_id"`
	HostHost      string `json:"host_host"`
	Hostname      string `json:"hostname"`
	HostsIP       string `json:"host_ip"`
	HostGroup     string `json:"host_group"`
	EventTime     string `json:"event_time"`
	Severity      string `json:"severity"`
	TriggerID     string `json:"trigger_id"`
	TriggerName   string `json:"trigger_name"`
	TriggerKey    string `json:"trigger_key"`
	TriggerValue  string `json:"trigger_value"`
	ItemID        string `json:"item_id"`
	ItemName      string `json:"item_name"`
	ItemValue     string `json:"item_value"`
	EventID       string `json:"event_id"`
	EventDuration string `json:"event_duration"`
}

func CreateProblemTpl() string {
	var tpl = EventTpl{
		HostsID:       "{HOST.ID}",
		HostHost:      "{HOST.HOST}",
		Hostname:      "{HOST.NAME}",
		HostsIP:       "{HOST.IP}",
		HostGroup:     "{TRIGGER.HOSTGROUP.NAME}",
		EventTime:     "{EVENT.DATE} {EVENT.TIME}",
		Severity:      "{TRIGGER.NSEVERITY}",
		TriggerID:     "{TRIGGER.ID}",
		TriggerName:   "{TRIGGER.NAME}",
		TriggerKey:    "{TRIGGER.KEY}",
		TriggerValue:  "{TRIGGER.VALUE}",
		ItemID:        "{ITEM.ID}",
		ItemName:      "{ITEM.NAME}",
		ItemValue:     "{ITEM.VALUE}",
		EventID:       "{EVENT.ID}",
		EventDuration: "{EVENT.DURATION}",
	}
	TPl, _ := json.MarshalIndent(tpl, "", "    ")
	tp := strings.ReplaceAll(string(TPl), `"`, `¦`)
	return tp
}
func CreateRecoveryTpl() string {
	var tpl = EventTpl{
		HostsID:       "{HOST.ID}",
		HostHost:      "{HOST.HOST}",
		Hostname:      "{HOST.NAME}",
		HostsIP:       "{HOST.IP}",
		HostGroup:     "{TRIGGER.HOSTGROUP.NAME}",
		EventTime:     "{EVENT.RECOVERY.DATE} {EVENT.RECOVERY.TIME}",
		Severity:      "{TRIGGER.NSEVERITY}",
		TriggerID:     "{TRIGGER.ID}",
		TriggerName:   "{TRIGGER.NAME}",
		TriggerKey:    "{TRIGGER.KEY}",
		TriggerValue:  "{TRIGGER.VALUE}",
		ItemID:        "{ITEM.ID}",
		ItemName:      "{ITEM.NAME}",
		ItemValue:     "{ITEM.VALUE}",
		EventID:       "{EVENT.ID}",
		EventDuration: "{EVENT.DURATION}",
	}
	TPl, _ := json.MarshalIndent(tpl, "", "    ")
	tp := strings.ReplaceAll(string(TPl), `"`, `¦`)
	return tp
}

// GetStrongPasswordString 强密码生成
func GetStrongPasswordString(l int) string {
	//~!@#$%^&*()_+{}":?><;.,
	str := "123456789ABCDEFGHIJKLMNPQRSTUVWXYZabcdefghijklmnpqrstuvwxyz!@#$%&*"
	bytes := []byte(str)
	result := []byte{}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, bytes[r.Intn(len(bytes))])
	}

	ok1, _ := regexp.MatchString(".[1|2|3|4|5|6|7|8|9]", string(result))
	ok2, _ := regexp.MatchString(".[Z|X|C|V|B|N|M|A|S|D|F|G|H|J|K|L|Q|W|E|R|T|Y|U|I|P]", string(result))
	ok3, _ := regexp.MatchString(".[z|x|c|v|b|n|m|a|s|d|f|g|h|j|k|l|q|w|e|r|t|y|u|i|p]", string(result))
	ok4, _ := regexp.MatchString(".[!|@|#|$|%|&|*]", string(result))
	if ok1 && ok2 && ok3 && ok4 {
		return string(result)
	} else {
		return GetStrongPasswordString(l)
	}
}

// ms-aget const
var (
	MSName   = "ms-agent"
	MSUser   = "ms-agent"
	MSGroup  = "MS-Agent Group"
	MSMedia  = "MS-Agent Media"
	MSAction = "MS-Agent Action"
)

var (
	// Install cli
	Install = &cli.Command{
		Name:   "install",
		Usage:  "Install ms-agent tools to Zabbix Server",
		Action: installAgent,
	}
	ZBV    bool
	ZBVVer string // Zabbix version string
)

func CheckZabbix() error {
	CheckConfExist()
	//login zabbix to get version
	version, err := CheckZabbixAPI(InitConfig("zabbix_web"), InitConfig("zabbix_user"),
		InitConfig("zabbix_pass"), InitConfig("zabbix_token"))
	if err != nil {
		logs.Error(err)
		return err
	}
	ZBVVer = version
	verArr := strings.Split(version, ".")
	ZbxMasterVer, _ := strconv.ParseInt(verArr[0], 10, 64)
	ZbxMiddleVer, _ := strconv.ParseInt(verArr[1], 10, 64)

	// 判断版本是否大于等于 7.0
	if ZbxMasterVer >= 7 || (ZbxMasterVer == 6 && ZbxMiddleVer >= 4) || (ZbxMasterVer == 5 && ZbxMiddleVer >= 4) {
		ZBV = true
	} else {
		ZBV = false
	}
	return nil
}

// installAagent
func installAgent(*cli.Context) error {
	err := CheckZabbix()
	if err != nil {
		return err
	}
	// 确保API对象已正确初始化并保持认证状态
	// 如果只配置了token，需要重新设置API.Auth
	token := InitConfig("zabbix_token")
	if token != "" {
		// 重新初始化API并设置token，确保认证状态正确
		zabbixWeb := InitConfig("zabbix_web")
		API = zabbix.NewAPI(zabbixWeb + "/api_jsonrpc.php")
		API.Auth = token
	}
	MediaParams := make(map[string]interface{}, 0)
	MediaParams["description"] = MSMedia
	MediaParams["name"] = MSMedia
	MediaParams["type"] = "1"
	MediaParams["status"] = "0"

	// 根据 Zabbix 版本设置不同的参数格式
	if ZBV {
		parameters := []map[string]interface{}{
			{
				"sortorder": "0",
				"value":     "{ALERT.SENDTO}",
			},
			{
				"sortorder": "1",
				"value":     "{ALERT.SUBJECT}",
			},
			{
				"sortorder": "2",
				"value":     "{ALERT.MESSAGE}",
			},
		}
		MediaParams["parameters"] = parameters
	} else {
		MediaParams["exec_params"] = "{ALERT.SENDTO}\n{ALERT.SUBJECT}\n{ALERT.MESSAGE}\n"
	}

	MediaParams["exec_path"] = MSName
	MediaParams["status"] = "0"
	ma, err := API.CallWithError("mediatype.create", MediaParams)
	if err != nil {
		logs.Error(err)
		return err
	}
	result := ma.Result.(map[string]interface{})
	mediatypeids := result["mediatypeids"].([]interface{})
	//var mediaid string
	mediaid := mediatypeids[0].(string)
	logs.Info("Create media type successfully!")
	//GroupParams create usergroup
	GroupParams := make(map[string]interface{}, 0)
	GroupParams["name"] = MSGroup
	group, err := API.CallWithError("usergroup.create", GroupParams)
	if err != nil {
		logs.Error(err)
		return err
	}
	resgroup := group.Result.(map[string]interface{})
	usrgrpids := resgroup["usrgrpids"].([]interface{})
	groupid := usrgrpids[0].(string)
	logs.Info("Create user group successfully!")
	//create user
	userpara := make(map[string]interface{}, 0)
	usrgrps := make(map[string]string, 0)
	usermepara := make(map[string]string, 0)
	usrgrps["usrgrpid"] = groupid
	a := make(map[int]interface{})
	a[0] = usrgrps
	usermepara["mediatypeid"] = mediaid
	usermepara["sendto"] = "v2"
	usermepara["active"] = "0"
	usermepara["severity"] = "63"
	usermepara["period"] = "1-7,00:00-24:00"
	b := make(map[int]interface{}, 0)
	b[0] = usermepara
	if ZBV {
		userpara["username"] = MSUser
	} else {
		userpara["alias"] = MSUser
		userpara["name"] = MSUser
	}
	//10位随机密码生成
	tPassword := GetStrongPasswordString(10)
	userpara["passwd"] = tPassword
	//5.2版本 取消type字段,切换为roleid，roleid=2默认为管理员角色
	//5.2以下版本为管理员角色type=3，5.2以上版本roleid=3，为超级管理员角色
	if ZBV {
		userpara["roleid"] = "3"
	} else {
		userpara["type"] = "3"
	}
	userpara["usrgrps"] = a
	if ZBV {
		userpara["medias"] = b
	} else {
		userpara["user_medias"] = b
	}
	user, err := API.CallWithError("user.create", userpara)
	if err != nil {
		logs.Error(err)
		return err
	}
	resuser := user.Result.(map[string]interface{})
	userids := resuser["userids"].([]interface{})
	userid := userids[0].(string)
	logs.Info("Create alarm user successfully!")
	logs.Info("Username :", MSUser)
	logs.Info("Password :", tPassword)
	actpara := make(map[string]interface{}, 0)
	actpara["name"] = MSAction
	actpara["eventsource"] = "0"
	actpara["status"] = "0"
	actpara["esc_period"] = "60"
	if !ZBV {
		actpara["def_shortdata"] = "{TRIGGER.STATUS}"
		actpara["def_longdata"] = CreateProblemTpl()
		actpara["r_shortdata"] = "{TRIGGER.STATUS}"
		actpara["r_longdata"] = CreateRecoveryTpl()
		actpara["recovery_msg"] = "1"
	}
	//operations
	operpara := make(map[string]interface{}, 0)
	operpara["operationtype"] = "0"
	use := make(map[string]string, 0)
	use["userid"] = userid
	v := make(map[int]interface{})
	v[0] = use
	opm := make(map[string]string, 0)
	//新版本
	opm["default_msg"] = "0"
	opm["subject"] = "{TRIGGER.STATUS}"
	opm["message"] = CreateProblemTpl()
	opm["mediatypeid"] = mediaid
	operpara["opmessage_usr"] = v
	operpara["opmessage"] = opm

	//recovery_operations
	recovpara := make(map[string]interface{}, 0)
	recovpara["operationtype"] = "0"
	use2 := make(map[string]string, 0)
	use2["userid"] = userid
	v2 := make(map[int]interface{})
	v2[0] = use2
	opm1 := make(map[string]string, 0)
	opm1["default_msg"] = "0"
	opm1["subject"] = "{TRIGGER.STATUS}"
	opm1["message"] = CreateRecoveryTpl()
	opm1["mediatypeid"] = mediaid
	recovpara["opmessage_usr"] = v2
	recovpara["opmessage"] = opm1

	//two operations
	reinter := make(map[int]interface{}, 0)
	reinter[0] = operpara
	actpara["operations"] = reinter
	reinter1 := make(map[int]interface{}, 0)
	reinter1[0] = recovpara
	actpara["recovery_operations"] = reinter1

	//action create
	_, err = API.CallWithError("action.create", actpara)
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Info("Create alarm action successfully!")
	logs.Info("MS-Agent plugin configured successfully!")
	logs.Info("MS-Agent token is", strings.ReplaceAll(uuid.New().String(), "-", ""))
	return nil
}

// GetActionID by actionname
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

// UserGrpInfo st
type UserInfo struct {
	Mediatypes []struct {
		Mediatypeid string `json:"mediatypeid"`
	} `json:"mediatypes"`
	Userid  string `json:"userid"`
	Usrgrps []struct {
		Usrgrpid string `json:"usrgrpid"`
	} `json:"usrgrps"`
}

// GetUserID by username
func GetUserID(Username string) (userinfo UserInfo, err error) {
	Params := make(map[string]interface{}, 0)
	FilterParams := make(map[string]string)
	if ZBV {
		FilterParams["username"] = Username
	} else {
		FilterParams["name"] = Username
	}
	Params["selectUsrgrps"] = []string{"usrgrpid"}
	Params["selectMediatypes"] = []string{"mediatypeid"}
	Params["filter"] = FilterParams
	Params["output"] = []string{"userid,usrgrps"}
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
