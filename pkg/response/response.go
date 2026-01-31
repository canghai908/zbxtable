package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// PageData 分页数据结构
type PageData struct {
	Items interface{} `json:"items"`
	Total int64       `json:"total"`
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data:    data,
	})
}

// SuccessWithMessage 成功响应（自定义消息）
func SuccessWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// SuccessWithPage 成功响应（分页数据）
func SuccessWithPage(c *gin.Context, items interface{}, total int64) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: "success",
		Data: PageData{
			Items: items,
			Total: total,
		},
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// ErrorWithData 错误响应（带数据）
func ErrorWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

// BadRequest 400 错误请求
func BadRequest(c *gin.Context, message string) {
	Error(c, CodeBadRequest, message)
}

// Unauthorized 401 未授权
func Unauthorized(c *gin.Context, message string) {
	Error(c, CodeUnauthorized, message)
}

// Forbidden 403 禁止访问
func Forbidden(c *gin.Context, message string) {
	Error(c, CodeForbidden, message)
}

// NotFound 404 未找到
func NotFound(c *gin.Context, message string) {
	Error(c, CodeNotFound, message)
}

// InternalError 500 内部错误
func InternalError(c *gin.Context, message string) {
	Error(c, CodeInternalError, message)
}

// DatabaseError 数据库错误
func DatabaseError(c *gin.Context, message string) {
	Error(c, CodeDatabaseError, message)
}

// ValidationError 验证错误
func ValidationError(c *gin.Context, message string) {
	Error(c, CodeValidationError, message)
}
