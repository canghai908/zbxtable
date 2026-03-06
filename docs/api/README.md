# ZbxTable API 文档

## 概述

ZbxTable 提供 RESTful API，支持通过 HTTP 请求访问系统功能。

## 基础信息

- **Base URL**: `http://your-domain:8088/api/v1`
- **认证方式**: Bearer Token
- **请求格式**: JSON
- **响应格式**: JSON

## 认证

所有 API 请求需要在 Header 中携带 Token：

```http
Tokenyour_token_here
```

获取 Token：

```http
POST /api/v1/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expire": "2024-01-31T23:59:59Z"
  }
}
```

## 通用响应格式

### 成功响应

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

### 错误响应

```json
{
  "code": 1001,
  "message": "error message",
  "data": null
}
```

### 错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1001 | 参数错误 |
| 1002 | 认证失败 |
| 1003 | 权限不足 |
| 1004 | 资源不存在 |
| 1005 | 服务器错误 |

## API 接口

### 用户管理

#### 登录

```http
POST /api/v1/login
```

#### 登出

```http
POST /api/v1/logout
```

#### 获取用户信息

```http
GET /api/v1/user/info
```

### 主机管理

#### 获取主机列表

```http
GET /api/v1/hosts?page=1&page_size=20
```

#### 获取主机详情

```http
GET /api/v1/hosts/:id
```

### 告警管理

#### 获取告警列表

```http
GET /api/v1/alarms?start_time=2024-01-01&end_time=2024-01-31
```

### 报表管理

#### 生成报表

```http
POST /api/v1/reports
```

更多 API 接口正在完善中...

