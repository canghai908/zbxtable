package utils

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"
)

//TimeFormat a
const TimeFormat = "2006-01-02 15:04:05"

//Md5 string
func Md5(buf []byte) string {
	hash := md5.New()
	hash.Write(buf)
	return fmt.Sprintf("%x", hash.Sum(nil))
}

//ParseTime func
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

//TimeFormater a
func TimeFormater(timer string) string {
	t, _ := time.Parse(TimeFormat, timer)
	return t.String()
}

//ParTime time
func ParTime(dataStr string) (timer time.Time, err error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	layout := "2006-01-02 15:04:05"
	if dataStr == "" {
		b := time.Now()
		return b, nil
	}
	btime := strings.TrimSpace(strings.Replace(dataStr, ".", "-", -1))
	ot, err := time.ParseInLocation(layout, string(btime), loc)
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
