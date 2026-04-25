# FAQ · 设计决策与常见疑问

本文档收录 ThinkGin 在演进过程中反复被问到、而文档里又没地方专门讲清楚的
问题。每条都尽量把"**当前实现**"和"**为什么这样**"一起写明，避免只看代码
推测设计意图。

---

## 1. 为什么没有再封装一层"类 ThinkPHP 的 Model 链式调用"？

### 当前实现

框架**直接暴露** `*gorm.DB`，业务层用 GORM 原生 API 做链式查询。既没有
`UserModel::where()->find()` 那种门面，也没有 `Db::name('user')` 这种 facade。

### 为什么不封装

**"连接污染"是 ThinkPHP 的历史债，不是 Go 世界的问题。**

| 维度 | ThinkPHP | GORM / ThinkGin |
|------|----------|----------------|
| 运行模型 | PHP 单请求进程，请求结束销毁 | 长驻进程，goroutine 并发 |
| 链式状态 | 静态属性挂条件 → 下次调用残留 | 每次 `.Where()` / `.Model()` 返回**新** `*gorm.DB` 实例 |
| 隔离手段 | 手动 `Db::table()->newQuery()` | 天然 immutable；`db.Session()` 可显式切隔离域 |

GORM 的链式本来就是 ThinkPHP 想要的那种：可读写分离、可注入 Context、可
自动携带事务——**而且没有污染**：

```go
// 每个 handler 一条独立链，互不干扰
users := []User{}
database.Default().
    WithContext(c.Request.Context()).   // 注入请求 ctx
    Where("status = ?", 1).
    Order("id desc").
    Limit(10).
    Find(&users)
```

如果想要 "用模型类代替 table 名" 的表达，GORM 原生就支持：

```go
type User struct {
    ID   int
    Name string
}
func (User) TableName() string { return "tg_users" }

var u User
database.Default().Where("name = ?", "foo").First(&u)
```

再封装一层反而容易**引入**问题：
- 很多 PHP 习惯的封装会把 Builder 状态挂到单例上，在 Go 里立刻变成 goroutine
  共享状态，必须加锁，性能反而下降。
- 屏蔽 GORM 原生能力（Preload、Joins、Scopes、Clauses）后，业务一旦遇到
  复杂查询，又得绕回 `*gorm.DB`，封装层反而成了双重心智负担。

### 什么时候自己再封装

只有当业务出现**高度重复**的查询模式（例如"每个模型都要软删除 + 按租户隔离 +
审计字段"）时，才值得用 GORM 的 `Scopes` / 插件机制抽象出来，而不是在
controller 里反复重写同一段 Where。这属于**业务沉淀**，不是框架义务。

---

## 2. 为什么 Redis 的连接定义放在 database.yaml，而不是 cache.yaml？

### 当前实现

- **连接定义**集中在 `database.yaml.connections` 下（含 MySQL / PostgreSQL /
  SQLite / Redis）
- **功能模块**（`cache.yaml` / `session.yaml` / 未来的限流器）通过
  `connection: "redis"` **按名引用**那个连接

```yaml
# config/database.yaml
database:
  connections:
    mysql: { driver: mysql, ... }
    redis: { driver: redis, host: 127.0.0.1, port: 6379 }   # 只定义一次

# config/cache.yaml
cache:
  stores:
    redis:
      driver: redis
      connection: "redis"     # 按名引用

# config/session.yaml
session:
  driver: redis
  redis:
    connection: "redis"       # 按名引用
```

### 为什么不搬到 cache.yaml

确实有人觉得"Redis 不是数据库，放 database.yaml 语义别扭"。但把它搬到
`cache.yaml` 会引入新的问题：

| 方案 | 优点 | 缺点 |
|------|------|------|
| **A. 当前（database.yaml 统一持有）** | 单一事实源；多用途共享同一连接；避免 host/port/password 在三个 YAML 里重复 | 语义上 Redis 不算"数据库"，有人困惑 |
| **B. 搬到 cache.yaml** | cache 域自包含，看 cache 模块不用跨文件 | Session 也用 Redis 时要么再写一份、要么让 `session.yaml` 反向引用 `cache.yaml`——形成"会话依赖缓存"的不自然耦合 |

现实里 Redis 几乎都是**跨用途共享资源**：Cache + Session + 限流 + pub/sub
常连同一个实例。集中声明、按名引用符合 DRY 原则；如果搬到 cache 再复制到
session，只是把"database 欠缺清晰度"换成"cache / session 各自欠缺清晰度"。

### 思考方向：未来可能的演进

如果你确实觉得 "database 里混着 redis" 不舒服，另一种可能的设计是**新增
`redis.yaml` 专门放 Redis 连接**，让语义完全独立于 SQL 数据库。这是个
合理演进方向，但会是一次破坏性改动，不在 v3.x 计划里。

---

## 3. 为什么默认配置不预置 MySQL / Redis 连接？

### 当前实现

`database.yaml.connections` 和 `cache.yaml.stores` 默认都是 `{}`（空），
真实连接示例全部塞在注释里供复制。

### 为什么

**零依赖启动**是 3.3.4 的一个显性目标：

```bash
git clone ... && cd thinkgin && go run main.go
# 立即得到一个能响应 /livez、/readyz、/metrics 的 HTTP 服务
```

预置连接会造成两类恶劣体验：

1. **本地连不上 MySQL**：进程启动时一堆 `database ping failed` 告警，新手
   以为框架有 bug。
2. **测试污染**：`go test` 在没有真实 MySQL 的 CI 环境里也会报告错误。

代价是：需要用数据库的时候，多花 30 秒从 YAML 注释里拷出来改一改。这个代价
远小于"新手第一次接触就看到一屏错误日志"的代价。

---

## 4. 为什么 `view.yaml` / `filesystem.yaml` / `lang.yaml` 里几乎是空的？

### 当前实现

三个 YAML 都只剩 `view: {}` / `filesystem: {}` / `lang: {}` 加一段注释，
说明"当前未实现运行时逻辑"。

### 为什么

**不想再留"配置了也没效果"的陷阱。**

之前版本这些 YAML 里写满了 `driver / paths / compile / assets / helpers`
等字段，但框架运行时根本不消费它们——模板路径硬编码为 `app/*/view/*.html`，
filesystem 抽象层完全没实现。保留这些字段反而让用户误以为"改了就会生效"，
是典型的**纸面配置陷阱**。

3.3.4 做的事就是**把未消费字段全部删掉**，只保留空骨架 + 注释说明替代方案：

- 模板：继续走 `html/template` + `app/*/view/*.html` 硬约定
- 文件存储：推荐直接用官方 SDK（aliyun-oss-go-sdk / aws-sdk-go-v2 等）
- 多语言：推荐 `nicksnyder/go-i18n` 或 `golang.org/x/text/message` 自行集成

等未来真正实现这些模块时，再往 YAML 里加字段——**"先实现，再暴露"**。

---

## 5. 为什么 `init()` 里还保留了自动加载配置？

### 当前实现

```go
// app/bootstrap.go
func init() {
    bootstrapOnce.Do(func() {
        if _, err := os.Stat(defaultConfigDir); os.IsNotExist(err) {
            setDefaultConfig()
            InitLogger()
            return
        }
        if err := Bootstrap(defaultConfigDir); err != nil { ... }
    })
}
```

### 为什么不完全显式化

纯粹论"工程洁癖"，`init()` 应该尽量干净。但测试下来，完全去掉 `init()` 后
会出现这种情况：

```go
// 业务代码
cfg := app.GetConfig()   // 此时 cfg 可能还是零值
```

用户必须**记得**在 `main` 里先调 `Bootstrap`，否则就是一段运行时 panic 或
看起来诡异的零值行为。对"开箱即用"的定位而言，代价太大。

当前折中：
- `init()` 仅做兜底，且 `sync.Once` 保证只执行一次
- 工作目录**没有** `config/` 时静默退回 defaults + logger，不污染 `go test`
  输出
- 显式场景（自定义配置路径、测试注入）调用 `Bootstrap(customDir)` 即可

---

如果还有其他反复被问到的设计决策，欢迎补充到这份文档——一次写清楚，省掉
未来每次都重复解释。
