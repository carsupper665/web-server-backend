package controllers

import (
	"net/http"

	"github.com/example/minecraft-server-controller/internal/services"
	"github.com/gin-gonic/gin"
)

func GetLogs(c *gin.Context) {
	logs := services.ReadLogs()
	c.String(http.StatusOK, logs)
}
