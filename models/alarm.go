package models

import (
	"strconv"
	"time"

	"github.com/astaxie/beego/orm"
)

//Alarm struct
type Alarm struct {
	ID int `orm:"column(id);auto" json:"id"`
	//v2 add begin
	TenantID  string `orm:"column(tenant_id);size(255)" json:"tenant_id"`
	HostID    string `orm:"column(host_id);size(255)" json:"host_id"`
	Hostname  string `orm:"column(hostname);size(200)" json:"hostname"`
	Host      string `orm:"column(host);size(200)" json:"host"`
	HostsIP   string `orm:"column(host_ip);size(200)" json:"host_ip"`
	TriggerID int64  `orm:"column(trigger_id);size(200)" json:"trigger_id"`
	ItemID    int64  `orm:"column(item_id);size(200)" json:"item_id"`
	ItemName  string `orm:"column(item_name);size(200)" json:"item_name"`
	ItemValue string `orm:"column(item_value);size(200)" json:"item_value"`
	//v2 add endss
	Hgroup    string    `orm:"column(hgroup);size(200)" json:"hgroup"`
	Occurtime time.Time `orm:"column(occurtime);type(datetime)" json:"occurtime"`
	Level     string    `orm:"column(level);size(200)" json:"level"`
	Message   string    `orm:"column(message);size(200)" json:"message"`
	Hkey      string    `orm:"column(hkey);size(200)" json:"hkey"`
	Detail    string    `orm:"column(detail);size(200)" json:"detail"`
	Status    string    `orm:"column(status);size(200)" json:"status"`
	EventID   int64     `orm:"column(event_id);size(200)" json:"eventid"`
}
type EventTpl struct {
	HostsID      string `json:"host_id"`
	HostHost     string `json:"host_host"`
	Hostname     string `json:"hostname"`
	HostsIP      string `json:"host_ip"`
	HostGroup    string `json:"host_group"`
	EventTime    string `json:"event_time"`
	Severity     string `json:"severity"`
	TriggerID    int64  `json:"trigger_id"`
	TriggerName  string `json:"trigger_name"`
	TriggerKey   string `json:"trigger_key"`
	TriggerValue string `json:"trigger_value"`
	ItemID       int64  `json:"item_id"`
	ItemName     string `json:"item_name"`
	ItemValue    string `json:"item_value"`
	EventID      int64  `json:"event_id"`
}

//ListQueryAlarm query
type ListQueryAlarm struct {
	Host   string   `json:"host"`
	Period []string `json:"period"`
}

//ListExportAlarm struct
type ListExportAlarm struct {
	Begin    string `json:"begin"`
	End      string `json:"end"`
	Hosts    string `json:"hosts"`
	TenantID string `json:"tenant_id"`
	Status   string `json:"status"`
	Level    string `json:"level"`
}

//ListAnalysisAlarm qu
type ListAnalysisAlarm struct {
	Begin    string `json:"begin"`
	End      string `json:"end"`
	TenantID string `json:"tenant_id"`
}

//SendALarm struct
type SendALarm struct {
	ID        int       `orm:"column(id);auto" json:"id"`
	Host      string    `orm:"column(host);size(255)" json:"host"`
	Hgroup    string    `orm:"column(hgroup);size(255)" json:"hgroup"`
	Occurtime time.Time `orm:"column(occurtime);type(datetime)" json:"occurtime"`
	Level     string    `orm:"column(level);size(255)" json:"level"`
	Message   string    `orm:"column(message);size(255)" json:"message"`
	Hkey      string    `orm:"column(hkey);size(255)" json:"hkey"`
	Detail    string    `orm:"column(detail);size(255)" json:"detail"`
	Status    string    `orm:"column(status);size(255)" json:"status"`
	EventID   string    `orm:"column(event_id);size(255)" json:"eventid"`
	Mail      []string  `orm:"column(event_id);size(255)" json:"mail"`
	Weixin    []string  `orm:"column(event_id);size(255)" json:"weixin"`
	Sms       []string  `orm:"column(event_id);size(255)" json:"sms"`
}

//WeixinMessage struct
type WeixinMessage struct {
	ID        int       `orm:"column(id);auto" json:"id"`
	Host      string    `orm:"column(host);size(255)" json:"host"`
	Hgroup    string    `orm:"column(hgroup);size(255)" json:"hgroup"`
	Occurtime time.Time `orm:"column(occurtime);type(datetime)" json:"occurtime"`
	Level     string    `orm:"column(level);size(255)" json:"level"`
	Message   string    `orm:"column(message);size(255)" json:"message"`
	Hkey      string    `orm:"column(hkey);size(255)" json:"hkey"`
	Detail    string    `orm:"column(detail);size(255)" json:"detail"`
	Status    string    `orm:"column(status);size(255)" json:"status"`
	EventID   string    `orm:"column(event_id);size(255)" json:"eventid"`
	Weixin    string    `orm:"column(event_id);size(255)" json:"weixin"`
}

//MailMessage struct
type MailMessage struct {
	ID        int       `orm:"column(id);auto" json:"id"`
	Host      string    `orm:"column(host);size(255)" json:"host"`
	Hgroup    string    `orm:"column(hgroup);size(255)" json:"hgroup"`
	Occurtime time.Time `orm:"column(occurtime);type(datetime)" json:"occurtime"`
	Level     string    `orm:"column(level);size(255)" json:"level"`
	Message   string    `orm:"column(message);size(255)" json:"message"`
	Hkey      string    `orm:"column(hkey);size(255)" json:"hkey"`
	Detail    string    `orm:"column(detail);size(255)" json:"detail"`
	Status    string    `orm:"column(status);size(255)" json:"status"`
	EventID   string    `orm:"column(event_id);size(255)" json:"eventid"`
	Mail      string    `orm:"column(event_id);size(255)" json:"mail"`
}

//SmsMessage struct
type SmsMessage struct {
	ID        int       `orm:"column(id);auto" json:"id"`
	Host      string    `orm:"column(host);size(255)" json:"host"`
	Hgroup    string    `orm:"column(hgroup);size(255)" json:"hgroup"`
	Occurtime time.Time `orm:"column(occurtime);type(datetime)" json:"occurtime"`
	Level     string    `orm:"column(level);size(255)" json:"level"`
	Message   string    `orm:"column(message);size(255)" json:"message"`
	Hkey      string    `orm:"column(hkey);size(255)" json:"hkey"`
	Detail    string    `orm:"column(detail);size(255)" json:"detail"`
	Status    string    `orm:"column(status);size(255)" json:"status"`
	EventID   string    `orm:"column(event_id);size(255)" json:"eventid"`
	Sms       string    `orm:"column(event_id);size(255)" json:"sms"`
}

//Pie struct
type Pie struct {
	Value int    `json:"value"`
	Name  string `json:"name"`
}

//AlarmList struct
type AlarmList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []Alarm `json:"items"`
		Total int64   `json:"total"`
	} `json:"data"`
}

//AlarmList struct
type AlarmTendantList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}

//AnalysisList struct
type AnalysisList struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Level      []string `json:"level"`
		LevelCount []Pie    `json:"level_count"`
		Host       []string `json:"host"`
		HostCount  []int    `json:"host_count"`
	} `json:"data"`
}

//TableName alarm
func (t *Alarm) TableName() string {
	return TableName("alarm")
}

// AddAlarm insert a new Alarm into database and returns
// last inserted Id on success.
func AddAlarm(m *Alarm) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

// GetAlarmByID retrieves Alarm by Id. Returns error if
// Id doesn't exist
func GetAlarmByID(id int) (v *Alarm, err error) {
	o := orm.NewOrm()
	v = &Alarm{ID: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// GetAllAlarm retrieves all Alarm matches certain condition. Returns empty list if
// no records exist
func GetAllAlarm(begin, end time.Time, page, limit,
	hosts, tenant_id, status, level string) (cnt int64, alarm []Alarm, err error) {
	o := orm.NewOrm()
	var alarms []Alarm
	var CountAlarms []Alarm
	al := new(Alarm)
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	//count alarms
	cond := orm.NewCondition()
	if hosts != "" {
		cond = cond.And("host__icontains", hosts)
	}
	if tenant_id != "" {
		cond = cond.And("tenant_id", tenant_id)
	}
	if status != "" {
		cond = cond.And("status", status)
	}
	if level != "" {
		cond = cond.And("level", level)
	}
	_, err = o.QueryTable(al).Filter("occurtime__gte", begin).Filter("occurtime__lte", end).
		SetCond(cond).
		All(&CountAlarms)
	_, err = o.QueryTable(al).Filter("occurtime__gte", begin).Filter("occurtime__lte", end).
		Limit(limits, (pages-1)*limits).OrderBy("-occurtime").SetCond(cond).
		All(&alarms)
	if err != nil {
		return 0, []Alarm{}, err
	}
	cnt = int64(len(CountAlarms))
	return cnt, alarms, nil
}

//get alarm tenant list
func GetAlarmTenant() (cnt int64, data interface{}, err error) {
	o := orm.NewOrm()
	var maps []orm.Params
	num, err := o.Raw("select distinct zbxtable_alarm.tenant_id from zbxtable_alarm;").Values(&maps)
	if err != nil {
		return 0, []Alarm{}, err
	}
	type list struct {
		ID       int    `json:"id"`
		TenantID string `json:"tenant_id"`
	}
	var ss []list
	var p list
	if err == nil && num > 0 {
		for i := 0; i < len(maps); i++ {
			p.ID = i
			p.TenantID = maps[i]["tenant_id"].(string)
			ss = append(ss, p)
		}
	}
	return num, ss, nil
}

//ExportAlarm export
func ExportAlarm(begin, end time.Time,
	hosts, tenant_id, status, level string) ([]byte, error) {
	o := orm.NewOrm()
	var alarms []Alarm
	al := new(Alarm)
	intbegin := begin.Unix()
	intend := end.Unix()
	//count alarms
	cond := orm.NewCondition()
	if hosts != "" {
		cond = cond.And("host__icontains", hosts)
	}
	if tenant_id != "" {
		cond = cond.And("tenant_id", tenant_id)
	}
	if status != "" {
		cond = cond.And("status", status)
	}
	if level != "" {
		cond = cond.And("level", level)
	}
	_, err := o.QueryTable(al).Filter("occurtime__gte", begin).Filter("occurtime__lte", end).
		SetCond(cond).
		OrderBy("-occurtime").All(&alarms)
	if err != nil {
		return []byte{}, err
	}
	cnt := int64(len(alarms))
	pbye, err := CreateAlarmXlsx(alarms, cnt, intbegin, intend)
	if err != nil {
		return []byte{}, err
	}
	return pbye, nil
}

//AnalysisAlarm all alarm
func AnalysisAlarm(begin, end time.Time, tenant_id string) (arrytile []string, pie []Pie, na []string, va []int, err error) {
	o := orm.NewOrm()
	strbeing := begin.Format("2006-01-02 15:04:05")
	strend := end.Format("2006-01-02 15:04:05")
	var maps []orm.Params
	var ss []string
	dpie := []Pie{}
	//饼图数据
	var num int64
	if tenant_id == "" {
		num, err = o.Raw("SELECT level, COUNT(DISTINCT id) AS level_count FROM zbxtable_alarm  WHERE occurtime >='" +
			strbeing + "' and occurtime <='" + strend + "' AND (STATUS='故障' or  STATUS='1') GROUP BY level;").Values(&maps)
	} else {
		num, err = o.Raw("SELECT level, COUNT(DISTINCT id) AS level_count FROM zbxtable_alarm  WHERE occurtime >='" +
			strbeing + "' and occurtime <='" + strend + "' AND (STATUS='故障' or  STATUS='1') AND tenant_id ='" +
			tenant_id + "' GROUP BY level;").Values(&maps)
	}
	if err == nil && num > 0 {
		for i := 0; i < len(maps); i++ {
			ss = append(ss, maps[i]["level"].(string))
			va, _ := strconv.Atoi(maps[i]["level_count"].(string))
			n := Pie{Value: va, Name: maps[i]["level"].(string)}
			dpie = append(dpie, n)
		}
	}
	//top10数据
	var map1s []orm.Params
	var name []string
	var values []int
	if tenant_id == "" {
		_, err = o.Raw("SELECT hostname, COUNT(DISTINCT id) AS host_count FROM zbxtable_alarm WHERE  occurtime >='" +
			strbeing + "' and occurtime <='" + strend + "' AND (STATUS='故障' or STATUS='1') GROUP BY host order by host_count desc limit 10;").Values(&map1s)
	} else {
		_, err = o.Raw("SELECT hostname, COUNT(DISTINCT id) AS host_count FROM zbxtable_alarm WHERE  occurtime >='" +
			strbeing + "' and occurtime <='" + strend + "' AND (STATUS='故障' or STATUS='1') AND  tenant_id ='" +
			tenant_id + "' GROUP BY host order by host_count desc limit 10;").Values(&map1s)
	}
	if err == nil && num > 0 {
		if len(map1s) <= 10 {
			for i := 0; i < len(map1s); i++ {
				name = append(name, map1s[i]["hostname"].(string))
				va, _ := strconv.Atoi(map1s[i]["host_count"].(string))
				values = append(values, va)
			}
		} else {
			for i := 0; i <= 10; i++ {
				name = append(name, map1s[i]["hostname"].(string))
				va, _ := strconv.Atoi(map1s[i]["host_count"].(string))
				values = append(values, va)
			}
		}
	}
	return ss, dpie, name, values, nil
}
