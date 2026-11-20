package handler

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
)

// PreviewRuleDispatch evaluates one draft rule without enqueueing notifications.
// It includes scope, timing, on-call delivery and current mute rules.
func PreviewRuleDispatch(c *gin.Context) {
	var req struct {
		Rule model.Rule     `json:"rule"`
		At   string         `json:"at"`
		ZID  int            `json:"zid"`
		Data map[string]any `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "无效预览请求")
		return
	}
	at, err := time.Parse(time.RFC3339, req.At)
	if err != nil {
		response.ValidationError(c, "请选择带时区的测试时间")
		return
	}
	r := req.Rule
	if r.MType != "1" && r.MType != "2" {
		response.ValidationError(c, "仅支持分发规则预览")
		return
	}
	if err := model.ValidateRuleConditions(r.Conditions, r.MType); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	if err := model.ValidateScheduleReferences(r.ScheduleIDs); err != nil {
		response.ValidationError(c, err.Error())
		return
	}
	result := struct {
		Matched  bool                    `json:"matched"`
		Reason   string                  `json:"reason"`
		Channels []model.DispatchChannel `json:"channels"`
	}{Channels: []model.DispatchChannel{}}
	// Legacy rule time windows are evaluated in the server timezone in production.
	at = at.In(time.Local)
	if r.Status != "0" {
		result.Reason = "规则未启用"
		response.Success(c, result)
		return
	}
	scoped := r.ZIDs == "*" || r.ZIDs == "* "
	for _, z := range strings.Split(r.ZIDs, ",") {
		if strings.TrimSpace(z) == strconv.Itoa(req.ZID) && req.ZID > 0 {
			scoped = true
		}
	}
	if !scoped {
		result.Reason = "测试实例不在规则范围内"
		response.Success(c, result)
		return
	}
	alarm := model.Alarm{ZID: req.ZID, OccurTime: at}
	value := func(k string) string { s, _ := req.Data[k].(string); return s }
	alarm.Host = value("host")
	alarm.Hgroup = value("group")
	alarm.ItemName = value("item")
	alarm.Hkey = value("key")
	alarm.Message = value("trigger")
	alarm.Level = value("severity")
	if !model.MeetConditions(&alarm, &r) {
		result.Reason = "告警不满足匹配条件"
		response.Success(c, result)
		return
	}
	if !model.DispatchRuleTimeMatches(at, r) {
		result.Reason = "不在分发规则生效时间内"
		response.Success(c, result)
		return
	}
	groups, err := model.LoadDispatchGroups([]model.Rule{r})
	if err != nil {
		response.InternalError(c, fmt.Sprintf("读取接收组失败: %v", err))
		return
	}
	channels, err := model.ResolveDispatchRecipients([]model.Rule{r}, groups, at, model.LoadScheduleDelivery)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	for i := range channels {
		event := model.Event{ZID: req.ZID, Host: alarm.Host, Hgroup: alarm.Hgroup, ItemName: alarm.ItemName, Hkey: alarm.Hkey, Message: alarm.Message, Level: alarm.Level, OccurTime: at, Channel: channels[i].Channel}
		channels[i].Muted = model.IsMuted(&event)
	}
	result.Matched = true
	result.Reason = "规则命中；仅预览，不发送通知。排班按测试时间解析，实际重试可能沿用事件收件人快照。"
	result.Channels = channels
	response.Success(c, result)
}
