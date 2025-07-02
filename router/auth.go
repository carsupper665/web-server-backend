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
	)
	{
		auth.POST("/login", controller.Login)
	}

}
