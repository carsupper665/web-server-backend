// router/auth.go

package router

import (
	"go-backend/controller"
	"go-backend/middleware"

	// "go-backend/middleware"
	"go-backend/common"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetAuthRouter(router *gin.Engine) {
	router.Use(middleware.CORS())

	auth := router.Group("/Authentication")
	auth.Use(
		gzip.Gzip(gzip.DefaultCompression),
		middleware.IpRateLimiter(common.GlobalApiRateLimitNum, common.GlobalApiRateLimitDuration),
		middleware.UserAgentFilter(),
		middleware.GloabalIPFilter(),
	)
	{
		auth.POST("/login", controller.Login)
		auth.POST("/verify", controller.VerifyLogin)
		auth.POST("/app/verify", controller.VerifyLogin)
		auth.POST("/app/login", controller.AppLogin)
	}
}
