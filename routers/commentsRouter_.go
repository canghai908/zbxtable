package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["zbxtable/controllers:AlarmController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:AlarmController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:AlarmController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:AlarmController"],
		beego.ControllerComments{
			Method:           "GetOne",
			Router:           "/:id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:AlarmController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:AlarmController"],
		beego.ControllerComments{
			Method:           "Analysis",
			Router:           "/analysis",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:AlarmController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:AlarmController"],
		beego.ControllerComments{
			Method:           "Export",
			Router:           "/export",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:AlarmController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:AlarmController"],
		beego.ControllerComments{
			Method:           "GetTenant",
			Router:           "/tenant",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:BeforeUserController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:BeforeUserController"],
		beego.ControllerComments{
			Method:           "Login",
			Router:           "/login",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:BeforeUserController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:BeforeUserController"],
		beego.ControllerComments{
			Method:           "Logout",
			Router:           "/logout",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:BeforeUserController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:BeforeUserController"],
		beego.ControllerComments{
			Method:           "Receive",
			Router:           "/receive",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:BeforeUserController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:BeforeUserController"],
		beego.ControllerComments{
			Method:           "Webhook",
			Router:           "/webhook",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:EchartController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:EchartController"],
		beego.ControllerComments{
			Method:           "GetHistory",
			Router:           "/history",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ExpController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ExpController"],
		beego.ControllerComments{
			Method:           "GetItemHistory",
			Router:           "/history",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ExpController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ExpController"],
		beego.ControllerComments{
			Method:           "ExpHostList",
			Router:           "/hosts",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ExpController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ExpController"],
		beego.ControllerComments{
			Method:           "Inspect",
			Router:           "/inspect",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ExpController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ExpController"],
		beego.ControllerComments{
			Method:           "GetItemTrend",
			Router:           "/trend",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:GraphController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:GraphController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           "/:hostid",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:GraphController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:GraphController"],
		beego.ControllerComments{
			Method:           "Exp",
			Router:           "/exp",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:HistoryController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:HistoryController"],
		beego.ControllerComments{
			Method:           "GetHistoryByItemID",
			Router:           "/",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:HostController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:HostController"],
		beego.ControllerComments{
			Method:           "Post",
			Router:           "/",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:HostController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:HostController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:HostController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:HostController"],
		beego.ControllerComments{
			Method:           "GetOne",
			Router:           "/:hostid",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:HostController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:HostController"],
		beego.ControllerComments{
			Method:           "GetMonInterface",
			Router:           "/interface/:hostid",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:HostController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:HostController"],
		beego.ControllerComments{
			Method:           "GetMonLinFileSystem",
			Router:           "/linmon/:hostid",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:HostController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:HostController"],
		beego.ControllerComments{
			Method:           "GetMonItem",
			Router:           "/monitem/:hostid",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:HostController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:HostController"],
		beego.ControllerComments{
			Method:           "Search",
			Router:           "/search",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:HostController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:HostController"],
		beego.ControllerComments{
			Method:           "GetMonWinFileSystem",
			Router:           "/winmon/:hostid",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:HostGroupsController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:HostGroupsController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:HostGroupsController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:HostGroupsController"],
		beego.ControllerComments{
			Method:           "GetGroup",
			Router:           "/all",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:HostGroupsController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:HostGroupsController"],
		beego.ControllerComments{
			Method:           "GetList",
			Router:           "/list",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:HostGroupsController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:HostGroupsController"],
		beego.ControllerComments{
			Method:           "GetHostByGroupID",
			Router:           "/list/:id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ImagesController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ImagesController"],
		beego.ControllerComments{
			Method:           "GetOne",
			Router:           "/:id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:IndexController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:IndexController"],
		beego.ControllerComments{
			Method:           "GetInfo",
			Router:           "/baseinfo/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:IndexController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:IndexController"],
		beego.ControllerComments{
			Method:           "GetEgress",
			Router:           "/egress",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:IndexController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:IndexController"],
		beego.ControllerComments{
			Method:           "GetInventory",
			Router:           "/inventory",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:IndexController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:IndexController"],
		beego.ControllerComments{
			Method:           "GetLinuxCPUTop",
			Router:           "/lincputop",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:IndexController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:IndexController"],
		beego.ControllerComments{
			Method:           "GetOverview",
			Router:           "/overview",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:IndexController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:IndexController"],
		beego.ControllerComments{
			Method:           "GetResrouceTop",
			Router:           "/restop",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:IndexController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:IndexController"],
		beego.ControllerComments{
			Method:           "GetVersion",
			Router:           "/version",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ItemController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ItemController"],
		beego.ControllerComments{
			Method:           "GetItemByKey",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ItemController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ItemController"],
		beego.ControllerComments{
			Method:           "GetAllItemByKey",
			Router:           "/list",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ItemController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ItemController"],
		beego.ControllerComments{
			Method:           "GetAllTraffficByKey",
			Router:           "/traffic",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ManagerController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ManagerController"],
		beego.ControllerComments{
			Method:           "GetOne",
			Router:           "/:id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ManagerController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ManagerController"],
		beego.ControllerComments{
			Method:           "Chpwd",
			Router:           "/chpwd",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ManagerController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ManagerController"],
		beego.ControllerComments{
			Method:           "Info",
			Router:           "/info",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ProblemsController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ProblemsController"],
		beego.ControllerComments{
			Method:           "GetInfo",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ReportController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ReportController"],
		beego.ControllerComments{
			Method:           "GetReportAll",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ReportController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ReportController"],
		beego.ControllerComments{
			Method:           "CreateReport",
			Router:           "/",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ReportController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ReportController"],
		beego.ControllerComments{
			Method:           "GetReportOne",
			Router:           "/:id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ReportController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ReportController"],
		beego.ControllerComments{
			Method:           "Put",
			Router:           "/:id",
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ReportController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ReportController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           "/:id",
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:ReportController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:ReportController"],
		beego.ControllerComments{
			Method:           "UpdateReportStatus",
			Router:           "/status",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:SystemController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:SystemController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:SystemController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:SystemController"],
		beego.ControllerComments{
			Method:           "GetOne",
			Router:           "/:id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:SystemController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:SystemController"],
		beego.ControllerComments{
			Method:           "PutOne",
			Router:           "/:id",
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:SystemController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:SystemController"],
		beego.ControllerComments{
			Method:           "GetEgress",
			Router:           "/egress/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:SystemController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:SystemController"],
		beego.ControllerComments{
			Method:           "PutEgress",
			Router:           "/egress/",
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:SystemController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:SystemController"],
		beego.ControllerComments{
			Method:           "DeployInit",
			Router:           "/init/:id",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TaskLogController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TaskLogController"],
		beego.ControllerComments{
			Method:           "GetTaskByReportID",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TaskLogController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TaskLogController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           "/:id",
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TemplateController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TemplateController"],
		beego.ControllerComments{
			Method:           "GetInfo",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TemplateController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TemplateController"],
		beego.ControllerComments{
			Method:           "GetAll",
			Router:           "/all",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TemplateController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TemplateController"],
		beego.ControllerComments{
			Method:           "GetItemByTempID",
			Router:           "/item/:templateid",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TemplateController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TemplateController"],
		beego.ControllerComments{
			Method:           "GetAllList",
			Router:           "/list",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TopoDataController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TopoDataController"],
		beego.ControllerComments{
			Method:           "CreateData",
			Router:           "/",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TopoDataController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TopoDataController"],
		beego.ControllerComments{
			Method:           "GetTopologyOne",
			Router:           "/:id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TopologyController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TopologyController"],
		beego.ControllerComments{
			Method:           "GetTopologyAll",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TopologyController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TopologyController"],
		beego.ControllerComments{
			Method:           "CreateTopology",
			Router:           "/",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TopologyController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TopologyController"],
		beego.ControllerComments{
			Method:           "Put",
			Router:           "/",
			AllowHTTPMethods: []string{"put"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TopologyController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TopologyController"],
		beego.ControllerComments{
			Method:           "GetTopologyOne",
			Router:           "/:id",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TopologyController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TopologyController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           "/:id",
			AllowHTTPMethods: []string{"delete"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TopologyController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TopologyController"],
		beego.ControllerComments{
			Method:           "DeployTopology",
			Router:           "/deploy",
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TrendController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TrendController"],
		beego.ControllerComments{
			Method:           "GetTrendByItemID",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TriggersController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TriggersController"],
		beego.ControllerComments{
			Method:           "GetInfo",
			Router:           "/",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["zbxtable/controllers:TriggersController"] = append(beego.GlobalControllerRouter["zbxtable/controllers:TriggersController"],
		beego.ControllerComments{
			Method:           "GetOne",
			Router:           "/list",
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
