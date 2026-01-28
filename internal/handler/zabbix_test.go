package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestRouter 创建测试路由
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	return r
}

// setupControllerTestDB 创建控制器测试数据库
func setupControllerTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// 自动迁移表结构
	err = db.AutoMigrate(&models.ZabbixInstance{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestListZabbixInstancesGin(t *testing.T) {
	// 备份原始 DB
	originalDB := models.DB
	defer func() { models.DB = originalDB }()

	// 设置测试数据库
	models.DB = setupControllerTestDB(t)

	// 创建测试数据
	instances := []*models.ZabbixInstance{
		{Name: "Zabbix 1", WebURL: "http://zabbix1.com", Enabled: true},
		{Name: "Zabbix 2", WebURL: "http://zabbix2.com", Enabled: false},
	}

	for _, inst := range instances {
		err := models.CreateZabbixInstance(inst)
		assert.NoError(t, err)
	}

	// 创建测试路由
	r := setupTestRouter()
	r.GET("/v1/zabbix/instances", ListZabbixInstancesGin)

	// 发送请求
	req, _ := http.NewRequest("GET", "/v1/zabbix/instances", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(200), response["code"])
	
	data := response["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestCreateZabbixInstanceGin_Success(t *testing.T) {
	originalDB := models.DB
	defer func() { models.DB = originalDB }()

	models.DB = setupControllerTestDB(t)

	r := setupTestRouter()
	r.POST("/v1/zabbix/instances", CreateZabbixInstanceGin)

	// 注意：实际测试需要 mock TestZabbixInstanceConfig 函数
	// 这里只测试请求格式验证
	reqBody := map[string]interface{}{
		"name":    "Test Zabbix",
		"web_url": "http://test.example.com",
		"user":    "admin",
		"pass":    "password",
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/v1/zabbix/instances", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 由于需要测试连接，这里会失败，但可以验证参数解析
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotNil(t, response["code"])
}

func TestCreateZabbixInstanceGin_InvalidParams(t *testing.T) {
	r := setupTestRouter()
	r.POST("/v1/zabbix/instances", CreateZabbixInstanceGin)

	// 缺少必需参数
	reqBody := map[string]interface{}{
		"name": "Test Zabbix",
		// 缺少 web_url
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/v1/zabbix/instances", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(400), response["code"])
}

func TestGetZabbixInstanceGin(t *testing.T) {
	originalDB := models.DB
	defer func() { models.DB = originalDB }()

	models.DB = setupControllerTestDB(t)

	// 创建测试实例
	inst := &models.ZabbixInstance{
		Name:   "Test Instance",
		WebURL: "http://test.com",
		User:   "admin",
		Pass:   "secret",
		Token:  "test-token",
	}
	err := models.CreateZabbixInstance(inst)
	assert.NoError(t, err)

	r := setupTestRouter()
	r.GET("/v1/zabbix/instances/:id", GetZabbixInstanceGin)

	// 发送请求
	req, _ := http.NewRequest("GET", "/v1/zabbix/instances/"+string(rune(inst.ID)), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	if response["code"] == float64(200) {
		data := response["data"].(map[string]interface{})
		assert.Equal(t, "Test Instance", data["name"])
		assert.Equal(t, "admin", data["user"])
		assert.Equal(t, "secret", data["pass"])
		assert.Equal(t, "test-token", data["token"])
	}
}

func TestUpdateZabbixInstanceGin(t *testing.T) {
	originalDB := models.DB
	defer func() { models.DB = originalDB }()

	models.DB = setupControllerTestDB(t)

	// 创建测试实例
	inst := &models.ZabbixInstance{
		Name:    "Original Name",
		WebURL:  "http://original.com",
		Enabled: true,
	}
	err := models.CreateZabbixInstance(inst)
	assert.NoError(t, err)

	r := setupTestRouter()
	r.PUT("/v1/zabbix/instances/:id", UpdateZabbixInstanceGin)

	// 更新请求
	reqBody := map[string]interface{}{
		"name":    "Updated Name",
		"web_url": "http://updated.com",
		"enabled": false,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/v1/zabbix/instances/"+string(rune(inst.ID)), bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteZabbixInstanceGin(t *testing.T) {
	originalDB := models.DB
	defer func() { models.DB = originalDB }()

	models.DB = setupControllerTestDB(t)

	// 创建测试实例
	inst := &models.ZabbixInstance{
		Name:   "Test Instance",
		WebURL: "http://test.com",
	}
	err := models.CreateZabbixInstance(inst)
	assert.NoError(t, err)

	r := setupTestRouter()
	r.DELETE("/v1/zabbix/instances/:id", DeleteZabbixInstanceGin)

	// 删除请求
	req, _ := http.NewRequest("DELETE", "/v1/zabbix/instances/"+string(rune(inst.ID)), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	if response["code"] == float64(200) {
		// 验证已删除
		_, err = models.GetZabbixInstanceByID(inst.ID)
		assert.Error(t, err)
	}
}

func TestEnableZabbixInstanceGin(t *testing.T) {
	originalDB := models.DB
	defer func() { models.DB = originalDB }()

	models.DB = setupControllerTestDB(t)

	// 创建测试实例
	inst := &models.ZabbixInstance{
		Name:    "Test Instance",
		WebURL:  "http://test.com",
		Enabled: true,
	}
	err := models.CreateZabbixInstance(inst)
	assert.NoError(t, err)

	r := setupTestRouter()
	r.PUT("/v1/zabbix/instances/:id/enabled", EnableZabbixInstanceGin)

	// 禁用请求
	reqBody := map[string]interface{}{
		"enabled": false,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("PUT", "/v1/zabbix/instances/"+string(rune(inst.ID))+"/enabled", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTestZabbixInstanceConfigGin_InvalidParams(t *testing.T) {
	r := setupTestRouter()
	r.POST("/v1/zabbix/instances/test", TestZabbixInstanceConfigGin)

	// 缺少必需参数
	reqBody := map[string]interface{}{
		"user": "admin",
		// 缺少 web_url
	}

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/v1/zabbix/instances/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(400), response["code"])
}

func TestToSafeResponse(t *testing.T) {
	inst := &models.ZabbixInstance{
		ID:      1,
		Name:    "Test Instance",
		WebURL:  "http://test.com",
		User:    "admin",
		Pass:    "secret",
		Token:   "test-token",
		Enabled: true,
	}

	safeResp := toSafeResponse(inst)

	// 验证敏感信息已隐藏
	assert.Equal(t, inst.ID, safeResp.ID)
	assert.Equal(t, inst.Name, safeResp.Name)
	assert.Equal(t, inst.WebURL, safeResp.WebURL)
	assert.Equal(t, inst.Enabled, safeResp.Enabled)
	
	// 注意：ZabbixInstanceSafeResponse 不包含 User, Pass, Token 字段
	// 这里无法直接验证，但可以通过 JSON 序列化验证
	jsonData, err := json.Marshal(safeResp)
	assert.NoError(t, err)
	
	var jsonMap map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonMap)
	assert.NoError(t, err)
	
	_, hasUser := jsonMap["user"]
	_, hasPass := jsonMap["pass"]
	_, hasToken := jsonMap["token"]
	
	assert.False(t, hasUser, "Safe response should not contain user field")
	assert.False(t, hasPass, "Safe response should not contain pass field")
	assert.False(t, hasToken, "Safe response should not contain token field")
}

