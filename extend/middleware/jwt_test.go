package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"

	"thinkgin/app"
)

// mintExpiredToken 直接用库签一个已过期 token，绕开 JWTIssue 的 ttl 兜底。
func mintExpiredToken(t *testing.T, secret string) string {
	t.Helper()
	now := time.Now()
	claims := JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "u1",
			IssuedAt:  jwt.NewNumericDate(now.Add(-2 * time.Hour)),
			ExpiresAt: jwt.NewNumericDate(now.Add(-time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign: %v", err)
	}
	return s
}

func init() { gin.SetMode(gin.TestMode) }

// resetJWTConfig 把 JWT 配置重置到可预测状态，避免用例间干扰。
func resetJWTConfig(t *testing.T, secret string, expireSec int) {
	t.Helper()
	app.Config = &app.GlobalConfig{}
	app.Config.App.JWT.Secret = secret
	app.Config.App.JWT.Expire = expireSec
}

func TestExtractBearer(t *testing.T) {
	cases := map[string]string{
		"Bearer abc.def.ghi":   "abc.def.ghi",
		"bearer abc":           "abc",
		"  Bearer   x  ":       "x",
		"Basic abc":            "",
		"":                     "",
		"Bearer":               "",
		"Bearer a b":           "",
	}
	for in, want := range cases {
		if got := extractBearer(in); got != want {
			t.Errorf("extractBearer(%q)=%q, want %q", in, got, want)
		}
	}
}

func TestJWTIssueAndParse_Roundtrip(t *testing.T) {
	resetJWTConfig(t, "test-secret", 3600)
	token, err := JWTIssue("u42", map[string]any{"role": "admin"}, 0)
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if strings.Count(token, ".") != 2 {
		t.Fatalf("token shape wrong: %q", token)
	}
	claims, err := JWTParse(token)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if claims.Subject != "u42" {
		t.Errorf("subject=%q, want u42", claims.Subject)
	}
	if claims.Extra["role"] != "admin" {
		t.Errorf("extra.role=%v, want admin", claims.Extra["role"])
	}
}

func TestJWTIssue_EmptySecretRefuses(t *testing.T) {
	resetJWTConfig(t, "", 3600)
	if _, err := JWTIssue("u1", nil, 0); err == nil {
		t.Fatal("expected error when secret empty")
	}
}

func TestJWTParse_ExpiredReturnsError(t *testing.T) {
	resetJWTConfig(t, "test-secret", 3600)
	token := mintExpiredToken(t, "test-secret")
	if _, err := JWTParse(token); err == nil {
		t.Fatal("expected expired error")
	}
}

func TestJWTParse_TamperedSignatureRejected(t *testing.T) {
	resetJWTConfig(t, "test-secret", 3600)
	token, _ := JWTIssue("u1", nil, 0)
	// 翻转签名部分多个字符，确保签名一定失效
	parts := strings.SplitN(token, ".", 3)
	sig := parts[2]
	flipped := make([]byte, len(sig))
	for i, b := range []byte(sig) {
		if b == 'A' {
			flipped[i] = 'B'
		} else {
			flipped[i] = 'A'
		}
	}
	tampered := parts[0] + "." + parts[1] + "." + string(flipped)
	if _, err := JWTParse(tampered); err == nil {
		t.Fatal("expected signature error")
	}
}

func TestJWTAuth_MissingHeaderReturns401(t *testing.T) {
	resetJWTConfig(t, "test-secret", 3600)

	r := gin.New()
	r.Use(JWTAuth())
	r.GET("/secure", func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/secure", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401", w.Code)
	}
	if !strings.Contains(w.Body.String(), "Authorization") {
		t.Errorf("body should mention Authorization, got %q", w.Body.String())
	}
}

func TestJWTAuth_ValidTokenPassesAndExposesClaims(t *testing.T) {
	resetJWTConfig(t, "test-secret", 3600)
	token, _ := JWTIssue("u42", map[string]any{"role": "admin"}, 0)

	r := gin.New()
	r.Use(JWTAuth())
	r.GET("/me", func(c *gin.Context) {
		claims := JWTFromContext(c)
		if claims == nil {
			t.Fatal("claims should be set")
		}
		c.String(200, claims.Subject)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("status=%d, body=%s", w.Code, w.Body.String())
	}
	if w.Body.String() != "u42" {
		t.Errorf("body=%q, want u42", w.Body.String())
	}
}

func TestJWTAuth_ExpiredTokenReturns401(t *testing.T) {
	resetJWTConfig(t, "test-secret", 3600)
	token := mintExpiredToken(t, "test-secret")

	r := gin.New()
	r.Use(JWTAuth())
	r.GET("/me", func(c *gin.Context) { c.String(200, "ok") })

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status=%d, want 401", w.Code)
	}
	if !strings.Contains(w.Body.String(), "expired") {
		t.Errorf("body should mention expired, got %q", w.Body.String())
	}
}

// oppositeChar 返回 base64url 字符集里与 c 不同的另一字符，用于伪造签名。
func oppositeChar(c byte) string {
	if c == 'A' {
		return "B"
	}
	return "A"
}
