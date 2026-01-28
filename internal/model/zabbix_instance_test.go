package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 创建测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&ZabbixInstance{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestZabbixInstance_TableName(t *testing.T) {
	inst := &ZabbixInstance{}
	tableName := inst.TableName()
	assert.Contains(t, tableName, "zabbix_instance")
}

func TestCreateZabbixInstance(t *testing.T) {
	// 备份原始 DB
	originalDB := DB
	defer func() { DB = originalDB }()

	// 设置测试数据库
	DB = setupTestDB(t)

	inst := &ZabbixInstance{
		Name:    "Test Zabbix",
		WebURL:  "http://test.example.com",
		User:    "admin",
		Pass:    "password",
		Token:   "",
		Enabled: true,
	}

	err := CreateZabbixInstance(inst)
	assert.NoError(t, err)
	assert.NotZero(t, inst.ID)
	assert.Equal(t, "http://test.example.com", inst.WebURL)
}

func TestListZabbixInstances(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupTestDB(t)

	// 创建测试数据
	instances := []*ZabbixInstance{
		{Name: "Zabbix 1", WebURL: "http://zabbix1.com", Enabled: true},
		{Name: "Zabbix 2", WebURL: "http://zabbix2.com", Enabled: false},
	}

	for _, inst := range instances {
		err := CreateZabbixInstance(inst)
		assert.NoError(t, err)
	}

	// 测试列表查询
	list, err := ListZabbixInstances()
	assert.NoError(t, err)
	assert.Len(t, list, 2)
}

func TestGetZabbixInstanceByID(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupTestDB(t)

	// 创建测试实例
	inst := &ZabbixInstance{
		Name:   "Test Instance",
		WebURL: "http://test.com",
		User:   "admin",
		Pass:   "secret",
		Token:  "test-token",
	}
	err := CreateZabbixInstance(inst)
	assert.NoError(t, err)

	// 测试获取
	retrieved, err := GetZabbixInstanceByID(inst.ID)
	assert.NoError(t, err)
	assert.Equal(t, inst.Name, retrieved.Name)
	assert.Equal(t, inst.WebURL, retrieved.WebURL)
	assert.Equal(t, inst.User, retrieved.User)
	assert.Equal(t, inst.Pass, retrieved.Pass)
	assert.Equal(t, inst.Token, retrieved.Token)
}

func TestSetZabbixInstanceEnabled(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupTestDB(t)

	inst := &ZabbixInstance{
		Name:     "Test Instance",
		WebURL:   "http://test.com",
		Enabled:  true,
		IsActive: true,
	}
	err := CreateZabbixInstance(inst)
	assert.NoError(t, err)

	// 测试禁用
	updated, err := SetZabbixInstanceEnabled(inst.ID, false)
	assert.NoError(t, err)
	assert.False(t, updated.Enabled)
	assert.False(t, updated.IsActive) // 禁用时应该取消激活状态
}

func TestUpdateZabbixInstance(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupTestDB(t)

	inst := &ZabbixInstance{
		Name:    "Original Name",
		WebURL:  "http://original.com",
		User:    "admin",
		Pass:    "pass",
		Enabled: true,
	}
	err := CreateZabbixInstance(inst)
	assert.NoError(t, err)

	// 测试更新
	patch := &ZabbixInstance{
		Name:    "Updated Name",
		WebURL:  "http://updated.com",
		User:    "newuser",
		Pass:    "newpass",
		Enabled: false,
	}

	updated, err := UpdateZabbixInstance(inst.ID, patch)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "http://updated.com", updated.WebURL)
	assert.Equal(t, "newuser", updated.User)
	assert.Equal(t, "newpass", updated.Pass)
	assert.False(t, updated.Enabled)
}

func TestDeleteZabbixInstance(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupTestDB(t)

	inst := &ZabbixInstance{
		Name:     "Test Instance",
		WebURL:   "http://test.com",
		IsActive: true,
	}
	err := CreateZabbixInstance(inst)
	assert.NoError(t, err)

	// 测试删除
	err = DeleteZabbixInstance(inst.ID)
	assert.NoError(t, err)

	// 验证已删除
	_, err = GetZabbixInstanceByID(inst.ID)
	assert.Error(t, err)
}

func TestActivateZabbixInstance(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupTestDB(t)

	// 创建两个实例
	inst1 := &ZabbixInstance{
		Name:     "Instance 1",
		WebURL:   "http://inst1.com",
		Enabled:  true,
		IsActive: true,
	}
	inst2 := &ZabbixInstance{
		Name:    "Instance 2",
		WebURL:  "http://inst2.com",
		Enabled: true,
	}

	err := CreateZabbixInstance(inst1)
	assert.NoError(t, err)
	err = CreateZabbixInstance(inst2)
	assert.NoError(t, err)

	// 激活第二个实例（需要 mock API，这里只测试数据库逻辑）
	// 注意：实际测试需要 mock ApplyActiveZabbix 函数
	// 这里我们只测试禁用实例无法激活的情况
	inst2.Enabled = false
	DB.Save(inst2)

	_, err = ActivateZabbixInstance(inst2.ID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "已禁用")
}

func TestGetActiveZabbixInstance(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupTestDB(t)

	// 创建激活的实例
	inst := &ZabbixInstance{
		Name:     "Active Instance",
		WebURL:   "http://active.com",
		Enabled:  true,
		IsActive: true,
	}
	err := CreateZabbixInstance(inst)
	assert.NoError(t, err)

	// 测试获取激活实例
	active, err := GetActiveZabbixInstance()
	assert.NoError(t, err)
	assert.NotNil(t, active)
	assert.Equal(t, inst.ID, active.ID)
	assert.True(t, active.IsActive)
}

func TestZabbixInstance_Timestamps(t *testing.T) {
	originalDB := DB
	defer func() { DB = originalDB }()

	DB = setupTestDB(t)

	inst := &ZabbixInstance{
		Name:   "Test Instance",
		WebURL: "http://test.com",
	}

	beforeCreate := time.Now()
	err := CreateZabbixInstance(inst)
	assert.NoError(t, err)
	afterCreate := time.Now()

	// 验证时间戳
	assert.True(t, inst.CreatedAt.After(beforeCreate) || inst.CreatedAt.Equal(beforeCreate))
	assert.True(t, inst.CreatedAt.Before(afterCreate) || inst.CreatedAt.Equal(afterCreate))
	assert.True(t, inst.UpdatedAt.After(beforeCreate) || inst.UpdatedAt.Equal(beforeCreate))
	assert.True(t, inst.UpdatedAt.Before(afterCreate) || inst.UpdatedAt.Equal(afterCreate))
}

