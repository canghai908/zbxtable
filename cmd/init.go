package cmd

import (
	"errors"
	"fmt"
	zabbix "github.com/canghai908/zabbix-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/manifoldco/promptui"
	"github.com/urfave/cli/v2"
	"gopkg.in/ini.v1"
	"os"
	"strconv"
	"strings"
)

const qrencode = `######################################################################
######################################################################
######################################################################
######              ######  ######    ##        ##              ######
######  ##########  ####  ########      ##    ####  ##########  ######
######  ##      ##  ##  ####  ####      ####    ##  ##      ##  ######
######  ##      ##  ####  ####  ##  ########    ##  ##      ##  ######
######  ##      ##  ####  ####    ######  ##  ####  ##      ##  ######
######  ##########  ######        ##      ##    ##  ##########  ######
######              ##  ##  ##  ##  ##  ##  ##  ##              ######
######################  ####  ##  ##  ##    ##  ######################
######      ##          ######          ####        ######  ##########
######  ########  ####    ##            ##  ####    ####  ####  ######
##########      ##    ####        ####  ##  ####  ##    ##      ######
########  ######  ######    ##    ######  ##              ##  ########
##########  ##  ##      ##    ##  ####  ####    ##  ####  ##    ######
######  ######  ####  ####    ##  ##        ##      ####    ##  ######
##########  ####            ####    ##  ########  ####  ####    ######
##########  ##    ####  ##    ##  ##  ##    ##            ##  ########
##########    ####  ##      ##    ##      ##          ######    ######
##########      ####  ##        ##        ####  ##  ######  ##  ######
######  ##            ##        ##  ##  ##########        ##    ######
########    ##    ##    ######      ##    ##########      ##  ########
######  ########    ##  ####  ######    ######          ##############
######################  ##    ##  ##      ##    ######  ##      ######
######              ##  ########    ####  ##    ##  ##    ##    ######
######  ##########  ##    ##  ##  ##        ##  ######  ######  ######
######  ##      ##  ##      ##    ####    ####            ############
######  ##      ##  ####        ##        ##    ####    ####    ######
######  ##      ##  ##          ##  ##        ##  ##      ####  ######
######  ##########  ##    ####    ######  ##      ##########  ########
######              ##  ##      ##        ####          ####    ######
######################################################################
######################################################################
######################################################################`

var (
	// Init cli
	Init = &cli.Command{
		Name:   "init",
		Usage:  "Init config file",
		Action: AppInit,
	}
	API = &zabbix.API{}
	Cfg = &ini.File{}
)

//AppInit
func AppInit(*cli.Context) error {
DB:
	validate := func(input string) error {
		_, err := strconv.ParseFloat(input, 64)
		if err != nil {
			return errors.New("Invalid number")
		}
		return nil
	}
	//db type
	ProDbtype := promptui.Select{
		Label:    "Select ZbxTable DB Type",
		Items:    []string{"mysql", "postgresql"},
		HideHelp: true,
	}
	_, dbtype, err := ProDbtype.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}
	//db host
	ProDbhost := promptui.Prompt{
		Label:     "DBHost",
		Default:   "localhost",
		AllowEdit: true,
	}
	dbhost, err := ProDbhost.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}
	//db name
	ProDBname := promptui.Prompt{
		Label:     "DBName",
		Default:   "zbxtable",
		AllowEdit: true,
	}
	dbname, err := ProDBname.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}
	//db user
	ProDBuser := promptui.Prompt{
		Label:     "DBUser",
		Default:   "zbxtable",
		AllowEdit: true,
	}
	dbuser, err := ProDBuser.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}
	//db pass
	ProDBpass := promptui.Prompt{
		Label:     "DBPass",
		Default:   "zbxtablepwd123",
		AllowEdit: true,
	}
	dbpass, err := ProDBpass.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}
	//switch DefaultPort
	var DefaultPort string
	switch dbtype {
	case "mysql":
		DefaultPort = "3306"
	case "postgresql":
		DefaultPort = "5432"
	}
	//db port
	ProDBport := promptui.Prompt{
		Label:     "DBPort",
		Default:   DefaultPort,
		AllowEdit: true,
		Validate:  validate,
	}
	dbport, err := ProDBport.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}
	//db check
	err = CheckDb(dbtype, dbhost, dbuser, dbpass, dbname, dbport)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Connection to database " + dbname + "  failed,please reconfigure the database connection information.")
		goto DB
	}
	fmt.Println("Connected to database " + dbname + " successfully!")
WEB:
	//zabbix web url
	ProZbxWeb := promptui.Prompt{
		Label:     "Zabbix Web URL",
		Default:   "http://",
		AllowEdit: true,
	}
	zabbix_web, err := ProZbxWeb.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}
	//zabbix username
	ProZbxUser := promptui.Prompt{
		Label:     "Zabbix Username",
		Default:   "Admin",
		AllowEdit: true,
	}
	zabbix_user, err := ProZbxUser.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}
	//zabbix Password
	ProZbxPass := promptui.Prompt{
		Label:     "Zabbix Password",
		Default:   "zabbix",
		AllowEdit: true,
	}
	zabbix_pass, err := ProZbxPass.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}
	//check zabbix connection
	version, err := CheckZabbixAPI(zabbix_web, zabbix_user, zabbix_pass)
	if err != nil {
		fmt.Println(err)
		fmt.Println("Connection to Zabbix API " + zabbix_web + " /api_jsonrpc.php failed,please reconfigure the zabbix web connection information.")
		goto WEB
	}
	fmt.Println("Connected to Zabbix API successfully！Zabbix version is ", version)
	fmt.Println("The configuration information is as follows:")
	fmt.Println("ZbxTable dbtype:", dbtype)
	fmt.Println("ZbxTable dbhost:", dbhost)
	fmt.Println("ZbxTable dbname:", dbname)
	fmt.Println("ZbxTable dbuser:", dbuser)
	fmt.Println("ZbxTable dbpass:", dbpass)
	fmt.Println("ZbxTable dbport:", dbport)
	fmt.Println("Zabbix Web URL:", zabbix_web)
	fmt.Println("Zabbix Username:", zabbix_user)
	fmt.Println("Zabbix Password:", zabbix_pass)
	prompt := promptui.Select{
		Label:    "Is the configuration information correct[Yes/No]?",
		Items:    []string{"Yes", "No"},
		HideHelp: true,
	}
	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return err
	}
	switch result {
	case "Yes":
		token := strings.ReplaceAll(uuid.New().String(), "-", "")
		err = WriteConf(zabbix_web, zabbix_user, zabbix_pass,
			dbtype, dbhost, dbuser, dbpass, dbname, dbport,
			"", "", "", token)
		if err != nil {
			fmt.Printf("write config file failed %v\n", err)
			return err
		}
		fmt.Println("The configuration file ./conf/app.conf is generated successfully!")
		fmt.Println("Follow WeChat public account to get the latest news!")
		fmt.Println(qrencode)
	case "No":
		goto DB
	}
	return nil
}

//Write config files
func WriteConf(zabbix_web, zabbix_user, zabbix_pass,
	dbtype, dbhost, dbuser, dbpass, dbname, dbport,
	httpport, runmode, timeout, token string) error {
	cfg := ini.Empty()
	//zbxtable info
	cfg.Section("").Key("appname").Comment = "zbxtable"
	//migrate  httpport
	if httpport == "" {
		cfg.Section("").NewKey("httpport", "8084")
	} else {
		cfg.Section("").NewKey("httpport", httpport)
	}
	//migrate  runmode
	if runmode == "" {
		cfg.Section("").NewKey("runmode", "prod")
	} else {
		cfg.Section("").NewKey("runmode", runmode)
	}
	//migrate  timeout
	if timeout == "" {
		cfg.Section("").NewKey("timeout", "12")
	} else {
		cfg.Section("").NewKey("timeout", timeout)
	}
	cfg.Section("").NewKey("appname", "zbxtable")
	cfg.Section("").NewKey("token", token)
	cfg.Section("").NewKey("copyrequestbody", "true")
	//database info
	cfg.Section("").Key("dbtype").Comment = "database"
	cfg.Section("").NewKey("dbtype", dbtype)
	cfg.Section("").NewKey("dbhost", dbhost)
	cfg.Section("").NewKey("dbuser", dbuser)
	cfg.Section("").NewKey("dbpass", dbpass)
	cfg.Section("").NewKey("dbname", dbname)
	cfg.Section("").NewKey("dbport", dbport)
	//zabbix info
	cfg.Section("").Key("zabbix_web").Comment = "zabbix"
	cfg.Section("").NewKey("zabbix_web", zabbix_web)
	cfg.Section("").NewKey("zabbix_user", zabbix_user)
	cfg.Section("").NewKey("zabbix_pass", zabbix_pass)
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
