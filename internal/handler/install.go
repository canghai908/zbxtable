package handler

import (
	"crypto/tls"
	"database/sql"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"zbxtable/internal/model"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/response"

	"github.com/canghai908/zabbix-go"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"gopkg.in/ini.v1"
)

// ZabbixAPIError Zabbix API 错误
type ZabbixAPIError struct {
	Message string
}

func (e *ZabbixAPIError) Error() string {
	return e.Message
}

// checkZabbixAPISafe 安全地检查 Zabbix API（不退出程序）
func checkZabbixAPISafe(address, user, pass, token string) (string, error) {
	// TLS SkipVerify
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	// 判断API地址是否正确，http get访问api地址判断状态码是不是412
	addURL := address + "/api_jsonrpc.php"
	dClient := http.Client{
		Transport: transport,
		Timeout:   3 * time.Second,
	}
	resp, err := dClient.Get(addURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPreconditionFailed {
		return "", &ZabbixAPIError{Message: "Zabbix API 地址不正确，状态码: " + strconv.Itoa(resp.StatusCode)}
	}
	// api定义
	api := zabbix.NewAPI(addURL)
	if token != "" {
		api.SetAuth(token)
	} else {
		_, err := api.Login(user, pass)
		if err != nil {
			return "", err
		}
	}
	// zabbix api data get test
	outputPar := []string{"hostid", "host", "available", "status", "name", "error"}
	type params map[string]interface{}
	_, err = api.CallWithError("host.get", params{
		"output":  outputPar,
		"hostids": "10084",
	})
	if err != nil {
		return "", err
	}
	// version get
	version, err := api.Version()
	if err != nil {
		return "", err
	}
	return version, nil
}

// checkDatabaseConnection 检查数据库连接
func checkDatabaseConnection(dbdriver, dbhost, dbuser, dbpass, dbname, dbport string) error {
	switch dbdriver {
	case "mysql":
		dbURL := dbuser + ":" + dbpass + "@tcp(" + dbhost + ":" +
			dbport + ")/" + dbname + "?parseTime=true&loc=Asia%2FShanghai&timeout=5s&charset=utf8&collation=utf8_general_ci"
		db, err := sql.Open("mysql", dbURL)
		if err != nil {
			return err
		}
		defer db.Close()
		err = db.Ping()
		if err != nil {
			return err
		}
	case "postgresql":
		dbURL := "postgres://" + dbuser + ":" + url.QueryEscape(dbpass) + "@" + dbhost + ":" + dbport + "/" + dbname + "?sslmode=disable"
		connString, err := url.Parse(dbURL)
		if err != nil {
			return err
		}
		db, err := sql.Open("postgres", connString.String())
		if err != nil {
			return err
		}
		defer db.Close()
		err = db.Ping()
		if err != nil {
			return err
		}
	}
	return nil
}

// writeConfigFile 写入配置文件
func writeConfigFile(
	dbtype, dbhost, dbuser, dbpass, dbname, dbport,
	httpport, runmode, timeout string) error {
	cfg := ini.Empty()
	// zbxtable info
	//cfg.Section("").Key("appname").Comment = "zbxtable"
	// migrate httpport
	if httpport == "" {
		cfg.Section("").NewKey("httpport", "8088")
	} else {
		cfg.Section("").NewKey("httpport", httpport)
	}
	// migrate runmode
	if runmode == "" {
		cfg.Section("").NewKey("runmode", "prod")
	} else {
		cfg.Section("").NewKey("runmode", runmode)
	}
	// migrate timeout
	if timeout == "" {
		cfg.Section("").NewKey("timeout", "12")
	} else {
		cfg.Section("").NewKey("timeout", timeout)
	}
	// logger defaults (copyrequestbody 已废弃，不再写入)
	cfg.Section("").NewKey("log_level", "3")
	cfg.Section("").NewKey("log_path", "log")
	cfg.Section("").NewKey("maxlines", "1000")
	cfg.Section("").NewKey("maxsize", "0")
	cfg.Section("").NewKey("maxdays", "10")
	cfg.Section("").NewKey("daily", "true")
	// database info
	cfg.Section("").Key("dbtype").Comment = "database"
	cfg.Section("").NewKey("dbtype", dbtype)
	cfg.Section("").NewKey("dbhost", dbhost)
	cfg.Section("").NewKey("dbuser", dbuser)
	var newpass string
	if strings.Contains(dbpass, `#`) || strings.Contains(dbpass, `$`) || strings.Contains(dbpass, `!`) {
		newpass = strings.ReplaceAll(dbpass, "'", "")
	} else {
		newpass = dbpass
	}
	cfg.Section("").NewKey("dbpass", newpass)
	cfg.Section("").NewKey("dbname", dbname)
	cfg.Section("").NewKey("dbport", dbport)
	// zabbix info（安装阶段不再写入，用户可在系统设置里管理多个 Zabbix）
	// check
	confpath := "./config"
	_, err := os.Stat(confpath)
	if err != nil {
		err := os.MkdirAll(confpath, 0755)
		if err != nil {
			return err
		}
	}
	err = cfg.SaveTo("./config/app.conf")
	if err != nil {
		return err
	}
	return nil
}

// InstallStatusResponse 安装状态响应
type InstallStatusResponse struct {
	Installed bool `json:"installed"`
}

// GetInstallStatus 获取安装状态
func GetInstallStatus(c *gin.Context) {
	// 检查配置文件是否存在
	confPath := "./config/app.conf"
	_, err := os.Stat(confPath)
	if err != nil {
		response.Success(c, gin.H{
			"installed": false,
		})
		return
	}
	// 检查数据库连接是否配置
	cfg, err := ini.Load(confPath)
	if err != nil {
		response.Success(c, gin.H{
			"installed": false,
		})
		return
	}
	dbtype := cfg.Section("").Key("dbtype").String()
	dbname := cfg.Section("").Key("dbname").String()
	// 检查数据库类型和数据库名
	if dbtype == "" || dbname == "" {
		response.Success(c, gin.H{
			"installed": false,
		})
		return
	}
	// 检查 dbhost
	dbhost := cfg.Section("").Key("dbhost").String()
	if dbhost == "" {
		response.Success(c, gin.H{
			"installed": false,
		})
		return
	}
	// 这里简化处理，实际可以检查特定表是否存在
	response.Success(c, gin.H{
		"installed": true,
	})
}

// CheckDatabaseRequest 数据库检查请求
type CheckDatabaseRequest struct {
	DBType string `json:"dbtype" binding:"required"`
	DBHost string `json:"dbhost"`
	DBUser string `json:"dbuser"`
	DBPass string `json:"dbpass"`
	DBName string `json:"dbname" binding:"required"`
	DBPort string `json:"dbport"`
}

// CheckDatabase 检查数据库连接
func CheckDatabase(c *gin.Context) {
	var req CheckDatabaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 验证必需字段
	if req.DBHost == "" || req.DBUser == "" || req.DBPass == "" || req.DBPort == "" {
		response.BadRequest(c, "参数错误: 数据库需要 dbhost, dbuser, dbpass, dbport")
		return
	}

	err := checkDatabaseConnection(req.DBType, req.DBHost, req.DBUser, req.DBPass, req.DBName, req.DBPort)
	if err != nil {
		response.InternalError(c, "数据库连接失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "数据库连接成功", gin.H{
		"success": true,
	})
}

// CheckZabbixRequest Zabbix 检查请求（安装不再强制，仅用于后续系统设置里测试连接）
type CheckZabbixRequest struct {
	ZabbixWeb  string `json:"zabbix_web" binding:"required"`
	ZabbixUser string `json:"zabbix_user"`
	ZabbixPass string `json:"zabbix_pass"`
	Token      string `json:"token"`
}

// CheckZabbixAPI 检查 Zabbix API 连接
func CheckZabbixAPI(c *gin.Context) {
	var req CheckZabbixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	version, err := checkZabbixAPISafe(req.ZabbixWeb, req.ZabbixUser, req.ZabbixPass, req.Token)
	if err != nil {
		response.InternalError(c, "Zabbix API 连接失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "Zabbix API 连接成功", gin.H{
		"success": true,
		"version": version,
	})
}

// InstallRequest 安装请求
type InstallRequest struct {
	CheckDatabaseRequest
	HTTPPort string `json:"httpport"`
	RunMode  string `json:"runmode"`
	Timeout  string `json:"timeout"`
}

// DoInstall 执行安装
func DoInstall(c *gin.Context) {
	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 验证必需字段
	if req.DBHost == "" || req.DBUser == "" || req.DBPass == "" || req.DBPort == "" {
		response.BadRequest(c, "参数错误: 数据库需要 dbhost, dbuser, dbpass, dbport")
		return
	}

	// 检查是否已安装
	confPath := "./config/app.conf"
	_, err := os.Stat(confPath)
	if err == nil {
		// 配置文件已存在，检查是否已初始化数据库
		cfg, err := ini.Load(confPath)
		if err == nil {
			dbtype := cfg.Section("").Key("dbtype").String()
			if dbtype != "" {
				response.BadRequest(c, "系统已安装，如需重新安装请先删除配置文件")
				return
			}
		}
	}

	// 再次验证数据库连接
	err = checkDatabaseConnection(req.DBType, req.DBHost, req.DBUser, req.DBPass, req.DBName, req.DBPort)
	if err != nil {
		response.InternalError(c, "数据库连接失败: "+err.Error())
		return
	}

	// 设置默认值
	if req.HTTPPort == "" {
		req.HTTPPort = "8088"
	}
	if req.RunMode == "" {
		req.RunMode = "prod"
	}
	if req.Timeout == "" {
		req.Timeout = "12"
	}

	// 写入配置文件
	err = writeConfigFile(
		req.DBType, req.DBHost, req.DBUser, req.DBPass, req.DBName, req.DBPort,
		req.HTTPPort, req.RunMode, req.Timeout,
	)
	if err != nil {
		response.InternalError(c, "写入配置文件失败: "+err.Error())
		return
	}

	// 初始化数据库
	err = model.ModelInit(
		req.DBType,
		req.DBHost,
		req.DBUser,
		req.DBPass,
		req.DBName,
		req.DBPort)
	if err != nil {
		response.InternalError(c, "数据库初始化失败: "+err.Error())
		return
	}
	logger.Log.Info("配置文件已生成，数据库已初始化")

	// 检查端口是否发生变化
	currentPort := c.Request.Host
	if strings.Contains(currentPort, ":") {
		currentPort = strings.Split(currentPort, ":")[1]
	} else {
		currentPort = "8088" // 默认端口
	}
	portChanged := req.HTTPPort != "" && req.HTTPPort != currentPort

	response.SuccessWithMessage(c, "安装成功！请手动重启程序以加载配置", gin.H{
		"success":      true,
		"need_restart": true,
		"port_changed": portChanged,
		"old_port":     currentPort,
		"new_port":     req.HTTPPort,
	})
}
