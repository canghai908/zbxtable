package models

import (
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/astaxie/beego"
)

//Jar struct
type Jar struct {
	cookies []*http.Cookie
}

//SetCookies a
func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.cookies = cookies
}

//Cookies func
func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies
}

//JAR st
var JAR = new(Jar)

//Intt a
func Intt() {
	v := url.Values{}
	ZabbixServer := beego.AppConfig.String("zabbix_server")
	ZabbixUser := beego.AppConfig.String("zabbix_user")
	ZabbixPass := beego.AppConfig.String("zabbix_pass")
	v.Set("name", ZabbixUser)
	v.Add("password", ZabbixPass)
	v.Add("autologin", "1")
	v.Add("enter", "Sign in")

	client := &http.Client{
		nil, nil, JAR, 99999999999999}
	reqest, err := http.NewRequest("POST", ZabbixServer+"/index.php", strings.NewReader(v.Encode()))
	if err != nil {
		beego.Error("Fatal error ", err.Error())
	}
	reqest.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	reqest.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	reqest.Header.Add("Accept-Encoding", "gzip, deflate")
	reqest.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
	reqest.Header.Add("Connection", "keep-alive")
	reqest.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")
	response, err := client.Do(reqest)
	if err != nil {
		beego.Error("Fatal error ", err.Error())
	}
	defer response.Body.Close()
	if beego.BConfig.RunMode == "dev" {
		beego.Info("Login to zabbix response.StatusCode is ", response.StatusCode)
	}
	if response.StatusCode == 200 {
		var reader io.Reader
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			reader, _ = gzip.NewReader(response.Body)
		default:
			reader = response.Body
		}
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			beego.Error("Failed to read response data: %+v", err)
		}
		if beego.BConfig.RunMode == "dev" {
			beego.Info("Login to zabbix response body is:", string(data))
		}
		if !strings.Contains(string(data), "Dashboard") {
			beego.Error("Login to Zabbix failed!")
			os.Exit(1)
		}
	}
}
