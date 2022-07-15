package models

//TrendRes resp
//IndexInfo struct
type IndexInfo struct {
	Hosts    int64 `json:"hosts"`
	Items    int64 `json:"items"`
	Triggers int64 `json:"triggers"`
	Problems int64 `json:"problems"`
	WinCount int64 `json:"win_count"` //Windows主机
	LinCount int64 `json:"lin_count"` //Linux主机
	NETCount int64 `json:"net_count"` //网络设备
	SRVCount int64 `json:"srv_count"` //硬件服务器
}

type InfoRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items IndexInfo `json:"items"`
	} `json:"data"`
}
type VerRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items bool `json:"items"`
	} `json:"data"`
}

type TopRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []TopList `json:"top_list"`
	} `json:"data"`
}
type TreeRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []Treeinventory `json:"items"`
	} `json:"data"`
}
type TopList struct {
	Hostname string  `json:"hostname"`
	Score    float64 `json:"score"`
}

type Treeinventory struct {
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	TwoChildren []TwoChildren `json:"children"`
}
type TwoChildren struct {
	ID           int64          `json:"id"`
	Name         string         `json:"name"`
	TreeChildren []TreeChildren `json:"children"`
}
type TreeChildren struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
type OverviewRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items OverviewList `json:"items"`
	} `json:"data"`
}
type OverviewList struct {
	Win []Hosts `json:"vm_win"`
	Lin []Hosts `json:"vm_lin"`
	NET []Hosts `json:"hw_net"`
	SRV []Hosts `json:"hw_srv"`
}
type EgressRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
	} `json:"data"`
}
type EgressList struct {
	NameOne string `json:"name_one"`
	InOne   string `json:"in_one"`
	OutOne  string `json:"out_one"`
	NameTwo string `json:"name_two"`
	InTwo   string `json:"in_two"`
	OutTwo  string `json:"out_two"`
	Date    string `json:"date"`
}
