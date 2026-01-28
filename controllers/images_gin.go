package controllers

import (
	"compress/gzip"
	"crypto/tls"
	"io"
	"net/http"
	"net/url"
	"strings"
	"zbxtable/models"
	"zbxtable/utils"

	"github.com/gin-gonic/gin"
)

// GetImage 获取图形图片
func GetImage(c *gin.Context) {
	idStr := c.Param("id")
	var StartTime, EndTime string
	all := c.Query("from")
	b := strings.Split(all, "?")
	if len(b) > 0 {
		StartTime = b[0]
	}
	if len(b) > 1 {
		parts := strings.Split(b[1], "=")
		if len(parts) > 1 {
			EndTime = parts[1]
		}
	}
	GraphID := idStr
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, post-check=0, pre-check=0")
	c.Header("Pragma", "no-cache, value")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client1 := &http.Client{
		Transport: tr,
		Jar:       models.JAR,
		Timeout:   99999999999999,
	}
	ZabbixWeb := models.GetConfKey("zabbix_web")
	imgurl := ZabbixWeb + "/chart2.php?"
	data := url.Values{}
	URL, err := url.Parse(imgurl)
	if err != nil {
		utils.Log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	data.Set("graphid", GraphID)
	data.Set("from", StartTime)
	data.Set("to", EndTime)
	data.Set("profileIdx", "web.graphs.filter")
	data.Set("height", "200")
	data.Set("width", "400")
	URL.RawQuery = data.Encode()
	urlPath := URL.String()
	reqest1, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		utils.Log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	reqest1.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	reqest1.Header.Add("Accept-Encoding", "gzip, deflate")
	reqest1.Header.Add("Accept-Language", "zh-cn,zh;q=0.8,en-us;q=0.5,en;q=0.3")
	reqest1.Header.Add("Connection", "keep-alive")
	reqest1.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 6.1; WOW64; rv:12.0) Gecko/20100101 Firefox/12.0")
	response1, err := client1.Do(reqest1)
	if err != nil {
		utils.Log.Error(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response1.Body.Close()
	if response1.StatusCode == 200 {
		var reader io.Reader
		switch response1.Header.Get("Content-Encoding") {
		case "gzip":
			reader, _ = gzip.NewReader(response1.Body)
		default:
			reader = response1.Body
		}
		data, err := io.ReadAll(reader)
		if err != nil {
			utils.Log.Error(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.Data(http.StatusOK, "image/png", data)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch image"})
	}
}
