package models

//TrendRes resp
//IndexInfo struct
type IndexInfo struct {
	Hosts    int64 `json:"hosts"`
	Items    int64 `json:"items"`
	Triggers int64 `json:"triggers"`
	Problems int64 `json:"problems"`
}

type InfoRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items IndexInfo `json:"items"`
	} `json:"data"`
}
