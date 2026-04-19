// 本文件包含各驱动的 DSN 构造与 GORM 打开逻辑。
//
// 设计说明：
//   - SQLite 使用 github.com/glebarez/sqlite（纯 Go，无 cgo 依赖），
//     避免 Windows 用户需要 gcc 工具链。
//   - MySQL / PostgreSQL 使用 GORM 官方 driver，生态最好。
//   - 连接池参数（max_open / max_idle / max_lifetime）统一在 applyPool 中消费。
package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// openConnection 按 driver 分发，构造 GORM 实例并应用连接池配置。
func openConnection(driver string, conn map[string]interface{}) (*gorm.DB, error) {
	gormCfg := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: strOr(conn, "table_prefix", ""),
		},
		Logger: logger.Default.LogMode(logger.Warn),
	}

	var (
		db  *gorm.DB
		err error
	)

	switch driver {
	case "mysql":
		db, err = gorm.Open(mysql.Open(buildMySQLDSN(conn)), gormCfg)
	case "postgres", "postgresql":
		db, err = gorm.Open(postgres.Open(buildPostgresDSN(conn)), gormCfg)
	case "sqlite", "sqlite3":
		path := resolveSQLitePath(conn)
		db, err = gorm.Open(sqlite.Open(path), gormCfg)
	default:
		return nil, fmt.Errorf("unsupported driver: %q", driver)
	}

	if err != nil {
		return nil, err
	}

	if err := applyPool(db, conn); err != nil {
		return nil, fmt.Errorf("apply pool: %w", err)
	}
	return db, nil
}

// buildMySQLDSN 按 github.com/go-sql-driver/mysql 约定拼接 DSN。
// 形如：user:pass@tcp(host:port)/db?charset=utf8mb4&parseTime=true&loc=Local
func buildMySQLDSN(conn map[string]interface{}) string {
	user := strOr(conn, "username", "root")
	pass := strOr(conn, "password", "")
	host := strOr(conn, "host", "127.0.0.1")
	port := intOr(conn, "port", 3306)
	db := strOr(conn, "database", "")
	charset := strOr(conn, "charset", "utf8mb4")

	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		user, pass, host, port, db, charset)
}

// buildPostgresDSN 使用空格分隔的 libpq 格式。
// PostgreSQL 的 timezone 走 Local，便于与应用的 Asia/Shanghai 对齐（由业务层校正时区）。
func buildPostgresDSN(conn map[string]interface{}) string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		strOr(conn, "host", "127.0.0.1"),
		intOr(conn, "port", 5432),
		strOr(conn, "username", "postgres"),
		strOr(conn, "password", ""),
		strOr(conn, "database", ""),
		strOr(conn, "sslmode", "disable"),
	)
}

// resolveSQLitePath 确保 SQLite 文件所在目录存在；不存在时自动创建。
// 这避免了用户第一次启动就崩在 "unable to open database file"。
func resolveSQLitePath(conn map[string]interface{}) string {
	path := strOr(conn, "database", "thinkgin.db")
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	return path
}

// applyPool 读取 pool.max_open / pool.max_idle / pool.max_lifetime 并应用。
// 任一字段缺失时使用 GORM/driver 默认值，不做强制要求。
func applyPool(db *gorm.DB, conn map[string]interface{}) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	pool, ok := conn["pool"].(map[string]interface{})
	if !ok {
		return nil
	}
	if v, ok := toInt(pool["max_open"]); ok {
		sqlDB.SetMaxOpenConns(v)
	}
	if v, ok := toInt(pool["max_idle"]); ok {
		sqlDB.SetMaxIdleConns(v)
	}
	if v, ok := toInt(pool["max_lifetime"]); ok {
		sqlDB.SetConnMaxLifetime(time.Duration(v) * time.Second)
	}
	return nil
}

// 以下是小工具函数，纯粹为了把 map[string]interface{} 安全转换为需要的类型。
// 因为 YAML 解析后的数值类型可能是 int / int64 / float64，需要统一归一。

func strOr(m map[string]interface{}, key string, fallback string) string {
	if v, ok := m[key].(string); ok && v != "" {
		return v
	}
	return fallback
}

func intOr(m map[string]interface{}, key string, fallback int) int {
	if v, ok := toInt(m[key]); ok {
		return v
	}
	return fallback
}

func toInt(v interface{}) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case float64:
		return int(n), true
	default:
		return 0, false
	}
}
