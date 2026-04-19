// Package database 负责从 config/database.yaml 初始化 GORM 多数据源。
//
// 设计目标：
//   - "零配置可跑"：启动即使没有任何 connection 也不会 panic，失败只记录错误。
//   - 多连接：支持在 connections 映射里声明任意多个命名连接。
//   - 驱动解耦：通过 driver 字段分发到具体 dialector（mysql/postgres/sqlite）。
//   - 可关闭：main 可在优雅停机阶段调用 CloseAll 释放连接池。
//
// 典型调用：
//
//	if err := database.Init(); err != nil { ... }
//	defer database.CloseAll()
//	db := database.Default()      // *gorm.DB
//	db2 := database.Get("pgsql")  // *gorm.DB
package database

import (
	"errors"
	"fmt"
	"sync"

	"gorm.io/gorm"

	"thinkgin/app"
)

// ErrNotFound 表示请求的命名连接不存在。
var ErrNotFound = errors.New("database: connection not found")

// dbs 持有所有已成功建立的连接；写保护由 mu 负责。
var (
	mu      sync.RWMutex
	dbs     = map[string]*gorm.DB{}
	defName string // 默认连接名，来自 config.database.default
)

// Init 根据 app.GetConfig().Database 构造所有非 Redis 的 SQL 连接。
// 支持的 driver：mysql / postgres / sqlite / sqlite3。
// Redis 条目（driver=redis）在此被明确忽略，应由 cache 包接管（暂未实现）。
//
// 返回的 error 聚合了所有失败的连接，但不会阻止其他连接初始化。
// 即使全部连接都失败，Init 也会返回错误但进程仍可启动（便于开发期本地无数据库场景）。
func Init() error {
	cfg := app.GetConfig()
	if cfg == nil {
		return errors.New("database: global config is nil, did you forget Bootstrap?")
	}

	mu.Lock()
	defer mu.Unlock()
	defName = cfg.Database.Default

	var errs []error
	for name, raw := range cfg.Database.Connections {
		conn, ok := raw.(map[string]interface{})
		if !ok {
			errs = append(errs, fmt.Errorf("%s: invalid connection structure", name))
			continue
		}

		driver, _ := conn["driver"].(string)
		if driver == "redis" {
			// Redis 不属于本包职责，静默跳过。
			continue
		}

		db, err := openConnection(driver, conn)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s(%s): %w", name, driver, err))
			continue
		}
		dbs[name] = db
	}

	return errors.Join(errs...)
}

// Get 按名称返回已注册的数据库连接。
// 名称不存在时返回 ErrNotFound，调用方应当判空或判错。
func Get(name string) (*gorm.DB, error) {
	mu.RLock()
	defer mu.RUnlock()
	db, ok := dbs[name]
	if !ok {
		return nil, fmt.Errorf("%w: %q", ErrNotFound, name)
	}
	return db, nil
}

// MustGet 同 Get，不存在时 panic。用于启动期可预知必然存在的连接。
func MustGet(name string) *gorm.DB {
	db, err := Get(name)
	if err != nil {
		panic(err)
	}
	return db
}

// Default 返回配置里 default 指向的连接；不存在时返回 nil。
// 这是为了方便"90% 业务只用一个数据库"的常见场景。
func Default() *gorm.DB {
	mu.RLock()
	defer mu.RUnlock()
	return dbs[defName]
}

// CloseAll 关闭所有已建立的连接。幂等：重复调用不会出错。
// 适合在 main 的 defer / signal 监听中调用。
func CloseAll() error {
	mu.Lock()
	defer mu.Unlock()

	var errs []error
	for name, db := range dbs {
		sqlDB, err := db.DB()
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: get underlying *sql.DB: %w", name, err))
			continue
		}
		if err := sqlDB.Close(); err != nil {
			errs = append(errs, fmt.Errorf("%s: close: %w", name, err))
		}
	}
	dbs = map[string]*gorm.DB{}
	return errors.Join(errs...)
}
