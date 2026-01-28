package models

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"zbxtable/pkg/utils"
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
	fmt.Println(m.Items, m.Cycle)
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
		"name":         m.Name,
		"emails":       m.Emails,
		"items":        m.Items,
		"linkbandwidth": m.LinkBandWidth,
		"host_ids":     m.HostIds,
		"item_ids":     m.ItemIds,
		"cycle":        m.Cycle,
		"status":       m.Status,
		"desc":         m.Desc,
		"start":        m.Start,
		"end":          m.End,
		"report_mode":  m.ReportMode,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

// UpdateTopologyByID updates Alarm by Id and returns error if
func CheckNowByID(m *Report) (err error) {
	var v Report
	err = DB.Where("id = ?", m.ID).First(&v).Error
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
			utils.Log.Error(err)
			return err
		}

		// 实时报表使用配置的开始和结束时间，设置一个临时的Cycle用于文件命名
		vTemp := v
		if len(vTemp.Cycle) == 0 {
			vTemp.Cycle = "realtime" // 用于文件命名区分
		}
		taskErr := TaskHostReport(vTemp)
		if taskErr != nil {
			utils.Log.Error(taskErr)
			v.ExecStatus = strconv.Itoa(Failed)
			v.EndAt = time.Now()
			UpdateReportExecStatusByID(&v)
			return taskErr
		}
		// 更新 report 状态
		v.ExecStatus = strconv.Itoa(Success)
		v.EndAt = time.Now()
		if err := UpdateReportExecStatusByID(&v); err != nil {
			utils.Log.Error(err)
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
				utils.Log.Error(taskErr)
				return taskErr
			}
			// 更新 report 状态
			v.ExecStatus = strconv.Itoa(Success)
			v.StartAt = start
			v.EndAt = time.Now()
			if err := UpdateReportExecStatusByID(&v); err != nil {
				utils.Log.Error(err)
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
				utils.Log.Error(taskErr)
				return taskErr
			}
			// 更新 report 状态
			v.ExecStatus = strconv.Itoa(Success)
			v.StartAt = start
			v.EndAt = time.Now()
			if err := UpdateReportExecStatusByID(&v); err != nil {
				utils.Log.Error(err)
				return err
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
			utils.Log.Error(err)
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
			utils.Log.Error(err)
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
		utils.Log.Debug("Number of records deleted in database:", result.RowsAffected)
	}
	return
}
