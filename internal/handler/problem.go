package handler

import (
	"net/http"
	"zbxtable/internal/model"

	"github.com/gin-gonic/gin"
)

// GetAllProblem 获取未恢复告警
func GetAllProblem(c *gin.Context) {
	var ProblemsRes model.ProblemsRes
	b, cnt, err := model.GetProblems()
	if err != nil {
		ProblemsRes.Code = 500
		ProblemsRes.Message = err.Error()
	} else {
		ProblemsRes.Code = 200
		ProblemsRes.Message = "获取成功"
		ProblemsRes.Data.Items = b
		ProblemsRes.Data.Total = cnt
	}
	c.JSON(http.StatusOK, ProblemsRes)
}
