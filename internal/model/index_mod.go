package model

// TrendRes resp
// IndexInfo struct（动态主机数量通过 HostCounts map 返回）
type IndexInfo struct {
	Hosts           int64            `json:"hosts"`
	Items           int64            `json:"items"`
	Triggers        int64            `json:"triggers"`
	Problems        int64            `json:"problems"`
	LinCount        int64            `json:"lin_count"`
	WinCount        int64            `json:"win_count"`
	SrvCount        int64            `json:"srv_count"`
	NetCount        int64            `json:"net_count"`
	TotalCount      int64            `json:"total_count"`
	HostCounts      map[string]int64 `json:"host_counts"` // key=type_code, value=数量
	AssetTypeCounts []AssetTypeCount `json:"asset_type_counts"`
}

type AssetTypeCount struct {
	TypeCode string `json:"type_code"`
	Name     string `json:"name"`
	Icon     string `json:"icon"`
	Count    int64  `json:"count"`
}

type RouRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
	} `json:"data"`
}
type RouterRes struct {
	Router   string     `json:"router"`
	Children []MenuItem `json:"children"`
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
		Items version `json:"items"`
	} `json:"data"`
}
type version struct {
	ZabbixVersion string `json:"zabbixVersion"`
	Version       string `json:"version"`
	GitHash       string `json:"gitHash"`
	BuildTime     string `json:"buildTime"`
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
	Hostname     string  `json:"hostname"`
	Score        float64 `json:"score"`
	CPU          float64 `json:"cpu"`
	MEM          float64 `json:"mem"`
	InstanceName string  `json:"instance_name,omitempty"` // 实例名称
}

type Treeinventory struct {
	ID          int64         `json:"id"`
	Name        string        `json:"name"`
	TwoChildren []TwoChildren `json:"children"`
}
type TwoChildren struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	TypeCode string `json:"type_code"`
	Icon     string `json:"icon"`
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

// OverviewList 动态资产类型列表，key=type_code（如 VM_LIN），value=主机列表
type OverviewList map[string][]Hosts
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

type SessionRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items string `json:"items"`
	} `json:"data"`
}
