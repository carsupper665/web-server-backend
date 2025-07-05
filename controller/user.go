// controller/user.go

package controller

import (
	"fmt"
	"go-backend/common"
	"go-backend/model"
	"math/rand"
	"net/http"
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
	var user model.User
	var err error
	clientIP := c.ClientIP()

	if err := c.ShouldBindJSON(&req); err != nil {
		common.LogError(c.Request.Context(), "Login request binding error: "+err.Error())
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	_, err = c.Cookie(common.JwtCookieName)
	if err == nil { // 已經登入了
		c.JSON(200, gin.H{"message": "Already logged in"})
		return
	}

	// 0) 檢查暫時封鎖

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
		_ = model.RecordAttempt(clientIP, false)
		c.JSON(401, gin.H{"error": "Invalid password"})
		return
	}
	UA := c.GetHeader("User-Agent")
	now := time.Now()
	msg := fmt.Sprintf(
		"User: %s, IP: %s, User-Agent: %s login at %s",
		user.Username,
		clientIP,
		UA,
		now.Format("2006/01/02-15:04:05"),
	)
	common.LogDebug(c.Request.Context(), msg)
	_ = model.ReSetFail(clientIP)
	clinetDeviceID, err := c.Cookie("device_id")
	if err != nil {
		clinetDeviceID = common.GenerateDeviceIDWithIP(clientIP)
		c.SetCookie( // 先寫進餅乾裡 下次直接取
			"device_id",    // name
			clinetDeviceID, // value
			60*60*24*365,   // maxAge (秒)：一年
			"/",            // path
			"",             // domain (留空為當前 host)
			false,          // secure (https 才送)
			true,           // httpOnly
		)
		CreateVerificationCode(c, user) // 發送驗證碼
		c.JSON(202, gin.H{"message": "verification code sent, for new device"})
		return
	}
	// 檢查是否已存在此裝置的登入紀錄
	isExists, _ := model.IsDeviceExists(clinetDeviceID)
	if !isExists { //不存在就跑email 驗證
		CreateVerificationCode(c, user) // 發送驗證碼
		c.JSON(202, gin.H{"message": "verification code sent"})
		return
	}
	// 如果存在就直接登入

	SetUpJWT(c, user) // 設置 JWT

}

func CreateVerificationCode(c *gin.Context, user model.User) {
	// 這個函式可以用來發送驗證碼到使用者的 email
	// 目前只是回傳一個訊息，實際上應該要發送 email
	common.LogDebug(c.Request.Context(), "CreateVerificationCode called for user: "+user.Username)
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	err := model.SetVerificationCode(user.ID, code)
	if err != nil {
		if err.Error() == "verification code already set and not expired" {
			return

		}
		common.LogError(c.Request.Context(), "SetVerificationCode error: "+err.Error())
		return
	}

	htmlMsg := fmt.Sprintf(
		`<!DOCTYPE html>
	<html>
	<head>
	<meta charset="UTF-8">
	<title>Verification Code</title>
	</head>
	<body style="font-family: sans-serif; line-height:1.4;">
	<p>Hello %s,</p>
	<p>Your verification code is: <strong>%s</strong></p>
	<p>
		Please use this code to verify your login. 
		This code will expire in <strong>5 minutes</strong>.
	</p>
	<p>If you did not request this, please ignore this email.</p>
	<br>
	<p>Thank you,<br>The %s Team</p>
	</body>
	</html>`,
		user.DisplayName,
		code,
		common.SystemName,
	)

	err = common.SendEmail(
		"Login Verification Code",
		user.Email, // 使用者的 email
		htmlMsg,
	)

	if err != nil {
		common.LogError(c.Request.Context(), "SendEmail error: "+err.Error()+" SMTP account: "+common.SMTPAccount)
		common.LogError(c.Request.Context(), "Failed to send verification code to user: "+user.Username)
	}

	c.SetCookie(
		"email",
		user.Email, // 使用者的 email
		60*5,
		"/",   // path
		"",    // domain (留空為當前 host)
		false, // secure (https 才送)
		true,  // httpOnly
	)
}

func VerifyLogin(c *gin.Context) {
	var req struct {
		Code string `json:"code" binding:"required"`
	}
	clientIP := c.ClientIP()
	if err := c.ShouldBindJSON(&req); err != nil {
		common.LogError(c.Request.Context(), "VerifyLogin request binding error: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid request"})
		return
	}

	_, err := c.Cookie(common.JwtCookieName)
	if err == nil { // 已經登入了
		c.JSON(200, gin.H{"message": "Already logged in"})
		return
	}

	email, err := c.Cookie("email")
	if err != nil {
		common.LogError(c.Request.Context(), "Failed to get email from cookie: "+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request."})
		return
	}

	code, sendAt, err := model.GetVerificationCode(email)
	if err != nil {
		common.LogError(c.Request.Context(), "Verify Error"+err.Error())
		_ = model.RecordAttempt(clientIP, false)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if code != req.Code || time.Since(sendAt) > 5*time.Minute {
		_ = model.RecordAttempt(clientIP, false)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired verification code" + code + " " + req.Code})
		return
	}
	_ = model.ReSetFail(clientIP)
	_ = model.ClearVerificationCode(email) // 重置驗證碼
	user, _ := model.GetUserByEmail(email)
	SetUpJWT(c, user)
}

func SetUpJWT(c *gin.Context, user model.User) {

	payload := map[string]interface{}{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"Login_IP": c.ClientIP(),
	}

	t, err := common.GenerateJWTToken(payload)

	if err != nil {
		common.LogError(c.Request.Context(), "GenerateJWTToken error: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	clinetDeviceID, _ := c.Cookie("device_id")

	ua := c.GetHeader("User-Agent")

	ip := c.ClientIP()

	model.SaveDevice(
		clinetDeviceID,
		ua,
		ip,
		user.ID)

	c.SetCookie(
		common.JwtCookieName,    // name
		t,                       // value
		common.JwtExpireSeconds, // maxAge (秒)：7天
		"/",                     // path
		"",                      // domain (留空為當前 host)
		false,                   // secure (https 才送)
		true,                    // httpOnly
	)
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful"})
}

func Logout(c *gin.Context) {
	// 清除 JWT Cookie
	c.SetCookie(
		common.JwtCookieName, // name
		"",                   // value
		-1,                   // maxAge (負值表示刪除)
		"/",                  // path
		"",                   // domain (留空為當前 host)
		false,                // secure (https 才送)
		true,                 // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}
