package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupRuleTestDB 创建规则测试数据库
func setupRuleTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&Rule{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestRule_TableName(t *testing.T) {
	rule := &Rule{}
	tableName := rule.TableName()
	assert.Contains(t, tableName, "rule")
}

func TestCreateRule(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupRuleTestDB(t)

	rule := &Rule{
		Name:       "Test Rule",
		TenantID:   "tenant1",
		MType:      "1",
		Conditions: `[{"r_type":"host","r_func":"==","r_value":"test"}]`,
		Sweek:      "1,2,3,4,5",
		Stime:      "09:00",
		Etime:      "18:00",
		Channel:    "mail,wechat",
		UserIds:    "1,2,3",
		GroupIds:   "1,2",
		Note:       "Test note",
		Status:     "0",
	}

	err := DB.Create(rule).Error
	assert.NoError(t, err)
	assert.NotZero(t, rule.ID)
	assert.Equal(t, "Test Rule", rule.Name)
	assert.Equal(t, "tenant1", rule.TenantID)
}

func TestRule_DefaultRule(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupRuleTestDB(t)

	// 创建默认规则 (m_type = 2)
	defaultRule := &Rule{
		Name:       "Default Rule",
		TenantID:   "*",
		MType:      "2",
		Conditions: `[]`,
		Channel:    "mail",
		Status:     "0",
	}

	err := DB.Create(defaultRule).Error
	assert.NoError(t, err)
	assert.Equal(t, "2", defaultRule.MType)
	assert.Equal(t, "*", defaultRule.TenantID)
}

func TestRule_MultiTenant(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupRuleTestDB(t)

	// 创建多租户规则
	rule := &Rule{
		Name:     "Multi-Tenant Rule",
		TenantID: "tenant1,tenant2,tenant3",
		MType:    "1",
		Channel:  "mail",
		Status:   "0",
	}

	err := DB.Create(rule).Error
	assert.NoError(t, err)
	assert.Contains(t, rule.TenantID, "tenant1")
	assert.Contains(t, rule.TenantID, "tenant2")
	assert.Contains(t, rule.TenantID, "tenant3")
}

func TestRule_StatusToggle(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupRuleTestDB(t)

	rule := &Rule{
		Name:    "Test Rule",
		MType:   "1",
		Channel: "mail",
		Status:  "0", // 启用
	}

	err := DB.Create(rule).Error
	assert.NoError(t, err)

	// 禁用规则
	rule.Status = "1"
	err = DB.Save(rule).Error
	assert.NoError(t, err)

	// 验证状态
	var retrieved Rule
	err = DB.First(&retrieved, rule.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "1", retrieved.Status)
}

func TestRule_QueryByTenant(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupRuleTestDB(t)

	// 创建多个规则
	rules := []*Rule{
		{Name: "Rule 1", TenantID: "tenant1", MType: "1", Channel: "mail", Status: "0"},
		{Name: "Rule 2", TenantID: "tenant2", MType: "1", Channel: "mail", Status: "0"},
		{Name: "Rule 3", TenantID: "tenant1,tenant2", MType: "1", Channel: "mail", Status: "0"},
	}

	for _, rule := range rules {
		err := DB.Create(rule).Error
		assert.NoError(t, err)
	}

	// 查询 tenant1 的规则
	var tenant1Rules []Rule
	err := DB.Where("tenant_id LIKE ?", "%tenant1%").Find(&tenant1Rules).Error
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(tenant1Rules), 2)
}

func TestRule_ConditionsJSON(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupRuleTestDB(t)

	conditions := `[
		{"r_type":"host","r_func":"==","r_value":"server1"},
		{"r_type":"severity","r_func":"==","r_value":"4"}
	]`

	rule := &Rule{
		Name:       "Complex Rule",
		TenantID:   "tenant1",
		MType:      "1",
		Conditions: conditions,
		Channel:    "mail,wechat",
		Status:     "0",
	}

	err := DB.Create(rule).Error
	assert.NoError(t, err)

	// 验证 JSON 存储
	var retrieved Rule
	err = DB.First(&retrieved, rule.ID).Error
	assert.NoError(t, err)
	assert.JSONEq(t, conditions, retrieved.Conditions)
}

func TestRule_DeleteCascade(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupRuleTestDB(t)

	rule := &Rule{
		Name:    "Test Rule",
		MType:   "1",
		Channel: "mail",
		Status:  "0",
	}

	err := DB.Create(rule).Error
	assert.NoError(t, err)

	ruleID := rule.ID

	// 删除规则
	err = DB.Delete(rule).Error
	assert.NoError(t, err)

	// 验证已删除
	var count int64
	err = DB.Model(&Rule{}).Where("id = ?", ruleID).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestRule_UpdateFields(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupRuleTestDB(t)

	rule := &Rule{
		Name:     "Original Name",
		TenantID: "tenant1",
		MType:    "1",
		Channel:  "mail",
		UserIds:  "1,2",
		GroupIds: "1",
		Status:   "0",
	}

	err := DB.Create(rule).Error
	assert.NoError(t, err)

	// 更新字段
	updates := map[string]interface{}{
		"name":      "Updated Name",
		"tenant_id": "tenant2",
		"channel":   "mail,wechat",
		"user_ids":  "1,2,3,4",
		"group_ids": "1,2,3",
	}

	err = DB.Model(rule).Updates(updates).Error
	assert.NoError(t, err)

	// 验证更新
	var retrieved Rule
	err = DB.First(&retrieved, rule.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", retrieved.Name)
	assert.Equal(t, "tenant2", retrieved.TenantID)
	assert.Equal(t, "mail,wechat", retrieved.Channel)
	assert.Equal(t, "1,2,3,4", retrieved.UserIds)
	assert.Equal(t, "1,2,3", retrieved.GroupIds)
}

