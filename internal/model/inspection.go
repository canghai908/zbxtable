package model

import (
	"errors"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"
)

const (
	defaultInspectCpuThreshold  = 80.0
	defaultInspectMemThreshold  = 80.0
	defaultInspectDiskThreshold = 85.0
)

var (
	inspectionRunMu    sync.Mutex
	runningInspectTask = make(map[int]bool)
)

func markInspectionTaskRunning(taskID int) bool {
	inspectionRunMu.Lock()
	defer inspectionRunMu.Unlock()
	if runningInspectTask[taskID] {
		return false
	}
	runningInspectTask[taskID] = true
	return true
}

func unmarkInspectionTaskRunning(taskID int) {
	inspectionRunMu.Lock()
	delete(runningInspectTask, taskID)
	inspectionRunMu.Unlock()
}

func cleanInspectionCSV(raw string, dedupe bool) string {
	s := utils.VAarToStr(raw)
	parts := strings.Split(s, ",")
	seen := make(map[string]struct{}, len(parts))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed == "" {
			continue
		}
		if dedupe {
			if _, ok := seen[trimmed]; ok {
				continue
			}
			seen[trimmed] = struct{}{}
		}
		out = append(out, trimmed)
	}
	return strings.Join(out, ",")
}

func splitInspectionCSV(raw string) []string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func normalizeInspectionChannelsCSV(raw string) string {
	parts := splitInspectionCSV(utils.VAarToStr(raw))
	out := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, ch := range parts {
		norm := normalizeChannel(ch)
		if norm == "" || !IsSupportedChannel(norm) || seen[norm] {
			continue
		}
		seen[norm] = true
		out = append(out, norm)
	}
	return strings.Join(out, ",")
}

func defaultCronForInspectionCycle(cycle string) string {
	switch strings.ToLower(strings.TrimSpace(cycle)) {
	case "week":
		return "0 0 8 * * 1"
	default:
		return "0 0 8 * * *"
	}
}

func normalizeInspectionTaskDefaults(t *InspectionTask) error {
	if t == nil {
		return errors.New("巡检任务参数不能为空")
	}
	t.Name = strings.TrimSpace(t.Name)
	if t.Name == "" {
		return errors.New("巡检任务名称不能为空")
	}
	t.InspectMode = strings.ToLower(strings.TrimSpace(t.InspectMode))
	if t.InspectMode == "" {
		t.InspectMode = "scheduled"
	}
	if t.InspectMode != "scheduled" && t.InspectMode != "manual" {
		return errors.New("巡检模式仅支持 scheduled 或 manual")
	}

	t.Cycle = strings.ToLower(strings.TrimSpace(t.Cycle))
	if t.Cycle == "" {
		t.Cycle = "day"
	}
	if t.Cycle != "day" && t.Cycle != "week" && t.Cycle != "custom" {
		t.Cycle = "day"
	}

	t.CronExpr = strings.TrimSpace(t.CronExpr)
	if t.InspectMode == "scheduled" {
		if t.CronExpr == "" {
			t.CronExpr = defaultCronForInspectionCycle(t.Cycle)
		}
		if _, err := taskCronParser.Parse(t.CronExpr); err != nil {
			return fmt.Errorf("Cron 表达式无效（需 6 段：秒 分 时 日 月 周）: %v", err)
		}
	}

	if t.CpuThreshold <= 0 {
		t.CpuThreshold = defaultInspectCpuThreshold
	} else if t.CpuThreshold > 100 {
		t.CpuThreshold = 100
	}
	if t.MemThreshold <= 0 {
		t.MemThreshold = defaultInspectMemThreshold
	} else if t.MemThreshold > 100 {
		t.MemThreshold = 100
	}
	if t.DiskThreshold <= 0 {
		t.DiskThreshold = defaultInspectDiskThreshold
	} else if t.DiskThreshold > 100 {
		t.DiskThreshold = 100
	}

	t.ZIDs = cleanInspectionCSV(t.ZIDs, true)
	if t.ZIDs == "" {
		t.ZIDs = "*"
	}
	t.HostTypes = cleanInspectionCSV(t.HostTypes, true)
	if t.HostTypes == "" {
		t.HostTypes = "*"
	}
	t.GroupIDs = cleanInspectionCSV(t.GroupIDs, true)
	t.NotifyChannels = normalizeInspectionChannelsCSV(t.NotifyChannels)
	t.NotifyUserIDs = cleanInspectionCSV(t.NotifyUserIDs, true)
	t.NotifyGroupIDs = cleanInspectionCSV(t.NotifyGroupIDs, true)
	t.NotifyEmails = cleanInspectionCSV(t.NotifyEmails, true)

	t.Status = strings.TrimSpace(t.Status)
	if t.Status != "0" && t.Status != "1" {
		t.Status = "1"
	}
	t.Desc = strings.TrimSpace(t.Desc)
	return nil
}

// ListInspectionTasks 分页查询巡检任务
func ListInspectionTasks(page, limit, name, status string) (int64, []InspectionTask, error) {
	if DB == nil {
		return 0, []InspectionTask{}, errors.New("database not initialized")
	}
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	if pages <= 0 {
		pages = 1
	}
	if limits <= 0 {
		limits = 10
	}

	query := DB.Model(&InspectionTask{})
	if name = strings.TrimSpace(name); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status = strings.TrimSpace(status); status != "" {
		query = query.Where("status = ?", status)
	}

	var cnt int64
	if err := query.Count(&cnt).Error; err != nil {
		return 0, []InspectionTask{}, err
	}

	var list []InspectionTask
	offset := (pages - 1) * limits
	if err := query.Order("id DESC").Limit(limits).Offset(offset).Find(&list).Error; err != nil {
		return 0, []InspectionTask{}, err
	}
	return cnt, list, nil
}

// GetInspectionTaskByID 根据 ID 获取巡检任务
func GetInspectionTaskByID(id int) (*InspectionTask, error) {
	if DB == nil {
		return nil, errors.New("database not initialized")
	}
	var task InspectionTask
	if err := DB.Where("id = ?", id).First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// CreateInspectionTask 创建巡检任务
func CreateInspectionTask(t *InspectionTask) (int, error) {
	if DB == nil {
		return 0, errors.New("database not initialized")
	}
	if err := normalizeInspectionTaskDefaults(t); err != nil {
		return 0, err
	}
	t.ExecStatus = NoBegin
	if err := DB.Create(t).Error; err != nil {
		return 0, err
	}
	ReloadTaskSchedulerIfRunning()
	return t.ID, nil
}

// UpdateInspectionTask 更新巡检任务
func UpdateInspectionTask(t *InspectionTask) error {
	if DB == nil {
		return errors.New("database not initialized")
	}
	if t == nil || t.ID <= 0 {
		return errors.New("无效的巡检任务 ID")
	}
	var existing InspectionTask
	if err := DB.Where("id = ?", t.ID).First(&existing).Error; err != nil {
		return err
	}
	if err := normalizeInspectionTaskDefaults(t); err != nil {
		return err
	}
	updates := map[string]interface{}{
		"name":             t.Name,
		"z_ids":            t.ZIDs,
		"host_types":       t.HostTypes,
		"group_ids":        t.GroupIDs,
		"inspect_mode":     t.InspectMode,
		"cycle":            t.Cycle,
		"cron_expr":        t.CronExpr,
		"cpu_threshold":    t.CpuThreshold,
		"mem_threshold":    t.MemThreshold,
		"disk_threshold":   t.DiskThreshold,
		"notify_channels":  t.NotifyChannels,
		"notify_user_ids":  t.NotifyUserIDs,
		"notify_group_ids": t.NotifyGroupIDs,
		"notify_emails":    t.NotifyEmails,
		"status":           t.Status,
		"desc":             t.Desc,
	}
	if err := DB.Model(&InspectionTask{}).Where("id = ?", t.ID).Updates(updates).Error; err != nil {
		return err
	}
	ReloadTaskSchedulerIfRunning()
	return nil
}

// UpdateInspectionTaskStatus 启用或停用巡检任务
func UpdateInspectionTaskStatus(id int, status string) error {
	if DB == nil {
		return errors.New("database not initialized")
	}
	s := strings.TrimSpace(status)
	if s != "0" && s != "1" {
		var existing InspectionTask
		if err := DB.Where("id = ?", id).First(&existing).Error; err != nil {
			return err
		}
		if existing.Status == "1" {
			s = "0"
		} else {
			s = "1"
		}
	}
	if err := DB.Model(&InspectionTask{}).Where("id = ?", id).Update("status", s).Error; err != nil {
		return err
	}
	ReloadTaskSchedulerIfRunning()
	return nil
}

// DeleteInspectionTask 删除巡检任务及其关联历史报告
func DeleteInspectionTask(id int) error {
	if DB == nil {
		return errors.New("database not initialized")
	}
	var reports []InspectionReport
	if err := DB.Where("task_id = ?", id).Find(&reports).Error; err == nil {
		for _, r := range reports {
			removeInspectionReportFiles(&r)
		}
	}
	if err := DB.Where("task_id = ?", id).Delete(&InspectionReport{}).Error; err != nil {
		return err
	}
	if err := DB.Where("id = ?", id).Delete(&InspectionTask{}).Error; err != nil {
		return err
	}
	ReloadTaskSchedulerIfRunning()
	return nil
}

// ListInspectionReports 分页查询巡检报告记录
func ListInspectionReports(page, limit, taskID, grade, status string) (int64, []InspectionReport, error) {
	if DB == nil {
		return 0, []InspectionReport{}, errors.New("database not initialized")
	}
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	if pages <= 0 {
		pages = 1
	}
	if limits <= 0 {
		limits = 10
	}

	query := DB.Model(&InspectionReport{})
	if tid := strings.TrimSpace(taskID); tid != "" {
		query = query.Where("task_id = ?", tid)
	}
	if g := strings.TrimSpace(grade); g != "" {
		query = query.Where("health_grade = ?", g)
	}
	if st := strings.TrimSpace(status); st != "" {
		query = query.Where("status = ?", st)
	}

	var cnt int64
	if err := query.Count(&cnt).Error; err != nil {
		return 0, []InspectionReport{}, err
	}

	var list []InspectionReport
	offset := (pages - 1) * limits
	if err := query.Omit("hosts_json").Order("id DESC").Limit(limits).Offset(offset).Find(&list).Error; err != nil {
		return 0, []InspectionReport{}, err
	}
	return cnt, list, nil
}

// GetInspectionReportDetail 获取完整巡检报告详情
func GetInspectionReportDetail(id int) (*InspectionReportDetail, error) {
	if DB == nil {
		return nil, errors.New("database not initialized")
	}
	var r InspectionReport
	if err := DB.Where("id = ?", id).First(&r).Error; err != nil {
		return nil, err
	}
	detail := &InspectionReportDetail{
		InspectionReport: r,
		Summary: InspectionSummary{
			InstanceStats:   []InspectionInstanceStat{},
			TypeStats:       []InspectionTypeStat{},
			Recommendations: []string{},
			TopAlarms:       []InspectionAlarmItem{},
		},
		Hosts: []InspectionHostItem{},
	}
	if strings.TrimSpace(r.SummaryJSON) != "" {
		if err := json.Unmarshal([]byte(r.SummaryJSON), &detail.Summary); err != nil {
			logger.Log.Warnf("解析巡检报告摘要 JSON 失败 (id=%d): %v", r.ID, err)
		}
		if detail.Summary.InstanceStats == nil {
			detail.Summary.InstanceStats = []InspectionInstanceStat{}
		}
		if detail.Summary.TypeStats == nil {
			detail.Summary.TypeStats = []InspectionTypeStat{}
		}
		if detail.Summary.Recommendations == nil {
			detail.Summary.Recommendations = []string{}
		}
		if detail.Summary.TopAlarms == nil {
			detail.Summary.TopAlarms = []InspectionAlarmItem{}
		}
	}
	if strings.TrimSpace(r.HostsJSON) != "" {
		if err := json.Unmarshal([]byte(r.HostsJSON), &detail.Hosts); err != nil {
			logger.Log.Warnf("解析巡检报告主机 JSON 失败 (id=%d): %v", r.ID, err)
		}
		if detail.Hosts == nil {
			detail.Hosts = []InspectionHostItem{}
		}
	}
	return detail, nil
}

func removeInspectionReportFiles(r *InspectionReport) {
	if r == nil {
		return
	}
	for _, name := range []string{r.Files, r.ExcelFile, r.HtmlFile, r.PdfFile} {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		clean := filepath.Base(name)
		if clean == "." || clean == ".." {
			continue
		}
		_ = os.Remove(filepath.Join(DownloadPath, clean))
	}
}

// DeleteInspectionReport 删除指定巡检报告及其本地文件
func DeleteInspectionReport(id int) error {
	if DB == nil {
		return errors.New("database not initialized")
	}
	var r InspectionReport
	if err := DB.Where("id = ?", id).First(&r).Error; err != nil {
		return err
	}
	removeInspectionReportFiles(&r)
	if err := DB.Where("id = ?", id).Delete(&InspectionReport{}).Error; err != nil {
		return err
	}
	_ = DB.Model(&InspectionTask{}).Where("last_report_id = ?", id).Update("last_report_id", 0).Error
	return nil
}

// GetInspectionOverview 获取巡检中心统计概览
func GetInspectionOverview() (*InspectionOverview, error) {
	if DB == nil {
		return nil, errors.New("database not initialized")
	}
	overview := &InspectionOverview{
		GradeCounts: map[string]int{
			"excellent": 0,
			"good":      0,
			"warning":   0,
			"critical":  0,
		},
		RecentReports: []InspectionReport{},
	}

	if err := DB.Model(&InspectionTask{}).Count(&overview.TotalTasks).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&InspectionTask{}).Where("status = ?", "1").Count(&overview.EnabledTasks).Error; err != nil {
		return nil, err
	}
	if err := DB.Model(&InspectionReport{}).Where("status = ?", Success).Count(&overview.TotalReports).Error; err != nil {
		return nil, err
	}

	type gradeRow struct {
		HealthGrade string `gorm:"column:health_grade"`
		Cnt         int    `gorm:"column:cnt"`
	}
	var rows []gradeRow
	if err := DB.Model(&InspectionReport{}).
		Select("health_grade, COUNT(*) AS cnt").
		Where("status = ?", Success).
		Group("health_grade").Find(&rows).Error; err == nil {
		for _, row := range rows {
			if row.HealthGrade != "" {
				overview.GradeCounts[row.HealthGrade] = row.Cnt
			}
		}
	}

	var recent []InspectionReport
	if err := DB.Model(&InspectionReport{}).Omit("hosts_json").Order("id DESC").Limit(10).Find(&recent).Error; err == nil && recent != nil {
		overview.RecentReports = recent
	}

	var tasks []InspectionTask
	_ = DB.Where("last_report_id > 0").Find(&tasks).Error
	scoreSum := 0
	scoreCount := 0
	for _, t := range tasks {
		scoreSum += t.LastScore
		scoreCount++
	}
	if scoreCount == 0 {
		for _, r := range overview.RecentReports {
			if r.Status == Success {
				scoreSum += r.HealthScore
				scoreCount++
			}
		}
	}
	if scoreCount > 0 {
		overview.LatestAvgScore = int(math.Round(float64(scoreSum) / float64(scoreCount)))
		overview.LatestGrade = healthGradeFromScore(overview.LatestAvgScore)
	}
	return overview, nil
}

// ResetStaleRunningInspections 启动恢复或清理超时的运行中巡检任务与报告
func ResetStaleRunningInspections(timeout time.Duration) error {
	if DB == nil {
		return nil
	}
	now := time.Now()
	taskQuery := DB.Model(&InspectionTask{}).Where("exec_status = ?", Running)
	reportQuery := DB.Model(&InspectionReport{}).Where("status = ?", Running)
	if timeout > 0 {
		cutoff := now.Add(-timeout)
		taskQuery = taskQuery.Where("last_run_at IS NULL OR last_run_at < ?", cutoff)
		reportQuery = reportQuery.Where("start_time < ?", cutoff)
	}
	if err := taskQuery.Updates(map[string]interface{}{
		"exec_status":    Failed,
		"last_finish_at": now,
	}).Error; err != nil {
		return err
	}
	return reportQuery.Updates(map[string]interface{}{
		"status":    Failed,
		"error_msg": "巡检任务异常中断，已自动标记为失败",
		"end_time":  now,
	}).Error
}

// ========== 指标解析与健康度评分 ==========

func parseMetricPercent(raw string) (float64, bool) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return 0, false
	}
	s = strings.TrimSpace(strings.TrimSuffix(s, "%"))
	s = strings.ReplaceAll(s, ",", "")
	val, err := strconv.ParseFloat(s, 64)
	if err != nil || math.IsNaN(val) || math.IsInf(val, 0) || val < 0 {
		return 0, false
	}
	if val > 100 {
		val = 100
	}
	return math.Round(val*100) / 100, true
}

func extractMountFromFSKey(key, name string) string {
	start := strings.Index(key, "[")
	end := strings.LastIndex(key, "]")
	if start >= 0 && end > start+1 {
		inner := key[start+1 : end]
		if comma := strings.Index(inner, ","); comma >= 0 {
			mount := strings.TrimSpace(inner[:comma])
			if mount != "" {
				return mount
			}
		}
		if trimmed := strings.TrimSpace(inner); trimmed != "" {
			return trimmed
		}
	}
	return strings.TrimSpace(name)
}

func extractDiskUsageFromItems(items []Item) (float64, string, bool) {
	var maxPct float64
	var maxMount string
	found := false

	for _, it := range items {
		k := strings.ToLower(strings.TrimSpace(it.Key))
		if !strings.Contains(k, "vfs.fs.") || strings.Contains(k, "inode") {
			continue
		}
		var pct float64
		var ok bool
		switch {
		case strings.Contains(k, ",pused]"):
			pct, ok = parseMetricPercent(it.Lastvalue)
		case strings.Contains(k, ",pfree]"):
			var freePct float64
			freePct, ok = parseMetricPercent(it.Lastvalue)
			if ok {
				pct = math.Round((100.0-freePct)*100) / 100
			}
		default:
			continue
		}
		if !ok {
			continue
		}
		if !found || pct > maxPct {
			found = true
			maxPct = pct
			maxMount = extractMountFromFSKey(it.Key, it.Name)
		}
	}
	return maxPct, maxMount, found
}

func extractCPUMemFromItems(items []Item, existingCPU, existingMem float64, hasCPU, hasMem bool) (float64, bool, float64, bool) {
	cpu, gotCPU := existingCPU, hasCPU
	mem, gotMem := existingMem, hasMem

	var fallbackCPU float64
	hasFallbackCPU := false

	var memTotal, memAvail, memUsed, memFree float64
	hasTotal, hasAvail, hasUsed, hasFree := false, false, false, false

	for _, it := range items {
		k := strings.ToLower(strings.TrimSpace(it.Key))
		if !gotCPU && strings.HasPrefix(k, "system.cpu.util") {
			val, ok := parseMetricPercent(it.Lastvalue)
			if ok {
				switch {
				case strings.Contains(k, ",idle"):
					cpu = math.Round((100.0-val)*100) / 100
					gotCPU = true
				case k == "system.cpu.util" || k == "system.cpu.util[]" || strings.HasPrefix(k, "system.cpu.util[,") || strings.HasPrefix(k, "system.cpu.util[all"):
					// 若是 user/system/iowait 单独项，仅作候选；整体 utilization 直接采纳
					if !strings.Contains(k, ",user") && !strings.Contains(k, ",system") && !strings.Contains(k, ",iowait") && !strings.Contains(k, ",nice") {
						cpu = val
						gotCPU = true
					} else if !hasFallbackCPU {
						fallbackCPU = val
						hasFallbackCPU = true
					}
				default:
					if !hasFallbackCPU {
						fallbackCPU = val
						hasFallbackCPU = true
					}
				}
			}
		}

		if !gotMem && strings.HasPrefix(k, "vm.memory.") {
			switch {
			case strings.HasPrefix(k, "vm.memory.utilization"):
				if val, ok := parseMetricPercent(it.Lastvalue); ok {
					mem = val
					gotMem = true
				}
			case k == "vm.memory.size[pused]":
				if val, ok := parseMetricPercent(it.Lastvalue); ok {
					mem = val
					gotMem = true
				}
			case k == "vm.memory.size[pavailable]":
				if val, ok := parseMetricPercent(it.Lastvalue); ok {
					mem = math.Round((100.0-val)*100) / 100
					gotMem = true
				}
			case k == "vm.memory.size[total]":
				if v, err := strconv.ParseFloat(strings.TrimSpace(it.Lastvalue), 64); err == nil && v > 0 {
					memTotal = v
					hasTotal = true
				}
			case k == "vm.memory.size[available]":
				if v, err := strconv.ParseFloat(strings.TrimSpace(it.Lastvalue), 64); err == nil && v >= 0 {
					memAvail = v
					hasAvail = true
				}
			case k == "vm.memory.size[used]":
				if v, err := strconv.ParseFloat(strings.TrimSpace(it.Lastvalue), 64); err == nil && v >= 0 {
					memUsed = v
					hasUsed = true
				}
			case k == "vm.memory.size[free]":
				if v, err := strconv.ParseFloat(strings.TrimSpace(it.Lastvalue), 64); err == nil && v >= 0 {
					memFree = v
					hasFree = true
				}
			}
		}
	}

	if !gotCPU && hasFallbackCPU {
		cpu = fallbackCPU
		gotCPU = true
	}
	if !gotMem && hasTotal && memTotal > 0 {
		switch {
		case hasUsed:
			mem = math.Round((memUsed/memTotal*100.0)*100) / 100
			gotMem = true
		case hasAvail:
			mem = math.Round(((memTotal-memAvail)/memTotal*100.0)*100) / 100
			gotMem = true
		case hasFree:
			mem = math.Round(((memTotal-memFree)/memTotal*100.0)*100) / 100
			gotMem = true
		}
		if mem < 0 {
			mem = 0
		} else if mem > 100 {
			mem = 100
		}
	}
	return cpu, gotCPU, mem, gotMem
}

func extractPingFromItems(items []Item) string {
	for _, it := range items {
		k := strings.ToLower(strings.TrimSpace(it.Key))
		if k == "icmpping" || strings.HasPrefix(k, "icmpping[") {
			return strings.TrimSpace(it.Lastvalue)
		}
	}
	return ""
}

func healthGradeFromScore(score int) string {
	switch {
	case score >= 90:
		return "excellent"
	case score >= 75:
		return "good"
	case score >= 60:
		return "warning"
	default:
		return "critical"
	}
}

func healthGradeLabel(grade string) string {
	switch grade {
	case "excellent":
		return "优秀"
	case "good":
		return "良好"
	case "warning":
		return "亚健康"
	case "critical":
		return "高危"
	default:
		return "未知"
	}
}

func severityToNumAndLabel(raw string) (int, string) {
	s := strings.TrimSpace(raw)
	switch strings.ToLower(s) {
	case "5", "disaster", "灾难":
		return 5, "灾难"
	case "4", "high", "严重":
		return 4, "严重"
	case "3", "average", "一般严重", "一般":
		return 3, "一般严重"
	case "2", "warning", "警告":
		return 2, "警告"
	case "1", "information", "信息":
		return 1, "信息"
	case "0", "not classified", "未分类":
		return 0, "未分类"
	default:
		return 0, s
	}
}

func evaluateHostHealth(h *InspectionHostItem, cpuThreshold, memThreshold, diskThreshold float64) {
	h.Issues = make([]string, 0, 5)
	h.CpuHigh = h.HasCpu && h.CpuPct >= cpuThreshold
	h.MemHigh = h.HasMem && h.MemPct >= memThreshold
	h.DiskHigh = h.HasDisk && h.DiskPct >= diskThreshold

	isOffline := strings.TrimSpace(h.Available) == "2"
	if isOffline {
		h.Issues = append(h.Issues, "采集失联/不可用")
	}
	if h.CpuHigh {
		h.Issues = append(h.Issues, fmt.Sprintf("CPU使用率%.1f%%(≥%.0f%%)", h.CpuPct, cpuThreshold))
	}
	if h.MemHigh {
		h.Issues = append(h.Issues, fmt.Sprintf("内存使用率%.1f%%(≥%.0f%%)", h.MemPct, memThreshold))
	}
	if h.DiskHigh {
		mount := h.DiskMount
		if mount == "" {
			mount = "/"
		}
		h.Issues = append(h.Issues, fmt.Sprintf("磁盘[%s]使用率%.1f%%(≥%.0f%%)", mount, h.DiskPct, diskThreshold))
	}
	if h.CriticalAlarms > 0 {
		h.Issues = append(h.Issues, fmt.Sprintf("%d条高危告警", h.CriticalAlarms))
	} else if h.ActiveAlarms > 0 {
		h.Issues = append(h.Issues, fmt.Sprintf("%d条未恢复告警", h.ActiveAlarms))
	}

	switch {
	case isOffline:
		h.Status = "offline"
	case h.CriticalAlarms > 0 || (h.HasCpu && h.CpuPct >= 95) || (h.HasMem && h.MemPct >= 95) || (h.HasDisk && h.DiskPct >= 95):
		h.Status = "critical"
	case len(h.Issues) > 0:
		h.Status = "warning"
	default:
		h.Status = "normal"
	}

	score := 100
	if isOffline {
		score -= 40
	}
	if h.CpuHigh {
		if h.CpuPct >= 95 {
			score -= 25
		} else {
			score -= 15
		}
	}
	if h.MemHigh {
		if h.MemPct >= 95 {
			score -= 25
		} else {
			score -= 15
		}
	}
	if h.DiskHigh {
		if h.DiskPct >= 95 {
			score -= 25
		} else {
			score -= 15
		}
	}
	critDeduct := h.CriticalAlarms * 15
	if critDeduct > 30 {
		critDeduct = 30
	}
	otherAlarms := h.ActiveAlarms - h.CriticalAlarms
	if otherAlarms < 0 {
		otherAlarms = 0
	}
	otherDeduct := otherAlarms * 5
	if otherDeduct > 15 {
		otherDeduct = 15
	}
	score -= critDeduct + otherDeduct
	if score < 0 {
		score = 0
	}
	h.HostScore = score
}

// ComputeInspectionScore 基于主机巡检结果和活跃告警计算四维健康评分、等级及建议
func ComputeInspectionScore(hosts []InspectionHostItem, alarms []InspectionAlarmItem, cpuThreshold, memThreshold, diskThreshold float64) (int, string, InspectionScoreBreakdown, []string) {
	if cpuThreshold <= 0 {
		cpuThreshold = defaultInspectCpuThreshold
	}
	if memThreshold <= 0 {
		memThreshold = defaultInspectMemThreshold
	}
	if diskThreshold <= 0 {
		diskThreshold = defaultInspectDiskThreshold
	}

	breakdown := InspectionScoreBreakdown{
		AvailabilityMax: 35,
		ResourceMax:     30,
		AlarmMax:        25,
		StabilityMax:    10,
	}

	totalHosts := len(hosts)
	if totalHosts == 0 && len(alarms) == 0 {
		breakdown.AvailabilityScore = breakdown.AvailabilityMax
		breakdown.ResourceScore = breakdown.ResourceMax
		breakdown.AlarmScore = breakdown.AlarmMax
		breakdown.StabilityScore = breakdown.StabilityMax
		return 100, "excellent", breakdown, []string{"当前巡检范围内暂无纳管设备，建议检查实例启用状态、设备分组或资产类型筛选条件。"}
	}

	var offlineNames, cpuHighNames, memHighNames, diskHighNames []string
	var resourceWeighted float64
	var abnormalCount, multiIssueCount int

	for i := range hosts {
		evaluateHostHealth(&hosts[i], cpuThreshold, memThreshold, diskThreshold)
		h := hosts[i]
		if h.Status == "offline" {
			offlineNames = append(offlineNames, h.Name)
		}
		if h.CpuHigh {
			cpuHighNames = append(cpuHighNames, fmt.Sprintf("%s(%.1f%%)", h.Name, h.CpuPct))
			resourceWeighted += 1.0
		}
		if h.MemHigh {
			memHighNames = append(memHighNames, fmt.Sprintf("%s(%.1f%%)", h.Name, h.MemPct))
			resourceWeighted += 1.0
		}
		if h.DiskHigh {
			mount := h.DiskMount
			if mount == "" {
				mount = "/"
			}
			diskHighNames = append(diskHighNames, fmt.Sprintf("%s[%s](%.1f%%)", h.Name, mount, h.DiskPct))
			resourceWeighted += 1.2
		}
		if h.Status != "normal" {
			abnormalCount++
		}
		if len(h.Issues) >= 2 {
			multiIssueCount++
		}
	}

	denom := float64(totalHosts)
	if denom < 1 {
		denom = 1
	}

	// 1. 可用性健康度 (35分)
	availDeduct := 0
	if len(offlineNames) > 0 {
		ratio := float64(len(offlineNames)) / denom
		availDeduct = int(math.Round(ratio * 70.0))
		if availDeduct < 5 {
			availDeduct = 5
		}
		if availDeduct > breakdown.AvailabilityMax {
			availDeduct = breakdown.AvailabilityMax
		}
	}
	breakdown.AvailabilityDeduct = availDeduct
	breakdown.AvailabilityScore = breakdown.AvailabilityMax - availDeduct

	// 2. 资源水位健康度 (30分)
	resDeduct := 0
	if resourceWeighted > 0 {
		resDeduct = int(math.Round((resourceWeighted / denom) * 20.0))
		if resDeduct < 2 {
			resDeduct = 2
		}
		if resDeduct > breakdown.ResourceMax {
			resDeduct = breakdown.ResourceMax
		}
	}
	breakdown.ResourceDeduct = resDeduct
	breakdown.ResourceScore = breakdown.ResourceMax - resDeduct

	// 3. 告警风险健康度 (25分)
	activeCount := len(alarms)
	criticalCount := 0
	for _, a := range alarms {
		if a.SeverityNum >= 4 {
			criticalCount++
		}
	}
	alarmDeduct := 0
	if activeCount > 0 {
		raw := float64(criticalCount*5 + (activeCount-criticalCount)*2)
		if denom > 20 {
			raw = raw * (20.0 / denom)
		}
		alarmDeduct = int(math.Round(raw))
		if criticalCount > 0 && alarmDeduct < 3 {
			alarmDeduct = 3
		} else if alarmDeduct < 1 {
			alarmDeduct = 1
		}
		if alarmDeduct > breakdown.AlarmMax {
			alarmDeduct = breakdown.AlarmMax
		}
	}
	breakdown.AlarmDeduct = alarmDeduct
	breakdown.AlarmScore = breakdown.AlarmMax - alarmDeduct

	// 4. 集群稳定度 (10分)
	stabDeduct := 0
	if abnormalCount > 0 {
		stabDeduct = int(math.Round((float64(abnormalCount)/denom)*10.0)) + multiIssueCount
		if stabDeduct < 1 {
			stabDeduct = 1
		}
		if stabDeduct > breakdown.StabilityMax {
			stabDeduct = breakdown.StabilityMax
		}
	}
	breakdown.StabilityDeduct = stabDeduct
	breakdown.StabilityScore = breakdown.StabilityMax - stabDeduct

	totalScore := breakdown.AvailabilityScore + breakdown.ResourceScore + breakdown.AlarmScore + breakdown.StabilityScore
	if totalScore < 0 {
		totalScore = 0
	} else if totalScore > 100 {
		totalScore = 100
	}
	grade := healthGradeFromScore(totalScore)

	recs := make([]string, 0, 5)
	if len(offlineNames) > 0 {
		recs = append(recs, fmt.Sprintf("共有 %d 台设备采集失联/不可用（%s），请优先排查 Zabbix Agent/SNMP 服务状态与网络连通性。",
			len(offlineNames), joinTopSample(offlineNames, 3)))
	}
	if len(diskHighNames) > 0 {
		recs = append(recs, fmt.Sprintf("共有 %d 台设备磁盘分区超过 %.0f%% 阈值（%s），建议及时清理过期日志、大文件或扩容存储。",
			len(diskHighNames), diskThreshold, joinTopSample(diskHighNames, 3)))
	}
	if len(cpuHighNames) > 0 {
		recs = append(recs, fmt.Sprintf("共有 %d 台设备 CPU 利用率超过 %.0f%% 阈值（%s），建议排查高负载进程及突发流量。",
			len(cpuHighNames), cpuThreshold, joinTopSample(cpuHighNames, 3)))
	}
	if len(memHighNames) > 0 {
		recs = append(recs, fmt.Sprintf("共有 %d 台设备内存利用率超过 %.0f%% 阈值（%s），建议检查 JVM/数据库内存配置及是否存在内存泄漏。",
			len(memHighNames), memThreshold, joinTopSample(memHighNames, 3)))
	}
	if criticalCount > 0 {
		recs = append(recs, fmt.Sprintf("当前仍有 %d 条高危未恢复告警（全部活跃告警共 %d 条），请安排值班人员及时认领与处置。",
			criticalCount, activeCount))
	} else if activeCount > 0 {
		recs = append(recs, fmt.Sprintf("当前存在 %d 条未恢复一般告警，建议结合告警中心核查是否需要收敛或处理。", activeCount))
	}
	if len(recs) == 0 {
		recs = append(recs, "本次巡检整体运行健康，所有纳管主机可用性及核心资源水位均处于安全阈值范围内，建议保持例行巡检节奏。")
	}

	return totalScore, grade, breakdown, recs
}

func joinTopSample(items []string, max int) string {
	if len(items) <= max {
		return strings.Join(items, "、")
	}
	return strings.Join(items[:max], "、") + fmt.Sprintf(" 等%d项", len(items))
}

// ========== Zabbix 数据采集与巡检执行引擎 ==========

type inspectRawHost struct {
	ListHost
	Groups []struct {
		GroupID string `json:"groupid"`
		Name    string `json:"name"`
	} `json:"groups"`
}

func resolveInspectionTargetInstances(zidsCSV string) ([]ZabbixInstance, error) {
	if DB == nil {
		return nil, errors.New("database not initialized")
	}
	var all []ZabbixInstance
	if err := DB.Where("enabled = ?", true).Order("id asc").Find(&all).Error; err != nil {
		return nil, err
	}
	trimmed := strings.TrimSpace(zidsCSV)
	if trimmed == "" || trimmed == "*" {
		return all, nil
	}
	wanted := make(map[string]bool)
	for _, token := range splitInspectionCSV(trimmed) {
		if token == "*" {
			return all, nil
		}
		wanted[token] = true
	}
	var filtered []ZabbixInstance
	for _, inst := range all {
		if wanted[strconv.Itoa(inst.ID)] || wanted[inst.Instance] {
			filtered = append(filtered, inst)
		}
	}
	return filtered, nil
}

func loadAssetTypeNameMap() map[string]string {
	res := make(map[string]string)
	if DB == nil {
		return res
	}
	types, err := GetAllAssetTypes()
	if err != nil {
		return res
	}
	for _, at := range types {
		res[strings.ToUpper(at.TypeCode)] = at.Name
	}
	return res
}

func batchFetchInstanceHostMetrics(inst *APIInstance, hostIDs []string) map[string][]Item {
	result := make(map[string][]Item)
	if inst == nil || inst.API == nil || len(hostIDs) == 0 {
		return result
	}
	searchKeys := []string{
		"system.cpu.util",
		"vm.memory.utilization",
		"vm.memory.size",
		"vfs.fs.size",
		"vfs.fs.dependent.size",
		"icmpping",
	}
	const chunkSize = 100
	for start := 0; start < len(hostIDs); start += chunkSize {
		end := start + chunkSize
		if end > len(hostIDs) {
			end = len(hostIDs)
		}
		rep, err := inst.API.CallWithError("item.get", Params{
			"output":      []string{"itemid", "hostid", "name", "key_", "lastvalue", "units", "value_type"},
			"hostids":     hostIDs[start:end],
			"filter":      map[string]string{"status": "0"},
			"searchByAny": true,
			"search":      map[string][]string{"key_": searchKeys},
		})
		if err != nil {
			logger.Log.Warnf("巡检批量获取实例 %s 监控项失败: %v", inst.Name, err)
			continue
		}
		items, err := unmarshalItemsResult(rep.Result)
		if err != nil {
			continue
		}
		for _, it := range items {
			if strings.TrimSpace(it.Hostid) == "" {
				continue
			}
			result[it.Hostid] = append(result[it.Hostid], it)
		}
	}
	return result
}

func getActiveTriggersFromInstanceSafe(inst *APIInstance) []InspectionAlarmItem {
	if inst == nil || inst.API == nil {
		return nil
	}
	rep, err := inst.API.CallWithError("trigger.get", Params{
		"output":          "extend",
		"sortfield":       "lastchange",
		"sortorder":       "DESC",
		"selectHosts":     []string{"hostid", "name", "host"},
		"selectLastEvent": "extend",
		"filter":          map[string]string{"value": "1"},
		"maintenance":     false,
		"only_true":       true,
		"monitored":       true,
	})
	if err != nil {
		logger.Log.Warnf("巡检获取实例 %s 活跃触发器失败: %v", inst.Name, err)
		return nil
	}
	raw, err := json.Marshal(rep.Result)
	if err != nil {
		return nil
	}
	var triggers []LastTriggers
	if err := json.Unmarshal(raw, &triggers); err != nil {
		return nil
	}
	out := make([]InspectionAlarmItem, 0, len(triggers))
	for _, tr := range triggers {
		if len(tr.Hosts) == 0 {
			continue
		}
		hostID := tr.Hosts[0].Hostid
		hostname := tr.Hosts[0].Name
		if hostname == "" {
			hostname = tr.Description
		}
		sevStr := tr.LastEvent.Severity
		if sevStr == "" {
			sevStr = tr.Priority
		}
		sevNum, sevLabel := severityToNumAndLabel(sevStr)
		msg := strings.TrimSpace(tr.LastEvent.Name)
		if msg == "" {
			msg = strings.TrimSpace(tr.Description)
		}
		clockStr := tr.LastEvent.Clock
		if clockStr == "" {
			clockStr = tr.Lastchange
		}
		occurTime := ""
		if ts, parseErr := strconv.ParseInt(clockStr, 10, 64); parseErr == nil && ts > 0 {
			occurTime = time.Unix(ts, 0).Format("2006-01-02 15:04:05")
		}
		out = append(out, InspectionAlarmItem{
			ZID:          inst.ZID,
			InstanceName: inst.Name,
			HostID:       hostID,
			Host:         hostname,
			Hostname:     hostname,
			Level:        sevLabel,
			SeverityNum:  sevNum,
			Message:      msg,
			OccurTime:    occurTime,
			EventID:      tr.LastEvent.Eventid,
			Acknowledged: tr.LastEvent.Acknowledged == "1",
		})
	}
	return out
}

func loadLocalUnclosedAlarms(zid int) []InspectionAlarmItem {
	if DB == nil {
		return nil
	}
	since := time.Now().Add(-7 * 24 * time.Hour)
	var rows []Alarm
	q := DB.Model(&Alarm{}).
		Where("escalation_step >= 0 AND (status = ? OR status = ?) AND occurtime >= ?", "1", "故障", since)
	if zid > 0 {
		q = q.Where("zid = ?", zid)
	}
	if err := q.Order("occurtime DESC").Limit(500).Find(&rows).Error; err != nil {
		return nil
	}
	out := make([]InspectionAlarmItem, 0, len(rows))
	for _, a := range rows {
		sevNum, sevLabel := severityToNumAndLabel(a.Level)
		hName := strings.TrimSpace(a.Hostname)
		if hName == "" {
			hName = strings.TrimSpace(a.Host)
		}
		out = append(out, InspectionAlarmItem{
			ZID:          a.ZID,
			HostID:       a.HostID,
			Host:         a.Host,
			Hostname:     hName,
			HostIP:       a.HostsIP,
			Level:        sevLabel,
			SeverityNum:  sevNum,
			Message:      a.Message,
			ItemName:     a.ItemName,
			OccurTime:    a.OccurTime.Format("2006-01-02 15:04:05"),
			EventID:      strconv.FormatInt(a.EventID, 10),
			Acknowledged: a.AckStatus == 1,
		})
	}
	return out
}

func collectInstanceInspectionData(instCfg ZabbixInstance, allowedTypes map[string]bool, groupIDs []string, typeNameMap map[string]string) (InspectionInstanceStat, []InspectionHostItem, []InspectionAlarmItem) {
	stat := InspectionInstanceStat{
		ZID:          instCfg.ID,
		InstanceName: instCfg.Name,
	}

	inst, err := GetAPIByZID(instCfg.ID)
	if err != nil || inst == nil || inst.API == nil {
		stat.Reachable = false
		if err != nil {
			stat.Error = err.Error()
		} else {
			stat.Error = "实例 API 不可用"
		}
		// 连接失败时尝试回退读取本地缓存及本地告警
		fallbackHosts := loadFallbackHostsFromCache(instCfg, allowedTypes, typeNameMap)
		localAlarms := loadLocalUnclosedAlarms(instCfg.ID)
		for i := range localAlarms {
			localAlarms[i].InstanceName = instCfg.Name
		}
		stat.TotalHosts = len(fallbackHosts)
		stat.ActiveAlarms = len(localAlarms)
		return stat, fallbackHosts, localAlarms
	}

	stat.Reachable = true
	params := Params{
		"output":           "extend",
		"filter":           map[string]string{"status": "0"},
		"selectInventory":  "extend",
		"selectInterfaces": []string{"ip", "port", "available", "error"},
	}
	if len(groupIDs) > 0 {
		params["groupids"] = groupIDs
	}
	rep, err := inst.API.CallWithError("host.get", params)
	if err != nil {
		stat.Reachable = false
		stat.Error = err.Error()
		fallbackHosts := loadFallbackHostsFromCache(instCfg, allowedTypes, typeNameMap)
		localAlarms := loadLocalUnclosedAlarms(instCfg.ID)
		for i := range localAlarms {
			localAlarms[i].InstanceName = instCfg.Name
		}
		stat.TotalHosts = len(fallbackHosts)
		stat.ActiveAlarms = len(localAlarms)
		return stat, fallbackHosts, localAlarms
	}

	rawBytes, err := json.Marshal(rep.Result)
	if err != nil {
		stat.Error = err.Error()
		return stat, nil, nil
	}
	var rawHosts []inspectRawHost
	if err := json.Unmarshal(rawBytes, &rawHosts); err != nil {
		stat.Error = err.Error()
		return stat, nil, nil
	}

	matchedRaw := make([]inspectRawHost, 0, len(rawHosts))
	hostIDs := make([]string, 0, len(rawHosts))
	for _, rh := range rawHosts {
		tc := strings.ToUpper(strings.TrimSpace(rh.Inventory.Type))
		if len(allowedTypes) > 0 && !allowedTypes[tc] {
			continue
		}
		matchedRaw = append(matchedRaw, rh)
		hostIDs = append(hostIDs, rh.Hostid)
	}

	itemsByHost := batchFetchInstanceHostMetrics(inst, hostIDs)
	triggerAlarms := getActiveTriggersFromInstanceSafe(inst)
	localAlarms := loadLocalUnclosedAlarms(inst.ZID)

	// 合并活跃告警，按 hostID + eventid/message 去重
	alarmMap := make(map[string]InspectionAlarmItem)
	allAlarms := make([]InspectionAlarmItem, 0, len(triggerAlarms)+len(localAlarms))
	for _, a := range triggerAlarms {
		key := fmt.Sprintf("%d|%s|%s", inst.ZID, a.EventID, a.HostID)
		if a.EventID == "" || a.EventID == "0" {
			key = fmt.Sprintf("%d|%s|%s", inst.ZID, a.HostID, a.Message)
		}
		alarmMap[key] = a
		allAlarms = append(allAlarms, a)
	}
	for _, a := range localAlarms {
		a.InstanceName = inst.Name
		key := fmt.Sprintf("%d|%s|%s", inst.ZID, a.EventID, a.HostID)
		if a.EventID == "" || a.EventID == "0" {
			key = fmt.Sprintf("%d|%s|%s", inst.ZID, a.HostID, a.Message)
		}
		if _, exists := alarmMap[key]; !exists {
			alarmMap[key] = a
			allAlarms = append(allAlarms, a)
		}
	}

	// 按主机索引告警数
	type alarmCounter struct {
		total    int
		critical int
		ip       string
	}
	hostAlarmByID := make(map[string]*alarmCounter)
	hostAlarmByName := make(map[string]*alarmCounter)
	for i := range allAlarms {
		a := &allAlarms[i]
		if a.HostID != "" {
			c := hostAlarmByID[a.HostID]
			if c == nil {
				c = &alarmCounter{}
				hostAlarmByID[a.HostID] = c
			}
			c.total++
			if a.SeverityNum >= 4 {
				c.critical++
			}
		} else if a.Host != "" {
			c := hostAlarmByName[a.Host]
			if c == nil {
				c = &alarmCounter{}
				hostAlarmByName[a.Host] = c
			}
			c.total++
			if a.SeverityNum >= 4 {
				c.critical++
			}
		}
	}

	hosts := make([]InspectionHostItem, 0, len(matchedRaw))
	for _, rh := range matchedRaw {
		tc := strings.TrimSpace(rh.Inventory.Type)
		tcUpper := strings.ToUpper(tc)
		if tc == "" {
			tc = "UNCLASSIFIED"
			tcUpper = "UNCLASSIFIED"
		}
		monitorType := GetMonitorType(tc)
		hSummary := buildHostFromListHost(inst, rh.ListHost, tc, monitorType)
		typeName := typeNameMap[tcUpper]
		if typeName == "" {
			if tc == "UNCLASSIFIED" {
				typeName = "未分类设备"
			} else {
				typeName = tc
			}
		}

		cpuVal, hasCPU := parseMetricPercent(hSummary.CPUUtilization)
		memVal, hasMem := parseMetricPercent(hSummary.MemoryUtilization)
		hostItems := itemsByHost[rh.Hostid]
		cpuVal, hasCPU, memVal, hasMem = extractCPUMemFromItems(hostItems, cpuVal, memVal, hasCPU, hasMem)
		diskVal, diskMount, hasDisk := extractDiskUsageFromItems(hostItems)
		pingVal := strings.TrimSpace(hSummary.Ping)
		if pingVal == "" {
			pingVal = extractPingFromItems(hostItems)
		}

		activeCnt, critCnt := 0, 0
		if c, ok := hostAlarmByID[rh.Hostid]; ok {
			activeCnt += c.total
			critCnt += c.critical
		} else if c, ok := hostAlarmByName[rh.Host]; ok {
			activeCnt += c.total
			critCnt += c.critical
		} else if c, ok := hostAlarmByName[rh.Name]; ok {
			activeCnt += c.total
			critCnt += c.critical
		}

		// 给关联告警回填主机 IP
		for i := range allAlarms {
			if allAlarms[i].HostIP == "" && ((allAlarms[i].HostID != "" && allAlarms[i].HostID == rh.Hostid) || allAlarms[i].Host == rh.Host || allAlarms[i].Hostname == rh.Name) {
				allAlarms[i].HostIP = hSummary.Interfaces
			}
		}

		displayName := strings.TrimSpace(hSummary.Name)
		if displayName == "" {
			displayName = hSummary.Host
		}
		hosts = append(hosts, InspectionHostItem{
			ZID:            inst.ZID,
			InstanceName:   inst.Name,
			HostID:         hSummary.HostID,
			Host:           hSummary.Host,
			Name:           displayName,
			IP:             hSummary.Interfaces,
			TypeCode:       tc,
			TypeName:       typeName,
			OS:             hSummary.OS,
			Available:      hSummary.Available,
			Ping:           pingVal,
			CpuPct:         cpuVal,
			HasCpu:         hasCPU,
			MemPct:         memVal,
			HasMem:         hasMem,
			DiskPct:        diskVal,
			HasDisk:        hasDisk,
			DiskMount:      diskMount,
			Uptime:         hSummary.Uptime,
			ActiveAlarms:   activeCnt,
			CriticalAlarms: critCnt,
		})
	}

	stat.TotalHosts = len(hosts)
	stat.ActiveAlarms = len(allAlarms)
	return stat, hosts, allAlarms
}

func loadFallbackHostsFromCache(instCfg ZabbixInstance, allowedTypes map[string]bool, typeNameMap map[string]string) []InspectionHostItem {
	assetTypes, err := GetAllAssetTypes()
	if err != nil {
		return nil
	}
	var out []InspectionHostItem
	for _, at := range assetTypes {
		tcUpper := strings.ToUpper(at.TypeCode)
		if len(allowedTypes) > 0 && !allowedTypes[tcUpper] {
			continue
		}
		cached, cErr := CacheGet(at.TypeCode + "_OVERVIEW")
		if cErr != nil || cached == "" {
			continue
		}
		var hs []Hosts
		if err := json.Unmarshal([]byte(cached), &hs); err != nil {
			continue
		}
		for _, h := range hs {
			if h.ZID != instCfg.ID {
				continue
			}
			cpuVal, hasCPU := parseMetricPercent(h.CPUUtilization)
			memVal, hasMem := parseMetricPercent(h.MemoryUtilization)
			tName := typeNameMap[tcUpper]
			if tName == "" {
				tName = at.Name
			}
			displayName := strings.TrimSpace(h.Name)
			if displayName == "" {
				displayName = h.Host
			}
			out = append(out, InspectionHostItem{
				ZID:          instCfg.ID,
				InstanceName: instCfg.Name,
				HostID:       h.HostID,
				Host:         h.Host,
				Name:         displayName,
				IP:           h.Interfaces,
				TypeCode:     at.TypeCode,
				TypeName:     tName,
				OS:           h.OS,
				Available:    h.Available,
				Ping:         h.Ping,
				CpuPct:       cpuVal,
				HasCpu:       hasCPU,
				MemPct:       memVal,
				HasMem:       hasMem,
				Uptime:       h.Uptime,
			})
		}
	}
	return out
}

// RunInspectionTask 同步执行一次完整的自动化巡检并保存报告与健康评分
func RunInspectionTask(taskID int, triggerType string) (*InspectionReport, error) {
	if DB == nil {
		return nil, errors.New("database not initialized")
	}
	if !markInspectionTaskRunning(taskID) {
		return nil, errors.New("该巡检任务正在执行中，请勿重复触发")
	}
	defer unmarkInspectionTaskRunning(taskID)

	_ = ResetStaleRunningInspections(15 * time.Minute)
	task, err := GetInspectionTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	if task.ExecStatus == Running {
		return nil, errors.New("该巡检任务正在执行中，请稍候")
	}

	if triggerType == "" {
		triggerType = "manual"
	}
	startedAt := time.Now()
	report := &InspectionReport{
		TaskID:      task.ID,
		TaskName:    task.Name,
		TriggerType: triggerType,
		Status:      Running,
		HealthGrade: "excellent",
		StartTime:   startedAt,
	}
	if err := DB.Create(report).Error; err != nil {
		return nil, err
	}
	_ = DB.Model(&InspectionTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"exec_status": Running,
		"last_run_at": startedAt,
	}).Error

	failReport := func(runErr error) (*InspectionReport, error) {
		finished := time.Now()
		report.Status = Failed
		report.EndTime = &finished
		report.TotalTime = int64(finished.Sub(startedAt).Seconds())
		if runErr != nil {
			report.ErrorMsg = runErr.Error()
			if len(report.ErrorMsg) > 480 {
				report.ErrorMsg = report.ErrorMsg[:480]
			}
		}
		_ = DB.Model(&InspectionReport{}).Where("id = ?", report.ID).Updates(map[string]interface{}{
			"status":     report.Status,
			"error_msg":  report.ErrorMsg,
			"end_time":   report.EndTime,
			"total_time": report.TotalTime,
		}).Error
		_ = DB.Model(&InspectionTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
			"exec_status":    Failed,
			"last_finish_at": finished,
		}).Error
		return report, runErr
	}

	instances, err := resolveInspectionTargetInstances(task.ZIDs)
	if err != nil {
		return failReport(err)
	}

	allowedTypes := make(map[string]bool)
	if ht := strings.TrimSpace(task.HostTypes); ht != "" && ht != "*" {
		for _, code := range splitInspectionCSV(ht) {
			if code == "*" {
				allowedTypes = make(map[string]bool)
				break
			}
			allowedTypes[strings.ToUpper(code)] = true
		}
	}
	groupIDs := splitInspectionCSV(task.GroupIDs)
	typeNameMap := loadAssetTypeNameMap()

	allHosts := make([]InspectionHostItem, 0)
	allAlarms := make([]InspectionAlarmItem, 0)
	instanceStats := make([]InspectionInstanceStat, 0, len(instances))

	if len(instances) == 0 {
		// 即使尚未配置启用任何 Zabbix 实例，也检查本地未关闭告警并给出完整健康报告
		localAlarms := loadLocalUnclosedAlarms(0)
		allAlarms = append(allAlarms, localAlarms...)
	} else {
		var wg sync.WaitGroup
		var mu sync.Mutex
		for _, instCfg := range instances {
			wg.Add(1)
			go func(cfg ZabbixInstance) {
				defer wg.Done()
				stat, hosts, alarms := collectInstanceInspectionData(cfg, allowedTypes, groupIDs, typeNameMap)
				mu.Lock()
				instanceStats = append(instanceStats, stat)
				allHosts = append(allHosts, hosts...)
				allAlarms = append(allAlarms, alarms...)
				mu.Unlock()
			}(instCfg)
		}
		wg.Wait()
	}

	sort.Slice(instanceStats, func(i, j int) bool { return instanceStats[i].ZID < instanceStats[j].ZID })
	sort.Slice(allAlarms, func(i, j int) bool {
		if allAlarms[i].SeverityNum != allAlarms[j].SeverityNum {
			return allAlarms[i].SeverityNum > allAlarms[j].SeverityNum
		}
		return allAlarms[i].OccurTime > allAlarms[j].OccurTime
	})

	score, grade, breakdown, recs := ComputeInspectionScore(allHosts, allAlarms, task.CpuThreshold, task.MemThreshold, task.DiskThreshold)

	// 若存在不可达实例或无启用实例，补充提示
	for _, st := range instanceStats {
		if !st.Reachable {
			recs = append([]string{fmt.Sprintf("Zabbix 实例 [%s] 连接异常（%s），当前部分主机数据可能依赖缓存或缺失，请及时核查。", st.InstanceName, st.Error)}, recs...)
		}
	}
	if len(instances) == 0 {
		recs = append([]string{"当前尚未配置或启用任何 Zabbix 实例，请前往【系统管理 -> Zabbix配置】添加并启用实例以获取实时设备巡检数据。"}, recs...)
	}

	// 计算主机与维度汇总统计
	availHosts, unavailHosts, highCPU, highMem, highDisk, abnormalHosts := 0, 0, 0, 0, 0, 0
	typeStatMap := make(map[string]*InspectionTypeStat)
	instHostAgg := make(map[int]struct {
		avail     int
		unavail   int
		abnormal  int
		scoreSum  int
		hostCount int
	})

	for _, h := range allHosts {
		agg := instHostAgg[h.ZID]
		agg.hostCount++
		agg.scoreSum += h.HostScore
		if h.Available == "2" {
			unavailHosts++
			agg.unavail++
		} else {
			availHosts++
			agg.avail++
		}
		if h.CpuHigh {
			highCPU++
		}
		if h.MemHigh {
			highMem++
		}
		if h.DiskHigh {
			highDisk++
		}
		if h.Status != "normal" {
			abnormalHosts++
			agg.abnormal++
		}
		instHostAgg[h.ZID] = agg

		ts := typeStatMap[h.TypeCode]
		if ts == nil {
			ts = &InspectionTypeStat{
				TypeCode: h.TypeCode,
				TypeName: h.TypeName,
			}
			typeStatMap[h.TypeCode] = ts
		}
		ts.TotalHosts++
		if h.Available != "2" {
			ts.AvailableHosts++
		}
		if h.Status != "normal" {
			ts.AbnormalHosts++
		}
	}

	for i := range instanceStats {
		agg := instHostAgg[instanceStats[i].ZID]
		instanceStats[i].AvailableHosts = agg.avail
		instanceStats[i].UnavailableHosts = agg.unavail
		instanceStats[i].AbnormalHosts = agg.abnormal
		if agg.hostCount > 0 {
			instanceStats[i].AvgScore = int(math.Round(float64(agg.scoreSum) / float64(agg.hostCount)))
		} else if instanceStats[i].Reachable {
			instanceStats[i].AvgScore = 100
		}
	}

	typeStats := make([]InspectionTypeStat, 0, len(typeStatMap))
	for _, ts := range typeStatMap {
		typeStats = append(typeStats, *ts)
	}
	sort.Slice(typeStats, func(i, j int) bool { return typeStats[i].TotalHosts > typeStats[j].TotalHosts })

	// 异常主机优先排在明细前列，便于巡检阅读
	sort.Slice(allHosts, func(i, j int) bool {
		if allHosts[i].HostScore != allHosts[j].HostScore {
			return allHosts[i].HostScore < allHosts[j].HostScore
		}
		return allHosts[i].Name < allHosts[j].Name
	})

	critAlarms := 0
	for _, a := range allAlarms {
		if a.SeverityNum >= 4 {
			critAlarms++
		}
	}
	topAlarms := allAlarms
	if len(topAlarms) > 50 {
		topAlarms = allAlarms[:50]
	}

	summary := InspectionSummary{
		ScoreBreakdown:  breakdown,
		InstanceStats:   instanceStats,
		TypeStats:       typeStats,
		Recommendations: recs,
		TopAlarms:       topAlarms,
		CpuThreshold:    task.CpuThreshold,
		MemThreshold:    task.MemThreshold,
		DiskThreshold:   task.DiskThreshold,
	}

	summaryJSONBytes, _ := json.Marshal(summary)
	hostsJSONBytes, _ := json.Marshal(allHosts)

	report.TotalInstances = len(instanceStats)
	report.TotalHosts = len(allHosts)
	report.AvailableHosts = availHosts
	report.UnavailableHosts = unavailHosts
	report.HighCpuHosts = highCPU
	report.HighMemHosts = highMem
	report.HighDiskHosts = highDisk
	report.ActiveAlarms = len(allAlarms)
	report.CriticalAlarms = critAlarms
	report.AbnormalHosts = abnormalHosts
	report.HealthScore = score
	report.HealthGrade = grade
	report.SummaryJSON = string(summaryJSONBytes)
	report.HostsJSON = string(hostsJSONBytes)

	finished := time.Now()
	report.EndTime = &finished
	report.TotalTime = int64(finished.Sub(startedAt).Seconds())
	report.Status = Success

	// 生成多格式巡检报告（Excel / HTML / PDF / ZIP）
	if fileErr := generateInspectionReportFiles(task, report, &summary, allHosts); fileErr != nil {
		logger.Log.Warnf("生成巡检报告文件部分失败 (task=%s): %v", task.Name, fileErr)
	}

	if err := DB.Model(&InspectionReport{}).Where("id = ?", report.ID).Updates(map[string]interface{}{
		"status":            report.Status,
		"health_score":      report.HealthScore,
		"health_grade":      report.HealthGrade,
		"total_instances":   report.TotalInstances,
		"total_hosts":       report.TotalHosts,
		"available_hosts":   report.AvailableHosts,
		"unavailable_hosts": report.UnavailableHosts,
		"high_cpu_hosts":    report.HighCpuHosts,
		"high_mem_hosts":    report.HighMemHosts,
		"high_disk_hosts":   report.HighDiskHosts,
		"active_alarms":     report.ActiveAlarms,
		"critical_alarms":   report.CriticalAlarms,
		"abnormal_hosts":    report.AbnormalHosts,
		"summary_json":      report.SummaryJSON,
		"hosts_json":        report.HostsJSON,
		"files":             report.Files,
		"excel_file":        report.ExcelFile,
		"html_file":         report.HtmlFile,
		"pdf_file":          report.PdfFile,
		"end_time":          report.EndTime,
		"total_time":        report.TotalTime,
	}).Error; err != nil {
		return failReport(err)
	}

	_ = DB.Model(&InspectionTask{}).Where("id = ?", task.ID).Updates(map[string]interface{}{
		"exec_status":    Success,
		"last_score":     report.HealthScore,
		"last_grade":     report.HealthGrade,
		"last_report_id": report.ID,
		"last_finish_at": finished,
	}).Error

	NotifyInspectionReport(task, report, &summary)
	return report, nil
}

// TriggerInspectionTaskNow 触发一次立即巡检；若 2.5s 内完成则直接返回完整报告，否则返回运行中的报告
func TriggerInspectionTaskNow(taskID int) (*InspectionReport, error) {
	if DB == nil {
		return nil, errors.New("database not initialized")
	}
	task, err := GetInspectionTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	_ = ResetStaleRunningInspections(15 * time.Minute)
	task, err = GetInspectionTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	if task.ExecStatus == Running {
		return nil, errors.New("该巡检任务正在执行中，请稍候")
	}

	type runOutcome struct {
		report *InspectionReport
		err    error
	}
	ch := make(chan runOutcome, 1)
	go func() {
		r, rErr := RunInspectionTask(taskID, "manual")
		ch <- runOutcome{report: r, err: rErr}
	}()

	select {
	case out := <-ch:
		return out.report, out.err
	case <-time.After(2500 * time.Millisecond):
		var latest InspectionReport
		if qErr := DB.Where("task_id = ?", taskID).Order("id DESC").Omit("hosts_json").First(&latest).Error; qErr == nil {
			return &latest, nil
		}
		return &InspectionReport{
			TaskID:      taskID,
			TaskName:    task.Name,
			TriggerType: "manual",
			Status:      Running,
			StartTime:   time.Now(),
		}, nil
	}
}
