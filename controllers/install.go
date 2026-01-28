package controllers

import (
	"crypto/tls"
	"database/sql"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"zbxtable/models"

	"github.com/canghai908/zabbix-go"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
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
	case "sqlite":
		// SQLite 使用文件路径作为数据库名
		dbPath := dbname
		if dbPath == "" {
			dbPath = "./data/zbxtable.db"
		}
		// 确保目录存在
		if err := os.MkdirAll("./data", 0755); err != nil {
			return err
		}
		// SQLite 不需要连接测试，文件会在首次使用时创建
		// 这里只检查目录是否可写
		return nil
	}
	return nil
}

// writeConfigFile 写入配置文件
func writeConfigFile(zabbix_web, zabbix_user, zabbix_pass,
	dbtype, dbhost, dbuser, dbpass, dbname, dbport,
	httpport, runmode, timeout, token string) error {
	cfg := ini.Empty()
	// zbxtable info
	cfg.Section("").Key("appname").Comment = "zbxtable"
	// migrate httpport
	if httpport == "" {
		cfg.Section("").NewKey("httpport", "8085")
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
	cfg.Section("").NewKey("appname", "zbxtable")
	cfg.Section("").NewKey("token", token)
	// logger defaults (copyrequestbody 已废弃，不再写入)
	cfg.Section("").NewKey("log_level", "6")
	cfg.Section("").NewKey("log_path", "logs/app.log")
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
	confpath := "./conf"
	_, err := os.Stat(confpath)
	if err != nil {
		err := os.MkdirAll(confpath, 0755)
		if err != nil {
			return err
		}
	}
	err = cfg.SaveTo("./conf/app.conf")
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
	confPath := "./conf/app.conf"
	_, err := os.Stat(confPath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "ok",
			"data": gin.H{
				"installed": false,
			},
		})
		return
	}

	// 检查数据库连接是否配置
	cfg, err := ini.Load(confPath)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "ok",
			"data": gin.H{
				"installed": false,
			},
		})
		return
	}

	dbtype := cfg.Section("").Key("dbtype").String()
	dbhost := cfg.Section("").Key("dbhost").String()
	dbname := cfg.Section("").Key("dbname").String()

	// sqlite 不需要 dbhost
	if dbtype == "" || dbname == "" || (dbtype != "sqlite" && dbhost == "") {
		c.JSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "ok",
			"data": gin.H{
				"installed": false,
			},
		})
		return
	}

	// 尝试连接数据库，检查表是否存在
	// 这里简化处理，实际可以检查特定表是否存在
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "ok",
		"data": gin.H{
			"installed": true,
		},
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
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
			"data": gin.H{
				"success": false,
			},
		})
		return
	}

	// 对于非 SQLite 数据库，验证必需字段
	if req.DBType != "sqlite" {
		if req.DBHost == "" || req.DBUser == "" || req.DBPass == "" || req.DBPort == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "参数错误: " + req.DBType + " 数据库需要 dbhost, dbuser, dbpass, dbport",
				"data": gin.H{
					"success": false,
				},
			})
			return
		}
	}

	err := checkDatabaseConnection(req.DBType, req.DBHost, req.DBUser, req.DBPass, req.DBName, req.DBPort)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "数据库连接失败: " + err.Error(),
			"data": gin.H{
				"success": false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "数据库连接成功",
		"data": gin.H{
			"success": true,
		},
	})
}

// CheckZabbixRequest Zabbix 检查请求（安装不再强制，仅用于后续系统设置里测试连接）
type CheckZabbixRequest struct {
	ZabbixWeb   string `json:"zabbix_web" binding:"required"`
	ZabbixUser  string `json:"zabbix_user"`
	ZabbixPass  string `json:"zabbix_pass"`
	ZabbixToken string `json:"zabbix_token"`
}

// CheckZabbixAPI 检查 Zabbix API 连接
func CheckZabbixAPI(c *gin.Context) {
	var req CheckZabbixRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	version, err := checkZabbixAPISafe(req.ZabbixWeb, req.ZabbixUser, req.ZabbixPass, req.ZabbixToken)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "Zabbix API 连接失败: " + err.Error(),
			"data": gin.H{
				"success": false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Zabbix API 连接成功",
		"data": gin.H{
			"success": true,
			"version": version,
		},
	})
}

// CheckRedisRequest Redis 检查请求
type CheckRedisRequest struct {
	RedisHost string `json:"redis_host" binding:"required"`
	RedisPort string `json:"redis_port" binding:"required"`
	RedisPass string `json:"redis_pass"`
	RedisDB   string `json:"redis_db"`
}

// InstallRedisRequest 安装时的 Redis 请求（可选）
type InstallRedisRequest struct {
	RedisHost string `json:"redis_host"`
	RedisPort string `json:"redis_port"`
	RedisPass string `json:"redis_pass"`
	RedisDB   string `json:"redis_db"`
}

// CheckRedis 检查 Redis 连接
func CheckRedis(c *gin.Context) {
	var req CheckRedisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 设置默认值
	if req.RedisPort == "" {
		req.RedisPort = "6379"
	}
	if req.RedisDB == "" {
		req.RedisDB = "0"
	}

	err := checkRedisConnection(req.RedisHost, req.RedisPort, req.RedisPass, req.RedisDB)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    500,
			"message": "Redis 连接失败: " + err.Error(),
			"data": gin.H{
				"success": false,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Redis 连接成功",
		"data": gin.H{
			"success": true,
		},
	})
}

// checkRedisConnection 检查 Redis 连接（已废弃，现在使用go-cache，不再需要Redis）
func checkRedisConnection(redisHost, redisPort, redisPass, redisDB string) error {
	// 使用go-cache后不再需要Redis连接，直接返回成功
	// 保留此函数以保持API兼容性
	return nil
}

// InstallRequest 安装请求
type InstallRequest struct {
	CheckDatabaseRequest
	InstallRedisRequest
	HTTPPort string `json:"httpport"`
	RunMode  string `json:"runmode"`
	Timeout  string `json:"timeout"`
}

// DoInstall 执行安装
func DoInstall(c *gin.Context) {
	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "参数错误: " + err.Error(),
			"data": gin.H{
				"success": false,
			},
		})
		return
	}

	// 检查是否已安装
	confPath := "./conf/app.conf"
	_, err := os.Stat(confPath)
	if err == nil {
		// 配置文件已存在，检查是否已初始化数据库
		cfg, err := ini.Load(confPath)
		if err == nil {
			dbtype := cfg.Section("").Key("dbtype").String()
			if dbtype != "" {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "系统已安装，如需重新安装请先删除配置文件",
					"data": gin.H{
						"success": false,
					},
				})
				return
			}
		}
	}

	// 再次验证数据库连接
	err = checkDatabaseConnection(req.DBType, req.DBHost, req.DBUser, req.DBPass, req.DBName, req.DBPort)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    500,
			"message": "数据库连接失败: " + err.Error(),
			"data": gin.H{
				"success": false,
			},
		})
		return
	}

	// 安装阶段不再强制验证 Zabbix（用户可在系统设置里添加多个 Zabbix）

	// 生成 token
	token := strings.ReplaceAll(uuid.New().String(), "-", "")

	// 设置默认值
	if req.HTTPPort == "" {
		req.HTTPPort = "8085"
	}
	if req.RunMode == "" {
		req.RunMode = "prod"
	}
	if req.Timeout == "" {
		req.Timeout = "12"
	}

	// 设置 Redis 默认值（在写入配置文件之前）
	if req.RedisHost == "" {
		req.RedisHost = "localhost"
	}
	if req.RedisPort == "" {
		req.RedisPort = "6379"
	}
	if req.RedisDB == "" {
		req.RedisDB = "0"
	}

	// 写入配置文件
	err = writeConfigFile(
		"", "", "",
		req.DBType, req.DBHost, req.DBUser, req.DBPass, req.DBName, req.DBPort,
		req.HTTPPort, req.RunMode, req.Timeout, token,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "写入配置文件失败: " + err.Error(),
			"data": gin.H{
				"success": false,
			},
		})
		return
	}

	// 再次验证 Redis 连接（如果提供了配置）
	// 注意：Redis 连接失败不应该阻止安装，只记录警告
	if req.RedisHost != "" && req.RedisPort != "" {
		err = checkRedisConnection(req.RedisHost, req.RedisPort, req.RedisPass, req.RedisDB)
		if err != nil {
			// Redis 连接失败只记录日志，不阻止安装
			// 用户可以在安装后配置 Redis
		}
	}

	// 初始化数据库
	models.ModelsInit(
		"", "", "", "",
		req.DBType, req.DBHost, req.DBUser, req.DBPass, req.DBName, req.DBPort,
		req.RedisHost, req.RedisPort, req.RedisPass, req.RedisDB,
	)

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "安装成功",
		"data": gin.H{
			"success": true,
		},
	})
}
