package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"strconv"
	"time"
)

//TableName alarm
func (t *Report) TableName() string {
	return TableName("report")
}

//get id
func GetReportsByID(id int) (v *Report, err error) {
	o := orm.NewOrm()
	v = &Report{ID: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

//get all
func GetALlReport() (cnt int64, system []Report, err error) {
	o := orm.NewOrm()
	var sys []Report
	al := new(Report)
	_, err = o.QueryTable(al).All(&sys)
	if err != nil {
		return 0, []Report{}, err
	}
	cnt = int64(len(sys))
	return cnt, sys, nil
}

// GetAllTopology t
func GetAllReportsLimt(page, limit, name string) (cnt int64, topo []Report, err error) {
	o := orm.NewOrm()
	var topologys []Report
	var CountTopologys []Report
	al := new(Report)
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	//count topology
	_, err = o.QueryTable(al).Filter("name__contains", name).All(&CountTopologys)
	_, err = o.QueryTable(al).Limit(limits, (pages-1)*limits).OrderBy("created_at").Filter("name__contains", name).All(&topologys)
	if err != nil {
		logs.Debug(err)
		return 0, []Report{}, err
	}
	cnt = int64(len(CountTopologys))
	return cnt, topologys, nil
}

// AddTopology insert a new ZmsTopology into database and returns
// last inserted Id on success.
func AddReport(m *Report) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	if err != nil {
		logs.Debug(err)
		return 0, err
	}
	return id, err
}

// UpdateTopologyByID updates Alarm by Id and returns error if
func UpdateReportByID(m *Report) (err error) {
	o := orm.NewOrm()
	v := Report{ID: m.ID}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		v.Name = m.Name
		v.Emails = m.Emails
		v.Items = m.Items
		v.LinkBandWidth = m.LinkBandWidth
		v.Cycle = m.Cycle
		v.Status = m.Status
		v.Desc = m.Desc
		_, err = o.Update(m, "Name", "Emails", "Items",
			"LinkBandWidth", "Cycle", "Status", "Desc")
		if err != nil {
			return err
		}
	}
	return nil
}

// UpdateTopologyByID updates Alarm by Id and returns error if
func CheckNowByID(m *Report) (err error) {
	o := orm.NewOrm()
	v := Report{ID: m.ID}
	// ascertain id exists in the database
	err = o.Read(&v)
	if err != nil {
		return err
	}
	if len(v.Cycle) != 0 {
		var cycle []string
		if err := json.Unmarshal([]byte(v.Cycle), &cycle); err != nil {
			logs.Error(err)
			return err
		}
		for _, vv := range cycle {
			//week
			if vv == "week" {
				start := time.Now()
				err := TaskWeekReport(v)
				if err != nil {
					logs.Error(err)
				}
				//更新report状态
				v.ExecStatus = strconv.Itoa(Success)
				v.StartAt = start
				v.EndAt = time.Now()
				err = UpdateReportExecStatusByID(&v)
				if err != nil {
					logs.Error(err)
				}
			}
			//day
			if vv == "day" {
				start := time.Now()
				err := TaskDayReport(v)
				if err != nil {
					logs.Error(err)
				}
				//更新report状态
				v.ExecStatus = strconv.Itoa(Success)
				v.StartAt = start
				v.EndAt = time.Now()
				err = UpdateReportExecStatusByID(&v)
				if err != nil {
					logs.Error(err)
				}
			}
		}
	}
	return
}

// UpdateTopologyByID updates Alarm by Id and returns error if
func UpdateReportExecStatusByID(m *Report) (err error) {
	o := orm.NewOrm()
	v := Report{ID: m.ID}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		v.ExecStatus = m.ExecStatus
		v.StartAt = m.StartAt
		v.EndAt = m.EndAt
		_, err = o.Update(m, "ExecStatus", "StartAt", "EndAt")
		if err != nil {
			logs.Error(err)
		}
	}
	return
}

// UpdateTopologyEdgesByID updates Alarm by Id and returns error if
func UpdateReportsStatusByID(m *Report) (err error) {
	o := orm.NewOrm()
	v := Report{ID: m.ID}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		if v.Status == "0" {
			m.Status = "1"
		} else {
			m.Status = "0"
		}
		_, err = o.Update(m, "Status")
		if err != nil {
			logs.Error(err)
		}
	}
	return
}

// DeleteAlarm deletes Alarm by Id and returns error if
// the record to be deleted doesn't exist
func DeleteReport(id int) (err error) {
	o := orm.NewOrm()
	v := Report{ID: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&Report{ID: id}); err == nil {
			logs.Debug("Number of records deleted in database:", num)
		}
	}
	return
}
