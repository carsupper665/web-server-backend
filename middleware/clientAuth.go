// middleware/clientAuth.go

package middleware

import (
	"go-backend/common"
	"go-backend/model"
	"strings"

	"github.com/gin-gonic/gin"
)

func ClientAppAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		appHeader := c.GetHeader(common.ClientHeader)
		if appHeader == "" {
			c.AbortWithStatusJSON(403, gin.H{"error": "unknow error"})
			return
		}

		exist, err := model.IsDeviceExists(appHeader)

		if err != nil {
			c.AbortWithStatusJSON(500, gin.H{"error": "internal error"})
			return
		}

		if !exist {
			c.AbortWithStatusJSON(403, gin.H{"error": "device not recognized"})
			common.LogWarn(c.Request.Context(), "Unrecognized device id: "+appHeader)
			return
		}

		ua := strings.ToLower(c.GetHeader("User-Agent"))
		if ua == "" || !strings.Contains(ua, "mpmc client ua") {
			c.AbortWithStatus(401)
			return
		}

		parts := strings.Split(ua, "-") // userId-ver-ua
		if len(parts) != 3 {
			c.AbortWithStatusJSON(502, gin.H{"error": "format error"})
			return
		}
		uid, ver := parts[0], parts[1]
		if ver != common.LatestClientVersion {
			c.AbortWithStatusJSON(502, gin.H{"error": "outdated"})
			return
		}

		tokenHeader := c.GetHeader("Authorization")
		if tokenHeader == "" || !strings.HasPrefix(tokenHeader, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid"})
			return
		}
		token := strings.TrimPrefix(tokenHeader, "Bearer ")

		payload, err := common.GetJWTPayload(token)
		if err != nil {
			c.AbortWithStatusJSON(403, gin.H{"error": "invalid token"})
			return
		}

		ip := c.ClientIP()
		// p := appHeader + uid + ip + ua
		p := ua + ip + uid + appHeader
		h, _ := payload["tid"].(string)
		v := common.ValidatePasswordAndHash(p, h)

		common.SysDebug("tid:" + p)

		if !v {
			c.AbortWithStatus(401)
			return
		}

		c.Set("ip", ip)
		c.Set("user", uid)

		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid, exists := c.Get("user")

		if !exists {
			c.AbortWithStatus(401)
			return
		}

		role, err := model.GetRole(uid.(uint))
		if err != nil || role != 6 {
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}

		c.Next()
	}
}
