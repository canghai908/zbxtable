package models

import (
	"github.com/astaxie/beego/logs"
	"github.com/astaxie/beego/orm"
	"strconv"
	"time"
)

type TaskLog struct {
	Id        int       `orm:"column(id);auto" json:"id"`
	ReportID  int       `orm:"column(report_id);default(0)" json:"report_id"`            // 任务id
	Name      string    `orm:"column(name);varchar(200);null" json:"name"`               // 任务名称
	Cycle     string    `orm:"column(cycle);varchar(64);null" json:"cycle"`              // crontab
	StartTime time.Time `orm:"column(start_time);type(datetime);null" json:"start_time"` // 开始执行时间
	EndTime   time.Time `orm:"column(end_time);type(datetime);null" json:"end_time"`     // 执行完成（失败）时间
	Status    int       `orm:"column(status);default(0)" json:"status"`                  // 状态 0:执行失败 1:执行中  2:执行完毕 3:任务取消(上次任务未执行完成) 4:异步执行
	Result    string    `orm:"column(result);size(200);null" json:"result"`
	Files     string    `orm:"column(files);size(200);null" json:"files"`
	TotalTime int64     `orm:"column(total_time);default(0)"json:"total_time"` // 执行总时长=
}

//SystemList struct
type TaskRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}

//TableName alarm
func (t *TaskLog) TableName() string {
	return TableName("task_log")
}

const (
	NoBegin = iota
	Running
	Success
	Failed
)

func CreateTaskLog(taskModel Report, status int) (int64, error) {
	taskLogModel := new(TaskLog)
	taskLogModel.ReportID = taskModel.ID
	taskLogModel.Name = taskModel.Name
	taskLogModel.Cycle = taskModel.Cycle
	taskLogModel.StartTime = time.Now()
	taskLogModel.Status = status
	insertId, err := taskLogModel.Create()
	return insertId, err
}

// AddTopology insert a new ZmsTopology into database and returns
// last inserted Id on success.
func (m *TaskLog) Create() (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	if err != nil {
		logs.Debug(err)
		return 0, err
	}
	return id, err
}

func (taskLog *TaskLog) Update(m *Report) (int64, error) {
	// UpdateTopologyByID updates Alarm by Id and returns error if
	o := orm.NewOrm()
	v := Report{ID: m.ID}
	var err error
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Update(m); err == nil {
			logs.Debug("Number of records updated in database:", num)
		}
		return int64(m.ID), nil
	}
	return 0, err
}

func GetTaskLogList(page, limit, report_id string) (cnt int64, topo []TaskLog, err error) {
	o := orm.NewOrm()
	var tasklog []TaskLog
	var count []TaskLog
	al := new(TaskLog)
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	//count topology
	_, err = o.QueryTable(al).Filter("report_id", report_id).All(&count)
	_, err = o.QueryTable(al).Limit(limits, (pages-1)*limits).OrderBy("-start_time").Filter("report_id", report_id).All(&tasklog)
	if err != nil {
		logs.Debug(err)
		return 0, []TaskLog{}, err
	}
	cnt = int64(len(count))
	return cnt, tasklog, nil
}

func (taskLog *TaskLog) Clear() (int64, error) {
	o := orm.NewOrm()
	al := new(TaskLog)
	id, err := o.QueryTable(al).Delete()
	if err != nil {
		logs.Debug(err)
		return 0, err
	}
	return id, err
}

//// 删除N个月前的日志
//func (taskLog *TaskLog) Remove(id int) (int64, error) {
//	t := time.Now().AddDate(0, -id, 0)
//	return Db.Where("start_time <= ?", t.Format(DefaultTimeFormat)).Delete(taskLog)
//}
//
//func (taskLog *TaskLog) Total(params CommonMap) (int64, error) {
//	session := Db.NewSession()
//	defer session.Close()
//	taskLog.parseWhere(session, params)
//	return session.Count(taskLog)
//}

// DeleteAlarm deletes Alarm by Id and returns error if
// the record to be deleted doesn't exist
func DeleteTaskLog(id int) (err error) {
	o := orm.NewOrm()
	v := TaskLog{Id: id}
	// ascertain id exists in the database
	if err = o.Read(&v); err == nil {
		_, err = o.Delete(&TaskLog{Id: id})
		if err != nil {
			return err
		}
		return nil
	}
	return nil
}
