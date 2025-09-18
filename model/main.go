// model/main.go
package model

import (
	"errors"
	"go-backend/common"

	// "os"
	// "strings"
	// "sync"
	"time"

	// "github.com/glebarez/sqlite"
	// "gorm.io/driver/mysql"
	// "gorm.io/driver/postgres"
	// "gorm.io/gorm"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

var LOG_DB *gorm.DB

func createRootAccountForTest() error {
	var user User
	//if user.Status != common.UserStatusEnabled {
	if err := DB.First(&user).Error; err != nil {
		userEmail := common.GetEnvOrDefaultString("ROOT_USER_EMAIL", "")

		if userEmail == "" {
			return errors.New("ROOT_USER_EMAIL is not set, please set it in .env file")
		}
		password := common.GetEnvOrDefaultString("ROOT_USER_PASSWORD", "123456")
		username := common.GetEnvOrDefaultString("ROOT_USER_NAME", "root")
		salt := common.GetRandomString(16)
		hashedPassword, err := common.Password2Hash(password + salt)
		if err != nil {
			return err
		}
		rootUser := User{
			Username:    username,
			Password:    hashedPassword,
			Email:       userEmail,
			Role:        common.RoleRootUser,
			Salt:        salt,
			DisplayName: "Root User",
			AccessToken: nil,
		}
		DB.Create(&rootUser)
		common.SysLog("no user exists, create a root user for you: username is " + username + ", password is " + password + ", email is:" + userEmail)
	}
	return nil
}

func InitSqliteDB(isLog bool) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(common.SQLitePath), &gorm.Config{
		PrepareStmt: true, // precompile SQL
	})

}

func InitDB() error {
	db, err := InitSqliteDB(false)
	if err != nil {
		return err
	}

	DB = db
	sqlDB, err := DB.DB()
	if err != nil {
		return err
	}
	sqlDB.SetMaxIdleConns(common.GetEnvOrDefault("SQL_MAX_IDLE_CONNS", 100))
	sqlDB.SetMaxOpenConns(common.GetEnvOrDefault("SQL_MAX_OPEN_CONNS", 1000))
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(common.GetEnvOrDefault("SQL_MAX_LIFETIME", 60)))
	common.SysLog("database migration started")
	err = migrateDB()
	return err
}

func migrateDB() error {
	err := DB.AutoMigrate(
		&User{},
		&UserDevice{},
		&UserMinecraftServer{},
		&BlockedIP{},
		&LoginAttempt{},
		&Book{},
	)

	if err != nil {
		return err
	}

	return nil
}

// 之後搞一個可以第一次啟動跳註冊的東東，現在先自動創建
// 改成有 error return
func CheckRootUser() error {
	createRoot := common.GetEnvOrDefaultBool("CREATE_ROOT_USER", false)
	if !createRoot {
		common.SysLog("CREATE_ROOT_USER disabled, skip root user check")
		return nil
	}

	if RootUserExists() {
		common.SysLog("Root user already exists, skip creating root user")
		return nil
	}

	if err := createRootAccountForTest(); err != nil {
		return err
	}

	common.SysLog("Root user created successfully")
	return nil
}
