// Package pagination 提供统一的分页请求解析与响应封装。
//
// 设计目标：
//   - 从 query string 自动解析 page / page_size 参数。
//   - 尊重 config/app.yaml 中的 pagination.page_size 和 pagination.max_page_size 配置。
//   - 提供 PageResult 泛型结构体，统一所有列表接口的 JSON 响应格式。
//
// 用法：
//
//	func ListUsers(c *gin.Context) {
//	    p := pagination.FromQuery(c)
//	    var users []User
//	    var total int64
//	    db.Model(&User{}).Count(&total).Offset(p.Offset()).Limit(p.PageSize).Find(&users)
//	    middleware.APISuccess(c, pagination.NewResult(p, total, users))
//	}
package pagination

import (
	"strconv"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
)

// Params 分页请求参数（已归一化）。
type Params struct {
	Page     int `json:"page"`      // 当前页，从 1 开始
	PageSize int `json:"page_size"` // 每页数量
}

// Offset 返回 SQL OFFSET 值。
func (p Params) Offset() int {
	return (p.Page - 1) * p.PageSize
}

// Result 统一的分页列表响应结构。
type Result struct {
	List       interface{} `json:"list"`        // 当前页数据
	Total      int64       `json:"total"`       // 总记录数
	Page       int         `json:"page"`        // 当前页码
	PageSize   int         `json:"page_size"`   // 每页数量
	TotalPages int         `json:"total_pages"` // 总页数
}

// FromQuery 从 gin.Context 的 query string 解析分页参数。
// 缺省值来自 app.yaml 的 pagination 配置。
func FromQuery(c *gin.Context) Params {
	cfg := app.GetConfig()
	defaultSize := 20
	maxSize := 100

	if cfg != nil {
		if cfg.App.Pagination.PageSize > 0 {
			defaultSize = cfg.App.Pagination.PageSize
		}
		if cfg.App.Pagination.MaxPageSize > 0 {
			maxSize = cfg.App.Pagination.MaxPageSize
		}
	}

	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "page_size", defaultSize)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultSize
	}
	if pageSize > maxSize {
		pageSize = maxSize
	}

	return Params{Page: page, PageSize: pageSize}
}

// NewResult 创建分页结果。
func NewResult(p Params, total int64, list interface{}) Result {
	totalPages := 0
	if p.PageSize > 0 {
		totalPages = int((total + int64(p.PageSize) - 1) / int64(p.PageSize))
	}
	return Result{
		List:       list,
		Total:      total,
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalPages: totalPages,
	}
}

func queryInt(c *gin.Context, key string, def int) int {
	s := c.Query(key)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
}
