package model

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"

	"github.com/signintech/gopdf"
	"github.com/xuri/excelize/v2"
)

func sanitizeReportFilename(name string) string {
	cleaned := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	).Replace(strings.TrimSpace(name))
	if cleaned == "" {
		cleaned = "inspection"
	}
	runes := []rune(cleaned)
	if len(runes) > 40 {
		cleaned = string(runes[:40])
	}
	return cleaned
}

func generateInspectionReportFiles(task *InspectionTask, report *InspectionReport, summary *InspectionSummary, hosts []InspectionHostItem) error {
	if err := utils.Mkdir(DownloadPath); err != nil {
		return err
	}

	baseName := fmt.Sprintf("%s_inspect_%s_%d", sanitizeReportFilename(task.Name), time.Now().Format("20060102_150405"), report.ID)
	filesToZip := make([]string, 0, 3)

	if xlsxName, err := CreateInspectionReportXlsx(task, report, summary, hosts, baseName+".xlsx"); err != nil {
		logger.Log.Warnf("生成巡检 Excel 失败: %v", err)
	} else if xlsxName != "" {
		report.ExcelFile = xlsxName
		filesToZip = append(filesToZip, filepath.Join(DownloadPath, xlsxName))
	}

	if htmlName, err := CreateInspectionReportHTML(task, report, summary, hosts, baseName+".html"); err != nil {
		logger.Log.Warnf("生成巡检 HTML 失败: %v", err)
	} else if htmlName != "" {
		report.HtmlFile = htmlName
		filesToZip = append(filesToZip, filepath.Join(DownloadPath, htmlName))
	}

	if pdfName, err := CreateInspectionReportPDF(task, report, summary, hosts, baseName+".pdf"); err != nil {
		logger.Log.Warnf("生成巡检 PDF 跳过或失败: %v", err)
	} else if pdfName != "" {
		report.PdfFile = pdfName
		filesToZip = append(filesToZip, filepath.Join(DownloadPath, pdfName))
	}

	if len(filesToZip) > 0 {
		zipName := baseName + ".zip"
		if err := utils.ZipFiles(filepath.Join(DownloadPath, zipName), filesToZip, DownloadPath, baseName+"/"); err != nil {
			logger.Log.Warnf("打包巡检 ZIP 失败: %v", err)
		} else {
			report.Files = zipName
		}
	}
	return nil
}

// CreateInspectionReportXlsx 生成多 Sheet、带异常自动标红的 Excel 巡检报告，返回文件名（位于 DownloadPath 下）
func CreateInspectionReportXlsx(task *InspectionTask, report *InspectionReport, summary *InspectionSummary, hosts []InspectionHostItem, filename string) (string, error) {
	if err := utils.Mkdir(DownloadPath); err != nil {
		return "", err
	}
	xlsx := excelize.NewFile()
	defer func() { _ = xlsx.Close() }()

	const (
		sheetOverview = "巡检总览与评分"
		sheetHosts    = "主机巡检明细"
		sheetAlarms   = "活跃告警明细"
	)

	defaultSheet := xlsx.GetSheetName(0)
	_ = xlsx.SetSheetName(defaultSheet, sheetOverview)
	_, _ = xlsx.NewSheet(sheetHosts)
	_, _ = xlsx.NewSheet(sheetAlarms)

	titleStyle, _ := xlsx.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 15, Color: "#1F2937"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})
	headerStyle, _ := xlsx.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 11, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1890FF"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "#D9D9D9", Style: 1},
			{Type: "right", Color: "#D9D9D9", Style: 1},
			{Type: "top", Color: "#D9D9D9", Style: 1},
			{Type: "bottom", Color: "#D9D9D9", Style: 1},
		},
	})
	normalStyle, _ := xlsx.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#E5E7EB", Style: 1},
			{Type: "right", Color: "#E5E7EB", Style: 1},
			{Type: "top", Color: "#E5E7EB", Style: 1},
			{Type: "bottom", Color: "#E5E7EB", Style: 1},
		},
	})
	alertStyle, _ := xlsx.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#CF1322"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#FFF1F0"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#FFA39E", Style: 1},
			{Type: "right", Color: "#FFA39E", Style: 1},
			{Type: "top", Color: "#FFA39E", Style: 1},
			{Type: "bottom", Color: "#FFA39E", Style: 1},
		},
	})
	warnStyle, _ := xlsx.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#D46B08"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#FFF7E6"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "#FFD591", Style: 1},
			{Type: "right", Color: "#FFD591", Style: 1},
			{Type: "top", Color: "#FFD591", Style: 1},
			{Type: "bottom", Color: "#FFD591", Style: 1},
		},
	})

	// ---- Sheet 1: 巡检总览与评分 ----
	_ = xlsx.SetColWidth(sheetOverview, "A", "A", 22)
	_ = xlsx.SetColWidth(sheetOverview, "B", "D", 22)
	_ = xlsx.MergeCell(sheetOverview, "A1", "D1")
	_ = xlsx.SetCellValue(sheetOverview, "A1", fmt.Sprintf("ZbxTable 自动化巡检健康评分报告 - %s", task.Name))
	_ = xlsx.SetCellStyle(sheetOverview, "A1", "D1", titleStyle)
	_ = xlsx.SetRowHeight(sheetOverview, 1, 28)

	overviewRows := [][]interface{}{
		{"巡检项目", "巡检结果", "说明/阈值", "状态评价"},
		{"综合健康评分", fmt.Sprintf("%d 分", report.HealthScore), "满分 100 分", healthGradeLabel(report.HealthGrade)},
		{"巡检时间", report.StartTime.Format("2006-01-02 15:04:05"), fmt.Sprintf("耗时 %d 秒", report.TotalTime), report.TriggerType},
		{"纳管实例 / 设备总数", fmt.Sprintf("%d 实例 / %d 台", report.TotalInstances, report.TotalHosts), fmt.Sprintf("正常 %d / 失联 %d", report.AvailableHosts, report.UnavailableHosts), boolStatusLabel(report.UnavailableHosts == 0)},
		{"CPU 超标主机", report.HighCpuHosts, fmt.Sprintf("阈值 ≥ %.0f%%", summary.CpuThreshold), boolStatusLabel(report.HighCpuHosts == 0)},
		{"内存超标主机", report.HighMemHosts, fmt.Sprintf("阈值 ≥ %.0f%%", summary.MemThreshold), boolStatusLabel(report.HighMemHosts == 0)},
		{"磁盘超标主机", report.HighDiskHosts, fmt.Sprintf("阈值 ≥ %.0f%%", summary.DiskThreshold), boolStatusLabel(report.HighDiskHosts == 0)},
		{"活跃告警 / 高危告警", fmt.Sprintf("%d / %d", report.ActiveAlarms, report.CriticalAlarms), "含未恢复触发器与本地未关闭告警", boolStatusLabel(report.CriticalAlarms == 0)},
		{"可用性健康度", fmt.Sprintf("%d / %d", summary.ScoreBreakdown.AvailabilityScore, summary.ScoreBreakdown.AvailabilityMax), fmt.Sprintf("扣 %d 分", summary.ScoreBreakdown.AvailabilityDeduct), ""},
		{"资源水位健康度", fmt.Sprintf("%d / %d", summary.ScoreBreakdown.ResourceScore, summary.ScoreBreakdown.ResourceMax), fmt.Sprintf("扣 %d 分", summary.ScoreBreakdown.ResourceDeduct), ""},
		{"告警风险健康度", fmt.Sprintf("%d / %d", summary.ScoreBreakdown.AlarmScore, summary.ScoreBreakdown.AlarmMax), fmt.Sprintf("扣 %d 分", summary.ScoreBreakdown.AlarmDeduct), ""},
		{"集群稳定度", fmt.Sprintf("%d / %d", summary.ScoreBreakdown.StabilityScore, summary.ScoreBreakdown.StabilityMax), fmt.Sprintf("扣 %d 分", summary.ScoreBreakdown.StabilityDeduct), ""},
	}
	for idx, row := range overviewRows {
		rNum := idx + 3
		for cIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, rNum)
			_ = xlsx.SetCellValue(sheetOverview, cell, val)
			if idx == 0 {
				_ = xlsx.SetCellStyle(sheetOverview, cell, cell, headerStyle)
			} else if len(row) == 4 && row[3] == "异常" {
				_ = xlsx.SetCellStyle(sheetOverview, cell, cell, alertStyle)
			} else {
				_ = xlsx.SetCellStyle(sheetOverview, cell, cell, normalStyle)
			}
		}
	}
	recStartRow := len(overviewRows) + 5
	_ = xlsx.MergeCell(sheetOverview, fmt.Sprintf("A%d", recStartRow), fmt.Sprintf("D%d", recStartRow))
	_ = xlsx.SetCellValue(sheetOverview, fmt.Sprintf("A%d", recStartRow), "巡检结论与运维处置建议")
	_ = xlsx.SetCellStyle(sheetOverview, fmt.Sprintf("A%d", recStartRow), fmt.Sprintf("D%d", recStartRow), headerStyle)
	for idx, rec := range summary.Recommendations {
		r := recStartRow + 1 + idx
		_ = xlsx.MergeCell(sheetOverview, fmt.Sprintf("A%d", r), fmt.Sprintf("D%d", r))
		_ = xlsx.SetCellValue(sheetOverview, fmt.Sprintf("A%d", r), fmt.Sprintf("%d. %s", idx+1, rec))
		_ = xlsx.SetCellStyle(sheetOverview, fmt.Sprintf("A%d", r), fmt.Sprintf("D%d", r), normalStyle)
	}

	// ---- Sheet 2: 主机巡检明细（异常自动标红） ----
	hostHeaders := []string{"所属实例", "设备名称", "IP地址", "设备类型", "采集状态", "CPU(%)", "内存(%)", "最大磁盘(%)", "磁盘挂载", "活跃告警", "高危告警", "单机评分", "巡检结论", "异常明细"}
	for i, h := range hostHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = xlsx.SetCellValue(sheetHosts, cell, h)
		_ = xlsx.SetCellStyle(sheetHosts, cell, cell, headerStyle)
	}
	_ = xlsx.SetColWidth(sheetHosts, "A", "D", 18)
	_ = xlsx.SetColWidth(sheetHosts, "E", "M", 12)
	_ = xlsx.SetColWidth(sheetHosts, "N", "N", 36)

	for idx, h := range hosts {
		r := idx + 2
		availText := "正常"
		if h.Available == "2" {
			availText = "失联"
		} else if h.Available == "0" || h.Available == "" {
			availText = "未知"
		}
		cpuText := "--"
		if h.HasCpu {
			cpuText = fmt.Sprintf("%.1f", h.CpuPct)
		}
		memText := "--"
		if h.HasMem {
			memText = fmt.Sprintf("%.1f", h.MemPct)
		}
		diskText := "--"
		if h.HasDisk {
			diskText = fmt.Sprintf("%.1f", h.DiskPct)
		}
		statusLabel := mapHostStatusLabel(h.Status)
		issuesText := "健康"
		if len(h.Issues) > 0 {
			issuesText = strings.Join(h.Issues, "; ")
		}

		vals := []interface{}{
			h.InstanceName, h.Name, h.IP, h.TypeName, availText,
			cpuText, memText, diskText, h.DiskMount,
			h.ActiveAlarms, h.CriticalAlarms, h.HostScore, statusLabel, issuesText,
		}
		rowStyle := normalStyle
		if h.Status == "offline" || h.Status == "critical" {
			rowStyle = alertStyle
		} else if h.Status == "warning" {
			rowStyle = warnStyle
		}
		for cIdx, val := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, r)
			_ = xlsx.SetCellValue(sheetHosts, cell, val)
			cellStyle := rowStyle
			// 单独高亮异常指标单元格
			if (cIdx == 4 && h.Available == "2") || (cIdx == 5 && h.CpuHigh) || (cIdx == 6 && h.MemHigh) || (cIdx == 7 && h.DiskHigh) || (cIdx == 10 && h.CriticalAlarms > 0) {
				cellStyle = alertStyle
			}
			_ = xlsx.SetCellStyle(sheetHosts, cell, cell, cellStyle)
		}
	}

	// ---- Sheet 3: 活跃告警明细 ----
	alarmHeaders := []string{"所属实例", "主机名称", "IP地址", "告警级别", "告警内容", "发生时间", "确认状态"}
	for i, h := range alarmHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = xlsx.SetCellValue(sheetAlarms, cell, h)
		_ = xlsx.SetCellStyle(sheetAlarms, cell, cell, headerStyle)
	}
	_ = xlsx.SetColWidth(sheetAlarms, "A", "D", 18)
	_ = xlsx.SetColWidth(sheetAlarms, "E", "E", 42)
	_ = xlsx.SetColWidth(sheetAlarms, "F", "G", 18)
	for idx, a := range summary.TopAlarms {
		r := idx + 2
		ackStr := "未确认"
		if a.Acknowledged {
			ackStr = "已确认"
		}
		vals := []interface{}{a.InstanceName, a.Hostname, a.HostIP, a.Level, a.Message, a.OccurTime, ackStr}
		style := normalStyle
		if a.SeverityNum >= 4 {
			style = alertStyle
		} else if a.SeverityNum >= 2 {
			style = warnStyle
		}
		for cIdx, val := range vals {
			cell, _ := excelize.CoordinatesToCellName(cIdx+1, r)
			_ = xlsx.SetCellValue(sheetAlarms, cell, val)
			_ = xlsx.SetCellStyle(sheetAlarms, cell, cell, style)
		}
	}

	fullPath := filepath.Join(DownloadPath, filename)
	if err := xlsx.SaveAs(fullPath); err != nil {
		return "", err
	}
	return filename, nil
}

func boolStatusLabel(ok bool) string {
	if ok {
		return "正常"
	}
	return "异常"
}

func mapHostStatusLabel(status string) string {
	switch status {
	case "normal":
		return "正常"
	case "warning":
		return "亚健康"
	case "critical":
		return "严重"
	case "offline":
		return "失联"
	default:
		return status
	}
}

type inspectHTMLViewData struct {
	TaskName        string
	ReportID        int
	HealthScore     int
	GradeLabel      string
	GradeColor      string
	TriggerType     string
	StartTime       string
	TotalTime       int64
	TotalInstances  int
	TotalHosts      int
	AvailableHosts  int
	UnavailableHost int
	HighCpuHosts    int
	HighMemHosts    int
	HighDiskHosts   int
	ActiveAlarms    int
	CriticalAlarms  int
	AbnormalHosts   int
	Breakdown       InspectionScoreBreakdown
	Recommendations []string
	InstanceStats   []InspectionInstanceStat
	Hosts           []inspectHTMLHostRow
	Alarms          []InspectionAlarmItem
	CpuThreshold    float64
	MemThreshold    float64
	DiskThreshold   float64
}

type inspectHTMLHostRow struct {
	InstanceName   string
	Name           string
	IP             string
	TypeName       string
	AvailLabel     string
	CpuText        string
	MemText        string
	DiskText       string
	CpuHigh        bool
	MemHigh        bool
	DiskHigh       bool
	ActiveAlarms   int
	CriticalAlarms int
	HostScore      int
	StatusLabel    string
	RowClass       string
	IssuesText     string
}

func gradeBadgeColor(grade string) string {
	switch grade {
	case "excellent":
		return "#52c41a"
	case "good":
		return "#1890ff"
	case "warning":
		return "#fa8c16"
	default:
		return "#f5222f"
	}
}

func renderInspectionHTMLBytes(task *InspectionTask, report *InspectionReport, summary *InspectionSummary, hosts []InspectionHostItem) ([]byte, error) {
	rows := make([]inspectHTMLHostRow, 0, len(hosts))
	for _, h := range hosts {
		avail := "正常"
		if h.Available == "2" {
			avail = "失联"
		} else if h.Available == "0" || h.Available == "" {
			avail = "未知"
		}
		cpuText, memText, diskText := "--", "--", "--"
		if h.HasCpu {
			cpuText = fmt.Sprintf("%.1f%%", h.CpuPct)
		}
		if h.HasMem {
			memText = fmt.Sprintf("%.1f%%", h.MemPct)
		}
		if h.HasDisk {
			if h.DiskMount != "" {
				diskText = fmt.Sprintf("%.1f%% (%s)", h.DiskPct, h.DiskMount)
			} else {
				diskText = fmt.Sprintf("%.1f%%", h.DiskPct)
			}
		}
		rowClass := "row-normal"
		if h.Status == "offline" || h.Status == "critical" {
			rowClass = "row-critical"
		} else if h.Status == "warning" {
			rowClass = "row-warning"
		}
		issuesText := "运行正常"
		if len(h.Issues) > 0 {
			issuesText = strings.Join(h.Issues, "；")
		}
		rows = append(rows, inspectHTMLHostRow{
			InstanceName:   h.InstanceName,
			Name:           h.Name,
			IP:             h.IP,
			TypeName:       h.TypeName,
			AvailLabel:     avail,
			CpuText:        cpuText,
			MemText:        memText,
			DiskText:       diskText,
			CpuHigh:        h.CpuHigh,
			MemHigh:        h.MemHigh,
			DiskHigh:       h.DiskHigh,
			ActiveAlarms:   h.ActiveAlarms,
			CriticalAlarms: h.CriticalAlarms,
			HostScore:      h.HostScore,
			StatusLabel:    mapHostStatusLabel(h.Status),
			RowClass:       rowClass,
			IssuesText:     issuesText,
		})
	}

	view := inspectHTMLViewData{
		TaskName:        task.Name,
		ReportID:        report.ID,
		HealthScore:     report.HealthScore,
		GradeLabel:      healthGradeLabel(report.HealthGrade),
		GradeColor:      gradeBadgeColor(report.HealthGrade),
		TriggerType:     report.TriggerType,
		StartTime:       report.StartTime.Format("2006-01-02 15:04:05"),
		TotalTime:       report.TotalTime,
		TotalInstances:  report.TotalInstances,
		TotalHosts:      report.TotalHosts,
		AvailableHosts:  report.AvailableHosts,
		UnavailableHost: report.UnavailableHosts,
		HighCpuHosts:    report.HighCpuHosts,
		HighMemHosts:    report.HighMemHosts,
		HighDiskHosts:   report.HighDiskHosts,
		ActiveAlarms:    report.ActiveAlarms,
		CriticalAlarms:  report.CriticalAlarms,
		AbnormalHosts:   report.AbnormalHosts,
		Breakdown:       summary.ScoreBreakdown,
		Recommendations: summary.Recommendations,
		InstanceStats:   summary.InstanceStats,
		Hosts:           rows,
		Alarms:          summary.TopAlarms,
		CpuThreshold:    summary.CpuThreshold,
		MemThreshold:    summary.MemThreshold,
		DiskThreshold:   summary.DiskThreshold,
	}

	tmpl, err := template.New("inspectReportHTML").Parse(inspectionReportHTMLTemplate)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, view); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// CreateInspectionReportHTML 生成独立可视化 HTML 巡检报告
func CreateInspectionReportHTML(task *InspectionTask, report *InspectionReport, summary *InspectionSummary, hosts []InspectionHostItem, filename string) (string, error) {
	if err := utils.Mkdir(DownloadPath); err != nil {
		return "", err
	}
	htmlBytes, err := renderInspectionHTMLBytes(task, report, summary, hosts)
	if err != nil {
		return "", err
	}
	fullPath := filepath.Join(DownloadPath, filename)
	if err := os.WriteFile(fullPath, htmlBytes, 0644); err != nil {
		return "", err
	}
	return filename, nil
}

// CreateInspectionReportPDF 生成 A4 PDF 巡检报告；若系统缺少字体文件则跳过并返回空文件名而不阻断主流程
func CreateInspectionReportPDF(task *InspectionTask, report *InspectionReport, summary *InspectionSummary, hosts []InspectionHostItem, filename string) (string, error) {
	if err := utils.Mkdir(DownloadPath); err != nil {
		return "", err
	}
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595.28, H: 841.89}})
	pdf.AddPage()

	fontPaths := []string{"./msty.ttf", "msty.ttf", "fonts/msty.ttf", "./assets/fonts/msty.ttf"}
	fontAdded := false
	for _, fp := range fontPaths {
		if err := pdf.AddTTFFont("msty", fp); err == nil {
			fontAdded = true
			break
		}
	}
	if !fontAdded {
		if logger.Log != nil {
			logger.Log.Info("未检测到中文字体文件(msty.ttf)，跳过巡检 PDF 报告生成")
		}
		return "", nil
	}
	if err := pdf.SetFont("msty", "", 11); err != nil {
		return "", err
	}

	writeLine := func(y float64, text string) float64 {
		if y > 790 {
			pdf.AddPage()
			y = 30
		}
		pdf.SetX(24)
		pdf.SetY(y)
		_ = pdf.Cell(nil, text)
		return y + 18
	}

	y := 30.0
	_ = pdf.SetFont("msty", "", 15)
	y = writeLine(y, fmt.Sprintf("ZbxTable 自动化巡检与健康评分报告 - %s", task.Name))
	_ = pdf.SetFont("msty", "", 11)
	y = writeLine(y, fmt.Sprintf("巡检时间: %s   耗时: %ds   综合健康分: %d (%s)",
		report.StartTime.Format("2006-01-02 15:04:05"), report.TotalTime, report.HealthScore, healthGradeLabel(report.HealthGrade)))
	y = writeLine(y, fmt.Sprintf("设备概览: 实例 %d 个 | 主机总数 %d 台 | 正常 %d | 失联 %d | 异常主机 %d",
		report.TotalInstances, report.TotalHosts, report.AvailableHosts, report.UnavailableHosts, report.AbnormalHosts))
	y = writeLine(y, fmt.Sprintf("水位与告警: CPU超标 %d 台 | 内存超标 %d 台 | 磁盘超标 %d 台 | 活跃告警 %d (高危 %d)",
		report.HighCpuHosts, report.HighMemHosts, report.HighDiskHosts, report.ActiveAlarms, report.CriticalAlarms))
	y = writeLine(y, fmt.Sprintf("四维得分: 可用性 %d/%d | 资源水位 %d/%d | 告警风险 %d/%d | 稳定度 %d/%d",
		summary.ScoreBreakdown.AvailabilityScore, summary.ScoreBreakdown.AvailabilityMax,
		summary.ScoreBreakdown.ResourceScore, summary.ScoreBreakdown.ResourceMax,
		summary.ScoreBreakdown.AlarmScore, summary.ScoreBreakdown.AlarmMax,
		summary.ScoreBreakdown.StabilityScore, summary.ScoreBreakdown.StabilityMax))
	pdf.Line(24, y, 570, y)
	y += 14

	y = writeLine(y, "【巡检结论与建议】")
	for i, rec := range summary.Recommendations {
		y = writeLine(y, fmt.Sprintf("%d. %s", i+1, rec))
	}
	y += 6

	if len(hosts) > 0 {
		y = writeLine(y, "【主机巡检明细（异常主机优先展示）】")
		limit := len(hosts)
		if limit > 80 {
			limit = 80
		}
		for i := 0; i < limit; i++ {
			h := hosts[i]
			cpuS, memS, diskS := "--", "--", "--"
			if h.HasCpu {
				cpuS = fmt.Sprintf("%.1f%%", h.CpuPct)
			}
			if h.HasMem {
				memS = fmt.Sprintf("%.1f%%", h.MemPct)
			}
			if h.HasDisk {
				diskS = fmt.Sprintf("%.1f%%", h.DiskPct)
			}
			line := fmt.Sprintf("[%s] %s (%s) | CPU:%s 内存:%s 磁盘:%s | 告警:%d | %d分 [%s]",
				mapHostStatusLabel(h.Status), h.Name, h.IP, cpuS, memS, diskS, h.ActiveAlarms, h.HostScore, strings.Join(h.Issues, ";"))
			y = writeLine(y, line)
		}
	}

	var b bytes.Buffer
	if err := pdf.Write(&b); err != nil {
		return "", err
	}
	fullPath := filepath.Join(DownloadPath, filename)
	if err := os.WriteFile(fullPath, b.Bytes(), 0644); err != nil {
		return "", err
	}
	return filename, nil
}

// ========== 巡检报告自动推送 ==========

// NotifyInspectionReport 将巡检结果与健康评分报告推送至配置的渠道
func NotifyInspectionReport(task *InspectionTask, report *InspectionReport, summary *InspectionSummary) {
	if task == nil || report == nil || DB == nil {
		return
	}
	channels := splitInspectionCSV(task.NotifyChannels)
	if len(channels) == 0 && strings.TrimSpace(task.NotifyEmails) == "" {
		return
	}

	userIDs := resolveInspectionRecipientUserIDs(task)
	var users []User
	if len(userIDs) > 0 {
		_ = DB.Where("id IN ? AND status = 0", userIDs).
			Select("id", "username", "email", "wechat", "wechat_robot_key", "phone", "ding_talk", "ding_talk_secret", "feishu_robot_key", "feishu_secret", "webhook_url", "webhook_token").
			Find(&users).Error
	}

	// 1. 邮件推送（含 HTML 正文与压缩包附件）
	hasMailChannel := false
	for _, ch := range channels {
		if ch == "mail" {
			hasMailChannel = true
			break
		}
	}
	if hasMailChannel || strings.TrimSpace(task.NotifyEmails) != "" {
		emailSet := make(map[string]bool)
		var toEmails []string
		for _, em := range splitInspectionCSV(task.NotifyEmails) {
			if !emailSet[em] {
				emailSet[em] = true
				toEmails = append(toEmails, em)
			}
		}
		for _, u := range users {
			em := strings.TrimSpace(u.Email)
			if em != "" && !emailSet[em] {
				emailSet[em] = true
				toEmails = append(toEmails, em)
			}
		}
		if len(toEmails) > 0 && report.Files != "" {
			smtpFrom := GetConfigValueByKey("email_from", GetConfKey("email_from"))
			smtpHost := GetConfigValueByKey("email_host", GetConfKey("email_host"))
			smtpSecret := GetConfigValueByKey("email_secret", GetConfKey("email_secret"))
			if smtpFrom != "" && smtpHost != "" && smtpSecret != "" {
				var hosts []InspectionHostItem
				_ = json.Unmarshal([]byte(report.HostsJSON), &hosts)
				htmlBody, renderErr := renderInspectionHTMLBytes(task, report, summary, hosts)
				if renderErr == nil {
					subject := fmt.Sprintf("[ZbxTable巡检报告][%s] 健康评分 %d分(%s) - %s",
						task.Name, report.HealthScore, healthGradeLabel(report.HealthGrade), report.StartTime.Format("2006-01-02"))
					if mailErr := Sendmail(toEmails, subject, report.Files, htmlBody); mailErr != nil {
						logger.Log.Warnf("巡检报告邮件发送失败 (task=%s): %v", task.Name, mailErr)
					}
				}
			}
		}
	}

	// 2. IM 机器人与 Webhook 推送
	summaryText := buildInspectionNotifyText(task, report, summary)
	encKey := GetEncryptionKey()
	dummyEvent := &Event{
		ID:           report.ID,
		EventID:      int64(report.ID),
		HostID:       "inspection",
		Hostname:     task.Name,
		Host:         task.Name,
		ItemName:     fmt.Sprintf("健康评分 %d 分 (%s)", report.HealthScore, healthGradeLabel(report.HealthGrade)),
		Level:        inspectionGradeToSeverity(report.HealthGrade),
		Status:       "1",
		Message:      fmt.Sprintf("[自动巡检] %s 健康评分 %d 分 (%s)", task.Name, report.HealthScore, healthGradeLabel(report.HealthGrade)),
		Detail:       summaryText,
		OccurTime:    report.StartTime,
		Rule:         "[巡检任务]" + task.Name,
		InstanceName: fmt.Sprintf("%d个实例/%d台主机", report.TotalInstances, report.TotalHosts),
	}

	for _, ch := range channels {
		if ch == "mail" {
			continue
		}
		for _, u := range users {
			switch ch {
			case "wechat":
				if strings.TrimSpace(u.Wechat) != "" && WeApp != nil {
					_ = SendWechatAlert(u, dummyEvent, summaryText)
				}
			case "wechat_robot":
				if strings.TrimSpace(u.WechatRobotKey) != "" {
					key := decryptUserCredential(u.WechatRobotKey, encKey)
					_ = SendWechatRobotAlert(u, dummyEvent, summaryText, key)
				}
			case "dingtalk_robot":
				if strings.TrimSpace(u.DingTalk) != "" {
					key := decryptUserCredential(u.DingTalk, encKey)
					sec := decryptUserCredential(u.DingTalkSecret, encKey)
					_ = SendDingTalkRobotAlert(u, dummyEvent, summaryText, key, sec)
				}
			case "feishu_robot":
				if strings.TrimSpace(u.FeishuRobotKey) != "" {
					key := decryptUserCredential(u.FeishuRobotKey, encKey)
					sec := decryptUserCredential(u.FeishuSecret, encKey)
					_ = SendFeishuRobotAlert(u, dummyEvent, summaryText, key, sec)
				}
			case "webhook":
				if strings.TrimSpace(u.WebhookURL) != "" {
					tok := decryptUserCredential(u.WebhookToken, encKey)
					_ = SendWebhookChannelAlert(u, dummyEvent, summaryText, u.WebhookURL, tok)
				}
			}
		}
	}
}

func resolveInspectionRecipientUserIDs(task *InspectionTask) []int {
	gids := map[string]struct{}{}
	for _, gid := range splitInspectionCSV(task.NotifyGroupIDs) {
		gids[gid] = struct{}{}
	}
	groupMap, err := getBatchUsers(nil, gids)
	if err != nil {
		groupMap = map[string][]string{}
	}
	idStrs := getRuleUserIDs(task.NotifyUserIDs, task.NotifyGroupIDs, groupMap)
	seen := make(map[int]bool, len(idStrs))
	out := make([]int, 0, len(idStrs))
	for _, s := range idStrs {
		if id, err := strconv.Atoi(strings.TrimSpace(s)); err == nil && id > 0 && !seen[id] {
			seen[id] = true
			out = append(out, id)
		}
	}
	return out
}

func inspectionGradeToSeverity(grade string) string {
	switch grade {
	case "critical":
		return "5"
	case "warning":
		return "3"
	case "good":
		return "1"
	default:
		return "0"
	}
}

func buildInspectionNotifyText(task *InspectionTask, report *InspectionReport, summary *InspectionSummary) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("【ZbxTable 自动巡检报告】%s\n", task.Name))
	b.WriteString("=================================\n")
	b.WriteString(fmt.Sprintf("综合健康评分：%d 分（%s）\n", report.HealthScore, healthGradeLabel(report.HealthGrade)))
	b.WriteString(fmt.Sprintf("巡检时间：%s（耗时 %ds）\n", report.StartTime.Format("2006-01-02 15:04:05"), report.TotalTime))
	b.WriteString(fmt.Sprintf("设备统计：共 %d 实例 / %d 台主机（正常 %d，失联 %d，异常 %d）\n",
		report.TotalInstances, report.TotalHosts, report.AvailableHosts, report.UnavailableHosts, report.AbnormalHosts))
	b.WriteString(fmt.Sprintf("资源超标：CPU≥%.0f%% %d台 | 内存≥%.0f%% %d台 | 磁盘≥%.0f%% %d台\n",
		summary.CpuThreshold, report.HighCpuHosts, summary.MemThreshold, report.HighMemHosts, summary.DiskThreshold, report.HighDiskHosts))
	b.WriteString(fmt.Sprintf("活跃告警：共 %d 条（高危 %d 条）\n", report.ActiveAlarms, report.CriticalAlarms))
	if len(summary.Recommendations) > 0 {
		b.WriteString("巡检建议：\n")
		limit := len(summary.Recommendations)
		if limit > 3 {
			limit = 3
		}
		for i := 0; i < limit; i++ {
			b.WriteString(fmt.Sprintf("  %d) %s\n", i+1, summary.Recommendations[i]))
		}
	}
	return strings.TrimSpace(b.String())
}

const inspectionReportHTMLTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<title>巡检健康评分报告 - {{.TaskName}}</title>
<style>
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "PingFang SC", "Microsoft YaHei", sans-serif; background: #f5f7fa; color: #1f2937; margin: 0; padding: 24px; }
  .wrap { max-width: 1120px; margin: 0 auto; background: #fff; border-radius: 10px; box-shadow: 0 4px 16px rgba(0,0,0,0.06); overflow: hidden; }
  .banner { background: linear-gradient(135deg, #1d4ed8 0%, #2563eb 60%, #3b82f6 100%); color: #fff; padding: 24px 32px; display: flex; justify-content: space-between; align-items: center; }
  .banner h1 { margin: 0 0 6px; font-size: 22px; font-weight: 600; }
  .banner .meta { font-size: 13px; opacity: 0.9; }
  .score-box { background: rgba(255,255,255,0.16); border-radius: 10px; padding: 12px 22px; text-align: center; min-width: 120px; }
  .score-num { font-size: 34px; font-weight: 700; line-height: 1.1; }
  .score-grade { display: inline-block; margin-top: 4px; padding: 2px 10px; border-radius: 12px; font-size: 12px; font-weight: 600; background: {{.GradeColor}}; color: #fff; }
  .section { padding: 20px 32px; border-bottom: 1px solid #f0f2f5; }
  .section h2 { margin: 0 0 14px; font-size: 16px; color: #111827; border-left: 4px solid #2563eb; padding-left: 10px; }
  .kpi-grid { display: grid; grid-template-columns: repeat(4, 1fr); gap: 14px; }
  .kpi-card { background: #f9fafb; border: 1px solid #e5e7eb; border-radius: 8px; padding: 14px; }
  .kpi-title { font-size: 12px; color: #6b7280; margin-bottom: 6px; }
  .kpi-val { font-size: 20px; font-weight: 600; color: #111827; }
  .kpi-sub { font-size: 12px; color: #6b7280; margin-top: 4px; }
  .rec-list { margin: 0; padding-left: 18px; line-height: 1.8; }
  table { width: 100%; border-collapse: collapse; font-size: 13px; }
  th, td { border: 1px solid #e5e7eb; padding: 8px 10px; text-align: left; }
  th { background: #f3f4f6; font-weight: 600; color: #374151; }
  .row-critical { background: #fff1f0; color: #cf1322; }
  .row-warning { background: #fffbe6; color: #ad6800; }
  .cell-alert { color: #cf1322; font-weight: 700; }
  .footer { text-align: center; padding: 16px; font-size: 12px; color: #9ca3af; }
</style>
</head>
<body>
<div class="wrap">
  <div class="banner">
    <div>
      <h1>ZbxTable 自动化巡检与健康评分报告：{{.TaskName}}</h1>
      <div class="meta">报告编号 #{{.ReportID}} | 巡检时间：{{.StartTime}} | 耗时：{{.TotalTime}}s | 阈值：CPU {{.CpuThreshold}}% / 内存 {{.MemThreshold}}% / 磁盘 {{.DiskThreshold}}%</div>
    </div>
    <div class="score-box">
      <div class="score-num">{{.HealthScore}}</div>
      <div class="score-grade">综合评级：{{.GradeLabel}}</div>
    </div>
  </div>

  <div class="section">
    <h2>健康度四维得分明细</h2>
    <div class="kpi-grid">
      <div class="kpi-card">
        <div class="kpi-title">可用性健康度</div>
        <div class="kpi-val">{{.Breakdown.AvailabilityScore}} / {{.Breakdown.AvailabilityMax}}</div>
        <div class="kpi-sub">在线 {{.AvailableHosts}} 台 / 失联 {{.UnavailableHost}} 台</div>
      </div>
      <div class="kpi-card">
        <div class="kpi-title">资源水位健康度</div>
        <div class="kpi-val">{{.Breakdown.ResourceScore}} / {{.Breakdown.ResourceMax}}</div>
        <div class="kpi-sub">CPU超标 {{.HighCpuHosts}} | 内存超标 {{.HighMemHosts}} | 磁盘超标 {{.HighDiskHosts}}</div>
      </div>
      <div class="kpi-card">
        <div class="kpi-title">告警风险健康度</div>
        <div class="kpi-val">{{.Breakdown.AlarmScore}} / {{.Breakdown.AlarmMax}}</div>
        <div class="kpi-sub">活跃告警 {{.ActiveAlarms}} 条（高危 {{.CriticalAlarms}} 条）</div>
      </div>
      <div class="kpi-card">
        <div class="kpi-title">集群稳定度</div>
        <div class="kpi-val">{{.Breakdown.StabilityScore}} / {{.Breakdown.StabilityMax}}</div>
        <div class="kpi-sub">总主机 {{.TotalHosts}} 台 / 异常主机 {{.AbnormalHosts}} 台</div>
      </div>
    </div>
  </div>

  <div class="section">
    <h2>巡检结论与运维建议</h2>
    <ul class="rec-list">
      {{range .Recommendations}}
      <li>{{.}}</li>
      {{end}}
    </ul>
  </div>

  {{if .Hosts}}
  <div class="section">
    <h2>主机巡检明细（共 {{.TotalHosts}} 台）</h2>
    <table>
      <thead>
        <tr>
          <th>实例</th><th>主机名称</th><th>IP地址</th><th>类型</th><th>采集状态</th>
          <th>CPU</th><th>内存</th><th>最大磁盘</th><th>告警(高危)</th><th>得分</th><th>结论</th><th>异常明细</th>
        </tr>
      </thead>
      <tbody>
        {{range .Hosts}}
        <tr class="{{.RowClass}}">
          <td>{{.InstanceName}}</td>
          <td>{{.Name}}</td>
          <td>{{.IP}}</td>
          <td>{{.TypeName}}</td>
          <td>{{.AvailLabel}}</td>
          <td class="{{if .CpuHigh}}cell-alert{{end}}">{{.CpuText}}</td>
          <td class="{{if .MemHigh}}cell-alert{{end}}">{{.MemText}}</td>
          <td class="{{if .DiskHigh}}cell-alert{{end}}">{{.DiskText}}</td>
          <td>{{.ActiveAlarms}} ({{.CriticalAlarms}})</td>
          <td>{{.HostScore}}</td>
          <td>{{.StatusLabel}}</td>
          <td>{{.IssuesText}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>
  {{end}}

  {{if .Alarms}}
  <div class="section">
    <h2>未恢复活跃告警明细（前 {{len .Alarms}} 条）</h2>
    <table>
      <thead>
        <tr><th>实例</th><th>主机</th><th>IP</th><th>级别</th><th>告警描述</th><th>发生时间</th></tr>
      </thead>
      <tbody>
        {{range .Alarms}}
        <tr>
          <td>{{.InstanceName}}</td><td>{{.Hostname}}</td><td>{{.HostIP}}</td>
          <td>{{.Level}}</td><td>{{.Message}}</td><td>{{.OccurTime}}</td>
        </tr>
        {{end}}
      </tbody>
    </table>
  </div>
  {{end}}

  <div class="footer">本报告由 ZbxTable 自动化巡检中心生成</div>
</div>
</body>
</html>`
