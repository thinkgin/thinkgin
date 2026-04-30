// 本文件提供基于 OpenTelemetry 的链路追踪能力。
//
// 支持的导出器（由 trace.driver 配置决定）：
//   - "stdout"（默认）：输出到 runtime/log/trace.log，适合本地开发。
//   - "otlp"：通过 OTLP 协议发送到 Jaeger / Tempo / SigNoz 等后端。
//     根据 trace.otel.use_grpc 自动选择 gRPC（默认 4317）或 HTTP（默认 4318）。
//   - "jaeger"：作为 "otlp" 的别名（Jaeger >= 1.35 原生支持 OTLP）。
//
// Gin 中间件：直接复用官方 otelgin.Middleware，不自己造轮子。
package middleware

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// tracerProvider 持有全局 TracerProvider，方便 main 注册 Shutdown。
var tracerProvider *sdktrace.TracerProvider

// noopShutdown 用于在追踪未启用时返回的无副作用关闭函数。
func noopShutdown(context.Context) error { return nil }

// InitTracer 根据 config/trace.yaml 初始化 TracerProvider。
// 返回值是 Shutdown 函数，建议在 main 的 defer 中调用以落盘缓冲数据。
func InitTracer() func(context.Context) error {
	cfg := app.GetConfig()
	if cfg == nil || !cfg.Trace.Enabled {
		return noopShutdown
	}

	exporter, driverName, err := newExporter(cfg)
	if err != nil {
		app.GetLogger().Warnf("[trace] 初始化导出器失败: %v", err)
		return noopShutdown
	}

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(buildSampler(cfg.Trace.SampleRate)),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(buildResource(cfg)),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	app.GetLogger().Infof("[trace] 已启用 (driver=%s, sample_rate=%.2f)", driverName, cfg.Trace.SampleRate)
	return tracerProvider.Shutdown
}

// TraceMiddleware 返回 Gin 链路追踪中间件。
// 未启用时返回空操作，避免额外开销。
func TraceMiddleware() gin.HandlerFunc {
	cfg := app.GetConfig()
	if cfg == nil || !cfg.Trace.Enabled {
		return func(c *gin.Context) { c.Next() }
	}
	// 使用官方 otelgin，自动处理 span 生命周期、路由模板、HTTP 语义字段。
	return otelgin.Middleware(cfg.Trace.ServiceName)
}

// newExporter 根据 trace.driver 创建对应的 SpanExporter。
// 返回值包含实际使用的 driver 名称（用于日志）。
func newExporter(cfg *app.GlobalConfig) (sdktrace.SpanExporter, string, error) {
	driver := strings.ToLower(strings.TrimSpace(cfg.Trace.Driver))

	switch driver {
	case "otlp", "jaeger", "otel":
		exp, err := newOTLPExporter(cfg)
		if err != nil {
			return nil, driver, err
		}
		return exp, driver, nil

	case "stdout", "":
		exp, err := newStdoutExporter()
		if err != nil {
			return nil, "stdout", err
		}
		return exp, "stdout", nil

	default:
		return nil, driver, fmt.Errorf("unsupported trace driver: %q (supported: stdout, otlp, jaeger)", driver)
	}
}

// newOTLPExporter 创建 OTLP 导出器。
// trace.otel.use_grpc=true → gRPC（默认端口 4317）；否则 HTTP（默认端口 4318）。
// 默认使用 insecure 连接，生产环境应在端点前配置 TLS 或网关。
func newOTLPExporter(cfg *app.GlobalConfig) (sdktrace.SpanExporter, error) {
	endpoint := cfg.Trace.Otel.Endpoint
	if endpoint == "" {
		if cfg.Trace.Otel.UseGrpc {
			endpoint = "localhost:4317"
		} else {
			endpoint = "localhost:4318"
		}
	}
	// 去掉 http:// / https:// 前缀，OTLP SDK 自己处理 scheme。
	endpoint = strings.TrimPrefix(endpoint, "http://")
	endpoint = strings.TrimPrefix(endpoint, "https://")

	ctx := context.Background()
	if cfg.Trace.Otel.UseGrpc {
		return otlptracegrpc.New(ctx,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithInsecure(),
		)
	}
	return otlptracehttp.New(ctx,
		otlptracehttp.WithEndpoint(endpoint),
		otlptracehttp.WithInsecure(),
	)
}

// newStdoutExporter 创建默认的 stdout 导出器，优先写入 runtime/log/trace.log。
// 目录不存在时自动创建；文件打开失败时回落到标准输出。
func newStdoutExporter() (sdktrace.SpanExporter, error) {
	const logPath = "runtime/log/trace.log"

	var w io.Writer = os.Stdout
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err == nil {
		if f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			w = f
		}
	}
	return stdouttrace.New(stdouttrace.WithWriter(w))
}

// buildSampler 构建采样器：
//   - rate <= 0 或 >= 1：全量采样（便于开发 / 默认）。
//   - 0 < rate < 1：按 TraceID 比例采样。
func buildSampler(rate float64) sdktrace.Sampler {
	if rate > 0 && rate < 1 {
		return sdktrace.TraceIDRatioBased(rate)
	}
	return sdktrace.AlwaysSample()
}

// buildResource 构建资源属性（服务名 + 用户自定义标签）。
func buildResource(cfg *app.GlobalConfig) *resource.Resource {
	attrs := []attribute.KeyValue{semconv.ServiceName(cfg.Trace.ServiceName)}
	for k, v := range cfg.Trace.Tags {
		attrs = append(attrs, attribute.String(k, v))
	}
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)
	return res
}
