package models

//TemplateList struct
type TemplateList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []Template `json:"items"`
		Total int64      `json:"total"`
	} `json:"data"`
}

//Template struct a
type Template struct {
	Host       string `json:"host"`
	Templateid string `json:"templateid"`
	Name       string `json:"name"`
	Hosts      []struct {
		Host   string `json:"host"`
		Name   string `json:"name"`
		HostID string `json:"hostid"`
	} `json:"hosts"`
	Applications string `json:"applications"`
	Triggers     string `json:"triggers"`
	Items        string `json:"items"`
	Graphs       string `json:"graphs"`
	Screens      string `json:"screens"`
	Discoveries  string `json:"discoveries"`
}
