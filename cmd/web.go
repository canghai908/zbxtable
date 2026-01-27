package cmd

import (
	"bufio"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
	"zbxtable/models"
	"zbxtable/packfile"
	"zbxtable/routers"
	"zbxtable/utils"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/toolbox"
	"github.com/canghai908/zabbix-go"
	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
)

const motd = `

$$$$$$$$\ $$$$$$$\  $$\   $$\ $$$$$$$$\  $$$$$$\  $$$$$$$\  $$\       $$$$$$$$\ 
\____$$  |$$  __$$\ $$ |  $$ |\__$$  __|$$  __$$\ $$  __$$\ $$ |      $$  _____|
    $$  / $$ |  $$ |\$$\ $$  |   $$ |   $$ /  $$ |$$ |  $$ |$$ |      $$ |      
   $$  /  $$$$$$$\ | \$$$$  /    $$ |   $$$$$$$$ |$$$$$$$\ |$$ |      $$$$$\    
  $$  /   $$  __$$\  $$  $$<     $$ |   $$  __$$ |$$  __$$\ $$ |      $$  __|   
 $$  /    $$ |  $$ |$$  /\$$\    $$ |   $$ |  $$ |$$ |  $$ |$$ |      $$ |      
$$$$$$$$\ $$$$$$$  |$$ /  $$ |   $$ |   $$ |  $$ |$$$$$$$  |$$$$$$$$\ $$$$$$$$\ 
\________|\_______/ \__|  \__|   \__|   \__|  \__|\_______/ \________|\________|
`

var (
	//Web 配置
	Web = &cli.Command{
		Name:   "web",
		Usage:  "Start web server",
		Action: runWeb,
	}

	envConfig   map[string]string
	initEnvOnce sync.Once
	envFileExists bool
)

// runWeb 启动web
func runWeb(*cli.Context) error {
	// 日志初始化
	InitLogger()
	logs.Info(motd)
	// 释放静态资源目录
	restoreAssets()
	//检查配置文件是否存在
	CheckConfExist()
	//dev模式下开启swagger
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	models.ModelsInit(InitConfig("zabbix_web"), InitConfig("zabbix_user"), InitConfig("zabbix_pass"),
		InitConfig("zabbix_token"),
		InitConfig("dbtype"), InitConfig("dbhost"), InitConfig("dbuser"),
		InitConfig("dbpass"), InitConfig("dbname"), InitConfig("dbport"),
		InitConfig("redis_host"), InitConfig("redis_port"),
		InitConfig("redis_pass"), InitConfig("redis_db"))
	routers.RouterInit()
	models.InitTask()
	toolbox.StartTask()
	defer toolbox.StopTask()
	models.InitSenderWorker()
	go models.ConsumeMail()
	go models.ConsumeWechat()
	go models.ConsumeWechatRobot()
	beego.Run()
	return nil
}

// checkWeb 是否需要释放web资源目录
func restoreAssets() error {
	//判断静态资源目录是否存在，不存在则释放
	files := []string{"web", "template", "conf"} // 设置需要释放的目录
	for _, file := range files {
		// 判断目录是否存在
		ex, err := utils.PathExists("./" + file)
		//目录不存在释放静态资源
		if !ex && err == nil {
			// 解压目录到当前目录
			if err := packfile.RestoreAssets("./", file); err != nil {
				logs.Error(err)
				break
			}
		}
	}
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

// LoginZabbixAPI Check
func CheckZabbixAPI(args ...string) (string, error) {
	address := args[0]
	user := args[1]
	pass := args[2]
	token := args[3]

	//TLS SkipVerify
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	//判断API地址是否正确，http get访问访问api地址判断状态码是不是412
	addURL := address + "/api_jsonrpc.php"
	dClient := http.Client{
		Transport: transport,
		Timeout:   3 * time.Second, // 设置超时时间为 3 秒
	}
	resp, err := dClient.Get(addURL)
	if err != nil {
		logs.Error("Zabbix Web get request failed:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusPreconditionFailed {
		logs.Error("Zabbix Web is incorrectly!")
		os.Exit(1)
	}
	// api定义
	API = zabbix.NewAPI(address + "/api_jsonrpc.php")
	if token != "" {
		API.SetAuth(token)
	} else {
		_, err := API.Login(user, pass)
		if err != nil {
			logs.Error("connect Zabbix API failed", err)
			os.Exit(1)
		}
	}
	//zabbix api data get test
	OutputPar := []string{"hostid", "host", "available", "status", "name", "error"}
	type params map[string]interface{}
	_, err = API.CallWithError("host.get", params{
		"output":  OutputPar,
		"hostids": "10084",
	})
	if err != nil {
		logs.Error("connect Zabbix API failed", err)
		os.Exit(1)
	}
	//version get
	version, err := API.Version()
	if err != nil {
		logs.Error("connect Zabbix API failed", err)
		os.Exit(1)
	}
	return version, nil
}

// CheckConfExist config
func CheckConfExist() {
	_, err := os.Stat("./conf/app.conf")
	if err != nil {
		logs.Error(err)
		logs.Error("Please run 'zbxtable init' to create app.conf")
		os.Exit(1)
	}
	Cfg, err = ini.Load("./conf/app.conf")
	if err != nil {
		logs.Error(err)
		logs.Error("Please run 'zbxtable init' to create app.conf")
		os.Exit(1)
	}
}

// init config files
func InitConfig(v string) string {
	// 为了与 models.GetConfKey 行为一致，这里也实现：
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
	p, err := Cfg.Section("").GetKey(v)
	if err != nil {
		logs.Error(err)
		return ""
	}
	return p.String()
}
func InitLogger() (err error) {
	BConfig, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		return errors.New("config init error:" + err.Error())
	}
	logConf := make(map[string]interface{})
	logConf["filename"] = BConfig.String("log_path")
	level, _ := BConfig.Int("log_level")
	maxday, _ := BConfig.Int("maxdays")
	maxlines, _ := BConfig.Int("maxlines")
	maxsize, _ := BConfig.Int("maxsize")
	daily, _ := BConfig.Bool("daily")
	logConf["level"] = level
	logConf["maxlines"] = maxlines
	logConf["maxsize"] = maxsize
	logConf["maxday"] = maxday
	logConf["daily"] = daily
	logConf["perm"] = "0755"
	confStr, err := json.Marshal(logConf)
	if err != nil {
		return errors.New("marshal failed,err:" + err.Error())
	}
	logs.SetLogger(logs.AdapterFile, string(confStr))
	logs.SetLogFuncCall(true)
	return
}

//
