// model/loginAttempt.go

package model

import (
	"go-backend/common"
	"time"

	"gorm.io/gorm"
)

// LoginAttempt 用來記錄每一次登入嘗試
type LoginAttempt struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	IP        string    `gorm:"size:45;index;not null"` // IPv4/6
	Success   bool      `gorm:"not null"`               // true=成功, false=失敗
	CreatedAt time.Time `gorm:"autoCreateTime"`         // 嘗試時間
}

type BlockedIP struct {
	IP        string         `gorm:"primaryKey;size:45;not null" json:"ip"`
	Reason    string         `gorm:"size:255;not null" json:"reason"`
	BannedAt  time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// CountRecentFails 統計 ip 在 since 之後失敗的次數
func CountRecentFails(ip string, since time.Time) (int64, error) {
	var cnt int64
	err := DB.Model(&LoginAttempt{}).
		Where("ip = ? AND success = ? AND created_at >= ?", ip, false, since).
		Count(&cnt).Error
	return cnt, err
}

// RecordAttempt 記錄一次嘗試
func RecordAttempt(ip string, success bool) error {
	return DB.Create(&LoginAttempt{
		IP:      ip,
		Success: success,
	}).Error
}

func ReSetFail(ip string) error {
	return DB.Where("ip = ?", ip).Delete(&LoginAttempt{}).Error
}

// IsIPBanned 回傳這個 IP 是否存在於 BlockedIP 表中
func IsIPBanned(ip string) (bool, error) {
	var count int64
	// 正確取得 .Error
	if err := DB.
		Model(&BlockedIP{}).
		Where("ip = ?", ip).
		Count(&count).
		Error; err != nil {
		// log 也可以放在 controller 層，model 層只傳錯
		common.SysError("Failed to check banned IP: " + err.Error())
		return false, err
	}
	return count > 0, nil
}

// BanIP 封鎖指定 IP，並記錄原因
func BanIP(ip, reason string) error {
	blocked := BlockedIP{
		IP:       ip,
		Reason:   reason,
		BannedAt: time.Now(),
	}
	if err := DB.Create(&blocked).Error; err != nil {
		common.SysError("Failed to ban IP: " + err.Error())
		return err
	}
	return nil
}
