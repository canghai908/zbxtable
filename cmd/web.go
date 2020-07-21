package cmd

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
	"github.com/canghai908/zbxtable/models"
	"github.com/canghai908/zbxtable/routers"
	"github.com/urfave/cli"
)

//Web 配置
var Web = cli.Command{
	Name:        "web",
	Usage:       "Start web server",
	Description: "ZbxTable web server",
	Action:      runWeb,
}

//runWeb 启动web
func runWeb(c *cli.Context) {
	logs.SetLogger(logs.AdapterFile, `{"filename":"logs/zbxtable.log","level":7,"maxlines":0,
		"maxsize":0,"daily":true,"maxdays":10,"color":true}`)

	if beego.BConfig.RunMode == "dev" {
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}
	models.ModelsInit()
	routers.RouterInit()
	beego.Run()
}
