// JWT 鉴权中间件。
//
// 设计：
//   - 签名算法固定 HS256，密钥从 app.yaml 的 app.jwt.secret 读取。
//   - 过期时间默认 app.jwt.expire 秒；Issue 时可按需覆盖。
//   - 中间件从 Authorization: Bearer <token> 解析；失败走 APIError(401)。
//   - 解析成功后把 *JWTClaims 放进 gin.Context，业务用 JWTFromContext(c) 取。
//
// 不自动实现的能力（留给业务）：
//   - 刷新 token：刷新策略因业务而异（例如双 token、滑动窗口），不做统一。
//   - 黑名单 / 撤销：需外部存储配合（Redis），避免强耦合。
package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"thinkgin/app"
)

// jwtContextKey 在 gin.Context.Keys 中标识 *JWTClaims；未导出常量避免外部覆盖。
const jwtContextKey = "thinkgin:jwt_claims"

// JWTClaims 是 ThinkGin 默认的 Claims 结构。
// Subject 复用 RFC 7519 的 sub 字段，业务可把用户主键塞进来。
// Extra 用于放非敏感的附加信息（角色、租户），签名会覆盖这些字段。
type JWTClaims struct {
	Extra map[string]any `json:"extra,omitempty"`
	jwt.RegisteredClaims
}

// JWTIssue 签发 token。
//
//	subject  - 一般放用户 ID
//	extra    - 附加声明；nil 表示不附加
//	ttl      - 有效期；<=0 时使用 app.jwt.expire，仍为 0 则兜底 2 小时
//
// 返回编码后的字符串 token。密钥缺失时返回错误，避免签出"空密钥"token 造成安全陷阱。
func JWTIssue(subject string, extra map[string]any, ttl time.Duration) (string, error) {
	secret, fallbackTTL, err := jwtConfig()
	if err != nil {
		return "", err
	}
	if ttl <= 0 {
		ttl = fallbackTTL
	}
	now := time.Now()
	claims := JWTClaims{
		Extra: extra,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// JWTParse 解析并校验 token 字符串。
// 严格校验签名算法，避免 alg=none 之类的经典漏洞。
func JWTParse(tokenStr string) (*JWTClaims, error) {
	secret, _, err := jwtConfig()
	if err != nil {
		return nil, err
	}
	claims := &JWTClaims{}
	_, err = jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return secret, nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// JWTAuth 返回 Gin 中间件：强制校验 Bearer token。
// 失败时立即 401 + 统一错误 JSON，并 Abort 后续 handler。
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := extractBearer(c.GetHeader("Authorization"))
		if raw == "" {
			APIError(c, http.StatusUnauthorized, http.StatusUnauthorized, "missing Authorization header")
			c.Abort()
			return
		}
		claims, err := JWTParse(raw)
		if err != nil {
			APIError(c, http.StatusUnauthorized, http.StatusUnauthorized, jwtErrorMessage(err))
			c.Abort()
			return
		}
		c.Set(jwtContextKey, claims)
		c.Next()
	}
}

// JWTFromContext 从 gin.Context 取出已校验的 Claims。未挂载时返回 nil。
func JWTFromContext(c *gin.Context) *JWTClaims {
	v, ok := c.Get(jwtContextKey)
	if !ok {
		return nil
	}
	claims, _ := v.(*JWTClaims)
	return claims
}

// jwtConfig 读取 secret 与默认 TTL。secret 为空视为未配置。
// 默认 TTL 在配置中为 0 或负值时兜底 2 小时，与 session 默认一致。
func jwtConfig() ([]byte, time.Duration, error) {
	cfg := app.GetConfig()
	if cfg == nil {
		return nil, 0, errors.New("jwt: global config is nil")
	}
	if cfg.App.JWT.Secret == "" {
		return nil, 0, errors.New("jwt: app.jwt.secret is empty; refuse to sign/verify")
	}
	ttl := time.Duration(cfg.App.JWT.Expire) * time.Second
	if ttl <= 0 {
		ttl = 2 * time.Hour
	}
	return []byte(cfg.App.JWT.Secret), ttl, nil
}

// extractBearer 容忍大小写与多余空白的 "Bearer xxx" 头。
// 非 Bearer 方案直接返回空串，让上层当作未认证处理。
func extractBearer(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	parts := strings.Fields(header)
	if len(parts) != 2 {
		return ""
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

// jwtErrorMessage 把底层库错误映射为对外可见的简短提示，
// 避免把内部实现细节泄漏给客户端。
func jwtErrorMessage(err error) string {
	switch {
	case errors.Is(err, jwt.ErrTokenExpired):
		return "token expired"
	case errors.Is(err, jwt.ErrTokenNotValidYet):
		return "token not yet valid"
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return "invalid token signature"
	case errors.Is(err, jwt.ErrTokenMalformed):
		return "malformed token"
	default:
		return "invalid token"
	}
}
