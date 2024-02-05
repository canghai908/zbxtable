package models

import (
	"context"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	zabbix "github.com/canghai908/zabbix-go"
	redis "github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
	workwx "github.com/xen0n/go-workwx"
	ini "gopkg.in/ini.v1"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"zbxtable/utils"

	jsoniter "github.com/json-iterator/go"
	_ "github.com/lib/pq"
	"os"
)

var (
	API  = &zabbix.API{}
	json = jsoniter.Config{
		EscapeHTML:             false,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
	}.Froze()
	RDB        = &redis.Client{}
	ZBX_VER    string
	ZBX_V      bool
	Version    string
	GitHash    string
	BuildTime  string
	AssetsHost string
	WeApp      = &workwx.WorkwxApp{}
)

// TableName 表名前缀
func TableName(str string) string {
	return fmt.Sprintf("%s%s", "zbxtable_", str)
}

// GetAssetsHost
func GetAssetsHost() string {
	AssetsHost = beego.AppConfig.String("AssetsHost")
	if AssetsHost == "" {
		AssetsHost = "http://dl.cactifans.com/assets/"
	}
	return AssetsHost
}

// ModelsInit  p
func ModelsInit(zabbix_web, zabbix_user, zabbix_pass, zabbix_token,
	dbtype, dbhost, dbuser, dbpass, dbname, dbport,
	redis_host, redis_port, redis_pass, redis_db string) {
	//GetAssetsHost
	GetAssetsHost()
	//database chechek
	switch dbtype {
	case "mysql":
		dbURL := dbuser + ":" + dbpass + "@tcp(" + dbhost + ":" +
			dbport + ")/" + dbname + "?parseTime=true&loc=Asia%2FShanghai&timeout=5s&charset=utf8&collation=utf8_general_ci"
		err := orm.RegisterDataBase("default", "mysql", dbURL)
		if err != nil {
			logs.Error(err)
			os.Exit(1)
		}
	case "postgresql":
		dbURl := "postgres://" + dbuser + ":" + url.QueryEscape(dbpass) + "@" + dbhost + ":" + dbport + "/" + dbname + "?sslmode=disable"
		connString, err := url.Parse(dbURl)
		if err != nil {
			logs.Error(err)
			os.Exit(1)
		}
		err = orm.RegisterDataBase("default", "postgres", connString.String())
		if err != nil {
			logs.Error(err)
			os.Exit(1)
		}
	default:
		dbURL := dbuser + ":" + dbpass + "@tcp(" + dbhost + ":" +
			dbport + ")/" + dbname + "?parseTime=true&loc=Asia%2FShanghai&timeout=5s&charset=utf8&collation=utf8_general_ci"
		err := orm.RegisterDataBase("default", "mysql", dbURL)
		if err != nil {
			logs.Error(err)
			os.Exit(1)
		}
	}
	//创建表
	orm.RegisterModel(
		new(Alarm), new(Manager), new(Topology),
		new(System), new(Report), new(Egress),
		new(TaskLog), new(Rule), new(UserGroup),
		new(EventLog))
	err := orm.RunSyncdb("default", false, true)
	if err != nil {
		logs.Error(err)
		os.Exit(1)
	}
	if GetConfKey("runmode") == "dev" {
		orm.Debug = true
	}
	// init admin
	DatabaseInit()
	//判断API地址是否正确，http get访问访问api地址判断状态码是不是412
	addURL := zabbix_web + "/api_jsonrpc.php"
	dClient := http.Client{
		Timeout: 3 * time.Second, // 设置超时时间为 3 秒
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
	//api变量
	API = zabbix.NewAPI(zabbix_web + "/api_jsonrpc.php")
	if zabbix_token != "" {
		API.Auth = zabbix_token
	} else {
		_, err = API.Login(zabbix_user, zabbix_pass)
		if err != nil {
			logs.Error(err)
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
		logs.Error("connect Zabbix API failed:", err)
		os.Exit(1)
	}
	//Zabbix version
	version, err := API.Version()
	if err != nil {
		logs.Error(err)
		os.Exit(1)
	}
	ZBX_VER = version
	verArr := strings.Split(ZBX_VER, ".")
	ZbxMasterVer, _ := strconv.ParseInt(verArr[0], 10, 64)
	ZbxMiddleVer, _ := strconv.ParseInt(verArr[1], 10, 64)
	if ZbxMasterVer >= 6 || (ZbxMasterVer == 5 && ZbxMiddleVer == 4) {
		ZBX_V = true
	} else {
		ZBX_V = false
	}
	logs.Info("Zabbix API connected！Zabbix version:", version)

	//zabbix web login
	//	LoginZabbixWeb(zabbix_web, zabbix_user, zabbix_pass)

	//redis
	res_db, err := strconv.Atoi(redis_db)
	if err != nil {
		logs.Error(err)
		os.Exit(1)
	}
	RDB = redis.NewClient(&redis.Options{
		Addr:     redis_host + ":" + redis_port,
		Password: redis_pass, // no password set
		DB:       res_db,     // use default DB
	})
	var ctx = context.Background()
	_, err = RDB.Ping(ctx).Result()
	if err != nil {
		logs.Error(err)
		os.Exit(1)
	}
	logs.Info("Redis connected!")
	//gen tpl

	//
	AgentId, err := beego.AppConfig.Int64("wechat_agentid")
	if err != nil {
		logs.Error(err)
		os.Exit(1)
	}
	client := workwx.New(beego.AppConfig.String("wechat_corpid"))
	WeApp = client.WithApp(beego.AppConfig.String("wechat_secret"), AgentId)
	WeApp.SpawnAccessTokenRefresher()
	logs.Info("WeChat inited!")
}

// DatabaseInit 数据初始化
func DatabaseInit() {
	//数据初始化操作
	o := orm.NewOrm()
	v := &Manager{Username: "admin"}
	err := o.Read(v, "username")
	//检查权限
	if err == nil && v.Operation == "" {
		var manager Manager
		manager.ID = v.ID
		manager.Operation = "['add', 'edit', 'delete','update']"
		_, err := o.Update(&manager, "Operation")
		if err != nil {
			logs.Info(err)
			return
		}
		logs.Info("update admin operation successfully")
	}
	//添加管理员账号
	if err == orm.ErrNoRows {
		logs.Info("the admin user does not exist, create a new admin account later!")
		var manager Manager
		manager.Username = "admin"
		manager.Password, _ = utils.PasswordHash("Zbxtable")
		manager.Avatar = "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif"
		manager.Role = "admin"
		manager.Operation = "['add', 'edit', 'delete','update']"
		manager.Status = 0
		id, err := o.Insert(&manager)
		if err != nil {
			logs.Info(err)
			return
		}
		logs.Info("create an administrator account successfully, the admin ID is:", id)
	}
	//初始化系统数据
	var cnt []System
	al := new(System)
	_, err = o.QueryTable(al).All(&cnt)
	if err != nil {
		logs.Info(err)
		return
	}
	if len(cnt) == 0 {
		sys := []System{
			{Name: "Linux操作系统", Status: 0},
			{Name: "Windows操作系统", Status: 0},
			{Name: "网络设备", Status: 0},
			{Name: "物理服务器", Status: 0},
		}
		_, err := o.InsertMulti(len(sys), sys)
		if err != nil {
			logs.Info("Init system info error！")
			return
		}
		logs.Info("Init system data successfully!")
	}
	//出口
	//初始化系统数据
	var cne []Egress
	all := new(Egress)
	_, err = o.QueryTable(all).All(&cne)
	if err != nil {
		logs.Info(err)
		return
	}
	if len(cne) == 0 {
		egre := []Egress{
			{NameOne: "电信100M", NameTwo: "移动100M", Status: 0},
		}
		_, err := o.InsertMulti(len(egre), egre)
		if err != nil {
			logs.Info("Init egress info error！")
			return
		}
		logs.Info("Init egress data successfully!")
	}
	//init default rule
	var cRule []Rule
	allRule := new(Rule)
	_, err = o.QueryTable(allRule).Filter("MType", "2").All(&cRule)
	if err != nil {
		logs.Info(err)
		return
	}
	if len(cRule) == 0 {
		defaultRule := []Rule{
			{Name: "默认规则", TenantID: "zabbix01", MType: "2",
				Channel: "wechat", UserIds: "1", Status: "0"},
		}
		_, err := o.InsertMulti(len(defaultRule), defaultRule)
		if err != nil {
			logs.Info("Init default rule error！")
			return
		}
		logs.Info("Init default rule successfully!")
	}
}

func GetConfKey(v string) string {
	cfg, err := ini.Load("./conf/app.conf")
	if err != nil {
		logs.Error(err)
		logs.Error("Please run 'zbxtable init' to create app.conf")
		return ""
	}
	p, err := cfg.Section("").GetKey(v)
	if err != nil {
		logs.Error(err)
		return ""
	}
	return p.String()
}
