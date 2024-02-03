package models

// ApplicationList struct
type ApplicationList struct {
	Code    int    `json:"code"`
	Message string `json:"mess age"`
	Data    struct {
		Items []Application `json:"items"`
		Total int64         `json:"total"`
	} `json:"data"`
}

// Application struct
type Application struct {
	Applicationid string            `json:"applicationid"`
	Name          string            `json:"name"`
	Items         []ApplicationItem `json:"items"`
}

// ApplicationItem struct
type ApplicationItem struct {
	ItemID    string `json:"itemid"`
	Name      string `json:"name"`
	ValutType string `json:"value_type"`
	Units     string `json:"units"`
}

type MonIts struct {
	Itemid    string `json:"itemid"`
	Key       string `json:"key_"`
	Lastclock string `json:"lastclock"`
	Lastvalue string `json:"lastvalue"`
	Name      string `json:"name"`
	Tags      []struct {
		Tag   string `json:"tag"`
		Value string `json:"value"`
	} `json:"tags"`
	SNMPOid    string `json:"snmp_oid"`
	Trends     string `json:"trends"`
	Units      string `json:"units"`
	ValueType  string `json:"value_type"`
	Delay      string `json:"delay"`
	ValuemapID string `json:"valuemapid"`
}

// Application struct
type TagsItems struct {
	Name  string   `json:"name"`
	Items []MonIts `json:"items"`
}
