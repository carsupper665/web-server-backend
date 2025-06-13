package controllers

import (
	"net/http"

	"github.com/example/minecraft-server-controller/internal/services"
	"github.com/gin-gonic/gin"
)

func StartServer(c *gin.Context) {
	services.StartServer()
	c.JSON(http.StatusOK, gin.H{"status": "starting"})
}

func StopServer(c *gin.Context) {
	services.StopServer()
	c.JSON(http.StatusOK, gin.H{"status": "stopping"})
}

func SwitchVersion(c *gin.Context) {
	version := c.PostForm("version")
	services.SwitchVersion(version)
	c.JSON(http.StatusOK, gin.H{"version": version})
}

func BackupWorld(c *gin.Context) {
	services.BackupWorld()
	c.JSON(http.StatusOK, gin.H{"status": "backup started"})
}
