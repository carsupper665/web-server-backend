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
	Username         string         `gorm:"size:12;not null;uniqueIndex" json:"username"`
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

func LoginByName(username string) (User, error) {

	var user User
	err := DB.
		Where("username = ?", username).
		First(&user).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return User{}, err
		}
		return User{}, err
	}
	return user, nil
}

// LoginByEmail retrieves a user by email.
// It only does the DB lookup; password verification should be done elsewhere.
func LoginByEmail(email string) (User, error) {

	var user User
	err := DB.
		Where("email = ?", email).
		First(&user).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return User{}, err
		}
		return User{}, err
	}
	return user, nil
}
