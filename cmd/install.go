package cmd

import (
	"math/rand"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	zabbix "github.com/canghai908/zabbix-go"
	"github.com/urfave/cli"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

//RandStringRunes 随机密码生成
func RandStringRunes() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

var (
	// Install cli
	Install = cli.Command{
		Name:        "install",
		Usage:       "Install ms-agent tools to Zabbix Server",
		Description: "A tools to send alarm message to ZbxTable",
		Action:      installAagent,
	}
)

func installAagent(c *cli.Context) {
	var ZabbixAddress, ZabbixAdmin, ZabbixPasswd string

	ZabbixAdd := beego.AppConfig.String("zabbix_web")
	ZabbixAddress = ZabbixAdd + "/api_jsonrpc.php"
	ZabbixAdmin = beego.AppConfig.String("zabbix_user")
	ZabbixPasswd = beego.AppConfig.String("zabbix_pass")

	logs.Info("Zabbix API Address:", ZabbixAddress)
	logs.Info("Zabbix Admin User:", ZabbixAdmin)
	logs.Info("Zabbix Admin Password:", ZabbixPasswd)
	API := zabbix.NewAPI(ZabbixAddress)
	_, err := API.Login(ZabbixAdmin, ZabbixPasswd)
	if err != nil {
		logs.Error(err)
		return
	}
	logs.Info("登录zabbix平台成功!")
	version, err := API.Version()
	if err != nil {
		logs.Error(err)
		return
	}
	logs.Info("zabbix版本为:", version)
	//MediaParams mediatype new
	MediaParams := make(map[string]interface{}, 0)
	//对zabbix版本进行判断，5版本配置有所变化
	if strings.HasPrefix(version, "5") {
		//5版本配置
		MediaParams["description"] = "MS-Agent Media"
		MediaParams["name"] = "MS-Agent Media"
		MediaParams["type"] = "1"
		MediaParams["exec_params"] = "{ALERT.SENDTO}\n{ALERT.SUBJECT}\n{ALERT.MESSAGE}\n"
		MediaParams["exec_path"] = "ms-agent"
		MediaParams["maxattempts"] = "3"
		MediaParams["attempt_interval"] = "3s"
		MessageTemplates := make(map[int]interface{}, 1)
		MediaParams["message_templates"] = MessageTemplates
		Operations := make(map[string]interface{}, 0)
		Operations["eventsource"] = "0"
		Operations["recovery"] = "0"
		Operations["subject"] = "[{TRIGGER.SEVERITY}]服务器:{HOSTNAME1}发生:{TRIGGER.NAME}故障！"
		Operations["message"] = "告警主机: {HOSTNAME1}\n主机分组: {TRIGGER.HOSTGROUP.NAME}\n告警时间: {EVENT.DATE} {EVENT.TIME}\n告警等级: {TRIGGER.SEVERITY}\n告警信息: {TRIGGER.NAME}\n告警项目: {TRIGGER.KEY1}\n问题详情: {ITEM.NAME}:{ITEM.VALUE}\n当前状态: {TRIGGER.STATUS}\n事件ID: {EVENT.ID}"
		RecoveryOperations := make(map[string]interface{}, 0)
		RecoveryOperations["eventsource"] = "0"
		RecoveryOperations["recovery"] = "1"
		RecoveryOperations["subject"] = "[{TRIGGER.SEVERITY}]服务器:{HOSTNAME1}发生:{TRIGGER.NAME}恢复！"
		RecoveryOperations["message"] = "告警主机: {HOSTNAME1}\n主机分组: {TRIGGER.HOSTGROUP.NAME}\n告警时间: {EVENT.DATE} {EVENT.TIME}\n告警等级: {TRIGGER.SEVERITY}\n告警信息: {TRIGGER.NAME}\n告警项目: {TRIGGER.KEY1}\n问题详情: {ITEM.NAME}:{ITEM.VALUE}\n当前状态: {TRIGGER.STATUS}\n事件ID: {EVENT.ID}"
		MessageTemplates[0] = Operations
		MessageTemplates[1] = RecoveryOperations
	} else {
		//其他版本
		MediaParams["description"] = "MS-Agent Media"
		MediaParams["name"] = "MS-Agent Media"
		MediaParams["type"] = "1"
		MediaParams["exec_params"] = "{ALERT.SENDTO}\n{ALERT.SUBJECT}\n{ALERT.MESSAGE}\n"
		MediaParams["exec_path"] = "ms-agent"
	}
	ma, err := API.CallWithError("mediatype.create", MediaParams)
	if err != nil {
		logs.Error(err)
		return
	}
	result := ma.Result.(map[string]interface{})
	mediatypeids := result["mediatypeids"].([]interface{})
	//var mediaid string
	mediaid := mediatypeids[0].(string)
	logs.Info("创建告警媒介成功!")

	//GroupParams create usergroup
	GroupParams := make(map[string]interface{}, 0)
	GroupParams["name"] = "MS-Agent Group"
	group, err := API.CallWithError("usergroup.create", GroupParams)
	if err != nil {
		logs.Error(err)
		return
	}
	resgroup := group.Result.(map[string]interface{})
	usrgrpids := resgroup["usrgrpids"].([]interface{})
	groupid := usrgrpids[0].(string)
	logs.Info("创建告警用户组成功!")

	//create user
	userpara := make(map[string]interface{}, 0)
	usrgrps := make(map[string]string, 0)
	usermepara := make(map[string]string, 0)
	usrgrps["usrgrpid"] = groupid
	a := make(map[int]interface{})
	a[0] = usrgrps
	usermepara["mediatypeid"] = mediaid
	usermepara["sendto"] = "ms-agent"
	usermepara["active"] = "0"
	usermepara["severity"] = "63"
	usermepara["period"] = "1-7,00:00-24:00"
	b := make(map[int]interface{}, 0)
	b[0] = usermepara
	userpara["alias"] = "ms-agent"
	userpara["name"] = "ms-agent"
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
		return
	}
	resuser := user.Result.(map[string]interface{})
	userids := resuser["userids"].([]interface{})
	userid := userids[0].(string)
	logs.Info("创建告警用户成功!")
	logs.Info("用户名:ms-agent")
	logs.Info("密码:", tpasswdord)
	actpara := make(map[string]interface{}, 0)
	actpara["name"] = "MS-Agent Action"
	actpara["eventsource"] = "0"
	actpara["status"] = "0"
	actpara["esc_period"] = "60"
	actpara["def_longdata"] = "告警主机: {HOSTNAME1}\n主机分组: {TRIGGER.HOSTGROUP.NAME}\n告警时间: {EVENT.DATE} {EVENT.TIME}\n告警等级: {TRIGGER.SEVERITY}\n告警信息: {TRIGGER.NAME}\n告警项目: {TRIGGER.KEY1}\n问题详情: {ITEM.NAME}:{ITEM.VALUE}\n当前状态: {TRIGGER.STATUS}\n事件ID: {EVENT.ID}"
	actpara["def_shortdata"] = "[{TRIGGER.SEVERITY}]服务器:{HOSTNAME1}发生:{TRIGGER.NAME}故障！"
	actpara["recovery_msg"] = "1"
	actpara["r_longdata"] = "告警主机: {HOSTNAME1}\n主机分组: {TRIGGER.HOSTGROUP.NAME}\n告警时间: {EVENT.DATE} {EVENT.TIME}\n告警等级: {TRIGGER.SEVERITY}\n告警信息: {TRIGGER.NAME}\n告警项目: {TRIGGER.KEY1}\n问题详情: {ITEM.NAME}:{ITEM.VALUE}\n当前状态: {TRIGGER.STATUS}\n事件ID: {EVENT.ID}"
	actpara["r_shortdata"] = "[{TRIGGER.SEVERITY}]服务器:{HOSTNAME1}{TRIGGER.NAME}已恢复！"

	//operations
	operpara := make(map[string]interface{}, 0)
	operpara["operationtype"] = "0"
	use := make(map[string]string, 0)
	use["userid"] = userid
	v := make(map[int]interface{})
	v[0] = use
	opm := make(map[string]string, 0)
	opm["default_msg"] = "1"
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
	opm1["default_msg"] = "1"
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
		return
	}
	logs.Info("创建告警动作成功!")
	logs.Info("插件安装完成!")
}
