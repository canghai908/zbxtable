package models

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/logger"

	zabbix "github.com/canghai908/zabbix-go"
	"gorm.io/gorm"
)

// TestZabbixTenantConfig 测试 Zabbix 配置能否连通
func TestZabbixTenantConfig(webURL, user, pass, zabbixToken string) (string, error) {
	web := strings.TrimRight(strings.TrimSpace(webURL), "/")
	if web == "" {
		return "", errors.New("web_url is empty")
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
	if strings.TrimSpace(zabbixToken) != "" {
		api.SetAuth(strings.TrimSpace(zabbixToken))
	} else {
		if strings.TrimSpace(user) == "" || strings.TrimSpace(pass) == "" {
			return "", errors.New("请输入 zabbix_token，或提供 user/pass")
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

// ListZabbixTenants 列出所有 Zabbix 租户
func ListZabbixTenants() ([]ZabbixTenant, error) {
	var list []ZabbixTenant
	err := DB.Order("id desc").Find(&list).Error
	return list, err
}

// GetZabbixTenantByID 根据ID获取
func GetZabbixTenantByID(id int64) (*ZabbixTenant, error) {
	var tenant ZabbixTenant
	err := DB.First(&tenant, id).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetZabbixTenantByTenantID 根据 tenant_id 获取
func GetZabbixTenantByTenantID(tenantID string) (*ZabbixTenant, error) {
	tid := strings.TrimSpace(tenantID)
	if tid == "" {
		return nil, errors.New("tenant_id is empty")
	}
	var tenant ZabbixTenant
	err := DB.Where("tenant_id = ?", tid).First(&tenant).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// CreateZabbixTenant 创建 Zabbix 租户
func CreateZabbixTenant(m *ZabbixTenant) error {
	m.WebURL = strings.TrimRight(strings.TrimSpace(m.WebURL), "/")
	m.TenantID = strings.TrimSpace(m.TenantID)
	m.Name = strings.TrimSpace(m.Name)

	if m.TenantID == "" {
		return errors.New("tenant_id is required")
	}
	if m.Name == "" {
		return errors.New("name is required")
	}
	if m.WebURL == "" {
		return errors.New("web_url is required")
	}

	m.UpdatedAt = time.Now()
	return DB.Create(m).Error
}

// UpdateZabbixTenant 更新 Zabbix 租户
func UpdateZabbixTenant(id int64, patch *ZabbixTenant) (*ZabbixTenant, error) {
	var tenant ZabbixTenant
	if err := DB.First(&tenant, id).Error; err != nil {
		return nil, err
	}

	newWeb := strings.TrimRight(strings.TrimSpace(patch.WebURL), "/")
	changedConn := false
	if newWeb != "" && newWeb != strings.TrimRight(strings.TrimSpace(tenant.WebURL), "/") {
		changedConn = true
	}
	if strings.TrimSpace(patch.User) != strings.TrimSpace(tenant.User) ||
		strings.TrimSpace(patch.Pass) != strings.TrimSpace(tenant.Pass) ||
		strings.TrimSpace(patch.ZabbixToken) != strings.TrimSpace(tenant.ZabbixToken) {
		changedConn = true
	}

	updates := map[string]interface{}{}
	if strings.TrimSpace(patch.TenantID) != "" {
		updates["tenant_id"] = strings.TrimSpace(patch.TenantID)
	}
	if strings.TrimSpace(patch.Name) != "" {
		updates["name"] = strings.TrimSpace(patch.Name)
	}
	if newWeb != "" {
		updates["web_url"] = newWeb
	}
	updates["user"] = patch.User
	updates["pass"] = patch.Pass
	updates["zabbix_token"] = patch.ZabbixToken
	updates["enabled"] = patch.Enabled
	updates["notify_method"] = patch.NotifyMethod
	updates["updated_at"] = time.Now()

	if changedConn {
		updates["last_test_ok"] = false
		updates["last_test_message"] = "未验证"
		updates["last_test_at"] = nil
		updates["version"] = ""
	}

	if err := DB.Model(&ZabbixTenant{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	if err := DB.First(&tenant, id).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

// DeleteZabbixTenant 删除租户
func DeleteZabbixTenant(id int64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var tenant ZabbixTenant
		if err := tx.First(&tenant, id).Error; err != nil {
			return err
		}
		if tenant.IsActive {
			if err := tx.Model(&ZabbixTenant{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&ZabbixTenant{}, id).Error
	})
}

// SetZabbixTenantEnabled 启用/禁用
func SetZabbixTenantEnabled(id int64, enabled bool) (*ZabbixTenant, error) {
	var tenant ZabbixTenant
	if err := DB.First(&tenant, id).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&ZabbixTenant{}).Where("id = ?", id).Update("enabled", enabled).Error; err != nil {
		return nil, err
	}
	tenant.Enabled = enabled
	if !enabled && tenant.IsActive {
		_ = DB.Model(&ZabbixTenant{}).Where("id = ?", id).Update("is_active", false).Error
		tenant.IsActive = false
	}
	return &tenant, nil
}

// TestAndUpdateZabbixTenant 测试并更新
func TestAndUpdateZabbixTenant(id int64) (*ZabbixTenant, string, error) {
	var tenant ZabbixTenant
	if err := DB.First(&tenant, id).Error; err != nil {
		return nil, "", err
	}
	now := time.Now()
	ver, err := TestZabbixTenantConfig(tenant.WebURL, tenant.User, tenant.Pass, tenant.ZabbixToken)
	if err != nil {
		_ = DB.Model(&ZabbixTenant{}).Where("id = ?", id).Updates(map[string]interface{}{
			"last_test_ok":      false,
			"last_test_message": err.Error(),
			"last_test_at":      &now,
		}).Error
		tenant.LastTestOk = false
		tenant.LastTestMessage = err.Error()
		tenant.LastTestAt = &now
		return &tenant, "", err
	}
	_ = DB.Model(&ZabbixTenant{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_test_ok":      true,
		"last_test_message": "连接成功",
		"last_test_at":      &now,
		"version":           ver,
	}).Error
	tenant.LastTestOk = true
	tenant.LastTestMessage = "连接成功"
	tenant.LastTestAt = &now
	tenant.Version = ver
	return &tenant, ver, nil
}

// ActivateZabbixTenant 设置为当前激活的 Zabbix
func ActivateZabbixTenant(id int64) (*ZabbixTenant, error) {
	var tenant ZabbixTenant
	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&tenant, id).Error; err != nil {
			return err
		}
		if !tenant.Enabled {
			return errors.New("该 Zabbix 已禁用，无法设为当前")
		}
		if err := tx.Model(&ZabbixTenant{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&ZabbixTenant{}).Where("id = ?", id).Update("is_active", true).Error; err != nil {
			return err
		}
		tenant.IsActive = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	if err := ApplyActiveZabbixTenant(&tenant); err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetActiveZabbixTenant 获取当前激活的租户
func GetActiveZabbixTenant() (*ZabbixTenant, error) {
	var tenant ZabbixTenant
	err := DB.Where("is_active = ?", true).Order("id desc").First(&tenant).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// ApplyActiveZabbixTenant 应用到全局 API
func ApplyActiveZabbixTenant(tenant *ZabbixTenant) error {
	if tenant == nil {
		return nil
	}
	if !tenant.Enabled {
		return errors.New("该 Zabbix 已禁用")
	}
	web := strings.TrimRight(strings.TrimSpace(tenant.WebURL), "/")
	if web == "" {
		return nil
	}
	apiURL := web + "/api_jsonrpc.php"
	api := zabbix.NewAPI(apiURL)
	if strings.TrimSpace(tenant.ZabbixToken) != "" {
		api.SetAuth(strings.TrimSpace(tenant.ZabbixToken))
	} else if strings.TrimSpace(tenant.User) != "" || strings.TrimSpace(tenant.Pass) != "" {
		_, err := api.Login(strings.TrimSpace(tenant.User), strings.TrimSpace(tenant.Pass))
		if err != nil {
			return err
		}
	}
	API = api

	if ver, err := API.Version(); err == nil {
		ZBX_VER = ver
		now := time.Now()
		_ = DB.Model(&ZabbixTenant{}).Where("id = ?", tenant.ID).Updates(map[string]interface{}{
			"version":           ver,
			"last_test_ok":      true,
			"last_test_message": "连接成功",
			"last_test_at":      &now,
		}).Error
		verArr := strings.Split(ZBX_VER, ".")
		if len(verArr) >= 2 {
			ZbxMasterVer, _ := strconv.ParseInt(verArr[0], 10, 64)
			ZbxMiddleVer, _ := strconv.ParseInt(verArr[1], 10, 64)
			if ZbxMasterVer >= 6 || (ZbxMasterVer == 5 && ZbxMiddleVer == 4) {
				ZBX_V = true
			} else {
				ZBX_V = false
			}
		}
	}
	return nil
}

// TryInitZabbixTenantFromDB 启动时自动初始化
func TryInitZabbixTenantFromDB() {
	tenant, err := GetActiveZabbixTenant()
	if err != nil || tenant == nil {
		return
	}
	_ = ApplyActiveZabbixTenant(tenant)
}

// UninstallMSAgentFromZabbixTenant 从 Zabbix 中卸载 MS-Agent 配置
func UninstallMSAgentFromZabbixTenant(id int64) error {
	tenant, err := GetZabbixTenantByID(id)
	if err != nil {
		return fmt.Errorf("获取租户失败: %w", err)
	}

	api := zabbix.NewAPI(tenant.WebURL + "/api_jsonrpc.php")
	if tenant.Token != "" {
		api.Auth = tenant.Token
	} else {
		_, err := api.Login(tenant.User, tenant.Pass)
		if err != nil {
			return fmt.Errorf("登录 Zabbix 失败: %w", err)
		}
	}

	const (
		MSUser   = "ms-agent"
		MSAction = "MS-Agent"
	)

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

	_ = DB.Model(&ZabbixTenant{}).Where("id = ?", id).Updates(map[string]interface{}{
		"ms_agent_installed": false,
		"ms_agent_version":   "",
	}).Error

	logger.Log.Info("MS-Agent 配置卸载完成！")
	return nil
}

func TryInitZabbixFromDB() {
	inst, err := GetActiveZabbixTenant()
	if err != nil || inst == nil {
		return
	}
	_ = ApplyActiveZabbixTenant(inst)
}
