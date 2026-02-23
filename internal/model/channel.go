package model

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"

	"github.com/google/cel-go/cel"
)

var (
	celMatcher *CELMatcher
	celOnce    sync.Once
)

type CELMatcher struct {
	env   *cel.Env
	cache sync.Map // map[string]cel.Program
}

func GetCELMatcher() *CELMatcher {
	celOnce.Do(func() {
		env, _ := cel.NewEnv(
			cel.Variable("host", cel.StringType),
			cel.Variable("group", cel.StringType),
			cel.Variable("item", cel.StringType),
			cel.Variable("key", cel.StringType),
			cel.Variable("trigger", cel.StringType),
			cel.Variable("severity", cel.StringType),
		)
		celMatcher = &CELMatcher{env: env}
	})
	return celMatcher
}

func (m *CELMatcher) Match(conds []Conditions, data map[string]any) bool {
	if len(conds) == 0 {
		return false
	}
	for _, v := range conds {
		// 映射字段名
		var field string
		switch v.RType {
		case "host":
			field = "host"
		case "group":
			field = "group"
		case "item":
			field = "item"
		case "key":
			field = "key"
		case "trigger":
			field = "trigger"
		case "severity":
			field = "severity"
		default:
			continue
		}

		// 构建表达式，注意处理包含关系和正则匹配
		var expr string
		op := strings.TrimSpace(v.RFunc)
		safeValue := strings.ReplaceAll(v.Rvalue, "'", "\\'")
		switch op {
		case "like":
			expr = fmt.Sprintf("%s.contains('%s')", field, safeValue)
		case "=~":
			// 兼容旧规则正则表达式匹配
			expr = fmt.Sprintf("%s.matches('%s')", field, safeValue)
		case "=":
			expr = fmt.Sprintf("%s == '%s'", field, safeValue)
		case "!=":
			expr = fmt.Sprintf("%s != '%s'", field, safeValue)
		default:
			expr = fmt.Sprintf("%s %s '%s'", field, op, safeValue)
		}

		prog, err := m.getProgram(expr)
		if err != nil {
			logger.Log.Errorf("CEL compile error: %v, expr: %s", err, expr)
			return false
		}

		out, _, err := prog.Eval(data)
		if err != nil {
			logger.Log.Errorf("CEL eval error: %v", err)
			return false
		}
		if res, ok := out.Value().(bool); !ok || !res {
			return false
		}
	}
	return true
}

func (m *CELMatcher) getProgram(expr string) (cel.Program, error) {
	if p, ok := m.cache.Load(expr); ok {
		return p.(cel.Program), nil
	}
	ast, issues := m.env.Compile(expr)
	if issues.Err() != nil {
		return nil, issues.Err()
	}
	prog, err := m.env.Program(ast)
	if err != nil {
		return nil, err
	}
	m.cache.Store(expr, prog)
	return prog, nil
}

// 根据规则生成告警
func GenAlert(alarm *Alarm) {
	// 1. 获取分发规则（遵循 实例普通 -> 实例默认 -> 全局默认 优先级）
	rules, ruleType := getDispatchRules(alarm.ZID)
	if len(rules) == 0 {
		logger.Log.Errorf("no rules found for zid: %d", alarm.ZID)
		return
	}

	// 准备匹配数据
	matchData := map[string]any{
		"host":     alarm.Host,
		"group":    alarm.Hgroup,
		"item":     alarm.ItemName,
		"key":      alarm.Hkey,
		"trigger":  alarm.Message,
		"severity": alarm.Level,
	}

	matcher := GetCELMatcher()
	var matchedRules []Rule
	var firstRule *Rule

	for i := range rules {
		var conds []Conditions
		if err := json.Unmarshal([]byte(rules[i].Conditions), &conds); err != nil {
			continue
		}
		if matcher.Match(conds, matchData) && !isNoneAlarm(alarm.OccurTime, &rules[i]) {
			matchedRules = append(matchedRules, rules[i])
			if firstRule == nil {
				firstRule = &rules[i]
			}
		}
	}

	if len(matchedRules) == 0 {
		return
	}

	// 3. 聚合去重：同人同渠道只生成一条，优化 GetEventUser 调用
	allUserIDs := make(map[string]struct{})
	allGroupIDs := make(map[string]struct{})
	for _, r := range matchedRules {
		if r.UserIds != "" {
			for _, id := range strings.Split(r.UserIds, ",") {
				allUserIDs[id] = struct{}{}
			}
		}
		if r.GroupIds != "" {
			for _, id := range strings.Split(r.GroupIds, ",") {
				allGroupIDs[id] = struct{}{}
			}
		}
	}

	// 统一获取用户信息并建立映射
	userMap, err := getBatchUsers(allUserIDs, allGroupIDs)
	if err != nil {
		logger.Log.Error("getBatchUsers error:", err)
		return
	}

	// 聚合 渠道 -> 用户集合
	chanRecipients := make(map[string]map[string]struct{})
	for _, r := range matchedRules {
		chans := strings.Split(r.Channel, ",")
		rUsers := getRuleUserIDs(r.UserIds, r.GroupIds, userMap)
		for _, c := range chans {
			c = strings.TrimSpace(c)
			if c == "" {
				continue
			}
			if _, ok := chanRecipients[c]; !ok {
				chanRecipients[c] = make(map[string]struct{})
			}
			for _, uid := range rUsers {
				chanRecipients[c][uid] = struct{}{}
			}
		}
	}

	// 4. 按渠道静默检查并推送到缓存
	sentCount := 0
	mutedCount := 0

	for channel, users := range chanRecipients {
		if len(users) == 0 {
			continue
		}

		// 构造 Event 用于静默检查
		event := &Event{
			ID:            alarm.ID,
			ZID:           alarm.ZID,
			Host:          alarm.Host,
			Hgroup:        alarm.Hgroup,
			ItemName:      alarm.ItemName,
			Hkey:          alarm.Hkey,
			Message:       alarm.Message,
			Level:         alarm.Level,
			OccurTime:     alarm.OccurTime,
			Channel:       channel,
			Rule:          firstRule.Name,
			RuleType:      ruleType,
			Hostname:      alarm.Hostname,
			HostsIP:       alarm.HostsIP,
			TriggerID:     alarm.TriggerID,
			ItemID:        alarm.ItemID,
			ItemValue:     alarm.ItemValue,
			Detail:        alarm.Detail,
			Status:        alarm.Status,
			EventID:       alarm.EventID,
			EventDuration: alarm.EventDuration,
		}

		if IsMuted(event) {
			mutedCount++
			continue
		}

		// 发送
		uList := make([]string, 0, len(users))
		for uid := range users {
			uList = append(uList, uid)
		}
		event.ToUsers = strings.Join(uList, ",")

		// 填充实例信息并入队
		sendToQueue(event)
		sentCount++
	}

	// 5. 更新告警状态
	finalStatus := strconv.Itoa(NotifySuccess)
	if ruleType == strconv.Itoa(RuleDefault) {
		finalStatus = strconv.Itoa(NotifyDefault)
	}
	if sentCount == 0 && mutedCount > 0 {
		finalStatus = strconv.Itoa(NotifyMuted)
	}

	ala := Alarm{ID: alarm.ID, NotifyStatus: finalStatus}
	UpdateAlarmStatus(&ala)
}

func getDispatchRules(zid int) ([]Rule, string) {
	var rules []Rule
	// 1. 实例普通规则
	rules = queryRules("1", zid, false)
	if len(rules) > 0 {
		return rules, strconv.Itoa(RuleCust)
	}
	// 2. 实例默认规则
	rules = queryRules("2", zid, false)
	if len(rules) > 0 {
		return rules, strconv.Itoa(RuleDefault)
	}
	// 3. 全局默认规则
	rules = queryRules("2", 0, true)
	return rules, strconv.Itoa(RuleDefault)
}

func queryRules(mType string, zid int, isGlobal bool) []Rule {
	var rules []Rule
	query := DB.Model(&Rule{}).Where("m_type = ? AND status = '0'", mType)
	if isGlobal {
		query = query.Where("z_ids = '* ' OR z_ids = '*'")
	} else if zid > 0 {
		zidStr := strconv.Itoa(zid)
		if DB.Dialector.Name() == "postgres" {
			query = query.Where("(',' || z_ids || ',') LIKE ?", "%,"+zidStr+",%")
		} else {
			query = query.Where("(FIND_IN_SET(?, z_ids) > 0 OR z_ids LIKE ?)", zidStr, "%"+zidStr+"%")
		}
	} else {
		return nil
	}
	query.Find(&rules)
	return rules
}

func getBatchUsers(uids, gids map[string]struct{}) (map[string][]string, error) {
	// 获取所有组的成员
	groupMemberMap := make(map[string][]string)
	if len(gids) > 0 {
		var groups []UserGroup
		keys := make([]string, 0, len(gids))
		for k := range gids {
			keys = append(keys, k)
		}
		if err := DB.Where("id IN ?", keys).Find(&groups).Error; err == nil {
			for _, g := range groups {
				groupMemberMap[strconv.Itoa(g.ID)] = strings.Split(g.Member, ",")
			}
		}
	}
	return groupMemberMap, nil
}

func getRuleUserIDs(uidsStr, gidsStr string, groupMap map[string][]string) []string {
	resMap := make(map[string]struct{})
	if uidsStr != "" {
		for _, id := range strings.Split(uidsStr, ",") {
			resMap[id] = struct{}{}
		}
	}
	if gidsStr != "" {
		for _, gid := range strings.Split(gidsStr, ",") {
			if members, ok := groupMap[gid]; ok {
				for _, mid := range members {
					resMap[mid] = struct{}{}
				}
			}
		}
	}
	res := make([]string, 0, len(resMap))
	for k := range resMap {
		res = append(res, k)
	}
	return res
}

func sendToQueue(event *Event) {
	if event.ZID > 0 {
		instance, err := GetZabbixInstanceByZID(event.ZID)
		if err == nil && instance != nil {
			event.Instance = instance.Instance
			event.InstanceName = instance.Name
		}
	}
	p, _ := json.Marshal(event)
	CacheLPush(event.Channel, p)
}

// mut
func IsMuted(event *Event) bool {
	// 查询匹配当前实例的屏蔽规则（m_type = 3）
	var rules []Rule
	var err error
	if event.ZID > 0 {
		// 根据数据库类型使用不同的查询方式
		zidStr := strconv.Itoa(event.ZID)
		query := DB.Model(&Rule{}).
			Where("m_type = ?", "3").
			Where("status = ?", "0")

		// 检查数据库类型
		if DB.Dialector.Name() == "postgres" {
			// PostgreSQL: 使用 LIKE 匹配（在逗号分隔的字符串中查找）
			query = query.Where("(',' || z_ids || ',') LIKE ?", "%,"+zidStr+",%")
		} else {
			// MySQL: 使用 FIND_IN_SET
			query = query.Where("(FIND_IN_SET(?, z_ids) > 0 OR z_ids LIKE ?)", zidStr, "%"+zidStr+"%")
		}

		err = query.Select("id", "name", "conditions", "z_ids", "note", "s_week",
			"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created").
			Find(&rules).Error
	} else {
		// 如果没有实例ID，查询所有屏蔽规则
		query := DB.Model(&Rule{}).
			Where("m_type = ?", "3").
			Where("status = ?", "0")
		err = query.Select("id", "name", "conditions", "z_ids", "note", "s_week",
			"s_time", "e_time", "user_ids", "group_ids", "channel", "status", "created").
			Find(&rules).Error
	}

	if err != nil {
		logger.Log.Error("查询屏蔽规则失败:", err)
		return false
	}
	if len(rules) == 0 {
		return false
	}
	//mute rule
	for _, v := range rules {
		//not match
		if !MeetEventConditions(event, &v) {
			continue
		}
		//not match
		if IsMuteTime(event, &v) {
			continue
		}
		//not match
		if isMuteChannel(event, &v) {
			continue
		}
		//not muted
		return true
	}
	return false
}

// IsMuteTime not
func IsMuteTime(event *Event, rule *Rule) bool {
	stime, _ := utils.ParTime(rule.Stime)
	etime, _ := utils.ParTime(rule.Etime)
	triggerTime := event.OccurTime.Unix()
	//stime
	if stime.Unix() <= etime.Unix() {
		if triggerTime < stime.Unix() || triggerTime > etime.Unix() {
			return true
		}
	} else {
		if triggerTime < stime.Unix() && triggerTime > etime.Unix() {
			return true
		}
	}
	return false

}

func MeetEventConditions(event *Event, rule *Rule) bool {
	var conds []Conditions
	if err := json.Unmarshal([]byte(rule.Conditions), &conds); err != nil {
		return false
	}
	matchData := map[string]any{
		"host":     event.Host,
		"group":    event.Hgroup,
		"item":     event.ItemName,
		"key":      event.Hkey,
		"trigger":  event.Message,
		"severity": event.Level,
	}
	return GetCELMatcher().Match(conds, matchData)
}

func isMuteChannel(event *Event, rule *Rule) bool {
	return !strings.Contains(rule.Channel, event.Channel)
}

// no alarm
func isNoneAlarm(occurtime time.Time, rule *Rule) bool {
	triggerTime := occurtime.Format("15:04")
	triggerWeek := strconv.Itoa(int(occurtime.Weekday()))
	//stime
	if rule.Stime <= rule.Etime {
		if triggerTime < rule.Stime || triggerTime > rule.Etime {
			return true
		}
	} else {
		if triggerTime < rule.Stime && triggerTime > rule.Etime {
			return true
		}
	}
	//sweek
	return !strings.Contains(rule.Sweek, triggerWeek)
}

func MeetConditions(alarm *Alarm, rule *Rule) bool {
	var conds []Conditions
	if err := json.Unmarshal([]byte(rule.Conditions), &conds); err != nil {
		return false
	}
	matchData := map[string]any{
		"host":     alarm.Host,
		"group":    alarm.Hgroup,
		"item":     alarm.ItemName,
		"key":      alarm.Hkey,
		"trigger":  alarm.Message,
		"severity": alarm.Level,
	}
	return GetCELMatcher().Match(conds, matchData)
}
