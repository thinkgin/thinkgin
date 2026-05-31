package service

import (
	"context"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"thinkgin/app/article/model"
	"thinkgin/app/article/repository"
	apperrors "thinkgin/app/errors"
)

// newTestService 用共享内存 SQLite 构造一个真实的 service，覆盖完整数据链路。
// 使用内存库避免 Windows 下临时文件被占用无法清理的问题；
// 通过 t.Cleanup 关闭连接，确保内存库随用例释放。
func newTestService(t *testing.T) *ArticleService {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	// 单连接：保证 :memory: 库在用例期间始终是同一个实例。
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("db handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	if err := db.AutoMigrate(&model.Article{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return New(repository.New(db))
}

func TestCreate_Success(t *testing.T) {
	svc := newTestService(t)
	a, err := svc.Create(context.Background(), CreateInput{Title: "Hello", Author: "alice"})
	if err != nil {
		t.Fatalf("Create error: %v", err)
	}
	if a.ID == 0 {
		t.Error("expected ID to be assigned")
	}
	if a.Status != model.StatusDraft {
		t.Errorf("default status = %q, want draft", a.Status)
	}
}

func TestCreate_EmptyTitle(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.Create(context.Background(), CreateInput{Title: "   "})
	if !errors.Is(err, ErrTitleRequired) {
		t.Errorf("err = %v, want ErrTitleRequired", err)
	}
}

func TestCreate_InvalidStatus(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.Create(context.Background(), CreateInput{Title: "x", Status: "bogus"})
	if !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("err = %v, want ErrInvalidStatus", err)
	}
}

func TestGet_NotFound(t *testing.T) {
	svc := newTestService(t)
	_, err := svc.Get(context.Background(), 999)
	if !errors.Is(err, ErrArticleNotFound) {
		t.Errorf("err = %v, want ErrArticleNotFound", err)
	}
	// 校验 HTTP 状态码映射正确。
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) && appErr.HTTPStatus != 404 {
		t.Errorf("HTTPStatus = %d, want 404", appErr.HTTPStatus)
	}
}

func TestUpdate_PartialFields(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	a, _ := svc.Create(ctx, CreateInput{Title: "old", Content: "c"})

	newTitle := "new"
	updated, err := svc.Update(ctx, a.ID, UpdateInput{Title: &newTitle})
	if err != nil {
		t.Fatalf("Update error: %v", err)
	}
	if updated.Title != "new" {
		t.Errorf("title = %q, want new", updated.Title)
	}
	if updated.Content != "c" {
		t.Errorf("content should be unchanged, got %q", updated.Content)
	}
}

func TestPublish_Flow(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	a, _ := svc.Create(ctx, CreateInput{Title: "to publish"})

	pub, err := svc.Publish(ctx, a.ID)
	if err != nil {
		t.Fatalf("Publish error: %v", err)
	}
	if pub.Status != model.StatusPublished {
		t.Errorf("status = %q, want published", pub.Status)
	}

	// 二次发布应冲突。
	_, err = svc.Publish(ctx, a.ID)
	if !errors.Is(err, ErrAlreadyPublished) {
		t.Errorf("err = %v, want ErrAlreadyPublished", err)
	}
}

func TestDelete_ThenNotFound(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	a, _ := svc.Create(ctx, CreateInput{Title: "doomed"})

	if err := svc.Delete(ctx, a.ID); err != nil {
		t.Fatalf("Delete error: %v", err)
	}
	if _, err := svc.Get(ctx, a.ID); !errors.Is(err, ErrArticleNotFound) {
		t.Errorf("after delete, Get err = %v, want ErrArticleNotFound", err)
	}
	// 删除不存在的记录应返回 404。
	if err := svc.Delete(ctx, 12345); !errors.Is(err, ErrArticleNotFound) {
		t.Errorf("delete missing err = %v, want ErrArticleNotFound", err)
	}
}

func TestList_FilterAndPaginate(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_, _ = svc.Create(ctx, CreateInput{Title: "draft", Author: "alice"})
	}
	for i := 0; i < 3; i++ {
		_, _ = svc.Create(ctx, CreateInput{Title: "pub", Author: "bob", Status: model.StatusPublished})
	}

	// 按状态过滤
	list, total, err := svc.List(ctx, ListInput{Status: model.StatusPublished, Limit: 10})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if total != 3 || len(list) != 3 {
		t.Errorf("published total=%d len=%d, want 3/3", total, len(list))
	}

	// 分页：每页 2 条，第一页应返回 2 条但总数仍为 5。
	list, total, err = svc.List(ctx, ListInput{Author: "alice", Offset: 0, Limit: 2})
	if err != nil {
		t.Fatalf("List error: %v", err)
	}
	if total != 5 || len(list) != 2 {
		t.Errorf("alice page1 total=%d len=%d, want 5/2", total, len(list))
	}
}

func TestList_InvalidStatusFilter(t *testing.T) {
	svc := newTestService(t)
	_, _, err := svc.List(context.Background(), ListInput{Status: "bogus", Limit: 10})
	if !errors.Is(err, ErrInvalidStatus) {
		t.Errorf("err = %v, want ErrInvalidStatus", err)
	}
}
