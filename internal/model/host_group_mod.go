package model

// HostGroups struct
type HostGroups struct {
	GroupID  string `json:"groupid"`
	Name     string `json:"name"`
	Internal string `json:"internal"`
	Flags    string `json:"flags"`
	Hosts    string `json:"hosts"`
}

// HostGroups struct
type HostGroupList struct {
	GroupID string `json:"groupid"`
	Name    string `json:"name"`
}

// HostGroupsPlist list
type HostGroupsPlist struct {
	GroupID  string `json:"groupid,omitempty"`
	Name     string `json:"name,omitempty"`
	Internal string `json:"internal,omitempty"`
	Flags    string `json:"flags,omitempty"`
	Hosts    []Host `json:"hosts,omitempty"`
}

// HostTree struct
type HostTree struct {
	GroupID   string `json:"groupid"`
	Name      string `json:"name"`
	Chrildren []struct {
		HostID string `json:"hostid"`
		Name   string `json:"name"`
	} `json:"hosts"`
}
type GroupHosts struct {
	HostID string `json:"hostid,omitempty"`
	Name   string `json:"name,omitempty"`
	Status string `json:"status"`
}

// HostGroupBYGroupID struct
type HostGroupBYGroupID struct {
	GroupID string       `json:"groupid,omitempty"`
	Name    string       `json:"name"`
	Hosts   []GroupHosts `json:"hosts"`
}
