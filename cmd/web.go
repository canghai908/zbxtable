package cmd

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/canghai908/zbxtable/models"
	"github.com/canghai908/zbxtable/routers"
	"github.com/urfave/cli"
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
\________|\_______/ \__|  \__|   \__|   \__|  \__|\_______/ \________|\________|
`

//Web 配置
var Web = cli.Command{
	Name:        "web",
	Usage:       "Start web server",
	Description: "ZbxTable web server",
	Action:      runWeb,
}

//runWeb 启动web
func runWeb(c *cli.Context) {
	fmt.Println(prompt)
	logs.SetLogger(logs.AdapterFile, `{"filename":"logs/zbxtable.log","level":7,"maxlines":0,
		"maxsize":0,"daily":true,"maxdays":10,
		"color":true,"perm":"0755"}`)
	_, err := os.Stat("./conf/app.conf")
	if err != nil {
		fmt.Println(err)
		fmt.Println("Please run 'zbxtable init' to create app.conf")
		return
	}
	err = CheckConf()
	if err != nil {
		fmt.Println(err)
		return
	}
	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	models.ModelsInit()
	routers.RouterInit()
	beego.Run()
}
