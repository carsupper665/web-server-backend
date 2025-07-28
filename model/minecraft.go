//model/minecraft.go

package model

import (
	"time"
)

type UserMinecraftServer struct {
	OnwerID    uint      `gorm:"primaryKey;not null" json:"owner_id"`
	ServerID   string    `gorm:"primaryKey;size:32;not null" json:"server_id"`
	SystemPath string    `gorm:"size:255;not null" json:"system_path"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
}

func AddServerToUser(userID uint, serverID, systemPath string) error {
	userServer := UserMinecraftServer{
		OnwerID:    userID,
		ServerID:   serverID,
		SystemPath: systemPath,
	}
	return DB.Create(&userServer).Error
}

func GetUserServers(userID uint) ([]UserMinecraftServer, error) {
	var servers []UserMinecraftServer
	err := DB.Where("owner_id = ?", userID).Find(&servers).Error
	if err != nil {
		return nil, err
	}
	return servers, nil
}

func RemoveServerByServerID(userID uint, serverID string) error {
	return DB.Where("owner_id = ? AND server_id = ?", userID, serverID).Delete(&UserMinecraftServer{}).Error
}
