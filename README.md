# ThinkGin
![img.png](img.png)

#### 介绍
实现一个Go语言基于gin的增删改查基础框架

#### 软件架构
```azure
thinkgin  应用部署目录
├─app                应用目录（可设置）
  └─index            默认模块
   ├─controller      控制器
   ├─model           模型
   └─view            视图
├─extend             扩展目录
├─public             公共目录
├─route              路由
└─runtime            运行日志
```


#### 安装教程

```azure

C:\> $env:GO111MODULE = "on"
C:\> $env:GOPROXY = "https://goproxy.cn,direct"
或者
C:\> $env:GOPROXY = "https://goproxy.io,direct"

go mod init thinkgin
go mod tidy
go run main.go

访问：
http://127.0.0.1:8000/  就可以看到上面的欢迎界面
```


#### 常见问题
- goland导入包爆红的问题的解决方案（成功解决Nice）：
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

#### Linux打包
```azure
go env -w GO111MODULE=on 
go env -w GOPROXY=https://goproxy.cn,direct
go build -o thinkgin.sh
```

##### 直接运行即可：
`./thinkgin.sh`
##### 或后台执行：
`nohup ./thinkgin.sh 1>info.log 2>&1 &`
