//controller/minecraft.go

package controller

import (
	"fmt"
	"go-backend/common"
	"go-backend/model"
	"go-backend/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

func CreateServer(c *gin.Context) {
	var req service.CreateServerRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		common.LogError(c.Request.Context(), "CreateMinecraftServer request binding error: "+err.Error())
		c.JSON(400, gin.H{"error": "Invalid request"})
		return
	}

	token, err := c.Cookie(common.JwtCookieName)
	if err != nil {
		c.JSON(403, gin.H{"error": "Unauthorized"})
		return
	}

	payload, err := common.GetJWTPayload(token)
	if err != nil {
		common.LogDebug(c.Request.Context(), "JWT payload error: "+err.Error())
		c.JSON(403, gin.H{"error": "Invalid token"})
		return
	}

	rawUID, _ := payload["user_id"]

	serverID, err := service.CreateServer(rawUID.(string), req.ServerType, req.ServerVer, req.FabricLoader, req.FabricInstaller)
	if err != nil {
		common.LogError(c.Request.Context(), "CreateMinecraftServer error: "+err.Error())
		c.JSON(500, gin.H{"error": "Failed to create server"})
		return
	}
	uid, parseErr := strconv.ParseUint(rawUID.(string), 10, 32)
	if parseErr != nil {
		service.ErrorFileClear(common.MinecraftServerPath + "/" + serverID)
		common.LogError(c.Request.Context(), "Parsing user id error: "+parseErr.Error())
		c.JSON(500, gin.H{"error": "Internal Server Error"})
		common.SendErrorToDc(fmt.Sprintf(
			"CreateServer failed: cannot parse user_id %q: %v",
			rawUID, parseErr,
		))
		return
	}

	model.AddServerToUser(uint(uid), serverID, common.MinecraftServerPath+"/"+serverID)

	c.JSON(200, gin.H{"server_id": serverID})
}

func GetAllVanillaVersions(c *gin.Context) {
	versions, err := service.GetAllVanillaVersions()
	if len(versions) == 0 || err != nil {
		c.JSON(404, gin.H{"error": ""})
		return
	}
	c.JSON(200, gin.H{"versions": versions})
}

func GetAllFabricVersions(c *gin.Context) {
	versions, err := service.GetAllFabricVersions()
	if len(versions) == 0 || err != nil {
		common.LogError(c.Request.Context(), "GetAllFabricVersions error: "+err.Error())
		c.JSON(404, gin.H{"error": ""})
		return
	}
	c.JSON(200, gin.H{"versions": versions})
}
