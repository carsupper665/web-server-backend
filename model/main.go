// model/main.go
package model

import (
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
		common.SysLog("no user exists, create a root user for you: username is root, password is 123456")
		hashedPassword, err := common.Password2Hash("123456")
		if err != nil {
			return err
		}
		rootUser := User{
			Username:    "root",
			Password:    hashedPassword,
			Role:        common.RoleRootUser,
			DisplayName: "Root User",
			AccessToken: nil,
		}
		DB.Create(&rootUser)
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
	)

	if err != nil {
		return err
	}

	return nil
}

func CheckRootUser() {
	createRoot := common.GetEnvOrDefaultBool("CREATE_ROOT_USER", true)
	if !createRoot {
		common.SysLog("CREATE_ROOT_USER disabled, skip root user check")
	}

	if RootUserExists() {
		common.SysLog("Root user already exists, skip creating root user")
	}

	if err := createRootAccountForTest(); err != nil {
	}

	common.SysLog("Root user created successfully")
}
