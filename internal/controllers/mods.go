package controllers

import (
	"net/http"

	"github.com/example/minecraft-server-controller/internal/services"
	"github.com/gin-gonic/gin"
)

func AddMod(c *gin.Context) {
	name := c.PostForm("name")
	services.AddMod(name)
	c.JSON(http.StatusOK, gin.H{"mod": name})
}

func DeleteMod(c *gin.Context) {
	name := c.Param("name")
	services.DeleteMod(name)
	c.Status(http.StatusOK)
}
