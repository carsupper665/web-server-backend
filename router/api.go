// router/api.go

package router

import (
	"go-backend/common"
	"go-backend/controller"
	"go-backend/middleware"
	"go-backend/service"

	// "go-backend/middleware"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

func SetAPIRouter(router *gin.Engine) {
	pl := common.GetPortList(30000, 30050)

	mgr := service.NewServerManager(pl)
	svc := service.NewServerService(mgr)
	c := controller.NewServerController(svc)

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
	amcapi := mcapi.Group("/a")
	amcapi.Use(middleware.ValidateJWT())
	{
		amcapi.POST("/status/:server_id", c.GetStatus)
	}

	testApi := router.Group("/test-api")
	testApi.Use(gzip.Gzip(gzip.DefaultCompression),
		middleware.DebugMode(),
	)
	{
		testApi.POST("/mc-server/create", controller.CreateServer)
		testApi.POST("/status/:server_id", c.GetStatus)
		testApi.POST("/startmyserver/:server_id", c.Start)
		testApi.POST("/stopmyserver/:server_id", c.Stop)
	}

}
