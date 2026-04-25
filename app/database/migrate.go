// 本文件提供基于 GORM 的数据库迁移系统。
//
// 设计思路：
//   - 使用单独的 _migrations 表记录已执行的迁移，支持版本排序。
//   - 迁移以 Go 函数形式注册，保持类型安全与编译期检查。
//   - 支持 Up（升级）和 Down（回退），按版本号字典序执行。
//   - 不引入第三方迁移库（如 golang-migrate），保持框架零额外依赖。
package database

import (
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

// MigrationRecord 对应 _migrations 表中的一行。
type MigrationRecord struct {
	ID        uint      `gorm:"primaryKey"`
	Version   string    `gorm:"uniqueIndex;size:255"`
	AppliedAt time.Time `gorm:"autoCreateTime"`
}

func (MigrationRecord) TableName() string {
	return "_migrations"
}

// Migration 定义单个迁移。
type Migration struct {
	Version string
	Up      func(db *gorm.DB) error
	Down    func(db *gorm.DB) error
}

// Migrator 管理迁移的注册与执行。
type Migrator struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrator 创建迁移管理器。
func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{db: db}
}

// Register 注册一个或多个迁移。
func (m *Migrator) Register(migrations ...Migration) {
	m.migrations = append(m.migrations, migrations...)
}

// Migrate 执行所有未应用的迁移（按版本字典序升序）。
func (m *Migrator) Migrate() error {
	if err := m.db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("migrate: create _migrations table: %w", err)
	}

	applied, err := m.appliedVersions()
	if err != nil {
		return err
	}

	// 按版本排序
	sorted := make([]Migration, len(m.migrations))
	copy(sorted, m.migrations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Version < sorted[j].Version
	})

	for _, mg := range sorted {
		if applied[mg.Version] {
			continue
		}
		if mg.Up == nil {
			continue
		}

		if err := mg.Up(m.db); err != nil {
			return fmt.Errorf("migrate up %s: %w", mg.Version, err)
		}

		record := MigrationRecord{Version: mg.Version}
		if err := m.db.Create(&record).Error; err != nil {
			return fmt.Errorf("migrate record %s: %w", mg.Version, err)
		}
	}

	return nil
}

// Rollback 回退最后 n 个已应用的迁移。
// n <= 0 时回退全部。
func (m *Migrator) Rollback(n int) error {
	if err := m.db.AutoMigrate(&MigrationRecord{}); err != nil {
		return fmt.Errorf("rollback: create _migrations table: %w", err)
	}

	var records []MigrationRecord
	q := m.db.Order("version DESC")
	if n > 0 {
		q = q.Limit(n)
	}
	if err := q.Find(&records).Error; err != nil {
		return fmt.Errorf("rollback: query records: %w", err)
	}

	// 建立 version→Migration 映射
	mgMap := make(map[string]Migration, len(m.migrations))
	for _, mg := range m.migrations {
		mgMap[mg.Version] = mg
	}

	for _, rec := range records {
		mg, ok := mgMap[rec.Version]
		if !ok || mg.Down == nil {
			continue
		}

		if err := mg.Down(m.db); err != nil {
			return fmt.Errorf("rollback down %s: %w", rec.Version, err)
		}

		if err := m.db.Where("version = ?", rec.Version).Delete(&MigrationRecord{}).Error; err != nil {
			return fmt.Errorf("rollback delete record %s: %w", rec.Version, err)
		}
	}

	return nil
}

// Status 返回所有已注册迁移的版本及其是否已应用。
func (m *Migrator) Status() ([]MigrationStatus, error) {
	if err := m.db.AutoMigrate(&MigrationRecord{}); err != nil {
		return nil, err
	}

	applied, err := m.appliedVersions()
	if err != nil {
		return nil, err
	}

	sorted := make([]Migration, len(m.migrations))
	copy(sorted, m.migrations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Version < sorted[j].Version
	})

	out := make([]MigrationStatus, len(sorted))
	for i, mg := range sorted {
		out[i] = MigrationStatus{
			Version: mg.Version,
			Applied: applied[mg.Version],
		}
	}
	return out, nil
}

// MigrationStatus 描述单个迁移的状态。
type MigrationStatus struct {
	Version string
	Applied bool
}

func (m *Migrator) appliedVersions() (map[string]bool, error) {
	var records []MigrationRecord
	if err := m.db.Find(&records).Error; err != nil {
		return nil, fmt.Errorf("migrate: query applied versions: %w", err)
	}
	applied := make(map[string]bool, len(records))
	for _, r := range records {
		applied[r.Version] = true
	}
	return applied, nil
}
