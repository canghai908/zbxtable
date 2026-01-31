package response

// 错误码定义
const (
	// 成功
	CodeSuccess = 200

	// 客户端错误 4xx
	CodeBadRequest      = 400 // 错误的请求
	CodeUnauthorized    = 401 // 未授权
	CodeForbidden       = 403 // 禁止访问
	CodeNotFound        = 404 // 资源未找到
	CodeValidationError = 422 // 验证错误

	// 服务器错误 5xx
	CodeInternalError = 500 // 内部服务器错误
	CodeDatabaseError = 501 // 数据库错误
	CodeCacheError    = 502 // 缓存错误
	CodeThirdPartyAPI = 503 // 第三方API错误

	// 业务错误 1xxx
	CodeBusinessError      = 1000 // 通用业务错误
	CodeInstanceNotFound   = 1001 // 实例不存在
	CodeInstanceDisabled   = 1002 // 实例已禁用
	CodeMappingNotFound    = 1003 // 映射配置不存在
	CodeMappingExists      = 1004 // 映射配置已存在
	CodeExecutionFailed    = 1005 // 执行失败
	CodeConfigInvalid      = 1006 // 配置无效
	CodeHostGroupNotFound  = 1007 // 主机组不存在
	CodeTemplateNotFound   = 1008 // 模板不存在
	CodeItemNotFound       = 1009 // 监控项不存在
	CodeZabbixAPIError     = 1010 // Zabbix API错误
	CodePermissionDenied   = 1011 // 权限不足
	CodeResourceInUse      = 1012 // 资源正在使用中
	CodeDuplicateEntry     = 1013 // 重复的记录
)

// 错误消息映射
var ErrorMessages = map[int]string{
	CodeSuccess:            "操作成功",
	CodeBadRequest:         "请求参数错误",
	CodeUnauthorized:       "未授权，请先登录",
	CodeForbidden:          "没有权限访问该资源",
	CodeNotFound:           "请求的资源不存在",
	CodeValidationError:    "数据验证失败",
	CodeInternalError:      "服务器内部错误",
	CodeDatabaseError:      "数据库操作失败",
	CodeCacheError:         "缓存操作失败",
	CodeThirdPartyAPI:      "第三方API调用失败",
	CodeBusinessError:      "业务处理失败",
	CodeInstanceNotFound:   "实例不存在",
	CodeInstanceDisabled:   "实例已被禁用",
	CodeMappingNotFound:    "映射配置不存在",
	CodeMappingExists:      "映射配置已存在",
	CodeExecutionFailed:    "执行失败",
	CodeConfigInvalid:      "配置格式无效",
	CodeHostGroupNotFound:  "主机组不存在",
	CodeTemplateNotFound:   "模板不存在",
	CodeItemNotFound:       "监控项不存在",
	CodeZabbixAPIError:     "Zabbix API调用失败",
	CodePermissionDenied:   "权限不足",
	CodeResourceInUse:      "资源正在使用中，无法删除",
	CodeDuplicateEntry:     "记录已存在",
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(code int) string {
	if msg, ok := ErrorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
