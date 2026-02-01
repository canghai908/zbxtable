package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"zbxtable/pkg/logger"

	zabbix "github.com/canghai908/zabbix-go"
	"github.com/google/uuid"
)

// InstallWebhookToZabbix 在 Zabbix 中安装 Webhook 配置
func InstallWebhookToZabbix(zid int, zbxtableURL string) error {
	// 常量定义

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

	// Webhook 需要 Zabbix 4.4+
	if zbxMasterVer < 4 {
		return fmt.Errorf("webhook 需要 Zabbix 4.4 或更高版本，当前版本: %s", version)
	}

	// 生成新的 WebhookToken
	webhookToken := strings.ReplaceAll(uuid.New().String(), "-", "")
	logger.Log.Info("生成 Webhook WebhookToken:", webhookToken)

	// 构建 Webhook URL - 从数据库配置读取
	configuredURL := GetConfigValueByKey("webhook_url", "")
	if configuredURL != "" {
		zbxtableURL = configuredURL
		logger.Log.Info("使用配置的 Webhook URL:", zbxtableURL)
	}
	webhookURL := fmt.Sprintf("%s/v1/receive", strings.TrimRight(zbxtableURL, "/"))
	logger.Log.Info("Webhook URL:", webhookURL)

	// Webhook 脚本 - 使用 Parameters 传递配置，消息体使用 {ALERT.MESSAGE}
	webhookScript := `try {
    var params = JSON.parse(value);
    var req = new HttpRequest();
    req.addHeader('Content-Type: application/json');
    req.addHeader('X-Instance: ' + params.instance_id);
    req.addHeader('X-Token: ' + params.webhook_token);
    
    // 直接使用 {ALERT.MESSAGE} 作为消息体
    var response = req.post(params.webhook_url, params.message);
    
    if (req.getStatus() !== 200) {
        throw 'Response code: ' + req.getStatus();
    }
    
    return 'OK';
} catch (error) {
    Zabbix.log(4, 'ZbxTable webhook error: ' + error);
    throw error;
}`

	// 准备 Media Type 参数
	mediaParams := make(map[string]interface{})
	mediaParams["name"] = WebhookName
	mediaParams["type"] = "4" // Webhook type
	mediaParams["status"] = "0"
	mediaParams["script"] = webhookScript

	// Webhook 参数 - 只保留 webhook_url、tenant_id、webhook_token 和 message
	parameters := []map[string]interface{}{
		{"name": "webhook_url", "value": webhookURL},
		{"name": "instance_id", "value": instance.InstanceID},
		{"name": "webhook_token", "value": webhookToken},
		{"name": "message", "value": "{ALERT.MESSAGE}"},
	}
	mediaParams["parameters"] = parameters

	// 1. 创建或更新 Webhook Media Type
	logger.Log.Info("检查 Webhook Media Type...")

	// 先查询是否已存在
	getParams := map[string]interface{}{
		"output": "extend",
		"filter": map[string]string{
			"name": WebhookName,
		},
	}

	existingMedia, err := api.CallWithError("mediatype.get", getParams)
	var mediaid string

	if err == nil && existingMedia.Result != nil {
		resultArray, ok := existingMedia.Result.([]interface{})
		if ok && len(resultArray) > 0 {
			// Media Type 已存在，更新它
			existingMediaMap := resultArray[0].(map[string]interface{})
			mediaid = existingMediaMap["mediatypeid"].(string)
			logger.Log.Info("Webhook Media Type 已存在, ID:", mediaid, "，正在更新...")

			mediaParams["mediatypeid"] = mediaid
			_, err = api.CallWithError("mediatype.update", mediaParams)
			if err != nil {
				return fmt.Errorf("更新 Webhook Media Type 失败: %w", err)
			}
			logger.Log.Info("Webhook Media Type 更新成功")
		} else {
			// 不存在，创建新的
			logger.Log.Info("创建 Webhook Media Type...")
			ma, err := api.CallWithError("mediatype.create", mediaParams)
			if err != nil {
				return fmt.Errorf("创建 Webhook Media Type 失败: %w", err)
			}
			result := ma.Result.(map[string]interface{})
			mediatypeids := result["mediatypeids"].([]interface{})
			mediaid = mediatypeids[0].(string)
			logger.Log.Info("Webhook Media Type 创建成功, ID:", mediaid)
		}
	} else {
		// 查询失败，尝试创建
		logger.Log.Info("创建 Webhook Media Type...")
		ma, err := api.CallWithError("mediatype.create", mediaParams)
		if err != nil {
			return fmt.Errorf("创建 Webhook Media Type 失败: %w", err)
		}
		result := ma.Result.(map[string]interface{})
		mediatypeids := result["mediatypeids"].([]interface{})
		mediaid = mediatypeids[0].(string)
		logger.Log.Info("Webhook Media Type 创建成功, ID:", mediaid)
	}

	// 2. 创建用户组（参考 ms-agent）
	logger.Log.Info("创建 Webhook 用户组...")

	groupParams := make(map[string]interface{})
	groupParams["name"] = WebhookGroup
	group, err := api.CallWithError("usergroup.create", groupParams)
	if err != nil {
		return fmt.Errorf("创建用户组失败: %w", err)
	}

	resgroup := group.Result.(map[string]interface{})
	usrgrpids := resgroup["usrgrpids"].([]interface{})
	groupid := usrgrpids[0].(string)
	logger.Log.Info("用户组创建成功, ID:", groupid)

	// 3. 创建用户（参考 ms-agent）
	logger.Log.Info("创建 Webhook 用户...")
	userpara := make(map[string]interface{})
	usrgrps := make(map[string]string)
	usermepara := make(map[string]string)
	usrgrps["usrgrpid"] = groupid
	a := make(map[int]interface{})
	a[0] = usrgrps
	usermepara["mediatypeid"] = mediaid
	usermepara["sendto"] = "webhook"
	usermepara["active"] = "0"
	usermepara["severity"] = "63"
	usermepara["period"] = "1-7,00:00-24:00"
	b := make(map[int]interface{})
	b[0] = usermepara

	// 检查 Zabbix 版本
	//5.4版本，以后user 增加roleid，之前为type表示 有user参数有区别
	isNewVersion := zbxMasterVer >= 6 || (zbxMasterVer == 5 && zbxMiddleVer == 4)

	if isNewVersion {
		userpara["username"] = WebhookUser
	} else {
		userpara["alias"] = WebhookUser
		userpara["name"] = WebhookUser
	}

	// 生成强密码
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
	logger.Log.Infof("用户创建成功, 用户名: %s, 密码: %s, ID: %s", WebhookUser, tPassword, userid)

	// 4. 创建 Action
	logger.Log.Info("创建 Webhook Action...")
	actpara := make(map[string]interface{})
	actpara["name"] = WebhookAction
	actpara["eventsource"] = "0"
	actpara["status"] = "0"
	actpara["esc_period"] = "60"

	// 生成故障消息模板（参考 ms-agent 格式）
	problemMessage := `{
    ¦host_id¦: ¦{HOST.ID}¦,
    ¦host_host¦: ¦{HOST.HOST}¦,
    ¦hostname¦: ¦{HOST.NAME}¦,
    ¦host_ip¦: ¦{HOST.IP}¦,
    ¦host_group¦: ¦{TRIGGER.HOSTGROUP.NAME}¦,
    ¦event_time¦: ¦{EVENT.DATE} {EVENT.TIME}¦,
    ¦severity¦: ¦{TRIGGER.NSEVERITY}¦,
    ¦trigger_id¦: ¦{TRIGGER.ID}¦,
    ¦trigger_name¦: ¦{TRIGGER.NAME}¦,
    ¦trigger_key¦: ¦{TRIGGER.KEY}¦,
    ¦trigger_value¦: ¦{TRIGGER.VALUE}¦,
    ¦item_id¦: ¦{ITEM.ID}¦,
    ¦item_name¦: ¦{ITEM.NAME}¦,
    ¦item_value¦: ¦{ITEM.VALUE}¦,
    ¦event_id¦: ¦{EVENT.ID}¦,
    ¦event_duration¦: ¦{EVENT.DURATION}¦
}`

	// 生成恢复消息模板（参考 ms-agent 格式）
	recoveryMessage := `{
    ¦host_id¦: ¦{HOST.ID}¦,
    ¦host_host¦: ¦{HOST.HOST}¦,
    ¦hostname¦: ¦{HOST.NAME}¦,
    ¦host_ip¦: ¦{HOST.IP}¦,
    ¦host_group¦: ¦{TRIGGER.HOSTGROUP.NAME}¦,
    ¦event_time¦: ¦{EVENT.RECOVERY.DATE} {EVENT.RECOVERY.TIME}¦,
    ¦severity¦: ¦{TRIGGER.NSEVERITY}¦,
    ¦trigger_id¦: ¦{TRIGGER.ID}¦,
    ¦trigger_name¦: ¦{TRIGGER.NAME}¦,
    ¦trigger_key¦: ¦{TRIGGER.KEY}¦,
    ¦trigger_value¦: ¦{TRIGGER.VALUE}¦,
    ¦item_id¦: ¦{ITEM.ID}¦,
    ¦item_name¦: ¦{ITEM.NAME}¦,
    ¦item_value¦: ¦{ITEM.VALUE}¦,
    ¦event_id¦: ¦{EVENT.ID}¦,
    ¦event_duration¦: ¦{EVENT.DURATION}¦
}`

	// 旧版本需要设置默认消息
	if !isNewVersion {
		actpara["def_shortdata"] = "{TRIGGER.STATUS}"
		actpara["def_longdata"] = problemMessage
		actpara["r_shortdata"] = "{TRIGGER.STATUS}"
		actpara["r_longdata"] = recoveryMessage
		actpara["recovery_msg"] = "1"
	}

	// operations (告警操作) - 发送给创建的用户
	operpara := make(map[string]interface{})
	operpara["operationtype"] = "0"

	use := make(map[string]string)
	use["userid"] = userid
	v := make(map[int]interface{})
	v[0] = use

	opm := make(map[string]string)
	opm["default_msg"] = "0"
	opm["subject"] = "Alert"
	opm["message"] = problemMessage
	opm["mediatypeid"] = mediaid

	operpara["opmessage_usr"] = v
	operpara["opmessage"] = opm

	// recovery_operations (恢复操作) - 发送给创建的用户
	recovpara := make(map[string]interface{})
	recovpara["operationtype"] = "0"

	use2 := make(map[string]string)
	use2["userid"] = userid
	v2 := make(map[int]interface{})
	v2[0] = use2

	opm1 := make(map[string]string)
	opm1["default_msg"] = "0"
	opm1["subject"] = "Resolved"
	opm1["message"] = recoveryMessage
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
		return fmt.Errorf("创建 Webhook Action 失败: %w", err)
	}
	logger.Log.Info("Webhook Action 创建成功")

	// 更新租户信息
	instance.WebhookToken = webhookToken
	instance.NotifyMethod = "webhook"
	instance.WebhookInstalled = true
	instance.WebhookURL = webhookURL

	err = DB.Model(&ZabbixInstance{}).Where("id = ?", instance.ID).Updates(map[string]interface{}{
		"webhook_token":     webhookToken,
		"notify_method":     "webhook",
		"webhook_installed": true,
		"webhook_url":       webhookURL,
	}).Error
	if err != nil {
		return fmt.Errorf("更新租户信息失败: %w", err)
	}

	logger.Log.Info("Webhook 配置安装完成！")
	return nil
}

// GetWebhookInfo 获取 Webhook 配置信息
func GetWebhookInfo(zid int, zbxtableURL string) (map[string]string, error) {
	instance, err := GetZabbixInstanceByZID(zid)
	if err != nil {
		return nil, fmt.Errorf("获取租户失败: %w", err)
	}

	if !instance.WebhookInstalled {
		return nil, errors.New("webhook 未安装")
	}

	webhookURL := fmt.Sprintf("%s/v1/receive", strings.TrimRight(zbxtableURL, "/"))

	info := map[string]string{
		"webhook_url":   webhookURL,
		"instance_id":   instance.InstanceID,
		"webhook_token": instance.WebhookToken,
		"method":        "POST",
		"content_type":  "application/json",
		"headers":       fmt.Sprintf("X-Instance: %s\nX-Token: %s", instance.InstanceID, instance.WebhookToken),
	}

	return info, nil
}

// UninstallWebhookFromZabbixInstance 从 Zabbix 中卸载 Webhook 配置
func UninstallWebhookFromZabbixInstance(id int) error {
	tenant, err := GetZabbixInstanceByZID(int(id))
	if err != nil {
		return fmt.Errorf("获取租户失败: %w", err)
	}

	// 解密密码和Token
	decryptedPass, decryptedToken := DecryptInstanceCredentials(tenant)

	api := zabbix.NewAPI(tenant.URL + "/api_jsonrpc.php")
	if decryptedToken != "" {
		api.Auth = decryptedToken
	} else {
		_, err := api.Login(tenant.User, decryptedPass)
		if err != nil {
			return fmt.Errorf("登录 Zabbix 失败: %w", err)
		}
	}
	// 删除 Action
	logger.Log.Info("删除 Webhook Action...")
	actionParams := map[string]interface{}{
		"output": []string{"actionid"},
		"filter": map[string]string{"name": WebhookAction},
	}
	actionRes, err := api.CallWithError("action.get", actionParams)
	if err == nil && actionRes.Result != nil {
		resultArray, ok := actionRes.Result.([]interface{})
		if ok && len(resultArray) > 0 {
			actionMap := resultArray[0].(map[string]interface{})
			actionID := actionMap["actionid"].(string)
			_, _ = api.CallWithError("action.delete", []string{actionID})
			logger.Log.Info("Action 删除成功")
		}
	}

	// 获取 Zabbix 版本以确定使用 username 还是 alias
	version, err := api.Version()
	if err != nil {
		logger.Log.Warn("获取 Zabbix 版本失败:", err)
	}

	var zbxMasterVer, zbxMiddleVer int64
	if version != "" {
		verArr := strings.Split(version, ".")
		zbxMasterVer, _ = strconv.ParseInt(verArr[0], 10, 64)
		if len(verArr) > 1 {
			zbxMiddleVer, _ = strconv.ParseInt(verArr[1], 10, 64)
		}
	}
	isNewVersion := zbxMasterVer >= 6 || (zbxMasterVer == 5 && zbxMiddleVer == 4)

	// 获取并删除用户
	logger.Log.Info("查询 Webhook 用户...")
	userParams := map[string]interface{}{
		"output":        []string{"userid"},
		"selectUsrgrps": []string{"usrgrpid"},
		"selectMedias":  []string{"mediatypeid"},
	}

	// 根据版本使用不同的过滤字段
	if isNewVersion {
		userParams["filter"] = map[string]string{"username": WebhookUser}
	} else {
		userParams["filter"] = map[string]string{"alias": WebhookUser}
	}

	userRes, err := api.CallWithError("user.get", userParams)
	var usergroupID, mediatypeID string
	if err == nil && userRes.Result != nil {
		resultArray, ok := userRes.Result.([]interface{})
		if ok && len(resultArray) > 0 {
			userMap := resultArray[0].(map[string]interface{})
			userID := userMap["userid"].(string)

			// 获取用户组ID
			if usrgrps, ok := userMap["usrgrps"].([]interface{}); ok && len(usrgrps) > 0 {
				grpMap := usrgrps[0].(map[string]interface{})
				usergroupID = grpMap["usrgrpid"].(string)
			}

			// 获取 Media Type ID
			if medias, ok := userMap["medias"].([]interface{}); ok && len(medias) > 0 {
				mediaMap := medias[0].(map[string]interface{})
				mediatypeID = mediaMap["mediatypeid"].(string)
			}

			// 删除用户
			_, err = api.CallWithError("user.delete", []string{userID})
			if err != nil {
				logger.Log.Warn("删除用户失败:", err)
			} else {
				logger.Log.Info("用户删除成功, ID:", userID)
			}
		} else {
			logger.Log.Warn("未找到 Webhook 用户")
		}
	} else {
		logger.Log.Warn("查询用户失败:", err)
	}

	// 删除用户组
	if usergroupID != "" {
		_, err = api.CallWithError("usergroup.delete", []string{usergroupID})
		if err != nil {
			logger.Log.Warn("删除用户组失败:", err)
		} else {
			logger.Log.Info("用户组删除成功, ID:", usergroupID)
		}
	} else {
		logger.Log.Warn("未找到用户组ID，跳过删除")
	}

	// 删除 Media Type
	if mediatypeID != "" {
		_, err = api.CallWithError("mediatype.delete", []string{mediatypeID})
		if err != nil {
			logger.Log.Warn("删除 Media Type 失败:", err)
		} else {
			logger.Log.Info("Media Type 删除成功, ID:", mediatypeID)
		}
	} else {
		logger.Log.Warn("未找到 Media Type ID，跳过删除")
	}

	_ = DB.Model(&ZabbixInstance{}).Where("id = ?", id).Updates(map[string]interface{}{
		"webhook_installed": false,
		"webhook_url":       "",
	}).Error

	logger.Log.Info("Webhook 配置卸载完成！")
	return nil
}
