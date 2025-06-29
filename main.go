// main.go
// 啟動流程
package main

import (
	// "embed"
	"fmt"
	"go-backend/common"
	"go-backend/middleware"
	"go-backend/router"
	"net/http"
	"os"
	"strconv"

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
	// init DB (use SQLite)

	// init HTTP server
	server := gin.New()
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
