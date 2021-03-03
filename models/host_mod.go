package models

//HostList struct
type HostList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []Hosts `json:"items"`
		Total int64   `json:"total"`
	} `json:"data"`
}

//Params map
type Params map[string]interface{}

//HostGet struct
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

//const aliav
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

//HostGroup struct
type HostGroup struct {
	GroupID  string       `json:"groupid,omitempty"`
	Name     string       `json:"name"`
	Internal InternalType `json:"internal,omitempty"`
}

//HostGroupID struct
type HostGroupID struct {
	GroupID string `json:"groupid"`
}

//HostGroupIds type
type HostGroupIds []HostGroupID

//HostInterfaces type
type HostInterfaces []HostInterface

//HostInterface type
type HostInterface struct {
	DNS   string        `json:"dns"`
	IP    string        `json:"ip"`
	Main  int           `json:"main"`
	Port  string        `json:"port"`
	Type  InterfaceType `json:"type"`
	UseIP int           `json:"useip"`
}

// type Hosts struct {
// 	HostId     string         `json:"hostid,omitempty"`
// 	Host       string         `json:"host"`
// 	Available  AvailableType  `json:"available"`
// 	Error      string         `json:"error"`
// 	Name       string         `json:"name"`
// 	Status     StatusType     `json:"status"`
// 	GroupIds   HostGroupIds   `json:"groups,omitempty"`
// 	Interfaces HostInterfaces `json:"interfaces,omitempty"`
// }

//Hosts struct
type Hosts struct {
	HostID     string   `json:"hostid,omitempty"`
	Host       string   `json:"host",omitempty`
	Available  string   `json:"available,omitempty"`
	Error      string   `json:"error"`
	Name       string   `json:"name,omitempty"`
	Status     string   `json:"status,omitempty"`
	Groups     string   `json:"groups,omitempty"`
	Interfaces string   `json:"interfaces,omitempty"`
	Template   []string `json:"template,omitempty"`
}

//Host struct
type Host struct {
	HostID     string   `json:"hostid,omitempty"`
	Host       string   `json:"host,omitempty"`
	Available  string   `json:"available,omitempty"`
	Error      string   `json:"error"`
	Name       string   `json:"name,omitempty"`
	Status     string   `json:"status,omitempty"`
	Groups     string   `json:"groups,omitempty"`
	Interfaces string   `json:"interfaces,omitempty"`
	Template   []string `json:"template,omitempty"`
}

//ListHosts struct
type ListHosts []struct {
	AutoCompress string `json:"auto_compress"`
	Available    string `json:"available"`
	Description  string `json:"description"`
	DisableUntil string `json:"disable_until"`
	Error        string `json:"error"`
	ErrorsFrom   string `json:"errors_from"`
	Flags        string `json:"flags"`
	Groups       []struct {
		Flags    string `json:"flags"`
		Groupid  string `json:"groupid"`
		Internal string `json:"internal"`
		Name     string `json:"name"`
	} `json:"groups"`
	Host       string `json:"host"`
	Hostid     string `json:"hostid"`
	Interfaces []struct {
		Bulk        string `json:"bulk"`
		DNS         string `json:"dns"`
		Hostid      string `json:"hostid"`
		Interfaceid string `json:"interfaceid"`
		IP          string `json:"ip"`
		Main        string `json:"main"`
		Port        string `json:"port"`
		Type        string `json:"type"`
		Useip       string `json:"useip"`
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
	ParentTemplates   []struct {
		Name       string `json:"name"`
		Templateid string `json:"templateid"`
	} `json:"parentTemplates"`
	ProxyAddress     string `json:"proxy_address"`
	ProxyHostid      string `json:"proxy_hostid"`
	SnmpAvailable    string `json:"snmp_available"`
	SnmpDisableUntil string `json:"snmp_disable_until"`
	SnmpError        string `json:"snmp_error"`
	SnmpErrorsFrom   string `json:"snmp_errors_from"`
	Status           string `json:"status"`
	Templateid       string `json:"templateid"`
	TLSAccept        string `json:"tls_accept"`
	TLSConnect       string `json:"tls_connect"`
	TLSIssuer        string `json:"tls_issuer"`
	TLSPsk           string `json:"tls_psk"`
	TLSPskIdentity   string `json:"tls_psk_identity"`
	TLSSubject       string `json:"tls_subject"`
}
