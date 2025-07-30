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
	mcapi := router.Group("/mc-api")
	mcapi.Use(gzip.Gzip(gzip.DefaultCompression),
		middleware.IpRateLimiter(common.GlobalApiRateLimitNum, common.GlobalApiRateLimitDuration),
		middleware.GloabalIPFilter(),
		middleware.UserAgentFilter(),
	)
	{
		mcapi.GET("/finfo", controller.GetAllFabricVersions)
		mcapi.GET("/vinfo", controller.GetAllVanillaVersions)
	}

	testApi := router.Group("/test-api")
	testApi.Use(gzip.Gzip(gzip.DefaultCompression),
		middleware.DebugMode(),
	)
	{
		testApi.POST("/mc-server/create", controller.CreateServer)
	}

}
