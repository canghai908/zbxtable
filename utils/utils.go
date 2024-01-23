package utils

import (
	"archive/zip"
	"crypto/md5"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

// TimeFormat a
const TimeFormat = "2006-01-02 15:04:05"

// Md5 string
func Md5(buf []byte) string {
	hash := md5.New()
	hash.Write(buf)
	return fmt.Sprintf("%x", hash.Sum(nil))
}
func VAarToStr(str string) string {
	new1 := strings.Replace(str, "[", "", -1)
	new2 := strings.Replace(new1, "]", "", -1)
	new3 := strings.Replace(new2, `\`, "", -1)
	return strings.TrimSuffix(strings.Replace(new3, `"`, ``, -1), `,`)
}

// Mkdir mkdir
func Mkdir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}

// PathExists 文件目录是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func ComparePass(encodePW, passwordOK string) error {
	err := bcrypt.CompareHashAndPassword([]byte(encodePW), []byte(passwordOK))
	if err != nil {
		return err
	}
	return nil
}

// PasswordHash
func PasswordHash(pwd string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(pwd), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), err
}

// ParseTime func
func ParseTime(strtime string) (end time.Time, err error) {
	timeLayout := "2006-01-02 15:04:05"
	loc, _ := time.LoadLocation("Asia/Shanghai")
	theTime, err := time.Parse(timeLayout, strtime)
	if err != nil {
		return time.Now(), err
	}
	etime := theTime.In(loc)
	return etime, nil
}

// TimeFormater a
func UnixTimeFormater(timer string) string {
	i, err := strconv.ParseInt(timer, 10, 64)
	if err != nil {
		return ""
	}
	tm := time.Unix(i, 0).Format(TimeFormat)
	return tm
}

// ParTime time
func ParTime(dataStr string) (timer time.Time, err error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	layout := "2006-01-02 15:04:05"
	if dataStr == "" {
		b := time.Now()
		return b, nil
	}
	btime := strings.TrimSpace(strings.Replace(dataStr, ".", "-", -1))
	ot, err := time.ParseInLocation(layout, btime, loc)
	if err != nil {
		b := time.Now()
		return b, nil
	}
	return ot, nil
}

// RemoveRepByLoop a
func RemoveRepByLoop(slc []string) []string {
	result := []string{} // 存放结果
	for i := range slc {
		flag := true
		for j := range result {
			if slc[i] == result[j] {
				flag = false // 存在重复元素，标识为false
				break
			}
		}
		if flag { // 标识为false，不添加进结果
			result = append(result, slc[i])
		}
	}
	return result
}

// RemoveRepByMap a
func RemoveRepByMap(slc []string) []string {
	result := []string{}
	tempMap := map[string]byte{} // 存放不重复主键
	for _, e := range slc {
		l := len(tempMap)
		tempMap[e] = 0
		if len(tempMap) != l { // 加入map后，map长度变化，则元素不重复
			result = append(result, e)
		}
	}
	return result
}

// 合并
func MergeArr(a, b []string) []string {
	var arr []string
	for _, i := range a {
		arr = append(arr, i)
	}
	for _, j := range b {
		arr = append(arr, j)
	}
	return arr
}

// 去重
func UniqueArr(m []string) []string {
	d := make([]string, 0)
	tempMap := make(map[string]bool, len(m))
	for _, v := range m { // 以值作为键名
		if tempMap[v] == false {
			tempMap[v] = true
			d = append(d, v)
		}
	}
	return d
}

func FormatTraffic(traf string) (size string) {
	traffic, err := strconv.ParseFloat(traf, 64)
	if err != nil {
		return "0KB"
	}
	switch {
	case traffic < 1000:
		return fmt.Sprintf("%.2fBps", float64(traffic)/float64(1))
	case traffic < (1000 * 1000):
		return fmt.Sprintf("%.2fKBps", float64(traffic)/float64(1000))
	case traffic < (1000 * 1000 * 1000):
		return fmt.Sprintf("%.2fMBps", float64(traffic)/float64(1000*1000))
	case traffic < (1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%.2fGBps", float64(traffic)/float64(1000*1000*1000))
	case traffic < (1000 * 1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%.2fTBps", float64(traffic)/float64(1000*1000*1000*1000))
	case traffic < (1000 * 1000 * 1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%.2fPBps", float64(traffic)/float64(1000*1000*1000*1000*1000))
	case traffic < (1000 * 1000 * 1000 * 1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%.2fEBps", float64(traffic)/float64(1000*1000*1000*1000*1000*1000))
	default:
		return fmt.Sprintf("%.2fBps", float64(traffic)/float64(1))
	}
}

func FormatTrafficXlsx(traf string) (size string) {
	traffic, err := strconv.ParseFloat(traf, 64)
	if err != nil {
		return "0KB"
	}
	switch {
	case traffic < 1000:
		return fmt.Sprintf("%.6fBps", float64(traffic)/float64(1))
	case traffic < (1000 * 1000):
		return fmt.Sprintf("%.6fKBps", float64(traffic)/float64(1000))
	case traffic < (1000 * 1000 * 1000):
		return fmt.Sprintf("%.6fMBps", float64(traffic)/float64(1000*1000))
	case traffic < (1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%.6fGBps", float64(traffic)/float64(1000*1000*1000))
	case traffic < (1000 * 1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%.6fTBps", float64(traffic)/float64(1000*1000*1000*1000))
	case traffic < (1000 * 1000 * 1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%.6fPBps", float64(traffic)/float64(1000*1000*1000*1000*1000))
	case traffic < (1000 * 1000 * 1000 * 1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%.6fEBps", float64(traffic)/float64(1000*1000*1000*1000*1000*1000))
	default:
		return fmt.Sprintf("%.2fBps", float64(traffic)/float64(1))
	}
}
func FormatTrafficFloat64(traffic float64) (size string) {
	switch {
	case traffic < 1000:
		return fmt.Sprintf("%.4fB", float64(traffic)/float64(1))
	case traffic < (1000 * 1000):
		return fmt.Sprintf("%.4fK", float64(traffic)/float64(1000))
	case traffic < (1000 * 1000 * 1000):
		return fmt.Sprintf("%.4fM", float64(traffic)/float64(1000*1000))
	case traffic < (1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%.4fG", float64(traffic)/float64(1000*1000*1000))
	case traffic < (1000 * 1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%.4fT", float64(traffic)/float64(1000*1000*1000*1000))
	case traffic < (1000 * 1000 * 1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%.4fP", float64(traffic)/float64(1000*1000*1000*1000*1000))
	case traffic < (1000 * 1000 * 1000 * 1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%.4fE", float64(traffic)/float64(1000*1000*1000*1000*1000*1000))
	default:
		return fmt.Sprintf("%.2fBps", float64(traffic)/float64(1))
	}
}

func FormatSpeed(traf string) (size string) {
	traffic, err := strconv.ParseInt(traf, 10, 64)
	if err != nil {
		return "0K"
	}
	switch {
	case traffic < 1000:
		return fmt.Sprintf("%d%s", traffic/int64(1), "B")
	case traffic < (1000 * 1000):
		return fmt.Sprintf("%d%s", traffic/int64(1000), "K")
	case traffic < (1000 * 1000 * 1000):
		return fmt.Sprintf("%d%s", traffic/int64(1000*1000), "M")
	case traffic < (1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%d%s", traffic/int64(1000*1000*1000), "G")
	case traffic < (1000 * 1000 * 1000 * 1000 * 1000):
		return fmt.Sprintf("%d%s", traffic/int64(1000*1000*1000*1000), "T")
	default:
		return fmt.Sprintf("%d%s", "1G")
	}
}

func InterfaceTrafficeStrTofloat64(val string) (value float64) {
	t, err := strconv.ParseFloat(val, 64)
	if err != nil {
		logs.Error(err)
		return 0
	}
	return t
}

func InterfaceStrToInt64(val string) (value int64) {
	t, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		logs.Error(err)
		return 0
	}
	return t
}

func Int64ToTime(t string) (val string) {
	Tint, err := strconv.ParseInt(t, 10, 64)
	if err != nil {
		return "-"
	}
	tm := time.Unix(Tint, 0).Format(TimeFormat)
	return tm
}
func DecFloat64Round2(val string) (vaule float64) {
	Tfloat64, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0.00
	}
	vaule, _ = decimal.NewFromFloat(Tfloat64).Round(2).Float64()
	return vaule
}
func Float64Round2(val float64) (vaule float64) {
	vaule, _ = decimal.NewFromFloat(val).Round(2).Float64()
	return vaule
}

func ZipFiles(filename string, files []string, oldform, newform string) error {
	newZipFile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer newZipFile.Close()

	zipWriter := zip.NewWriter(newZipFile)
	defer zipWriter.Close()
	// 把files添加到zip中
	for _, file := range files {
		zipfile, err := os.Open(file)
		if err != nil {
			return err
		}
		defer zipfile.Close()
		info, err := zipfile.Stat()
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = strings.Replace(file, oldform, newform, -1)

		header.Method = zip.Deflate
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}
		if _, err = io.Copy(writer, zipfile); err != nil {
			return err
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// AlertSeverity
func AlertSeverityTo(v string) string {
	switch v {
	case "0":
		return "Not classified"
	case "1":
		return "Information"
	case "2":
		return "Warning"
	case "3":
		return "Average"
	case "4":
		return "High"
	case "5":
		return "Disaster"
	}
	return "Not classified"
}

func AlertType(v string) string {
	switch v {
	case "0":
		return "恢复"
	case "1":
		return "故障"
	}
	return "恢复"
}
