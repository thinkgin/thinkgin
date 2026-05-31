// Package controller 是 article 模块的 HTTP 处理层。
//
// 职责边界：
//   - 绑定与校验请求参数（middleware.BindAndValidate）。
//   - 解析分页（pagination.FromQuery）。
//   - 调用 service，并把返回的 *AppError 统一交给 APIAppError 输出。
//   - 不含业务规则，保持 handler 轻薄。
//
// 依赖注入：控制器持有 *ArticleService，由 route 层在装配时构造并注入，
// 便于测试时替换为基于内存 DB 的服务实例。
package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"thinkgin/app/article/service"
	apperrors "thinkgin/app/errors"
	"thinkgin/app/pagination"
	"thinkgin/extend/middleware"
)

// ArticleController 持有业务服务依赖。
type ArticleController struct {
	svc *service.ArticleService
}

// New 创建控制器。
func New(svc *service.ArticleService) *ArticleController {
	return &ArticleController{svc: svc}
}

// createRequest 创建文章的请求体。binding tag 由 validator 校验。
type createRequest struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content"`
	Author  string `json:"author" binding:"max=64"`
	Status  string `json:"status" binding:"omitempty,oneof=draft published archived"`
}

// updateRequest 更新文章的请求体，指针字段区分"未传"与"传空值"。
type updateRequest struct {
	Title   *string `json:"title" binding:"omitempty,min=1,max=200"`
	Content *string `json:"content"`
	Status  *string `json:"status" binding:"omitempty,oneof=draft published archived"`
}

// Create 处理 POST /api/v1/articles。
func (ctl *ArticleController) Create(c *gin.Context) {
	var req createRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return // BindAndValidate 已写入 422 响应
	}
	a, err := ctl.svc.Create(c.Request.Context(), service.CreateInput{
		Title:   req.Title,
		Content: req.Content,
		Author:  req.Author,
		Status:  req.Status,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	middleware.APISuccess(c, a)
}

// Get 处理 GET /api/v1/articles/:id。
func (ctl *ArticleController) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	a, err := ctl.svc.Get(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	middleware.APISuccess(c, a)
}

// List 处理 GET /api/v1/articles，支持 ?status= &author= 过滤与分页。
func (ctl *ArticleController) List(c *gin.Context) {
	p := pagination.FromQuery(c)
	list, total, err := ctl.svc.List(c.Request.Context(), service.ListInput{
		Status: c.Query("status"),
		Author: c.Query("author"),
		Offset: p.Offset(),
		Limit:  p.PageSize,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	middleware.APISuccess(c, pagination.NewResult(p, total, list))
}

// Update 处理 PUT /api/v1/articles/:id。
func (ctl *ArticleController) Update(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req updateRequest
	if err := middleware.BindAndValidate(c, &req); err != nil {
		return
	}
	a, err := ctl.svc.Update(c.Request.Context(), id, service.UpdateInput{
		Title:   req.Title,
		Content: req.Content,
		Status:  req.Status,
	})
	if err != nil {
		respondError(c, err)
		return
	}
	middleware.APISuccess(c, a)
}

// Publish 处理 POST /api/v1/articles/:id/publish。
func (ctl *ArticleController) Publish(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	a, err := ctl.svc.Publish(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	middleware.APISuccess(c, a)
}

// Delete 处理 DELETE /api/v1/articles/:id。
func (ctl *ArticleController) Delete(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := ctl.svc.Delete(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	middleware.APISuccess(c, gin.H{"deleted": id})
}

// parseID 解析 :id 路径参数，非法时写入 400 并返回 ok=false。
func parseID(c *gin.Context) (uint, bool) {
	raw := c.Param("id")
	v, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || v == 0 {
		middleware.APIError(c, http.StatusBadRequest, apperrors.ErrBadRequest.Code, "非法的资源 ID")
		return 0, false
	}
	return uint(v), true
}

// respondError 把 service 返回的错误统一格式化。
// *AppError 走结构化输出，其余兜底为 500。
func respondError(c *gin.Context, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		middleware.APIAppError(c, appErr)
		return
	}
	middleware.APIError(c, http.StatusInternalServerError, apperrors.ErrInternal.Code, "服务内部错误")
}
