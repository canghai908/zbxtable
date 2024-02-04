package models

// HostList struct
type HostList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []Hosts `json:"items"`
		Total int64   `json:"total"`
	} `json:"data"`
}

type HostInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items Hosts `json:"items"`
		Total int64 `json:"total"`
	} `json:"data"`
}

type HostInterfaceInfo struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}

// Params map
type Params map[string]interface{}

// HostGet struct
type HostGet struct {
	ID      int    `json:"id"`
	Jsonrpc string `json:"jsonrpc"`
	Result  []struct {
		Available        string        `json:"available,omitempty"`
		Description      string        `json:"description"`
		DisableUntil     string        `json:"disable_until"`
		Error            string        `json:"error"`
		ErrorsFrom       string        `json:"errors_from"`
		Host             string        `json:"host"`
		Hostid           string        `json:"hostid"`
		Maintenances     []interface{} `json:"maintenances"`
		Name             string        `json:"name"`
		ProxyHostid      string        `json:"proxy_hostid"`
		SnmpAvailable    string        `json:"snmp_available"`
		SnmpDisableUntil string        `json:"snmp_disable_until"`
		SnmpError        string        `json:"snmp_error"`
		SnmpErrorsFrom   string        `json:"snmp_errors_from"`
		Status           string        `json:"status"`
	} `json:"result"`
}

type (
	// AvailableType sd
	AvailableType int
	// StatusType type int
	StatusType int
	// InterfaceType int
	InterfaceType int
	// InternalType int
	InternalType int
)

// const aliav
const (
	Available   AvailableType = 1
	Unavailable AvailableType = 2
	Monitored   StatusType    = 0
	Unmonitored StatusType    = 1
	Agent       InterfaceType = 1
	SNMP        InterfaceType = 2
	IPMI        InterfaceType = 3
	JMX         InterfaceType = 4
	NotInternal InternalType  = 0
	Internal    InternalType  = 1
)

// HostGroup struct
type HostGroup struct {
	GroupID  string       `json:"groupid,omitempty"`
	Name     string       `json:"name"`
	Internal InternalType `json:"internal,omitempty"`
}

// HostGroupID struct
type HostGroupID struct {
	GroupID string `json:"groupid"`
}

// HostGroupIds type
type HostGroupIds []HostGroupID

// HostInterfaces type
type HostInterfaces []HostInterface

// HostInterface type
type HostInterface struct {
	DNS   string        `json:"dns"`
	IP    string        `json:"ip"`
	Main  int           `json:"main"`
	Port  string        `json:"port"`
	Type  InterfaceType `json:"type"`
	UseIP int           `json:"useip"`
}

// Hosts struct
type Hosts struct {
	HostID            string `json:"hostid"` //主机
	Host              string `json:"host"`
	Available         string `json:"available"` //设备状态
	Error             string `json:"error"`
	Name              string `json:"name"` //设备名称
	Status            string `json:"status"`
	Interfaces        string `json:"interfaces"`         //IP地址
	NumberOfCores     string `json:"number_of_cores"`    //核心数
	CPUUtilization    string `json:"cpu_utilization"`    //cpu使用率
	MemoryUtilization string `json:"memory_utilization"` //内存使用率
	MemoryUsed        string `json:"memory_used"`        //内存使用大小
	MemoryTotal       string `json:"memory_total"`       //内存总大小
	Uptime            string `json:"uptime"`             //运行时长
	OS                string `json:"os"`                 //操作系统版本
	SystemName        string `json:"system_name"`        //主机名
	SerialNo          string `json:"serial_no"`          //序列号
	Model             string `json:"model"`              //设备类型
	Location          string `json:"location"`           //位置
	DateHwExpiry      string `json:"date_hw_expiry"`     //维保到期时间
	DateHwInstall     string `json:"date_hw_install"`    //设备安装时间
	Vendor            string `json:"vendor"`             //备注
	ResourceID        string `json:"resource_id"`        //资产编号
	MAC               string `json:"mac"`                //mac地址
	Department        string `json:"department"`         //部门
	Ping              string `json:"ping"`               //ping
	PingLoss          string `json:"ping_loss"`          //ping丢包率
	PingSec           string `json:"ping_sec"`           //ping时延
	Alarm             string `json:"alarm"`              //告警总数
}

// Host struct
type Host struct {
	HostID     string `json:"hostid,omitempty"`
	Host       string `json:"host,omitempty"`
	Available  string `json:"available,omitempty"`
	Error      string `json:"error"`
	Name       string `json:"name,omitempty"`
	Status     string `json:"status,omitempty"`
	Interfaces string `json:"interfaces,omitempty"`
}
type Inventory struct {
	Type             string `json:"type"`
	TypeFull         string `json:"type_full"`
	Name             string `json:"name"`
	Alias            string `json:"alias"`
	Os               string `json:"os"`
	OsFull           string `json:"os_full"`
	OsShort          string `json:"os_short"`
	SerialnoA        string `json:"serialno_a"`
	SerialnoB        string `json:"serialno_b"`
	Tag              string `json:"tag"`
	AssetTag         string `json:"asset_tag"`
	MacaddressA      string `json:"macaddress_a"`
	MacaddressB      string `json:"macaddress_b"`
	Hardware         string `json:"hardware"`
	HardwareFull     string `json:"hardware_full"`
	Software         string `json:"software"`
	SoftwareFull     string `json:"software_full"`
	SoftwareAppA     string `json:"software_app_a"`
	SoftwareAppB     string `json:"software_app_b"`
	SoftwareAppC     string `json:"software_app_c"`
	SoftwareAppD     string `json:"software_app_d"`
	SoftwareAppE     string `json:"software_app_e"`
	Contact          string `json:"contact"`
	Location         string `json:"location"`
	LocationLat      string `json:"location_lat"`
	LocationLon      string `json:"location_lon"`
	Notes            string `json:"notes"`
	Chassis          string `json:"chassis"`
	Model            string `json:"model"`
	HwArch           string `json:"hw_arch"`
	Vendor           string `json:"vendor"`
	ContractNumber   string `json:"contract_number"`
	InstallerName    string `json:"installer_name"`
	DeploymentStatus string `json:"deployment_status"`
	URLA             string `json:"url_a"`
	URLB             string `json:"url_b"`
	URLC             string `json:"url_c"`
	HostNetworks     string `json:"host_networks"`
	HostNetmask      string `json:"host_netmask"`
	HostRouter       string `json:"host_router"`
	OobIP            string `json:"oob_ip"`
	OobNetmask       string `json:"oob_netmask"`
	OobRouter        string `json:"oob_router"`
	DateHwPurchase   string `json:"date_hw_purchase"`
	DateHwInstall    string `json:"date_hw_install"`
	DateHwExpiry     string `json:"date_hw_expiry"`
	DateHwDecomm     string `json:"date_hw_decomm"`
	SiteAddressA     string `json:"site_address_a"`
	SiteAddressB     string `json:"site_address_b"`
	SiteAddressC     string `json:"site_address_c"`
	SiteCity         string `json:"site_city"`
	SiteState        string `json:"site_state"`
	SiteCountry      string `json:"site_country"`
	SiteZip          string `json:"site_zip"`
	SiteRack         string `json:"site_rack"`
	SiteNotes        string `json:"site_notes"`
	Poc1Name         string `json:"poc_1_name"`
	Poc1Email        string `json:"poc_1_email"`
	Poc1PhoneA       string `json:"poc_1_phone_a"`
	Poc1PhoneB       string `json:"poc_1_phone_b"`
	Poc1Cell         string `json:"poc_1_cell"`
	Poc1Screen       string `json:"poc_1_screen"`
	Poc1Notes        string `json:"poc_1_notes"`
	Poc2Name         string `json:"poc_2_name"`
	Poc2Email        string `json:"poc_2_email"`
	Poc2PhoneA       string `json:"poc_2_phone_a"`
	Poc2PhoneB       string `json:"poc_2_phone_b"`
	Poc2Cell         string `json:"poc_2_cell"`
	Poc2Screen       string `json:"poc_2_screen"`
	Poc2Notes        string `json:"poc_2_notes"`
}

// ListHosts struct
type ListHosts []struct {
	AutoCompress string    `json:"auto_compress"`
	Available    string    `json:"available"`
	Description  string    `json:"description"`
	DisableUntil string    `json:"disable_until"`
	Error        string    `json:"error"`
	Host         string    `json:"host"`
	Hostid       string    `json:"hostid"`
	Inventory    Inventory `json:"inventory"`
	Interfaces   []struct {
		Bulk        string `json:"bulk"`
		DNS         string `json:"dns"`
		Hostid      string `json:"hostid"`
		Interfaceid string `json:"interfaceid"`
		IP          string `json:"ip"`
		Main        string `json:"main"`
		Port        string `json:"port"`
		Type        string `json:"type"`
		Useip       string `json:"useip"`
		Error       string `json:"error"`
		Available   string `json:"available"`
	} `json:"interfaces"`
	IpmiAuthtype      string `json:"ipmi_authtype"`
	IpmiAvailable     string `json:"ipmi_available"`
	IpmiDisableUntil  string `json:"ipmi_disable_until"`
	IpmiError         string `json:"ipmi_error"`
	IpmiErrorsFrom    string `json:"ipmi_errors_from"`
	IpmiPassword      string `json:"ipmi_password"`
	IpmiPrivilege     string `json:"ipmi_privilege"`
	IpmiUsername      string `json:"ipmi_username"`
	JmxAvailable      string `json:"jmx_available"`
	JmxDisableUntil   string `json:"jmx_disable_until"`
	JmxError          string `json:"jmx_error"`
	JmxErrorsFrom     string `json:"jmx_errors_from"`
	Lastaccess        string `json:"lastaccess"`
	MaintenanceFrom   string `json:"maintenance_from"`
	MaintenanceStatus string `json:"maintenance_status"`
	MaintenanceType   string `json:"maintenance_type"`
	Maintenanceid     string `json:"maintenanceid"`
	Name              string `json:"name"`
	ProxyAddress      string `json:"proxy_address"`
	ProxyHostid       string `json:"proxy_hostid"`
	SnmpAvailable     string `json:"snmp_available"`
	SnmpDisableUntil  string `json:"snmp_disable_until"`
	SnmpError         string `json:"snmp_error"`
	SnmpErrorsFrom    string `json:"snmp_errors_from"`
	Status            string `json:"status"`
	Templateid        string `json:"templateid"`
	TLSAccept         string `json:"tls_accept"`
	TLSConnect        string `json:"tls_connect"`
	TLSIssuer         string `json:"tls_issuer"`
	TLSPsk            string `json:"tls_psk"`
	TLSPskIdentity    string `json:"tls_psk_identity"`
	TLSSubject        string `json:"tls_subject"`
}
type MonItemList []MonItem
type MonItem struct {
	Applicationid string `json:"applicationid"`
	Name          string `json:"name"`
	Items         []struct {
		Itemid     string `json:"itemid"`
		Name       string `json:"name"`
		Key        string `json:"key_"`
		ValueType  string `json:"value_type"`
		Delay      string `json:"delay"`
		Units      string `json:"units"`
		Lastvalue  string `json:"lastvalue"`
		Lastclock  string `json:"lastclock"`
		ValuemapID string `json:"valuemapid"`
	} `json:"items"`
}

type InterfaceDataList []InterfaceData
type InterfaceData struct {
	Index                      string  `json:"index"`
	Name                       string  `json:"name"`
	InDiscarded                int64   `json:"in_discarded"`
	InDiscardedItemId          string  `json:"in_discarded_itemid"`
	InDiscardedValueType       string  `json:"in_discarded_value_type"`
	InErrors                   int64   `json:"in_errors"`
	InErrorsItemId             string  `json:"in_errors_itemid"`
	InErrorsValueType          string  `json:"in_errors_value_type"`
	BitsReceived               float64 `json:"bits_received"`
	BitsReceivedItemId         string  `json:"bits_received_itemid"`
	BitsReceivedValueType      string  `json:"bits_received_value_type"`
	BitsSent                   float64 `json:"bits_sent"`
	BitsSentItemId             string  `json:"bits_sent_itemid"`
	BitsSentValueType          string  `json:"bits_sent_value_type"`
	OutDiscarded               int64   `json:"out_discarded"`
	OutDiscardedItemId         string  `json:"out_discarded_itemid"`
	OutDiscardedValueType      string  `json:"out_discarded_value_type"`
	OutErrors                  int64   `json:"out_errors"`
	OutErrorsItemId            string  `json:"out_errors_itemid"`
	OutErrorsValueType         string  `json:"out_errors_value_type"`
	Speed                      string  `json:"speed"`
	OperationalStatus          string  `json:"operational_status"`
	OperationalStatusItemId    string  `json:"operational_status_itemid"`
	OperationalStatusValueType string  `json:"operational_status_value_type"`
	Lastclock                  string  `json:"lastclock"`
	Begin                      string  `json:"begin"`
	End                        string  `json:"end"`
}
type WinFilesSystemData struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	SpaceUtilization float64 `json:"space_utilization"`
	TotalSpace       int64   `json:"total_space"`
	UsedSpace        int64   `json:"used_space"`
	Lastclock        string  `json:"lastclock"`
}

type LinFilesSystemData struct {
	ID               int64   `json:"id"`
	Name             string  `json:"name"`
	InodesPUsed      float64 `json:"inodes_pused"`
	SpaceUtilization float64 `json:"space_utilization"`
	TotalSpace       int64   `json:"total_space"`
	UsedSpace        int64   `json:"used_space"`
	Lastclock        string  `json:"lastclock"`
}

type MonWinData struct {
	FileSystem      []WinFilesSystemData `json:"filesystem"`
	FileSystemTotal int64                `json:"filesystem_total"`
	Interfaces      []InterfaceData      `json:"interfaces"`
	InterfacesTotal int64                `json:"interfaces_total"`
}
type MonLinData struct {
	FileSystem      []LinFilesSystemData `json:"filesystem"`
	FileSystemTotal int64                `json:"filesystem_total"`
	Interfaces      []InterfaceData      `json:"interfaces"`
	InterfacesTotal int64                `json:"interfaces_total"`
}
