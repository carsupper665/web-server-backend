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

	_, uid_str, uid_uint, err := getPayloadAndId(c)

	serverID, err := service.CreateServer(uid_str, req.ServerType, req.ServerVer, req.FabricLoader, req.FabricInstaller)
	if err != nil {
		common.LogError(c.Request.Context(), "CreateMinecraftServer error: "+err.Error())
		c.JSON(500, gin.H{"error": "Failed to create server"})
		return
	}

	modelErr := model.AddServerToUser(uid_uint, serverID, req.DisplayName, common.MinecraftServerPath+"/"+serverID)
	if modelErr != nil {
		common.LogError(c.Request.Context(), "AddServerToUser error: "+err.Error())
		service.ErrorFileClear(common.MinecraftServerPath + "/" + serverID)
		c.JSON(500, gin.H{"error": "Failed to add server to user"})
		return
	}

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

func MyServers(c *gin.Context) {
	_, _, uid, err := getPayloadAndId(c)

	servers, err := model.GetUserServers(uid)
	if err != nil {
		common.LogError(c.Request.Context(), "GetUserServers error: "+err.Error())
		c.JSON(500, gin.H{"error": "Failed to retrieve servers"})
		return
	}

	c.JSON(200, servers)
}

func DeleteServerById(c *gin.Context) {
	serverID := c.Param("server_id")
	if serverID == "" {
		c.JSON(400, gin.H{"error": "Server ID is required"})
		return
	}

	_, _, id_uint, err := getPayloadAndId(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "Internal Server Error: " + err.Error()})
		return
	}

	err = model.RemoveServerByServerID(id_uint, serverID)
	if err != nil {
		common.LogError(c.Request.Context(), "RemoveServerByServerID error: "+err.Error())
		c.JSON(500, gin.H{"error": "Failed to delete server"})
		return
	}

	c.JSON(200, gin.H{"message": "Server deleted successfully"})
}

func getPayloadAndId(c *gin.Context) (map[string]interface{}, string, uint, error) {
	token, err := c.Cookie(common.JwtCookieName)
	if err != nil {
		return nil, "", 0, fmt.Errorf("failed to get JWT cookie: %w", err)
	}
	payload, err := common.GetJWTPayload(token)
	if err != nil {
		common.LogDebug(c.Request.Context(), "JWT error: "+err.Error())
		clearCookies(c)
		return nil, "", 0, fmt.Errorf("invalid token: %w", err)
	}

	rawUID, _ := payload["user_id"]
	uid, parseErr := strconv.ParseUint(rawUID.(string), 10, 32)

	if parseErr != nil {
		common.LogError(c.Request.Context(), "Parsing user id error: "+parseErr.Error())
		return nil, "", 0, fmt.Errorf("failed to parse user ID: %w", parseErr)
	}
	return payload, rawUID.(string), uint(uid), nil
}
