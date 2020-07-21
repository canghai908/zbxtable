package models

import (
	"fmt"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	zabbix "github.com/canghai908/zabbix-go"
	"github.com/canghai908/zbxtable/utils"
	_ "github.com/go-sql-driver/mysql"
)

//TableName func
func TableName(str string) string {
	return fmt.Sprintf("%s%s", beego.AppConfig.String("dbprefix"), str)
}

//API for zabbix
var API = &zabbix.API{}

//ModelsInit init
func ModelsInit() {
	//database init
	dbhost := beego.AppConfig.String("hostname")
	dbuser := beego.AppConfig.String("username")
	dbpass := beego.AppConfig.String("dbpsword")
	dbname := beego.AppConfig.String("database")
	dbport := beego.AppConfig.String("port")
	if dbport == "" {
		dbport = "3306"
	}
	dburl := dbuser + ":" + dbpass + "@tcp(" + dbhost + ":" +
		dbport + ")/" + dbname + "?parseTime=true&loc=Asia%2FShanghai&timeout=5s&charset=utf8&collation=utf8_general_ci"
	orm.RegisterDataBase("default", "mysql", dburl)
	orm.RegisterModel(new(Alarm), new(Manager))
	orm.RunSyncdb("default", false, true)
	if beego.AppConfig.String("runmode") == "dev" {
		orm.Debug = true
	}
	DatabaseInit()
	//zabbix init
	ZabbixServer := beego.AppConfig.String("zabbix_server")
	ZabbixUser := beego.AppConfig.String("zabbix_user")
	ZabbixPass := beego.AppConfig.String("zabbix_pass")
	API = zabbix.NewAPI(ZabbixServer + "/api_jsonrpc.php")
	API.Login(ZabbixUser, ZabbixPass)
	Intt()
}

//DatabaseInit 数据初始化
func DatabaseInit() {
	//数据初始化操作
	//添加管理员账号
	o := orm.NewOrm()
	v := &Manager{Username: "admin"}
	err = o.Read(v, "Username")
	if err == orm.ErrNoRows {
		beego.Info("the admin user does not exist, create a new admin account later!")
		var manager Manager
		manager.Username = "admin"
		manager.Password = utils.Md5([]byte("Zbxtable"))
		manager.Avatar = "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif"
		manager.Role = "admin"
		manager.Status = 0
		manager.Created = time.Now()
		id, err := o.Insert(&manager)
		if err != nil {
			beego.Error(err)
		}
		beego.Info("create an administrator account successfully, the admin ID is:", id)
	}
}
