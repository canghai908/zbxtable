package cmd

import (
	"github.com/astaxie/beego/logs"
	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
)

var (
	//Web 配置
	Uc = &cli.Command{
		Name:   "uc",
		Usage:  "Update config file",
		Action: updateconfig,
	}
)

//runWeb 启动web
func updateconfig(*cli.Context) error {
	logs.Info("Start upgrading the old configuration file!")
	cfg, err := ini.Load("./conf/app.conf")
	if err != nil {
		logs.Error(err)
		return err
	}
	conf := make(map[string]string)
	var b = []string{"appname", "httpport", "runmode", "copyrequestbody", "session_timeout",
		"dbdriver", "hostname", "username", "dbpsword", "database", "port",
		"zabbix_web", "zabbix_user", "zabbix_pass", "token"}
	for _, v := range b {
		p, err := cfg.Section("").GetKey(v)
		if err != nil {
			logs.Error(err)
			return err
		}
		conf[v] = p.String()
	}
	err = CheckConf(conf["zabbix_web"], conf["zabbix_user"], conf["zabbix_pass"],
		conf["dbdriver"], conf["hostname"], conf["username"], conf["dbpsword"], conf["database"], conf["port"])
	if err != nil {
		logs.Error(err)
		return err
	}
	err = WriteConf(conf["zabbix_web"], conf["zabbix_user"], conf["zabbix_pass"],
		conf["dbdriver"], conf["hostname"], conf["username"], conf["dbpsword"], conf["database"], conf["port"],
		conf["httpport"], conf["runmode"], conf["session_timeout"], conf["token"])
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Info("Successfully upgraded the old configuration file!")
	return nil
}

//check conf
func CheckConf(zabbix_web, zabbix_user, zabbix_pass,
	dbtype, dbhost, dbuser, dbpass, dbname, dbport string) error {
	err := CheckDb(dbtype, dbhost, dbuser, dbpass, dbname, dbport)
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Info("Connected to database " + dbname + " successfully!")
	version, err := CheckZabbix(zabbix_web, zabbix_user, zabbix_pass)
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Info("Connected to zabbix web successfully！Zabbix version is :", version)
	return nil
}

//check conf
func PreCheckConf() error {
	cfg, err := ini.Load("./conf/app.conf")
	if err != nil {
		logs.Error(err)
		return err
	}
	conf := make(map[string]string)
	var b = []string{"dbtype", "dbhost", "dbuser", "dbpass", "dbname", "dbport",
		"zabbix_web", "zabbix_user", "zabbix_pass"}
	for _, v := range b {
		p, err := cfg.Section("").GetKey(v)
		if err != nil {
			logs.Error(err)
			return err
		}
		conf[v] = p.String()
	}
	err = CheckDb(conf["dbtype"], conf["dbhost"], conf["dbuser"], conf["dbpass"], conf["dbname"], conf["dbport"])
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Info("Connected to database " + conf["dbname"] + " successfully!")
	version, err := CheckZabbix(conf["zabbix_web"], conf["zabbix_user"], conf["zabbix_pass"])
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Info("Connected to zabbix web successfully！Zabbix version is :", version)
	return nil
}
