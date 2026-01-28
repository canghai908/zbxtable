package models

type GraphList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []GraphInfo `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}

//GraphListQuery struct
type GraphListQuery struct {
	Hostid string   `json:"hostid"`
	Period []string `json:"period"`
}

//GraphExpQuery struc
type GraphExpQuery struct {
	Hostids []string `json:"hostids"`
	Period  []string `json:"period"`
}

//GraphInfo struct
type GraphInfo struct {
	GraphID string `json:"graphid"`
	Name    string `json:"name"`
}

//GraphIDList as
type GraphIDList struct {
	Hosts     string      `json:"hosts"`
	GraphList []GraphInfo `json:"graphid"`
}

//GIDList struct
type GIDList struct {
	GIDList []GraphIDList `json:"gidlist"`
}

//GraphByteInfo struct
type GraphByteInfo struct {
	GraphID   string `json:"graphid"`
	GraphByte []byte `json:"graphbyte"`
	Name      string `json:"name"`
}
