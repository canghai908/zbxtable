package model

import (
	"compress/gzip"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"zbxtable/pkg/logger"
)

// GetPNGGraphFromInstance 从指定实例获取图形（使用该实例的 JAR）
func GetPNGGraphFromInstance(inst *APIInstance, GraphID, start, end string) (png string, err error) {
	// 获取实例配置
	tenant, err := GetZabbixInstanceByZID(inst.ZID)
	if err != nil {
		return "", err
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client1 := &http.Client{
		Transport: tr,
		Jar:       inst.JAR, // 使用该实例的 JAR
		Timeout:   DefaultZabbixWebTimeout,
	}

	imgurl := inst.WebURL + "/chart2.php?"
	data := url.Values{}
	URL, err := url.Parse(imgurl)
	if err != nil {
		logger.Log.Error(err)
		return "", err
	}
	data.Set("graphid", GraphID)
	data.Set("from", start)
	data.Set("to", end)
	data.Set("profileIdx", "web.graphs.filter")
	data.Set("high", "200")
	data.Set("width", "800")
	data.Set("widget_view", "1")
	data.Set("resolve_macros", "1")

	URL.RawQuery = data.Encode()
	urlPath := URL.String()
	request, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		logger.Log.Error(err)
		return "", err
	}
	request.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	request.Header.Add("Accept-Encoding", "gzip, deflate")
	request.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
	request.Header.Add("Connection", "keep-alive")
	request.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")

	// 如果 JAR 中没有 cookie，先登录到该实例自己的 jar，禁止覆盖为全局 JAR
	if len(inst.JAR.cookies) == 0 && tenant.User != "" && tenant.Pass != "" {
		if err := LoginZabbixWebToJar(inst.WebURL, tenant.User, tenant.Pass, inst.JAR); err != nil {
			logger.Log.Error("LoginZabbixWebToJar failed:", err)
			return "", err
		}
	}

	response1, err := client1.Do(request)
	if err != nil {
		logger.Log.Error(err)
		return "", err
	}
	defer response1.Body.Close()

	if response1.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Zabbix Web returned HTTP %d for graph %s", response1.StatusCode, GraphID)
	}

	var reader io.Reader = response1.Body
	if response1.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(response1.Body)
		if err != nil {
			return "", fmt.Errorf("open compressed graph response: %w", err)
		}
		defer gzipReader.Close()
		reader = gzipReader
	}
	body, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	if http.DetectContentType(body) != "image/png" {
		return "", fmt.Errorf("Zabbix Web returned a non-PNG response for graph %s", GraphID)
	}
	return base64.StdEncoding.EncodeToString(body), nil
}
