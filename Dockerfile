# ThinkGin 应用镜像（多阶段构建）
#
# 构建：docker build -t thinkgin:latest .
# 运行：docker run -p 8000:8000 -e THINKGIN_APP_JWT_SECRET=<your-secret> thinkgin:latest
#
# 设计要点：
#   - 多阶段：builder 阶段编译，最终镜像仅含静态二进制 + 配置，体积小。
#   - 静态编译（CGO_ENABLED=0）：SQLite 驱动用纯 Go 实现（glebarez/sqlite），无需 libc。
#   - 非 root 运行：降低容器逃逸风险。
#   - 版本注入：通过 ldflags 把 VERSION 写入 app.Version。

# ---------- 构建阶段 ----------
FROM golang:1.25-alpine AS builder

# 构建参数：版本号（CI 可传 --build-arg VERSION=3.x.y）
ARG VERSION=dev

WORKDIR /src

# 国内构建可解开下面一行加速依赖拉取
# ENV GOPROXY=https://goproxy.cn,direct

# 先拷贝 go.mod/go.sum 以最大化利用层缓存
COPY go.mod go.sum ./
RUN go mod download

# 拷贝源码并编译为静态二进制
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags "-s -w -X thinkgin/app.Version=${VERSION}" \
    -o /out/thinkgin ./main.go

# ---------- 运行阶段 ----------
FROM alpine:3.20

# ca-certificates 用于 HTTPS 出站；tzdata 让 Asia/Shanghai 时区可用
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S -G app app

WORKDIR /app

# 拷贝二进制与默认配置
COPY --from=builder /out/thinkgin /app/thinkgin
COPY --from=builder /src/config /app/config

# 运行期目录（日志等），赋权给非 root 用户
RUN mkdir -p /app/runtime/log && chown -R app:app /app

USER app

EXPOSE 8000

# 健康检查：复用框架的 /livez 探针
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8000/livez || exit 1

# 生产建议以 release 模式运行，并通过环境变量注入敏感配置
ENV THINKGIN_SERVER_MODE=release \
    THINKGIN_APP_DEBUG=false

ENTRYPOINT ["/app/thinkgin"]
CMD ["serve"]
