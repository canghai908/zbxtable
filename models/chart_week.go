package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/components"
	"github.com/go-echarts/go-echarts/v2/opts"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	"zbxtable/utils"
)

//

func TaskWeekReport(m Report) error {
	Tend := time.Now()
	var task TaskLog
	task.ReportID = m.ID
	task.Name = m.Name
	task.Cycle = "week"
	task.StartTime = Tend
	tstart := Tend.Add(-120 * time.Hour)
	start := Tend.Add(-120 * time.Hour).Unix()
	end := Tend.Unix()
	StrStart := tstart.Format("2006-01-02 15:04:05")
	StrEnd := Tend.Format("2006-01-02 15:04:05")
	var itemlist []string
	err := utils.Mkdir(DowloadPath)
	if err != nil {
		taskend := time.Now()
		task.EndTime = taskend
		task.Status = Failed
		task.TotalTime = taskend.Unix() - Tend.Unix()
		task.Result = err.Error()
		_, err = task.Create()
		if err != nil {
			logs.Error(err)
			return err
		}
		logs.Error(err)
		return err
	}
	err = json.Unmarshal([]byte(m.Items), &itemlist)
	if err != nil {
		//写入日志
		taskend := time.Now()
		task.EndTime = taskend
		task.Status = Failed
		task.TotalTime = taskend.Unix() - Tend.Unix()
		task.Result = err.Error()
		_, err = task.Create()
		if err != nil {
			return err
		}
		return err
	}
	var bindlist []string
	if len(m.LinkBandWidth) == 0 {
		for i := 0; i < len(itemlist); i++ {
			bindlist = append(bindlist, "0")
		}
	}
	if strings.Contains(m.LinkBandWidth, ",") {
		bindlist = strings.Split(m.LinkBandWidth, ",")
	} else {
		bindlist = append(bindlist, m.LinkBandWidth)
	}
	if len(bindlist) < len(itemlist) {
		for i := 0; i < len(itemlist)-len(bindlist); i++ {
			//bindlist[k] = "0"
			bindlist = append(bindlist, "0")
		}
	}
	var filelist []string
	var ChartList []ChartData
	var OneChartData ChartData
	//plist := make(map[][]opts.LineData)
	for k, v := range itemlist {
		var bind float64
		var err error
		bind, err = strconv.ParseFloat(bindlist[k], 64)
		if err != nil {
			bind = 0
		}
		//获取item信息
		ItemInfo, _ := GetItemByID(v)
		//	fmt.Println(ItemInfo[0])
		hostinfo, err := GetHost(ItemInfo[0].Hostid)
		if err != nil {
			//写入日志
			taskend := time.Now()
			task.EndTime = taskend
			task.Status = Failed
			task.TotalTime = taskend.Unix() - Tend.Unix()
			task.Result = err.Error()
			_, err = task.Create()
			if err != nil {
				return err
			}
		}
		OneChartData.Host = hostinfo.Name
		OneChartData.IP = hostinfo.Interfaces
		OneChartData.Name = ItemInfo[0].Name
		OneChartData.Units = ItemInfo[0].Units
		p, err := GetHistoryByItemIDTTTT(ItemInfo[0].Itemid, ItemInfo[0].ValueType, start, end)
		if err != nil {
			//写入日志
			taskend := time.Now()
			task.EndTime = taskend
			task.Status = Failed
			task.TotalTime = taskend.Unix() - Tend.Unix()
			task.Result = err.Error()
			_, err = task.Create()
			if err != nil {
				return err
			}
			return err
		}
		xlsfilename, err := CreateHistoryReportXlsx(p, m.Name, hostinfo.Name, ItemInfo[0].Name,
			ItemInfo[0].Itemid, ItemInfo[0].Key, ItemInfo[0].Units, "week", StrStart, StrEnd)
		if err != nil {
			//写入日志
			taskend := time.Now()
			task.EndTime = taskend
			task.Status = Failed
			task.TotalTime = taskend.Unix() - Tend.Unix()
			task.Result = err.Error()
			_, err = task.Create()
			if err != nil {
				return err
			}
			return err
		}
		filelist = append(filelist, xlsfilename)
		var datelist []string
		var vallist []opts.LineData
		var vbllist []opts.LineData
		for _, vv := range p {
			//err := CreateHistoryXlsx(vv)
			tclock, _ := strconv.ParseInt(vv.Clock, 10, 64)
			date := time.Unix(tclock, 0)
			//fmt.Println(date.Format("2006-01-02 15:04:05"))
			datelist = append(datelist, date.Format("2006-01-02 15:04:05"))
			floaval, _ := strconv.ParseFloat(vv.Value, 64)
			vallist = append(vallist, opts.LineData{Value: floaval})
			vbllist = append(vbllist, opts.LineData{Value: bind * 1000 * 1000})
		}
		OneChartData.Start = StrStart
		OneChartData.End = StrEnd
		OneChartData.Date = datelist
		OneChartData.Data = vallist
		OneChartData.LinkBandWidth = vbllist
		ChartList = append(ChartList, OneChartData)
	}

	htmlname, err := ChartWeekHtml(m, ChartList)
	if err != nil {
		//写入日志
		taskend := time.Now()
		task.EndTime = taskend
		task.Status = Failed
		task.TotalTime = taskend.Unix() - Tend.Unix()
		task.Result = err.Error()
		_, err = task.Create()
		if err != nil {
			logs.Error(err)
			return err
		}
		logs.Error(err)
		return err
	}
	filelist = append(filelist, htmlname)
	dirdata := Tend.Format("2006-01-02_15_04_05")
	dirname := m.Name + "_week_" + dirdata + "/"
	Subject := "[周报]" + "[" + m.Name + "]" + "[" + time.Now().Format("2006-01-02") + "]"
	zipfilename := m.Name + "_week_" + dirdata + ".zip"

	err = utils.ZipFiles(DowloadPath+zipfilename, filelist, DowloadPath, dirname)
	if err != nil {
		//写入日志
		taskend := time.Now()
		task.EndTime = taskend
		task.Status = Failed
		task.TotalTime = taskend.Unix() - Tend.Unix()
		task.Result = err.Error()
		_, err = task.Create()
		if err != nil {
			logs.Error(err)
			return err
		}
		logs.Error(err)
		return err
	}
	//remote file
	for _, v := range filelist {
		err := os.Remove(v)
		if err != nil {
			//写入日志
			taskend := time.Now()
			task.EndTime = taskend
			task.Status = Failed
			task.TotalTime = taskend.Unix() - Tend.Unix()
			task.Result = err.Error()
			_, err = task.Create()
			if err != nil {
				logs.Error(err)
				return err
			}
		}
	}
	//如果邮件为空 记录日志 返回
	if m.Emails == "" {
		taskend := time.Now()
		task.EndTime = taskend
		task.Status = Success
		task.TotalTime = taskend.Unix() - Tend.Unix()
		task.Result = "执行成功"
		task.Files = zipfilename
		_, err = task.Create()
		if err != nil {
			return err
		}
		return nil
	}
	//email html templ
	byhtml, err := CreateMailTable(m, ChartList)
	if err != nil {
		//写入日志
		taskend := time.Now()
		task.EndTime = taskend
		task.Status = Failed
		task.TotalTime = taskend.Unix() - Tend.Unix()
		task.Result = err.Error()
		_, err = task.Create()
		if err != nil {
			return err
		}
		return err
	}

	//发邮件
	tolist := strings.Split(m.Emails, ",")
	err = Sendmail(tolist, Subject, zipfilename, byhtml)
	if err != nil {
		//写入日志
		taskend := time.Now()
		task.EndTime = taskend
		task.Status = Failed
		task.TotalTime = taskend.Unix() - Tend.Unix()
		task.Files = zipfilename
		task.Result = err.Error()
		_, err = task.Create()
		logs.Error(err)
		if err != nil {
			return err
		}
		return err
	}
	//写入日志
	taskend := time.Now()
	task.EndTime = taskend
	task.Status = Success
	task.TotalTime = taskend.Unix() - Tend.Unix()
	task.Result = "执行成功"
	task.Files = zipfilename
	_, err = task.Create()
	if err != nil {
		return err
	}

	return nil
}

func CreateWeekChart(data ChartData) *charts.Line {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTooltipOpts(opts.Tooltip{
			Show:      true,
			Trigger:   "axis",
			TriggerOn: "mousemove",
		}),
		charts.WithDataZoomOpts(
			opts.DataZoom{
				Type: "slider",
			},
		),
		charts.WithInitializationOpts(opts.Initialization{
			Width:  "1200px",
			Height: "600px",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:   true,
			Orient: "vertical",
			Left:   "auto"}),
		charts.WithYAxisOpts(opts.YAxis{
			AxisLabel: &opts.AxisLabel{
				Show:      true,
				Formatter: opts.FuncOpts(Formatter),
			},
		}),
		charts.WithToolboxOpts(opts.Toolbox{
			Show:   true,
			Orient: "horizontal",
			Left:   "right",
			Feature: &opts.ToolBoxFeature{
				SaveAsImage: &opts.ToolBoxFeatureSaveAsImage{
					Show: true, Title: "Save as image"},
				Restore: &opts.ToolBoxFeatureRestore{
					Show: true, Title: "Reset"},
			}}),
		charts.WithInitializationOpts(
			opts.Initialization{
				Theme: "shine"}),

		//Theme: types.ThemeWesteros}),
		charts.WithTitleOpts(opts.Title{
			Title:    data.Host,
			Subtitle: data.Name + "\n" + data.Start + "--" + data.End,
			Left:     "center",
			TitleStyle: &opts.TextStyle{
				FontSize: 20,
			},
			SubtitleStyle: &opts.TextStyle{
				FontSize: 12,
			},
		}))
	// Put data into instance
	line.SetXAxis(data.Date).
		AddSeries(data.Name, data.Data).
		AddSeries("带宽", data.LinkBandWidth).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	line.PageTitle = data.Name
	return line
}

func ChartWeekHtml(m Report, data []ChartData) (string, error) {
	page := components.NewPage()
	for _, v := range data {
		//fmt.Println(v.LinkBandWidth)
		page.AddCharts(
			CreateWeekChart(v),
		)
	}
	page.Initialization.AssetsHost = AssetsHost
	date := time.Now().Format("2006-01-02_15_04_05")
	page.PageTitle = m.Name
	filename := DowloadPath + m.Name + "_week_" + date + ".html"
	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	page.Render(io.MultiWriter(f))
	return filename, nil
}
