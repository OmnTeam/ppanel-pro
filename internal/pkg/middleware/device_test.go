package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/OmnTeam/npanel-pro/internal/conf"
	"github.com/go-kratos/kratos/v2/log"
)

func TestDeviceMiddlewareRejectsCoveredRequestsWhenSecretEmpty(t *testing.T) {
	mw := DeviceMiddleware(&conf.Device{
		Enable:         true,
		SecuritySecret: "",
	}, log.NewStdLogger(io.Discard))

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("covered request without secret should not reach downstream handler")
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
	req.Header.Set("Login-Type", "web")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "{\"code\":40007,\"msg\":\"Secret is empty\"}" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestDeviceMiddlewareSkipsPreflightRequests(t *testing.T) {
	mw := DeviceMiddleware(&conf.Device{
		Enable:         true,
		SecuritySecret: "",
	}, log.NewStdLogger(io.Discard))

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodOptions, "/v1/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", http.MethodPost)

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusNoContent)
	}
}

func TestDeviceMiddlewareRejectsDeviceRequestsWithoutSecret(t *testing.T) {
	mw := DeviceMiddleware(&conf.Device{
		Enable:         true,
		SecuritySecret: "",
	}, log.NewStdLogger(io.Discard))

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("device request without secret should not reach downstream handler")
	}))

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
	req.Header.Set("Login-Type", "device")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}
	if body := strings.TrimSpace(rec.Body.String()); body != "{\"code\":40007,\"msg\":\"Secret is empty\"}" {
		t.Fatalf("unexpected body: %q", body)
	}
}

func TestNewResponseWriterInitializesDefaultStatus(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)

	rw := NewResponseWriter(rec, req, "secret", log.NewStdLogger(io.Discard))

	if rw.status != defaultStatus {
		t.Fatalf("unexpected default status: got %d want %d", rw.status, defaultStatus)
	}
	if rw.size != noWritten {
		t.Fatalf("unexpected initial size: got %d want %d", rw.size, noWritten)
	}
}

func TestDeviceMiddlewarePreservesCORSWhenRejected(t *testing.T) {
	handler := CORSFilter(&conf.Server_CORS{
		Enable:           true,
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           86400,
	})(DeviceMiddleware(&conf.Device{
		Enable:         true,
		SecuritySecret: "",
	}, log.NewStdLogger(io.Discard))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("device request without secret should not reach downstream handler")
	})))

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Login-Type", "device")

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("unexpected allow origin: got %q", got)
	}
}
