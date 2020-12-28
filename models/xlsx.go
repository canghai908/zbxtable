package models

import (
	"bytes"
	"fmt"
	"github.com/astaxie/beego/logs"
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
)

//Crt excel table
func Crt(Filedata []FileSystemDataALL, host, itemtype string, start, end int64) ([]byte, error) {

	loc, _ := time.LoadLocation("Asia/Shanghai")
	StartUnix := time.Unix(start, 0).In(loc)
	StrStart := StartUnix.Format("2006-01-02 15:04:05")

	EndUnix := time.Unix(end, 0).In(loc)
	StrEnd := EndUnix.Format("2006-01-02 15:04:05")

	//dataname string
	//var dataname, sourcename, vunit string
	var sourcename, vunit string
	var vfun1, vfun2 float64
	switch itemtype {
	case "cpu":
		// dataname = "cpu使用率"
		sourcename = "CPU"
		vunit = "%"
		vfun1 = 1
		vfun2 = 1
	case "mem":
		// dataname = "内存使用"
		// sourcename = "CPU"
		vunit = "MB"
		vfun1 = 1024
		vfun2 = 1024
	case "disk":
		// dataname = "磁盘空间"
		sourcename = "挂载点"
		vfun1 = 1024
		vfun2 = 1024
		vunit = "MB"
	case "net_in":
		// dataname = "网卡流量"
		sourcename = "网卡"
		vunit = "KB"
		vfun1 = 1024
		vfun2 = 1024
	case "net_out":
		// dataname = "网卡流量"
		sourcename = "网卡"
		vunit = "KB"
		vfun1 = 1024
		vfun2 = 1024
	default:
		sourcename = ""
		vunit = "KB"
		vfun1 = 1
		vfun2 = 1
	}

	xlsx := excelize.NewFile()
	// 创建一个工作表
	index := xlsx.NewSheet("Sheet1")
	//设置列宽
	xlsx.SetColWidth("Sheet1", "A", "A", 20)
	xlsx.SetColWidth("Sheet1", "B", "D", 15)

	//表头设计
	//主机名
	xlsx.SetCellValue("Sheet1", "A1", "主机名称")
	xlsx.SetCellValue("Sheet1", "B1", host)
	//指标类型
	xlsx.SetCellValue("Sheet1", "A2", "指标类型")
	xlsx.SetCellValue("Sheet1", "B2", itemtype)
	//开始时间
	xlsx.SetCellValue("Sheet1", "A3", "开始时间")
	xlsx.SetCellValue("Sheet1", "B3", StrStart)
	//结束时间
	xlsx.SetCellValue("Sheet1", "A4", "结束时间")
	xlsx.SetCellValue("Sheet1", "B4", StrEnd)

	//数据样式设置
	stylecenter, err := xlsx.NewStyle(`{"alignment":{"horizontal":"center"}}`)
	if err != nil {
		fmt.Println("创建样式失败", err)
	}
	styleleft, err := xlsx.NewStyle(`{"alignment":{"horizontal":"left"}}`)
	if err != nil {
		fmt.Println("创建样式失败", err)
	}
	lea := len(Filedata[0].FileSystemDataADD)
	//设置单元格对其方式
	for i := 0; i < 5; i++ {
		xlsx.SetCellStyle("Sheet1", "A5", "A"+strconv.Itoa(lea+9), stylecenter)
		xlsx.SetCellStyle("Sheet1", "B5", "B5", stylecenter)
		xlsx.SetCellStyle("Sheet1", "B6", "B"+strconv.Itoa(lea+9), styleleft)
		xlsx.SetCellStyle("Sheet1", "C5", "C5", stylecenter)
		xlsx.SetCellStyle("Sheet1", "C6", "C"+strconv.Itoa(lea+9), styleleft)
		xlsx.SetCellStyle("Sheet1", "D5", "D5", stylecenter)
		xlsx.SetCellStyle("Sheet1", "D6", "D"+strconv.Itoa(lea+9), styleleft)
	}
	//遍历数据
	for k, v := range Filedata {
		//数据分类遍历
		xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(7+(lea*k)), sourcename)
		xlsx.SetCellValue("Sheet1", "B"+strconv.Itoa(7+(lea*k)), v.MountPoint)
		xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(8+(lea*k)), "时间")
		xlsx.SetCellValue("Sheet1", "B"+strconv.Itoa(8+(lea*k)), "平均"+"("+vunit+")")
		xlsx.SetCellValue("Sheet1", "C"+strconv.Itoa(8+(lea*k)), "最大"+"("+vunit+")")
		xlsx.SetCellValue("Sheet1", "D"+strconv.Itoa(8+(lea*k)), "最小"+"("+vunit+")")
		//数据具体数据遍历
		for kk, vv := range Filedata[k].FileSystemDataADD {
			xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(kk+9+(lea*k)), vv.Clock)
			floatValueAvg, _ := strconv.ParseFloat(vv.ValueAvg, 64)
			xlsx.SetCellValue("Sheet1", "B"+strconv.Itoa(kk+9+(lea*k)), Round((floatValueAvg/vfun1/vfun2), 3))
			floatValueMax, _ := strconv.ParseFloat(vv.ValueMax, 64)
			xlsx.SetCellValue("Sheet1", "C"+strconv.Itoa(kk+9+(lea*k)), Round((floatValueMax/vfun1/vfun2), 3))
			floatValueMin, _ := strconv.ParseFloat(vv.ValueAvg, 64)
			xlsx.SetCellValue("Sheet1", "D"+strconv.Itoa(kk+9+(lea*k)), Round((floatValueMin/vfun1/vfun2), 3))
		}
	}

	xlsx.SetActiveSheet(index)
	var b bytes.Buffer
	err = xlsx.Write(&b)
	if err != nil {
		return []byte{}, nil
	}
	return b.Bytes(), nil

}

//CreateTrenXlsx excel table
func CreateTrenXlsx(Filedata []Trend, v ListQueryAll, start, end int64) ([]byte, error) {

	StartUnix := time.Unix(start, 0)
	StrStart := StartUnix.Format("2006-01-02 15:04:05")

	EndUnix := time.Unix(end, 0)
	StrEnd := EndUnix.Format("2006-01-02 15:04:05")

	xlsx := excelize.NewFile()
	// 创建一个工作表
	index := xlsx.NewSheet("Sheet1")
	//设置列宽
	xlsx.SetColWidth("Sheet1", "A", "A", 20)
	xlsx.SetColWidth("Sheet1", "B", "D", 15)

	//表头设计
	//主机名
	xlsx.SetCellValue("Sheet1", "A1", "主机名称")
	xlsx.SetCellValue("Sheet1", "B1", v.Host.Name)
	//指标类型
	xlsx.SetCellValue("Sheet1", "A2", "指标名称")
	xlsx.SetCellValue("Sheet1", "B2", v.Item.Name)
	//指标key
	xlsx.SetCellValue("Sheet1", "A3", "指标Key")
	xlsx.SetCellValue("Sheet1", "B3", v.Item.Key)

	//指标类型
	var ValueTypeStr string
	switch v.Item.ValueType {
	case "0":
		ValueTypeStr = "浮点型"
	case "1":
		ValueTypeStr = "字符型"
	case "2":
		ValueTypeStr = "日志型"
	case "3":
		ValueTypeStr = "整型"
	case "4":
		ValueTypeStr = "文本型"
	default:
		ValueTypeStr = "整型"
	}
	//指标ID
	xlsx.SetCellValue("Sheet1", "A4", "指标ID")
	xlsx.SetCellValue("Sheet1", "B4", v.Item.Itemid)
	//指标类型
	xlsx.SetCellValue("Sheet1", "A5", "数据类型")
	xlsx.SetCellValue("Sheet1", "B5", ValueTypeStr)
	//开始时间
	xlsx.SetCellValue("Sheet1", "A6", "开始时间")
	xlsx.SetCellValue("Sheet1", "B6", StrStart)
	//结束时间
	xlsx.SetCellValue("Sheet1", "A7", "结束时间")
	xlsx.SetCellValue("Sheet1", "B7", StrEnd)

	//数据样式设置
	stylecenter, err := xlsx.NewStyle(`{"alignment":{"horizontal":"center"}}`)
	if err != nil {
		logs.Error(err)
	}
	lea := len(Filedata)
	//设置单元格对其方式
	xlsx.SetCellStyle("Sheet1", "A8", "A"+strconv.Itoa(lea+9), stylecenter)
	xlsx.SetCellStyle("Sheet1", "B8", "B"+strconv.Itoa(lea+9), stylecenter)
	xlsx.SetCellStyle("Sheet1", "C8", "C"+strconv.Itoa(lea+9), stylecenter)
	xlsx.SetCellStyle("Sheet1", "D8", "D"+strconv.Itoa(lea+9), stylecenter)
	//遍历数据
	xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(8), "时间")
	xlsx.SetCellValue("Sheet1", "B"+strconv.Itoa(8), "平均"+"("+v.Item.Units+")")
	xlsx.SetCellValue("Sheet1", "C"+strconv.Itoa(8), "最大"+"("+v.Item.Units+")")
	xlsx.SetCellValue("Sheet1", "D"+strconv.Itoa(8), "最小"+"("+v.Item.Units+")")
	for k, v := range Filedata {
		//数据分类遍历
		//数据具体数据遍历
		loc, _ := time.LoadLocation("Asia/Shanghai")
		timeint64, _ := strconv.ParseInt(v.Clock, 10, 64)
		TimeUnix := time.Unix(timeint64, 0).In(loc)
		StrTime := TimeUnix.Format("2006-01-02 15:04:05")
		xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(k+9), StrTime)
		xlsx.SetCellValue("Sheet1", "B"+strconv.Itoa(k+9), v.ValueAvg)
		xlsx.SetCellValue("Sheet1", "C"+strconv.Itoa(k+9), v.ValueMax)
		xlsx.SetCellValue("Sheet1", "D"+strconv.Itoa(k+9), v.ValueMin)
	}

	xlsx.SetActiveSheet(index)
	var b bytes.Buffer
	err = xlsx.Write(&b)
	if err != nil {
		return []byte{}, nil
	}
	return b.Bytes(), nil
}

//CreateHistoryXlsx excel table
func CreateHistoryXlsx(Filedata []History, v ListQueryAll, start, end int64) ([]byte, error) {
	StartUnix := time.Unix(start, 0)
	StrStart := StartUnix.Format("2006-01-02 15:04:05")
	EndUnix := time.Unix(end, 0)
	StrEnd := EndUnix.Format("2006-01-02 15:04:05")

	xlsx := excelize.NewFile()
	// 创建一个工作表
	index := xlsx.NewSheet("Sheet1")
	//设置列宽
	xlsx.SetColWidth("Sheet1", "A", "B", 20)
	xlsx.SetColWidth("Sheet1", "B", "B", 30)

	//表头设计
	//主机名
	xlsx.SetCellValue("Sheet1", "A1", "主机名称")
	xlsx.SetCellValue("Sheet1", "B1", v.Host.Name)
	//指标名
	xlsx.SetCellValue("Sheet1", "A2", "指标名称")
	xlsx.SetCellValue("Sheet1", "B2", v.Item.Name)
	//指标key
	xlsx.SetCellValue("Sheet1", "A3", "指标Key")
	xlsx.SetCellValue("Sheet1", "B3", v.Item.Key)
	//指标类型
	var ValueTypeStr string
	switch v.Item.ValueType {
	case "0":
		ValueTypeStr = "浮点型"
	case "1":
		ValueTypeStr = "字符型"
	case "2":
		ValueTypeStr = "日志型"
	case "3":
		ValueTypeStr = "整型"
	case "4":
		ValueTypeStr = "文本型"
	default:
		ValueTypeStr = "整型"
	}
	//指标ID
	xlsx.SetCellValue("Sheet1", "A4", "指标ID")
	xlsx.SetCellValue("Sheet1", "B4", v.Item.Itemid)
	//指标类型
	xlsx.SetCellValue("Sheet1", "A5", "数据类型")
	xlsx.SetCellValue("Sheet1", "B5", ValueTypeStr)
	//开始时间
	xlsx.SetCellValue("Sheet1", "A6", "开始时间")
	xlsx.SetCellValue("Sheet1", "B6", StrStart)
	//结束时间
	xlsx.SetCellValue("Sheet1", "A7", "结束时间")
	xlsx.SetCellValue("Sheet1", "B7", StrEnd)

	//数据样式设置
	stylecenter, err := xlsx.NewStyle(`{"alignment":{"horizontal":"center"}}`)
	if err != nil {
		logs.Error(err)
	}
	lea := len(Filedata)
	//设置单元格对其方式
	for i := 0; i < 5; i++ {
		xlsx.SetCellStyle("Sheet1", "A8", "A"+strconv.Itoa(lea+9), stylecenter)
		xlsx.SetCellStyle("Sheet1", "B8", "B"+strconv.Itoa(lea+9), stylecenter)
	}
	//遍历数据
	xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(8), "时间")
	xlsx.SetCellValue("Sheet1", "B"+strconv.Itoa(8), "数值"+"("+v.Item.Units+")")
	for k, v := range Filedata {
		loc, _ := time.LoadLocation("Asia/Shanghai")
		timeint64, _ := strconv.ParseInt(v.Clock, 10, 64)
		TimeUnix := time.Unix(timeint64, 0).In(loc)
		StrTime := TimeUnix.Format("2006-01-02 15:04:05")
		xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(k+9), StrTime)
		xlsx.SetCellValue("Sheet1", "B"+strconv.Itoa(k+9), v.Value)
	}

	xlsx.SetActiveSheet(index)
	var b bytes.Buffer
	err = xlsx.Write(&b)
	if err != nil {
		return []byte{}, nil
	}
	return b.Bytes(), nil
}

//CreateHistoryXlsx excel table

//CreateAlarmXlsx excel table
func CreateAlarmXlsx(Filedata []Alarm, cnt, start, end int64) ([]byte, error) {
	StartUnix := time.Unix(start, 0)
	StrStart := StartUnix.Format("2006-01-02 15:04:05")
	EndUnix := time.Unix(end, 0)
	StrEnd := EndUnix.Format("2006-01-02 15:04:05")

	xlsx := excelize.NewFile()
	// 创建一个工作表
	index := xlsx.NewSheet("Sheet1")
	//设置列宽
	xlsx.SetColWidth("Sheet1", "A", "B", 30)
	xlsx.SetColWidth("Sheet1", "B", "B", 20)
	xlsx.SetColWidth("Sheet1", "C", "C", 20)
	xlsx.SetColWidth("Sheet1", "E", "E", 20)
	xlsx.SetColWidth("Sheet1", "F", "F", 40)
	xlsx.SetColWidth("Sheet1", "G", "G", 40)

	//表头设计
	//主机名
	xlsx.SetCellValue("Sheet1", "A1", "开始时间")
	xlsx.SetCellValue("Sheet1", "B1", StrStart)
	//指标名
	xlsx.SetCellValue("Sheet1", "A2", "结束时间")
	xlsx.SetCellValue("Sheet1", "B2", StrEnd)
	xlsx.SetCellValue("Sheet1", "A3", "告警共计")
	xlsx.SetCellValue("Sheet1", "B3", cnt)
	//指标key
	xlsx.SetCellValue("Sheet1", "A5", "主机名")
	xlsx.SetCellValue("Sheet1", "B5", "主机组")
	xlsx.SetCellValue("Sheet1", "C5", "告警时间")
	xlsx.SetCellValue("Sheet1", "D5", "告警等级")
	xlsx.SetCellValue("Sheet1", "E5", "告警Key")
	xlsx.SetCellValue("Sheet1", "F5", "告警摘要")
	xlsx.SetCellValue("Sheet1", "G5", "告警详情")
	xlsx.SetCellValue("Sheet1", "H5", "告警类型")
	xlsx.SetCellValue("Sheet1", "I5", "事件ID")
	for k, v := range Filedata {
		xlsx.SetCellValue("Sheet1", "A"+strconv.Itoa(k+6), v.Host)
		xlsx.SetCellValue("Sheet1", "B"+strconv.Itoa(k+6), v.Hgroup)
		xlsx.SetCellValue("Sheet1", "C"+strconv.Itoa(k+6), v.Occurtime.Format("2006-01-02 15:04:05"))
		xlsx.SetCellValue("Sheet1", "D"+strconv.Itoa(k+6), v.Level)
		xlsx.SetCellValue("Sheet1", "E"+strconv.Itoa(k+6), v.Hkey)
		xlsx.SetCellValue("Sheet1", "F"+strconv.Itoa(k+6), v.Message)
		xlsx.SetCellValue("Sheet1", "G"+strconv.Itoa(k+6), v.Detail)
		xlsx.SetCellValue("Sheet1", "H"+strconv.Itoa(k+6), v.Status)
		xlsx.SetCellValue("Sheet1", "I"+strconv.Itoa(k+6), v.EventID)
	}
	xlsx.SetActiveSheet(index)
	var b bytes.Buffer
	err := xlsx.Write(&b)
	if err != nil {
		return []byte{}, err
	}
	return b.Bytes(), nil
}
