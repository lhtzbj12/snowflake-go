package actuator

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Health 健康检查
func Health(ctx *gin.Context) {
	ctx.String(http.StatusOK, "Hi Golang, I Feel Great!!!")
}
