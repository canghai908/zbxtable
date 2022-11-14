package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/toolbox"
	"github.com/canghai908/zabbix-go"
	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
	"os"
	"zbxtable/models"
	"zbxtable/routers"
)

const motd = `
$$$$$$$$\ $$$$$$$\  $$\   $$\ $$$$$$$$\  $$$$$$\  $$$$$$$\  $$\       $$$$$$$$\ 
\____$$  |$$  __$$\ $$ |  $$ |\__$$  __|$$  __$$\ $$  __$$\ $$ |      $$  _____|
    $$  / $$ |  $$ |\$$\ $$  |   $$ |   $$ /  $$ |$$ |  $$ |$$ |      $$ |      
   $$  /  $$$$$$$\ | \$$$$  /    $$ |   $$$$$$$$ |$$$$$$$\ |$$ |      $$$$$\    
  $$  /   $$  __$$\  $$  $$<     $$ |   $$  __$$ |$$  __$$\ $$ |      $$  __|   
 $$  /    $$ |  $$ |$$  /\$$\    $$ |   $$ |  $$ |$$ |  $$ |$$ |      $$ |      
$$$$$$$$\ $$$$$$$  |$$ /  $$ |   $$ |   $$ |  $$ |$$$$$$$  |$$$$$$$$\ $$$$$$$$\ 
\________|\_______/ \__|  \__|   \__|   \__|  \__|\_______/ \________|\________|`

var (
	//Web 配置
	Web = &cli.Command{
		Name:   "web",
		Usage:  "Start web server",
		Action: runWeb,
	}
)

//runWeb 启动web
func runWeb(*cli.Context) error {
	logs.Info(motd)
	CheckConfExist()
	Initlogger()
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
	beego.Run()
	return nil
}

//CheckDb
func CheckDb(dbdriver, dbhost, dbuser, dbpass, dbname string, dbport string) error {
	//database type
	switch dbdriver {
	case "mysql":
		dburl := dbuser + ":" + dbpass + "@tcp(" + dbhost + ":" +
			dbport + ")/" + dbname + "?parseTime=true&loc=Asia%2FShanghai&timeout=5s&charset=utf8&collation=utf8_general_ci"
		db, err := sql.Open("mysql", dburl)
		if err != nil {
			return err
		}
		err = db.Ping()
		if err != nil {
			return err
		}
	case "postgresql":
		dburl := "postgres://" + dbuser + ":" + dbpass + "@" + dbhost + ":" + dbport + "/" + dbname + "?sslmode=disable"
		db, err := sql.Open("postgres", dburl)
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

//LoginZabbixAPI Check
func CheckZabbixAPI(args ...string) (string, error) {
	address := args[0]
	user := args[1]
	pass := args[2]
	token := args[3]
	API = zabbix.NewAPI(address + "/api_jsonrpc.php")
	address = args[0]
	if token != "" {
		API.Auth = token
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
	_, err := API.CallWithError("host.get", params{
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

//CheckConfExist config
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

//init config files
func InitConfig(v string) string {
	p, err := Cfg.Section("").GetKey(v)
	if err != nil {
		logs.Error(err)
		return ""
		os.Exit(1)
	}
	return p.String()
}
func Initlogger() (err error) {
	BConfig, err := config.NewConfig("ini", "conf/app.conf")
	if err != nil {
		fmt.Println("config init error:", err)
		return
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
		fmt.Println("marshal failed,err:", err)
		return
	}
	logs.SetLogger(logs.AdapterFile, string(confStr))
	logs.SetLogFuncCall(true)
	return
}

//
