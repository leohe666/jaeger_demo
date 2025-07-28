package main

import (
	"log" // 标准库 log，用于日志输出
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	otlog "github.com/opentracing/opentracing-go/log" // 为 opentracing 的 log 包设置别名，避免冲突
	"github.com/uber/jaeger-client-go"
	jaegercfg "github.com/uber/jaeger-client-go/config"
)

func main() {
	// 初始化 Jaeger
	cfg := jaegercfg.Configuration{
		ServiceName: "gin-jaeger-demo",
		Sampler: &jaegercfg.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegercfg.ReporterConfig{
			LogSpans:          true,
			CollectorEndpoint: "http://localhost:14268/api/traces",
		},
	}

	tracer, closer, err := cfg.NewTracer()
	if err != nil {
		log.Fatalf("初始化 Jaeger 失败: %v", err)
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	// 创建 Gin 路由
	r := gin.Default()

	// 添加 Jaeger 中间件
	r.Use(func(c *gin.Context) {
		// 从请求头提取追踪上下文
		spanCtx, _ := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(c.Request.Header))
		span := tracer.StartSpan(c.Request.URL.Path, ext.RPCServerOption(spanCtx))
		defer span.Finish()

		// 设置标准 HTTP 相关 Tag
		ext.HTTPMethod.Set(span, c.Request.Method)
		ext.HTTPUrl.Set(span, c.Request.URL.String())
		ext.Component.Set(span, "gin")

		// 记录查询参数
		for key, values := range c.Request.URL.Query() {
			span.SetTag("http.query."+key, values)
		}

		// 记录部分请求头（可选）
		for _, header := range []string{"User-Agent", "X-Request-Id"} {
			if value := c.GetHeader(header); value != "" {
				span.SetTag("http.header."+header, value)
			}
		}

		// 将 Span 注入到 Gin 上下文
		c.Request = c.Request.WithContext(opentracing.ContextWithSpan(c.Request.Context(), span))
		c.Next()

		// 记录响应状态码
		ext.HTTPStatusCode.Set(span, uint16(c.Writer.Status()))
	})

	// 测试路由
	r.GET("/hello", func(c *gin.Context) {
		span, _ := opentracing.StartSpanFromContext(c.Request.Context(), "hello-handler")
		defer span.Finish()

		// 记录响应数据（使用 opentracing 的 log 包）
		span.LogFields(otlog.String("response", `{"message":"Hello Jaeger!"}`))

		c.JSON(http.StatusOK, gin.H{
			"message": "Hello Jaeger!",
		})
	})

	// 启动服务
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
