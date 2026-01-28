package models

import (
	"strconv"
	"time"
	"zbxtable/utils"
)

type TaskLog struct {
	Id        int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ReportID  int       `gorm:"column:report_id;default:0" json:"report_id"`
	Name      string    `gorm:"column:name;size:200" json:"name"`
	Cycle     string    `gorm:"column:cycle;size:64" json:"cycle"`
	StartTime time.Time `gorm:"column:start_time;type:datetime" json:"start_time"`
	EndTime   time.Time `gorm:"column:end_time;type:datetime" json:"end_time"`
	Status    int       `gorm:"column:status;default:0" json:"status"`
	Result    string    `gorm:"column:result;size:200" json:"result"`
	Files     string    `gorm:"column:files;size:200" json:"files"`
	TotalTime int64     `gorm:"column:total_time;default:0" json:"total_time"`
}

// SystemList struct
type TaskRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}

// TableName alarm
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
	err = DB.Create(m).Error
	if err != nil {
		utils.Log.Debug(err)
		return 0, err
	}
	return int64(m.Id), err
}

func (taskLog *TaskLog) Update(m *Report) (int64, error) {
	// UpdateTopologyByID updates Alarm by Id and returns error if
	var v Report
	err := DB.Where("id = ?", m.ID).First(&v).Error
	if err == nil {
		result := DB.Model(&Report{}).Where("id = ?", m.ID).Updates(m)
		if result.Error == nil {
			utils.Log.Debug("Number of records updated in database:", result.RowsAffected)
		}
		return int64(m.ID), nil
	}
	return 0, err
}

func GetTaskLogList(page, limit, report_id string) (cnt int64, topo []TaskLog, err error) {
	var tasklog []TaskLog
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	if limits == 0 {
		limits = 10
	}
	if pages == 0 {
		pages = 1
	}

	query := DB.Model(&TaskLog{}).Where("report_id = ?", report_id)

	// 获取总数
	err = query.Count(&cnt).Error
	if err != nil {
		utils.Log.Debug(err)
		return 0, []TaskLog{}, err
	}

	// 获取分页数据
	offset := (pages - 1) * limits
	err = query.Order("start_time DESC").Limit(limits).Offset(offset).Find(&tasklog).Error
	if err != nil {
		utils.Log.Debug(err)
		return 0, []TaskLog{}, err
	}
	return cnt, tasklog, nil
}

func (taskLog *TaskLog) Clear() (int64, error) {
	result := DB.Delete(&TaskLog{})
	if result.Error != nil {
		utils.Log.Debug(result.Error)
		return 0, result.Error
	}
	return result.RowsAffected, nil
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
	err = DB.Delete(&TaskLog{}, id).Error
	if err != nil {
		return err
	}
	return nil
}
