package database

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	return db
}

func TestMigrator_MigrateAndRollback(t *testing.T) {
	db := testDB(t)

	m := NewMigrator(db)
	m.Register(
		Migration{
			Version: "001_create_users",
			Up: func(db *gorm.DB) error {
				return db.Exec("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)").Error
			},
			Down: func(db *gorm.DB) error {
				return db.Exec("DROP TABLE IF EXISTS users").Error
			},
		},
		Migration{
			Version: "002_add_email",
			Up: func(db *gorm.DB) error {
				return db.Exec("ALTER TABLE users ADD COLUMN email TEXT").Error
			},
			Down: func(db *gorm.DB) error {
				// SQLite 不支持 DROP COLUMN，跳过
				return nil
			},
		},
	)

	// 执行全部迁移
	if err := m.Migrate(); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	// 再次执行应为幂等
	if err := m.Migrate(); err != nil {
		t.Fatalf("Migrate (idempotent): %v", err)
	}

	// 检查状态
	status, err := m.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status) != 2 {
		t.Fatalf("Status len = %d, want 2", len(status))
	}
	for _, s := range status {
		if !s.Applied {
			t.Errorf("migration %s should be applied", s.Version)
		}
	}

	// 回退 1 个
	if err := m.Rollback(1); err != nil {
		t.Fatalf("Rollback(1): %v", err)
	}

	status, _ = m.Status()
	if status[0].Applied != true {
		t.Error("001 should still be applied")
	}
	if status[1].Applied != false {
		t.Error("002 should be rolled back")
	}

	// 回退全部
	if err := m.Rollback(0); err != nil {
		t.Fatalf("Rollback(0): %v", err)
	}

	status, _ = m.Status()
	for _, s := range status {
		if s.Applied {
			t.Errorf("migration %s should not be applied after full rollback", s.Version)
		}
	}
}

func TestMigrator_EmptyMigrationsIsNoop(t *testing.T) {
	db := testDB(t)
	m := NewMigrator(db)

	if err := m.Migrate(); err != nil {
		t.Fatalf("empty Migrate: %v", err)
	}

	status, err := m.Status()
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if len(status) != 0 {
		t.Errorf("len = %d, want 0", len(status))
	}
}
