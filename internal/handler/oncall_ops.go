package handler

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"zbxtable/internal/middleware"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"
)

type onCallViewer struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func requireOnCallReader() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := onCallActor(c)
		if name == "" {
			response.Unauthorized(c, "登录状态失效")
			c.Abort()
			return
		}
		var user model.User
		if err := model.DB.Select("id", "username", "role", "status").Where("username = ?", name).First(&user).Error; err != nil || user.Status != 0 || (user.Role != "admin" && user.Role != "user") {
			response.Forbidden(c, "无权限查看排班")
			c.Abort()
			return
		}
		c.Set("oncall_viewer", onCallViewer{ID: user.ID, Username: user.Username, Role: user.Role})
		c.Next()
	}
}

// RegisterOnCallRoutes keeps read and write authorization identical in tests and production.
// The caller's parent API group must already apply token authentication.
func RegisterOnCallRoutes(group *gin.RouterGroup) {
	group.Use(requireOnCallReader())
	group.GET("", ListOnCallSchedules)
	group.GET("/directory", OnCallDirectory)
	group.GET("/overview", OnCallOverview)
	group.GET("/:id", GetOnCallSchedule)
	group.POST("/:id/preview", PreviewOnCallSchedule)
	group.POST("", middleware.RequireAdmin(), CreateOnCallSchedule)
	group.PUT("/:id", middleware.RequireAdmin(), UpdateOnCallSchedule)
	group.DELETE("/:id", middleware.RequireAdmin(), DeleteOnCallSchedule)
	group.POST("/:id/overrides", middleware.RequireAdmin(), CreateOnCallOverride)
	group.DELETE("/:id/overrides/:override_id", middleware.RequireAdmin(), DeleteOnCallOverride)
	group.GET("/:id/audit", middleware.RequireAdmin(), ListOnCallAudit)
}

func OnCallDirectory(c *gin.Context) {
	type channels struct {
		Mail   bool `json:"mail"`
		Wechat bool `json:"wechat"`
		Robot  bool `json:"wechat_robot"`
	}
	type account struct {
		ID       int      `json:"id"`
		Username string   `json:"username"`
		Status   int64    `json:"status"`
		Channels channels `json:"channels"`
	}
	type group struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	users := []model.User{}
	if err := model.DB.Select("id", "username", "status", "email", "wechat", "wechat_robot_key").Order("id").Find(&users).Error; err != nil {
		response.DatabaseError(c, "读取排班人员失败")
		return
	}
	safe := make([]account, 0, len(users))
	for _, u := range users {
		safe = append(safe, account{ID: u.ID, Username: u.Username, Status: u.Status, Channels: channels{Mail: strings.TrimSpace(u.Email) != "", Wechat: strings.TrimSpace(u.Wechat) != "", Robot: strings.TrimSpace(u.WechatRobotKey) != ""}})
	}
	groups := []group{}
	if err := model.DB.Model(&model.UserGroup{}).Select("id", "name").Order("id").Scan(&groups).Error; err != nil {
		response.DatabaseError(c, "读取排班用户组失败")
		return
	}
	response.Success(c, gin.H{"viewer": c.MustGet("oncall_viewer"), "users": safe, "groups": groups})
}

func OnCallOverview(c *gin.Context) {
	viewer := c.MustGet("oncall_viewer").(onCallViewer)
	schedules := []model.OnCallSchedule{}
	query := model.DB.Model(&model.OnCallSchedule{})
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if err := query.Order("id DESC").Find(&schedules).Error; err != nil {
		response.DatabaseError(c, "读取排班失败")
		return
	}
	overrides := []model.OnCallOverride{}
	if len(schedules) > 0 {
		ids := make([]int, 0, len(schedules))
		for _, s := range schedules {
			ids = append(ids, s.ID)
		}
		if err := model.DB.Where("schedule_id IN ?", ids).Order("id").Find(&overrides).Error; err != nil {
			response.DatabaseError(c, "读取临时代班失败")
			return
		}
	}
	bySchedule := map[int][]model.OnCallOverride{}
	for _, o := range overrides {
		bySchedule[o.ScheduleID] = append(bySchedule[o.ScheduleID], o)
	}
	groupRows := []model.UserGroup{}
	if err := model.DB.Select("id", "member").Find(&groupRows).Error; err != nil {
		response.DatabaseError(c, "读取备用用户组失败")
		return
	}
	groups := map[int][]int{}
	for _, g := range groupRows {
		for _, text := range strings.Split(g.Member, ",") {
			id, e := strconv.Atoi(strings.TrimSpace(text))
			if e == nil && id > 0 {
				groups[g.ID] = append(groups[g.ID], id)
			}
		}
	}
	now := time.Now()
	mine := c.Query("mine") == "1" || c.Query("mine") == "true"
	filtered := make([]model.OnCallSchedule, 0, len(schedules))
	for _, s := range schedules {
		if !mine || model.OnCallViewerRole(s, bySchedule[s.ID], viewer.ID, groups, now) != "" {
			filtered = append(filtered, s)
		}
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 20
	}
	rows := []model.OnCallOverview{}
	// Check the page against the number of items before multiplication so hostile
	// page values cannot overflow into negative slice indexes.
	if page <= (len(filtered)+limit-1)/limit {
		start := (page - 1) * limit
		end := start + limit
		if end > len(filtered) {
			end = len(filtered)
		}
		for _, s := range filtered[start:end] {
			rows = append(rows, model.BuildOnCallOverview(s, bySchedule[s.ID], viewer.ID, groups, now))
		}
	}
	response.Success(c, gin.H{"items": rows, "total": len(filtered), "at": now.UTC().Format(time.RFC3339Nano)})
}
