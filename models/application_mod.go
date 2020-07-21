package models

//ApplicationList struct
type ApplicationList struct {
	Code    int    `json:"code"`
	Message string `json:"mess age"`
	Data    struct {
		Items []Application `json:"items"`
		Total int64         `json:"total"`
	} `json:"data"`
}

//Application struct
type Application struct {
	Applicationid string            `json:"applicationid"`
	Name          string            `json:"name"`
	Items         []ApplicationItem `json:"items"`
}

//ApplicationItem struct
type ApplicationItem struct {
	ItemID    string `json:"itemid"`
	Name      string `json:"name"`
	ValutType string `json:"value_type"`
	Units     string `json:"units"`
}
