// controller/clientApp.go

package controller

import (
	"fmt"
	"go-backend/common"
	"go-backend/model"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AppLogin(c *gin.Context) {
	var req loginRequest
	var user model.User
	var err error
	clientIP := c.ClientIP()

	if err := c.ShouldBindJSON(&req); err != nil {
		common.LogError(c.Request.Context(), "Login request binding error: "+err.Error())
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

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
			_ = model.RecordAttempt(clientIP, false)
		} else {
			c.JSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}

	v := common.ValidatePasswordAndHash(req.Password+user.Salt, user.Password)

	_ = model.RecordAttempt(clientIP, false)

	if !v {
		c.JSON(401, gin.H{"error": "Invalid password"})
		return
	}

	UA := c.GetHeader("User-Agent")
	if UA == "" || !strings.Contains(UA, "mpmc client ua") {
		c.JSON(502, gin.H{"error": "bad req"})
		return
	}
	parts := strings.Split(UA, "-") // userId-ver-ua
	if len(parts) != 3 {
		c.AbortWithStatusJSON(502, gin.H{"error": "format error"})
		return
	}
	_, ver := parts[0], parts[1]
	if ver != common.LatestClientVersion {
		c.AbortWithStatusJSON(502, gin.H{"error": "outdated"})
		return
	}

	now := time.Now()
	msg := fmt.Sprintf(
		"User: %s, IP: %s, User-Agent: %s login at %s",
		user.Username,
		clientIP,
		UA,
		now.Format("2006/01/02-15:04:05"),
	)
	common.LogDebug(c.Request.Context(), msg)

	appHeadertDeviceID := c.GetHeader(common.ClientHeader)

	if appHeadertDeviceID == "" {
		c.JSON(401, gin.H{"error": ""})
		return
	}

	_ = model.ReSetFail(clientIP)
	if appHeadertDeviceID == "mpmc HUNS" {
		appHeadertDeviceID = common.GenerateDeviceIDWithIP(clientIP)
		CreateVerificationCode(c, user) // 發送驗證碼
		c.Writer.Header().Set(common.ClientHeader, appHeadertDeviceID)
		c.JSON(202, gin.H{"message": "verification code sent, for new device"})
		return
	}
	isExists, dberr := model.IsDeviceExists(appHeadertDeviceID)
	if !isExists || dberr != nil { //不存在就跑email 驗證
		CreateVerificationCode(c, user) // 發送驗證碼
		c.JSON(202, gin.H{"message": "verification code sent"})
		return
	}
	SetUpAppJWT(c, user)
}

func SetUpAppJWT(c *gin.Context, user model.User) {

	exp := time.Now().Add(common.JwtExpireSeconds * time.Second).Unix()

	appHeadertDeviceID := c.GetHeader(common.ClientHeader)
	ua := strings.ToLower(c.GetHeader("User-Agent"))
	if !strings.Contains(ua, "mpmc client ua") && appHeadertDeviceID == "" {
		c.JSON(401, gin.H{"error": "Invalid id and ua"})
		return
	}

	tid, err := common.Password2Hash(appHeadertDeviceID + fmt.Sprint(user.ID) + c.ClientIP() + ua)

	if err != nil {
		c.JSON(502, gin.H{"error": "Server Error."})
		return
	}

	payload := map[string]interface{}{
		"user_id":  fmt.Sprint(user.ID),
		"username": user.Username,
		"role":     user.Role,
		"tid":      tid,
		"exp":      exp,
	}

	t, err := common.GenerateJWTToken(payload)

	if err != nil {
		common.LogError(c.Request.Context(), "GenerateJWTToken error: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	ip := c.ClientIP()

	model.SaveDevice(
		appHeadertDeviceID,
		ua,
		ip,
		user.ID)

	c.SetCookie("email", "", -1, "/", "", false, true)
	c.JSON(200, gin.H{
		"message": "Login successful", "token": t})
}
