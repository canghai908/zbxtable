package model

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/logger"

	zabbix "github.com/canghai908/zabbix-go"
	"github.com/google/uuid"
)

// MSAgentConfig MS-Agent 配置信息
type MSAgentConfig struct {
	ZbxTableURL  string `json:"zbxtable_url"`  // ZbxTable 服务地址
	Instance     string `json:"instance"`      // 实例 ID
	WebhookToken string `json:"webhook_token"` // Webhook 认证 Token
}

// MSAgentInstallScript MS-Agent 安装脚本
type MSAgentInstallScript struct {
	CurlCommand    string `json:"curl_command"`    // curl 安装命令
	InstallCommand string `json:"install_command"` // 完整安装命令
	ConfigContent  string `json:"config_content"`  // 配置文件内容
}

// CreateProblemTpl 创建问题模板
func CreateProblemTpl() string {
	var tpl = EventTpl{
		HostsID:       "{HOST.ID}",
		HostHost:      "{HOST.HOST}",
		Hostname:      "{HOST.NAME}",
		HostsIP:       "{HOST.IP}",
		HostGroup:     "{TRIGGER.HOSTGROUP.NAME}",
		EventTime:     "{EVENT.DATE} {EVENT.TIME}",
		Severity:      "{TRIGGER.NSEVERITY}",
		TriggerID:     0,
		TriggerName:   "{TRIGGER.NAME}",
		TriggerKey:    "{TRIGGER.KEY}",
		TriggerValue:  "{TRIGGER.VALUE}",
		ItemID:        0,
		ItemName:      "{ITEM.NAME}",
		ItemValue:     "{ITEM.VALUE}",
		EventID:       0,
		EventDuration: "{EVENT.DURATION}",
	}
	TPl, _ := json.MarshalIndent(tpl, "", "    ")
	tp := strings.ReplaceAll(string(TPl), `"`, `¦`)
	// 替换占位符
	tp = strings.ReplaceAll(tp, `¦trigger_id¦: 0`, `¦trigger_id¦: ¦{TRIGGER.ID}¦`)
	tp = strings.ReplaceAll(tp, `¦item_id¦: 0`, `¦item_id¦: ¦{ITEM.ID}¦`)
	tp = strings.ReplaceAll(tp, `¦event_id¦: 0`, `¦event_id¦: ¦{EVENT.ID}¦`)
	return tp
}

// CreateRecoveryTpl 创建恢复模板
func CreateRecoveryTpl() string {
	var tpl = EventTpl{
		HostsID:       "{HOST.ID}",
		HostHost:      "{HOST.HOST}",
		Hostname:      "{HOST.NAME}",
		HostsIP:       "{HOST.IP}",
		HostGroup:     "{TRIGGER.HOSTGROUP.NAME}",
		EventTime:     "{EVENT.RECOVERY.DATE} {EVENT.RECOVERY.TIME}",
		Severity:      "{TRIGGER.NSEVERITY}",
		TriggerID:     0,
		TriggerName:   "{TRIGGER.NAME}",
		TriggerKey:    "{TRIGGER.KEY}",
		TriggerValue:  "{TRIGGER.VALUE}",
		ItemID:        0,
		ItemName:      "{ITEM.NAME}",
		ItemValue:     "{ITEM.VALUE}",
		EventID:       0,
		EventDuration: "{EVENT.DURATION}",
	}
	TPl, _ := json.MarshalIndent(tpl, "", "    ")
	tp := strings.ReplaceAll(string(TPl), `"`, `¦`)
	// 替换占位符
	tp = strings.ReplaceAll(tp, `¦trigger_id¦: 0`, `¦trigger_id¦: ¦{TRIGGER.ID}¦`)
	tp = strings.ReplaceAll(tp, `¦item_id¦: 0`, `¦item_id¦: ¦{ITEM.ID}¦`)
	tp = strings.ReplaceAll(tp, `¦event_id¦: 0`, `¦event_id¦: ¦{EVENT.ID}¦`)
	return tp
}

// GetStrongPasswordString 生成强密码
func GetStrongPasswordString(l int) string {
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
	}
	return GetStrongPasswordString(l)
}

// GenerateMSAgentInstallScript 生成 MS-Agent 安装脚本
func GenerateMSAgentInstallScript(config *MSAgentConfig) (*MSAgentInstallScript, error) {
	if config == nil {
		return nil, errors.New("config is nil")
	}

	// 生成配置文件内容
	configContent := fmt.Sprintf(`# MS-Agent Configuration
# ZbxTable Server URL
zbxtable_url: %s

# Tenant ID
tenant_id: %s

# Authentication Token
webhook_token: %s

# Log Level (debug, info, warn, error)
log_level: info

# Log File Path
log_path: /var/log/ms-agent/ms-agent.log
`, config.ZbxTableURL, config.Instance, config.WebhookToken)

	// 生成 curl 下载和安装命令
	curlCommand := `curl -fsSL https://raw.githubusercontent.com/canghai908/ms-agent/main/install.sh | bash`

	// 生成完整的安装命令（包含配置）
	installCommand := fmt.Sprintf(`#!/bin/bash
# MS-Agent 自动安装脚本
# 生成时间: %s

set -e

echo "开始安装 MS-Agent..."

# 下载并安装 MS-Agent
curl -fsSL https://raw.githubusercontent.com/canghai908/ms-agent/main/install.sh | bash

# 创建配置目录
mkdir -p /etc/ms-agent

# 写入配置文件
cat > /etc/ms-agent/config.yml << 'EOF'
%s
EOF

# 重启 MS-Agent 服务
systemctl restart ms-agent
systemctl enable ms-agent

echo "MS-Agent 安装完成！"
echo "配置文件: /etc/ms-agent/config.yml"
echo "日志文件: /var/log/ms-agent/ms-agent.log"
echo "查看状态: systemctl status ms-agent"
echo "查看日志: tail -f /var/log/ms-agent/ms-agent.log"
`, time.Now().Format("2006-01-02 15:04:05"), configContent)

	return &MSAgentInstallScript{
		CurlCommand:    curlCommand,
		InstallCommand: installCommand,
		ConfigContent:  configContent,
	}, nil
}

// InstallMSAgentToZabbix 在 Zabbix 中安装 MS-Agent 配置
func InstallMSAgentToZabbix(zid int) error {
	// 获取租户信息
	instance, err := GetZabbixInstanceByZID(zid)
	if err != nil {
		return fmt.Errorf("获取租户失败: %w", err)
	}

	// 解密密码和Token
	decryptedPass, decryptedToken := DecryptInstanceCredentials(instance)

	// 初始化 Zabbix API
	api := zabbix.NewAPI(instance.URL + "/api_jsonrpc.php")
	if decryptedToken != "" {
		api.Auth = decryptedToken
	} else {
		_, err := api.Login(instance.User, decryptedPass)
		if err != nil {
			return fmt.Errorf("登录 Zabbix 失败: %w", err)
		}
	}

	// 检查 Zabbix 版本
	version, err := api.Version()
	if err != nil {
		return fmt.Errorf("获取 Zabbix 版本失败: %w", err)
	}

	verArr := strings.Split(version, ".")
	zbxMasterVer, _ := strconv.ParseInt(verArr[0], 10, 64)
	zbxMiddleVer, _ := strconv.ParseInt(verArr[1], 10, 64)
	isNewVersion := zbxMasterVer >= 7 || (zbxMasterVer == 6 && zbxMiddleVer >= 4) || (zbxMasterVer == 5 && zbxMiddleVer >= 4)

	// 准备 Media Type 参数
	mediaParams := make(map[string]interface{})
	mediaParams["description"] = MSMedia
	mediaParams["name"] = MSMedia
	mediaParams["type"] = "1"
	mediaParams["status"] = "0"

	if isNewVersion {
		parameters := []map[string]interface{}{
			{"sortorder": "0", "value": "{ALERT.SENDTO}"},
			{"sortorder": "1", "value": "{ALERT.SUBJECT}"},
			{"sortorder": "2", "value": "{ALERT.MESSAGE}"},
		}
		mediaParams["parameters"] = parameters
	} else {
		mediaParams["exec_params"] = "{ALERT.SENDTO}\n{ALERT.SUBJECT}\n{ALERT.MESSAGE}\n"
	}

	mediaParams["exec_path"] = MSName

	// 1. 创建或更新 Media Type
	logger.Log.Info("检查 Media Type...")

	// 先查询是否已存在
	getMediaParams := map[string]interface{}{
		"output": "extend",
		"filter": map[string]string{
			"name": MSMedia,
		},
	}

	existingMedia, err := api.CallWithError("mediatype.get", getMediaParams)
	var mediaid string

	if err == nil && existingMedia.Result != nil {
		resultArray, ok := existingMedia.Result.([]interface{})
		if ok && len(resultArray) > 0 {
			// Media Type 已存在，更新它
			existingMediaMap := resultArray[0].(map[string]interface{})
			mediaid = existingMediaMap["mediatypeid"].(string)
			logger.Log.Info("Media Type 已存在, ID:", mediaid, "，正在更新...")

			mediaParams["mediatypeid"] = mediaid
			_, err = api.CallWithError("mediatype.update", mediaParams)
			if err != nil {
				return fmt.Errorf("更新 Media Type 失败: %w", err)
			}
			logger.Log.Info("Media Type 更新成功")
		} else {
			// 不存在，创建新的
			logger.Log.Info("创建 Media Type...")
			ma, err := api.CallWithError("mediatype.create", mediaParams)
			if err != nil {
				return fmt.Errorf("创建 Media Type 失败: %w", err)
			}
			result := ma.Result.(map[string]interface{})
			mediatypeids := result["mediatypeids"].([]interface{})
			mediaid = mediatypeids[0].(string)
			logger.Log.Info("Media Type 创建成功, ID:", mediaid)
		}
	} else {
		// 查询失败，尝试创建
		logger.Log.Info("创建 Media Type...")
		ma, err := api.CallWithError("mediatype.create", mediaParams)
		if err != nil {
			return fmt.Errorf("创建 Media Type 失败: %w", err)
		}
		result := ma.Result.(map[string]interface{})
		mediatypeids := result["mediatypeids"].([]interface{})
		mediaid = mediatypeids[0].(string)
		logger.Log.Info("Media Type 创建成功, ID:", mediaid)
	}

	// 2. 创建用户组
	logger.Log.Info("创建用户组...")
	groupParams := make(map[string]interface{})
	groupParams["name"] = MSGroup
	group, err := api.CallWithError("usergroup.create", groupParams)
	if err != nil {
		return fmt.Errorf("创建用户组失败: %w", err)
	}

	resgroup := group.Result.(map[string]interface{})
	usrgrpids := resgroup["usrgrpids"].([]interface{})
	groupid := usrgrpids[0].(string)
	logger.Log.Info("用户组创建成功, ID:", groupid)

	// 3. 创建用户
	logger.Log.Info("创建用户...")
	userpara := make(map[string]interface{})
	usrgrps := make(map[string]string)
	usermepara := make(map[string]string)
	usrgrps["usrgrpid"] = groupid
	a := make(map[int]interface{})
	a[0] = usrgrps
	usermepara["mediatypeid"] = mediaid
	usermepara["sendto"] = "v2"
	usermepara["active"] = "0"
	usermepara["severity"] = "63"
	usermepara["period"] = "1-7,00:00-24:00"
	b := make(map[int]interface{})
	b[0] = usermepara

	if isNewVersion {
		userpara["username"] = MSUser
	} else {
		userpara["alias"] = MSUser
		userpara["name"] = MSUser
	}

	tPassword := GetStrongPasswordString(10)
	userpara["passwd"] = tPassword

	if isNewVersion {
		userpara["roleid"] = "3"
	} else {
		userpara["type"] = "3"
	}

	userpara["usrgrps"] = a
	if isNewVersion {
		userpara["medias"] = b
	} else {
		userpara["user_medias"] = b
	}

	user, err := api.CallWithError("user.create", userpara)
	if err != nil {
		return fmt.Errorf("创建用户失败: %w", err)
	}

	resuser := user.Result.(map[string]interface{})
	userids := resuser["userids"].([]interface{})
	userid := userids[0].(string)
	logger.Log.Infof("用户创建成功, 用户名: %s, 密码: %s, ID: %s", MSUser, tPassword, userid)

	// 4. 创建 Action
	logger.Log.Info("创建 Action...")
	actpara := make(map[string]interface{})
	actpara["name"] = MSAction
	actpara["eventsource"] = "0"
	actpara["status"] = "0"
	actpara["esc_period"] = "60"

	if !isNewVersion {
		actpara["def_shortdata"] = "{TRIGGER.STATUS}"
		actpara["def_longdata"] = CreateProblemTpl()
		actpara["r_shortdata"] = "{TRIGGER.STATUS}"
		actpara["r_longdata"] = CreateRecoveryTpl()
		actpara["recovery_msg"] = "1"
	}

	// operations
	operpara := make(map[string]interface{})
	operpara["operationtype"] = "0"
	use := make(map[string]string)
	use["userid"] = userid
	v := make(map[int]interface{})
	v[0] = use
	opm := make(map[string]string)
	opm["default_msg"] = "0"
	opm["subject"] = "{TRIGGER.STATUS}"
	opm["message"] = CreateProblemTpl()
	opm["mediatypeid"] = mediaid
	operpara["opmessage_usr"] = v
	operpara["opmessage"] = opm

	// recovery_operations
	recovpara := make(map[string]interface{})
	recovpara["operationtype"] = "0"
	use2 := make(map[string]string)
	use2["userid"] = userid
	v2 := make(map[int]interface{})
	v2[0] = use2
	opm1 := make(map[string]string)
	opm1["default_msg"] = "0"
	opm1["subject"] = "{TRIGGER.STATUS}"
	opm1["message"] = CreateRecoveryTpl()
	opm1["mediatypeid"] = mediaid
	recovpara["opmessage_usr"] = v2
	recovpara["opmessage"] = opm1

	reinter := make(map[int]interface{})
	reinter[0] = operpara
	actpara["operations"] = reinter
	reinter1 := make(map[int]interface{})
	reinter1[0] = recovpara
	actpara["recovery_operations"] = reinter1

	_, err = api.CallWithError("action.create", actpara)
	if err != nil {
		return fmt.Errorf("创建 Action 失败: %w", err)
	}
	logger.Log.Info("Action 创建成功")

	// 生成新的 Token
	webhookToken := strings.ReplaceAll(uuid.New().String(), "-", "")
	logger.Log.Info("生成 MS-Agent WebhookToken:", webhookToken)

	// 更新租户信息
	instance.WebhookToken = webhookToken
	instance.MSAgentInstalled = true

	err = DB.Model(&ZabbixInstance{}).Where("id = ?", instance.ID).Updates(map[string]interface{}{
		"webhook_token":      webhookToken,
		"ms_agent_installed": true,
	}).Error
	if err != nil {
		return fmt.Errorf("更新租户信息失败: %w", err)
	}

	logger.Log.Info("MS-Agent 配置安装完成！")
	return nil
}
