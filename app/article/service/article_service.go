// Package service 是 article 模块的业务逻辑层。
//
// 职责边界：
//   - 实现业务规则（如标题非空、状态合法、不可重复发布等）。
//   - 把 repository 返回的原始错误翻译为统一的 *errors.AppError。
//   - 不感知 HTTP（无 gin.Context），便于被 cron、队列、gRPC 等多入口复用。
//
// 错误码约定：本模块使用 4060xx 段（与框架预定义码段不冲突）。
package service

import (
	"context"
	stderrors "errors"
	"net/http"
	"strings"

	"gorm.io/gorm"

	"thinkgin/app/article/model"
	"thinkgin/app/article/repository"
	apperrors "thinkgin/app/errors"
)

// 模块级错误码（4060xx 段，避免与框架通用码冲突）。
var (
	ErrArticleNotFound  = apperrors.New(http.StatusNotFound, 406001, "文章不存在")
	ErrTitleRequired    = apperrors.New(http.StatusUnprocessableEntity, 406002, "标题不能为空")
	ErrInvalidStatus    = apperrors.New(http.StatusUnprocessableEntity, 406003, "非法的文章状态")
	ErrAlreadyPublished = apperrors.New(http.StatusConflict, 406004, "文章已是发布状态")
)

// ArticleService 承载文章相关业务逻辑。
type ArticleService struct {
	repo *repository.ArticleRepository
}

// New 创建服务实例。
func New(repo *repository.ArticleRepository) *ArticleService {
	return &ArticleService{repo: repo}
}

// CreateInput 创建文章的入参（已与 HTTP 解耦）。
type CreateInput struct {
	Title   string
	Content string
	Author  string
	Status  string
}

// Create 校验并创建文章。状态留空时默认草稿。
func (s *ArticleService) Create(ctx context.Context, in CreateInput) (*model.Article, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, ErrTitleRequired
	}
	status := in.Status
	if status == "" {
		status = model.StatusDraft
	}
	if !model.ValidStatus(status) {
		return nil, ErrInvalidStatus.WithData(map[string]any{"allowed": []string{
			model.StatusDraft, model.StatusPublished, model.StatusArchived,
		}})
	}

	a := &model.Article{
		Title:   title,
		Content: in.Content,
		Author:  strings.TrimSpace(in.Author),
		Status:  status,
	}
	if err := s.repo.Create(ctx, a); err != nil {
		return nil, apperrors.ErrInternal.WithCause(err)
	}
	return a, nil
}

// Get 按 ID 获取文章，不存在时返回 ErrArticleNotFound。
func (s *ArticleService) Get(ctx context.Context, id uint) (*model.Article, error) {
	a, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrArticleNotFound.WithData(map[string]any{"id": id})
		}
		return nil, apperrors.ErrInternal.WithCause(err)
	}
	return a, nil
}

// UpdateInput 更新文章的入参，nil 字段表示不更新该字段。
type UpdateInput struct {
	Title   *string
	Content *string
	Status  *string
}

// Update 局部更新文章字段，并做业务校验。
func (s *ArticleService) Update(ctx context.Context, id uint, in UpdateInput) (*model.Article, error) {
	a, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return nil, ErrTitleRequired
		}
		a.Title = title
	}
	if in.Content != nil {
		a.Content = *in.Content
	}
	if in.Status != nil {
		if !model.ValidStatus(*in.Status) {
			return nil, ErrInvalidStatus
		}
		a.Status = *in.Status
	}

	if err := s.repo.Update(ctx, a); err != nil {
		return nil, apperrors.ErrInternal.WithCause(err)
	}
	return a, nil
}

// Publish 将文章置为已发布；若已发布则返回 ErrAlreadyPublished。
func (s *ArticleService) Publish(ctx context.Context, id uint) (*model.Article, error) {
	a, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if a.Status == model.StatusPublished {
		return nil, ErrAlreadyPublished.WithData(map[string]any{"id": id})
	}
	a.Status = model.StatusPublished
	if err := s.repo.Update(ctx, a); err != nil {
		return nil, apperrors.ErrInternal.WithCause(err)
	}
	return a, nil
}

// Delete 删除文章（软删除）。删除前确认存在，以便返回 404 而非静默成功。
func (s *ArticleService) Delete(ctx context.Context, id uint) error {
	if _, err := s.Get(ctx, id); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return apperrors.ErrInternal.WithCause(err)
	}
	return nil
}

// ListInput 列表查询入参。
type ListInput struct {
	Status string
	Author string
	Offset int
	Limit  int
}

// List 分页查询文章，返回数据与总数。
func (s *ArticleService) List(ctx context.Context, in ListInput) ([]model.Article, int64, error) {
	if in.Status != "" && !model.ValidStatus(in.Status) {
		return nil, 0, ErrInvalidStatus
	}
	list, total, err := s.repo.List(ctx, repository.ListFilter{
		Status: in.Status,
		Author: in.Author,
		Offset: in.Offset,
		Limit:  in.Limit,
	})
	if err != nil {
		return nil, 0, apperrors.ErrInternal.WithCause(err)
	}
	return list, total, nil
}
