package models

import (
	"archive/zip"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/astaxie/beego"
	"github.com/tealeg/xlsx"
)

var file *xlsx.File
var sheet *xlsx.Sheet
var row, row1, row2 *xlsx.Row
var cell *xlsx.Cell
var err error

//CreateOneXls a
func CreateOneXls(RRD []FileSystemDataALL, host, itemtype string, start, end int64) (filename string, err error) {
	newtitle := strings.Replace(host, "/", "_", -1)
	file = xlsx.NewFile()
	sheet, _ = file.AddSheet("Sheet1")

	row = sheet.AddRow()
	row.SetHeightCM(1)
	cell = row.AddCell()
	cell.Value = "主机名称"
	cell = row.AddCell()
	cell.Value = newtitle

	//dataname string
	var dataname, sourcename, unit string
	switch itemtype {
	case "cpu":
		dataname = "cpu使用率"
		sourcename = "CPU"
		unit = "%"
	case "mem":
		dataname = "内存使用"
		sourcename = "CPU"
		unit = "MB"
	case "disk":
		dataname = "磁盘空间"
		sourcename = "挂载点"
		unit = "MB"
	case "net":
		dataname = "网卡流量"
		sourcename = "网卡"
		unit = "KB"
	}

	row2 = sheet.AddRow()
	row2.SetHeightCM(1)
	cell = row2.AddCell()
	cell.Value = "指标类型"
	cell = row2.AddCell()
	cell.Value = dataname

	loc, _ := time.LoadLocation("Asia/Shanghai")
	StartUnix := time.Unix(start, 0).In(loc)
	StrStart := StartUnix.Format("2006-01-02 15:04:05")
	FileStart := StartUnix.Format("2006-01-02_15-04-05")

	row2 = sheet.AddRow()
	row2.SetHeightCM(1)
	cell = row2.AddCell()
	cell.Value = "开始时间"
	cell = row2.AddCell()
	cell.Value = StrStart

	EndUnix := time.Unix(end, 0).In(loc)
	StrEnd := EndUnix.Format("2006-01-02 15:04:05")
	FileEnd := EndUnix.Format("2006-01-02_15-04-05")

	row2 = sheet.AddRow()
	row2.SetHeightCM(1)
	cell = row2.AddCell()
	cell.Value = "结束时间"
	cell = row2.AddCell()
	cell.Value = StrEnd

	row2 = sheet.AddRow()
	row2.SetHeightCM(1)
	cell = row2.AddCell()
	cell.Value = ""
	cell = row2.AddCell()
	cell.Value = ""

	for i := 0; i < len(RRD); i++ {
		row2 = sheet.AddRow()
		row2.SetHeightCM(1)
		cell = row2.AddCell()
		cell.Value = sourcename
		cell = row2.AddCell()
		cell.Value = RRD[i].MountPoint

		row2 = sheet.AddRow()
		row2.SetHeightCM(1)
		cell = row2.AddCell()
		cell.Value = "时间"
		cell = row2.AddCell()
		cell.Value = "平均(" + unit + ")"
		cell = row2.AddCell()
		cell.Value = "最大(" + unit + ")"
		cell = row2.AddCell()
		cell.Value = "最小(" + unit + ")"

		for _, v := range RRD[i].FileSystemDataADD {
			row2 = sheet.AddRow()
			row2.SetHeightCM(1)
			cell = row2.AddCell()

			int64Clock, _ := strconv.ParseInt(v.Clock, 10, 64)
			cell.Value = time.Unix(int64Clock, 0).In(loc).Format("2006-01-02 15:04:05")
			//avg
			floatValueAvg, _ := strconv.ParseFloat(v.ValueAvg, 64)
			cell = row2.AddCell()
			cell.SetFloat(floatValueAvg / 1024 / 1024)
			//max
			floatValueMax, _ := strconv.ParseFloat(v.ValueMax, 64)
			cell = row2.AddCell()
			cell.SetFloat(floatValueMax / 1024 / 1024)
			//min
			floatValueMin, _ := strconv.ParseFloat(v.ValueAvg, 64)
			cell = row2.AddCell()
			cell.SetFloat(floatValueMin / 1024 / 1024)
		}

	}

	//filename string
	var filesname string
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	//og.Println(dir)

	//path := "download"
	path := dir + "\\download"
	switch runtime.GOOS {
	case "windows":
		_, err = os.Stat(path)
		if err != nil {
			beego.Error(err)
			os.MkdirAll(path, 0777)
		}
		filesname = path + "\\" + "[" + newtitle + "]" + FileStart + "-" + FileEnd + ".xlsx"
	case "linux":
		_, err = os.Stat(path)
		if err != nil {
			beego.Error(err)
			os.MkdirAll(path, 0777)
		}
		filesname = path + "/" + "[" + newtitle + "]" + FileStart + "-" + FileEnd + ".xlsx"
	case "darwin":
		_, err = os.Stat(path)
		if err != nil {
			beego.Error(err)
			os.MkdirAll(path, 0777)
		}
		filesname = path + "/" + "[" + newtitle + "]" + FileStart + "-" + FileEnd + ".xlsx"
	}
	//log.Println(runtime.GOOS)

	err = file.Save(filesname)
	if err != nil {
		beego.Error(err)
		return "", err
	}
	return filesname, nil
}

//FileCommon aa
func FileCommon(filelist []string, name string) error {
	var fe []*os.File
	for _, v := range filelist {
		f1, err := os.Open(v)
		if err != nil {
			beego.Error(err)
			return err
		}
		defer f1.Close()
		fe = append(fe, f1)
	}
	err = Compress(fe, name)
	if err != nil {
		beego.Error(err)
		return err
	}
	return nil

}

//Compress a
func Compress(files []*os.File, dest string) error {
	d, _ := os.Create(dest)
	defer d.Close()
	w := zip.NewWriter(d)
	defer w.Close()
	for _, file := range files {
		err := compress(file, "", w)
		if err != nil {
			beego.Error(err)
			return err
		}
	}
	return nil
}

//compress a
func compress(file *os.File, prefix string, zw *zip.Writer) error {
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		prefix = prefix + "/" + info.Name()
		fileInfos, err := file.Readdir(-1)
		if err != nil {
			beego.Error(err)
			return err
		}
		for _, fi := range fileInfos {
			f, err := os.Open(file.Name() + "/" + fi.Name())
			if err != nil {
				beego.Error(err)
				return err
			}
			err = compress(f, prefix, zw)
			if err != nil {
				beego.Error(err)
				return err
			}
		}
	} else {
		header, err := zip.FileInfoHeader(info)
		header.Name = prefix + "/" + header.Name
		if err != nil {
			beego.Error(err)
			return err
		}
		writer, err := zw.CreateHeader(header)
		if err != nil {
			beego.Error(err)
			return err
		}
		_, err = io.Copy(writer, file)
		file.Close()
		if err != nil {
			beego.Error(err)
			return err
		}
	}
	return nil
}
