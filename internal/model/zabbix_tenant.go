package models

import (
	"errors"
	"fmt"
	"strings"
	"zbxtable/pkg/utils"

	zabbix "github.com/canghai908/zabbix-go"
)

func ListZabbixTenantBindings() ([]ZabbixTenantBinding, error) {
	var list []ZabbixTenantBinding
	err := DB.Order("id desc").Find(&list).Error
	return list, err
}

func GetZabbixTenantBindingByTenant(tenantID string) (*ZabbixTenantBinding, error) {
	tid := strings.TrimSpace(tenantID)
	if tid == "" {
		return nil, errors.New("tenant_id is empty")
	}
	var v ZabbixTenantBinding
	err := DB.Where("tenant_id = ?", tid).First(&v).Error
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func UpsertZabbixTenantBinding(m *ZabbixTenantBinding) error {
	if m == nil {
		return errors.New("binding is nil")
	}
	m.TenantID = strings.TrimSpace(m.TenantID)
	m.Token = strings.TrimSpace(m.Token)
	if m.TenantID == "" {
		return errors.New("tenant_id is empty")
	}
	if m.ZabbixInstanceID == 0 {
		return errors.New("zabbix_instance_id is empty")
	}
	// upsert by tenant_id (unique)
	var exist ZabbixTenantBinding
	err := DB.Where("tenant_id = ?", m.TenantID).First(&exist).Error
	if err == nil {
		// 更新时保留安装状态字段
		updateData := map[string]interface{}{
			"zabbix_instance_id": m.ZabbixInstanceID,
			"token":              m.Token,
			"notify_method":      m.NotifyMethod,
			"enabled":            m.Enabled,
		}
		// 如果有安装状态更新，也一并更新（包括设置为 false 的情况）
		if m.MSAgentInstalled || exist.MSAgentInstalled {
			updateData["ms_agent_installed"] = m.MSAgentInstalled
		}
		if m.MSAgentVersion != "" || exist.MSAgentVersion != "" {
			updateData["ms_agent_version"] = m.MSAgentVersion
		}
		if m.WebhookInstalled || exist.WebhookInstalled {
			updateData["webhook_installed"] = m.WebhookInstalled
		}
		if m.WebhookURL != "" || exist.WebhookURL != "" {
			updateData["webhook_url"] = m.WebhookURL
		}
		return DB.Model(&ZabbixTenantBinding{}).
			Where("id = ?", exist.ID).
			Updates(updateData).Error
	}
	return DB.Create(m).Error
}

func DeleteZabbixTenantBinding(id int) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return DB.Delete(&ZabbixTenantBinding{}, id).Error
}

// UninstallMSAgentFromZabbix 从 Zabbix 中卸载 MS-Agent 配置
func UninstallMSAgentFromZabbix(instanceID int, tenantID string) error {
	// 获取 Zabbix 实例信息
	instance, err := GetZabbixInstanceByID(int64(instanceID))
	if err != nil {
		return fmt.Errorf("获取 Zabbix 实例失败: %w", err)
	}

	// 初始化 Zabbix API
	api := zabbix.NewAPI(instance.WebURL + "/api_jsonrpc.php")
	if instance.Token != "" {
		api.Auth = instance.Token
	} else {
		_, err := api.Login(instance.User, instance.Pass)
		if err != nil {
			return fmt.Errorf("登录 Zabbix 失败: %w", err)
		}
	}

	const (
		MSUser   = "ms-agent"
		MSGroup  = "MS-Agent Group"
		MSMedia  = "MS-Agent Media"
		MSAction = "MS-Agent Action"
	)

	// 1. 删除 Action
	utils.Log.Info("删除 MS-Agent Action...")
	actionParams := map[string]interface{}{
		"output": []string{"actionid"},
		"filter": map[string]string{
			"name": MSAction,
		},
	}

	actionRes, err := api.CallWithError("action.get", actionParams)
	if err == nil && actionRes.Result != nil {
		resultArray, ok := actionRes.Result.([]interface{})
		if ok && len(resultArray) > 0 {
			actionMap := resultArray[0].(map[string]interface{})
			actionID := actionMap["actionid"].(string)

			_, err = api.CallWithError("action.delete", []string{actionID})
			if err != nil {
				utils.Log.Warn("删除 Action 失败:", err)
			} else {
				utils.Log.Info("Action 删除成功")
			}
		}
	}

	// 2. 获取用户信息（包含 usergroup 和 mediatype）
	utils.Log.Info("查询 MS-Agent 用户...")
	userParams := map[string]interface{}{
		"output":        []string{"userid"},
		"selectUsrgrps": []string{"usrgrpid"},
		"selectMedias":  []string{"mediatypeid"},
		"filter": map[string]string{
			"username": MSUser,
		},
	}

	userRes, err := api.CallWithError("user.get", userParams)
	var usergroupID, mediatypeID string

	if err == nil && userRes.Result != nil {
		resultArray, ok := userRes.Result.([]interface{})
		if ok && len(resultArray) > 0 {
			userMap := resultArray[0].(map[string]interface{})
			userID := userMap["userid"].(string)

			// 获取 usergroup ID
			if usrgrps, ok := userMap["usrgrps"].([]interface{}); ok && len(usrgrps) > 0 {
				grpMap := usrgrps[0].(map[string]interface{})
				usergroupID = grpMap["usrgrpid"].(string)
			}

			// 获取 mediatype ID
			if medias, ok := userMap["medias"].([]interface{}); ok && len(medias) > 0 {
				mediaMap := medias[0].(map[string]interface{})
				mediatypeID = mediaMap["mediatypeid"].(string)
			}

			// 删除用户
			utils.Log.Info("删除 MS-Agent 用户...")
			_, err = api.CallWithError("user.delete", []string{userID})
			if err != nil {
				utils.Log.Warn("删除用户失败:", err)
			} else {
				utils.Log.Info("用户删除成功")
			}
		}
	}

	// 3. 删除 User Group
	if usergroupID != "" {
		utils.Log.Info("删除 MS-Agent 用户组...")
		_, err = api.CallWithError("usergroup.delete", []string{usergroupID})
		if err != nil {
			utils.Log.Warn("删除用户组失败:", err)
		} else {
			utils.Log.Info("用户组删除成功")
		}
	}

	// 4. 删除 Media Type
	if mediatypeID != "" {
		utils.Log.Info("删除 MS-Agent Media Type...")
		_, err = api.CallWithError("mediatype.delete", []string{mediatypeID})
		if err != nil {
			utils.Log.Warn("删除 Media Type 失败:", err)
		} else {
			utils.Log.Info("Media Type 删除成功")
		}
	}

	// 5. 更新租户绑定信息
	binding, err := GetZabbixTenantBindingByTenant(tenantID)
	if err != nil {
		return fmt.Errorf("获取租户绑定失败: %w", err)
	}

	binding.MSAgentInstalled = false
	binding.MSAgentVersion = ""
	err = UpsertZabbixTenantBinding(binding)
	if err != nil {
		return fmt.Errorf("更新租户绑定失败: %w", err)
	}

	utils.Log.Info("MS-Agent 配置卸载完成！")
	return nil
}

// UninstallWebhookFromZabbix 从 Zabbix 中卸载 Webhook 配置
func UninstallWebhookFromZabbix(instanceID int, tenantID string) error {
	// 获取 Zabbix 实例信息
	instance, err := GetZabbixInstanceByID(int64(instanceID))
	if err != nil {
		return fmt.Errorf("获取 Zabbix 实例失败: %w", err)
	}

	// 初始化 Zabbix API
	api := zabbix.NewAPI(instance.WebURL + "/api_jsonrpc.php")
	if instance.Token != "" {
		api.Auth = instance.Token
	} else {
		_, err := api.Login(instance.User, instance.Pass)
		if err != nil {
			return fmt.Errorf("登录 Zabbix 失败: %w", err)
		}
	}

	const (
		WebhookName   = "ZbxTable Webhook"
		WebhookAction = "ZbxTable Webhook Action"
		WebhookUser   = "zbxtable-webhook"
		WebhookGroup  = "ZbxTable Webhook Group"
	)

	// 1. 删除 Action
	utils.Log.Info("删除 Webhook Action...")
	actionParams := map[string]interface{}{
		"output": []string{"actionid"},
		"filter": map[string]string{
			"name": WebhookAction,
		},
	}

	actionRes, err := api.CallWithError("action.get", actionParams)
	if err == nil && actionRes.Result != nil {
		resultArray, ok := actionRes.Result.([]interface{})
		if ok && len(resultArray) > 0 {
			actionMap := resultArray[0].(map[string]interface{})
			actionID := actionMap["actionid"].(string)

			_, err = api.CallWithError("action.delete", []string{actionID})
			if err != nil {
				utils.Log.Warn("删除 Action 失败:", err)
			} else {
				utils.Log.Info("Action 删除成功")
			}
		}
	}

	// 2. 获取用户信息（包含 usergroup 和 mediatype）
	utils.Log.Info("查询 Webhook 用户...")
	userParams := map[string]interface{}{
		"output":        []string{"userid"},
		"selectUsrgrps": []string{"usrgrpid"},
		"selectMedias":  []string{"mediatypeid"},
		"filter": map[string]string{
			"username": WebhookUser,
		},
	}

	userRes, err := api.CallWithError("user.get", userParams)
	var usergroupID, mediatypeID string

	if err == nil && userRes.Result != nil {
		resultArray, ok := userRes.Result.([]interface{})
		if ok && len(resultArray) > 0 {
			userMap := resultArray[0].(map[string]interface{})
			userID := userMap["userid"].(string)

			// 获取 usergroup ID
			if usrgrps, ok := userMap["usrgrps"].([]interface{}); ok && len(usrgrps) > 0 {
				grpMap := usrgrps[0].(map[string]interface{})
				usergroupID = grpMap["usrgrpid"].(string)
			}

			// 获取 mediatype ID
			if medias, ok := userMap["medias"].([]interface{}); ok && len(medias) > 0 {
				mediaMap := medias[0].(map[string]interface{})
				mediatypeID = mediaMap["mediatypeid"].(string)
			}

			// 删除用户
			utils.Log.Info("删除 Webhook 用户...")
			_, err = api.CallWithError("user.delete", []string{userID})
			if err != nil {
				utils.Log.Warn("删除用户失败:", err)
			} else {
				utils.Log.Info("用户删除成功")
			}
		}
	}

	// 3. 删除 User Group
	if usergroupID != "" {
		utils.Log.Info("删除 Webhook 用户组...")
		_, err = api.CallWithError("usergroup.delete", []string{usergroupID})
		if err != nil {
			utils.Log.Warn("删除用户组失败:", err)
		} else {
			utils.Log.Info("用户组删除成功")
		}
	}

	// 4. 删除 Media Type
	if mediatypeID != "" {
		utils.Log.Info("删除 Webhook Media Type...")
		_, err = api.CallWithError("mediatype.delete", []string{mediatypeID})
		if err != nil {
			utils.Log.Warn("删除 Media Type 失败:", err)
		} else {
			utils.Log.Info("Media Type 删除成功")
		}
	}

	// 5. 更新租户绑定信息
	binding, err := GetZabbixTenantBindingByTenant(tenantID)
	if err != nil {
		return fmt.Errorf("获取租户绑定失败: %w", err)
	}

	binding.WebhookInstalled = false
	binding.WebhookURL = ""
	err = UpsertZabbixTenantBinding(binding)
	if err != nil {
		return fmt.Errorf("更新租户绑定失败: %w", err)
	}

	utils.Log.Info("Webhook 配置卸载完成！")
	return nil
}
