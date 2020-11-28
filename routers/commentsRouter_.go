package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:AlarmController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:AlarmController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:AlarmController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:AlarmController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: "/:id",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:AlarmController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:AlarmController"],
        beego.ControllerComments{
            Method: "Analysis",
            Router: "/analysis",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:AlarmController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:AlarmController"],
        beego.ControllerComments{
            Method: "Export",
            Router: "/export",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:BeforeUserController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:BeforeUserController"],
        beego.ControllerComments{
            Method: "Login",
            Router: "/login",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:BeforeUserController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:BeforeUserController"],
        beego.ControllerComments{
            Method: "Logout",
            Router: "/logout",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:BeforeUserController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:BeforeUserController"],
        beego.ControllerComments{
            Method: "Receive",
            Router: "/receive",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:BeforeUserController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:BeforeUserController"],
        beego.ControllerComments{
            Method: "Webhook",
            Router: "/webhook",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:EchartController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:EchartController"],
        beego.ControllerComments{
            Method: "GetHistory",
            Router: "/history",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ExpController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ExpController"],
        beego.ControllerComments{
            Method: "GetItemHistory",
            Router: "/history",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ExpController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ExpController"],
        beego.ControllerComments{
            Method: "Inspect",
            Router: "/inspect",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ExpController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ExpController"],
        beego.ControllerComments{
            Method: "GetItemTrend",
            Router: "/trend",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:GraphController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:GraphController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/:hostid",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:GraphController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:GraphController"],
        beego.ControllerComments{
            Method: "Exp",
            Router: "/exp",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HistoryController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HistoryController"],
        beego.ControllerComments{
            Method: "GetHistoryByItemID",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostController"],
        beego.ControllerComments{
            Method: "Post",
            Router: "/",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: "/:id",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostController"],
        beego.ControllerComments{
            Method: "GetApplication",
            Router: "/application/:hostid",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostGroupsController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostGroupsController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostGroupsController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostGroupsController"],
        beego.ControllerComments{
            Method: "GetList",
            Router: "/list",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostGroupsController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:HostGroupsController"],
        beego.ControllerComments{
            Method: "GetHostByGroupID",
            Router: "/list/:id",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ImagesController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ImagesController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: "/:id",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:IndexController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:IndexController"],
        beego.ControllerComments{
            Method: "GetInfo",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ItemController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ItemController"],
        beego.ControllerComments{
            Method: "GetItemByKey",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ItemController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ItemController"],
        beego.ControllerComments{
            Method: "GetAllItemByKey",
            Router: "/list",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ManagerController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ManagerController"],
        beego.ControllerComments{
            Method: "GetOne",
            Router: "/:id",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ManagerController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ManagerController"],
        beego.ControllerComments{
            Method: "Chpwd",
            Router: "/chpwd",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ManagerController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ManagerController"],
        beego.ControllerComments{
            Method: "Info",
            Router: "/info",
            AllowHTTPMethods: []string{"post"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ProblemsController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:ProblemsController"],
        beego.ControllerComments{
            Method: "GetInfo",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:TemplateController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:TemplateController"],
        beego.ControllerComments{
            Method: "GetInfo",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:TemplateController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:TemplateController"],
        beego.ControllerComments{
            Method: "GetAll",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:TrendController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:TrendController"],
        beego.ControllerComments{
            Method: "GetTrendByItemID",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

    beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:TriggersController"] = append(beego.GlobalControllerRouter["github.com/canghai908/zbxtable/controllers:TriggersController"],
        beego.ControllerComments{
            Method: "GetInfo",
            Router: "/",
            AllowHTTPMethods: []string{"get"},
            MethodParams: param.Make(),
            Filters: nil,
            Params: nil})

}
