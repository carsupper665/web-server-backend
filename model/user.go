// model/user.go

package model

import (
	"go-backend/common"
	"time"

	"gorm.io/gorm"
)

// 創表用，和新增用的結構
type User struct {
	ID               uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Username         string         `gorm:"size:12;not null" json:"username"`
	DisplayName      string         `json:"display_name" gorm:"index" validate:"max=20"`
	Role             int            `gorm:"default:1;not null" json:"role"`
	Email            string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	Password         string         `gorm:"size:255;not null" json:"-"`
	Salt             string         `gorm:"size:255;not null" json:"-"`
	VerificationCode string         `gorm:"-:all" json:"verification_code"`
	AccessToken      *string        `json:"access_token" gorm:"type:char(32);column:access_token;uniqueIndex"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func RootUserExists() bool {
	var user User
	err := DB.Where("role = ?", common.RoleRootUser).First(&user).Error
	if err != nil {
		return false
	}
	return true
}
