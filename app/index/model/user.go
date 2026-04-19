// Package model 提供 index 模块的 GORM Model 示例。
// 演示如何定义表结构并通过 database.Default() 进行 CRUD。
package model

import (
	"time"

	"gorm.io/gorm"
)

// User 示例用户模型。
// 使用内嵌 gorm.Model 获得 ID/CreatedAt/UpdatedAt/DeletedAt 四个标准字段。
type User struct {
	gorm.Model
	Name  string `gorm:"size:64;not null;uniqueIndex:uk_name"`
	Email string `gorm:"size:128;uniqueIndex:uk_email"`

	// LastLoginAt 用独立 *time.Time 以支持 NULL 语义。
	LastLoginAt *time.Time
}

// TableName 显式指定表名，避免 GORM 默认复数化在中文项目里造成困惑。
// 实际项目可直接删除该方法以使用 GORM 的约定命名（users）。
func (User) TableName() string {
	return "users"
}
