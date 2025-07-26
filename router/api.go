// router/api.go

package router

import (
	"go-backend/common"
	"go-backend/controller"
	"go-backend/middleware"

	// "go-backend/middleware"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetAPIRouter(router *gin.Engine) {
	api := router.Group("/api")
	api.Use(gzip.Gzip(gzip.DefaultCompression),
		middleware.IpRateLimiter(common.GlobalApiRateLimitNum, common.GlobalApiRateLimitDuration),
		middleware.GloabalIPFilter())
	{
		api.GET(("/test-Server"), controller.TestServer)
		api.POST("/logout", controller.Logout)
	}

}
