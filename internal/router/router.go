package router

import (
	"github.com/gin-gonic/gin"

	"github.com/example/minecraft-server-controller/internal/controllers"
	"github.com/example/minecraft-server-controller/internal/middleware"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.POST("/login", controllers.Login)

	authorized := r.Group("/")
	authorized.Use(middleware.AuthMiddleware())
	{
		authorized.POST("/server/start", controllers.StartServer)
		authorized.POST("/server/stop", controllers.StopServer)
		authorized.POST("/server/version", controllers.SwitchVersion)
		authorized.POST("/server/backup", controllers.BackupWorld)
		authorized.GET("/logs", controllers.GetLogs)
		authorized.POST("/mods", controllers.AddMod)
		authorized.DELETE("/mods/:name", controllers.DeleteMod)
	}

	return r
}
