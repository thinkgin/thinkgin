# 从 ThinkPHP 迁移到 ThinkGin

> 本文面向有 ThinkPHP 6.x / 8.x 经验的开发者。每一小节左边是 ThinkPHP 的做法，
> 右边是 ThinkGin 对应的实现。阅读顺序不重要，按需查阅即可。

---

## 零、心智转变

| 维度 | ThinkPHP | ThinkGin |
|------|----------|----------|
| 语言 | PHP 动态类型 | Go 静态类型 |
| 请求周期 | 每请求一次脚本 | 长驻进程 |
| 路由 | 文件名约定 + 注解 | 代码注册（`route.RegisterXxx`） |
| 依赖 | Composer autoload | `go.mod` 显式导入 |
| 错误处理 | try/catch + Exception | `error` 返回值 + `defer/recover` |
| 并发 | 无内置，靠多进程 | goroutine + context |

**最容易踩的坑**：不要再指望"每请求一次"的脚本心智。ThinkGin 所有包级变量都是**进程级单例**，
生命周期与进程同长。使用可变全局状态前一定要加锁或用 `sync.Once` 保护。

---

## 一、目录结构对照

### ThinkPHP 6

```
app/
├── controller/
│   └── Index.php
├── model/
│   └── User.php
├── middleware.php           # 全局中间件定义
└── route/
    └── app.php              # 路由
config/
└── database.php
```

### ThinkGin

```
app/
├── index/
│   ├── controller/
│   │   └── hello.go
│   ├── model/
│   │   └── user.go
│   └── view/
│       └── index.html
├── database/                # GORM 多数据源（不是模块）
└── cache/                   # Redis（不是模块）
config/
└── database.yaml
route/
├── web.go
└── api.go
extend/middleware/
├── cors.go
└── ratelimit.go
```

**关键差异**：ThinkGin 按"业务模块"划分（`app/index/`），每个模块自成 MVC；
`app/database/` 和 `app/cache/` 是横切基础设施，不是业务模块。

---

## 二、配置系统

### 读配置

```php
// ThinkPHP
$port = config('server.port');
$host = env('DB_HOST', '127.0.0.1');
```

```go
// ThinkGin
cfg := app.GetConfig()
port := cfg.Server.HTTP.Port
// 环境变量覆盖由框架自动完成：THINKGIN_SERVER_HTTP_PORT=9090
```

### 环境变量

```
# ThinkPHP .env
DB_HOST=127.0.0.1
DB_PORT=3306
```

```bash
# ThinkGin：THINKGIN_ 前缀覆盖 YAML
export THINKGIN_SERVER_MODE=release
export THINKGIN_SERVER_HTTP_PORT=9090
export THINKGIN_LOG_LEVEL=warn
```

---

## 三、路由

### ThinkPHP

```php
// route/app.php
Route::get('user/:id', 'User/read');
Route::group('api', function () {
    Route::get('users', 'api.User/index');
})->middleware(['auth']);
```

### ThinkGin

```go
// route/api.go
func RegisterAPIRoutes(r *gin.Engine) {
    api := r.Group("/api/v1")
    api.Use(middleware.JWTAuth())

    api.GET("/users/:id", controller.UserRead)
    api.GET("/users", controller.UserList)
}
```

**差异**：
- 没有"路由注解"——路由注册是普通代码
- 路径参数语法 `:id` 与 ThinkPHP 一致
- 分组中间件通过 `Group.Use()` 附加

---

## 四、Controller

### ThinkPHP

```php
class User extends BaseController
{
    public function read($id)
    {
        return json(['id' => $id]);
    }
}
```

### ThinkGin

```go
func UserRead(c *gin.Context) {
    id := c.Param("id")
    c.JSON(200, gin.H{"id": id})
}
```

**差异**：
- 没有"控制器类" —— 就是普通函数，接收 `*gin.Context`
- 参数提取用 `c.Param / c.Query / c.PostForm / c.ShouldBindJSON`
- 返回值不由 `return` 决定，而是显式调用 `c.JSON / c.HTML / c.Abort`

---

## 五、Model

### ThinkPHP

```php
class User extends Model
{
    protected $table = 'users';
    protected $schema = [
        'id'   => 'int',
        'name' => 'string',
    ];
}

// 使用
$u = User::find(1);
$u->save();
User::where('status', 1)->select();
```

### ThinkGin

```go
// app/index/model/user.go
type User struct {
    gorm.Model                        // 内嵌得到 ID/CreatedAt/UpdatedAt/DeletedAt
    Name  string `gorm:"size:64"`
    Email string `gorm:"size:128;uniqueIndex"`
}

func (User) TableName() string { return "users" }

// 使用
var u User
db := database.Default().WithContext(ctx)
db.First(&u, 1)
db.Save(&u)
db.Where("status = ?", 1).Find(&users)
```

**差异**：
- Go 没有 `static` 方法，所有 CRUD 走 `*gorm.DB` 实例
- 字段用 tag 描述（`gorm:"size:64;uniqueIndex"`），对应 ThinkPHP 的 `$schema`
- **强烈建议**始终传 `WithContext(ctx)`，让 SQL 纳入 OTel trace

---

## 六、中间件

### ThinkPHP

```php
namespace app\middleware;
class Check
{
    public function handle($request, \Closure $next)
    {
        if (!$request->isPost()) return redirect('/');
        return $next($request);
    }
}
```

### ThinkGin

```go
func Check() gin.HandlerFunc {
    return func(c *gin.Context) {
        if c.Request.Method != http.MethodPost {
            c.Redirect(302, "/")
            c.Abort()
            return
        }
        c.Next()
    }
}

// 注册
r.Use(Check())
```

**差异**：
- 没有"Closure 链"，Gin 用 `c.Next()` 显式传递
- 想中断处理链必须 `c.Abort()`，否则后续中间件仍会执行

---

## 七、数据库连接

### ThinkPHP

```php
// config/database.php
'mysql' => [
    'type'     => 'mysql',
    'hostname' => '127.0.0.1',
    'database' => 'thinkgin',
    'username' => 'root',
],

// 使用
Db::name('users')->select();
Db::connect('pgsql')->name('logs')->select();
```

### ThinkGin

```yaml
# config/database.yaml（已有）
database:
  default: mysql
  connections:
    mysql:
      driver: mysql
      host: 127.0.0.1
      database: thinkgin
    pgsql:
      driver: postgres
      host: 127.0.0.1
```

```go
// 使用
db := database.Default()
database.Default().Table("users").Find(&users)
pg, _ := database.Get("pgsql")
pg.Table("logs").Find(&logs)
```

---

## 八、缓存

### ThinkPHP

```php
Cache::set('key', 'value', 600);
$v = Cache::get('key');
Cache::delete('key');
```

### ThinkGin

```go
import "thinkgin/app/cache"

rdb := cache.Default()  // *redis.Client
rdb.Set(ctx, "key", "value", 10*time.Minute)
v, _ := rdb.Get(ctx, "key").Result()
rdb.Del(ctx, "key")
```

**差异**：Go 没有"静态 Facade"，直接用 `*redis.Client` 方法。

---

## 八点二、JWT 鉴权

### ThinkPHP

通常借助 `firebase/php-jwt` 或第三方包自行封装：

```php
$token = JWT::encode(['uid' => $user->id], $secret, 'HS256');
$claims = JWT::decode($token, new Key($secret, 'HS256'));
```

### ThinkGin

```go
import "thinkgin/extend/middleware"

// 签发
token, err := middleware.JWTIssue("42", map[string]any{"role": "admin"}, 0) // 0 = 使用 app.jwt.expire

// 保护路由：middleware.yaml 里 global 加 "jwt"，或分组 Use
api := r.Group("/api/v1", middleware.JWTAuth())
api.GET("/me", func(c *gin.Context) {
    claims := middleware.JWTFromContext(c)
    c.JSON(200, gin.H{"uid": claims.Subject, "role": claims.Extra["role"]})
})
```

**差异**：
- 密钥读 `config/app.yaml` 的 `app.jwt.secret`，为空时拒绝签发/校验。
- 算法固定 HS256，显式拒绝 `alg=none`、非 HMAC 攻击。
- 过期、签名非法、格式错误都统一走 `APIError(401)`，响应体对外隐藏内部细节。

---

## 八点五、Session

### ThinkPHP

```php
session('user_id', 42);
$id = session('user_id');
session('user_id', null); // 删除
```

### ThinkGin

```go
import "thinkgin/app/session"

// 在 config/middleware.yaml 的 global 列表里加上 "session"，或在某个 group 里 Use。
// 然后业务代码：
func Profile(c *gin.Context) {
    s := session.From(c)
    s.Set("user_id", 42)
    _ = s.Save(c.Request.Context())

    if v, ok := s.Get("user_id"); ok {
        c.JSON(200, gin.H{"uid": v})
    }
}

func Logout(c *gin.Context) {
    _ = session.From(c).Destroy(c.Request.Context())
    session.ClearCookie(c)
}
```

**差异**：
- Save 必须显式调用。框架不自动保存，避免遮蔽业务错误处理路径。
- `config/session.yaml` 支持 `memory` 与 `redis` 两种 driver；`file` / `database` 尚未实现。
- Cookie 的 `SameSite` / `HttpOnly` / `Secure` 标志与配置一一对应，默认就是安全取向。

---

## 九、视图

### ThinkPHP

```php
// Controller
return view('index', ['name' => 'Tom']);

// template
<h1>{{ $name }}</h1>
```

### ThinkGin

```go
// Controller
c.HTML(200, "index.html", gin.H{"name": "Tom"})
```

```html
<!-- app/index/view/index.html -->
<h1>{{ .name }}</h1>
```

**差异**：
- 模板引擎是 Go 标准库 `html/template`，语法 `{{ .var }}` 而非 `{{ $var }}`
- 默认启用 XSS 防护（HTML escape），输出原始 HTML 要 `{{ .html | safehtml }}`

---

## 十、JSON 请求 / 响应

### ThinkPHP

```php
$data = $request->post();
return json(['code' => 0, 'data' => $data]);
```

### ThinkGin

```go
type CreateReq struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

func Create(c *gin.Context) {
    var req CreateReq
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(400, gin.H{"code": 400, "msg": err.Error()})
        return
    }
    c.JSON(200, gin.H{"code": 0, "data": req})
}
```

**差异**：
- 请求体必须**先定义结构体**，参数校验通过 tag 声明
- 推荐用 `extend/middleware/api_response.go` 里的统一响应封装

---

## 十一、日志

| ThinkPHP | ThinkGin |
|----------|----------|
| `Log::info('msg')` | `app.GetLogger().Info("msg")` |
| `Log::error('e', $context)` | `app.GetLogger().WithFields(logrus.Fields{...}).Error("e")` |

---

## 十二、异常 / 错误

```php
// ThinkPHP
try {
    doSomething();
} catch (\Exception $e) {
    Log::error($e);
    return json_error($e->getMessage());
}
```

```go
// ThinkGin
if err := doSomething(); err != nil {
    app.GetLogger().WithError(err).Error("...")
    c.JSON(500, gin.H{"code": 500, "msg": err.Error()})
    return
}
```

Recovery 中间件会把 panic 兜底转换为 500，无需业务自己 `defer recover()`。

---

## 十三、常见坑

1. **Go 的 `map` 不是并发安全**。多 goroutine 读写共享 map 必须加锁或用 `sync.Map`。
2. **别把 `*gin.Context` 跨 goroutine 传**。要脱离请求生命周期就 `c.Copy()`。
3. **`time.Time` 的零值不是 NULL**。需要可空字段用 `*time.Time`。
4. **环境变量改不动配置？** 确认已经 Bootstrap 完成，且用的是 `THINKGIN_` 前缀。
5. **CI 挂了但本地通过？** 大概率是 `go vet` 或 `-race` 发现的问题，本地补 `CGO_ENABLED=1 go test -race ./...` 复现。

---

## 十四、还在路上的能力

以下 ThinkPHP 用户可能期待的能力**当前未实现**，计划在后续版本逐步补齐：

- [ ] Seeder / Migration 工具（当前只能 GORM `AutoMigrate`）
- [ ] 多环境配置（`config/dev/`、`config/prod/`）
- [ ] Session `file` / `database` driver（`memory` / `redis` 已可用）
- [ ] 多语言 i18n 的运行时实现

已落地：

- [x] 命令行脚手架：`go run ./cmd/scaffold new module user` 生成 MVC 骨架
- [x] Session 运行时：`memory` / `redis` driver、Gin 中间件、安全 Cookie

如果某项阻塞了你的迁移，欢迎提 Issue。
