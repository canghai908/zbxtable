package cmd

import (
	"bufio"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	v1 "zbxtable/api/v1"
	"zbxtable/pkg/assets"
	"zbxtable/pkg/logger"

	model "zbxtable/internal/model"

	zabbix "github.com/canghai908/zabbix-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
)

const motd = `
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║   ███████╗██████╗ ██╗  ██╗████████╗ █████╗ ██████╗ ██╗     ║
║   ╚══███╔╝██╔══██╗╚██╗██╔╝╚══██╔══╝██╔══██╗██╔══██╗██║     ║
║     ███╔╝ ██████╔╝ ╚███╔╝    ██║   ███████║██████╔╝██║     ║
║    ███╔╝  ██╔══██╗ ██╔██╗    ██║   ██╔══██║██╔══██╗██║     ║
║   ███████╗██████╔╝██╔╝ ██╗   ██║   ██║  ██║██████╔╝███████╗║
║   ╚══════╝╚═════╝ ╚═╝  ╚═╝   ╚═╝   ╚═╝  ╚═╝╚═════╝ ╚══════╝║
║                                                               ║
║              Zabbix Table Management System                   ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
`

var (
	//Web 配置
	Web = &cli.Command{
		Name:   "web",
		Usage:  "Start web server",
		Action: runWeb,
	}

	envConfig     map[string]string
	initEnvOnce   sync.Once
	envFileExists bool
	webCfg        *ini.File
	webAPI        *zabbix.API
)

// checkInstallStatus 检查安装状态
func checkInstallStatus() bool {
	confPath := "./config/app.conf"
	_, err := os.Stat(confPath)
	if err != nil {
		return false
	}

	cfg, err := ini.Load(confPath)
	if err != nil {
		return false
	}

	dbtype := cfg.Section("").Key("dbtype").String()
	dbhost := cfg.Section("").Key("dbhost").String()
	dbname := cfg.Section("").Key("dbname").String()

	if dbtype == "" || dbhost == "" || dbname == "" {
		return false
	}

	return true
}

// runWeb 启动web
func runWeb(*cli.Context) error {
	// 日志初始化（使用默认配置，不依赖配置文件）
	fmt.Println("Web started!")
	logErr := initLoggerSafe()
	if logErr != nil {
		// 如果日志初始化失败，至少输出到标准输出
		fmt.Println("Warning: Logger initialization failed:", logErr)
		fmt.Println("Using stdout for logging")
		// 确保 Log 不为 nil
		if logger.Log == nil {
			logger.Log = logrus.New()
			logger.Log.SetOutput(os.Stdout)
			logger.Log.SetLevel(logrus.InfoLevel)
			logger.Log.SetFormatter(&logrus.TextFormatter{
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
			})
		}
	}

	// 释放模板文件、js文件到assets目录下
	if err := assets.RestoreAssets(); err != nil {
		logger.Log.Error("Failed to restore template files:", err)
		// 不退出程序，继续运行
	}
	// 检查安装状态
	installed := checkInstallStatus()
	if !installed {
		logger.Log.Info("系统未安装，启动安装引导模式")
		// 未安装时，只启动 Web 服务器，不连接数据库
		r := v1.InitRouter()
		httpport := "8085"
		logger.Log.Info("Starting Gin server in installation mode on port:", httpport)
		logger.Log.Info("Please visit http://localhost:" + httpport + "/install to complete installation")
		r.Run(":" + httpport)
		return nil
	}

	// 已安装，加载配置文件并连接数据库
	logger.Log.Info("系统已安装，加载配置并连接数据库")
	var err error
	webCfg, err = ini.Load("./config/app.conf")
	if err != nil {
		logger.Log.Error("Failed to load config file:", err)
		os.Exit(1)
	}
	//打印motd
	logger.Log.Info(motd)
	//配置文件已经建立，从配置文件读取数据库配置，并初始化数据库
	model.ModelInit(
		GetConfKey("dbtype"),
		GetConfKey("dbhost"),
		GetConfKey("dbuser"),
		GetConfKey("dbpass"),
		GetConfKey("dbname"),
		GetConfKey("dbport"))
	//计划任务
	model.InitTask()
	//企业微信
	model.InitWechat()
	//更新检查器
	model.InitUpdateChecker()
	defer model.StopTask()
	defer model.StopUpdateChecker()
	model.InitSenderWorker()
	go model.ConsumeMail()
	go model.ConsumeWechat()
	go model.ConsumeWechatRobot()

	// 直接使用 Gin 框架
	r := v1.InitRouter()
	httpport := GetConfKey("httpport")
	if httpport == "" {
		httpport = "8085"
	}
	logger.Log.Info("Starting Gin server on port:", httpport)
	r.Run(":" + httpport)
	return nil
}

// initLoggerSafe 安全地初始化日志（不依赖配置文件）
func initLoggerSafe() error {
	// 尝试加载配置文件
	cfg, err := ini.Load("./config/app.conf")
	if err != nil {
		// 配置文件不存在，使用默认日志配置：./log/yyyy-MM-dd.log
		err = logger.InitLoggerWithRunMode("dev", "", 7, 10000, 100, true)
		if err != nil {
			// 如果日志初始化失败，使用标准输出
			logger.Log = logrus.New()
			logger.Log.SetOutput(os.Stdout)
			logger.Log.SetLevel(logrus.InfoLevel)
			logger.Log.SetFormatter(&logrus.TextFormatter{
				FullTimestamp:   true,
				TimestampFormat: "2006-01-02 15:04:05",
			})
		}
		return nil
	}

	// 配置文件存在，使用配置的日志设置
	logPath := cfg.Section("").Key("log_path").String()
	if logPath == "" {
		// 如果配置文件中未指定日志路径，使用默认路径：./log/yyyy-MM-dd.log
		logPath = ""
	}

	// 获取 runmode 配置
	runmode := cfg.Section("").Key("runmode").String()
	if runmode == "" {
		runmode = "prod" // 默认为生产模式
	}

	maxday, _ := cfg.Section("").Key("maxdays").Int()
	if maxday == 0 {
		maxday = 7
	}
	maxlines, _ := cfg.Section("").Key("maxlines").Int()
	if maxlines == 0 {
		maxlines = 10000
	}
	maxsize, _ := cfg.Section("").Key("maxsize").Int()
	if maxsize == 0 {
		maxsize = 100
	}
	daily, _ := cfg.Section("").Key("daily").Bool()
	if !daily {
		daily = true // 默认启用按天分割
	}

	// 检查是否手动指定了 log_level，如果指定了则使用手动配置，否则根据 runmode 自动设置
	level, err := cfg.Section("").Key("log_level").Int()
	if err != nil || level == 0 {
		// 未手动指定 log_level，根据 runmode 自动设置
		err = logger.InitLoggerWithRunMode(runmode, logPath, maxday, maxlines, maxsize, daily)
	} else {
		// 手动指定了 log_level，使用手动配置
		err = logger.InitLogger(logPath, level, maxday, maxlines, maxsize, daily)
	}

	if err != nil {
		// 如果日志初始化失败，使用标准输出
		logger.Log = logrus.New()
		logger.Log.SetOutput(os.Stdout)
		logger.Log.SetLevel(logrus.InfoLevel)
		logger.Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
		return nil
	}

	// 记录当前运行模式和日志级别
	logger.Log.Infof("System running in %s mode", runmode)

	return nil
}

// CheckDb
func CheckDb(dbdriver, dbhost, dbuser, dbpass, dbname string, dbport string) error {
	//database type
	switch dbdriver {
	case "mysql":
		dbURL := dbuser + ":" + dbpass + "@tcp(" + dbhost + ":" +
			dbport + ")/" + dbname + "?parseTime=true&loc=Asia%2FShanghai&timeout=5s&charset=utf8&collation=utf8_general_ci"
		db, err := sql.Open("mysql", dbURL)
		if err != nil {
			return err
		}
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
		err = db.Ping()
		if err != nil {
			return err
		}
	}
	return nil
}

// init config files
func GetConfKey(v string) string {
	// 为了与 model.GetConfKey 行为一致，这里也实现：
	// 1）如果 .env 存在，则所有配置完全由 .env 决定（可以是空字符串），不会再回退到 app.conf
	// 2）只有当 .env 不存在时，才从 app.conf 读取
	initEnvOnce.Do(func() {
		file, err := os.Open(".env")
		if err != nil {
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
			val = strings.Trim(val, `"'`)
			if key != "" {
				envConfig[key] = val
			}
		}
	})

	if envFileExists {
		if envConfig == nil {
			return ""
		}
		if val, ok := envConfig[v]; ok {
			return val
		}
		return ""
	}

	// 回退到 app.conf
	if webCfg == nil {
		return ""
	}
	p, err := webCfg.Section("").GetKey(v)
	if err != nil {
		logger.Log.Error(err)
		return ""
	}
	return p.String()
}

// InitLogger 初始化日志（已废弃，改为使用 initLoggerSafe）
func InitLogger() (err error) {
	return initLoggerSafe()
}
