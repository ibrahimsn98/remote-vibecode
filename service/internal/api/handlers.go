package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handlers struct{}

func New() *Handlers {
	return &Handlers{}
}

func (h *Handlers) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
