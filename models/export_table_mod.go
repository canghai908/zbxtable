package models

//ListQuery struct
type ListQuery struct {
	HostID   string   `json:"hostid"`
	ItemType string   `json:"itemtype"`
	Period   []string `json:"period"`
}

//ListQueryNew struct
type ListQueryNew struct {
	HostID string   `json:"hostid"`
	Item   Item     `json:"item"`
	Period []string `json:"period"`
}

//ListQueryAll struct
type ListQueryAll struct {
	Host   Host     `json:"host"`
	Item   Item     `json:"item"`
	Period []string `json:"period"`
}

//Itm struct
type Itm struct {
	Itemids  string `json:"itemids"`
	ItemName string `json:"itemname"`
	ItemKey  string `json:"itemkey"`
	Status   string `json:"status"`
	State    string `json:"state"`
}

//ExpList struct
type ExpList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items      []FileSystemDataALL `json:"items"`
		MountPoint []string            `json:"mountpoint"`
		FileName   string              `json:"filename"`
	} `json:"data"`
}

//FileSystemInfo data
type FileSystemInfo struct {
	MountPoint string `json:"mountpoint"`
	ItemID     string `json:"itemid"`
	ItemName   string `json:"itemname"`
	ItemKey    string `json:"itemkey"`
}

//FileSystemDataVue struct
type FileSystemDataVue struct {
	FileSystemDataADD []FileSystemData `json:"filesystemdata"`
}

//FileSystemDataALL struct
type FileSystemDataALL struct {
	MountPoint        string           `json:"mountpoint"`
	FileSystemDataADD []FileSystemData `json:"filesystemdata"`
}

//FileSystemData data
type FileSystemData struct {
	MountPoint string `json:"mountpoint"`
	ItemID     string `json:"itemid"`
	ItemName   string `json:"itemname"`
	ItemKey    string `json:"itemkey"`
	Clock      string `json:"clock"`
	Num        string `json:"num"`
	ValueMin   string `json:"value_min"`
	ValueAvg   string `json:"value_avg"`
	ValueMax   string `json:"value_max"`
}

//Insp a
type Insp struct {
	HostName string  `json:"hostname"`
	CPULoad  float64 `json:"cpuload"`
	MemPct   float64 `json:"mempct"`
}

//HostsData strunct
type HostsData struct {
	Hostid string `json:"hostid"`
	Host   string `json:"host"`
	Items  []struct {
		History string `json:"history"`
		Itemid  string `json:"itemid"`
		Key     string `json:"key_"`
		Name    string `json:"name"`
		State   string `json:"state"`
		Status  string `json:"status"`
		Trends  string `json:"trends"`
	} `json:"items"`
	ParentTemplates []struct {
		Name       string `json:"name"`
		Templateid string `json:"templateid"`
	} `json:"parentTemplates"`
}
