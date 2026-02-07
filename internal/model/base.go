package model

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"

	zabbix "github.com/canghai908/zabbix-go"
	_ "github.com/go-sql-driver/mysql"
	jsoniter "github.com/json-iterator/go"
	_ "github.com/lib/pq"
	workwx "github.com/xen0n/go-workwx"
	ini "gopkg.in/ini.v1"
	"gorm.io/gorm"
)

var (
	API  = &zabbix.API{}
	json = jsoniter.Config{
		EscapeHTML:             false,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
	}.Froze()
	ZBX_VER   string
	ZBX_V     bool
	Version   string
	GitHash   string
	BuildTime string
	WeApp     *workwx.WorkwxApp // 不初始化，默认为 nil

	envConfig     map[string]string
	envConfigOnce sync.Once
	envFileExists bool
)

const (
	//ms-agent 常量定义
	MSName   = "ms-agent"
	MSUser   = "ms-agent"
	MSGroup  = "MS-Agent"
	MSMedia  = "MS-Agent"
	MSAction = "MS-Agent"
	//webhook 常量定义
	WebhookName   = "ZbxTable"
	WebhookAction = "ZbxTable Webhook"
	WebhookGroup  = "ZbxTable Webhook"
	WebhookUser   = "zbxtable-webhook"
	//down
	DownloadPath = "./download/"
	TplPath      = "./assets/templates/"
)

// TableName 表名前缀
func TableName(str string) string {
	return fmt.Sprintf("%s%s", "zbxtable_", str)
}

// 接收数据库信息初始化
func ModelInit(dbtype, dbhost, dbuser, dbpass, dbname, dbport string) {
	// 直接使用 GORM
	runmode := GetConfKey("runmode")
	err := InitGormDB(dbtype, dbhost, dbuser, dbpass, dbname, dbport, runmode)
	if err != nil {
		logger.Log.Error("Failed to connect database: ", err)
		os.Exit(1)
	}
	logger.Log.Info("Database connected!")
	// 自动迁移表
	err = AutoMigrate()
	if err != nil {
		logger.Log.Error("Failed to auto migrate: ", err)
		os.Exit(1)
	}
	// 基础数据初始化
	DatabaseInit()
	// 使用go-cache替代Redis（不再需要Redis连接）
	InitCache()
}

func InitWechat() {
	// 首先检查企业微信开关是否启用
	wechatEnabled := GetConfigValueByKey("wechat_enabled", "0")
	if wechatEnabled != "1" {
		logger.Log.Info("WeChat is disabled in system config, skipping WeChat initialization")
		WeApp = nil // 明确设置为 nil
		return
	}
	// 开关已启用，从数据库读取企业微信配置
	agentIDStr := GetConfigValueByKey("wechat_agentid", "")
	corpid := GetConfigValueByKey("wechat_corpid", "")
	secret := GetConfigValueByKey("wechat_secret", "") // 这里会自动解密

	// 验证必填配置项
	if agentIDStr == "" || corpid == "" || secret == "" {
		logger.Log.Info("WeChat is enabled but configuration is incomplete (agentid/corpid/secret), WeChat app will not be initialized")
		WeApp = nil // 明确设置为 nil
		return
	}
	// 解析 AgentID
	AgentId, err := strconv.ParseInt(agentIDStr, 10, 64)
	if err != nil {
		logger.Log.Error("wechat_agentid parse error:", err)
		WeApp = nil // 明确设置为 nil
		return
	}

	// 初始化企业微信客户端
	client := workwx.New(corpid)
	WeApp = client.WithApp(secret, AgentId)
	WeApp.SpawnAccessTokenRefresher()
	logger.Log.Info("WeChat inited successfully!")
}

// DatabaseInit 数据初始化
func DatabaseInit() {
	//数据初始化操作 - 使用 GORM
	var v User
	err := DB.Where("username = ?", "admin").First(&v).Error
	//检查权限
	if err == nil && v.Operation == "" {
		err := DB.Model(&User{}).Where("id = ?", v.ID).Update("operation", "['add', 'edit', 'delete','update']").Error
		if err != nil {
			logger.Log.Info(err)
			return
		}
		logger.Log.Info("update admin operation successfully")
	}
	//添加管理员账号
	if errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Log.Info("the admin user does not exist, create a new admin account later!")

		// 构建默认主题配置（使用结构化方式，避免手写 JSON 出错）
		defaultThemeConfig := map[string]interface{}{
			"theme": map[string]string{
				"color":   "#1890ff",
				"mode":    "dark",
				"success": "#52c41a",
				"warning": "#faad14",
				"error":   "#f5222f",
			},
			"animate": map[string]interface{}{
				"disabled":  true,
				"name":      "lightSpeed",
				"direction": "left",
			},
		}

		// 将配置转换为 JSON 字符串
		themeJSON, err := json.Marshal(defaultThemeConfig)
		if err != nil {
			logger.Log.Error("Failed to marshal default theme config:", err)
			// 如果序列化失败，使用备用的硬编码 JSON
			themeJSON = []byte(`{"theme":{"color":"#1890ff","mode":"dark","success":"#52c41a","warning":"#faad14","error":"#f5222f"},"animate":{"disabled":true,"name":"lightSpeed","direction":"left"}}`)
		}

		var user User
		user.Username = "admin"
		user.Password, _ = utils.PasswordHash("Zbxtable")
		user.Avatar = "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif"
		user.Role = "admin"
		user.Operation = "['add', 'edit', 'delete','update']"
		user.Theme = string(themeJSON)
		user.Status = 0
		err = DB.Create(&user).Error
		if err != nil {
			logger.Log.Info(err)
			return
		}
		logger.Log.Info("create an administrator account successfully, the admin ID is:", user.ID)
	}
	//初始化系统数据
	var cnt []System
	err = DB.Find(&cnt).Error
	if err != nil {
		logger.Log.Info(err)
		return
	}
	if len(cnt) == 0 {
		now := time.Now()
		sys := []System{
			{Name: "Linux操作系统", Status: 0, InitedAt: &now},
			{Name: "Windows操作系统", Status: 0, InitedAt: &now},
			{Name: "网络设备", Status: 0, InitedAt: &now},
			{Name: "物理服务器", Status: 0, InitedAt: &now},
		}
		err := DB.Create(&sys).Error
		if err != nil {
			logger.Log.Info("Init system info error！", err)
			return
		}
		logger.Log.Info("Init system data successfully!")
	}
	//告警默认规则初始化 default rule
	var rules []Rule
	err = DB.Where("m_type = ?", "2").Find(&rules).Error
	if err != nil {
		logger.Log.Info(err)
		return
	}
	if len(rules) == 0 {
		// 使用 "*" 作为全局默认规则，匹配所有租户
		defaultRule := []Rule{
			{
				Name:    "全局默认规则",
				ZIDs:    "*",
				MType:   "2",
				Channel: "wechat_robot",
				UserIds: "1",
				Sweek:   "0,1,2,3,4,5,6",
				Stime:   "00:00",
				Etime:   "23.59",
				Status:  "0"},
		}
		err := DB.Create(&defaultRule).Error
		if err != nil {
			logger.Log.Info("Init default rule error！", err)
			return
		}
		logger.Log.Info("Init default rule successfully!")
	}
	// 初始化加密密钥（如果不存在）
	var encryptionKeyConfig Config
	err = DB.Where("`key` = ?", "encryption_key").First(&encryptionKeyConfig).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 生成随机加密密钥
		randomKey := utils.GenerateRandomKey()
		encryptionKeyConfig = Config{
			Name:    "加密密钥",
			Key:     "encryption_key",
			Value:   randomKey,
			Comment: "用于加密存储敏感信息（如Zabbix密码和Token）的密钥，系统自动生成，不可修改",
		}
		if insertErr := DB.Create(&encryptionKeyConfig).Error; insertErr != nil {
			logger.Log.Error("Init encryption_key error:", insertErr)
		} else {
			logger.Log.Info("Init encryption_key successfully! Key length:", len(randomKey))
		}
	}

	defaultConfigs := []Config{
		// 系统外观配置
		{Name: "系统名称", Key: "system_name", Value: "ZbxTable", Comment: "系统显示的名称"},
		{Name: "系统Logo", Key: "system_logo", Value: "data:image/png;base64,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", Comment: "系统Logo的base64编码数据"},
		// 初始配置状态
		{Name: "初始配置完成", Key: "initial_setup_completed", Value: "0", Comment: "标记系统是否完成初始配置：1 已完成,0 未完成"},
		// Dashboard 相关
		{Name: "数据面板", Key: "zbx_dash", Value: "0", Comment: "是否开启Zabbix看板：1 开启,0 关闭"},
		{Name: "面板配置", Key: "dash_id", Value: "1", Comment: "需要引入的Zabbix面板的ID，默认为1"},
		{Name: "主机分类同步", Key: "sync_inventory", Value: "1", Comment: "主机分类同步计划任务是否启用：1 启用,0 不启用"},
		{Name: "Webhook回调地址", Key: "webhook_url", Value: "", Comment: "webhook通知地址"},
		// 邮件配置
		{Name: "邮件发件人", Key: "email_from", Value: "", Comment: "告警邮件发件人邮箱地址"},
		{Name: "邮件昵称", Key: "email_nickname", Value: "ZbxTable", Comment: "告警邮件显示的发件人昵称"},
		{Name: "SMTP 密码/授权码", Key: "email_secret", Value: "", Comment: "SMTP 登录密码或授权码"},
		{Name: "SMTP 服务器", Key: "email_host", Value: "smtp.qq.com", Comment: "SMTP 服务器地址"},
		{Name: "SMTP 端口", Key: "email_port", Value: "465", Comment: "SMTP 端口号"},
		{Name: "SMTP 使用 SSL", Key: "email_isSSl", Value: "true", Comment: "是否启用 SSL：true/false"},
		// 企业微信配置
		{Name: "企业微信开关", Key: "wechat_enabled", Value: "0", Comment: "是否启用企业微信：1 启用,0 禁用"},
		{Name: "企业微信 AgentID", Key: "wechat_agentid", Value: "", Comment: "企业微信应用的 AgentID"},
		{Name: "企业微信 CorpID", Key: "wechat_corpid", Value: "", Comment: "企业微信企业ID"},
		{Name: "企业微信 Secret", Key: "wechat_secret", Value: "", Comment: "企业微信应用的 Secret"},
		// AI 配置
		{Name: "AI 类型", Key: "ai_type", Value: "ollama", Comment: "选择使用的 AI 服务类型：ollama 或 deepseek"},
		// Ollama 配置
		{Name: "Ollama Host", Key: "ollama_host", Value: "http://localhost:11434", Comment: "Ollama 服务地址，如 http://127.0.0.1:11434"},
		{Name: "Ollama Model", Key: "ollama_model", Value: "deepseek-r1:32b", Comment: "默认使用的大模型名称"},
		// Deepseek 配置
		{Name: "Deepseek API Key", Key: "deepseek_api_key", Value: "", Comment: "Deepseek API 密钥"},
		{Name: "Deepseek Model", Key: "deepseek_model", Value: "deepseek-chat", Comment: "Deepseek 模型名称，如 deepseek-chat"},
		{Name: "Deepseek Base URL", Key: "deepseek_base_url", Value: "https://api.deepseek.com", Comment: "Deepseek API 地址，默认为 https://api.deepseek.com"},
	}

	for _, cfgItem := range defaultConfigs {
		var existing Config
		err = DB.Where("`key` = ?", cfgItem.Key).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if insertErr := DB.Create(&cfgItem).Error; insertErr != nil {
				logger.Log.Infof("Init config key %s error: %v", cfgItem.Key, insertErr)
				continue
			}
			logger.Log.Infof("Init config key %s successfully!", cfgItem.Key)
		}
	}
	//默认菜单初始化
	InitMenuData()
	//检查并添加缺失的菜单项（用于版本升级）
	CheckAndAddMenus()

}

func GetConfKey(v string) string {
	// 优先从 .env 中读取（便于开发环境配置敏感信息且不提交到 git）
	envConfigOnce.Do(func() {
		file, err := os.Open(".env")
		if err != nil {
			// 没有 .env 文件时直接跳过，后面回退到 app.conf
			envFileExists = false
			return
		}
		defer file.Close()

		envFileExists = true
		envConfig = make(map[string]string)

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			// 去掉可能的引号
			val = strings.Trim(val, `"'`)
			if key != "" {
				envConfig[key] = val // 即使 val 为空也认为是有效配置
			}
		}
	})

	// 如果存在 .env 文件，则所有配置完全由 .env 决定：
	// - key 存在：返回对应值（可以是空字符串）
	// - key 不存在：返回空字符串
	if envFileExists {
		if envConfig == nil {
			return ""
		}
		if val, ok := envConfig[v]; ok {
			return val
		}
		return ""
	}
	// 其次从 config/app.conf 读取
	cfg, err := ini.Load("./config/app.conf")
	if err != nil {
		logger.Log.Error(err)
		return ""
	}
	p, err := cfg.Section("").GetKey(v)
	if err != nil {
		return ""
	}
	return p.String()
}

// IsPasswordConfigured 检查是否配置了密码（用于图形查看）
// 优先检查数据库中是否有激活的 Zabbix 实例配置了用户名和密码
// 注意：查看图形需要用户名和密码，Token 方式无法用于 Web 登录
func IsPasswordConfigured() bool {
	// 1. 优先检查数据库中是否有激活的 Zabbix 实例
	var activeInstance ZabbixInstance
	err := DB.Where("is_active = ?", true).First(&activeInstance).Error
	if err == nil {
		// 找到激活的实例，检查是否配置了用户名和密码（查看图形必须用密码，Token 无法用于 Web 登录）
		if strings.TrimSpace(activeInstance.User) != "" && strings.TrimSpace(activeInstance.Pass) != "" {
			return true
		}
	}

	// 2. 如果数据库中没有激活实例，检查是否有任何可用的实例配置了密码
	var anyInstance ZabbixInstance
	err = DB.Where("enabled = ?", true).First(&anyInstance).Error
	if err == nil {
		// 找到可用的实例，检查是否配置了用户名和密码
		if strings.TrimSpace(anyInstance.User) != "" && strings.TrimSpace(anyInstance.Pass) != "" {
			return true
		}
	}

	// 3. 回退到配置文件检查（兼容旧版本）
	pass := GetConfKey("zabbix_pass")
	user := GetConfKey("zabbix_user")
	return pass != "" && user != ""
}

// GetEncryptionKey 获取加密密钥
// 优先从数据库读取，如果数据库中没有则从环境变量读取，最后使用默认密钥
func GetEncryptionKey() string {
	// 1. 优先从数据库读取
	key := GetConfigValueByKey("encryption_key", "")
	if key != "" {
		return key
	}

	// 2. 从环境变量或配置文件读取
	key = GetConfKey("encryption_key")
	if key != "" {
		return key
	}

	// 3. 使用默认密钥（不应该到这里，因为 DatabaseInit 会生成）
	logger.Log.Warn("使用默认加密密钥，建议在数据库中配置 encryption_key")
	return "zbxtable-default-encryption-key-2024"
}
