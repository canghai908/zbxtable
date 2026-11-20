package model

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func LoadOnCallSchedule(id int) (OnCallSchedule, []OnCallOverride, error) {
	var s OnCallSchedule
	if id <= 0 {
		return s, nil, errors.New("无效排班ID")
	}
	if err := DB.First(&s, id).Error; err != nil {
		return s, nil, err
	}
	overrides := []OnCallOverride{}
	err := DB.Where("schedule_id = ?", id).Order("id ASC").Find(&overrides).Error
	return s, overrides, err
}
func ValidateScheduleReferences(ids []int) error {
	if err := validatePositiveIDs(ids, "排班", false); err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	var count int64
	if err := DB.Model(&OnCallSchedule{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return errors.New("引用的排班不存在")
	}
	return nil
}

// LockScheduleReferences validates and locks referenced rows in a stable order.
// The caller must keep this transaction open until its rule write completes,
// serializing with schedule deletion without requiring JSON foreign keys.
func LockScheduleReferences(tx *gorm.DB, ids []int) error {
	if err := validatePositiveIDs(ids, "排班", false); err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	var rows []OnCallSchedule
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").Where("id IN ?", ids).Order("id ASC").Find(&rows).Error; err != nil {
		return err
	}
	if len(rows) != len(ids) {
		return errors.New("引用的排班不存在")
	}
	return nil
}

func validateScheduleAccounts(tx *gorm.DB, s OnCallSchedule) error {
	users := map[int]bool{}
	for _, id := range append(append([]int{}, s.Members...), s.FallbackUserIDs...) {
		users[id] = true
	}
	ids := []int{}
	for id := range users {
		ids = append(ids, id)
	}
	var count int64
	if err := tx.Model(&User{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
		return err
	}
	if count != int64(len(ids)) {
		return errors.New("值班成员或兜底用户不存在")
	}
	if len(s.FallbackGroupIDs) > 0 {
		if err := tx.Model(&UserGroup{}).Where("id IN ?", s.FallbackGroupIDs).Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(s.FallbackGroupIDs)) {
			return errors.New("兜底用户组不存在")
		}
	}
	return nil
}
func writeScheduleAudit(tx *gorm.DB, id int, actor, action string, detail any) error {
	raw, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	return tx.Create(&OnCallAudit{ScheduleID: id, Actor: actor, Action: action, Detail: string(raw)}).Error
}
func SaveOnCallSchedule(s *OnCallSchedule, actor string) error {
	s.Name = strings.TrimSpace(s.Name)
	if err := ValidateOnCallSchedule(*s); err != nil {
		return err
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		if err := validateScheduleAccounts(tx, *s); err != nil {
			return err
		}
		action := "create"
		var before OnCallSchedule
		if s.ID > 0 {
			action = "update"
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&before, s.ID).Error; err != nil {
				return err
			}
			s.Created = before.Created
			if err := tx.Model(&before).Select("name", "timezone", "rotation", "start", "members", "weekdays", "window_start", "window_end", "fallback_user_ids", "fallback_group_ids", "enabled").Updates(s).Error; err != nil {
				return err
			}
		} else if err := tx.Create(s).Error; err != nil {
			return err
		}
		return writeScheduleAudit(tx, s.ID, actor, action, map[string]any{"before": before, "after": s})
	})
}
func DeleteOnCallSchedule(id int, actor string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var s OnCallSchedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&s, id).Error; err != nil {
			return err
		}
		var rules []Rule
		if err := tx.Select("id", "name", "schedule_ids").Find(&rules).Error; err != nil {
			return err
		}
		for _, r := range rules {
			for _, sid := range r.ScheduleIDs {
				if sid == id {
					return fmt.Errorf("排班仍被分发规则 %s 引用，请先解除引用", r.Name)
				}
			}
		}
		if err := tx.Where("schedule_id = ?", id).Delete(&OnCallOverride{}).Error; err != nil {
			return err
		}
		if err := tx.Delete(&s).Error; err != nil {
			return err
		}
		return writeScheduleAudit(tx, id, actor, "delete", s)
	})
}
func AddOnCallOverride(o *OnCallOverride, actor string) error {
	if err := ValidateOnCallOverride(*o); err != nil {
		return err
	}
	return DB.Transaction(func(tx *gorm.DB) error {
		var s OnCallSchedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&s, o.ScheduleID).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&User{}).Where("id IN ?", o.UserIDs).Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(o.UserIDs)) {
			return errors.New("代班用户不存在")
		}
		others := []OnCallOverride{}
		if err := tx.Where("schedule_id = ?", o.ScheduleID).Find(&others).Error; err != nil {
			return err
		}
		a, _ := timeParse(o.Start)
		b, _ := timeParse(o.End)
		for _, other := range others {
			x, e := timeParse(other.Start)
			if e != nil {
				return e
			}
			y, e := timeParse(other.End)
			if e != nil {
				return e
			}
			if a.Before(y) && x.Before(b) {
				return errors.New("代班时间与已有代班重叠")
			}
		}
		o.CreatedBy = actor
		if err := tx.Create(o).Error; err != nil {
			return err
		}
		return writeScheduleAudit(tx, o.ScheduleID, actor, "override_create", o)
	})
}
func RemoveOnCallOverride(scheduleID, id int, actor string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var s OnCallSchedule
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&s, scheduleID).Error; err != nil {
			return err
		}
		var o OnCallOverride
		if err := tx.Where("id = ? AND schedule_id = ?", id, scheduleID).First(&o).Error; err != nil {
			return err
		}
		if err := tx.Delete(&o).Error; err != nil {
			return err
		}
		return writeScheduleAudit(tx, scheduleID, actor, "override_delete", o)
	})
}

func timeParse(raw string) (time.Time, error) { return time.Parse(time.RFC3339, raw) }
