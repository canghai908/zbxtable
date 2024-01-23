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

const DowloadPath = "./download/"

type ChartData struct {
	Host          string          `json:"host"`
	IP            string          `json:"ip"`
	Name          string          `json:"name"`
	Units         string          `json:"units"`
	Start         string          `json:"start"`
	End           string          `json:"end"`
	Date          []string        `json:"date"`
	Data          []opts.LineData `json:"data"`
	LinkBandWidth []opts.LineData `json:"link_band_width"`
}

func TaskDayReport(m Report) error {
	Tend := time.Now()
	var task TaskLog
	task.ReportID = m.ID
	task.Name = m.Name
	task.Cycle = "day"
	task.StartTime = Tend
	tstart := Tend.Add(-24 * time.Hour)
	start := Tend.Add(-24 * time.Hour).Unix()
	end := Tend.Unix()
	StrStart := tstart.Format("2006-01-02 15:04:05")
	StrEnd := Tend.Format("2006-01-02 15:04:05")
	//创建目录
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
	itemsList := strings.Split(m.Items, ",")
	if len(itemsList) == 0 {
		taskend := time.Now()
		task.EndTime = taskend
		task.Status = Failed
		task.TotalTime = taskend.Unix() - Tend.Unix()
		task.Result = "监控项为空"
		_, err = task.Create()
		if err != nil {
			logs.Error(err)
			return err
		}
		logs.Error(err)
		return err
	}
	//获取bandwith 列表
	var bindlist []string
	//为空，全部填充为0
	if len(m.LinkBandWidth) == 0 {
		for i := 0; i < len(itemsList); {
			bindlist = append(bindlist, "0")
			i++
		}
	} else {
		bindlist = strings.Split(m.LinkBandWidth, ",")
		//少于
		if len(bindlist) < len(itemsList) {
			for i := 0; i <= len(itemsList)-len(bindlist)+1; {
				//bindlist[k] = "0"
				bindlist = append(bindlist, "0")
				i++
			}
		}
	}
	var filelist []string
	var ChartList []ChartData
	var OneChartData ChartData
	//plist := make(map[][]opts.LineData)
	for k, v := range itemsList {
		var bind float64
		var err error
		bind, err = strconv.ParseFloat(bindlist[k], 64)
		if err != nil {
			bind = 0
		}
		//获取item信息
		ItemInfo, err := GetItemByID(v)
		if err != nil {
			logs.Error(err)
			continue
		}
		//判断是否为空
		if len(ItemInfo) == 0 {
			logs.Error(err)
			continue
		}
		//	fmt.Println(ItemInfo[0])
		hostInfo, err := GetHost(ItemInfo[0].Hostid)
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
		OneChartData.Host = hostInfo.Name
		OneChartData.IP = hostInfo.Interfaces
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
		//生成xlsx文件
		xlsfilename, err := CreateHistoryReportXlsx(p, m.Name, hostInfo.Name, ItemInfo[0].Name,
			ItemInfo[0].Itemid, ItemInfo[0].Key, ItemInfo[0].Units, "day", StrStart, StrEnd)
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

	htmlname, err := Examples(m, ChartList)
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
	filelist = append(filelist, htmlname)
	dirdata := Tend.Format("2006-01-02_15_04_05")
	dirname := m.Name + "_day_" + dirdata + "/"
	Subject := "[日报]" + "[" + m.Name + "]" + "[" + time.Now().Format("2006-01-02") + "]"
	zipfilename := m.Name + "_day_" + dirdata + ".zip"
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
			return err
		}
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
		logs.Error(err)
		//写入日志
		taskend := time.Now()
		task.EndTime = taskend
		task.Status = Failed
		task.TotalTime = taskend.Unix() - Tend.Unix()
		task.Result = err.Error()
		task.Files = zipfilename
		_, err = task.Create()
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

var Formatter = `function(value) {
        return (value / (1024 * 1024)).toFixed(2) + 'Mbps';
    }`

func CreateDayChart(data ChartData) *charts.Line {
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
		//charts.WithToolboxOpts(
		//	opts.Toolbox{
		//		Show: true,
		//		Feature: &opts.ToolBoxFeature{
		//			DataZoom: &opts.ToolBoxFeatureDataZoom{
		//				Show:  true,
		//				Title: m,
		//			},
		//		},
		//	},
		//),
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
		//charts.WithLineStyleOpts(
		//	opts.LineStyle{Color: "#00FF00"})).
		//AddSeries("出流量", vcllist).
		////charts.WithLineStyleOpts(
		////	opts.LineStyle{Color: "#0000FF"})).
		//AddSeries("带宽", vbllist).
		//charts.WithLineStyleOpts(
		//	opts.LineStyle{Color: "#FF0000", Type: "dotted"})). \

		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true})) //charts.WithMarkLineNameTypeItemOpts(opts.MarkLineNameTypeItem{
	//	Name: "Average",
	//	Type: "average",
	//}),
	//charts.WithMarkLineNameTypeItemOpts(opts.MarkLineNameTypeItem{
	//	Name: "max",
	//	Type: "max",
	//}),

	line.PageTitle = data.Name
	return line
}

func Examples(m Report, data []ChartData) (string, error) {
	page := components.NewPage()
	for _, v := range data {
		//fmt.Println(v.LinkBandWidth)
		page.AddCharts(
			CreateDayChart(v),
		)
	}
	page.Initialization.AssetsHost = AssetsHost
	date := time.Now().Format("2006-01-02_15_04_05")
	page.PageTitle = m.Name
	filename := DowloadPath + m.Name + "_day_" + date + ".html"
	f, err := os.Create(filename)
	if err != nil {
		return "", err
	}
	defer f.Close()
	page.Render(io.MultiWriter(f))
	return filename, nil
}
