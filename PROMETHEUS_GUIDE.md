# ThinkGin 普罗米修斯监控服务完全指南

## 📋 目录

1. [什么是 Prometheus](#什么是-prometheus)
2. [为什么需要监控](#为什么需要监控)
3. [快速开始](#快速开始)
4. [详细配置说明](#详细配置说明)
5. [如何使用监控功能](#如何使用监控功能)
6. [设置 Prometheus 服务器](#设置-prometheus-服务器)
7. [集成 Grafana 仪表板](#集成-grafana-仪表板)
8. [业务代码中的监控](#业务代码中的监控)
9. [常用监控查询](#常用监控查询)
10. [故障排除](#故障排除)
11. [最佳实践](#最佳实践)

---

## 🎯 什么是 Prometheus

Prometheus 是一个开源的系统监控和告警工具，最初由 SoundCloud 开发。它现在是 Cloud Native Computing Foundation（CNCF）的毕业项目。

### 核心概念

- **指标（Metrics）**: 数值数据，如请求数量、响应时间、内存使用量
- **时间序列**: 带有时间戳的指标数据点序列
- **标签（Labels）**: 用于标识和分组指标的键值对
- **抓取（Scrape）**: Prometheus 从目标服务收集指标的过程

### 四种指标类型

1. **Counter（计数器）**: 只能递增的累计指标

   - 示例：HTTP 请求总数、错误总数

2. **Gauge（仪表盘）**: 可以任意上下变动的指标

   - 示例：当前内存使用量、在线用户数

3. **Histogram（直方图）**: 观察值的分布情况

   - 示例：请求响应时间分布

4. **Summary（摘要）**: 类似直方图，但计算的是分位数
   - 示例：响应时间的 50%、90%、99% 分位数

---

## 🚀 为什么需要监控

监控系统可以帮助您：

- **及时发现问题**: 在用户发现之前就知道系统异常
- **性能优化**: 找出系统瓶颈，优化性能
- **容量规划**: 了解资源使用趋势，合理规划容量
- **故障分析**: 快速定位问题根因
- **业务洞察**: 了解用户行为和业务趋势

---

## ⚡ 快速开始

### 步骤 1: 检查配置文件

确保以下配置文件存在且配置正确：

**`config/app.yaml`** - 总开关配置

```yaml
app:
  monitoring:
    # 这是总开关，必须设置为 true
    prometheus_enabled: true
```

**`config/prometheus.yaml`** - 详细配置

```yaml
prometheus:
  # 启用 Prometheus 监控
  enabled: true
  # 监控端点路径
  path: "/metrics"
  # 其他配置...
```

### 步骤 2: 启动应用

```bash
# 在项目根目录下执行
go run main.go
```

### 步骤 3: 验证监控是否工作

打开浏览器或使用 curl 访问：

```bash
# 检查监控端点
curl http://localhost:8000/metrics

# 如果看到类似下面的输出，说明监控正常工作
# HELP app_http_requests_total Total number of HTTP requests
# TYPE app_http_requests_total counter
# app_http_requests_total{method="GET",path="api",status="200"} 1
```

---

## ⚙️ 详细配置说明

### 总开关配置（config/app.yaml）

```yaml
app:
  monitoring:
    # 普罗米修斯监控总开关
    # true: 启用监控功能
    # false: 完全禁用监控功能（推荐用于生产环境临时关闭）
    prometheus_enabled: true
```

### 详细监控配置（config/prometheus.yaml）

```yaml
prometheus:
  # 是否启用 Prometheus 监控
  enabled: true

  # 监控指标暴露路径（URL路径）
  path: "/metrics"

  # 监控端口（0表示使用主服务端口）
  port: 0

  # 监控服务名称（用于标识不同的服务）
  service_name: "thinkgin"

  # 监控命名空间（用于组织指标）
  namespace: "app"

  # 自定义标签（会添加到所有指标上）
  labels:
    environment: "development" # 环境标识
    version: "2.0" # 版本号
    team: "backend" # 团队标识

  # 监控指标配置
  metrics:
    # HTTP请求相关指标
    http:
      enabled: true
      requests_total: "http_requests_total" # 请求总数指标名
      request_duration: "http_request_duration_seconds" # 请求耗时指标名
      request_size: "http_request_size_bytes" # 请求大小指标名
      response_size: "http_response_size_bytes" # 响应大小指标名
      include_path: false # 是否在指标中包含完整URL路径（可能产生大量标签）

    # 系统资源指标
    system:
      enabled: true
      cpu_usage: "system_cpu_usage_percent" # CPU使用率指标名
      memory_usage: "system_memory_usage_bytes" # 内存使用量指标名
      process_start_time: "process_start_time_seconds" # 进程启动时间

    # 业务指标
    business:
      enabled: true
      counter_prefix: "business_counter" # 业务计数器前缀
      histogram_prefix: "business_histogram" # 业务直方图前缀
      gauge_prefix: "business_gauge" # 业务仪表盘前缀

  # 监控数据采集间隔（秒）
  scrape_interval: 15

  # 是否启用基础认证（保护监控端点）
  auth:
    enabled: false # 是否启用认证
    username: "prometheus" # 用户名
    password: "secure_password" # 密码
```

### 配置优先级

监控功能需要同时满足以下条件才会启用：

1. **总开关**: `app.yaml` 中的 `monitoring.prometheus_enabled` = `true`
2. **功能开关**: `prometheus.yaml` 中的 `prometheus.enabled` = `true`

这种设计的优势：

- 可以快速关闭整个监控系统（修改总开关）
- 可以保留详细配置的同时临时禁用
- 适合不同环境的配置管理

---

## 🔧 如何使用监控功能

### 开启监控

1. **修改总开关**

   ```yaml
   # config/app.yaml
   app:
     monitoring:
       prometheus_enabled: true # 设置为 true
   ```

2. **确认详细配置**

   ```yaml
   # config/prometheus.yaml
   prometheus:
     enabled: true # 确保为 true
   ```

3. **重启应用**

   ```bash
   # 停止当前应用（如果正在运行）
   # 然后重新启动
   go run main.go
   ```

4. **验证监控**
   ```bash
   # 访问监控端点
   curl http://localhost:8000/metrics
   ```

### 关闭监控

#### 方法 1: 使用总开关（推荐）

```yaml
# config/app.yaml
app:
  monitoring:
    prometheus_enabled: false # 设置为 false
```

#### 方法 2: 关闭具体功能

```yaml
# config/prometheus.yaml
prometheus:
  enabled: false # 设置为 false
```

#### 两种方法的区别

- **方法 1**: 完全禁用，不会初始化任何监控组件，性能影响最小
- **方法 2**: 只是不暴露端点，但监控组件仍会初始化

### 测试监控是否生效

1. **检查监控端点**

   ```bash
   curl http://localhost:8000/metrics
   ```

   - 如果监控开启：返回大量指标数据
   - 如果监控关闭：返回 404 错误

2. **检查应用日志**

   ```bash
   # 启动应用时查看日志
   go run main.go

   # 如果看到类似输出，说明监控已启用
   # {"level":"info","msg":"🚀 服务器启动在: http://0.0.0.0:8000"}
   ```

3. **生成测试数据**

   ```bash
   # 访问几个页面生成监控数据
   curl http://localhost:8000/
   curl http://localhost:8000/index/hello

   # 再次检查监控端点
   curl http://localhost:8000/metrics | grep http_requests_total
   ```

---

## 🖥️ 设置 Prometheus 服务器

### 下载和安装

#### Windows 系统

1. **下载 Prometheus**

   ```bash
   # 访问官网下载
   # https://prometheus.io/download/
   # 选择 Windows 版本下载
   ```

2. **解压并配置**
   ```bash
   # 解压下载的文件
   # 进入解压目录
   cd prometheus-*
   ```

#### Linux/macOS 系统

```bash
# 下载
wget https://github.com/prometheus/prometheus/releases/download/v2.45.0/prometheus-2.45.0.linux-amd64.tar.gz

# 解压
tar xvfz prometheus-*.tar.gz
cd prometheus-*
```

### 配置 Prometheus

创建 `prometheus.yml` 配置文件：

```yaml
# 全局配置
global:
  # 抓取间隔
  scrape_interval: 15s
  # 评估规则间隔
  evaluation_interval: 15s

# 告警规则文件
rule_files:
  # - "first_rules.yml"
  # - "second_rules.yml"

# 抓取配置
scrape_configs:
  # Prometheus 自己的监控
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:9090"]

  # ThinkGin 应用监控
  - job_name: "thinkgin"
    static_configs:
      - targets: ["localhost:8000"] # 您的应用地址
    metrics_path: "/metrics" # 监控端点路径
    scrape_interval: 15s # 抓取间隔


    # 如果启用了认证，添加认证信息
    # basic_auth:
    #   username: 'prometheus'
    #   password: 'secure_password'
```

### 启动 Prometheus

```bash
# 启动 Prometheus 服务器
./prometheus --config.file=prometheus.yml

# 或者在后台运行
nohup ./prometheus --config.file=prometheus.yml > prometheus.log 2>&1 &
```

### 访问 Prometheus Web UI

打开浏览器访问：`http://localhost:9090`

在查询框中输入以下查询来测试：

```promql
# 查看应用的 HTTP 请求总数
app_http_requests_total

# 查看内存使用情况
app_system_memory_usage_bytes
```

---

## 📊 集成 Grafana 仪表板

### 安装 Grafana

#### 使用 Docker（推荐）

```bash
# 运行 Grafana 容器
docker run -d \
  -p 3000:3000 \
  --name=grafana \
  -e "GF_SECURITY_ADMIN_PASSWORD=admin" \
  grafana/grafana-enterprise
```

#### 手动安装

访问 [Grafana 官网](https://grafana.com/grafana/download) 下载对应系统版本。

### 配置 Grafana

1. **登录 Grafana**

   - 地址：`http://localhost:3000`
   - 用户名：`admin`
   - 密码：`admin`（首次登录会要求修改）

2. **添加数据源**

   - 点击左侧菜单 "Configuration" → "Data Sources"
   - 点击 "Add data source"
   - 选择 "Prometheus"
   - URL 填写：`http://localhost:9090`
   - 点击 "Save & Test"

3. **创建仪表板**

#### 方法 1: 导入预置仪表板

```json
{
  "dashboard": {
    "title": "ThinkGin 应用监控",
    "panels": [
      {
        "title": "HTTP 请求数量",
        "type": "stat",
        "targets": [
          {
            "expr": "sum(rate(app_http_requests_total[5m]))",
            "legendFormat": "请求/秒"
          }
        ]
      },
      {
        "title": "HTTP 响应时间",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(app_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95% 响应时间"
          }
        ]
      }
    ]
  }
}
```

#### 方法 2: 手动创建面板

1. 点击 "+" → "Dashboard"
2. 点击 "Add new panel"
3. 在查询框输入 PromQL 查询
4. 选择合适的可视化类型
5. 配置面板标题和样式
6. 保存仪表板

### 常用仪表板配置

#### HTTP 请求监控面板

```promql
# 请求速率
sum(rate(app_http_requests_total[5m])) by (status)

# 平均响应时间
rate(app_http_request_duration_seconds_sum[5m]) / rate(app_http_request_duration_seconds_count[5m])

# 错误率
sum(rate(app_http_requests_total{status=~"5.."}[5m])) / sum(rate(app_http_requests_total[5m])) * 100
```

#### 系统资源监控面板

```promql
# 内存使用量（MB）
app_system_memory_usage_bytes / 1024 / 1024

# 进程运行时间
time() - app_process_start_time_seconds
```

---

## 💻 业务代码中的监控

### 基本使用方法

在您的控制器或业务逻辑中添加监控代码：

```go
package controller

import (
    "time"
    "thinkgin/extend/middleware"
    "github.com/gin-gonic/gin"
)

func UserController(c *gin.Context) {
    start := time.Now()

    // 记录 API 调用次数
    if counter := middleware.BusinessCounter("api_calls", []string{"endpoint", "status"}); counter != nil {
        counter.WithLabelValues("/api/users", "success").Inc()
    }

    // 模拟业务逻辑
    users := getUserList()

    // 记录 API 响应时间
    if histogram := middleware.BusinessHistogram("api_duration", []string{"endpoint"}, nil); histogram != nil {
        histogram.WithLabelValues("/api/users").Observe(time.Since(start).Seconds())
    }

    // 记录当前在线用户数
    if gauge := middleware.BusinessGauge("online_users", []string{"type"}); gauge != nil {
        gauge.WithLabelValues("active").Set(float64(getOnlineUserCount()))
    }

    c.JSON(200, gin.H{"users": users})
}
```

### 高级使用示例

#### 1. 数据库操作监控

```go
func DatabaseOperation() {
    start := time.Now()

    // 记录数据库查询次数
    if counter := middleware.BusinessCounter("db_queries", []string{"operation", "table"}); counter != nil {
        counter.WithLabelValues("select", "users").Inc()
    }

    // 执行数据库操作
    result := db.Query("SELECT * FROM users")

    // 记录查询耗时
    if histogram := middleware.BusinessHistogram("db_query_duration", []string{"operation"}, nil); histogram != nil {
        histogram.WithLabelValues("select").Observe(time.Since(start).Seconds())
    }

    // 如果查询失败，记录错误
    if result.Error != nil {
        if counter := middleware.BusinessCounter("db_errors", []string{"operation"}); counter != nil {
            counter.WithLabelValues("select").Inc()
        }
    }
}
```

#### 2. 缓存操作监控

```go
func CacheOperation(key string) {
    // 记录缓存操作
    if counter := middleware.BusinessCounter("cache_operations", []string{"operation", "result"}); counter != nil {
        value := getFromCache(key)
        if value != nil {
            counter.WithLabelValues("get", "hit").Inc()
        } else {
            counter.WithLabelValues("get", "miss").Inc()
        }
    }

    // 记录缓存大小
    if gauge := middleware.BusinessGauge("cache_size", []string{"type"}); gauge != nil {
        gauge.WithLabelValues("redis").Set(float64(getCacheSize()))
    }
}
```

#### 3. 自定义业务指标

```go
func ProcessOrder(order Order) {
    start := time.Now()

    // 记录订单处理次数
    if counter := middleware.BusinessCounter("orders_processed", []string{"status", "payment_method"}); counter != nil {
        counter.WithLabelValues("success", order.PaymentMethod).Inc()
    }

    // 记录订单金额分布
    if histogram := middleware.BusinessHistogram("order_amount", []string{"currency"},
        []float64{10, 50, 100, 500, 1000, 5000}); histogram != nil {
        histogram.WithLabelValues(order.Currency).Observe(order.Amount)
    }

    // 记录处理时间
    if histogram := middleware.BusinessHistogram("order_processing_time", []string{"type"}, nil); histogram != nil {
        histogram.WithLabelValues("payment").Observe(time.Since(start).Seconds())
    }

    // 记录当前待处理订单数
    if gauge := middleware.BusinessGauge("pending_orders", []string{"priority"}); gauge != nil {
        gauge.WithLabelValues(order.Priority).Set(float64(getPendingOrderCount(order.Priority)))
    }
}
```

### 监控指标命名最佳实践

#### 1. 命名规范

```go
// 好的命名
middleware.BusinessCounter("user_login_attempts", []string{"result", "method"})
middleware.BusinessHistogram("payment_processing_duration", []string{"gateway"}, nil)
middleware.BusinessGauge("active_connections", []string{"protocol"})

// 不好的命名
middleware.BusinessCounter("counter1", []string{"a", "b"})
middleware.BusinessHistogram("time", []string{"x"}, nil)
```

#### 2. 标签设计

```go
// 好的标签设计 - 标签值数量有限且有意义
counter.WithLabelValues("login", "success", "email")  // 3个固定的值
counter.WithLabelValues("payment", "failed", "paypal") // 可预期的值

// 避免的标签设计 - 可能产生无限多的标签值
counter.WithLabelValues(userID)  // 每个用户一个标签值
counter.WithLabelValues(timestamp) // 每次调用不同的值
```

---

## 📈 常用监控查询

### 基础查询语法

```promql
# 简单查询 - 获取当前值
app_http_requests_total

# 带标签过滤
app_http_requests_total{status="200"}

# 标签匹配
app_http_requests_total{status=~"2.."}  # 2xx 状态码
app_http_requests_total{status!="200"} # 非 200 状态码
```

### HTTP 监控查询

```promql
# 1. 请求速率（每秒请求数）
sum(rate(app_http_requests_total[5m]))

# 2. 按状态码分组的请求速率
sum(rate(app_http_requests_total[5m])) by (status)

# 3. 错误率（百分比）
sum(rate(app_http_requests_total{status=~"5.."}[5m])) / sum(rate(app_http_requests_total[5m])) * 100

# 4. 平均响应时间
rate(app_http_request_duration_seconds_sum[5m]) / rate(app_http_request_duration_seconds_count[5m])

# 5. 95% 分位响应时间
histogram_quantile(0.95, rate(app_http_request_duration_seconds_bucket[5m]))

# 6. 请求大小中位数
histogram_quantile(0.5, rate(app_http_request_size_bytes_bucket[5m]))
```

### 系统资源查询

```promql
# 1. 内存使用量（MB）
app_system_memory_usage_bytes / 1024 / 1024

# 2. 内存使用趋势（过去1小时）
increase(app_system_memory_usage_bytes[1h])

# 3. 应用运行时间（小时）
(time() - app_process_start_time_seconds) / 3600
```

### 业务指标查询

```promql
# 1. API 调用成功率
sum(rate(app_business_counter_api_calls{status="success"}[5m])) /
sum(rate(app_business_counter_api_calls[5m])) * 100

# 2. 平均 API 响应时间
rate(app_business_histogram_api_duration_sum[5m]) /
rate(app_business_histogram_api_duration_count[5m])

# 3. 当前在线用户数
app_business_gauge_online_users

# 4. 每分钟处理的订单数
increase(app_business_counter_orders_processed[1m])
```

### 复合查询和告警

```promql
# 1. 应用是否正常运行（1=正常，0=异常）
up{job="thinkgin"}

# 2. 高错误率告警（错误率超过5%）
sum(rate(app_http_requests_total{status=~"5.."}[5m])) / sum(rate(app_http_requests_total[5m])) > 0.05

# 3. 响应时间过慢告警（95%请求超过1秒）
histogram_quantile(0.95, rate(app_http_request_duration_seconds_bucket[5m])) > 1

# 4. 内存使用量告警（超过500MB）
app_system_memory_usage_bytes > 500 * 1024 * 1024
```

---

## 🔍 故障排除

### 常见问题及解决方案

#### 1. 访问 /metrics 返回 404

**可能原因：**

- 监控功能未启用
- 配置文件错误
- 路径配置错误

**解决步骤：**

```bash
# 1. 检查应用总开关
cat config/app.yaml | grep -A3 monitoring

# 应该看到：
# monitoring:
#   prometheus_enabled: true

# 2. 检查详细配置
cat config/prometheus.yaml | grep -A3 prometheus

# 应该看到：
# prometheus:
#   enabled: true
#   path: "/metrics"

# 3. 重启应用
go run main.go

# 4. 检查应用日志
# 确保没有配置加载错误
```

#### 2. 监控端点无数据

**可能原因：**

- 没有 HTTP 请求产生监控数据
- 监控中间件未正确加载

**解决步骤：**

```bash
# 1. 先访问一些页面生成数据
curl http://localhost:8000/
curl http://localhost:8000/index/hello

# 2. 再检查监控端点
curl http://localhost:8000/metrics | grep http_requests_total

# 3. 如果仍无数据，检查路由配置
# 确保 middleware.PrometheusMiddleware() 已添加
```

#### 3. Prometheus 无法抓取数据

**可能原因：**

- 网络连接问题
- 认证配置错误
- 目标地址错误

**解决步骤：**

```bash
# 1. 检查网络连接
curl http://localhost:8000/metrics

# 2. 检查 Prometheus 配置
cat prometheus.yml

# 3. 查看 Prometheus 日志
# 在 Prometheus Web UI 的 Status -> Targets 页面查看
```

#### 4. 监控数据不准确

**可能原因：**

- 时间窗口设置不当
- 查询语法错误
- 多实例重复计数

**解决步骤：**

```promql
# 1. 调整时间窗口
rate(app_http_requests_total[1m])  # 1分钟窗口
rate(app_http_requests_total[5m])  # 5分钟窗口

# 2. 检查查询语法
# 使用 Prometheus Web UI 的查询功能验证

# 3. 避免重复计数
sum(rate(app_http_requests_total[5m])) by (instance)
```

### 调试技巧

#### 1. 启用详细日志

```yaml
# config/log.yaml
log:
  default:
    level: "debug" # 改为 debug 级别
```

#### 2. 检查配置加载

```go
// 在代码中添加调试输出
func debugConfig() {
    config := app.GetConfig()
    fmt.Printf("Prometheus enabled: %v\n", config.App.Monitoring.PrometheusEnabled)
    fmt.Printf("Prometheus config: %+v\n", config.Prometheus)
}
```

#### 3. 手动测试指标

```bash
# 使用 curl 详细检查响应
curl -v http://localhost:8000/metrics

# 查看响应头
curl -I http://localhost:8000/metrics

# 过滤特定指标
curl -s http://localhost:8000/metrics | grep "app_http_requests_total"
```

---

## 🏆 最佳实践

### 1. 监控指标设计

#### 选择合适的指标类型

```go
// ✅ 正确的指标选择
counter := middleware.BusinessCounter("orders_total", []string{"status"})        // Counter: 累计值
histogram := middleware.BusinessHistogram("response_time", []string{"api"}, nil) // Histogram: 分布
gauge := middleware.BusinessGauge("queue_size", []string{"type"})               // Gauge: 瞬时值

// ❌ 错误的指标选择
counter := middleware.BusinessCounter("current_temperature", []string{})         // 应该用 Gauge
gauge := middleware.BusinessGauge("total_requests", []string{})                 // 应该用 Counter
```

#### 标签设计原则

```go
// ✅ 好的标签设计
counter.WithLabelValues("api", "success", "POST")  // 有限的、有意义的标签值

// ❌ 避免的标签设计
counter.WithLabelValues(userID, timestamp)         // 无限增长的标签值
counter.WithLabelValues(requestID)                 // 每次都不同的值
```

### 2. 性能优化

#### 减少指标数量

```go
// ✅ 合理的指标数量
if counter := middleware.BusinessCounter("api_calls", []string{"endpoint", "status"}); counter != nil {
    counter.WithLabelValues("/users", "success").Inc()
}

// ❌ 过多的标签组合
if counter := middleware.BusinessCounter("api_calls",
    []string{"endpoint", "method", "status", "user_agent", "client_ip"}); counter != nil {
    // 可能产生成千上万个时间序列
}
```

#### 使用采样

```go
// 对于高频操作，使用采样
func highFrequencyOperation() {
    // 只记录 1% 的操作
    if rand.Float64() < 0.01 {
        if counter := middleware.BusinessCounter("high_freq_ops", []string{}); counter != nil {
            counter.WithLabelValues().Add(100) // 乘以采样率
        }
    }
}
```

### 3. 监控策略

#### 分层监控

```yaml
# 1. 基础设施层监控
# - CPU、内存、磁盘、网络
# - 进程状态、连接数

# 2. 应用层监控
# - HTTP 请求指标
# - 数据库连接和查询
# - 缓存命中率

# 3. 业务层监控
# - 用户行为指标
# - 业务流程指标
# - 收入和转化率
```

#### 告警策略

```promql
# 1. 错误率告警（5分钟内错误率超过5%）
sum(rate(app_http_requests_total{status=~"5.."}[5m])) /
sum(rate(app_http_requests_total[5m])) > 0.05

# 2. 响应时间告警（95%请求超过2秒）
histogram_quantile(0.95, rate(app_http_request_duration_seconds_bucket[5m])) > 2

# 3. 系统资源告警（内存使用超过80%）
app_system_memory_usage_bytes / (1024*1024*1024) > 0.8

# 4. 业务指标告警（在线用户数异常下降）
decrease(app_business_gauge_online_users[10m]) > 100
```

### 4. 安全考虑

#### 保护监控端点

```yaml
# 启用认证
prometheus:
  auth:
    enabled: true
    username: "monitor"
    password: "strong_password_here"
```

#### 网络访问控制

```bash
# 使用防火墙限制访问
# 只允许 Prometheus 服务器访问监控端点
iptables -A INPUT -p tcp --dport 8000 -s 10.0.1.100 -j ACCEPT
iptables -A INPUT -p tcp --dport 8000 -j DROP
```

### 5. 容量规划

#### 估算存储需求

```bash
# 计算公式：
# 存储大小 = 时间序列数量 × 数据点大小 × 保留时间 × 采样频率

# 示例：
# - 1000个时间序列
# - 每个数据点 2 字节
# - 保留 30 天
# - 15秒采样一次
# 存储需求 = 1000 × 2 × (30×24×3600/15) ≈ 345 MB
```

#### 监控系统本身

```promql
# Prometheus 存储使用量
prometheus_tsdb_symbol_table_size_bytes

# 查询性能
prometheus_engine_query_duration_seconds

# 抓取目标健康状态
up{job="thinkgin"}
```

---

## 📞 获取帮助

如果您在使用过程中遇到问题：

1. **检查配置**: 确保配置文件语法正确
2. **查看日志**: 检查应用启动日志中的错误信息
3. **验证网络**: 确保端口没有被防火墙阻止
4. **参考文档**: 查阅 Prometheus 官方文档
5. **社区支持**: 在相关技术论坛寻求帮助

## 🔗 相关资源

- [Prometheus 官方文档](https://prometheus.io/docs/)
- [Grafana 官方文档](https://grafana.com/docs/)
- [PromQL 查询语言](https://prometheus.io/docs/prometheus/latest/querying/)
- [监控最佳实践](https://prometheus.io/docs/practices/)

---

_最后更新：2025 年 6 月_
