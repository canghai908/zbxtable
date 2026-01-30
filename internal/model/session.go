package model

import (
	"compress/gzip"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"zbxtable/pkg/logger"
)

// Jar struct
type Jar struct {
	cookies []*http.Cookie
}

// SetCookies a
func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	jar.cookies = cookies
}

// Cookies func
func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
	return jar.cookies
}

// JAR st
var JAR = new(Jar)

// LoginZabbixWeb a
func LoginZabbixWeb(ZabbixWeb, ZabbixUser, ZabbixPass string) {
	v := url.Values{}
	v.Set("name", ZabbixUser)
	v.Add("password", ZabbixPass)
	v.Add("autologin", "1")
	v.Add("enter", "Sign in")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Jar:       JAR,
		Timeout:   99999999999999,
	}
	request, err := http.NewRequest("POST", ZabbixWeb+"/index.php", strings.NewReader(v.Encode()))
	if err != nil {
		logger.Log.Error("Fatal error ", err.Error())
		return
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded; param=value")
	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	request.Header.Add("Accept-Encoding", "gzip, deflate")
	request.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
	request.Header.Add("Connection", "keep-alive")
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")
	response, err := client.Do(request)
	if err != nil {
		logger.Log.Error("Fatal error ", err.Error())
		return
	}
	defer response.Body.Close()
	if response.StatusCode == 200 {
		var reader io.Reader
		switch response.Header.Get("Content-Encoding") {
		case "gzip":
			reader, _ = gzip.NewReader(response.Body)
		default:
			reader = response.Body
		}
		data, err := io.ReadAll(reader)
		if err != nil {
			logger.Log.Error("Failed to read response data: %+v", err)
		}
		// 开发模式下可以打印响应体用于调试
		// runmode := GetConfKey("runmode")
		// if runmode == "dev" {
		//	logger.Log.Info("Login to zabbix response body is:", string(data))
		// }
		if strings.Contains(string(data), "blocked") {
			logger.Log.Error("Login to Zabbix failed!")
			os.Exit(1)
		}
	} else {
		os.Exit(1)
	}

	// 开发模式下打印详细信息
	runmode := GetConfKey("runmode")
	if runmode == "dev" {
		logger.Log.Info("Login to zabbix  successfully！http status code:", response.StatusCode)
	} else {
		logger.Log.Info("Login to zabbix  successfully!")
	}
	//解析并把cookies存如redis
	u, err := url.Parse(ZabbixWeb)
	if err != nil {
		logger.Log.Error("Failed to parse URL: %v", err)
		return
	}
	//把cookies设置到缓存中
	cookies := JAR.Cookies(u) // 从 CookieJar 中获取指定 URL 的 Cookies
	for _, cookie := range cookies {
		value, err := url.QueryUnescape(cookie.Value)
		if err != nil {
			logger.Log.Error("解码失败:", err)
			return
		}
		err = CacheSet("zbx_session", value, 0)
		if err != nil {
			logger.Log.Error(err)
			return
		}
	}
}
