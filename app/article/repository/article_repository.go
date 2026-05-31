// Package repository 是 article 模块的数据访问层。
//
// 职责边界：
//   - 只负责"如何查数据库"，不含任何业务规则（业务规则在 service 层）。
//   - 接收 context.Context，所有查询都带上以支持超时与链路追踪。
//   - 返回原始 GORM 错误，由 service 层翻译为 *AppError。
//
// 这样分层的好处：repository 可被独立 mock 或替换（如换成内存实现做单测），
// service 层无需感知底层是 GORM 还是其他存储。
package repository

import (
	"context"

	"gorm.io/gorm"

	"thinkgin/app/article/model"
)

// ArticleRepository 封装文章的持久化操作。
type ArticleRepository struct {
	db *gorm.DB
}

// New 创建仓储实例。db 通常来自 database.Default()。
func New(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

// Create 插入一条文章记录，成功后 a.ID 被回填。
func (r *ArticleRepository) Create(ctx context.Context, a *model.Article) error {
	return r.db.WithContext(ctx).Create(a).Error
}

// FindByID 按主键查询。未找到时返回 gorm.ErrRecordNotFound。
func (r *ArticleRepository) FindByID(ctx context.Context, id uint) (*model.Article, error) {
	var a model.Article
	if err := r.db.WithContext(ctx).First(&a, id).Error; err != nil {
		return nil, err
	}
	return &a, nil
}

// Update 全量更新已存在的文章（基于主键）。
func (r *ArticleRepository) Update(ctx context.Context, a *model.Article) error {
	return r.db.WithContext(ctx).Save(a).Error
}

// Delete 软删除指定文章（gorm.Model 自带 DeletedAt）。
func (r *ArticleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Article{}, id).Error
}

// ListFilter 描述列表查询的过滤条件，零值字段表示不过滤。
type ListFilter struct {
	Status string // 按状态过滤
	Author string // 按作者过滤
	Offset int
	Limit  int
}

// List 按过滤条件分页查询，同时返回当前页数据与匹配的总数。
func (r *ArticleRepository) List(ctx context.Context, f ListFilter) ([]model.Article, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Article{})

	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.Author != "" {
		q = q.Where("author = ?", f.Author)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []model.Article
	if err := q.Order("id DESC").Offset(f.Offset).Limit(f.Limit).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}
