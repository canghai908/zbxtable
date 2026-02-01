package model

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"

	zabbix "github.com/canghai908/zabbix-go"
)

// TestZabbixInstanceConfig 测试 Zabbix 配置能否连通
func TestZabbixInstanceConfig(webURL, user, pass, token string) (string, error) {
	web := strings.TrimRight(strings.TrimSpace(webURL), "/")
	if web == "" {
		return "", errors.New("url is empty")
	}
	apiURL := web + "/api_jsonrpc.php"

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := http.Client{Transport: transport, Timeout: 5 * time.Second}
	resp, err := client.Get(apiURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPreconditionFailed {
		return "", errors.New("Zabbix API 地址不正确，状态码: " + strconv.Itoa(resp.StatusCode))
	}

	api := zabbix.NewAPI(apiURL)
	if strings.TrimSpace(token) != "" {
		api.SetAuth(strings.TrimSpace(token))
	} else {
		if strings.TrimSpace(user) == "" || strings.TrimSpace(pass) == "" {
			return "", errors.New("请输入 token，或提供 user/pass")
		}
		if _, err := api.Login(strings.TrimSpace(user), strings.TrimSpace(pass)); err != nil {
			return "", err
		}
	}

	ver, err := api.Version()
	if err != nil {
		return "", err
	}
	return ver, nil
}

// ListZabbixInstances 列出所有 Zabbix 实例
func ListZabbixInstance() ([]ZabbixInstance, error) {
	var list []ZabbixInstance
	err := DB.Order("id asc").Find(&list).Error
	return list, err
}

// GetZabbixInstanceByZID 根据ZID获取实例
func GetZabbixInstanceByZID(id int) (*ZabbixInstance, error) {
	var instance ZabbixInstance
	err := DB.First(&instance, id).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// GetZabbixInstanceByInstanceID 根据 instance_id 获取实例
func GetZabbixInstanceByInstanceID(instance_id string) (*ZabbixInstance, error) {
	iid := strings.TrimSpace(instance_id)
	if iid == "" {
		return nil, errors.New("instance_id is empty")
	}
	var instance ZabbixInstance
	err := DB.Where("instance_id = ?", iid).First(&instance).Error
	if err != nil {
		return nil, err
	}
	return &instance, nil
}

// CreateZabbixInstance 创建 Zabbix 实例
func CreateZabbixInstance(m *ZabbixInstance) error {
	m.URL = strings.TrimRight(strings.TrimSpace(m.URL), "/")
	m.InstanceID = strings.TrimSpace(m.InstanceID)
	m.Name = strings.TrimSpace(m.Name)
	fmt.Println(m.InstanceID)
	if m.InstanceID == "" {
		return errors.New("instance_id is required")
	}
	if m.Name == "" {
		return errors.New("name is required")
	}
	if m.URL == "" {
		return errors.New("URL is required")
	}

	// 加密密码和Token
	encryptionKey := GetEncryptionKey()
	if m.Pass != "" {
		encryptedPass, err := utils.EncryptString(m.Pass, encryptionKey)
		if err != nil {
			return fmt.Errorf("加密密码失败: %w", err)
		}
		m.Pass = encryptedPass
	}
	if m.Token != "" {
		encryptedToken, err := utils.EncryptString(m.Token, encryptionKey)
		if err != nil {
			return fmt.Errorf("加密Token失败: %w", err)
		}
		m.Token = encryptedToken
	}

	m.UpdatedAt = time.Now()
	return DB.Create(m).Error
}

// UpdateZabbixInstance 更新 Zabbix 实例
func UpdateZabbixInstance(zid int, patch *ZabbixInstance) (*ZabbixInstance, error) {
	var instance ZabbixInstance
	if err := DB.First(&instance, zid).Error; err != nil {
		return nil, err
	}

	// 解密现有的密码和Token用于比较
	encryptionKey := GetEncryptionKey()
	currentPass := instance.Pass
	currentToken := instance.Token
	if currentPass != "" {
		decryptedPass, _ := utils.DecryptString(currentPass, encryptionKey)
		currentPass = decryptedPass
	}
	if currentToken != "" {
		decryptedToken, _ := utils.DecryptString(currentToken, encryptionKey)
		currentToken = decryptedToken
	}

	newWeb := strings.TrimRight(strings.TrimSpace(patch.URL), "/")
	changedConn := false
	if newWeb != "" && newWeb != strings.TrimRight(strings.TrimSpace(instance.URL), "/") {
		changedConn = true
	}
	if strings.TrimSpace(patch.User) != strings.TrimSpace(instance.User) ||
		strings.TrimSpace(patch.Pass) != currentPass ||
		strings.TrimSpace(patch.Token) != currentToken {
		changedConn = true
	}

	updates := map[string]interface{}{}
	if strings.TrimSpace(patch.InstanceID) != "" {
		updates["instance_id"] = strings.TrimSpace(patch.InstanceID)
	}
	if strings.TrimSpace(patch.Name) != "" {
		updates["name"] = strings.TrimSpace(patch.Name)
	}
	if newWeb != "" {
		updates["url"] = newWeb
	}
	updates["user"] = patch.User
	
	// 加密密码和Token
	if patch.Pass != "" {
		encryptedPass, err := utils.EncryptString(patch.Pass, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("加密密码失败: %w", err)
		}
		updates["pass"] = encryptedPass
	} else {
		updates["pass"] = ""
	}
	
	if patch.Token != "" {
		encryptedToken, err := utils.EncryptString(patch.Token, encryptionKey)
		if err != nil {
			return nil, fmt.Errorf("加密Token失败: %w", err)
		}
		updates["token"] = encryptedToken
	} else {
		updates["token"] = ""
	}
	
	updates["enabled"] = patch.Enabled
	updates["notify_method"] = patch.NotifyMethod
	updates["updated_at"] = time.Now()

	if changedConn {
		updates["last_test_ok"] = false
		updates["last_test_message"] = "未验证"
		updates["last_test_at"] = nil
		updates["version"] = ""
	}

	if err := DB.Model(&ZabbixInstance{}).Where("id = ?", zid).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := DB.First(&instance, zid).Error; err != nil {
		return nil, err
	}
	return &instance, nil
}

// DeleteZabbixInstance 删除实例
func DeleteZabbixInstance(zid int) error {
	return DB.Delete(&ZabbixInstance{}, zid).Error
}

// SetZabbixInstanceEnabled 启用/禁用实例
func SetZabbixInstanceEnabled(zid int, enabled bool) (*ZabbixInstance, error) {
	var instance ZabbixInstance
	if err := DB.First(&instance, zid).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&ZabbixInstance{}).Where("id = ?", zid).Update("enabled", enabled).Error; err != nil {
		return nil, err
	}
	instance.Enabled = enabled
	return &instance, nil
}

// TestAndUpdateZabbixInstance 测试并更新实例
func TestAndUpdateZabbixInstance(zid int) (*ZabbixInstance, string, error) {
	var instance ZabbixInstance
	if err := DB.First(&instance, zid).Error; err != nil {
		return nil, "", err
	}
	
	// 解密密码和Token
	encryptionKey := GetEncryptionKey()
	decryptedPass := instance.Pass
	decryptedToken := instance.Token
	if decryptedPass != "" {
		pass, err := utils.DecryptString(decryptedPass, encryptionKey)
		if err != nil {
			logger.Log.Error("解密密码失败:", err)
		} else {
			decryptedPass = pass
		}
	}
	if decryptedToken != "" {
		token, err := utils.DecryptString(decryptedToken, encryptionKey)
		if err != nil {
			logger.Log.Error("解密Token失败:", err)
		} else {
			decryptedToken = token
		}
	}
	
	now := time.Now()
	ver, err := TestZabbixInstanceConfig(instance.URL, instance.User, decryptedPass, decryptedToken)
	if err != nil {
		_ = DB.Model(&ZabbixInstance{}).Where("id = ?", zid).Updates(map[string]interface{}{
			"last_test_ok":      false,
			"last_test_message": err.Error(),
			"last_test_at":      &now,
		}).Error
		instance.LastTestOk = false
		instance.LastTestMessage = err.Error()
		instance.LastTestAt = &now
		return &instance, "", err
	}
	_ = DB.Model(&ZabbixInstance{}).Where("id = ?", zid).Updates(map[string]interface{}{
		"last_test_ok":      true,
		"last_test_message": "连接成功",
		"last_test_at":      &now,
		"version":           ver,
	}).Error
	instance.LastTestOk = true
	instance.LastTestMessage = "连接成功"
	instance.LastTestAt = &now
	instance.Version = ver
	return &instance, ver, nil
}

// UninstallMSAgentFromZabbixInstance 从 Zabbix 实例中卸载 MS-Agent 配置
func UninstallMSAgentFromZabbixInstance(zid int) error {
	instance, err := GetZabbixInstanceByZID(zid)
	if err != nil {
		return fmt.Errorf("获取实例失败: %w", err)
	}

	// 解密密码和Token
	encryptionKey := GetEncryptionKey()
	decryptedPass := instance.Pass
	decryptedToken := instance.Token
	if decryptedPass != "" {
		pass, err := utils.DecryptString(decryptedPass, encryptionKey)
		if err != nil {
			logger.Log.Error("解密密码失败:", err)
		} else {
			decryptedPass = pass
		}
	}
	if decryptedToken != "" {
		token, err := utils.DecryptString(decryptedToken, encryptionKey)
		if err != nil {
			logger.Log.Error("解密Token失败:", err)
		} else {
			decryptedToken = token
		}
	}

	api := zabbix.NewAPI(instance.URL + "/api_jsonrpc.php")
	if decryptedToken != "" {
		api.Auth = decryptedToken
	} else {
		_, err := api.Login(instance.User, decryptedPass)
		if err != nil {
			return fmt.Errorf("登录 Zabbix 失败: %w", err)
		}
	}
	// 删除 Action
	logger.Log.Info("删除 MS-Agent Action...")
	actionParams := map[string]interface{}{
		"output": []string{"actionid"},
		"filter": map[string]string{"name": MSAction},
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

	// 获取并删除用户
	logger.Log.Info("查询 MS-Agent 用户...")
	userParams := map[string]interface{}{
		"output":        []string{"userid"},
		"selectUsrgrps": []string{"usrgrpid"},
		"selectMedias":  []string{"mediatypeid"},
		"filter":        map[string]string{"username": MSUser},
	}
	userRes, err := api.CallWithError("user.get", userParams)
	var usergroupID, mediatypeID string
	if err == nil && userRes.Result != nil {
		resultArray, ok := userRes.Result.([]interface{})
		if ok && len(resultArray) > 0 {
			userMap := resultArray[0].(map[string]interface{})
			userID := userMap["userid"].(string)
			if usrgrps, ok := userMap["usrgrps"].([]interface{}); ok && len(usrgrps) > 0 {
				grpMap := usrgrps[0].(map[string]interface{})
				usergroupID = grpMap["usrgrpid"].(string)
			}
			if medias, ok := userMap["medias"].([]interface{}); ok && len(medias) > 0 {
				mediaMap := medias[0].(map[string]interface{})
				mediatypeID = mediaMap["mediatypeid"].(string)
			}
			_, _ = api.CallWithError("user.delete", []string{userID})
			logger.Log.Info("用户删除成功")
		}
	}

	if usergroupID != "" {
		_, _ = api.CallWithError("usergroup.delete", []string{usergroupID})
		logger.Log.Info("用户组删除成功")
	}
	if mediatypeID != "" {
		_, _ = api.CallWithError("mediatype.delete", []string{mediatypeID})
		logger.Log.Info("Media Type 删除成功")
	}

	_ = DB.Model(&ZabbixInstance{}).Where("id = ?", zid).Updates(map[string]interface{}{
		"ms_agent_installed": false,
		"ms_agent_version":   "",
	}).Error

	logger.Log.Info("MS-Agent 配置卸载完成！")
	return nil
}
