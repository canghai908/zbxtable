package models

import (
	"log"
	"strconv"
	"time"

	"github.com/astaxie/beego/orm"
)

//Alarm struct
type Alarm struct {
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
}

//ListQueryAlarm query
type ListQueryAlarm struct {
	Host   string   `json:"host"`
	Period []string `json:"period"`
}

//ListExportAlarm struct
type ListExportAlarm struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
	Hosts string `json:"hosts"`
}

//ListAnalysisAlarm qu
type ListAnalysisAlarm struct {
	Begin string `json:"begin"`
	End   string `json:"end"`
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
func GetAllAlarm(begin, end time.Time, page, limit, hosts string) (cnt int64, alarm []Alarm, err error) {
	o := orm.NewOrm()
	var alarms []Alarm
	var CountAlarms []Alarm
	al := new(Alarm)
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	//count alarms
	_, err = o.QueryTable(al).Filter("occurtime__gte",
		begin).Filter("occurtime__lte", end).Filter("host__contains", hosts).All(&CountAlarms)
	// offset int64, limit int64, begin, end time.Time) (ml []interface{}, cnt int, err error) {
	_, err = o.QueryTable(al).Filter("occurtime__gte",
		begin).Filter("occurtime__lte", end).Limit(limits, (pages-1)*limits).OrderBy("-occurtime").Filter("host__contains", hosts).All(&alarms)
	if err != nil {
		return 0, []Alarm{}, err
	}
	cnt = int64(len(CountAlarms))
	return cnt, alarms, nil

}

//ExportAlarm export
func ExportAlarm(begin, end time.Time, hosts string) ([]byte, error) {
	o := orm.NewOrm()
	var alarms []Alarm
	al := new(Alarm)
	intbegin := begin.Unix()
	intend := end.Unix()
	// offset int64, limit int64, begin, end time.Time) (ml []interface{}, cnt int, err error) {
	_, err = o.QueryTable(al).Filter("occurtime__gte", begin).Filter("occurtime__lte",
		end).Filter("host__contains", hosts).OrderBy("-occurtime").All(&alarms)
	cnt := int64(len(alarms))
	if err != nil {
		return []byte{}, err
	}
	pbye, err := CreateAlarmXlsx(alarms, cnt, intbegin, intend)
	if err != nil {
		return []byte{}, err
	}
	return pbye, nil
}

//AnalysisAlarm all alarm
func AnalysisAlarm(begin, end time.Time) (arrytile []string, pie []Pie, na []string, va []int, err error) {
	o := orm.NewOrm()
	strbeing := begin.Format("2006-01-02 15:04:05")
	strend := end.Format("2006-01-02 15:04:05")
	var maps []orm.Params
	var ss []string
	dpie := []Pie{}
	//饼图数据
	num, err := o.Raw("SELECT level, COUNT(DISTINCT id) AS level_count FROM zbxtable_alarm  WHERE occurtime >='" +
		strbeing + "' and occurtime <='" + strend + "' AND STATUS='故障' GROUP BY level;").Values(&maps)
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
	_, err = o.Raw("SELECT host, COUNT(DISTINCT id) AS host_count FROM zbxtable_alarm WHERE  occurtime >='" +
		strbeing + "' and occurtime <='" + strend + "' AND STATUS='故障' GROUP BY host order by host_count desc limit 10;").Values(&map1s)
	if err == nil && num > 0 {
		if len(map1s) <= 10 {
			for i := 0; i < len(map1s); i++ {
				name = append(name, map1s[i]["host"].(string))
				va, _ := strconv.Atoi(map1s[i]["host_count"].(string))
				values = append(values, va)
			}
		} else {
			for i := 0; i <= 10; i++ {
				name = append(name, map1s[i]["host"].(string))
				va, _ := strconv.Atoi(map1s[i]["host_count"].(string))
				values = append(values, va)
			}
		}
	}
	return ss, dpie, name, values, nil
}

// UpdateAlarmByID updates Alarm by Id and returns error if
func UpdateAlarmByID(m *Alarm) (err error) {
	o := orm.NewOrm()
	v := Alarm{ID: m.ID}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			log.Println("Number of records updated in database:", num)
		}
	}
	return
}

// DeleteAlarm deletes Alarm by Id and returns error if
// the record to be deleted doesn't exist
func DeleteAlarm(id int) (err error) {
	o := orm.NewOrm()
	v := Alarm{ID: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&Alarm{ID: id}); err == nil {
			log.Println("Number of records deleted in database:", num)
		}
	}
	return
}
