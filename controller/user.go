// controller/user.go

package controller

import (
	"fmt"
	"go-backend/common"
	"go-backend/model"
	"time"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Username string `json:"username" binding:"omitempty"`
	Password string `json:"password" binding:"required"`
}

func Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.LogError(c.Request.Context(), "Login request binding error: "+err.Error())
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	var user model.User
	var err error
	switch {
	case req.Email != "":
		user, err = model.LoginByEmail(req.Email)
	case req.Username != "":
		user, err = model.LoginByName(req.Username)
	default:
		c.JSON(400, gin.H{"error": "Email or Username is required"})
		return
	}

	if err != nil {
		if err.Error() == "record not found" {
			c.JSON(401, gin.H{"error": "User not found"})
		} else {
			c.JSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}

	v := common.ValidatePasswordAndHash(req.Password+user.Salt, user.Password)

	if !v {
		c.JSON(401, gin.H{"error": "Invalid password"})
		return
	}
	clientIP := c.ClientIP()
	device := c.GetHeader("User-Agent")
	now := time.Now()
	msg := fmt.Sprintf(
		"User: %s, IP: %s, Device: %s login at %s",
		user.Username,
		clientIP,
		device,
		now.Format("2006/01/02-15:04:05"),
	)
	common.LogDebug(c.Request.Context(), msg)
	c.JSON(200, gin.H{
		"message": "Login successful"})

	return

}
