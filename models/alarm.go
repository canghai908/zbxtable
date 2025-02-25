package models

import (
	"strconv"
	"time"

	"github.com/astaxie/beego/orm"
)

// TableName alarm
func (t *Alarm) TableName() string {
	return TableName("alarm")
}

// AddAlarm insert a new Alarm into database and returns
// last inserted Id on success.
func AddAlarm(m *Alarm) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// //update alarm notifystatus
func UpdateAlarmStatus(m *Alarm) (id int64, err error) {
	o := orm.NewOrm()
	//user
	v := Alarm{ID: m.ID}
	err = o.Read(&v)
	if err != nil {
		return 0, err
	}
	v.NotifyStatus = m.NotifyStatus
	id, err = o.Update(m, "NotifyStatus")
	if err != nil {
		return 0, err
	}
	return id, nil
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
func GetAllAlarm(begin, end time.Time, page, limit, hosts, ip, tenant_id, status, level string) (cnt int64, al []Alarm, err error) {
	o := orm.NewOrm()
	var alarms []Alarm
	var CountAlarms []Alarm
	qs := o.QueryTable(new(Alarm))
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
	if ip != "" {
		cond = cond.And("host_ip__icontains", ip)
	}

	qs = qs.Filter("occurtime__gte", begin).Filter("occurtime__lte", end).SetCond(cond)

	// 获取总数
	cnt, err = qs.All(&CountAlarms)
	if err != nil {
		return 0, []Alarm{}, err
	}

	// 获取分页数据
	_, err = qs.Limit(limits, (pages-1)*limits).OrderBy("-occurtime").All(&alarms)
	if err != nil {
		return 0, []Alarm{}, err
	}

	return cnt, alarms, nil
}

// get alarm tenant list
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
	if num > 0 {
		for i := 0; i < len(maps); i++ {
			p.ID = i
			p.TenantID = maps[i]["tenant_id"].(string)
			ss = append(ss, p)
		}
	}
	return num, ss, nil
}

// ExportAlarm export
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

// AnalysisAlarm all alarm
func AnalysisAlarm(begin, end time.Time, tenant_id string) (arrytile []string, pie []Pie, na []string, va []int, err error) {
	o := orm.NewOrm()
	strbeing := begin.Format("2006-01-02 15:04:05")
	strend := end.Format("2006-01-02 15:04:05")
	var maps []orm.Params
	var ss []string
	dpie := []Pie{}
	//饼图数据
	var num int64

	baseQueryPie := "SELECT level, COUNT(DISTINCT id) AS level_count " +
		"FROM zbxtable_alarm WHERE occurtime >= ? AND occurtime <= ? " +
		"AND (STATUS='故障' or STATUS='1') "

	var paramsPie []interface{}
	paramsPie = append(paramsPie, strbeing, strend)

	if tenant_id != "" {
		baseQueryPie += "AND tenant_id = ? "
		paramsPie = append(paramsPie, tenant_id)
	}

	baseQueryPie += "GROUP BY level ORDER BY level_count DESC"

	num, err = o.Raw(baseQueryPie, paramsPie...).Values(&maps)

	if err == nil && num > 0 {
		for i := 0; i < len(maps); i++ {
			ss = append(ss, maps[i]["level"].(string))
			va, _ := strconv.Atoi(maps[i]["level_count"].(string))
			n := Pie{Value: va, Name: maps[i]["level"].(string)}
			dpie = append(dpie, n)
		}
	}

	//修改top10数据查询
	var map1s []orm.Params
	var name []string
	var values []int

	baseQuery := "SELECT host, hostname, COUNT(DISTINCT id) AS host_count " +
		"FROM zbxtable_alarm WHERE occurtime >= ? AND occurtime <= ? " +
		"AND (STATUS='故障' or STATUS='1') "

	var params []interface{}
	params = append(params, strbeing, strend)

	if tenant_id != "" {
		baseQuery += "AND tenant_id = ? "
		params = append(params, tenant_id)
	}

	// 修改排序，确保按照告警数量降序排列
	baseQuery += "GROUP BY host, hostname ORDER BY COUNT(DISTINCT id) DESC LIMIT 10"

	_, err = o.Raw(baseQuery, params...).Values(&map1s)

	if err == nil && len(map1s) > 0 {
		for i := 0; i < len(map1s); i++ {
			name = append(name, map1s[i]["hostname"].(string))
			va, _ := strconv.Atoi(map1s[i]["host_count"].(string))
			values = append(values, va)
		}
	}

	return ss, dpie, name, values, nil
}
