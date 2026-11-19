package model

import (
	"crypto/tls"
	"net/http"
	"time"

	zabbix "github.com/canghai908/zabbix-go"
)

// DefaultZabbixAPITimeout 防止 Zabbix JSON-RPC 挂起拖死请求与定时任务。
const DefaultZabbixAPITimeout = 30 * time.Second

// NewZabbixAPI 创建带超时与 TLS SkipVerify 的 Zabbix API 客户端。
func NewZabbixAPI(apiURL string) *zabbix.API {
	api := zabbix.NewAPI(apiURL)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	api.SetClient(&http.Client{
		Transport: transport,
		Timeout:   DefaultZabbixAPITimeout,
	})
	return api
}
