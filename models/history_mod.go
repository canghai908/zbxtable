package models

// HistoryList struct
type HistoryList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []History `json:"items"`
		Total int64     `json:"total"`
	} `json:"data"`
}

// History struct
type History struct {
	Itemid string `json:"itemid"`
	Clock  string `json:"clock"`
	Value  string `json:"value"`
	Ns     string `json:"ns"`
}

// ListQueryAll struct
type HistoryQuery struct {
	Itemids string   `json:"itemids"`
	History string   `json:"history"`
	Period  []string `json:"period"`
}

// XAxis x轴
type XAxis struct {
	Type string   `json:"type"`
	Data []string `json:"data"`
}

// YAxis y轴
type YAxis struct {
	Name      string    `json:"name"`
	Data      []float64 `json:"data"`
	Type      string    `json:"type"`
	ItemStyle ItemStyle `json:"itemStyle"`
}
type ItemStyle struct {
	Opacity float64 `json:"opacity"`
	Color   string  `json:"color"`
}

// Legend 图例
type Legend struct {
	Data []string `json:"data"`
}

// TrafficSeries 流量图数据
type TrafficSeries struct {
	XAxis        XAxis          `json:"xAxis"`
	YAxis        []YAxis        `json:"yAxis"`
	Legend       Legend         `json:"legend"`
	TrafficTable []TrafficTable `json:"table"`
}

// TrafficeTable 流量表数据
type TrafficTable struct {
	Name        string `json:"name"`
	Min         string `json:"min"`
	Max         string `json:"max"`
	Avg         string `json:"avg"`
	Th95PercAvg string `json:"th_perc_avg"`
	Th95PercVal string `json:"th_perc_val"`
}

// ErrorsSeries 错误包图数据
type ErrorsSeries struct {
	XAxis  XAxis   `json:"xAxis"`
	YAxis  []YAxis `json:"yAxis"`
	Legend Legend  `json:"legend"`
	Table  []Table `json:"table"`
}

// DiscardedSeries 丢弃包图数据
type DiscardedSeries struct {
	XAxis  XAxis   `json:"xAxis"`
	YAxis  []YAxis `json:"yAxis"`
	Legend Legend  `json:"legend"`
	Table  []Table `json:"table"`
}

// OperationalStatusSeries 运行状态图数据
type OperationalStatusSeries struct {
	XAxis  XAxis   `json:"xAxis"`
	YAxis  []YAxis `json:"yAxis"`
	Legend Legend  `json:"legend"`
	Table  []Table `json:"table"`
}
type Table struct {
	Name string `json:"name"`
	Min  string `json:"min"`
	Max  string `json:"max"`
	Avg  string `json:"avg"`
}

type TrafficData struct {
	TrafficSeries     TrafficSeries           `json:"traffic_series"`
	ErrorsSeries      ErrorsSeries            `json:"errors_series"`
	DiscardedSeries   DiscardedSeries         `json:"discarded_series"`
	OperationalSeries OperationalStatusSeries `json:"operational_status_series"`
}
