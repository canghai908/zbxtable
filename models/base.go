package models

import (
	fmt "fmt"
	"github.com/astaxie/beego/logs"
	"os"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	zabbix "github.com/canghai908/zabbix-go"
	"github.com/canghai908/zbxtable/utils"
	_ "github.com/go-sql-driver/mysql"
	jsoniter "github.com/json-iterator/go"
	_ "github.com/lib/pq"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

//TableName func
func TableName(str string) string {
	return fmt.Sprintf("%s%s", "zbxtable_", str)
}

//DatabaseInit 数据初始化
func DatabaseInit() {
	//数据初始化操作
	//添加管理员账号
	o := orm.NewOrm()
	v := &Manager{Username: "admin"}
	err := o.Read(v, "Username")
	if err == orm.ErrNoRows {
		logs.Info("the admin user does not exist, create a new admin account later!")
		var manager Manager
		manager.Username = "admin"
		manager.Password = utils.Md5([]byte("Zbxtable"))
		manager.Avatar = "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif"
		manager.Role = "admin"
		manager.Status = 0
		manager.Created = time.Now()
		id, err := o.Insert(&manager)
		if err != nil {
			logs.Error(err)
		}
		logs.Info("create an administrator account successfully, the admin ID is:", id)
	}
}

//API for zabbix
var API = &zabbix.API{}

//ModelsInit  p
func ModelsInit() {
	//database init
	dbtype := beego.AppConfig.String("dbtype")
	dbhost := beego.AppConfig.String("dbhost")
	dbuser := beego.AppConfig.String("dbuser")
	dbpass := beego.AppConfig.String("dbpass")
	dbname := beego.AppConfig.String("dbname")
	dbport := beego.AppConfig.String("dbport")
	//database type
	switch dbtype {
	case "mysql":
		dburl := dbuser + ":" + dbpass + "@tcp(" + dbhost + ":" +
			dbport + ")/" + dbname + "?parseTime=true&loc=Asia%2FShanghai&timeout=5s&charset=utf8&collation=utf8_general_ci"
		err := orm.RegisterDataBase("default", "mysql", dburl)
		if err != nil {
			logs.Error(err)
			os.Exit(1)
		}
	case "postgresql":
		dburl := "postgres://" + dbuser + ":" + dbpass + "@" + dbhost + ":" + dbport + "/" + dbname + "?sslmode=disable"
		err := orm.RegisterDataBase("default", "postgres", dburl)
		if err != nil {
			logs.Error(err)
			os.Exit(1)
		}
	default:
		dburl := dbuser + ":" + dbpass + "@tcp(" + dbhost + ":" +
			dbport + ")/" + dbname + "?parseTime=true&loc=Asia%2FShanghai&timeout=5s&charset=utf8&collation=utf8_general_ci"
		err := orm.RegisterDataBase("default", "mysql", dburl)
		if err != nil {
			logs.Error(err)
			os.Exit(1)
		}
	}
	orm.RegisterModel(new(Alarm), new(Manager))
	err := orm.RunSyncdb("default", false, true)
	if err != nil {
		logs.Error(err)
		os.Exit(1)
	}
	if beego.AppConfig.String("runmode") == "dev" {
		orm.Debug = true
	}
	// init admin
	DatabaseInit()
	//zabbix api login
	ZabbixWeb := beego.AppConfig.String("zabbix_web")
	ZabbixUser := beego.AppConfig.String("zabbix_user")
	ZabbixPass := beego.AppConfig.String("zabbix_pass")
	API = zabbix.NewAPI(ZabbixWeb + "/api_jsonrpc.php")
	_, err = API.Login(ZabbixUser, ZabbixPass)
	if err != nil {
		logs.Error(err)
		os.Exit(1)
	}
	//web login
	Intt()
}
