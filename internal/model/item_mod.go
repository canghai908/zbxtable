package models

//ItemList struct
type ItemList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []Item `json:"items"`
		Total int64  `json:"total"`
	} `json:"data"`
}

//ItemList struct
type ItemRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}

//Item struct
type Item struct {
	Itemid               string        `json:"itemid"`
	Type                 string        `json:"type,"`
	SnmpCommunity        string        `json:"snmp_community,omitempty"`
	SnmpOid              string        `json:"snmp_oid,omitempty"`
	Hostid               string        `json:"hostid,"`
	Name                 string        `json:"name,"`
	Key                  string        `json:"key_,"`
	Delay                string        `json:"delay,"`
	History              string        `json:"history"`
	Trends               string        `json:"trends"`
	Lastvalue            string        `json:"lastvalue"`
	Lastclock            string        `json:"lastclock"`
	Prevvalue            string        `json:"prevvalue,omitempty"`
	State                string        `json:"state,omitempty"`
	Status               string        `json:"status,omitempty"`
	ValueType            string        `json:"value_type,omitempty"`
	TrapperHosts         string        `json:"trapper_hosts,omitempty"`
	Units                string        `json:"units,omitempty"`
	Snmpv3Securityname   string        `json:"snmpv3_securityname,omitempty"`
	Snmpv3Securitylevel  string        `json:"snmpv3_securitylevel,omitempty"`
	Snmpv3Authpassphrase string        `json:"snmpv3_authpassphrase,omitempty"`
	Snmpv3Privpassphrase string        `json:"snmpv3_privpassphrase,omitempty"`
	Snmpv3Authprotocol   string        `json:"snmpv3_authprotocol,omitempty"`
	Snmpv3Privprotocol   string        `json:"snmpv3_privprotocol,omitempty"`
	Snmpv3Contextname    string        `json:"snmpv3_contextname,omitempty"`
	Error                string        `json:"error,omitempty"`
	Lastlogsize          string        `json:"lastlogsize,omitempty"`
	Logtimefmt           string        `json:"logtimefmt,omitempty"`
	Templateid           string        `json:"templateid,omitempty"`
	Valuemapid           string        `json:"valuemapid,omitempty"`
	Params               string        `json:"params,omitempty"`
	IpmiSensor           string        `json:"ipmi_sensor,omitempty"`
	Authtype             string        `json:"authtype,omitempty"`
	Username             string        `json:"username,omitempty"`
	Password             string        `json:"password,omitempty"`
	Publickey            string        `json:"publickey,omitempty"`
	Privatekey           string        `json:"privatekey,omitempty"`
	Mtime                string        `json:"mtime,omitempty"`
	Lastns               string        `json:"lastns,omitempty"`
	Flags                string        `json:"flags,omitempty"`
	Interfaceid          string        `json:"interfaceid,omitempty"`
	Port                 string        `json:"port,omitempty"`
	Description          string        `json:"description,omitempty"`
	InventoryLink        string        `json:"inventory_link,omitempty"`
	Lifetime             string        `json:"lifetime,omitempty"`
	Evaltype             string        `json:"evaltype,omitempty"`
	JmxEndpoint          string        `json:"jmx_endpoint,omitempty"`
	MasterItemid         string        `json:"master_itemid,omitempty"`
	Timeout              string        `json:"timeout,omitempty"`
	URL                  string        `json:"url,omitempty"`
	QueryFields          []interface{} `json:"query_fields,omitempty"`
	Posts                string        `json:"posts,omitempty"`
	StatusCodes          string        `json:"status_codes,omitempty"`
	FollowRedirects      string        `json:"follow_redirects,omitempty"`
	PostType             string        `json:"post_type,omitempty"`
	HTTPProxy            string        `json:"http_proxy,omitempty"`
	Headers              []interface{} `json:"headers,omitempty"`
	RetrieveMode         string        `json:"retrieve_mode,omitempty"`
	RequestMethod        string        `json:"request_method,omitempty"`
	OutputFormat         string        `json:"output_format,omitempty"`
	SslCertFile          string        `json:"ssl_cert_file,omitempty"`
	SslKeyFile           string        `json:"ssl_key_file,omitempty"`
	SslKeyPassword       string        `json:"ssl_key_password,omitempty"`
	VerifyPeer           string        `json:"verify_peer,omitempty"`
	VerifyHost           string        `json:"verify_host,omitempty"`
	AllowTraps           string        `json:"allow_traps,omitempty"`
}

// item get all
type Items struct {
	Itemid string `json:"itemid"`
	Name   string `json:"name"`
	Key    string `json:"key_"`
}

type ValueMap struct {
	Hostid   string `json:"hostid"`
	Mappings []struct {
		Newvalue string `json:"newvalue"`
		Type     string `json:"type"`
		Value    string `json:"value"`
	} `json:"mappings"`
	Name       string `json:"name"`
	UUID       string `json:"uuid"`
	Valuemapid string `json:"valuemapid"`
}
