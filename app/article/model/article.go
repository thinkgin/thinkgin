// Package model 提供 article 模块的 GORM 数据模型。
//
// 本模块是 ThinkGin 的「完整业务模块范例」，演示生产项目的标准分层：
//
//	model      —— 数据结构与表映射（本文件）
//	repository —— 数据访问层，封装所有 GORM 查询
//	service    —— 业务逻辑层，返回结构化 *AppError
//	controller —— HTTP 处理层，负责绑定/校验/分页/响应
//
// 仿照本模块即可快速搭建自己的 CRUD 模块。
package model

import "gorm.io/gorm"

// Article 文章模型。
// 内嵌 gorm.Model 获得 ID / CreatedAt / UpdatedAt / DeletedAt 标准字段（含软删除）。
type Article struct {
	gorm.Model
	Title   string `gorm:"size:200;not null;index" json:"title"`
	Content string `gorm:"type:text" json:"content"`
	Author  string `gorm:"size:64;index" json:"author"`
	// Status：draft（草稿）/ published（已发布）/ archived（归档）。
	Status string `gorm:"size:16;not null;default:draft;index" json:"status"`
}

// TableName 显式表名。
func (Article) TableName() string {
	return "articles"
}

// 文章状态常量，供 service / controller 校验复用。
const (
	StatusDraft     = "draft"
	StatusPublished = "published"
	StatusArchived  = "archived"
)

// ValidStatus 报告 s 是否为合法的文章状态。
func ValidStatus(s string) bool {
	switch s {
	case StatusDraft, StatusPublished, StatusArchived:
		return true
	default:
		return false
	}
}
