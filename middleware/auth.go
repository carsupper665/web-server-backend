// middleware/auth.go

package middleware

import (
	"go-backend/common"
	"go-backend/model"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	FailLimit  = 5
	FailWindow = 5 * time.Minute
)

func GloabalIPFilter() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		isBanned, err := model.IsIPBanned(ip)

		if err != nil {
			c.Abort()
			return
		}

		if isBanned {
			common.LogDebug(c.Request.Context(), "Blocked IP: "+ip)
			c.AbortWithStatus(403)
			return
		}

		windowStart := time.Now().Add(-FailWindow)
		fails, err := model.CountRecentFails(ip, windowStart)
		if err != nil {
			common.LogError(c.Request.Context(), "CountRecentFails error: "+err.Error())
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
		if fails >= FailLimit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "Too many login attempts."})
			_ = model.BanIP(ip, "Too many login or verify attempts")
			return
		}

		c.Next()
	}
}

func DebugMode() gin.HandlerFunc {
	return func(c *gin.Context) {
		if common.DebugMode {
			common.LogWarn(c.Request.Context(), "Warn: Debug mode is enabled, allowing all test requests.")
			c.Next()
		} else {
			c.AbortWithStatus(http.StatusBadGateway)
			return
		}
	}
}

// 初步過濾掉一些簡易爬蟲 只允許各大宗瀏覽器(但不含edge)
func UserAgentFilter() gin.HandlerFunc {
	// 定義允許的瀏覽器關鍵字
	allowed := []string{
		"chrome",
		"firefox",
		"safari",
		"opera",
		"mozilla",
		"mpmc client ua",
	}

	return func(c *gin.Context) {

		if common.DebugMode && !common.UaFilter {
			// 如果是DEBUG模式且UA過濾關閉，直接放行
			common.LogDebug(c.Request.Context(), "User-Agent filter is disabled.")
			c.Next()
			return
		}

		ua := strings.ToLower(c.GetHeader("User-Agent"))
		if ua == "" {

			c.Abort()
			return
		}

		if strings.Contains(ua, "edge") || strings.Contains(ua, "edg/") {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Forbidden"})
			return
		}

		for _, a := range allowed {
			if strings.Contains(ua, a) {
				common.LogDebug(c.Request.Context(), "User-Agent allowed: "+ua)
				c.Next()
				return
			}
		}

		// 其他一律拒絕
		c.AbortWithStatus(http.StatusForbidden)
	}
}

func ValidateJWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(common.JwtCookieName)
		if err != nil || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		Valid := common.ValidateUser(token, c.ClientIP())
		if !Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		c.Next()
	}
}
