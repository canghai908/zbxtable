package models

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
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
	ZBX_VER    string
	ZBX_V      bool
	Version    string
	GitHash    string
	BuildTime  string
	AssetsHost string
	WeApp      = &workwx.WorkwxApp{}

	envConfig     map[string]string
	envConfigOnce sync.Once
	envFileExists bool
)

// TableName 表名前缀
func TableName(str string) string {
	return fmt.Sprintf("%s%s", "zbxtable_", str)
}

// GetAssetsHost
func GetAssetsHost() string {
	AssetsHost = GetConfKey("AssetsHost")
	if AssetsHost == "" {
		AssetsHost = "http://dl.cactifans.com/assets/"
	}
	return AssetsHost
}

// ModelsInit  p
func ModelsInit(zabbix_web, zabbix_user, zabbix_pass, zabbix_token,
	dbtype, dbhost, dbuser, dbpass, dbname, dbport string) {

	//GetAssetsHost
	GetAssetsHost()

	// 直接使用 GORM
	runmode := GetConfKey("runmode")
	err := InitGormDB(dbtype, dbhost, dbuser, dbpass, dbname, dbport, runmode)
	if err != nil {
		utils.Log.Error("Failed to connect database: ", err)
		os.Exit(1)
	}
	utils.Log.Info("Database connected!")

	// 自动迁移表
	err = AutoMigrate()
	if err != nil {
		utils.Log.Error("Failed to auto migrate: ", err)
		os.Exit(1)
	}

	// 基础数据初始化
	DatabaseInit()
	// 使用go-cache替代Redis（不再需要Redis连接）
	InitCache()

	// 尝试从数据库加载“当前激活的 Zabbix”
	TryInitZabbixFromDB()

	// 安装后首次启动允许不配置 Zabbix：跳过 Zabbix 初始化
	if strings.TrimSpace(zabbix_web) == "" {
		utils.Log.Info("Zabbix is not configured, skipping Zabbix initialization")
		return
	}

	//TLS SkipVerify
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//判断API地址是否正确，http get访问访问api地址判断状态码是不是412
	addURL := zabbix_web + "/api_jsonrpc.php"
	dClient := http.Client{
		Transport: transport,
		Timeout:   3 * time.Second, // 设置超时时间为 3 秒
	}
	resp, err := dClient.Get(addURL)
	if err != nil {
		utils.Log.Error("Zabbix Web get request failed:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPreconditionFailed {
		utils.Log.Error("Zabbix Web is incorrectly!")
		os.Exit(1)
	}
	//api变量
	API = zabbix.NewAPI(zabbix_web + "/api_jsonrpc.php")
	if zabbix_token != "" {
		API.SetAuth(zabbix_token)
	} else {
		_, err = API.Login(zabbix_user, zabbix_pass)
		if err != nil {
			utils.Log.Error(err)
			os.Exit(1)
		}
	}
	//zabbix api data get test
	OutputPar := []string{"hostid", "host", "name", "error"}
	type params map[string]interface{}
	_, err = API.CallWithError("host.get", params{
		"output":  OutputPar,
		"hostids": "10084",
	})
	if err != nil {
		utils.Log.Error("connect Zabbix API failed:", err)
		os.Exit(1)
	}
	//Zabbix version
	ZBX_VER, err = API.Version()
	if err != nil {
		utils.Log.Error(err)
		os.Exit(1)
	}
	verArr := strings.Split(ZBX_VER, ".")
	ZbxMasterVer, _ := strconv.ParseInt(verArr[0], 10, 64)
	ZbxMiddleVer, _ := strconv.ParseInt(verArr[1], 10, 64)
	if ZbxMasterVer >= 6 || (ZbxMasterVer == 5 && ZbxMiddleVer == 4) {
		ZBX_V = true
	} else {
		ZBX_V = false
	}
	utils.Log.Info("Zabbix API connected！Zabbix version:", ZBX_VER)
	//	zabbix web login (only if token is not configured)
	if zabbix_pass != "" {
		LoginZabbixWeb(zabbix_web, zabbix_user, zabbix_pass)
	} else {
		utils.Log.Info("Zabbix pass is not configured, skipping web login")
	}

	//gen tpl (企业微信配置优先从系统配置表读取，其次回退到 app.conf)
	agentIDStr := GetConfigValueByKey("wechat_agentid", GetConfKey("wechat_agentid"))
	if agentIDStr == "" {
		utils.Log.Info("wechat_agentid is empty, WeChat app will not be initialized")
	} else {
		AgentId, err := strconv.ParseInt(agentIDStr, 10, 64)
		if err != nil {
			utils.Log.Error("wechat_agentid parse error:", err)
			os.Exit(1)
		}
		corpid := GetConfigValueByKey("wechat_corpid", GetConfKey("wechat_corpid"))
		secret := GetConfigValueByKey("wechat_secret", GetConfKey("wechat_secret"))
		if corpid == "" || secret == "" {
			utils.Log.Info("wechat_corpid or wechat_secret is empty, WeChat app will not be initialized")
		} else {
			client := workwx.New(corpid)
			WeApp = client.WithApp(secret, AgentId)
			WeApp.SpawnAccessTokenRefresher()
			utils.Log.Info("WeChat inited!")
		}
	}
}

// DatabaseInit 数据初始化
func DatabaseInit() {
	//数据初始化操作 - 使用 GORM
	var v Manager
	err := DB.Where("username = ?", "admin").First(&v).Error
	//检查权限
	if err == nil && v.Operation == "" {
		err := DB.Model(&Manager{}).Where("id = ?", v.ID).Update("operation", "['add', 'edit', 'delete','update']").Error
		if err != nil {
			utils.Log.Info(err)
			return
		}
		utils.Log.Info("update admin operation successfully")
	}
	//添加管理员账号
	if errors.Is(err, gorm.ErrRecordNotFound) {
		utils.Log.Info("the admin user does not exist, create a new admin account later!")
		var manager Manager
		manager.Username = "admin"
		manager.Password, _ = utils.PasswordHash("Zbxtable")
		manager.Avatar = "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif"
		manager.Role = "admin"
		manager.Operation = "['add', 'edit', 'delete','update']"
		manager.Status = 0
		err := DB.Create(&manager).Error
		if err != nil {
			utils.Log.Info(err)
			return
		}
		utils.Log.Info("create an administrator account successfully, the admin ID is:", manager.ID)
	}
	//初始化系统数据
	var cnt []System
	err = DB.Find(&cnt).Error
	if err != nil {
		utils.Log.Info(err)
		return
	}
	if len(cnt) == 0 {
		sys := []System{
			{Name: "Linux操作系统", Status: 0},
			{Name: "Windows操作系统", Status: 0},
			{Name: "网络设备", Status: 0},
			{Name: "物理服务器", Status: 0},
		}
		err := DB.Create(&sys).Error
		if err != nil {
			utils.Log.Info("Init system info error！")
			return
		}
		utils.Log.Info("Init system data successfully!")
	}
	//出口
	//初始化系统数据
	var cne []Egress
	err = DB.Find(&cne).Error
	if err != nil {
		utils.Log.Info(err)
		return
	}
	if len(cne) == 0 {
		egress := []Egress{
			{NameOne: "电信100M", NameTwo: "移动100M", Status: 0},
		}
		err := DB.Create(&egress).Error
		if err != nil {
			utils.Log.Info("Init egress info error！")
			return
		}
		utils.Log.Info("Init egress data successfully!")
	}
	// 默认配置初始化（包括面板、邮件、微信、Ollama 等）
	defaultConfigs := []Config{
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
		{Name: "企业微信 AgentID", Key: "wechat_agentid", Value: "", Comment: "企业微信应用的 AgentID"},
		{Name: "企业微信 CorpID", Key: "wechat_corpid", Value: "", Comment: "企业微信企业ID"},
		{Name: "企业微信 Secret", Key: "wechat_secret", Value: "", Comment: "企业微信应用的 Secret"},
		// Ollama 配置
		{Name: "Ollama Host", Key: "ollama_host", Value: "http://localhost:11434", Comment: "Ollama 服务地址，如 http://127.0.0.1:11434"},
		{Name: "Ollama Model", Key: "ollama_model", Value: "deepseek-r1:32b", Comment: "默认使用的大模型名称"},
	}

	for _, cfgItem := range defaultConfigs {
		var existing Config
		err = DB.Where("`key` = ?", cfgItem.Key).First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			if insertErr := DB.Create(&cfgItem).Error; insertErr != nil {
				utils.Log.Infof("Init config key %s error: %v", cfgItem.Key, insertErr)
				continue
			}
			utils.Log.Infof("Init config key %s successfully!", cfgItem.Key)
		}
	}
	//默认菜单初始化
	InitMenuData()
	//检查并添加缺失的菜单项（用于版本升级）
	CheckAndAddMenus()
	//告警默认规则初始化 default rule
	var cRule []Rule
	err = DB.Where("m_type = ?", "2").Find(&cRule).Error
	if err != nil {
		utils.Log.Info(err)
		return
	}
	if len(cRule) == 0 {
		// 使用 "*" 作为全局默认规则，匹配所有租户
		defaultRule := []Rule{
			{
				Name:     "全局默认规则",
				TenantID: "*",
				MType:    "2",
				Channel:  "wechat_robot",
				UserIds:  "1",
				Status:   "0"},
		}
		err := DB.Create(&defaultRule).Error
		if err != nil {
			utils.Log.Info("Init default rule error！", err)
			return
		}
		utils.Log.Info("Init default rule successfully!")
	}
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
		utils.Log.Error(err)
		utils.Log.Error("Please run 'zbxtable init' to create app.conf")
		return ""
	}
	p, err := cfg.Section("").GetKey(v)
	if err != nil {
		utils.Log.Error(err)
		return ""
	}
	return p.String()
}

// IsPasswordConfigured 检查是否配置了密码（用于图形查看）
// 优先检查数据库中是否有激活的 Zabbix 实例配置
// 如果数据库中没有，则回退到配置文件检查
func IsPasswordConfigured() bool {
	// 1. 优先检查数据库中是否有激活的 Zabbix 实例
	var activeInstance ZabbixInstance
	err := DB.Where("is_active = ?", true).First(&activeInstance).Error
	if err == nil {
		// 找到激活的实例，检查是否配置了密码或 token
		if activeInstance.Token != "" || activeInstance.Pass != "" {
			return true
		}
	}

	// 2. 如果数据库中没有激活实例，检查是否有任何可用的实例
	var anyInstance ZabbixInstance
	err = DB.Where("enabled = ?", true).First(&anyInstance).Error
	if err == nil {
		// 找到可用的实例，检查是否配置了密码或 token
		if anyInstance.Token != "" || anyInstance.Pass != "" {
			return true
		}
	}

	// 3. 回退到配置文件检查（兼容旧版本）
	pass := GetConfKey("zabbix_pass")
	token := GetConfKey("zabbix_token")
	return pass != "" || token != ""
}
