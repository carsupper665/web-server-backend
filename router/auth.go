// router/auth.go

package router

import (
	"go-backend/controller"
	"go-backend/middleware"

	// "go-backend/middleware"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetAuthRouter(router *gin.Engine) {
	auth := router.Group("/Authentication")
	auth.Use(
		gzip.Gzip(gzip.DefaultCompression),
		middleware.UserAgentFilter(),
		middleware.GloabalIPFilter(),
	)
	{
		auth.POST("/login", controller.Login)
		auth.POST("/verify", controller.VerifyLogin)
	}

}
