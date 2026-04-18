package middleware

import (
	"context"
	"io"
	"os"

	"thinkgin/app"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.opentelemetry.io/otel/trace"
)

var tracerProvider *sdktrace.TracerProvider

// InitTracer 初始化链路追踪
// 当前使用 stdout 导出器，生产环境可替换为 OTLP/Jaeger 导出器
func InitTracer() func(context.Context) error {
	cfg := app.GetConfig()
	if !cfg.Trace.Enabled {
		return func(ctx context.Context) error { return nil }
	}

	var exporter sdktrace.SpanExporter
	var err error

	// 根据配置选择导出方式
	switch cfg.Trace.Driver {
	default:
		// 默认使用 stdout，输出到日志目录
		var w io.Writer
		w, err = os.OpenFile("runtime/log/trace.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			w = os.Stdout
		}
		exporter, err = stdouttrace.New(stdouttrace.WithWriter(w))
	}

	if err != nil {
		logrus.Warnf("[Trace] 初始化导出器失败: %v", err)
		return func(ctx context.Context) error { return nil }
	}

	// 构建采样率
	sampler := sdktrace.AlwaysSample()
	if cfg.Trace.SampleRate > 0 && cfg.Trace.SampleRate < 1 {
		sampler = sdktrace.TraceIDRatioBased(cfg.Trace.SampleRate)
	}

	// 构建资源属性
	attrs := []attribute.KeyValue{
		semconv.ServiceName(cfg.Trace.ServiceName),
	}
	for k, v := range cfg.Trace.Tags {
		attrs = append(attrs, attribute.String(k, v))
	}

	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL, attrs...),
	)

	tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logrus.Infof("[Trace] 链路追踪已启用 (driver=%s, sample_rate=%.2f)", cfg.Trace.Driver, cfg.Trace.SampleRate)
	return tracerProvider.Shutdown
}

// TraceMiddleware 返回 Gin 链路追踪中间件
func TraceMiddleware() gin.HandlerFunc {
	cfg := app.GetConfig()
	if !cfg.Trace.Enabled {
		return func(c *gin.Context) { c.Next() }
	}

	tracer := otel.Tracer(cfg.Trace.ServiceName)

	return func(c *gin.Context) {
		spanName := c.Request.Method + " " + c.FullPath()
		if spanName == c.Request.Method+" " {
			spanName = c.Request.Method + " " + c.Request.URL.Path
		}

		ctx, span := tracer.Start(c.Request.Context(), spanName,
			trace.WithAttributes(
				semconv.HTTPMethod(c.Request.Method),
				semconv.HTTPTarget(c.Request.URL.Path),
				attribute.String("http.client_ip", c.ClientIP()),
			),
		)
		defer span.End()

		c.Request = c.Request.WithContext(ctx)
		c.Next()

		span.SetAttributes(
			semconv.HTTPStatusCode(c.Writer.Status()),
		)
	}
}
