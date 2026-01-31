package model

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"zbxtable/pkg/utils"

	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/jordan-wright/email"
)

func Sendmail(To []string, Subject, attach string, temp []byte) error {
	// 邮件配置优先从系统配置表读取，其次回退到 app.conf
	from := GetConfigValueByKey("email_from", GetConfKey("email_from"))
	nickname := GetConfigValueByKey("email_nickname", GetConfKey("email_nickname"))
	secret := GetConfigValueByKey("email_secret", GetConfKey("email_secret"))
	host := GetConfigValueByKey("email_host", GetConfKey("email_host"))
	portStr := GetConfigValueByKey("email_port", GetConfKey("email_port"))
	if portStr == "" {
		portStr = "465"
	}
	port, _ := strconv.Atoi(portStr)
	isSSlStr := GetConfigValueByKey("email_isSSl", GetConfKey("email_isSSl"))
	isSSL := strings.ToLower(isSSlStr) == "true"
	auth := smtp.PlainAuth("", from, secret, host)
	e := email.NewEmail()
	if nickname != "" {
		e.From = fmt.Sprintf("%s <%s>", nickname, from)
	} else {
		e.From = from
	}
	var err error
	file, err := os.Open(DownloadPath + attach)
	if err != nil {
		return err
	}
	attachment, err := e.Attach(bufio.NewReader(file), attach, "application/x-zip-compressed; charset=utf-8")
	if err != nil {
		return err
	}
	attachment.HTMLRelated = true
	e.Attachments[0] = attachment
	e.To = To
	e.Subject = Subject
	e.HTML = temp
	hostAddr := fmt.Sprintf("%s:%d", host, port)
	if isSSL {
		err = e.SendWithTLS(hostAddr, auth, &tls.Config{ServerName: host})
	} else {
		err = e.Send(hostAddr, auth)
	}
	if err != nil {
		return err
	}
	return nil
}

type ItemsHtml struct {
	Title     string          `json:"message"`
	LinkName  string          `json:"linkname"`
	Start     string          `json:"start"`
	End       string          `json:"end"`
	EndLine   string          `json:"endLine"`
	TableInfo []TableDataList `json:"device"`
}
type TableDataList struct {
	Host        string `json:"host"`
	IP          string `json:"ip"`
	ItemName    string `json:"itemname"`
	LinkBinWith string `json:"linkbinwith"`
	Avg         string `json:"avg"`
	AVgPre      string `json:"avgpre"`
}

func CreateMailTable(m Report, chartdata []ChartData) ([]byte, error) {
	// 没有任何图表数据时，直接返回错误，避免 panic
	if len(chartdata) == 0 {
		return nil, fmt.Errorf("CreateMailTable: no chart data for report %s", m.Name)
	}

	var data ItemsHtml
	data.Title = m.Name
	data.LinkName = m.Name
	data.Start = chartdata[0].Start
	data.End = chartdata[0].End
	data.EndLine = "ZMS运维平台"
	var plist []TableDataList
	var one TableDataList
	for _, v := range chartdata {
		// 跳过没有带宽信息或没有数据点的记录，避免越界或除零
		if len(v.LinkBandWidth) == 0 || len(v.Data) == 0 {
			continue
		}

		one.ItemName = v.Name
		one.IP = v.IP
		one.Host = v.Host
		one.LinkBinWith = utils.FormatTrafficFloat64(v.LinkBandWidth[0].Value.(float64))
		ttt := Avg(v.Data)
		one.Avg = utils.FormatTrafficFloat64(Avg(v.Data))
		one.AVgPre = AvgPer(ttt, v.LinkBandWidth[0].Value.(float64)) + "%"
		plist = append(plist, one)
	}
	// 如果所有记录都被过滤掉，也返回一个明确错误，给任务日志记录
	if len(plist) == 0 {
		return nil, fmt.Errorf("CreateMailTable: chart data has no valid rows for report %s", m.Name)
	}

	data.TableInfo = plist
	t, err := template.New("webpage").Parse(htmlReport)
	if err != nil {
		return []byte{}, err
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, data)
	if err != nil {
		return []byte{}, err
	}
	return buf.Bytes(), nil
}
func Avg(list []opts.LineData) (t float64) {
	var sum float64
	sum = float64(0)
	for _, v := range list {
		sum = sum + v.Value.(float64)
	}
	avg := sum / float64(len(list))
	return avg
}
func AvgPer(dataavg, bind float64) string {
	//per := utils.Float64Round2(datasum / bind)
	per := fmt.Sprintf("%.2f", float64(dataavg)/float64(bind)*100)
	return per
}
