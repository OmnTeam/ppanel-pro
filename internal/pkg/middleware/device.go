package middleware

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"github.com/OmnTeam/npanel-pro/internal/responsecode"
	"github.com/OmnTeam/npanel-pro/pkg/constant"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/OmnTeam/npanel-pro/internal/conf"
	pkgaes "github.com/OmnTeam/npanel-pro/pkg/aes"
)

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

// DeviceMiddleware 设备加密中间件
// 用于处理设备客户端的请求加密/解密
func DeviceMiddleware(deviceConfig *conf.Device, logger log.Logger) func(http.Handler) http.Handler {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 如果设备功能未启用，直接跳过
			if deviceConfig == nil || !deviceConfig.Enable {
				handler.ServeHTTP(w, r)
				return
			}

			// 全局 Filter 下，预检和非设备请求必须直接放行。
			if r.Method == http.MethodOptions {
				handler.ServeHTTP(w, r)
				return
			}

			if !shouldProcessDevicePath(r.URL.Path) {
				handler.ServeHTTP(w, r)
				return
			}

			loginType := strings.TrimSpace(r.Header.Get("Login-Type"))
			if loginType != "" && r.Context().Value(constant.CtxKeyUser) == nil {
				ctx := context.WithValue(r.Context(), constant.LoginType, loginType)
				r = r.WithContext(ctx)
			}

			// Old middleware validated secret for all covered paths once device mode was enabled,
			// before checking whether this specific request is a device login flow.
			if strings.TrimSpace(deviceConfig.SecuritySecret) == "" {
				writeLegacyDeviceError(w, responsecode.ErrSecretIsEmpty, "Secret is empty")
				return
			}

			// 检查是否为设备登录类型
			if loginType != "device" {
				handler.ServeHTTP(w, r)
				return
			}

			// 创建加密响应写入器
			rw := NewResponseWriter(w, r, deviceConfig.SecuritySecret, logger)
			if !rw.Decrypt() {
				writeLegacyDeviceError(w, responsecode.ErrInvalidCiphertext, "Invalid ciphertext")
				return
			}

			// 使用自定义的ResponseWriter
			handler.ServeHTTP(rw, r)
			rw.FlushAbort()
		})
	}
}

func writeLegacyDeviceError(w http.ResponseWriter, code int, msg string) {
	if strings.TrimSpace(msg) == "" {
		msg = "Internal Server Error"
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"code": code,
		"msg":  msg,
	})
}

func shouldProcessDevicePath(path string) bool {
	switch path {
	case "/v1/public/user/device_ws_connect":
		return false
	}

	for _, prefix := range []string{
		"/v1/auth/",
		"/v1/common/",
		"/v1/public/announcement/",
		"/v1/public/document/",
		"/v1/public/order/",
		"/v1/public/payment/",
		"/v1/public/portal/",
		"/v1/public/redemption/",
		"/v1/public/subscribe/",
		"/v1/public/ticket/",
		"/v1/public/user/",
	} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}

	return false
}

// NewResponseWriter 创建加密响应写入器
func NewResponseWriter(w http.ResponseWriter, r *http.Request, encryptionKey string, logger log.Logger) *ResponseWriter {
	return &ResponseWriter{
		w:                w,
		r:                r,
		size:             noWritten,
		status:           defaultStatus,
		body:             new(bytes.Buffer),
		ResponseWriter:   w,
		encryptionKey:    encryptionKey,
		encryptionMethod: "AES",
		encryption:       true,
	}
}

// ResponseWriter 加密响应写入器
type ResponseWriter struct {
	http.ResponseWriter
	size             int
	status           int
	flush            bool
	body             *bytes.Buffer
	w                http.ResponseWriter
	r                *http.Request
	encryption       bool
	encryptionKey    string
	encryptionMethod string
}

// Encrypt 加密响应体
func (rw *ResponseWriter) Encrypt() {
	if !rw.encryption {
		return
	}

	buf := rw.body.Bytes()
	params := map[string]interface{}{}
	err := json.Unmarshal(buf, &params)
	if err != nil {
		log.Errorf("[DeviceMiddleware] Failed to unmarshal response: %v", err)
		return
	}

	data := params["data"]
	if data != nil {
		var jsonData []byte
		str, ok := data.(string)
		if ok {
			jsonData = []byte(str)
		} else {
			jsonData, _ = json.Marshal(data)
		}

		encrypt, iv, err := pkgaes.Encrypt(jsonData, rw.encryptionKey)
		if err != nil {
			log.Errorf("[DeviceMiddleware] Failed to encrypt: %v", err)
			return
		}

		params["data"] = map[string]interface{}{
			"data": encrypt,
			"time": iv,
		}
	}

	marshal, _ := json.Marshal(params)
	rw.body.Reset()
	rw.body.Write(marshal)
}

// Decrypt 解密请求体
func (rw *ResponseWriter) Decrypt() bool {
	if !rw.encryption {
		return true
	}

	// 判断URL参数中是否存在data和iv数据，存在就进行解密并设置回去
	query := rw.r.URL.Query()
	dataStr := query.Get("data")
	timeStr := query.Get("time")

	if dataStr != "" && timeStr != "" {
		decrypt, err := pkgaes.Decrypt(dataStr, rw.encryptionKey, timeStr)
		if err == nil {
			params := map[string]interface{}{}
			err = json.Unmarshal([]byte(decrypt), &params)
			if err == nil {
				for k, v := range params {
					query.Set(k, fmt.Sprintf("%v", v))
				}
				query.Del("data")
				query.Del("time")

				// 重建RequestURI
				pathParts := strings.SplitN(rw.r.RequestURI, "?", 2)
				if len(pathParts) == 2 {
					rw.r.RequestURI = fmt.Sprintf("%s?%s", pathParts[0], query.Encode())
				}
				rw.r.URL.RawQuery = query.Encode()
			}
		}
	}

	// 判断body是否存在数据，存在就尝试解密，并设置回去
	body, err := io.ReadAll(rw.r.Body)
	if err != nil {
		return true
	}

	if len(body) == 0 {
		return true
	}

	params := map[string]interface{}{}
	err = json.Unmarshal(body, &params)
	if err != nil {
		return false
	}

	data := params["data"]
	nonce := params["time"]
	if data == nil || nonce == nil {
		return false
	}

	str, ok := data.(string)
	if !ok {
		return false
	}
	iv, ok := nonce.(string)
	if !ok {
		return false
	}

	decrypt, err := pkgaes.Decrypt(str, rw.encryptionKey, iv)
	if err != nil {
		log.Errorf("[DeviceMiddleware] Failed to decrypt request: %v", err)
		return false
	}

	rw.r.Body = io.NopCloser(bytes.NewBuffer([]byte(decrypt)))
	return true
}

// FlushAbort 刷新并中止
func (rw *ResponseWriter) FlushAbort() {
	responseBody := rw.body.String()
	log.Debugf("[DeviceMiddleware] Original Response Body: %s", responseBody)
	rw.flush = true

	if rw.encryption {
		rw.Encrypt()
	}

	_, err := rw.Write(rw.body.Bytes())
	if err != nil {
		log.Errorf("[DeviceMiddleware] Failed to write response: %v", err)
	}
}

// Unwrap 解包
func (rw *ResponseWriter) Unwrap() http.ResponseWriter {
	return rw.ResponseWriter
}

// WriteHeader 写入状态码
func (rw *ResponseWriter) WriteHeader(code int) {
	if code > 0 && rw.status != code {
		if rw.Written() {
			return
		}
		rw.status = code
	}
}

// WriteHeaderNow 立即写入状态码
func (rw *ResponseWriter) WriteHeaderNow() {
	if !rw.Written() {
		rw.size = 0
		rw.ResponseWriter.WriteHeader(rw.status)
	}
}

// Write 写入数据
func (rw *ResponseWriter) Write(data []byte) (n int, err error) {
	if rw.flush {
		rw.WriteHeaderNow()
		n, err = rw.ResponseWriter.Write(data)
		rw.size += n
	} else {
		rw.body.Write(data)
	}
	return
}

// WriteString 写入字符串
func (rw *ResponseWriter) WriteString(s string) (n int, err error) {
	if rw.flush {
		rw.WriteHeaderNow()
		n, err = rw.ResponseWriter.Write([]byte(s))
		rw.size += n
	} else {
		rw.body.Write([]byte(s))
	}
	return
}

// Status 获取状态码
func (rw *ResponseWriter) Status() int {
	return rw.status
}

// Size 获取大小
func (rw *ResponseWriter) Size() int {
	return rw.size
}

// Written 是否已写入
func (rw *ResponseWriter) Written() bool {
	return rw.size != noWritten
}

// Hijack 劫持连接
func (rw *ResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if rw.size < 0 {
		rw.size = 0
	}
	if hijacker, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("ResponseWriter does not implement http.Hijacker")
}

// CloseNotify 关闭通知
func (rw *ResponseWriter) CloseNotify() <-chan bool {
	done := rw.r.Context().Done()
	closed := make(chan bool)

	go func() {
		<-done
		closed <- true
	}()

	return closed
}

// Flush 刷新
func (rw *ResponseWriter) Flush() {
	rw.WriteHeaderNow()
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Pusher 推送器
func (rw *ResponseWriter) Pusher() (pusher http.Pusher) {
	if p, ok := rw.ResponseWriter.(http.Pusher); ok {
		return p
	}
	return nil
}
