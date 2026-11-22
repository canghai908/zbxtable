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
