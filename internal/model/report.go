package model

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/logger"
	"zbxtable/pkg/utils"

	"gorm.io/gorm"
)

// TableName alarm
func (t *Report) TableName() string {
	return TableName("report")
}

// GetReportsByID f
func GetReportsByID(id int) (v *Report, err error) {
	v = &Report{}
	err = DB.Where("id = ?", id).First(v).Error
	if err != nil {
		return nil, err
	}
	return v, nil
}

// get all
func GetALlReport() (cnt int64, system []Report, err error) {
	var sys []Report
	err = DB.Find(&sys).Error
	if err != nil {
		return 0, []Report{}, err
	}
	cnt = int64(len(sys))
	return cnt, sys, nil
}

// GetAllTopology t
func GetAllReportsLimt(page, limit, name, reportType string) (cnt int64, topo []Report, err error) {
	var topologys []Report
	pages, _ := strconv.Atoi(page)
	limits, _ := strconv.Atoi(limit)
	if limits == 0 {
		limits = 10
	}
	if pages == 0 {
		pages = 1
	}

	query := DB.Model(&Report{})

	if name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if reportType != "" {
		query = query.Where("report_type = ?", reportType)
	}

	// 获取总数
	err = query.Count(&cnt).Error
	if err != nil {
		return 0, []Report{}, err
	}

	// 获取分页数据
	offset := (pages - 1) * limits
	err = query.Order("created_at").Limit(limits).Offset(offset).Find(&topologys).Error
	if err != nil {
		return 0, []Report{}, err
	}
	return cnt, topologys, nil
}

// AddTopology insert a new ZmsTopology into database and returns
// last inserted Id on success.
func AddReport(m *Report) (id int64, err error) {
	m.Items = utils.VAarToStr(m.Items)
	m.Cycle = utils.VAarToStr(m.Cycle)
	err = DB.Create(m).Error
	if err != nil {
		return 0, err
	}
	return int64(m.ID), err
}

// UpdateTopologyByID updates Alarm by Id and returns error if
func UpdateReportByID(m *Report) (err error) {
	var v Report
	err = DB.Where("id = ?", m.ID).First(&v).Error
	if err != nil {
		return err
	}
	m.Items = utils.VAarToStr(m.Items)
	m.Cycle = utils.VAarToStr(m.Cycle)
	err = DB.Model(&Report{}).Where("id = ?", m.ID).Updates(map[string]interface{}{
		"name":            m.Name,
		"emails":          m.Emails,
		"items":           m.Items,
		"link_band_width": m.LinkBandWidth,
		"host_ids":        m.HostIds,
		"resolved_host_ids": func() string {
			if m.ReportType == "host" {
				return ""
			}
			return v.ResolvedHostIds
		}(),
		"item_ids":    m.ItemIds,
		"cycle":       m.Cycle,
		"status":      m.Status,
		"desc":        m.Desc,
		"start":       m.Start,
		"end":         m.End,
		"report_mode": m.ReportMode,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func UpdateReportHostConfigByID(id int, hostIDs, itemIDs string) error {
	return DB.Model(&Report{}).Where("id = ?", id).Updates(map[string]interface{}{
		"host_ids": hostIDs,
		"item_ids": itemIDs,
	}).Error
}

func UpdateReportItemIDsByID(id int, itemIDs string) error {
	return DB.Model(&Report{}).Where("id = ?", id).Update("item_ids", itemIDs).Error
}

func UpdateReportResolvedHostConfigByID(id int, resolvedHostIDs, itemIDs string) error {
	return DB.Model(&Report{}).Where("id = ?", id).Updates(map[string]interface{}{
		"resolved_host_ids": resolvedHostIDs,
		"item_ids":          itemIDs,
	}).Error
}

// UpdateTopologyByID updates Alarm by Id and returns error if
func CheckNowByID(m *Report) (taskLogID int64, err error) {
	var v Report
	err = DB.Where("id = ?", m.ID).First(&v).Error
	if err != nil {
		return 0, err
	}

	if err := ResetStaleRunningReports(15 * time.Minute); err != nil {
		logger.Log.Error(err)
	}
	if err := DB.Where("id = ?", m.ID).First(&v).Error; err != nil {
		return 0, err
	}

	if v.ExecStatus == strconv.Itoa(Running) {
		latestTaskLog, taskErr := GetLatestTaskLogByReportID(v.ID)
		if taskErr != nil {
			if !errors.Is(taskErr, gorm.ErrRecordNotFound) {
				return 0, taskErr
			}
		}
		if latestTaskLog == nil || latestTaskLog.Status != Running {
			finishedAt := time.Now()
			v.ExecStatus = strconv.Itoa(Failed)
			v.EndAt = &finishedAt
			if err := UpdateReportExecStatusByID(&v); err != nil {
				return 0, err
			}
		} else {
			return 0, fmt.Errorf("报表正在处理中，请稍后重试")
		}
	}

	if v.ReportType != "host" {
		start := time.Now()
		v.ExecStatus = strconv.Itoa(Running)
		v.StartAt = &start
		v.EndAt = nil
		if err := UpdateReportExecStatusByID(&v); err != nil {
			logger.Log.Error(err)
			return 0, err
		}
		runErr := runReportNow(&v, nil)
		finishedAt := time.Now()
		v.EndAt = &finishedAt
		if runErr != nil {
			v.ExecStatus = strconv.Itoa(Failed)
		} else {
			v.ExecStatus = strconv.Itoa(Success)
		}
		if err := UpdateReportExecStatusByID(&v); err != nil {
			logger.Log.Error(err)
			return 0, err
		}
		return 0, runErr
	}

	start := time.Now()
	v.ExecStatus = strconv.Itoa(Running)
	v.StartAt = &start
	v.EndAt = nil
	if err := UpdateReportExecStatusByID(&v); err != nil {
		logger.Log.Error(err)
		return 0, err
	}

	taskLogID, err = CreateTaskLog(v, Running)
	if err != nil {
		finishedAt := time.Now()
		v.ExecStatus = strconv.Itoa(Failed)
		v.EndAt = &finishedAt
		if updateErr := UpdateReportExecStatusByID(&v); updateErr != nil {
			logger.Log.Error(updateErr)
		}
		logger.Log.Error(err)
		return 0, err
	}

	go func(report Report, logID int64) {
		task := &TaskLog{
			Id:        int(logID),
			ReportID:  report.ID,
			Name:      report.Name,
			Cycle:     report.Cycle,
			StartTime: start,
			Status:    Running,
		}
		runErr := runReportNow(&report, task)
		finishedAt := time.Now()
		report.EndAt = &finishedAt
		if runErr != nil {
			report.ExecStatus = strconv.Itoa(Failed)
		} else if report.ReportType == "host" && report.ReportMode != "realtime" {
			report.ExecStatus = strconv.Itoa(NoBegin)
		} else {
			report.ExecStatus = strconv.Itoa(Success)
		}
		if updateErr := UpdateReportExecStatusByID(&report); updateErr != nil {
			logger.Log.Error(updateErr)
		}
	}(v, taskLogID)

	return taskLogID, nil
}

func ResetStaleRunningReports(timeout time.Duration) error {
	query := DB.Model(&Report{}).Where("exec_status = ?", strconv.Itoa(Running))
	if timeout > 0 {
		cutoff := time.Now().Add(-timeout)
		query = query.Where("start_at < ?", cutoff)
	}
	if err := query.Updates(map[string]interface{}{
		"exec_status": strconv.Itoa(Failed),
		"end_at":      time.Now(),
	}).Error; err != nil {
		return err
	}
	return MarkStaleRunningTaskLogsFailed(timeout)
}

func runReportNow(v *Report, task *TaskLog) error {
	if v == nil {
		return fmt.Errorf("报表不存在")
	}

	if v.ReportType == "host" && HasBulkHostReportConfig(v.HostIds) {
		if err := PrepareHostReportExecutionConfig(v); err != nil {
			return err
		}
	}

	// 实时报表：直接生成，不需要周期
	if v.ReportMode == "realtime" && v.ReportType == "host" {
		vTemp := *v
		if len(vTemp.Cycle) == 0 {
			vTemp.Cycle = "realtime" // 用于文件命名区分
		}
		return TaskHostReportWithTaskLog(vTemp, task)
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
			var taskErr error
			if v.ReportType == "host" {
				vDay := *v
				vDay.Cycle = "day"
				taskErr = TaskHostReportWithTaskLog(vDay, task)
			} else {
				taskErr = TaskDayReport(*v)
			}
			if taskErr != nil {
				logger.Log.Error(taskErr)
				return taskErr
			}
		}

		// week
		if vv == "week" {
			var taskErr error
			if v.ReportType == "host" {
				vWeek := *v
				vWeek.Cycle = "week"
				taskErr = TaskHostReportWithTaskLog(vWeek, task)
			} else {
				taskErr = TaskWeekReport(*v)
			}
			if taskErr != nil {
				logger.Log.Error(taskErr)
				return taskErr
			}
		}
	}

	return nil
}

// UpdateTopologyByID updates Alarm by Id and returns error if
func UpdateReportExecStatusByID(m *Report) (err error) {
	var v Report
	err = DB.Where("id = ?", m.ID).First(&v).Error
	if err == nil {
		err = DB.Model(&Report{}).Where("id = ?", m.ID).Updates(map[string]interface{}{
			"exec_status": m.ExecStatus,
			"start_at":    m.StartAt,
			"end_at":      m.EndAt,
		}).Error
		if err != nil {
			logger.Log.Error(err)
		}
	}
	return
}

// UpdateTopologyEdgesByID updates Alarm by Id and returns error if
func UpdateReportsStatusByID(m *Report) (err error) {
	var v Report
	err = DB.Where("id = ?", m.ID).First(&v).Error
	if err == nil {
		newStatus := "1"
		if v.Status == "0" {
			newStatus = "1"
		} else {
			newStatus = "0"
		}
		err = DB.Model(&Report{}).Where("id = ?", m.ID).Update("status", newStatus).Error
		if err != nil {
			logger.Log.Error(err)
		}
	}
	return
}

// DeleteAlarm deletes Alarm by Id and returns error if
// the record to be deleted doesn't exist
func DeleteReport(id int) (err error) {
	var v Report
	err = DB.Where("id = ?", id).First(&v).Error
	if err == nil {
		result := DB.Delete(&Report{}, id)
		if result.Error != nil {
			return result.Error
		}
		logger.Log.Debug("Number of records deleted in database:", result.RowsAffected)
	}
	return
}
