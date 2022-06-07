package heartbeat

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Heartbeat(c *gin.Context) {
	c.AbortWithStatus(http.StatusOK)
}
