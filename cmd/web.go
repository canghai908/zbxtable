package cmd

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/canghai908/zbxtable/models"
	"github.com/canghai908/zbxtable/routers"
	"github.com/urfave/cli/v2"
	"os"
)

const prompt = `
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
	logs.Info(prompt)
	if err != nil {
		logs.Info(err)
		return err
	}

	_, err = os.Stat("./conf/app.conf")
	if err != nil {
		logs.Error(err)
		logs.Error("Please run 'zbxtable init' to create app.conf")
		return err
	}
	err = PreCheckConf()
	if err != nil {
		logs.Error(err)
		return err
	}
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	models.ModelsInit()
	routers.RouterInit()
	beego.Run()
	return nil
}
