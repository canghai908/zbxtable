package models

import (
	"crypto/tls"
	"errors"
	"net/http"
	"strings"
	"time"

	"strconv"

	zabbix "github.com/canghai908/zabbix-go"
	"gorm.io/gorm"
)

// TestZabbixInstanceConfig 测试某个配置能否连通，并返回版本号
// 规则：
// - 先用 GET 检查 /api_jsonrpc.php 是否返回 412（与项目其它位置保持一致）
// - token 优先；否则 user/pass 登录
func TestZabbixInstanceConfig(webURL, user, pass, token string) (string, error) {
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
func ListZabbixInstances() ([]ZabbixInstance, error) {
	var list []ZabbixInstance
	err := DB.Order("id desc").Find(&list).Error
	return list, err
}

// GetZabbixInstanceByID 根据ID获取单个实例（包含敏感信息）
func GetZabbixInstanceByID(id int64) (*ZabbixInstance, error) {
	var inst ZabbixInstance
	err := DB.First(&inst, id).Error
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

// CreateZabbixInstance 创建 Zabbix 实例
func CreateZabbixInstance(m *ZabbixInstance) error {
	m.WebURL = strings.TrimRight(strings.TrimSpace(m.WebURL), "/")
	if !m.Enabled {
		// keep as is
	} else {
		// 默认启用
		m.Enabled = true
	}
	m.UpdatedAt = time.Now()
	return DB.Create(m).Error
}

// SetZabbixInstanceEnabled 启用/禁用
func SetZabbixInstanceEnabled(id int64, enabled bool) (*ZabbixInstance, error) {
	var inst ZabbixInstance
	if err := DB.First(&inst, id).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&ZabbixInstance{}).Where("id = ?", id).Update("enabled", enabled).Error; err != nil {
		return nil, err
	}
	inst.Enabled = enabled
	// 禁用时同时取消 active，避免“当前实例不可用”
	if !enabled && inst.IsActive {
		_ = DB.Model(&ZabbixInstance{}).Where("id = ?", id).Update("is_active", false).Error
		inst.IsActive = false
	}
	return &inst, nil
}

// UpdateZabbixInstance 更新实例（不强制测试；调用方可通过 test 接口更新版本/连接状态）
// 如果连接参数有变化，将 last_test_ok 置为 false，version 清空，避免显示过期的连接状态。
func UpdateZabbixInstance(id int64, patch *ZabbixInstance) (*ZabbixInstance, error) {
	var inst ZabbixInstance
	if err := DB.First(&inst, id).Error; err != nil {
		return nil, err
	}

	newWeb := strings.TrimRight(strings.TrimSpace(patch.WebURL), "/")
	changedConn := false
	if newWeb != "" && newWeb != strings.TrimRight(strings.TrimSpace(inst.WebURL), "/") {
		changedConn = true
	}
	if strings.TrimSpace(patch.User) != strings.TrimSpace(inst.User) ||
		strings.TrimSpace(patch.Pass) != strings.TrimSpace(inst.Pass) ||
		strings.TrimSpace(patch.Token) != strings.TrimSpace(inst.Token) {
		changedConn = true
	}

	updates := map[string]interface{}{}
	if strings.TrimSpace(patch.Name) != "" {
		updates["name"] = strings.TrimSpace(patch.Name)
	}
	if newWeb != "" {
		updates["web_url"] = newWeb
	}
	// 允许置空凭证
	updates["user"] = patch.User
	updates["pass"] = patch.Pass
	updates["token"] = patch.Token
	updates["enabled"] = patch.Enabled
	updates["updated_at"] = time.Now()

	if changedConn {
		updates["last_test_ok"] = false
		updates["last_test_message"] = "未验证"
		updates["last_test_at"] = nil
		updates["version"] = ""
	}

	if err := DB.Model(&ZabbixInstance{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新取一遍最新数据
	if err := DB.First(&inst, id).Error; err != nil {
		return nil, err
	}
	return &inst, nil
}

// DeleteZabbixInstance 删除实例
// 若删除的是当前 active，会自动取消 active（不自动切换到别的实例，避免误切换）。
func DeleteZabbixInstance(id int64) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var inst ZabbixInstance
		if err := tx.First(&inst, id).Error; err != nil {
			return err
		}
		if inst.IsActive {
			if err := tx.Model(&ZabbixInstance{}).Where("id = ?", id).Update("is_active", false).Error; err != nil {
				return err
			}
		}
		return tx.Delete(&ZabbixInstance{}, id).Error
	})
}

// TestAndUpdateZabbixInstance 测试并把结果写回数据库（版本/状态/时间/消息）
func TestAndUpdateZabbixInstance(id int64) (*ZabbixInstance, string, error) {
	var inst ZabbixInstance
	if err := DB.First(&inst, id).Error; err != nil {
		return nil, "", err
	}
	now := time.Now()
	ver, err := TestZabbixInstanceConfig(inst.WebURL, inst.User, inst.Pass, inst.Token)
	if err != nil {
		_ = DB.Model(&ZabbixInstance{}).Where("id = ?", id).Updates(map[string]interface{}{
			"last_test_ok":      false,
			"last_test_message": err.Error(),
			"last_test_at":      &now,
		}).Error
		inst.LastTestOk = false
		inst.LastTestMessage = err.Error()
		inst.LastTestAt = &now
		return &inst, "", err
	}
	_ = DB.Model(&ZabbixInstance{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_test_ok":      true,
		"last_test_message": "连接成功",
		"last_test_at":      &now,
		"version":           ver,
	}).Error
	inst.LastTestOk = true
	inst.LastTestMessage = "连接成功"
	inst.LastTestAt = &now
	inst.Version = ver
	return &inst, ver, nil
}

// ActivateZabbixInstance 设置某个实例为当前使用的 Zabbix
func ActivateZabbixInstance(id int64) (*ZabbixInstance, error) {
	var inst ZabbixInstance
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 只能激活 enabled 的实例
		if err := tx.First(&inst, id).Error; err != nil {
			return err
		}
		if !inst.Enabled {
			return errors.New("该 Zabbix 已禁用，无法设为当前")
		}
		if err := tx.Model(&ZabbixInstance{}).Where("is_active = ?", true).Update("is_active", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&ZabbixInstance{}).Where("id = ?", id).Update("is_active", true).Error; err != nil {
			return err
		}
		inst.IsActive = true
		return nil
	})
	if err != nil {
		return nil, err
	}
	// 立即应用到运行时全局 API（当前实现为全局生效）
	if err := ApplyActiveZabbix(&inst); err != nil {
		return nil, err
	}
	return &inst, nil
}

// GetActiveZabbixInstance 获取当前激活的实例
func GetActiveZabbixInstance() (*ZabbixInstance, error) {
	var inst ZabbixInstance
	err := DB.Where("is_active = ?", true).Order("id desc").First(&inst).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &inst, nil
}

// ApplyActiveZabbix 将实例应用到全局 Zabbix API 客户端
func ApplyActiveZabbix(inst *ZabbixInstance) error {
	if inst == nil {
		return nil
	}
	if !inst.Enabled {
		return errors.New("该 Zabbix 已禁用")
	}
	web := strings.TrimRight(strings.TrimSpace(inst.WebURL), "/")
	if web == "" {
		return nil
	}
	apiURL := web + "/api_jsonrpc.php"
	api := zabbix.NewAPI(apiURL)
	if strings.TrimSpace(inst.Token) != "" {
		api.SetAuth(strings.TrimSpace(inst.Token))
	} else if strings.TrimSpace(inst.User) != "" || strings.TrimSpace(inst.Pass) != "" {
		_, err := api.Login(strings.TrimSpace(inst.User), strings.TrimSpace(inst.Pass))
		if err != nil {
			return err
		}
	}
	API = api

	// 更新版本信息（失败不阻断，只是不更新）
	if ver, err := API.Version(); err == nil {
		ZBX_VER = ver
		// 同步写回实例版本/测试状态（不阻断）
		now := time.Now()
		_ = DB.Model(&ZabbixInstance{}).Where("id = ?", inst.ID).Updates(map[string]interface{}{
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

// TryInitZabbixFromDB 如果有激活实例，启动时自动初始化
func TryInitZabbixFromDB() {
	inst, err := GetActiveZabbixInstance()
	if err != nil || inst == nil {
		return
	}
	_ = ApplyActiveZabbix(inst)
}
