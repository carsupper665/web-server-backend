// controller/user.go

package controller

import (
	"fmt"
	"go-backend/common"
	"go-backend/model"
	"go-backend/service"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type loginRequest struct {
	Email    string `json:"email" binding:"omitempty,email"`
	Username string `json:"username" binding:"omitempty"`
	Password string `json:"password" binding:"required"`
}

func clearCookies(c *gin.Context) {
	c.SetCookie(common.JwtCookieName, "", -1, "/", "", false, true)
	c.SetCookie("email", "", -1, "/", "", false, true)
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
	clientDeviceID, err := c.Cookie("device_id")
	if err != nil {
		clientDeviceID = common.GenerateDeviceIDWithIP(clientIP)
		c.SetCookie( // 先寫進餅乾裡 下次直接取
			"device_id",    // name
			clientDeviceID, // value
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
	isExists, dberr := model.IsDeviceExists(clientDeviceID)
	if !isExists || dberr != nil { //不存在就跑email 驗證
		CreateVerificationCode(c, user) // 發送驗證碼
		c.JSON(202, gin.H{"message": "verification code sent"})
		return
	}
	// 如果存在就直接登入

	SetUpJWT(c, user) // 設置 JWT

}

func CreateVerificationCode(c *gin.Context, user model.User) {
	common.LogDebug(c.Request.Context(), "CreateVerificationCode called for user: "+user.Username)
	c.SetCookie(
		"email",
		user.Email, // 使用者的 email
		60*5,
		"/",   // path
		"",    // domain (留空為當前 host)
		false, // secure (https 才送)
		true,  // httpOnly
	)
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
	<style>
		body {
		margin: 0;
		padding: 0;
		font-family: sans-serif;
		line-height: 1.4;
		background: linear-gradient(-45deg,
			#ff0000, #ff7f00, #ffff00, #00ff00, #0000ff, #4b0082, #8f00ff);
		background-size: 400%% 400%%;
		animation: rainbow 15s ease infinite;
		}
		@keyframes rainbow {
		0%%   { background-position: 0%% 50%%; }
		50%%  { background-position: 100%% 50%%; }
		100%% { background-position: 0%% 50%%; }
		}
		.container {
		padding: 20px;
		background: rgba(255, 255, 255, 0.8);
		margin: 40px auto;
		max-width: 600px;
		border-radius: 8px;
		}
		p { margin: 1em 0; }
	</style>
	</head>
	<body>
	<div class="container">
		<p>Hello %s,</p>
		<p>Your verification code is: <strong>%s</strong></p>
		<p>
		Please use this code to verify your login. 
		This code will expire in <strong>5 minutes</strong>.
		</p>
		<p>If you did not request this, please ignore this email.</p>
		<br>
		<p>Thank you,<br>The %s Team</p>
	</div>
	</body>
	</html>`,
		user.DisplayName,
		code,
		common.SystemName,
	)

	email := user.Email

	if email == "" || email == "null" {
		common.LogError(c.Request.Context(), "User email is empty for user: "+user.Username)
	}

	common.LogDebug(c.Request.Context(), "Sending verification code to user: "+user.Username+" Email: "+email)

	err = common.SendEmail(
		"Login Verification Code",
		user.Email, // 使用者的 email
		htmlMsg,
	)

	if err != nil {
		common.LogError(c.Request.Context(), "SendEmail error: "+err.Error()+" SMTP account: "+common.SMTPAccount)
		common.LogError(c.Request.Context(), "Failed to send verification code to user: "+user.Username)
	}
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

	email, cookieErr := c.Cookie("email")
	if cookieErr != nil {
		common.LogError(c.Request.Context(), "Failed to get email from cookie: "+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request."})
		return
	}

	code, sendAt, ver_err := model.GetVerificationCode(email)
	if ver_err != nil {
		common.LogError(c.Request.Context(), "Verify Error"+err.Error())
		_ = model.RecordAttempt(clientIP, false)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	if code != req.Code || time.Since(sendAt) > 5*time.Minute {
		_ = model.RecordAttempt(clientIP, false)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired verification code"})
		return
	}
	_ = model.ReSetFail(clientIP)
	_ = model.ClearVerificationCode(email) // 重置驗證碼
	user, _ := model.GetUserByEmail(email)

	appHeadertDeviceID := c.GetHeader(common.ClientHeader)
	ua := strings.ToLower(c.GetHeader("User-Agent"))
	if strings.Contains(ua, "mpmc client ua") && appHeadertDeviceID != "" {
		SetUpAppJWT(c, user)
		return
	}

	SetUpJWT(c, user)
}

func SetUpJWT(c *gin.Context, user model.User) {

	exp := time.Now().Add(common.JwtExpireSeconds * time.Second).Unix()

	payload := map[string]interface{}{
		"user_id":  fmt.Sprint(user.ID),
		"username": user.Username,
		"role":     user.Role,
		"Login_IP": c.ClientIP(),
		"exp":      exp,
	}

	t, err := common.GenerateJWTToken(payload)

	if err != nil {
		common.LogError(c.Request.Context(), "GenerateJWTToken error: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	clientDeviceID, _ := c.Cookie("device_id")

	ua := c.GetHeader("User-Agent")

	ip := c.ClientIP()

	model.SaveDevice(
		clientDeviceID,
		ua,
		ip,
		user.ID)

	c.SetCookie(
		common.JwtCookieName,    // name
		t,                       // value
		common.JwtExpireSeconds, // maxAge
		"/",                     // path
		"",                      // domain (留空為當前 host)
		false,                   // secure (https 才送)
		true,                    // httpOnly
	)

	c.SetCookie("email", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful"})
}

func Logout(c *gin.Context) {
	// 清除 JWT Cookie
	clearCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "Logout successful"})
}

type AmongUs struct {
	agm *service.GameManager
}

func NewAmongUsController(agm *service.GameManager) *AmongUs {
	return &AmongUs{agm: agm}
}

func (a *AmongUs) Join(c *gin.Context) {
	gameId := c.Param("id")
	if gameId == "" {
		c.JSON(400, gin.H{"error": "傻逼"})
		return
	}

	clientIP := c.ClientIP()

	clientDeviceID, err := c.Cookie("device_id")
	if err != nil {
		clientDeviceID = common.GenerateDeviceIDWithIP(clientIP)
		c.SetCookie( // 先寫進餅乾裡 下次直接取
			"device_id",    // name
			clientDeviceID, // value
			60*60*24*365,   // maxAge (秒)：一年
			"/",            // path
			"",             // domain (留空為當前 host)
			false,          // secure (https 才送)
			true,           // httpOnly
		)
		err = nil
	}

	role, task, taskInfo, rt, err := a.agm.Join(clientDeviceID, gameId)
	// role, task, taskInfo, rt, err := a.agm.Join(common.GetRandomString(8), gameId)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": role, "Task": task, "TaskInfo": taskInfo, "RoundTasks": rt})
}

// admin method
func (a *AmongUs) AllGames(c *gin.Context) {
	//admin method

	gs := a.agm.List()

	c.JSON(200, gin.H{
		"message": "all games",
		"games":   gs,
	})

}

func (a *AmongUs) Create(c *gin.Context) {
	//admin method

	num := c.Param("num")

	if num == "" {
		num = "5"
	}

	gs, err := a.agm.Create(num)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Create sussues", "game_id": gs.ID()})
}

func (a *AmongUs) EndGame(c *gin.Context) {
	//admin method

	gameId := c.Param("id")
	if gameId == "" {
		c.JSON(400, gin.H{"error": "傻逼"})
		return
	}

	err := a.agm.EndGame(gameId)

	if err != nil {
		c.JSON(400, gin.H{"error": "傻逼"})
		return
	}

	c.JSON(200, gin.H{"message": "del"})
}

func (a *AmongUs) ListPlayers(c *gin.Context) {
	gameId := c.Param("id")
	if gameId == "" {
		c.JSON(400, gin.H{"error": "傻逼"})
		return
	}

	players, err := a.agm.ListPlayers(gameId)

	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": players})

}
