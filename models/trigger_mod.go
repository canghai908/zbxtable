package models

//TriggersRes rest
type TriggersRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []EndTrigger `json:"items"`
		Total int64        `json:"total"`
	} `json:"data"`
}

//TriggersRes rest
type TriggersListRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []TriggerListStr `json:"items"`
		Total int64            `json:"total"`
	} `json:"data"`
}

//LastTriggers struct
type LastTriggers struct {
	Comments        string `json:"comments"`
	CorrelationMode string `json:"correlation_mode"`
	CorrelationTag  string `json:"correlation_tag"`
	Description     string `json:"description"`
	Details         string `json:"details"`
	Error           string `json:"error"`
	Expression      string `json:"expression"`
	Flags           string `json:"flags"`
	Hosts           []struct {
		Hostid string `json:"hostid"`
		Name   string `json:"name"`
	} `json:"hosts"`
	LastEvent struct {
		Acknowledged string `json:"acknowledged"`
		Clock        string `json:"clock"`
		Eventid      string `json:"eventid"`
		Name         string `json:"name"`
		Ns           string `json:"ns"`
		Object       string `json:"object"`
		Objectid     string `json:"objectid"`
		Severity     string `json:"severity"`
		Source       string `json:"source"`
		Value        string `json:"value"`
	} `json:"lastEvent"`
	Lastchange         string `json:"lastchange"`
	ManualClose        string `json:"manual_close"`
	Priority           string `json:"priority"`
	RecoveryExpression string `json:"recovery_expression"`
	RecoveryMode       string `json:"recovery_mode"`
	State              string `json:"state"`
	Status             string `json:"status"`
	Templateid         string `json:"templateid"`
	Triggerid          string `json:"triggerid"`
	Type               string `json:"type"`
	URL                string `json:"url"`
	Value              string `json:"value"`
}

//LastTriggers struct
type TriggerListStr struct {
	Comments           string `json:"comments"`
	CorrelationMode    string `json:"correlation_mode"`
	CorrelationTag     string `json:"correlation_tag"`
	Description        string `json:"description"`
	Details            string `json:"details"`
	Error              string `json:"error"`
	Expression         string `json:"expression"`
	Flags              string `json:"flags"`
	Lastchange         string `json:"lastchange"`
	ManualClose        string `json:"manual_close"`
	Priority           string `json:"priority"`
	RecoveryExpression string `json:"recovery_expression"`
	RecoveryMode       string `json:"recovery_mode"`
	State              string `json:"state"`
	Status             string `json:"status"`
	Templateid         string `json:"templateid"`
	Triggerid          string `json:"triggerid"`
	Type               string `json:"type"`
	URL                string `json:"url"`
	Value              string `json:"value"`
}

//EndTrigger struct
type EndTrigger struct {
	Acknowledged  string `json:"acknowledged"`
	Hostid        string `json:"hostid"`
	Name          string `json:"name"`
	Lastchange    string `json:"lastchange"`
	LastEventName string `json:"lasteventname"`
	Severity      string `json:"severity"`
	Eventid       string `json:"eventid"`
	Objectid      string `json:"objectid"`
}
