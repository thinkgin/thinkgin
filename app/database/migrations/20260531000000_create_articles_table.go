package migrations

import (
	"thinkgin/app/database"

	"gorm.io/gorm"
)

// 创建 articles 表（article 业务模块范例）。
func init() {
	Register(database.Migration{
		Version: "20260531000000",
		Up: func(db *gorm.DB) error {
			type Article struct {
				ID        uint   `gorm:"primaryKey"`
				Title     string `gorm:"size:200;not null;index"`
				Content   string `gorm:"type:text"`
				Author    string `gorm:"size:64;index"`
				Status    string `gorm:"size:16;not null;default:draft;index"`
				CreatedAt int64  `gorm:"autoCreateTime"`
				UpdatedAt int64  `gorm:"autoUpdateTime"`
				DeletedAt *int64 `gorm:"index"`
			}
			return db.AutoMigrate(&Article{})
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable("articles")
		},
	})
}
