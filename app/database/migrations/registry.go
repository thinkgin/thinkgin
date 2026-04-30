// Package migrations 集中注册所有数据库迁移。
//
// 每个迁移文件以 YYYYMMDDHHMMSS_描述.go 命名，在 init() 中调用 Register()。
// 统一由 cmd/migrate.go 导入此包以触发注册。
//
// 示例迁移文件参见 20260430000000_create_users_table.go。
package migrations

import "thinkgin/app/database"

var registered []database.Migration

// Register 注册一个迁移。迁移文件的 init() 中调用。
func Register(m database.Migration) {
	registered = append(registered, m)
}

// All 返回所有已注册的迁移。
func All() []database.Migration {
	out := make([]database.Migration, len(registered))
	copy(out, registered)
	return out
}
