package routers

import (
	"strings"
	"zbxtable/controllers"
	"zbxtable/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// RouterInitGin 初始化 Gin 路由
func RouterInitGin() *gin.Engine {
	// 设置运行模式
	//gin.SetMode(gin.ReleaseMode)
	gin.SetMode(gin.DebugMode)
	r := gin.New()

	// 中间件
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// CORS 中间件
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "X-Token", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"}
	config.ExposeHeaders = []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// 静态文件
	r.Static("/download", "./download")

	// WebSocket（必须在 NoRoute 之前）
	r.GET("/ws/:id", controllers.WebSocketHandlerGin)

	// 安装引导 API（无需认证，必须在 NoRoute 之前）
	installGroup := r.Group("/install")
	{
		installGroup.GET("/status", controllers.GetInstallStatus)
		installGroup.POST("/check-db", controllers.CheckDatabase)
		installGroup.POST("/check-redis", controllers.CheckRedis)
		installGroup.POST("/install", controllers.DoInstall)
	}

	// 前端静态文件（SPA 应用）
	// 注意：在生产环境中，前端文件应该已经打包到 web 目录
	r.StaticFile("/", "./web/index.html")
	r.Static("/static", "./web/static")
	r.Static("/assets", "./web/assets")

	// SPA 路由支持：所有非 API 路由都返回 index.html（最后注册，作为兜底）
	r.NoRoute(func(c *gin.Context) {
		// 如果是 API 请求，返回 404
		if strings.HasPrefix(c.Request.URL.Path, "/v1") ||
			strings.HasPrefix(c.Request.URL.Path, "/install") ||
			strings.HasPrefix(c.Request.URL.Path, "/ws") {
			c.JSON(404, gin.H{"code": 404, "message": "Not Found"})
			return
		}
		// 否则返回前端页面（SPA 路由）
		c.File("./web/index.html")
	})

	// 安装状态检查中间件（除了安装路由）
	r.Use(middleware.CheckInstallStatusMiddleware())

	// API v1
	v1 := r.Group("/v1")
	{
		// 认证相关（无需 token）
		v1.POST("/login", controllers.LoginGin)
		v1.POST("/logout", controllers.LogoutGin)
		v1.POST("/receive", controllers.ReceiveGin)
		v1.POST("/webhook", controllers.WebhookGin)

		// 需要认证的路由
		api := v1.Group("")
		api.Use(middleware.JWTAuthMiddleware())
		{
			// Zabbix 实例管理（多 Zabbix）
			zabbixGroup := api.Group("/zabbix")
			{
				zabbixGroup.GET("/instances", controllers.ListZabbixInstancesGin)
				zabbixGroup.GET("/instances/:id", controllers.GetZabbixInstanceGin)
				zabbixGroup.POST("/instances", controllers.CreateZabbixInstanceGin)
				zabbixGroup.POST("/instances/test", controllers.TestZabbixInstanceConfigGin)
				zabbixGroup.POST("/instances/:id/test", controllers.TestZabbixInstanceGin)
				zabbixGroup.PUT("/instances/:id", controllers.UpdateZabbixInstanceGin)
				zabbixGroup.DELETE("/instances/:id", controllers.DeleteZabbixInstanceGin)
				zabbixGroup.PUT("/instances/:id/enabled", controllers.EnableZabbixInstanceGin)
				zabbixGroup.PUT("/instances/:id/activate", controllers.ActivateZabbixInstanceGin)
				zabbixGroup.GET("/active", controllers.GetActiveZabbixInstanceGin)

				// 租户绑定（tenant -> zabbix_instance + token）
				zabbixGroup.GET("/tenants", controllers.ListZabbixTenantBindingsGin)
				zabbixGroup.POST("/tenants", controllers.UpsertZabbixTenantBindingGin)
				zabbixGroup.DELETE("/tenants/:id", controllers.DeleteZabbixTenantBindingGin)
			}

			// 首页/仪表板
			indexGroup := api.Group("/index")
			{
				indexGroup.GET("/routers", controllers.GetRouters)
				indexGroup.GET("/baseinfo", controllers.GetBaseInfo)
				indexGroup.GET("/restop", controllers.GetResourceTop)
				indexGroup.GET("/inventory", controllers.GetInventory)
				indexGroup.GET("/overview", controllers.GetOverview)
				indexGroup.GET("/egress", controllers.GetEgressData)
				indexGroup.GET("/version", controllers.GetVersion)
				indexGroup.GET("/session", controllers.GetZbxSession)
			}

			// 告警管理
			alarmGroup := api.Group("/alarm")
			{
				alarmGroup.GET("", controllers.GetAllAlarm)
				alarmGroup.GET("/:id", controllers.GetAlarmByID)
				alarmGroup.GET("/tenant", controllers.GetAlarmTenant)
				alarmGroup.POST("/analysis", controllers.AnalysisAlarm)
				alarmGroup.POST("/export", controllers.ExportAlarm)
				alarmGroup.POST("", controllers.CreateAlarm)
				alarmGroup.PUT("/:id", controllers.UpdateAlarm)
				alarmGroup.DELETE("/:id", controllers.DeleteAlarm)
			}

			// 问题管理
			problemGroup := api.Group("/problem")
			{
				problemGroup.GET("", controllers.GetAllProblem)
			}

			// 触发器管理
			triggerGroup := api.Group("/trigger")
			{
				triggerGroup.GET("", controllers.GetAllTrigger)
				triggerGroup.GET("/list", controllers.GetTriggerList)
			}

			// 导出功能
			exportGroup := api.Group("/export")
			{
				exportGroup.POST("/trend", controllers.ExportTrend)
				exportGroup.POST("/history", controllers.ExportHistory)
				exportGroup.POST("/inspect", controllers.ExportInspect)
				exportGroup.POST("/hosts", controllers.ExportHosts)
				exportGroup.POST("/inventory", controllers.ExportInventory)
			}

			// 主机管理
			hostGroup := api.Group("/host")
			{
				hostGroup.GET("", controllers.GetAllHost)
				hostGroup.GET("/:hostid", controllers.GetHostByID)
				hostGroup.POST("", controllers.UpdateHost)
				hostGroup.GET("/search", controllers.SearchHost)
				hostGroup.GET("/monitem/:hostid", controllers.GetMonItem)
				hostGroup.GET("/interface/:hostid", controllers.GetMonInterface)
				hostGroup.POST("/interface/data", controllers.GetOneInterface)
				hostGroup.GET("/winmon/:hostid", controllers.GetMonWinFileSystem)
				hostGroup.GET("/linmon/:hostid", controllers.GetMonLinFileSystem)
				hostGroup.POST("/graph/:hostid", controllers.GetHostGraph)
			}

			// 主机组管理
			hostGroupGroup := api.Group("/host_group")
			{
				hostGroupGroup.GET("", controllers.GetAllHostGroup)
				hostGroupGroup.GET("/list", controllers.GetAllHostGroupsList)
				hostGroupGroup.GET("/all", controllers.GetAllGroupsList)
				hostGroupGroup.GET("/list/:id", controllers.GetHostsByGroupID)
			}

			// 模板管理
			templateGroup := api.Group("/template")
			{
				templateGroup.GET("", controllers.GetAllTemplate)
				templateGroup.GET("/all", controllers.GetAllTemplateAll)
				templateGroup.GET("/list", controllers.GetAllTemplateList)
				templateGroup.GET("/item/:templateid", controllers.GetItemByTemplateID)
			}

			// 监控项管理
			itemGroup := api.Group("/item")
			{
				itemGroup.GET("", controllers.GetItemByKey)
				itemGroup.GET("/list", controllers.GetAllItemByHostID)
				itemGroup.GET("/traffic", controllers.GetAllTrafficItem)
				itemGroup.GET("/topotraffic", controllers.GetAllReceiveTrafficItem)
			}

			// 历史数据
			historyGroup := api.Group("/history")
			{
				historyGroup.POST("", controllers.GetHistoryByItemID)
			}

			// 趋势数据
			trendGroup := api.Group("/trend")
			{
				trendGroup.GET("", controllers.GetTrendByItemID)
			}

			// 图表管理
			graphGroup := api.Group("/graph")
			{
				graphGroup.POST("", controllers.GetGraphByHostID)
				graphGroup.POST("/exp", controllers.ExportGraph)
			}

			// 图片处理
			imagesGroup := api.Group("/images")
			{
				imagesGroup.GET("/:id", controllers.GetImage)
			}

			// 拓扑管理
			topologyGroup := api.Group("/topology")
			{
				topologyGroup.GET("", controllers.GetAllTopology)
				topologyGroup.GET("/:id", controllers.GetTopologyByID)
				topologyGroup.POST("", controllers.CreateTopology)
				topologyGroup.PUT("/:id", controllers.UpdateTopology)
				topologyGroup.DELETE("/:id", controllers.DeleteTopology)
				topologyGroup.POST("/status", controllers.UpdateTopologyStatus)
			}

			// 拓扑数据
			topodataGroup := api.Group("/topodata")
			{
				topodataGroup.GET("/:id", controllers.GetTopoDataByID)
				topodataGroup.POST("", controllers.CreateTopoData)
			}

			// 系统配置
			systemGroup := api.Group("/system")
			{
				systemGroup.GET("", controllers.GetAllSystem)
				systemGroup.GET("/:id", controllers.GetSystemByID)
				systemGroup.PUT("/:id", controllers.UpdateSystem)
				systemGroup.POST("/init/:id", controllers.SystemInit)
				systemGroup.GET("/egress", controllers.GetEgress)
				systemGroup.PUT("/egress", controllers.UpdateEgress)
				systemGroup.GET("/config", controllers.GetAllConfig)
				systemGroup.PUT("/config/:id", controllers.UpdateConfig)
			}

			// 报表管理
			reportGroup := api.Group("/report")
			{
				reportGroup.GET("", controllers.GetReportGin)
				reportGroup.GET("/:id", controllers.GetReportOneGin)
				reportGroup.POST("", controllers.CreateReportGin)
				reportGroup.PUT("/:id", controllers.UpdateReportGin)
				reportGroup.DELETE("/:id", controllers.DeleteReportGin)
				reportGroup.POST("/checknow", controllers.CheckNowGin)
				reportGroup.POST("/status", controllers.UpdateReportStatusGin)
				reportGroup.GET("/hosts", controllers.GetReportHostsGin)
				reportGroup.GET("/items", controllers.GetReportItemsGin)
			}

			// 任务日志
			taskLogGroup := api.Group("/task_log")
			{
				taskLogGroup.GET("", controllers.GetTaskLogByReportID)
				taskLogGroup.DELETE("/:id", controllers.DeleteTaskLog)
			}

			// 事件日志
			eventLogGroup := api.Group("/event_log")
			{
				eventLogGroup.GET("/:id", controllers.GetEventLogByAlarmID)
			}

			// 规则管理
			ruleGroup := api.Group("/rule")
			{
				ruleGroup.GET("", controllers.GetAllRule)
				ruleGroup.GET("/:id", controllers.GetRuleByID)
				ruleGroup.POST("", controllers.CreateRule)
				ruleGroup.PUT("/:id", controllers.UpdateRule)
				ruleGroup.PUT("/status/:id", controllers.UpdateRuleStatus)
				ruleGroup.DELETE("/:id", controllers.DeleteRule)
			}

			// 用户管理
			userGroup := api.Group("/user")
			{
				userGroup.GET("", controllers.GetUserGin)
				userGroup.POST("", controllers.CreateUserGin)
				userGroup.PUT("/:id", controllers.UpdateUserGin)
				userGroup.DELETE("/:id", controllers.DeleteUserGin)
			}

			// 用户组管理
			groupGroup := api.Group("/group")
			{
				groupGroup.GET("", controllers.GetAllGroup)
				groupGroup.POST("", controllers.CreateGroup)
				groupGroup.PUT("/:id", controllers.UpdateGroup)
				groupGroup.PUT("/member/:id", controllers.UpdateGroupMember)
				groupGroup.DELETE("/:id", controllers.DeleteGroup)
			}

			// AI 聊天
			aiGroup := api.Group("/ai")
			{
				aiGroup.POST("/chat", controllers.AIChat)
			}
		}
	}

	return r
}
