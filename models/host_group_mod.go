package models

//HostGroupsList struct
type HostGroupsList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []HostGroups `json:"items"`
		Total int64        `json:"total"`
	} `json:"data"`
}

//HostGroupBYGroupIDList struct
type HostGroupBYGroupIDList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []HostGroupBYGroupID `json:"items"`
		Total int64                `json:"total"`
	} `json:"data"`
}

//HostGroups struct
type HostGroups struct {
	GroupID  string `json:"groupid"`
	Name     string `json:"name"`
	Internal string `json:"internal"`
	Flags    string `json:"flags"`
	Hosts    string `json:"hosts"`
}

//HostGroupsPlist list
type HostGroupsPlist struct {
	GroupID  string `json:"groupid,omitempy"`
	Name     string `json:"name,omitempy"`
	Internal string `json:"internal,omitempy"`
	Flags    string `json:"flags,omitempy"`
	Hosts    []Host `json:"hosts,omitempy"`
}

//HostTreeList sst
type HostTreeList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []HostTree `json:"items"`
		Total int64      `json:"total"`
	} `json:"data"`
}

//HostTree struct
type HostTree struct {
	GroupID   string `json:"groupid"`
	Name      string `json:"name"`
	Chrildren []struct {
		HostID string `json:"hostid"`
		Name   string `json:"name"`
	} `json:"hosts"`
}
type GroupHosts struct {
	HostID string `json:"hostid,omitempy"`
	Name   string `json:"name,omitempy"`
	Status string `json:"status"`
}

//HostGroupBYGroupID struct
type HostGroupBYGroupID struct {
	GroupID string       `json:"groupid,omitempy"`
	Name    string       `json:"name"`
	Hosts   []GroupHosts `json:"hosts"`
}
