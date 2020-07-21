// routers init
// @APIVersion 1.0.0
// @Title beego Test API
// @Description beego has a very cool tools to autogenerate documents for your API
// @Contact astaxie@gmail.com
// @TermsOfServiceUrl http://beego.me/
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html

package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
	"github.com/canghai908/zbxtable/controllers"
)

//RouterInit router
func RouterInit() {
	beego.SetStaticPath("/download", "download")
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "X-Token", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		AllowCredentials: true,
	}))
	ns := beego.NewNamespace("/v1",
		beego.NSRouter("/login", &controllers.BeforeUserController{}, "*:Login"),
		beego.NSRouter("/logout", &controllers.BeforeUserController{}, "*:Logout"),
		beego.NSRouter("/receive", &controllers.BeforeUserController{}, "*:Receive"),
		beego.NSRouter("/webhook", &controllers.BeforeUserController{}, "*:Webhook"),
		beego.NSNamespace("/index",
			beego.NSInclude(
				&controllers.IndexController{},
			),
		),
		beego.NSNamespace("/alarm",
			beego.NSInclude(
				&controllers.AlarmController{},
			),
		),
		beego.NSNamespace("/problem",
			beego.NSInclude(
				&controllers.ProblemsController{},
			),
		),
		beego.NSNamespace("/trigger",
			beego.NSInclude(
				&controllers.TriggersController{},
			),
		),
		beego.NSNamespace("/manager",
			beego.NSInclude(
				&controllers.ManagerController{},
			),
		),
		beego.NSNamespace("/export",
			beego.NSInclude(
				&controllers.ExpController{},
			),
		),
		beego.NSNamespace("/host",
			beego.NSInclude(
				&controllers.HostController{},
			),
		),
		beego.NSNamespace("/host_group",
			beego.NSInclude(
				&controllers.HostGroupsController{},
			),
		),
		beego.NSNamespace("/template",
			beego.NSInclude(
				&controllers.TemplateController{},
			),
		),
		beego.NSNamespace("/item",
			beego.NSInclude(
				&controllers.ItemController{},
			),
		),
		beego.NSNamespace("/history",
			beego.NSInclude(
				&controllers.HistoryController{},
			),
		),
		beego.NSNamespace("/trend",
			beego.NSInclude(
				&controllers.TrendController{},
			),
		),
		beego.NSNamespace("/graph",
			beego.NSInclude(
				&controllers.GraphController{},
			),
		),
		beego.NSNamespace("/images",
			beego.NSInclude(
				&controllers.ImagesController{},
			),
		),
		beego.NSNamespace("/echart",
			beego.NSInclude(
				&controllers.EchartController{},
			),
		),
	)
	beego.AddNamespace(ns)
}
