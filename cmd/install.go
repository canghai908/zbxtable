package cmd

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	zabbix "github.com/canghai908/zabbix-go"
	"github.com/urfave/cli/v2"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type EventTpl struct {
	HostsID      string `json:"host_id"`
	HostHost     string `json:"host_host"`
	Hostname     string `json:"hostname"`
	HostsIP      string `json:"host_ip"`
	HostGroup    string `json:"host_group"`
	EventTime    string `json:"event_time"`
	Severity     string `json:"severity"`
	TriggerID    string `json:"trigger_id"`
	TriggerName  string `json:"trigger_name"`
	TriggerKey   string `json:"trigger_key"`
	TriggerValue string `json:"trigger_value"`
	ItemID       string `json:"item_id"`
	ItemName     string `json:"item_name"`
	ItemValue    string `json:"item_value"`
	EventID      string `json:"event_id"`
}

func CreateEventTpl() string {
	var tpl = EventTpl{
		HostsID:      "{HOST.ID}",
		HostHost:     "{HOST.HOST}",
		Hostname:     "{HOST.NAME}",
		HostsIP:      "{HOST.IP}",
		HostGroup:    "{TRIGGER.HOSTGROUP.NAME}",
		EventTime:    "{EVENT.DATE} {EVENT.TIME}",
		Severity:     "{TRIGGER.NSEVERITY}",
		TriggerID:    "{TRIGGER.ID}",
		TriggerName:  "{TRIGGER.NAME}",
		TriggerKey:   "{TRIGGER.KEY}",
		TriggerValue: "{TRIGGER.VALUE}",
		ItemID:       "{ITEM.ID}",
		ItemName:     "{ITEM.NAME}",
		ItemValue:    "{ITEM.VALUE}",
		EventID:      "{EVENT.ID}",
	}
	TPl, _ := json.MarshalIndent(tpl, "", "    ")
	tp := strings.ReplaceAll(string(TPl), `"`, `¦`)
	return tp
}

//RandStringRunes 随机密码生成
func RandStringRunes() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

//const
var (
	letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	API         = &zabbix.API{}
)

//ms-aget const
var (
	MSName   = "ms-agent"
	MSUser   = "ms-agent"
	MSGroup  = "MS-Agent Group"
	MSMedia  = "MS-Agent Media"
	MSAction = "MS-Agent Action"
)

func LoginZabbix() (string, error) {
	var ZabbixAddress, ZabbixAdmin, ZabbixPasswd string

	ZabbixAdd := beego.AppConfig.String("zabbix_web")
	ZabbixAddress = ZabbixAdd + "/api_jsonrpc.php"
	ZabbixAdmin = beego.AppConfig.String("zabbix_user")
	ZabbixPasswd = beego.AppConfig.String("zabbix_pass")

	logs.Info("Zabbix API Address:", ZabbixAddress)
	logs.Info("Zabbix Admin User:", ZabbixAdmin)
	logs.Info("Zabbix Admin Password:", ZabbixPasswd)
	API = zabbix.NewAPI(ZabbixAddress)
	_, err := API.Login(ZabbixAdmin, ZabbixPasswd)
	if err != nil {
		logs.Error(err)
		return "", err
	}
	logs.Info("Login to zabbix successed")
	version, err := API.Version()
	if err != nil {
		logs.Error(err)
		return "", err
	}
	logs.Info("Zabbix version is", version)
	return version, nil
}

var (
	// Install cli
	Install = &cli.Command{
		Name:   "install",
		Usage:  "Install ms-agent tools to Zabbix Server",
		Action: installAagent,
	}
)

//installAagent
func installAagent(*cli.Context) error {
	//login zabbix to get version
	version, err := LoginZabbix()
	if err != nil {
		logs.Error(err)
		return err
	}
	MediaParams := make(map[string]interface{}, 0)
	MediaParams["description"] = MSMedia
	MediaParams["name"] = MSMedia
	MediaParams["type"] = "1"
	MediaParams["exec_params"] = "{ALERT.SENDTO}\n{ALERT.SUBJECT}\n{ALERT.MESSAGE}\n"
	MediaParams["exec_path"] = MSName
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
	userpara["alias"] = MSUser
	userpara["name"] = MSUser
	tpasswdord := RandStringRunes()
	userpara["passwd"] = tpasswdord
	//5.2版本 取消type字段,切换为roleid，roleid=2默认为管理员角色
	//5.2以下版本为管理员角色type=3，5.2以上版本roleid=3，为超级管理员角色
	if strings.HasPrefix(version, "5.2") {
		userpara["roleid"] = "3"
	} else {
		userpara["type"] = "3"
	}
	userpara["usrgrps"] = a
	userpara["user_medias"] = b
	user, err := API.CallWithError("user.create", userpara)
	if err != nil {
		logs.Error(err)
		return err
	}
	resuser := user.Result.(map[string]interface{})
	userids := resuser["userids"].([]interface{})
	userid := userids[0].(string)
	logs.Info("Create alarm user successfully!")
	logs.Info("Username :" + MSUser)
	logs.Info("Password :", tpasswdord)
	actpara := make(map[string]interface{}, 0)
	actpara["name"] = MSAction
	actpara["eventsource"] = "0"
	actpara["status"] = "0"
	actpara["esc_period"] = "60"
	actpara["def_longdata"] = CreateEventTpl()
	actpara["def_shortdata"] = "{TRIGGER.STATUS}"
	actpara["recovery_msg"] = "1"
	actpara["r_longdata"] = CreateEventTpl()
	actpara["r_shortdata"] = "{TRIGGER.STATUS}"

	//operations
	operpara := make(map[string]interface{}, 0)
	operpara["operationtype"] = "0"
	use := make(map[string]string, 0)
	use["userid"] = userid
	v := make(map[int]interface{})
	v[0] = use
	opm := make(map[string]string, 0)
	if strings.HasPrefix(version, "5") {
		opm["default_msg"] = "0"
		opm["subject"] = "{TRIGGER.STATUS}"
		opm["message"] = CreateEventTpl()
	} else {
		opm["default_msg"] = "1"
	}
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
	if strings.HasPrefix(version, "5") {
		opm1["default_msg"] = "0"
		opm1["subject"] = "{TRIGGER.STATUS}"
		opm1["message"] = CreateEventTpl()
	} else {
		opm1["default_msg"] = "1"
	}
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
	logs.Info("MS-Agent token is", beego.AppConfig.String("token"))
	return nil
}
