package model

import (
	"compress/gzip"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
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

// LoginZabbixWeb 登录 Zabbix Web（图表回退路径）。失败只记日志，绝不退出进程。
func LoginZabbixWeb(ZabbixWeb, ZabbixUser, ZabbixPass string) error {
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
		Timeout:   30 * time.Second,
	}
	request, err := http.NewRequest("POST", ZabbixWeb+"/index.php", strings.NewReader(v.Encode()))
	if err != nil {
		logger.Log.Error("Fatal error ", err.Error())
		return err
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
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		err := fmt.Errorf("login to zabbix web failed, status=%d", response.StatusCode)
		logger.Log.Error(err)
		return err
	}

	var reader io.Reader
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		reader, _ = gzip.NewReader(response.Body)
	default:
		reader = response.Body
	}
	data, err := io.ReadAll(reader)
	if err != nil {
		logger.Log.Errorf("failed to read response data: %+v", err)
		return err
	}
	if strings.Contains(string(data), "blocked") {
		err := fmt.Errorf("login to Zabbix failed: account blocked")
		logger.Log.Error(err)
		return err
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
		logger.Log.Errorf("failed to parse URL: %v", err)
		return err
	}
	//把cookies设置到缓存中
	cookies := JAR.Cookies(u) // 从 CookieJar 中获取指定 URL 的 Cookies
	for _, cookie := range cookies {
		value, err := url.QueryUnescape(cookie.Value)
		if err != nil {
			logger.Log.Error("解码失败:", err)
			return err
		}
		err = CacheSet("zbx_session", value, 0)
		if err != nil {
			logger.Log.Error(err)
			return err
		}
	}
	return nil
}
