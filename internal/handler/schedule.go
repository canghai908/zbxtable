package handler

import (
	"errors"
	"strconv"
	"strings"
	"time"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func onCallID(c *gin.Context, key string) (int, bool) {
	id, err := strconv.Atoi(c.Param(key))
	if err != nil || id <= 0 {
		response.BadRequest(c, "无效排班或代班ID")
		return 0, false
	}
	return id, true
}
func onCallActor(c *gin.Context) string {
	actor, _ := c.Get("username")
	s, _ := actor.(string)
	return s
}
func onCallError(c *gin.Context, err error) {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		response.NotFound(c, "排班或代班不存在")
		return
	}
	response.ValidationError(c, err.Error())
}
func ListOnCallSchedules(c *gin.Context) {
	items := []model.OnCallSchedule{}
	q := model.DB.Model(&model.OnCallSchedule{})
	if name := strings.TrimSpace(c.Query("name")); name != "" {
		q = q.Where("name LIKE ?", "%"+name+"%")
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 100
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.DatabaseError(c, "读取排班失败")
		return
	}
	if err := q.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&items).Error; err != nil {
		response.DatabaseError(c, "读取排班失败")
		return
	}
	response.SuccessWithPage(c, items, total)
}
func GetOnCallSchedule(c *gin.Context) {
	id, ok := onCallID(c, "id")
	if !ok {
		return
	}
	s, o, err := model.LoadOnCallSchedule(id)
	if err != nil {
		onCallError(c, err)
		return
	}
	response.Success(c, gin.H{"schedule": s, "overrides": o})
}
func CreateOnCallSchedule(c *gin.Context) {
	var s model.OnCallSchedule
	if err := c.ShouldBindJSON(&s); err != nil {
		response.BadRequest(c, "无效排班请求")
		return
	}
	s.ID = 0
	s.Created = time.Time{}
	s.Updated = time.Time{}
	if err := model.SaveOnCallSchedule(&s, onCallActor(c)); err != nil {
		onCallError(c, err)
		return
	}
	response.Success(c, s)
}
func UpdateOnCallSchedule(c *gin.Context) {
	id, ok := onCallID(c, "id")
	if !ok {
		return
	}
	var s model.OnCallSchedule
	if err := c.ShouldBindJSON(&s); err != nil {
		response.BadRequest(c, "无效排班请求")
		return
	}
	s.ID = id
	s.Created = time.Time{}
	s.Updated = time.Time{}
	if err := model.SaveOnCallSchedule(&s, onCallActor(c)); err != nil {
		onCallError(c, err)
		return
	}
	response.Success(c, s)
}
func DeleteOnCallSchedule(c *gin.Context) {
	id, ok := onCallID(c, "id")
	if !ok {
		return
	}
	if err := model.DeleteOnCallSchedule(id, onCallActor(c)); err != nil {
		onCallError(c, err)
		return
	}
	response.SuccessWithMessage(c, "排班已删除", nil)
}
func CreateOnCallOverride(c *gin.Context) {
	id, ok := onCallID(c, "id")
	if !ok {
		return
	}
	var o model.OnCallOverride
	if err := c.ShouldBindJSON(&o); err != nil {
		response.BadRequest(c, "无效代班请求")
		return
	}
	o.ID = 0
	o.ScheduleID = id
	o.Created = time.Time{}
	o.CreatedBy = ""
	if err := model.AddOnCallOverride(&o, onCallActor(c)); err != nil {
		onCallError(c, err)
		return
	}
	response.Success(c, o)
}
func DeleteOnCallOverride(c *gin.Context) {
	id, ok := onCallID(c, "id")
	if !ok {
		return
	}
	oid, ok := onCallID(c, "override_id")
	if !ok {
		return
	}
	if err := model.RemoveOnCallOverride(id, oid, onCallActor(c)); err != nil {
		onCallError(c, err)
		return
	}
	response.SuccessWithMessage(c, "已取消临时代班", nil)
}
func PreviewOnCallSchedule(c *gin.Context) {
	id, ok := onCallID(c, "id")
	if !ok {
		return
	}
	var req struct {
		At      string `json:"at"`
		Days    int    `json:"days"`
		Channel string `json:"channel"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效预览请求")
		return
	}
	at, err := time.Parse(time.RFC3339, req.At)
	if err != nil {
		response.ValidationError(c, "预览时间必须为包含时区的 RFC3339")
		return
	}
	if req.Days == 0 {
		req.Days = 7
	}
	if req.Days < 1 || req.Days > 31 {
		response.ValidationError(c, "预览天数必须为1至31天")
		return
	}
	if req.Channel != "" && req.Channel != "mail" && req.Channel != "wechat" && req.Channel != "wechat_robot" {
		response.ValidationError(c, "预览渠道必须为 mail/wechat/wechat_robot")
		return
	}
	s, o, err := model.LoadOnCallSchedule(id)
	if err != nil {
		onCallError(c, err)
		return
	}
	current, err := model.ResolveOnCall(s, o, at)
	if err != nil {
		onCallError(c, err)
		return
	}
	slots, err := model.PreviewOnCallSlots(s, o, at, req.Days)
	if err != nil {
		onCallError(c, err)
		return
	}
	if req.Channel != "" {
		users, groups, err := model.LoadScheduleDeliveryInputs(s, o)
		if err != nil {
			onCallError(c, err)
			return
		}
		decision, err := model.ResolveScheduleDelivery(s, o, at, req.Channel, users, groups)
		if err != nil {
			onCallError(c, err)
			return
		}
		current.UserIDs, current.Source, current.Reason = decision.UserIDs, decision.Source, decision.Reason
		for i := range slots {
			start, _ := time.Parse(time.RFC3339, slots[i].Start)
			d, err := model.ResolveScheduleDelivery(s, o, start, req.Channel, users, groups)
			if err != nil {
				onCallError(c, err)
				return
			}
			slots[i].UserIDs, slots[i].Source, slots[i].Reason = d.UserIDs, d.Source, d.Reason
		}
	}
	response.Success(c, gin.H{"current": current, "slots": slots, "timezone": s.Timezone})
}
func ListOnCallAudit(c *gin.Context) {
	id, ok := onCallID(c, "id")
	if !ok {
		return
	}
	items := []model.OnCallAudit{}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 200 {
		limit = 50
	}
	q := model.DB.Model(&model.OnCallAudit{}).Where("schedule_id = ?", id)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		response.DatabaseError(c, "读取排班审计失败")
		return
	}
	if err := q.Order("id DESC").Offset((page - 1) * limit).Limit(limit).Find(&items).Error; err != nil {
		response.DatabaseError(c, "读取排班审计失败")
		return
	}
	response.SuccessWithPage(c, items, total)
}
