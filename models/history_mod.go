package models

//HistoryList struct
type HistoryList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []History `json:"items"`
		Total int64     `json:"total"`
	} `json:"data"`
}

//History struct
type History struct {
	Itemid string `json:"itemid"`
	Clock  string `json:"clock"`
	Value  string `json:"value"`
	Ns     string `json:"ns"`
}
