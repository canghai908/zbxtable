package models

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/jordan-wright/email"
	"html/template"
	"net/smtp"
	"os"
	"zbxtable/utils"
)

func Sendmail(To []string, Subject, attach string, temp []byte) error {
	from := beego.AppConfig.String("email_from")
	nickname := beego.AppConfig.String("email_nickname")
	secret := beego.AppConfig.String("email_secret")
	host := beego.AppConfig.String("email_host")
	port, _ := beego.AppConfig.Int("email_port")
	isSSL, _ := beego.AppConfig.Bool("email_isSSl")
	auth := smtp.PlainAuth("", from, secret, host)
	e := email.NewEmail()
	if nickname != "" {
		e.From = fmt.Sprintf("%s <%s>", nickname, from)
	} else {
		e.From = from
	}
	var err error
	file, err := os.Open(DowloadPath + attach)
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
	var data ItemsHtml
	data.Title = m.Name
	data.LinkName = m.Name
	data.Start = chartdata[0].Start
	data.End = chartdata[0].End
	data.EndLine = "ZMS运维平台"
	var plist []TableDataList
	var one TableDataList
	for _, v := range chartdata {
		one.ItemName = v.Name
		one.IP = v.IP
		one.Host = v.Host
		one.LinkBinWith = utils.FormatTrafficFloat64(v.LinkBandWidth[0].Value.(float64))
		ttt := Avg(v.Data)
		one.Avg = utils.FormatTrafficFloat64(Avg(v.Data))
		one.AVgPre = AvgPer(ttt, v.LinkBandWidth[0].Value.(float64)) + "%"
		plist = append(plist, one)
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
