// middleware/auth.go

package middleware

import (
	"go-backend/common"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func GenToken() string {
	// 產生JWT並返回
	return "generated_token"
}

// 初步過濾掉一些簡易爬蟲 只允許各大宗瀏覽器(但不含edge)
func UserAgentFilter() gin.HandlerFunc {
	// 定義允許的瀏覽器關鍵字
	allowed := []string{
		"chrome",
		"firefox",
		"safari",
		"opera",
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
		c.Abort()
	}
}
