package models

//TrendList struct
type TrendList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []Trend `json:"items"`
		Total int64   `json:"total"`
	} `json:"data"`
}

//Trend struct
type Trend struct {
	Itemid   string `json:"itemid"`
	Clock    string `json:"clock"`
	Num      string `json:"num"`
	ValueMin string `json:"value_min"`
	ValueAvg string `json:"value_avg"`
	ValueMax string `json:"value_max"`
}
