# 数据库迁移指南（Goose）

ThinkGin **不内置**数据库迁移工具。本页给出推荐做法。

## 为什么不用 GORM 的 `AutoMigrate`

`db.AutoMigrate(&User{})` 适合**开发期快速试错**，但不能上生产：

- **无法删字段**：删了结构体里的字段，数据库列照样留着
- **无法改列类型**：类型不兼容时直接报错，无回退方案
- **没有版本号**：不知道哪台实例跑过哪一版，回滚全凭人脑
- **无法做数据迁移**：只能改 schema，不能顺便 `UPDATE` 旧数据

这些能力**不是框架该解决的**，已有成熟工具。

---

## 推荐工具：pressly/goose

[`github.com/pressly/goose/v3`](https://github.com/pressly/goose) 是目前 Go 社区使用最广的迁移工具，支持 SQL 和 Go 两种迁移文件格式。

### 安装

```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

### 目录约定

在项目根目录放一个 `db/migrations/` 存放迁移文件：

```
thinkgin/
├── db/
│   └── migrations/
│       ├── 00001_create_users_table.sql
│       ├── 00002_add_users_email_index.sql
│       └── ...
└── config/
    └── database.yaml
```

### 编写一个 SQL 迁移

```sql
-- db/migrations/00001_create_users_table.sql
-- +goose Up
CREATE TABLE users (
    id         BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    email      VARCHAR(128) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +goose Down
DROP TABLE users;
```

### 执行

```bash
# 升级到最新
goose -dir db/migrations mysql "thinkgin:thinkgin@tcp(127.0.0.1:3306)/thinkgin?parseTime=true" up

# 回滚一步
goose -dir db/migrations mysql "thinkgin:thinkgin@tcp(127.0.0.1:3306)/thinkgin?parseTime=true" down

# 查看当前版本
goose -dir db/migrations mysql "thinkgin:thinkgin@tcp(127.0.0.1:3306)/thinkgin?parseTime=true" status
```

### 新建一个迁移文件

```bash
goose -dir db/migrations create add_users_email_index sql
# 会生成 db/migrations/<timestamp>_add_users_email_index.sql
```

---

## 把迁移跑进 CI / 部署流水线

生产部署通常分两步：

1. **先迁移**：部署前用 goose 把 schema 改到新版本（向后兼容的改动）
2. **再发代码**：新代码读写新 schema

```yaml
# .github/workflows/deploy.yml 节选
- name: Migrate database
  run: goose -dir db/migrations mysql "${{ secrets.DB_DSN }}" up

- name: Deploy app
  run: kubectl rollout restart deploy/thinkgin
```

---

## 替代方案

如果你有其他偏好，这两个也经过生产验证：

- [`golang-migrate/migrate`](https://github.com/golang-migrate/migrate)：纯 CLI，多数据库支持更广
- [`sqlc`](https://github.com/sqlc-dev/sqlc) + 上述迁移工具：编译期生成类型安全的 Go 代码

ThinkGin 不与任一工具绑定 —— 选你团队最顺手的即可。

---

## 开发期快捷做法

本地起步或者写测试时，直接 `AutoMigrate` 确实方便：

```go
import "thinkgin/app/database"

func init() {
    if db := database.Default(); db != nil {
        _ = db.AutoMigrate(&model.User{}, &model.Post{})
    }
}
```

但务必**只在 dev 环境执行**。生产 env 应当：

- 把 `AutoMigrate` 块用 `if cfg.App.Debug` 包住
- 或者干脆删掉，全部走 goose
