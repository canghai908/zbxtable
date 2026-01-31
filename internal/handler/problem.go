package handler

import (
	"zbxtable/internal/model"
	"zbxtable/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetAllProblem 获取未恢复告警
func GetAllProblem(c *gin.Context) {
	b, cnt, err := model.GetProblems()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.SuccessWithPage(c, b, cnt)
}
