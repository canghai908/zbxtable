package model

import (
	"bytes"
	"compress/gzip"
	"crypto/tls"

	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/assets"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/signintech/gopdf"
	"github.com/xuri/excelize/v2"
)

// HostReportConfig 主机报表配置结构
type HostReportConfig struct {
	ZID     string   `json:"zid"`
	HostID  string   `json:"host_id"`
	ItemIDs []string `json:"item_ids"`
}

// ItemData 指标数据结构，用于生成多sheet Excel
type ItemData struct {
	HostInfo     Hosts
	ItemInfo     []Item
	HistoryData  []History
	InstanceName string
}

// HistoryStats 历史数据统计结果
type HistoryStats struct {
	Max   float64
	Min   float64
	Avg   float64
	Count int
}

// HasData 是否包含有效统计数据
func (s HistoryStats) HasData() bool {
	return s.Count > 0
}

// CalcHistoryStats 根据历史数据计算最大值、最小值、平均值
func CalcHistoryStats(historyData []History) HistoryStats {
	var stats HistoryStats
	var sum float64
	for _, h := range historyData {
		val, err := strconv.ParseFloat(h.Value, 64)
		if err != nil {
			continue
		}
		if stats.Count == 0 {
			stats.Max = val
			stats.Min = val
		} else {
			if val > stats.Max {
				stats.Max = val
			}
			if val < stats.Min {
				stats.Min = val
			}
		}
		sum += val
		stats.Count++
	}
	if stats.Count > 0 {
		stats.Avg = sum / float64(stats.Count)
	}
	return stats
}

// formatStatValue 格式化统计数值，保留两位小数并去除多余的零
func formatStatValue(v float64) string {
	s := strconv.FormatFloat(v, 'f', 2, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(s, "0")
		s = strings.TrimRight(s, ".")
	}
	return s
}

func buildHostReportSheetName(hostName, itemName string, existingSheetNames []string) string {
	sheetName := strings.TrimSpace(hostName) + "-" + strings.TrimSpace(itemName)
	sheetName = strings.NewReplacer(
		":", "_",
		"\\", "_",
		"/", "_",
		"?", "_",
		"*", "_",
		"[", "_",
		"]", "_",
	).Replace(sheetName)
	sheetName = strings.TrimSpace(sheetName)
	if sheetName == "" {
		sheetName = "Sheet"
	}

	baseRunes := []rune(sheetName)
	if len(baseRunes) > 31 {
		sheetName = string(baseRunes[:31])
	}

	originalSheetName := sheetName
	sheetIndex := 1
	for {
		exists := false
		for _, name := range existingSheetNames {
			if name == sheetName {
				exists = true
				break
			}
		}
		if !exists {
			return sheetName
		}

		suffix := strconv.Itoa(sheetIndex)
		maxBaseLen := 31 - len([]rune(suffix))
		trimmedRunes := []rune(originalSheetName)
		if len(trimmedRunes) > maxBaseLen {
			trimmedRunes = trimmedRunes[:maxBaseLen]
		}
		sheetName = string(trimmedRunes) + suffix
		sheetIndex++
	}
}

// TaskHostReport 生成主机报表
func TaskHostReport(m Report) error {
	return runTaskHostReport(m, nil)
}

func TaskHostReportWithTaskLog(m Report, task *TaskLog) error {
	return runTaskHostReport(m, task)
}

func runTaskHostReport(m Report, task *TaskLog) error {
	Tend := time.Now()

	if err := PrepareHostReportExecutionConfig(&m); err != nil {
		return err
	}

	if task == nil {
		task = &TaskLog{
			ReportID:  m.ID,
			Name:      m.Name,
			StartTime: Tend,
			Status:    Running,
			Result:    buildTaskLogResult(Running, 0, "等待执行"),
		}
		// 根据Cycle确定任务周期类型：day 或 week
		if strings.Contains(m.Cycle, "week") {
			task.Cycle = "week"
		} else if strings.Contains(m.Cycle, "day") {
			task.Cycle = "day"
		} else {
			task.Cycle = m.Cycle
		}
		if _, err := task.Create(); err != nil {
			return err
		}
	}
	updateProgress := func(progress int, detail string) {
		if task != nil && task.Id > 0 {
			if err := UpdateTaskLogProgress(int64(task.Id), progress, detail); err != nil {
				logger.Log.Error(err)
			}
		}
	}
	failTask := func(err error, detail string, files string) error {
		if task != nil && task.Id > 0 {
			if updateErr := FinishTaskLog(int64(task.Id), task.StartTime, Failed, detail, files); updateErr != nil {
				logger.Log.Error(updateErr)
			}
		}
		return err
	}

	updateProgress(2, "正在解析主机配置")

	// 解析主机和指标配置
	var hostConfigs []HostReportConfig
	if err := json.Unmarshal([]byte(m.HostIds), &hostConfigs); err != nil {
		return failTask(err, "主机配置解析失败: "+err.Error(), "")
	}

	if len(hostConfigs) == 0 {
		return failTask(fmt.Errorf("主机配置为空"), "主机配置为空", "")
	}
	updateProgress(8, "主机配置解析完成")

	// 确定时间范围
	var tstart, tend time.Time
	var start, end int64
	var StrStart, StrEnd string

	// 如果报表配置了开始和结束时间，使用配置的时间；否则根据周期计算
	if m.Start != nil && m.End != nil {
		tstart = *m.Start
		tend = *m.End
		start = tstart.Unix()
		end = tend.Unix()
		StrStart = tstart.Format("2006-01-02 15:04:05")
		StrEnd = tend.Format("2006-01-02 15:04:05")
	} else if strings.Contains(m.Cycle, "day") {
		tstart = Tend.Add(-24 * time.Hour)
		tend = Tend
		start = tstart.Unix()
		end = tend.Unix()
		StrStart = tstart.Format("2006-01-02 15:04:05")
		StrEnd = tend.Format("2006-01-02 15:04:05")
	} else if strings.Contains(m.Cycle, "week") {
		tstart = Tend.Add(-120 * time.Hour)
		tend = Tend
		start = tstart.Unix()
		end = tend.Unix()
		StrStart = tstart.Format("2006-01-02 15:04:05")
		StrEnd = tend.Format("2006-01-02 15:04:05")
	} else {
		tstart = Tend.Add(-24 * time.Hour)
		tend = Tend
		start = tstart.Unix()
		end = tend.Unix()
		StrStart = tstart.Format("2006-01-02 15:04:05")
		StrEnd = tend.Format("2006-01-02 15:04:05")
	}

	// 创建下载目录
	err := utils.Mkdir(DownloadPath)
	if err != nil {
		return failTask(err, err.Error(), "")
	}
	updateProgress(12, "已准备输出目录")

	var filelist []string
	var ChartList []ChartData

	// 收集所有指标数据，用于生成多sheet Excel
	var allItemsData []ItemData
	totalItemCount := 0
	for _, config := range hostConfigs {
		totalItemCount += len(config.ItemIDs)
	}
	if totalItemCount == 0 {
		totalItemCount = len(hostConfigs)
	}
	processedItemCount := 0
	lastProgress := 12

	// 遍历每个主机配置
	for _, v := range hostConfigs {
		// 检查实例ID是否存在
		if v.ZID == "" {
			logger.Log.Error("主机配置缺少实例ID")
			continue
		}

		// 获取该实例的 API 连接（使用 zid)
		id, err := strconv.Atoi(v.ZID)
		if err != nil {
			logger.Log.Error(err)
			continue
		}
		inst, err := GetAPIByZID(id)
		if err != nil {
			logger.Log.Error("获取实例API失败:", err)
			continue
		}

		// 获取主机信息（使用该实例的API）
		hostInfo, err := GetHostFromInstance(inst, v.HostID)
		if err != nil {
			logger.Log.Error("获取主机信息失败:", err)
			continue
		}

		// 遍历每个指标
		for _, itemID := range v.ItemIDs {
			// 获取指标信息（使用该实例的API）
			itemInfo, err := GetItemByIDFromInstance(inst, itemID)
			if err != nil || len(itemInfo) == 0 {
				logger.Log.Error("获取指标信息失败:", err)
				continue
			}

			// 获取历史数据（使用该实例的API）
			historyData, err := GetHistoryByItemIDFromInstance(inst, itemInfo[0].Itemid, itemInfo[0].ValueType, start, end)
			if err != nil {
				logger.Log.Error("获取历史数据失败:", err)
				processedItemCount++
				currentProgress := 12 + (processedItemCount * 58 / totalItemCount)
				if currentProgress > lastProgress {
					lastProgress = currentProgress
					updateProgress(currentProgress, fmt.Sprintf("正在采集指标数据 %d/%d", processedItemCount, totalItemCount))
				}
				continue
			}

			// 保存数据用于生成多sheet Excel
			allItemsData = append(allItemsData, ItemData{
				HostInfo:     hostInfo,
				ItemInfo:     itemInfo,
				HistoryData:  historyData,
				InstanceName: inst.Name, // 添加实例名称
			})

			// 准备图表数据
			var datelist []string
			var vallist []opts.LineData
			for _, vv := range historyData {
				tclock, _ := strconv.ParseInt(vv.Clock, 10, 64)
				date := time.Unix(tclock, 0)
				datelist = append(datelist, date.Format("2006-01-02 15:04:05"))
				floaval, _ := strconv.ParseFloat(vv.Value, 64)
				vallist = append(vallist, opts.LineData{Value: floaval})
			}

			// 根据时间范围内的历史数据计算最大、最小、平均值
			stats := CalcHistoryStats(historyData)

			chartData := ChartData{
				Host:         hostInfo.Name,
				IP:           hostInfo.Interfaces,
				Name:         itemInfo[0].Name,
				Units:        itemInfo[0].Units,
				Start:        StrStart,
				End:          StrEnd,
				Date:         datelist,
				Data:         vallist,
				ItemID:       itemInfo[0].Itemid,
				InstanceName: inst.Name, // 添加实例名称
				Instance:     inst,      // 添加实例对象
				Max:          stats.Max,
				Min:          stats.Min,
				Avg:          stats.Avg,
				HasStats:     stats.HasData(),
			}
			ChartList = append(ChartList, chartData)
			processedItemCount++
			currentProgress := 12 + (processedItemCount * 58 / totalItemCount)
			if currentProgress > lastProgress {
				lastProgress = currentProgress
				updateProgress(currentProgress, fmt.Sprintf("正在采集指标数据 %d/%d", processedItemCount, totalItemCount))
			}
		}
	}

	// 如果有多个指标，生成包含多个sheet的Excel文件
	updateProgress(75, "正在生成 Excel 报表")
	if len(allItemsData) > 0 {
		xlsfilename, err := CreateMultiSheetHostReportXlsx(allItemsData, m.Name, m.Cycle, StrStart, StrEnd)
		if err != nil {
			logger.Log.Error("生成多sheet Excel报表失败:", err)
		} else {
			filelist = append(filelist, xlsfilename)
		}
	}

	if len(ChartList) == 0 {
		return failTask(fmt.Errorf("没有可用的数据"), "没有可用的数据", "")
	}

	// 生成HTML报表
	updateProgress(82, "正在生成 HTML 报表")
	htmlname, err := CreateHostReportHTML(m, ChartList)
	if err != nil {
		return failTask(err, err.Error(), "")
	}
	filelist = append(filelist, htmlname)

	// 复制静态资源文件到HTML同目录，并添加到文件列表
	updateProgress(86, "正在整理静态资源")
	assetFiles, err := assets.CopyAssetsToDir(DownloadPath)
	if err != nil {
		logger.Log.Error("Failed to copy assets:", err)
	} else {
		filelist = append(filelist, assetFiles...)
	}

	// 生成PDF报表
	updateProgress(90, "正在生成 PDF 报表")
	pdfname, err := CreateHostReportPDF(m, ChartList, StrStart, StrEnd)
	if err != nil {
		logger.Log.Error("生成PDF报表失败:", err)
	} else {
		filelist = append(filelist, pdfname)
	}

	// 打包文件
	dirdata := Tend.Format("2006-01-02_15_04_05")
	// 根据Cycle确定文件名后缀：day、week 或 realtime
	cycleType := "day" // 默认使用day
	if strings.Contains(m.Cycle, "week") {
		cycleType = "week"
	} else if strings.Contains(m.Cycle, "day") {
		cycleType = "day"
	} else if strings.Contains(m.Cycle, "realtime") {
		cycleType = "realtime"
	}
	// dirname 应该和 zipfilename 保持一致（都包含 _host_）
	dirname := m.Name + "_host_" + cycleType + "_" + dirdata + "/"
	Subject := "[主机报表]" + "[" + m.Name + "]" + "[" + time.Now().Format("2006-01-02") + "]"
	zipfilename := m.Name + "_host_" + cycleType + "_" + dirdata + ".zip"

	updateProgress(95, "正在打包报表文件")
	err = utils.ZipFiles(DownloadPath+zipfilename, filelist, DownloadPath, dirname)
	if err != nil {
		return failTask(err, err.Error(), "")
	}

	// 清理临时文件
	for _, v := range filelist {
		_ = os.Remove(v)
	}

	// 发送邮件
	if m.Emails != "" {
		updateProgress(98, "正在发送邮件")
		byhtml, err := CreateHostMailTable(m, ChartList, StrStart, StrEnd)
		if err != nil {
			logger.Log.Error("生成邮件内容失败:", err)
		} else {
			tolist := strings.Split(m.Emails, ",")
			err = Sendmail(tolist, Subject, zipfilename, byhtml)
			if err != nil {
				logger.Log.Error("发送邮件失败:", err)
				return failTask(err, "发送邮件失败: "+err.Error(), zipfilename)
			}
		}
	}

	// 记录任务日志
	updateProgress(100, "执行完成")
	if err := FinishTaskLog(int64(task.Id), task.StartTime, Success, "执行成功", zipfilename); err != nil {
		logger.Log.Error(err)
	}

	return nil
}

// CreateHostReportHTML 创建主机报表HTML
func CreateHostReportHTML(m Report, data []ChartData) (string, error) {
	page := components.NewPage()
	for _, v := range data {
		page.AddCharts(CreateHostChart(v))
	}
	// 使用本地相对路径
	page.Initialization.AssetsHost = assets.GetLocalAssetsHost()
	date := time.Now().Format("2006-01-02_15_04_05")
	page.PageTitle = m.Name
	filename := DownloadPath + m.Name + "_host_" + date + ".html"
	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	page.Render(io.MultiWriter(f))

	return filename, nil
}

// buildChartSubtitle 构建图表副标题，附带最大/最小/平均统计
func buildChartSubtitle(data ChartData) string {
	subtitle := data.Name + "\n" + data.Start + "--" + data.End
	if data.HasStats {
		unit := strings.TrimSpace(data.Units)
		statLine := fmt.Sprintf("最大: %s   最小: %s   平均: %s",
			formatStatValue(data.Max), formatStatValue(data.Min), formatStatValue(data.Avg))
		if unit != "" {
			statLine += " " + unit
		}
		subtitle += "\n" + statLine
	}
	return subtitle
}

// CreateHostChart 创建主机图表
func CreateHostChart(data ChartData) *charts.Line {
	line := charts.NewLine()
	// 构建标题，包含实例名称
	titleText := data.Host
	if data.InstanceName != "" {
		titleText = "[" + data.InstanceName + "] " + data.Host
	}
	line.SetGlobalOptions(
		charts.WithTooltipOpts(opts.Tooltip{
			Show:      true,
			Trigger:   "axis",
			TriggerOn: "mousemove",
		}),
		charts.WithDataZoomOpts(opts.DataZoom{Type: "slider"}),
		charts.WithInitializationOpts(opts.Initialization{
			Width:           "1200px",
			Height:          "600px",
			Theme:           "white",   // 使用白色主题
			BackgroundColor: "#ffffff", // 白色背景
		}),
		// 下移绘图区，为标题/副标题（含统计信息）留出空间，避免重叠
		charts.WithGridOpts(opts.Grid{
			Top:          "22%",
			ContainLabel: true,
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:   true,
			Orient: "vertical",
			Left:   "auto",
		}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show:   true,
			Orient: "horizontal",
			Left:   "right",
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show: true, Title: "Save as image",
				},
				Restore: &opts.ToolBoxFeatureRestore{
					Show: true, Title: "Reset",
				},
			},
		}),
		charts.WithTitleOpts(opts.Title{
			Title:         titleText,
			Subtitle:      buildChartSubtitle(data),
			Left:          "center",
			TitleStyle:    &opts.TextStyle{FontSize: 20, Color: "#333"},
			SubtitleStyle: &opts.TextStyle{FontSize: 12, Color: "#666"},
		}),
		// 设置颜色为蓝色系，替代红色
		charts.WithColorsOpts(opts.Colors{"#5470c6"}),
	)
	line.SetXAxis(data.Date).AddSeries(data.Name, data.Data).
		SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{Smooth: true}),
			charts.WithLineStyleOpts(opts.LineStyle{
				Color: "#5470c6", // 蓝色线条
				Width: 2,
			}),
			charts.WithItemStyleOpts(opts.ItemStyle{
				Color: "#5470c6", // 蓝色数据点
			}),
		)
	line.PageTitle = data.Name
	return line
}

// GetItemChartImageFromInstance 从指定Zabbix实例获取item的图表图片
func GetItemChartImageFromInstance(inst *APIInstance, itemID, start, end string) (gopdf.ImageHolder, error) {
	if inst == nil {
		return nil, fmt.Errorf("实例参数为空")
	}

	if inst.WebURL == "" {
		return nil, fmt.Errorf("实例 %s 未配置 WebURL", inst.Name)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 使用实例的 JAR（如果有），否则使用全局 JAR
	jar := inst.JAR
	if jar == nil {
		jar = JAR
	}

	client1 := &http.Client{
		Transport: tr,
		Jar:       jar,
		Timeout:   99999999999992,
	}

	// 使用实例的 WebURL 获取 item 的图表
	imgurl := inst.WebURL + "/chart.php?"
	data := url.Values{}
	URL, err := url.Parse(imgurl)
	if err != nil {
		return nil, err
	}

	// 将时间字符串转换为 Unix 时间戳
	loc, _ := time.LoadLocation("Asia/Shanghai")
	startTime, err := time.ParseInLocation("2006-01-02 15:04:05", start, loc)
	if err != nil {
		return nil, fmt.Errorf("解析开始时间失败: %v", err)
	}
	endTime, err := time.ParseInLocation("2006-01-02 15:04:05", end, loc)
	if err != nil {
		return nil, fmt.Errorf("解析结束时间失败: %v", err)
	}

	data.Set("itemids[]", itemID)
	data.Set("from", startTime.Format("2006-01-02 15:04:05"))
	data.Set("to", endTime.Format("2006-01-02 15:04:05"))
	data.Set("width", "800")
	data.Set("height", "200")

	URL.RawQuery = data.Encode()
	urlPath := URL.String()
	request, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return nil, err
	}

	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	request.Header.Add("Accept-Encoding", "gzip, deflate")
	request.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
	request.Header.Add("Connection", "keep-alive")
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")

	response, err := client1.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == 200 {
		var reader io.Reader
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(response.Body)
			if err != nil {
				return nil, err
			}
		default:
			reader = response.Body
		}

		imgHolder, err := gopdf.ImageHolderByReader(reader)
		if err != nil {
			return nil, err
		}
		return imgHolder, nil
	}

	return nil, fmt.Errorf("获取图表失败，状态码: %d", response.StatusCode)
}

// CreateHostReportPDF 创建主机报表PDF
func CreateHostReportPDF(m Report, data []ChartData, start, end string) (string, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595.28, H: 841.89}})

	// 必须先添加页面才能写入内容
	pdf.AddPage()

	// 添加字体，尝试多个可能的路径
	fontPaths := []string{"./msty.ttf", "msty.ttf", "fonts/msty.ttf", "./assets/fonts/msty.ttf"}
	var fontErr error
	var fontAdded bool

	for _, fontPath := range fontPaths {
		err := pdf.AddTTFFont("msty", fontPath)
		if err == nil {
			fontAdded = true
			break
		}
		fontErr = err
	}

	if !fontAdded {
		logger.Log.Error("无法添加字体文件，尝试的路径:", fontPaths, "最后错误:", fontErr)
		return "", fmt.Errorf("无法添加字体文件: %v", fontErr)
	}

	// 设置字体
	err := pdf.SetFont("msty", "", 12)
	if err != nil {
		logger.Log.Error("设置字体失败:", err)
		return "", fmt.Errorf("设置字体失败: %v", err)
	}

	// 添加标题
	pdf.SetX(10)
	pdf.SetY(20)
	pdf.Cell(nil, "报表名称: "+m.Name)
	pdf.SetX(10)
	pdf.SetY(35)
	pdf.Cell(nil, "报表周期: "+start+" -- "+end)
	pdf.Line(10, 45, 585, 45)

	// 添加图表图片
	yPos := 60.0
	for i, chartData := range data {
		// 每页最多显示2个图表，如果超过则添加新页面
		if i > 0 && i%2 == 0 {
			pdf.AddPage()
			// 重新添加页面标题
			pdf.SetX(10)
			pdf.SetY(20)
			pdf.Cell(nil, "报表名称: "+m.Name)
			pdf.SetX(10)
			pdf.SetY(35)
			pdf.Cell(nil, "报表周期: "+start+" -- "+end)
			pdf.Line(10, 45, 585, 45)
			yPos = 60.0
		}

		// 添加图表标题
		pdf.SetX(10)
		pdf.SetY(yPos)
		// 构建标题，包含实例名称
		titleText := fmt.Sprintf("主机: %s | 指标: %s (%s)", chartData.Host, chartData.Name, chartData.Units)
		if chartData.InstanceName != "" {
			titleText = fmt.Sprintf("实例: %s | 主机: %s | 指标: %s (%s)", chartData.InstanceName, chartData.Host, chartData.Name, chartData.Units)
		}
		pdf.Cell(nil, titleText)

		// 在标题下方添加统计信息（最大/最小/平均）
		if chartData.HasStats {
			pdf.SetX(10)
			pdf.SetY(yPos + 14)
			statText := fmt.Sprintf("最大: %s | 最小: %s | 平均: %s %s",
				formatStatValue(chartData.Max), formatStatValue(chartData.Min),
				formatStatValue(chartData.Avg), strings.TrimSpace(chartData.Units))
			pdf.Cell(nil, strings.TrimSpace(statText))
			yPos += 14
		}

		// 获取图表图片
		if chartData.ItemID != "" && chartData.Instance != nil {
			// 先检查实例是否配置了用户名和密码
			zbxInstance, err := GetZabbixInstanceByZID(chartData.Instance.ZID)
			canGetChart := false
			isTokenOnly := false

			if err == nil && zbxInstance != nil {
				// 判断是否只配置了Token而没有配置用户名和密码
				if (zbxInstance.User == "" || zbxInstance.Pass == "") && zbxInstance.Token != "" {
					isTokenOnly = true
					canGetChart = false
				} else if zbxInstance.User != "" && zbxInstance.Pass != "" {
					canGetChart = true
				}
			}

			// 如果只配置了Token，直接显示提示，不尝试获取图表
			if isTokenOnly {
				logger.Log.Warn(fmt.Sprintf("实例 %s 只配置了Token认证，无法获取图表图片 (ItemID: %s)", chartData.InstanceName, chartData.ItemID))
				pdf.SetX(10)
				pdf.SetY(yPos + 20)
				pdf.Cell(nil, fmt.Sprintf("提示: 实例 %s 使用Token认证方式，无法获取图表图片。", chartData.InstanceName))
				pdf.SetX(10)
				pdf.SetY(yPos + 35)
				pdf.Cell(nil, "请配置实例的用户名和密码以获取图表图片。")
				yPos += 70
			} else if canGetChart {
				// 尝试获取图表图片
				imgHolder, imgErr := GetItemChartImageFromInstance(chartData.Instance, chartData.ItemID, start, end)
				if imgErr == nil && imgHolder != nil {
					// 在标题下方添加图表图片
					imageY := yPos + 20
					// 使用图片原始尺寸，位置在页面左侧，宽度限制在页面内
					pdf.ImageByHolder(imgHolder, 10, imageY, nil)
					// 估算图片高度（根据请求的height=200，加上一些边距）
					estimatedImageHeight := 250.0
					yPos = imageY + estimatedImageHeight + 20
				} else {
					// 如果获取图表失败，在 PDF 中输出提示信息
					logger.Log.Warn(fmt.Sprintf("获取图表图片失败 (ItemID: %s, Instance: %s): %v", chartData.ItemID, chartData.InstanceName, imgErr))
					pdf.SetX(10)
					pdf.SetY(yPos + 20)
					pdf.Cell(nil, fmt.Sprintf("提示: 无法从实例 %s 获取该指标的图表图片。", chartData.InstanceName))
					yPos += 60
				}
			} else {
				// 其他情况（无法获取实例信息等）
				pdf.SetX(10)
				pdf.SetY(yPos + 20)
				pdf.Cell(nil, fmt.Sprintf("提示: 无法从实例 %s 获取该指标的图表图片。", chartData.InstanceName))
				yPos += 60
			}
		} else {
			// 没有有效的 ItemID 或实例，给出提示
			pdf.SetX(10)
			pdf.SetY(yPos + 20)
			if chartData.Instance == nil {
				pdf.Cell(nil, "提示: 该指标缺少实例信息，无法获取图表图片。")
			} else {
				pdf.Cell(nil, "提示: 该指标缺少 ItemID，无法获取图表图片。")
			}
			yPos += 60
		}
	}

	// 保存PDF
	date := time.Now().Format("2006-01-02_15_04_05")
	filename := DownloadPath + m.Name + "_host_" + date + ".pdf"
	var b bytes.Buffer
	err = pdf.Write(&b)
	if err != nil {
		return "", err
	}

	err = os.WriteFile(filename, b.Bytes(), 0644)
	if err != nil {
		return "", err
	}

	return filename, nil
}

// CreateHostMailTable 创建主机报表邮件内容
func CreateHostMailTable(m Report, data []ChartData, start, end string) ([]byte, error) {
	type MailData struct {
		ReportName string
		Start      string
		End        string
		TableInfo  []struct {
			InstanceName string
			Host         string
			IP           string
			ItemName     string
			Units        string
			Max          string
			Min          string
			Avg          string
		}
	}

	mailData := MailData{
		ReportName: m.Name,
		Start:      start,
		End:        end,
	}

	for _, v := range data {
		maxStr, minStr, avgStr := "--", "--", "--"
		if v.HasStats {
			maxStr = formatStatValue(v.Max)
			minStr = formatStatValue(v.Min)
			avgStr = formatStatValue(v.Avg)
		}
		mailData.TableInfo = append(mailData.TableInfo, struct {
			InstanceName string
			Host         string
			IP           string
			ItemName     string
			Units        string
			Max          string
			Min          string
			Avg          string
		}{
			InstanceName: v.InstanceName,
			Host:         v.Host,
			IP:           v.IP,
			ItemName:     v.Name,
			Units:        v.Units,
			Max:          maxStr,
			Min:          minStr,
			Avg:          avgStr,
		})
	}

	tmpl, err := template.New("hostReport").Parse(htmlHostReport)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, mailData)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// CreateMultiSheetHostReportXlsx 创建包含多个sheet的主机报表Excel文件
func CreateMultiSheetHostReportXlsx(itemsData []ItemData, reportName, cycle, start, end string) (string, error) {
	xlsx := excelize.NewFile()

	// 获取默认Sheet1的名称
	defaultSheetName := xlsx.GetSheetName(0)
	firstSheetCreated := false

	// 为每个指标创建一个sheet
	for idx, itemData := range itemsData {
		if len(itemData.ItemInfo) == 0 {
			continue
		}

		itemInfo := itemData.ItemInfo[0]
		hostInfo := itemData.HostInfo

		sheetName := buildHostReportSheetName(hostInfo.Name, itemInfo.Name, xlsx.GetSheetList())

		// 创建新的sheet（如果是第一个，使用默认的Sheet1）
		var index int
		if !firstSheetCreated {
			// 重命名默认的Sheet1
			xlsx.SetSheetName(defaultSheetName, sheetName)
			index = 0
			firstSheetCreated = true
		} else {
			// 创建新的sheet
			index = xlsx.NewSheet(sheetName)
		}

		// 设置列宽
		xlsx.SetColWidth(sheetName, "A", "B", 20)
		xlsx.SetColWidth(sheetName, "B", "B", 30)

		// 写入表头信息
		xlsx.SetCellValue(sheetName, "A1", "实例名称")
		xlsx.SetCellValue(sheetName, "B1", itemData.InstanceName)
		xlsx.SetCellValue(sheetName, "A2", "主机名称")
		xlsx.SetCellValue(sheetName, "B2", hostInfo.Name)
		xlsx.SetCellValue(sheetName, "A3", "指标名称")
		xlsx.SetCellValue(sheetName, "B3", itemInfo.Name)
		xlsx.SetCellValue(sheetName, "A4", "指标ID")
		xlsx.SetCellValue(sheetName, "B4", itemInfo.Itemid)
		xlsx.SetCellValue(sheetName, "A5", "指标Key")
		xlsx.SetCellValue(sheetName, "B5", itemInfo.Key)
		xlsx.SetCellValue(sheetName, "A6", "开始时间")
		xlsx.SetCellValue(sheetName, "B6", start)
		xlsx.SetCellValue(sheetName, "A7", "结束时间")
		xlsx.SetCellValue(sheetName, "B7", end)

		// 写入统计信息（最大/最小/平均），紧跟在结束时间行之后，根据时间范围内的历史数据计算
		stats := CalcHistoryStats(itemData.HistoryData)
		unitSuffix := ""
		if u := strings.TrimSpace(itemInfo.Units); u != "" {
			unitSuffix = "(" + u + ")"
		}
		maxStr, minStr, avgStr := "--", "--", "--"
		if stats.HasData() {
			maxStr = formatStatValue(stats.Max)
			minStr = formatStatValue(stats.Min)
			avgStr = formatStatValue(stats.Avg)
		}
		xlsx.SetCellValue(sheetName, "A8", "最大值"+unitSuffix)
		xlsx.SetCellValue(sheetName, "B8", maxStr)
		xlsx.SetCellValue(sheetName, "A9", "最小值"+unitSuffix)
		xlsx.SetCellValue(sheetName, "B9", minStr)
		xlsx.SetCellValue(sheetName, "A10", "平均值"+unitSuffix)
		xlsx.SetCellValue(sheetName, "B10", avgStr)

		// 数据表头与数据起始行（统计信息后空一行）
		const dataHeaderRow = 12
		const dataStartRow = 13

		// 数据样式设置
		stylecenter, err := xlsx.NewStyle(`{"alignment":{"horizontal":"center"}}`)
		if err != nil {
			logger.Log.Error("创建样式失败:", err)
		}

		lea := len(itemData.HistoryData)
		// 设置单元格对齐方式
		if lea > 0 {
			lastRow := strconv.Itoa(dataStartRow + lea - 1)
			xlsx.SetCellStyle(sheetName, "A"+strconv.Itoa(dataHeaderRow), "A"+lastRow, stylecenter)
			xlsx.SetCellStyle(sheetName, "B"+strconv.Itoa(dataHeaderRow), "B"+lastRow, stylecenter)
		}

		// 写入数据表头
		xlsx.SetCellValue(sheetName, "A"+strconv.Itoa(dataHeaderRow), "时间")
		xlsx.SetCellValue(sheetName, "B"+strconv.Itoa(dataHeaderRow), "数值("+itemInfo.Units+")")

		// 写入历史数据
		for k, v := range itemData.HistoryData {
			loc, _ := time.LoadLocation("Asia/Shanghai")
			timeint64, _ := strconv.ParseInt(v.Clock, 10, 64)
			TimeUnix := time.Unix(timeint64, 0).In(loc)
			StrTime := TimeUnix.Format("2006-01-02 15:04:05")
			rowStr := strconv.Itoa(k + dataStartRow)
			xlsx.SetCellValue(sheetName, "A"+rowStr, StrTime)
			xlsx.SetCellValue(sheetName, "B"+rowStr, v.Value)
		}

		// 如果是第一个sheet，设置为活动sheet
		if idx == 0 {
			xlsx.SetActiveSheet(index)
		}
	}

	// 保存文件
	StrDate := time.Now().Format("2006-01-02_15_04_05")
	filename := DownloadPath + reportName + "_host_multi_" + StrDate + ".xlsx"
	if err := xlsx.SaveAs(filename); err != nil {
		return "", err
	}

	return filename, nil
}

var htmlHostReport = `<div>
<div style="font: Verdana normal 14px; color: #000">
    <div style="position: relative">
        <div class="eml-w eml-w-sys-layout">
            <div style="font-size: 0px">
                <div class="eml-w-sys-line">
                    <div class="eml-w-sys-line-left"></div>
                    <div class="eml-w-sys-line-right"></div>
                </div>
            </div>
            <div class="eml-w-sys-content">
                <div class="dragArea gen-group-list">
                    <div class="gen-item">
                        <div style="padding: 0px">
                            <div class="eml-w-phase-normal-16">
                                报表名称: {{.ReportName}}
                            </div>
                            <div class="eml-w-phase-normal-16">
                                报表周期：{{.Start}}--{{.End}}
                            </div>
                        </div>
                    </div>
                    <div class="gen-item">
                        <table border="1" style="width: 100%; border-collapse: collapse;">
                            <caption style="font-weight: bold; padding: 10px;">主机监控指标报表</caption>
                            <tr>
                                <th>实例名称</th>
                                <th>主机名称</th>
                                <th>IP地址</th>
                                <th>指标名称</th>
                                <th>单位</th>
                                <th>最大值</th>
                                <th>最小值</th>
                                <th>平均值</th>
                            </tr>
                            {{range .TableInfo}}
                            <tr>
                                <td>{{.InstanceName}}</td>
                                <td>{{.Host}}</td>
                                <td>{{.IP}}</td>
                                <td>{{.ItemName}}</td>
                                <td>{{.Units}}</td>
                                <td>{{.Max}}</td>
                                <td>{{.Min}}</td>
                                <td>{{.Avg}}</td>
                            </tr>
                            {{end}}
                        </table>
                    </div>
                </div>
            </div>
            <div class="eml-w-sys-footer">本报表由ZbxTable自动生成</div>
        </div>
    </div>
</div>
</div>
<style>
    .eml-w .eml-w-phase-normal-16 {
        color: #2b2b2b;
        font-size: 16px;
        line-height: 1.75;
    }
    .eml-w-sys-layout {
        background: #fff;
        box-shadow: 0 2px 8px 0 rgba(0, 0, 0, 0.2);
        border-radius: 4px;
        margin: 50px auto;
        max-width: 800px;
        overflow: hidden;
    }
    .eml-w-sys-line-left {
        display: inline-block;
        width: 88%;
        background: #2984ef;
        height: 3px;
    }
    .eml-w-sys-line-right {
        display: inline-block;
        width: 11.5%;
        height: 3px;
        background: #8bd5ff;
        margin-left: 1px;
    }
    .eml-w-sys-content {
        position: relative;
        padding: 20px 50px 0;
        min-height: 16px;
        word-break: break-all;
    }
    .eml-w-sys-footer {
        font-weight: 500;
        font-size: 12px;
        color: #bebebe;
        letter-spacing: 0.5px;
        padding: 0 0 30px 50px;
        margin-top: 60px;
    }
    .eml-w {
        font-family: Helvetica Neue, Arial, PingFang SC, Hiragino Sans GB, STHeiti,
            Microsoft YaHei, sans-serif;
        -webkit-font-smoothing: antialiased;
        color: #2b2b2b;
        font-size: 14px;
        line-height: 1.75;
    }
    table {
        border-collapse: collapse;
        width: 100%;
    }
    th, td {
        border: 1px solid #ddd;
        padding: 8px;
        text-align: left;
    }
    th {
        background-color: #f2f2f2;
        font-weight: bold;
    }
</style>`
