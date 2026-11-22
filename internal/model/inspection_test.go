package model

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestParseMetricPercentAndItemExtraction(t *testing.T) {
	cases := []struct {
		raw   string
		want  float64
		valid bool
	}{
		{"45.67%", 45.67, true},
		{" 82.125 % ", 82.13, true},
		{"105.4", 100, true},
		{"0", 0, true},
		{"-3.2", 0, false},
		{"N/A", 0, false},
		{"", 0, false},
	}
	for _, tc := range cases {
		got, ok := parseMetricPercent(tc.raw)
		if ok != tc.valid || got != tc.want {
			t.Fatalf("parseMetricPercent(%q) = (%.2f, %v), want (%.2f, %v)", tc.raw, got, ok, tc.want, tc.valid)
		}
	}

	items := []Item{
		{Key: "system.cpu.util[,idle]", Name: "CPU idle time", Lastvalue: "12.5"},
		{Key: "vm.memory.size[total]", Name: "Total memory", Lastvalue: "1000"},
		{Key: "vm.memory.size[available]", Name: "Available memory", Lastvalue: "150"},
		{Key: "vfs.fs.size[/,pused]", Name: "/: Space utilization", Lastvalue: "62.4"},
		{Key: "vfs.fs.dependent.size[/var/lib/docker,pfree]", Name: "/var/lib/docker: Free", Lastvalue: "8.5"},
		{Key: "vfs.fs.inode[/,pfree]", Name: "Inode free", Lastvalue: "90"},
		{Key: "icmpping", Name: "ICMP ping", Lastvalue: "1"},
	}

	cpu, gotCPU, mem, gotMem := extractCPUMemFromItems(items, 0, 0, false, false)
	if !gotCPU || cpu != 87.5 {
		t.Fatalf("CPU extraction = (%.2f, %v), want (87.50, true)", cpu, gotCPU)
	}
	if !gotMem || mem != 85.0 {
		t.Fatalf("Mem extraction = (%.2f, %v), want (85.00, true)", mem, gotMem)
	}

	disk, mount, gotDisk := extractDiskUsageFromItems(items)
	if !gotDisk || disk != 91.5 || mount != "/var/lib/docker" {
		t.Fatalf("Disk extraction = (%.2f, %q, %v), want (91.50, %q, true)", disk, mount, gotDisk, "/var/lib/docker")
	}

	if ping := extractPingFromItems(items); ping != "1" {
		t.Fatalf("extractPingFromItems() = %q, want %q", ping, "1")
	}
}

func TestComputeInspectionScoreScenarios(t *testing.T) {
	// 1. 全健康场景 -> 100 分
	healthyHosts := []InspectionHostItem{
		{Name: "web-01", IP: "10.0.0.1", Available: "1", CpuPct: 35, HasCpu: true, MemPct: 48, HasMem: true, DiskPct: 50, HasDisk: true, DiskMount: "/"},
		{Name: "web-02", IP: "10.0.0.2", Available: "1", CpuPct: 42, HasCpu: true, MemPct: 52, HasMem: true, DiskPct: 55, HasDisk: true, DiskMount: "/"},
	}
	score, grade, bd, recs := ComputeInspectionScore(healthyHosts, nil, 80, 80, 85)
	if score != 100 || grade != "excellent" {
		t.Fatalf("healthy score = (%d, %q), want (100, excellent)", score, grade)
	}
	if bd.AvailabilityScore != 35 || bd.ResourceScore != 30 || bd.AlarmScore != 25 || bd.StabilityScore != 10 {
		t.Fatalf("unexpected healthy breakdown: %+v", bd)
	}
	if len(recs) == 0 || !strings.Contains(recs[0], "整体运行健康") {
		t.Fatalf("unexpected healthy recommendation: %v", recs)
	}

	// 2. 失联 + 资源超标 + 高危告警混合场景
	mixedHosts := []InspectionHostItem{
		{Name: "web-01", IP: "10.0.0.1", Available: "1", CpuPct: 92, HasCpu: true, MemPct: 86, HasMem: true, DiskPct: 91, HasDisk: true, DiskMount: "/data", ActiveAlarms: 2, CriticalAlarms: 1},
		{Name: "db-01", IP: "10.0.0.2", Available: "2", CpuPct: 0, HasCpu: false, MemPct: 0, HasMem: false},
		{Name: "app-01", IP: "10.0.0.3", Available: "1", CpuPct: 20, HasCpu: true, MemPct: 40, HasMem: true, DiskPct: 35, HasDisk: true, DiskMount: "/"},
	}
	alarms := []InspectionAlarmItem{
		{Hostname: "web-01", Level: "灾难", SeverityNum: 5, Message: "Disk space low on /data"},
		{Hostname: "web-01", Level: "一般严重", SeverityNum: 3, Message: "CPU load high"},
		{Hostname: "db-01", Level: "严重", SeverityNum: 4, Message: "Zabbix agent unreachable"},
	}
	score2, grade2, bd2, recs2 := ComputeInspectionScore(mixedHosts, alarms, 80, 80, 85)
	if score2 >= 75 || (grade2 != "warning" && grade2 != "critical") {
		t.Fatalf("expected degraded score < 75, got (%d, %q)", score2, grade2)
	}
	if bd2.AvailabilityDeduct == 0 || bd2.ResourceDeduct == 0 || bd2.AlarmDeduct == 0 || bd2.StabilityDeduct == 0 {
		t.Fatalf("expected deductions across all 4 dimensions, got %+v", bd2)
	}
	if mixedHosts[0].HostScore >= mixedHosts[2].HostScore || mixedHosts[1].Status != "offline" || mixedHosts[0].Status != "critical" {
		t.Fatalf("unexpected host statuses/scores: web=%+v db=%+v app=%+v", mixedHosts[0], mixedHosts[1], mixedHosts[2])
	}
	joinedRecs := strings.Join(recs2, "\n")
	if !strings.Contains(joinedRecs, "db-01") || !strings.Contains(joinedRecs, "/data") || !strings.Contains(joinedRecs, "高危未恢复告警") {
		t.Fatalf("expected actionable details in recommendations, got:\n%s", joinedRecs)
	}
}

func TestNormalizeInspectionTaskDefaults(t *testing.T) {
	task := &InspectionTask{
		Name:           " 每日核心巡检 ",
		Cycle:          "week",
		NotifyChannels: `["Mail","dingtalk","wechat_robot"]`,
	}
	if err := normalizeInspectionTaskDefaults(task); err != nil {
		t.Fatalf("normalizeInspectionTaskDefaults failed: %v", err)
	}
	if task.Name != "每日核心巡检" || task.ZIDs != "*" || task.HostTypes != "*" {
		t.Fatalf("unexpected task scope defaults: %+v", task)
	}
	if task.CronExpr != "0 0 8 * * 1" {
		t.Fatalf("expected weekly default cron 0 0 8 * * 1, got %q", task.CronExpr)
	}
	if task.CpuThreshold != 80 || task.MemThreshold != 80 || task.DiskThreshold != 85 {
		t.Fatalf("unexpected thresholds: cpu=%.1f mem=%.1f disk=%.1f", task.CpuThreshold, task.MemThreshold, task.DiskThreshold)
	}
	if task.NotifyChannels != "mail,dingtalk_robot,wechat_robot" {
		t.Fatalf("unexpected normalized notify channels: %q", task.NotifyChannels)
	}
}

func TestCreateInspectionReportFiles(t *testing.T) {
	tmpDir := t.TempDir()
	origWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origWD) }()

	now := time.Now()
	task := &InspectionTask{
		ID:            1,
		Name:          "核心生产巡检",
		CpuThreshold:  80,
		MemThreshold:  80,
		DiskThreshold: 85,
	}
	hosts := []InspectionHostItem{
		{
			ZID: 1, InstanceName: "Prod-Zabbix", HostID: "1001", Name: "prod-db-01", IP: "192.168.10.11",
			TypeCode: "VM_LIN", TypeName: "Linux", Available: "1",
			CpuPct: 89.2, HasCpu: true, MemPct: 72.5, HasMem: true, DiskPct: 92.4, HasDisk: true, DiskMount: "/var",
			ActiveAlarms: 1, CriticalAlarms: 1,
		},
		{
			ZID: 1, InstanceName: "Prod-Zabbix", HostID: "1002", Name: "prod-web-01", IP: "192.168.10.12",
			TypeCode: "VM_LIN", TypeName: "Linux", Available: "1",
			CpuPct: 24.0, HasCpu: true, MemPct: 45.0, HasMem: true, DiskPct: 38.0, HasDisk: true, DiskMount: "/",
		},
	}
	alarms := []InspectionAlarmItem{
		{ZID: 1, InstanceName: "Prod-Zabbix", Hostname: "prod-db-01", HostIP: "192.168.10.11", Level: "严重", SeverityNum: 4, Message: "/var usage > 90%", OccurTime: now.Format("2006-01-02 15:04:05")},
	}
	score, grade, bd, recs := ComputeInspectionScore(hosts, alarms, 80, 80, 85)
	summary := &InspectionSummary{
		ScoreBreakdown:  bd,
		Recommendations: recs,
		TopAlarms:       alarms,
		CpuThreshold:    80,
		MemThreshold:    80,
		DiskThreshold:   85,
	}
	report := &InspectionReport{
		ID:             101,
		TaskID:         1,
		TaskName:       task.Name,
		TriggerType:    "manual",
		Status:         Success,
		HealthScore:    score,
		HealthGrade:    grade,
		TotalInstances: 1,
		TotalHosts:     2,
		AvailableHosts: 2,
		HighCpuHosts:   1,
		HighDiskHosts:  1,
		ActiveAlarms:   1,
		CriticalAlarms: 1,
		AbnormalHosts:  1,
		StartTime:      now,
		TotalTime:      2,
	}

	if err := generateInspectionReportFiles(task, report, summary, hosts); err != nil {
		t.Fatalf("generateInspectionReportFiles failed: %v", err)
	}
	if report.ExcelFile == "" || report.HtmlFile == "" || report.Files == "" {
		t.Fatalf("expected excel, html and zip files, got excel=%q html=%q zip=%q", report.ExcelFile, report.HtmlFile, report.Files)
	}

	xlsxFile, err := excelize.OpenFile(filepath.Join(DownloadPath, report.ExcelFile))
	if err != nil {
		t.Fatalf("failed to open generated xlsx: %v", err)
	}
	defer func() { _ = xlsxFile.Close() }()
	sheets := xlsxFile.GetSheetList()
	if len(sheets) != 3 {
		t.Fatalf("expected 3 sheets, got %v", sheets)
	}
	hostCell, err := xlsxFile.GetCellValue("主机巡检明细", "B2")
	if err != nil || hostCell != "prod-db-01" {
		t.Fatalf("expected prod-db-01 first in host sheet, got %q (err=%v)", hostCell, err)
	}

	htmlBytes, err := os.ReadFile(filepath.Join(DownloadPath, report.HtmlFile))
	if err != nil {
		t.Fatalf("failed to read generated html: %v", err)
	}
	htmlStr := string(htmlBytes)
	if !strings.Contains(htmlStr, "核心生产巡检") || !strings.Contains(htmlStr, "prod-db-01") || !strings.Contains(htmlStr, "/var usage") {
		t.Fatalf("generated html missing expected content")
	}
}

func TestFilterUnclosedLocalAlarmsExcludesRecovered(t *testing.T) {
	base := time.Now().Add(-2 * time.Hour)
	rows := []Alarm{
		// 1. 同一 EventID 已产生恢复记录 ("0") -> 应剔除
		{ID: 1, ZID: 1, HostID: "1001", Host: "web-01", Hostname: "web-01", TriggerID: 5001, EventID: 9001, Status: "1", Level: "4", Message: "High CPU", OccurTime: base},
		{ID: 2, ZID: 1, HostID: "1001", Host: "web-01", Hostname: "web-01", TriggerID: 5001, EventID: 9001, Status: "0", Level: "4", Message: "High CPU", OccurTime: base.Add(5 * time.Minute)},

		// 2. EventID 不同但同主机 + 同 TriggerID 且产生过中文"恢复" -> 应剔除
		{ID: 3, ZID: 1, HostID: "1002", Host: "db-01", Hostname: "db-01", TriggerID: 5002, EventID: 9002, Status: "故障", Level: "5", Message: "DB down", OccurTime: base.Add(10 * time.Minute)},
		{ID: 4, ZID: 1, HostID: "1002", Host: "db-01", Hostname: "db-01", TriggerID: 5002, EventID: 9099, Status: "恢复", Level: "5", Message: "DB down", OccurTime: base.Add(12 * time.Minute)},

		// 3. TriggerID 为 0，通过同主机 + 同消息在更晚时间恢复 -> 应剔除
		{ID: 5, ZID: 1, HostID: "1003", Host: "app-01", Hostname: "app-01", EventID: 9003, Status: "1", Level: "3", Message: "Port 8080 unreachable", OccurTime: base.Add(15 * time.Minute)},
		{ID: 6, ZID: 1, HostID: "1003", Host: "app-01", Hostname: "app-01", EventID: 9004, Status: "0", Level: "3", Message: "Port 8080 unreachable", OccurTime: base.Add(18 * time.Minute)},

		// 4. escalation_step = -1 已标记关闭 -> 应剔除
		{ID: 7, ZID: 1, HostID: "1004", Host: "cache-01", Hostname: "cache-01", TriggerID: 5004, EventID: 9005, Status: "1", Level: "2", Message: "Memory high", EscalationStep: -1, OccurTime: base.Add(20 * time.Minute)},

		// 5. 先恢复过一次，之后又再次触发且尚未恢复 + 同触发器重复触发一条 -> 应保留最新 1 条
		{ID: 8, ZID: 1, HostID: "1005", Host: "mq-01", Hostname: "mq-01", TriggerID: 5005, EventID: 9006, Status: "0", Level: "4", Message: "Disk > 90%", OccurTime: base.Add(22 * time.Minute)},
		{ID: 9, ZID: 1, HostID: "1005", Host: "mq-01", Hostname: "mq-01", TriggerID: 5005, EventID: 9007, Status: "1", Level: "4", Message: "Disk > 90%", OccurTime: base.Add(30 * time.Minute)},
		{ID: 10, ZID: 1, HostID: "1005", Host: "mq-01", Hostname: "mq-01", TriggerID: 5005, EventID: 9008, Status: "1", Level: "4", Message: "Disk > 90%", OccurTime: base.Add(35 * time.Minute)},

		// 6. 真正未恢复的高危告警（但在 allowedHostIDs 范围之外）
		{ID: 11, ZID: 1, HostID: "9999", Host: "other-01", Hostname: "other-01", TriggerID: 5006, EventID: 9009, Status: "1", Level: "5", Message: "Outside scope", OccurTime: base.Add(40 * time.Minute)},
	}

	gotAll := filterUnclosedLocalAlarms(rows, nil)
	if len(gotAll) != 2 {
		t.Fatalf("expected 2 unclosed alarms without host filter, got %d: %+v", len(gotAll), gotAll)
	}
	if gotAll[0].HostID != "9999" || gotAll[1].HostID != "1005" || gotAll[1].EventID != "9008" {
		t.Fatalf("unexpected remaining alarms: %+v", gotAll)
	}

	allowed := map[string]bool{"1001": true, "1002": true, "1003": true, "1004": true, "1005": true}
	gotScoped := filterUnclosedLocalAlarms(rows, allowed)
	if len(gotScoped) != 1 || gotScoped[0].HostID != "1005" || gotScoped[0].EventID != "9008" {
		t.Fatalf("expected only host 1005 latest active alarm in scope, got %+v", gotScoped)
	}
}

func TestExtractActiveTriggerAlarmsStrictFiltering(t *testing.T) {
	triggers := []inspectRawTrigger{
		{
			Triggerid:   "101",
			Description: "Agent unavailable",
			Priority:    "4",
			Value:       "1",
			Lastchange:  "1700000100",
			Hosts: []struct {
				Hostid string `json:"hostid"`
				Host   string `json:"host"`
				Name   string `json:"name"`
			}{{Hostid: "1001", Host: "prod-web-01", Name: "Web 01"}},
			LastEvent: struct {
				Acknowledged string `json:"acknowledged"`
				Clock        string `json:"clock"`
				Eventid      string `json:"eventid"`
				Name         string `json:"name"`
				Severity     string `json:"severity"`
				Value        string `json:"value"`
			}{Eventid: "7001", Value: "1", Severity: "4", Clock: "1700000100", Name: "Agent unavailable", Acknowledged: "0"},
		},
		{
			// 触发器当前值已恢复为 0 -> 应剔除
			Triggerid:   "102",
			Description: "Recovered trigger value",
			Priority:    "3",
			Value:       "0",
			Lastchange:  "1700000200",
			Hosts: []struct {
				Hostid string `json:"hostid"`
				Host   string `json:"host"`
				Name   string `json:"name"`
			}{{Hostid: "1001", Host: "prod-web-01", Name: "Web 01"}},
			LastEvent: struct {
				Acknowledged string `json:"acknowledged"`
				Clock        string `json:"clock"`
				Eventid      string `json:"eventid"`
				Name         string `json:"name"`
				Severity     string `json:"severity"`
				Value        string `json:"value"`
			}{Eventid: "7002", Value: "1", Severity: "3"},
		},
		{
			// LastEvent 已经是恢复事件 (Value == "0") -> 应剔除
			Triggerid:   "103",
			Description: "Resolved last event",
			Priority:    "3",
			Value:       "1",
			Lastchange:  "1700000300",
			Hosts: []struct {
				Hostid string `json:"hostid"`
				Host   string `json:"host"`
				Name   string `json:"name"`
			}{{Hostid: "1001", Host: "prod-web-01", Name: "Web 01"}},
			LastEvent: struct {
				Acknowledged string `json:"acknowledged"`
				Clock        string `json:"clock"`
				Eventid      string `json:"eventid"`
				Name         string `json:"name"`
				Severity     string `json:"severity"`
				Value        string `json:"value"`
			}{Eventid: "7003", Value: "0", Severity: "3"},
		},
	}

	got := extractActiveTriggerAlarms(1, "Prod", triggers, map[string]bool{"1001": true}, make(map[string]bool))
	if len(got) != 1 || got[0].EventID != "7001" || got[0].Host != "prod-web-01" || got[0].Hostname != "Web 01" {
		t.Fatalf("expected only active trigger 7001, got %+v", got)
	}
}
