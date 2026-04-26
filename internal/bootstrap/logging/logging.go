package logging

import (
	"fmt"
	stdlog "log"
	"os"
	"path/filepath"
	"strings"

	ppanellog "github.com/OmnTeam/ppanel-pro/pkg/logger"
	kratoslog "github.com/go-kratos/kratos/v2/log"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Level          string `json:"level" yaml:"level"`
	Path           string `json:"path" yaml:"path"`
	Format         string `json:"format" yaml:"format"`
	DisableConsole bool   `json:"disable_console" yaml:"disable_console"`
	MaxSizeMB      int    `json:"max_size_mb" yaml:"max_size_mb"`
	MaxBackups     int    `json:"max_backups" yaml:"max_backups"`
	MaxAgeDays     int    `json:"max_age_days" yaml:"max_age_days"`
	Compress       bool   `json:"compress" yaml:"compress"`
}

func DefaultConfig(serviceName string) Config {
	return Config{
		Level:      "info",
		Path:       filepath.Join("logs", serviceName+".log"),
		Format:     "json",
		MaxSizeMB:  100,
		MaxBackups: 30,
		MaxAgeDays: 7,
		Compress:   true,
	}
}

func New(cfg Config, serviceID, serviceName, serviceVersion string) (*zap.Logger, func() error, error) {
	if serviceName == "" {
		serviceName = "ppanel-pro"
	}
	if serviceVersion == "" {
		serviceVersion = "dev"
	}
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	if cfg.Format == "" {
		cfg.Format = "json"
	}
	cfg.Format = normalizeFormat(cfg.Format)
	if cfg.MaxSizeMB <= 0 {
		cfg.MaxSizeMB = 100
	}
	if cfg.MaxBackups <= 0 {
		cfg.MaxBackups = 30
	}
	if cfg.MaxAgeDays <= 0 {
		cfg.MaxAgeDays = 7
	}

	level := zap.NewAtomicLevel()
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return nil, nil, fmt.Errorf("parse log level %q: %w", cfg.Level, err)
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	encoder := newEncoder(cfg.Format, encoderConfig)

	cores := make([]zapcore.Core, 0, 2)

	if !cfg.DisableConsole {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
	}

	var rotateLogger *lumberjack.Logger
	if cfg.Path != "" {
		filePath, err := resolveLogPath(cfg.Path, serviceName)
		if err != nil {
			return nil, nil, err
		}
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
			return nil, nil, fmt.Errorf("create log directory: %w", err)
		}
		rotateLogger = &lumberjack.Logger{
			Filename:   filePath,
			MaxSize:    cfg.MaxSizeMB,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAgeDays,
			Compress:   cfg.Compress,
			LocalTime:  true,
		}
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(rotateLogger), level))
	}

	if len(cores) == 0 {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
	}

	base := zap.New(zapcore.NewTee(cores...), zap.AddStacktrace(zap.ErrorLevel)).With(
		zap.String("service.id", serviceID),
		zap.String("service.name", serviceName),
		zap.String("service.version", serviceVersion),
	)

	stdlog.SetFlags(0)
	stdlog.SetOutput(&stdLogSink{logger: base.Named("stdlib")})

	cleanup := func() error {
		var firstErr error
		if err := base.Sync(); err != nil && !isIgnorableSyncError(err) {
			firstErr = err
		}
		if rotateLogger != nil {
			if err := rotateLogger.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
		return firstErr
	}

	return base, cleanup, nil
}

func NewKratosLogger(base *zap.Logger) kratoslog.Logger {
	return &kratosZapLogger{logger: base.Named("kratos")}
}

func NewPPanelWriter(base *zap.Logger) ppanellog.Writer {
	return &ppanelZapWriter{logger: base.Named("ppanel")}
}

type kratosZapLogger struct {
	logger *zap.Logger
}

func (l *kratosZapLogger) Log(level kratoslog.Level, keyvals ...interface{}) error {
	msg, fields := splitKeyvals(keyvals)
	switch level {
	case kratoslog.LevelDebug:
		l.logger.Debug(msg, fields...)
	case kratoslog.LevelWarn:
		l.logger.Warn(msg, fields...)
	case kratoslog.LevelError:
		l.logger.Error(msg, fields...)
	case kratoslog.LevelFatal:
		l.logger.Error(msg, append(fields, zap.String("kind", "fatal"))...)
	default:
		l.logger.Info(msg, fields...)
	}
	return nil
}

type ppanelZapWriter struct {
	logger *zap.Logger
}

func (w *ppanelZapWriter) Alert(v any) {
	w.logger.Warn(stringify(v), zap.String("kind", "alert"))
}

func (w *ppanelZapWriter) Close() error {
	return nil
}

func (w *ppanelZapWriter) Debug(v any, fields ...ppanellog.LogField) {
	w.logger.Debug(stringify(v), toZapFields(fields)...)
}

func (w *ppanelZapWriter) Error(v any, fields ...ppanellog.LogField) {
	w.logger.Error(stringify(v), toZapFields(fields)...)
}

func (w *ppanelZapWriter) Info(v any, fields ...ppanellog.LogField) {
	w.logger.Info(stringify(v), toZapFields(fields)...)
}

func (w *ppanelZapWriter) Severe(v any) {
	w.logger.Error(stringify(v), zap.String("kind", "severe"))
}

func (w *ppanelZapWriter) Slow(v any, fields ...ppanellog.LogField) {
	baseFields := append([]zap.Field{zap.String("kind", "slow")}, toZapFields(fields)...)
	w.logger.Warn(stringify(v), baseFields...)
}

func (w *ppanelZapWriter) Stack(v any) {
	w.logger.Error(stringify(v), zap.String("kind", "stack"))
}

func (w *ppanelZapWriter) Stat(v any, fields ...ppanellog.LogField) {
	baseFields := append([]zap.Field{zap.String("kind", "stat")}, toZapFields(fields)...)
	w.logger.Info(stringify(v), baseFields...)
}

type stdLogSink struct {
	logger *zap.Logger
}

func (w *stdLogSink) Write(p []byte) (int, error) {
	w.logger.Info(strings.TrimSpace(string(p)))
	return len(p), nil
}

func splitKeyvals(keyvals []interface{}) (string, []zap.Field) {
	fields := make([]zap.Field, 0, len(keyvals)/2)
	msg := ""
	for i := 0; i < len(keyvals); i += 2 {
		key := fmt.Sprintf("key_%d", i)
		value := interface{}(nil)
		if i < len(keyvals) {
			key = fmt.Sprint(keyvals[i])
		}
		if i+1 < len(keyvals) {
			value = keyvals[i+1]
		}
		if key == "msg" || key == "message" {
			msg = stringify(value)
			continue
		}
		fields = append(fields, zap.Any(key, value))
	}
	if msg == "" {
		msg = "kratos log"
	}
	return msg, fields
}

func toZapFields(fields []ppanellog.LogField) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	zapFields := make([]zap.Field, 0, len(fields))
	for _, field := range fields {
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}
	return zapFields
}

func stringify(v any) string {
	switch value := v.(type) {
	case nil:
		return ""
	case string:
		return value
	case error:
		return value.Error()
	case fmt.Stringer:
		return value.String()
	default:
		return fmt.Sprint(value)
	}
}

func normalizeFormat(format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "json":
		return "json"
	case "console", "text", "plain":
		return "console"
	default:
		return "json"
	}
}

func newEncoder(format string, cfg zapcore.EncoderConfig) zapcore.Encoder {
	if normalizeFormat(format) == "console" {
		return zapcore.NewConsoleEncoder(cfg)
	}
	return zapcore.NewJSONEncoder(cfg)
}

func resolveLogPath(pathValue, serviceName string) (string, error) {
	cleaned := filepath.Clean(pathValue)
	if cleaned == "." || cleaned == string(filepath.Separator) || filepath.Ext(cleaned) == "" {
		return filepath.Join(cleaned, serviceName+".log"), nil
	}
	return cleaned, nil
}

func isIgnorableSyncError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "invalid argument") || strings.Contains(msg, "inappropriate ioctl for device")
}
