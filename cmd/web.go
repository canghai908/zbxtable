package cmd

import (
	"database/sql"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/canghai908/zabbix-go"
	"github.com/canghai908/zbxtable/models"
	"github.com/canghai908/zbxtable/routers"
	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
	"os"
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
	err := logs.SetLogger(logs.AdapterFile, `{"filename":"logs/zbxtable.log","level":7,"maxlines":0,
		"maxsize":0,"daily":true,"maxdays":10,
		"color":true,"perm":"0755"}`)
	logs.Info(motd)
	if err != nil {
		logs.Info(err)
		return err
	}
	//CheckConfExist
	CheckConfExist()
	////PreCheckConf
	//err = PreCheckConf(InitConfig("zabbix_web"), InitConfig("zabbix_user"), InitConfig("zabbix_pass"),
	//	InitConfig("dbtype"), InitConfig("dbhost"), InitConfig("dbuser"),
	//	InitConfig("dbpass"), InitConfig("dbname"), InitConfig("dbport"))
	//if err != nil {
	//	logs.Error(err)
	//	return err
	//}
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	models.ModelsInit(InitConfig("zabbix_web"), InitConfig("zabbix_user"), InitConfig("zabbix_pass"),
		InitConfig("dbtype"), InitConfig("dbhost"), InitConfig("dbuser"),
		InitConfig("dbpass"), InitConfig("dbname"), InitConfig("dbport"))
	routers.RouterInit()
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
func CheckZabbixAPI(address, user, pass string) (string, error) {
	API = zabbix.NewAPI(address + "/api_jsonrpc.php")
	_, err := API.Login(user, pass)
	if err != nil {
		return "", err
	}
	version, err := API.Version()
	if err != nil {
		return "", err
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

//
