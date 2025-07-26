// model/book.go

package model

import (
	"errors"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Book 對應 books 資料表，一張表儲存所有欄位
// 已在資料庫對 media_id 設定 UNIQUE INDEX
type Book struct {
	ID            int64          `gorm:"primaryKey" json:"id"`
	MediaID       int64          `gorm:"uniqueIndex;not null" json:"media_id"`
	EnglishTitle  string         `gorm:"column:english_title" json:"english_title"`
	JapaneseTitle string         `gorm:"column:japanese_title" json:"japanese_title"`
	PrettyTitle   string         `gorm:"column:pretty_title" json:"pretty_title"`
	Scanlator     string         `json:"scanlator"`
	UploadDate    time.Time      `gorm:"not null" json:"upload_date"`
	NumPages      int            `gorm:"not null" json:"num_pages"`
	NumFavorites  int            `gorm:"not null" json:"num_favorites"`
	Pages         datatypes.JSON `gorm:"type:jsonb" json:"pages"`
	Cover         datatypes.JSON `gorm:"type:jsonb" json:"cover"`
	Thumbnail     datatypes.JSON `gorm:"type:jsonb" json:"thumbnail"`
	Tags          datatypes.JSON `gorm:"type:jsonb" json:"tags"`
	CreatedAt     time.Time      `json:"-"`
	UpdatedAt     time.Time      `json:"-"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

// 這邊的檔案都還沒測試 先暫時擱置 等server開發完 再回頭做這個
// CreateBook 建立新書目，若 media_id 重複則跳過
func CreateBook(db *gorm.DB, book *Book) error {
	var existing Book
	err := db.Unscoped().Where("media_id = ?", book.MediaID).First(&existing).Error
	if err == nil {
		// 已存在，不重複建立
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	return db.Create(book).Error
}

// GetBookByID 依 ID 查詢
func GetBookByID(db *gorm.DB, id int64) (*Book, error) {
	var book Book
	if err := db.First(&book, id).Error; err != nil {
		return nil, err
	}
	return &book, nil
}

// GetAllBooks 查詢所有書目，可加分頁
func GetAllBooks(db *gorm.DB, limit, offset int) ([]Book, error) {
	var books []Book
	if err := db.Limit(limit).Offset(offset).Find(&books).Error; err != nil {
		return nil, err
	}
	return books, nil
}

// UpdateBook 更新書目資料
// updates 可為 map[string]interface{} 或 struct
func UpdateBook(db *gorm.DB, id int64, updates interface{}) error {
	return db.Model(&Book{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteBook 刪除書目（軟刪除）
func DeleteBook(db *gorm.DB, id int64) error {
	return db.Delete(&Book{}, id).Error
}
