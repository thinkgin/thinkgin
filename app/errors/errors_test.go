package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestNew_CreatesError(t *testing.T) {
	e := New(http.StatusBadRequest, 400001, "invalid param")
	if e.HTTPStatus != http.StatusBadRequest {
		t.Errorf("HTTPStatus = %d, want %d", e.HTTPStatus, http.StatusBadRequest)
	}
	if e.Code != 400001 {
		t.Errorf("Code = %d, want 400001", e.Code)
	}
	if e.Message != "invalid param" {
		t.Errorf("Message = %q, want %q", e.Message, "invalid param")
	}
}

func TestAppError_ErrorString(t *testing.T) {
	e := New(400, 400001, "bad request")
	want := "[400001] bad request"
	if e.Error() != want {
		t.Errorf("Error() = %q, want %q", e.Error(), want)
	}
}

func TestAppError_ErrorStringWithCause(t *testing.T) {
	cause := fmt.Errorf("db connection failed")
	e := New(500, 500001, "internal").WithCause(cause)
	if e.Error() != "[500001] internal: db connection failed" {
		t.Errorf("Error() = %q", e.Error())
	}
}

func TestAppError_WithMsg(t *testing.T) {
	original := New(400, 400001, "original")
	derived := original.WithMsg("overridden")

	if derived.Message != "overridden" {
		t.Errorf("derived.Message = %q, want %q", derived.Message, "overridden")
	}
	// 原始不被修改
	if original.Message != "original" {
		t.Errorf("original.Message = %q, should not change", original.Message)
	}
}

func TestAppError_WithData(t *testing.T) {
	e := New(400, 400001, "err").WithData(map[string]string{"field": "name"})
	if e.Data == nil {
		t.Error("Data should not be nil")
	}
}

func TestAppError_WithCause(t *testing.T) {
	cause := fmt.Errorf("root cause")
	e := New(500, 500001, "err").WithCause(cause)
	if e.Unwrap() != cause {
		t.Error("Unwrap should return the cause")
	}
}

func TestAppError_Is(t *testing.T) {
	e1 := New(400, 400001, "msg1")
	e2 := New(422, 400001, "msg2") // 同 Code，不同 HTTPStatus
	e3 := New(400, 400002, "msg3") // 不同 Code

	if !errors.Is(e1, e2) {
		t.Error("errors.Is should match same Code")
	}
	if errors.Is(e1, e3) {
		t.Error("errors.Is should not match different Code")
	}
}

func TestAppError_As(t *testing.T) {
	e := New(400, 400001, "test")
	var target *AppError
	if !errors.As(e, &target) {
		t.Error("errors.As should succeed")
	}
	if target.Code != 400001 {
		t.Errorf("target.Code = %d, want 400001", target.Code)
	}
}

func TestPredefinedErrors(t *testing.T) {
	tests := []struct {
		err  *AppError
		code int
	}{
		{ErrBadRequest, 400000},
		{ErrUnauthorized, 401000},
		{ErrTokenExpired, 401001},
		{ErrForbidden, 403000},
		{ErrNotFound, 404000},
		{ErrValidation, 422000},
		{ErrTooManyRequests, 429000},
		{ErrInternal, 500000},
		{ErrServiceUnavailable, 503000},
		{ErrTimeout, 504000},
	}
	for _, tt := range tests {
		if tt.err.Code != tt.code {
			t.Errorf("%v.Code = %d, want %d", tt.err, tt.err.Code, tt.code)
		}
	}
}
