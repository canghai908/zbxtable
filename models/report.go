package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"zbxtable/utils"

	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
)

// TableName alarm
func (t *Report) TableName() string {
	return TableName("report")
}

// GetReportsByID f
func GetReportsByID(id int) (v *Report, err error) {
	o := orm.NewOrm()
	v = &Report{ID: id}
	if err = o.Read(v); err == nil {
		return v, nil
	}
	return nil, err
}

// get all
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
func GetAllReportsLimt(page, limit, name, reportType string) (cnt int64, topo []Report, err error) {
	o := orm.NewOrm()
	var topologys []Report
	var CountTopologys []Report
	al := new(Report)
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	query := o.QueryTable(al)
	//count topology
	if name != "" {
		query = query.Filter("name__contains", name)
	}
	if reportType != "" {
		query = query.Filter("report_type", reportType)
	}
	_, err = query.All(&CountTopologys)
	if err != nil {
		return 0, []Report{}, err
	}
	query = o.QueryTable(al)
	if name != "" {
		query = query.Filter("name__contains", name)
	}
	if reportType != "" {
		query = query.Filter("report_type", reportType)
	}
	_, err = query.Limit(limits, (pages-1)*limits).OrderBy("created_at").All(&topologys)
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
	m.Items = utils.VAarToStr(m.Items)
	m.Cycle = utils.VAarToStr(m.Cycle)
	fmt.Println(m.Items, m.Cycle)
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
	err = o.Read(&v)
	if err != nil {
		return err
	}
	m.Items = utils.VAarToStr(m.Items)
	m.Cycle = utils.VAarToStr(m.Cycle)
	_, err = o.Update(m, "Name", "Emails", "Items",
		"LinkBandWidth", "HostIds", "ItemIds", "Cycle", "Status", "Desc", "Start", "End", "ReportMode")
	if err != nil {
		return err
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

	// 实时报表：直接生成，不需要周期
	if v.ReportMode == "realtime" && v.ReportType == "host" {
		start := time.Now()
		// 设置执行状态为处理中
		v.ExecStatus = strconv.Itoa(Running)
		v.StartAt = start
		if err := UpdateReportExecStatusByID(&v); err != nil {
			logs.Error(err)
			return err
		}

		// 实时报表使用配置的开始和结束时间，设置一个临时的Cycle用于文件命名
		vTemp := v
		if len(vTemp.Cycle) == 0 {
			vTemp.Cycle = "realtime" // 用于文件命名区分
		}
		taskErr := TaskHostReport(vTemp)
		if taskErr != nil {
			logs.Error(taskErr)
			v.ExecStatus = strconv.Itoa(Failed)
			v.EndAt = time.Now()
			UpdateReportExecStatusByID(&v)
			return taskErr
		}
		// 更新 report 状态
		v.ExecStatus = strconv.Itoa(Success)
		v.EndAt = time.Now()
		if err := UpdateReportExecStatusByID(&v); err != nil {
			logs.Error(err)
			return err
		}
		return nil
	}

	// 循环报表：需要配置周期
	if len(v.Cycle) == 0 {
		return nil
	}

	cycle := strings.Split(v.Cycle, ",")

	// 根据周期分别生成报表：
	// - 流量报表：day -> TaskDayReport，week -> TaskWeekReport
	// - 主机报表：day/week 都调用 TaskHostReport，但每次只传递当前周期，避免重复逻辑
	for _, vv := range cycle {
		// day
		if vv == "day" {
			start := time.Now()
			var taskErr error
			if v.ReportType == "host" {
				// 只针对当前周期生成主机报表
				vDay := v
				vDay.Cycle = "day"
				taskErr = TaskHostReport(vDay)
			} else {
				taskErr = TaskDayReport(v)
			}
			if taskErr != nil {
				logs.Error(taskErr)
				return taskErr
			}
			// 更新 report 状态
			v.ExecStatus = strconv.Itoa(Success)
			v.StartAt = start
			v.EndAt = time.Now()
			if err := UpdateReportExecStatusByID(&v); err != nil {
				logs.Error(err)
				return err
			}
		}

		// week
		if vv == "week" {
			start := time.Now()
			var taskErr error
			if v.ReportType == "host" {
				// 只针对当前周期生成主机报表
				vWeek := v
				vWeek.Cycle = "week"
				taskErr = TaskHostReport(vWeek)
			} else {
				taskErr = TaskWeekReport(v)
			}
			if taskErr != nil {
				logs.Error(taskErr)
				return taskErr
			}
			// 更新 report 状态
			v.ExecStatus = strconv.Itoa(Success)
			v.StartAt = start
			v.EndAt = time.Now()
			if err := UpdateReportExecStatusByID(&v); err != nil {
				logs.Error(err)
				return err
			}
		}
	}

	return nil
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
