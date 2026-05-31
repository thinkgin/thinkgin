// 本文件提供请求参数校验辅助函数。
//
// 设计思路：
//   - 基于 Gin 内置的 validator/v10，不引入额外依赖。
//   - BindAndValidate 封装 ShouldBind + 错误格式化，一行完成绑定+校验。
//   - 校验失败时自动返回 422 + 结构化字段级错误，符合 RFC 7807 精神。
//   - 与 app/errors 包集成：返回 *AppError 类型，可被 APIErrorHandler 自动捕获。
//
// 用法：
//
//	type CreateUserReq struct {
//	    Name  string `json:"name"  binding:"required,min=2,max=50"`
//	    Email string `json:"email" binding:"required,email"`
//	}
//
//	func CreateUser(c *gin.Context) {
//	    var req CreateUserReq
//	    if err := middleware.BindAndValidate(c, &req); err != nil {
//	        return // 已自动写入 422 响应
//	    }
//	    // req 已校验通过，直接使用
//	}
package middleware

import (
	"errors"
	"reflect"
	"strings"
	"sync"

	apperrors "thinkgin/app/errors"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// registerTagNameOnce 保证向 Gin 的 validator 注册 json tag 解析只执行一次。
var registerTagNameOnce sync.Once

// useJSONFieldNames 让 validator 的 fe.Field() 直接返回 struct 的 json tag 名，
// 从而错误信息里的字段名与请求/响应里的 JSON 字段一致（含 snake_case）。
// 仅作用于 Gin 默认的 validator 引擎；非该引擎时静默跳过。
func useJSONFieldNames() {
	registerTagNameOnce.Do(func() {
		v, ok := binding.Validator.Engine().(*validator.Validate)
		if !ok {
			return
		}
		v.RegisterTagNameFunc(func(field reflect.StructField) string {
			// 取 json tag 的第一段（去掉 ",omitempty" 等选项）。
			tag := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
			if tag == "-" || tag == "" {
				return field.Name
			}
			return tag
		})
	})
}

// FieldError 单个字段的校验错误详情。
type FieldError struct {
	Field   string `json:"field"`   // 字段名（JSON tag 优先）
	Tag     string `json:"tag"`     // 校验规则名（required/min/email 等）
	Message string `json:"message"` // 人可读的错误描述
}

// BindAndValidate 绑定请求参数并校验，失败时自动向 gin.Context 写入 422 响应。
// 返回 nil 表示绑定+校验成功；返回 error 表示已响应客户端，调用方应直接 return。
func BindAndValidate(c *gin.Context, obj interface{}) error {
	useJSONFieldNames()
	if err := c.ShouldBind(obj); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			fields := make([]FieldError, 0, len(ve))
			for _, fe := range ve {
				fields = append(fields, FieldError{
					Field:   jsonFieldName(fe),
					Tag:     fe.Tag(),
					Message: translateFieldError(fe),
				})
			}
			appErr := apperrors.ErrValidation.WithData(fields)
			APIAppError(c, appErr)
			return appErr
		}
		// 非校验错误（JSON 格式错误等）
		appErr := apperrors.ErrBadRequest.WithMsg(err.Error())
		APIAppError(c, appErr)
		return appErr
	}
	return nil
}

// jsonFieldName 返回字段名。配合 useJSONFieldNames 注册后，
// fe.Field() 已是 json tag 名；此处直接返回，保留函数以集中字段名取用逻辑。
func jsonFieldName(fe validator.FieldError) string {
	return fe.Field()
}

// translateFieldError 将 validator 错误翻译为中文消息。
// 覆盖最常用的规则，未匹配的回退到英文描述。
func translateFieldError(fe validator.FieldError) string {
	field := jsonFieldName(fe)
	switch fe.Tag() {
	case "required":
		return field + " 为必填项"
	case "min":
		return field + " 长度不能少于 " + fe.Param()
	case "max":
		return field + " 长度不能超过 " + fe.Param()
	case "email":
		return field + " 格式不正确"
	case "oneof":
		return field + " 必须是 [" + fe.Param() + "] 之一"
	case "gt":
		return field + " 必须大于 " + fe.Param()
	case "gte":
		return field + " 必须大于等于 " + fe.Param()
	case "lt":
		return field + " 必须小于 " + fe.Param()
	case "lte":
		return field + " 必须小于等于 " + fe.Param()
	case "len":
		return field + " 长度必须为 " + fe.Param()
	case "url":
		return field + " 必须是合法的 URL"
	case "numeric":
		return field + " 必须是数字"
	case "alpha":
		return field + " 只能包含字母"
	case "alphanum":
		return field + " 只能包含字母和数字"
	default:
		return field + " 校验失败: " + fe.Tag()
	}
}
