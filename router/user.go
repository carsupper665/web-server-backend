// router/user.go
package router

import (
	"go-backend/common"
	"go-backend/controller"
	"go-backend/middleware"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetUserRouter(router *gin.Engine) {
	router.Use(middleware.CORS())
	router.POST("/logout", controller.Logout)

	user := router.Group("/user")
	user.Use(
		gzip.Gzip(gzip.DefaultCompression),
		middleware.GloabalIPFilter(),
		middleware.UserAgentFilter(),
		middleware.IpRateLimiter(common.GlobalApiRateLimitNum, common.GlobalApiRateLimitDuration),
		middleware.ValidateJWT(),
	)
	{
		user.POST("/cs", controller.CreateServer)
		user.GET("/myservers", controller.MyServers)
	}

	admin := router.Group("/op")
	admin.Use(
		gzip.Gzip(gzip.DefaultCompression),
		middleware.GloabalIPFilter(),
		middleware.UserAgentFilter(),
		middleware.IpRateLimiter(common.GlobalApiRateLimitNum, common.GlobalApiRateLimitDuration),
		middleware.ClientAppAuth(),
		middleware.AdminOnly())

	{
		admin.GET("/add", controller.AddUser)
	}

}
