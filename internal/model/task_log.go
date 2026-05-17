package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/logger"
)

type TaskLog struct {
	Id        int       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ReportID  int       `gorm:"column:report_id;default:0" json:"report_id"`
	Name      string    `gorm:"column:name;size:200" json:"name"`
	Cycle     string    `gorm:"column:cycle;size:64" json:"cycle"`
	StartTime time.Time `gorm:"column:start_time;type:timestamp" json:"start_time"`
	EndTime   time.Time `gorm:"column:end_time;type:timestamp" json:"end_time"`
	Status    int       `gorm:"column:status;default:0" json:"status"`
	Result    string    `gorm:"column:result;size:200" json:"result"`
	Files     string    `gorm:"column:files;size:200" json:"files"`
	TotalTime int64     `gorm:"column:total_time;default:0" json:"total_time"`
	Progress  int       `gorm:"-" json:"progress"`
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

const taskLogProgressPrefix = "[progress:"

func formatTaskLogProgressResult(progress int, detail string) string {
	if progress < 0 {
		progress = 0
	}
	if progress > 100 {
		progress = 100
	}
	return fmt.Sprintf("%s%d]%s", taskLogProgressPrefix, progress, detail)
}

func parseTaskLogProgressResult(result string) (int, string) {
	if !strings.HasPrefix(result, taskLogProgressPrefix) {
		return 0, result
	}

	endIndex := strings.Index(result, "]")
	if endIndex == -1 {
		return 0, result
	}

	progressStr := strings.TrimPrefix(result[:endIndex], taskLogProgressPrefix)
	progress, err := strconv.Atoi(progressStr)
	if err != nil {
		return 0, result
	}

	return progress, strings.TrimSpace(result[endIndex+1:])
}

func buildTaskLogResult(status, progress int, detail string) string {
	if status == Running {
		return formatTaskLogProgressResult(progress, detail)
	}
	return detail
}

func CreateTaskLog(taskModel Report, status int) (int64, error) {
	taskLogModel := new(TaskLog)
	taskLogModel.ReportID = taskModel.ID
	taskLogModel.Name = taskModel.Name
	taskLogModel.Cycle = taskModel.Cycle
	taskLogModel.StartTime = time.Now()
	taskLogModel.Status = status
	taskLogModel.Result = buildTaskLogResult(status, 0, "等待执行")
	insertId, err := taskLogModel.Create()
	return insertId, err
}

// AddTopology insert a new ZmsTopology into database and returns
// last inserted Id on success.
func (m *TaskLog) Create() (id int64, err error) {
	query := DB
	if m.EndTime.IsZero() {
		query = query.Omit("end_time")
	}
	err = query.Create(m).Error
	if err != nil {
		logger.Log.Debug(err)
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
			logger.Log.Debug("Number of records updated in database:", result.RowsAffected)
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
		logger.Log.Debug(err)
		return 0, []TaskLog{}, err
	}

	// 获取分页数据
	offset := (pages - 1) * limits
	err = query.Order("start_time DESC").Limit(limits).Offset(offset).Find(&tasklog).Error
	if err != nil {
		logger.Log.Debug(err)
		return 0, []TaskLog{}, err
	}
	for index := range tasklog {
		progress, detail := parseTaskLogProgressResult(tasklog[index].Result)
		tasklog[index].Result = detail
		switch tasklog[index].Status {
		case Success:
			tasklog[index].Progress = 100
		case Failed:
			if progress > 0 {
				tasklog[index].Progress = progress
			}
		case Running:
			tasklog[index].Progress = progress
		default:
			tasklog[index].Progress = 0
		}
	}
	return cnt, tasklog, nil
}

func GetLatestTaskLogByReportID(reportID int) (*TaskLog, error) {
	taskLog := &TaskLog{}
	err := DB.Where("report_id = ?", reportID).Order("start_time DESC").First(taskLog).Error
	if err != nil {
		return nil, err
	}
	progress, detail := parseTaskLogProgressResult(taskLog.Result)
	taskLog.Result = detail
	switch taskLog.Status {
	case Success:
		taskLog.Progress = 100
	case Failed:
		if progress > 0 {
			taskLog.Progress = progress
		}
	case Running:
		taskLog.Progress = progress
	default:
		taskLog.Progress = 0
	}
	return taskLog, nil
}

func UpdateTaskLogProgress(id int64, progress int, detail string) error {
	return DB.Model(&TaskLog{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status": Running,
		"result": buildTaskLogResult(Running, progress, detail),
	}).Error
}

func FinishTaskLog(id int64, startedAt time.Time, status int, detail, files string) error {
	finishedAt := time.Now()
	return DB.Model(&TaskLog{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     status,
		"result":     buildTaskLogResult(status, 100, detail),
		"files":      files,
		"end_time":   finishedAt,
		"total_time": finishedAt.Unix() - startedAt.Unix(),
	}).Error
}

func MarkStaleRunningTaskLogsFailed(timeout time.Duration) error {
	query := DB.Model(&TaskLog{}).Where("status = ?", Running)
	if timeout > 0 {
		cutoff := time.Now().Add(-timeout)
		query = query.Where("start_time < ?", cutoff)
	}
	return query.Updates(map[string]interface{}{
		"status":     Failed,
		"result":     "任务异常中断，已自动标记为失败",
		"end_time":   time.Now(),
		"total_time": 0,
	}).Error
}

func (taskLog *TaskLog) Clear() (int64, error) {
	result := DB.Delete(&TaskLog{})
	if result.Error != nil {
		logger.Log.Debug(result.Error)
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
