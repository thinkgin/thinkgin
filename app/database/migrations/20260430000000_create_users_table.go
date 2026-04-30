package migrations

import (
	"thinkgin/app/database"

	"gorm.io/gorm"
)

func init() {
	Register(database.Migration{
		Version: "20260430000000",
		Up: func(db *gorm.DB) error {
			type User struct {
				ID        uint   `gorm:"primaryKey"`
				Name      string `gorm:"size:100;not null"`
				Email     string `gorm:"size:255;uniqueIndex"`
				Password  string `gorm:"size:255"`
				CreatedAt int64  `gorm:"autoCreateTime"`
				UpdatedAt int64  `gorm:"autoUpdateTime"`
			}
			return db.AutoMigrate(&User{})
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable("users")
		},
	})
}
