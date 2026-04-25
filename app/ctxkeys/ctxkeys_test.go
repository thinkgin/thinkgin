package ctxkeys

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRequestID_SetGet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 未设置时返回空字符串
	if got := GetRequestID(c); got != "" {
		t.Errorf("GetRequestID() = %q, want empty", got)
	}

	SetRequestID(c, "req-abc-123")
	if got := GetRequestID(c); got != "req-abc-123" {
		t.Errorf("GetRequestID() = %q, want req-abc-123", got)
	}
}

func TestCSRFToken_SetGet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	if got := GetCSRFToken(c); got != "" {
		t.Errorf("GetCSRFToken() = %q, want empty", got)
	}

	SetCSRFToken(c, "csrf-xyz-789")
	if got := GetCSRFToken(c); got != "csrf-xyz-789" {
		t.Errorf("GetCSRFToken() = %q, want csrf-xyz-789", got)
	}
}

func TestKeys_AreNamespaced(t *testing.T) {
	// 确保 key 常量使用了 thinkgin: 前缀避免冲突
	if RequestID != "thinkgin:request_id" {
		t.Errorf("RequestID = %q, want thinkgin:request_id", RequestID)
	}
	if CSRFToken != "thinkgin:csrf_token" {
		t.Errorf("CSRFToken = %q, want thinkgin:csrf_token", CSRFToken)
	}
}
