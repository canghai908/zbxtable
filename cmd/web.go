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

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
)

var (
	//Web config
	Web = &cli.Command{
		Name:   "web",
		Usage:  "Start web server",
		Action: runWeb,
	}

	envConfig     map[string]string
	initEnvOnce   sync.Once
	envFileExists bool
	webCfg        *ini.File
)

// checkInstallStatus check installation status
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

// runWeb start web server
func runWeb(*cli.Context) error {
	fmt.Println("🚀 Starting ZbxTable service...")

	// Initialize logger (using default config, not dependent on config file)
	logErr := initLoggerSafe()
	if logErr != nil {
		// If logger initialization fails, at least output to stdout
		fmt.Println("Warning: Logger initialization failed:", logErr)
		fmt.Println("Using standard output for logging")
		// Ensure Log is not nil
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

	// Extract template files and js files to assets directory
	logger.Log.Info("Extracting static resource files...")
	if err := assets.RestoreAssets(); err != nil {
		logger.Log.Error("Failed to extract template files:", err)
		// Don't exit, continue running
	} else {
		logger.Log.Info("Static resource files extracted successfully")
	}

	// Check installation status
	installed := checkInstallStatus()
	if !installed {
		logger.Log.Info("System not installed, starting installation wizard mode")
		// When not installed, only start web server, don't connect to database
		r := v1.InitRouter()
		httpport := "8088"
		fmt.Println("╔══════════════════════════════════════════════════════╗")
		fmt.Printf("║  Service listening on: http://0.0.0.0:%-11s    ║\n", httpport)
		fmt.Printf("║  Local access URL: http://localhost:%-13s    ║\n", httpport)
		fmt.Printf("║  Installation wizard: http://localhost:%s/install  ║\n", httpport)
		fmt.Println("║  Running mode: Installation wizard mode              ║")
		fmt.Println("╚══════════════════════════════════════════════════════╝")
		fmt.Println("ZbxTable service started successfully, waiting for installation configuration...")
		r.Run(":" + httpport)
		return nil
	}

	// Already installed, load config file and connect to database
	logger.Log.Info("System already installed, loading configuration...")

	var err error
	webCfg, err = ini.Load("./config/app.conf")
	if err != nil {
		logger.Log.Error("Failed to load config file:", err)
		os.Exit(1)
	}
	logger.Log.Info("Config file loaded successfully")

	// Get configuration information
	dbtype := GetConfKey("dbtype")
	dbhost := GetConfKey("dbhost")
	dbport := GetConfKey("dbport")
	dbname := GetConfKey("dbname")
	httpport := GetConfKey("httpport")
	if httpport == "" {
		httpport = "8088"
	}
	runmode := GetConfKey("runmode")
	if runmode == "" {
		runmode = "prod"
	}

	// Print configuration information
	logger.Log.Info("System configuration:")
	logger.Log.Infof("   - Running mode: %s", runmode)
	logger.Log.Infof("   - Database type: %s", dbtype)
	logger.Log.Infof("   - Database address: %s:%s", dbhost, dbport)
	logger.Log.Infof("   - Database name: %s", dbname)
	logger.Log.Infof("   - HTTP port: %s", httpport)

	//Config file already created, read database config from config file and initialize database
	logger.Log.Info("Connecting to database...")
	model.ModelInit(
		dbtype,
		dbhost,
		GetConfKey("dbuser"),
		GetConfKey("dbpass"),
		dbname,
		dbport)
	logger.Log.Info("Database connected successfully")

	//Scheduled tasks
	logger.Log.Info("Initializing scheduled tasks...")
	model.InitTask()
	logger.Log.Info("Scheduled tasks initialized successfully")

	//WeChat Work
	logger.Log.Info("Initializing WeChat...")
	model.InitWechat()
	logger.Log.Info("WeChat initialized successfully")

	//Update checker
	logger.Log.Info("Initializing update checker...")
	model.InitUpdateChecker()
	logger.Log.Info("Update checker initialized successfully")

	defer model.StopTask()
	defer model.StopUpdateChecker()

	logger.Log.Info("Initializing message sending service...")
	model.InitSenderWorker()
	go model.ConsumeMail()
	go model.ConsumeWechat()
	go model.ConsumeWechatRobot()
	logger.Log.Info("Message sending service started successfully")

	// Use Gin framework directly
	r := v1.InitRouter()

	fmt.Println("╔═════════════════════════════════════════════╗")
	fmt.Printf("║  Service listening on: http://0.0.0.0:%-1s  ║\n", httpport)
	fmt.Printf("║  Local access URL: http://localhost:%-1s    ║\n", httpport)
	fmt.Printf("║  Running mode: %-25s    ║\n", runmode)
	fmt.Printf("║  Database: %s@%s:%-16s║\n", dbtype, dbhost, dbport)
	fmt.Println("╚═════════════════════════════════════════════╝")
	fmt.Println("ZbxTable service started successfully!")

	r.Run(":" + httpport)
	return nil
}

// initLoggerSafe safely initialize logger (not dependent on config file)
func initLoggerSafe() error {
	// Try to load config file
	cfg, err := ini.Load("./config/app.conf")
	if err != nil {
		// Config file doesn't exist, use default log config: ./log/yyyy-MM-dd.log
		err = logger.InitLoggerWithRunMode("dev", "", 7, 10000, 100, true)
		if err != nil {
			// If logger initialization fails, use standard output
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

	// Config file exists, use configured log settings
	logPath := cfg.Section("").Key("log_path").String()
	if logPath == "" {
		// If log path not specified in config file, use default path: ./log/yyyy-MM-dd.log
		logPath = ""
	}

	// Get runmode config
	runmode := cfg.Section("").Key("runmode").String()
	if runmode == "" {
		runmode = "prod" // Default to production mode
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
		daily = true // Enable daily rotation by default
	}

	// Check if log_level is manually specified, if so use manual config, otherwise auto-set based on runmode
	level, err := cfg.Section("").Key("log_level").Int()
	if err != nil || level == 0 {
		// log_level not manually specified, auto-set based on runmode
		err = logger.InitLoggerWithRunMode(runmode, logPath, maxday, maxlines, maxsize, daily)
	} else {
		// log_level manually specified, use manual config
		err = logger.InitLogger(logPath, level, maxday, maxlines, maxsize, daily)
	}

	if err != nil {
		// If logger initialization fails, use standard output
		logger.Log = logrus.New()
		logger.Log.SetOutput(os.Stdout)
		logger.Log.SetLevel(logrus.InfoLevel)
		logger.Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
		return nil
	}

	// Log current running mode and log level
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

// GetConfKey init config files
func GetConfKey(v string) string {
	// To be consistent with model.GetConfKey behavior, implement here:
	// 1) If .env exists, all configuration is completely determined by .env (can be empty string), will not fall back to app.conf
	// 2) Only when .env doesn't exist, read from app.conf
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

	// Fall back to app.conf
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

// InitLogger initialize logger (deprecated, use initLoggerSafe instead)
func InitLogger() (err error) {
	return initLoggerSafe()
}
