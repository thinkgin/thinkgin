# ThinkGin2.0

![img.png](img.png)

#### 介绍

实现一个 Go 语言基于 gin 的增删改查基础框架

#### 软件架构

```azure
thinkgin  应用部署目录
├─app                应用目录（可设置）
├  └─index            默认模块
├   ├─controller      控制器
├   ├─model           模型
├   └─view            视图
├─extend             扩展目录
├─public             公共目录
├─route              路由
└─runtime            运行日志
```

#### 快速开始

**环境配置 (Go 1.13+)**

```bash
# 推荐方式
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
```

**macOS/Linux**

```bash
export GO111MODULE=on
export GOPROXY=https://goproxy.cn
```

**Windows**

```cmd
set GO111MODULE=on
set GOPROXY=https://goproxy.cn
```

**运行项目**

```bash
go mod init thinkgin
go mod tidy
go run main.go
```

访问：http://localhost:8000/

#### 常见问题

- goland 导入包爆红的问题的解决方案（成功解决 Nice）：
  `GOPROXY=https://goproxy.cn,direct`
  ![img_1.png](https://gitee.com/goubiwanyi/thinkgin/raw/master/img/img_1.png)

#### 路由规范

###### 路由编写文件：`route/router.go`

###### 访问示例：`http://thinkgin.cn:8000/index/hello`

```azure
{
    "code": 200,
    "data": {
        "Time": "2023-02-01T10:11:41.76127929+08:00",
        "data": "Hello ThinkGin!",
        "name": ""
    },
    "msg": "ok"
}
```

#### Linux 打包

```azure
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
go build -o thinkgin.sh
```

##### 直接运行即可：

`./thinkgin.sh`

##### 或后台执行：

`nohup ./thinkgin.sh 1>info.log 2>&1 &`
