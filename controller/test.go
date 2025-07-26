// controller/test

package controller

import (
	"go-backend/common"
	"go-backend/model"

	"github.com/gin-gonic/gin"
)

func TestServer(c *gin.Context) {
	if model.TestModelServer() && common.DebugMode {
		c.JSON(200, gin.H{
			"message": "Server is running correctly",
			"status":  200,
		})
		return
	} else {
		c.JSON(403, gin.H{
			"message": "Method not allowed.",
		})
		return
	}

}

func ErrorTest(c *gin.Context) {
	panic("This is a test panic to check error handling.")
}
