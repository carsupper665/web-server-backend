// main.go
// 啟動流程
package main

import (
	// "embed"
	"fmt"
	"go-backend/common"
	"go-backend/middleware"
	"go-backend/model"
	"go-backend/router"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// //go:embed web/dist
// var buildFS embed.FS

// //go:embed web/dist/index.html
// var indexPage []byte

func main() {
	// .env config load
	// Go 沒有例外（exception）機制，錯誤都是以 error 型別回傳
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println(err)
	}
	common.LoadEnv()
	// Setup logger
	common.SetupLogger()
	common.SysLog("Backend Server Engine | " + common.Version + "-" + common.Bulid + " started")
	if os.Getenv("DEBUG") != "true" { // gin 預設為 debug 所以要記得關
		common.SysLog(common.ColorGreen + "Running in Release Mode" + common.ColorReset)
		gin.SetMode(gin.ReleaseMode)
	} else {
		common.SysLog(common.ColorBrightCyan + "Debug mode is enabled, running in Debug Mode" + common.ColorReset)
	}
	// init DB (use SQLite)
	err = model.InitDB()
	if err != nil {
		// fatal log 會自己關程序
		common.FatalLog("failed to init DB: " + err.Error())
	}
	// check root user exists
	err = model.CheckRootUser()
	if err != nil {
		common.SysError("failed to create root user: " + err.Error())
	}
	// init HTTP server
	server := gin.New()
	// CustomRecovery 這邊的作用是超大 exception 機制 如果API哪裡繃了可以防程序崩 再以json回傳問題
	server.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		common.SysError(fmt.Sprintf("panic detected: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("Unknow Error: %v", err),
				"type":    "unknow_panic",
				"status":  500,
			},
		})
	}))

	server.Use(middleware.RequestId())
	middleware.SetUpLogger(server)

	// init middleware

	// init session store
	store := cookie.NewStore([]byte(common.SessionSecret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   2592000, // 30 days
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
	})
	server.Use(sessions.Sessions("session", store))
	// set router
	router.SetRouter(server)
	// get port and start server
	var port = os.Getenv("PORT")
	if port == "" {
		port = strconv.Itoa(*common.Port)
	}

	err = server.Run(":" + port)
	if err != nil {
		common.FatalLog("failed to start HTTP server: " + err.Error())
	}
}
