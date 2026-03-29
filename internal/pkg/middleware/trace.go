package middleware

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/OmnTeam/ppanel-pro/pkg/trace"
	"github.com/go-kratos/kratos/v2/log"
)

const (
	// RequestIdKey 请求ID的HTTP头key
	RequestIdKey = "X-Request-ID"
)

// statusByWriter 根据HTTP状态码返回span状态
// 400-499范围的状态码不会作为错误返回
func statusByWriter(code int) (codes.Code, string) {
	if code < 100 || code >= 600 {
		return codes.Error, fmt.Sprintf("Invalid HTTP status code %d", code)
	}
	if code >= 500 {
		return codes.Error, ""
	}
	return codes.Unset, ""
}

// requestAttributes 从HTTP请求中提取属性
func requestAttributes(req *http.Request) []attribute.KeyValue {
	var (
		protoName    string
		protoVersion string
		clientAddr   = req.RemoteAddr
		clientPort   string
	)

	if parts := strings.SplitN(req.Proto, "/", 2); len(parts) > 0 {
		protoName = parts[0]
		if len(parts) == 2 {
			protoVersion = parts[1]
		}
	}
	if host, port, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		clientAddr = host
		clientPort = port
	}

	attrs := []attribute.KeyValue{
		semconv.HTTPRequestMethodKey.String(req.Method),
		semconv.HTTPUserAgentKey.String(req.UserAgent()),
		semconv.HTTPRequestContentLengthKey.Int64(req.ContentLength),

		semconv.URLFullKey.String(req.URL.String()),
		semconv.URLSchemeKey.String(req.URL.Scheme),
		semconv.URLFragmentKey.String(req.URL.Fragment),
		semconv.URLPathKey.String(req.URL.Path),
		semconv.URLQueryKey.String(req.URL.RawQuery),
	}

	if protoName != "" {
		attrs = append(attrs, semconv.NetworkProtocolNameKey.String(strings.ToLower(protoName)))
	}
	if protoVersion != "" {
		attrs = append(attrs, semconv.NetworkProtocolVersionKey.String(protoVersion))
	}
	if clientAddr != "" {
		attrs = append(attrs, semconv.ClientAddressKey.String(clientAddr))
	}
	if clientPort != "" {
		attrs = append(attrs, semconv.ClientPortKey.String(clientPort))
	}

	return attrs
}

// TraceMiddleware 追踪中间件
// 使用OpenTelemetry进行分布式追踪
func TraceMiddleware(logger log.Logger) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tracer := trace.TracerFromContext(ctx)

			// 从请求路径中提取span名称
			spanName := r.URL.Path
			method := r.Method

			ctx, span := tracer.Start(
				ctx,
				fmt.Sprintf("%s %s", method, spanName),
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			)
			defer span.End()

			requestId := trace.TraceIDFromContext(ctx)

			// 设置请求ID到响应头
			w.Header().Set(RequestIdKey, requestId)

			// 设置请求属性
			span.SetAttributes(requestAttributes(r)...)
			span.SetAttributes(
				attribute.String("http.request_id", requestId),
				semconv.HTTPRouteKey.String(r.URL.Path),
			)

			// 将request host存入context
			ctx = context.WithValue(ctx, "requestHost", r.Host)
			ctx = context.WithValue(ctx, "requestId", requestId)

			// 使用更新后的context继续处理请求
			r = r.WithContext(ctx)
			handler.ServeHTTP(w, r)

			// 处理响应相关属性
			status := getResponseWriterStatus(w)
			span.SetStatus(statusByWriter(status))
			if status > 0 {
				span.SetAttributes(semconv.HTTPResponseStatusCodeKey.Int(status))
			}
		})
	}
}

// getResponseWriterStatus 获取响应写入器的状态码
// 这是一个辅助函数，用于从ResponseWriter中获取状态码
func getResponseWriterStatus(w http.ResponseWriter) int {
	// 如果ResponseWriter实现了statusGetter接口
	type statusGetter interface {
		Status() int
	}

	if sg, ok := w.(statusGetter); ok {
		return sg.Status()
	}

	// 默认返回200
	return http.StatusOK
}
