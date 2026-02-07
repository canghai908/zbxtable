# Zabbix 实例标识重构方案

## 重构目标

将多租户标识系统改为实例标识系统：
1. `ZabbixTenant` → `ZabbixInstance`
2. `ID` → `ZID` (主键)
3. `TenantID` → `InstanceID` (实例标识符)
4. 所有 API 参数 `tenant_id` → `zid`

## 数据库迁移

### 迁移 SQL

```sql
-- 1. 重命名列（保持数据）
ALTER TABLE zabbix CHANGE COLUMN tenant_id instance_id VARCHAR(255) NOT NULL;

-- 2. 更新索引
ALTER TABLE zabbix DROP INDEX tenant_id;
ALTER TABLE zabbix ADD UNIQUE INDEX instance_id (instance_id);

-- 3. 更新其他表中的外键引用（如果有）
-- 根据实际情况调整
```

## 后端修改清单

### 核心模型文件
- [x] `internal/model/zabbix_mod.go` - 结构体定义
- [x] `internal/model/zabbix.go` - CRUD 函数

### API 池管理
- [ ] `internal/model/api_pool.go` - API 连接池

### Handler 层
- [ ] `internal/handler/zabbix.go` - Zabbix 管理 API
- [ ] `internal/handler/host.go` - 主机相关 API
- [ ] `internal/handler/index.go` - 首页 API
- [ ] `internal/handler/system.go` - 系统配置 API
- [ ] 其他使用 tenant_id 的 handler

### Model 层业务逻辑
- [ ] `internal/model/egress.go` - 出口配置
- [ ] `internal/model/host.go` - 主机管理
- [ ] `internal/model/alarm.go` - 告警管理
- [ ] `internal/model/system.go` - 系统配置
- [ ] 其他使用 tenant_id 的 model

## 前端修改清单

### API 服务
- [ ] `src/services/api.js` - API 路径定义
- [ ] `src/services/admin.js` - API 调用函数
- [ ] `src/services/zabbix.js` - Zabbix 相关 API

### 页面组件
- [ ] `src/pages/system/zabbix.vue` - Zabbix 管理页面
- [ ] `src/pages/system/bandwidth.vue` - 出口配置页面
- [ ] `src/pages/dashboard/index.vue` - 首页
- [ ] 其他使用 tenant_id 的页面

### 组件
- [ ] `src/components/egress/EgressBandwidth.vue` - 出口带宽组件
- [ ] 其他使用 tenant_id 的组件

## 修改策略

### 1. 数据库字段映射
- 数据库列名：`instance_id` (从 `tenant_id` 重命名)
- Go 结构体字段：`InstanceID`
- JSON 字段：`instance_id`
- API 参数：`zid` (查询时使用主键 ID)

### 2. API 参数变更
```
旧: GET /v1/host/search?tenant_id=zabbix48
新: GET /v1/host/search?zid=1
```

### 3. 前端数据结构
```javascript
// 旧
{
  id: 1,
  tenant_id: "zabbix48",
  name: "Zabbix 4.8"
}

// 新
{
  zid: 1,
  instance_id: "zabbix48",
  name: "Zabbix 4.8"
}
```

## 关键修改点

### 1. API 池管理 (api_pool.go)
```go
// 旧
func GetZabbixInstanceAPI(tenantID string) (*APIInstance, error)

// 新
func GetZabbixInstanceAPI(zid int) (*APIInstance, error)
```

### 2. 查询方式变更
```go
// 旧：通过字符串 tenant_id 查询
DB.Where("tenant_id = ?", tenantID).First(&instance)

// 新：通过整数 zid 查询
DB.First(&instance, zid)
```

### 3. 前端 API 调用
```javascript
// 旧
hostSearch({ tenant_id: record.tenant_id })

// 新
hostSearch({ zid: record.zid })
```

## 测试计划

### 单元测试
- [ ] 测试实例 CRUD 操作
- [ ] 测试 API 池管理
- [ ] 测试主机搜索

### 集成测试
- [ ] 测试 Zabbix 实例管理页面
- [ ] 测试出口配置页面
- [ ] 测试首页数据展示
- [ ] 测试多实例切换

### 回归测试
- [ ] 测试所有使用实例的功能
- [ ] 测试数据迁移正确性

## 风险评估

### 高风险
1. 数据库迁移可能导致数据丢失
2. API 参数变更可能导致前后端不兼容
3. 大量文件修改可能引入新 bug

### 缓解措施
1. 备份数据库
2. 分阶段提交代码
3. 充分测试后再部署

## 实施步骤

1. ✅ 修改核心模型定义
2. ⏳ 修改 API 池管理
3. ⏳ 修改 Handler 层
4. ⏳ 修改 Model 层
5. ⏳ 修改前端 API 服务
6. ⏳ 修改前端页面组件
7. ⏳ 数据库迁移
8. ⏳ 编译测试
9. ⏳ 生成文档

## 预计工作量

- 后端修改：约 50+ 文件
- 前端修改：约 20+ 文件
- 测试验证：2-3 小时
- 总计：4-6 小时
