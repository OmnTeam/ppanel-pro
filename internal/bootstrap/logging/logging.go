package logging

import (
	"fmt"
	stdlog "log"
	"os"
	"path/filepath"
	"strings"

	npanellog "github.com/OmnTeam/npanel-pro/pkg/logger"
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
		serviceName = "npanel-pro"
	}
	if serviceVersion == "" {
		serviceVersion = "dev"
	}
	if cfg.Level == "" {
		cfg.Level = "info"
	}
	format, err := parseFormat(cfg.Format)
	if err != nil {
		return nil, nil, err
	}
	cfg.Format = format
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
	var fileLoggers *leveledLoggers

	if !cfg.DisableConsole {
		cores = append(cores, zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), level))
	}

	if cfg.Path != "" {
		filePath, err := resolveLogPath(cfg.Path, serviceName)
		if err != nil {
			return nil, nil, err
		}
		fileLoggers, err = newLeveledLoggers(filePath, cfg)
		if err != nil {
			return nil, nil, err
		}
		cores = append(cores,
			zapcore.NewCore(encoder, zapcore.AddSync(fileLoggers.debug), exactLevelEnabler(level, zap.DebugLevel)),
			zapcore.NewCore(encoder, zapcore.AddSync(fileLoggers.info), exactLevelEnabler(level, zap.InfoLevel)),
			zapcore.NewCore(encoder, zapcore.AddSync(fileLoggers.warn), exactLevelEnabler(level, zap.WarnLevel)),
			zapcore.NewCore(encoder, zapcore.AddSync(fileLoggers.error), errorLevelEnabler(level)),
		)
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
		if fileLoggers != nil {
			if err := fileLoggers.Close(); err != nil && firstErr == nil {
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

func NewNPanelWriter(base *zap.Logger) npanellog.Writer {
	return &npanelZapWriter{logger: base.Named("npanel")}
}

type leveledLoggers struct {
	debug *lumberjack.Logger
	info  *lumberjack.Logger
	warn  *lumberjack.Logger
	error *lumberjack.Logger
}

func (l *leveledLoggers) Close() error {
	if l == nil {
		return nil
	}

	var firstErr error
	for _, logger := range []*lumberjack.Logger{l.debug, l.info, l.warn, l.error} {
		if logger == nil {
			continue
		}
		if err := logger.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
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

type npanelZapWriter struct {
	logger *zap.Logger
}

func (w *npanelZapWriter) Alert(v any) {
	w.logger.Warn(stringify(v), zap.String("kind", "alert"))
}

func (w *npanelZapWriter) Close() error {
	return nil
}

func (w *npanelZapWriter) Debug(v any, fields ...npanellog.LogField) {
	w.logger.Debug(stringify(v), toZapFields(fields)...)
}

func (w *npanelZapWriter) Error(v any, fields ...npanellog.LogField) {
	w.logger.Error(stringify(v), toZapFields(fields)...)
}

func (w *npanelZapWriter) Info(v any, fields ...npanellog.LogField) {
	w.logger.Info(stringify(v), toZapFields(fields)...)
}

func (w *npanelZapWriter) Severe(v any) {
	w.logger.Error(stringify(v), zap.String("kind", "severe"))
}

func (w *npanelZapWriter) Slow(v any, fields ...npanellog.LogField) {
	baseFields := append([]zap.Field{zap.String("kind", "slow")}, toZapFields(fields)...)
	w.logger.Warn(stringify(v), baseFields...)
}

func (w *npanelZapWriter) Stack(v any) {
	w.logger.Error(stringify(v), zap.String("kind", "stack"))
}

func (w *npanelZapWriter) Stat(v any, fields ...npanellog.LogField) {
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

func toZapFields(fields []npanellog.LogField) []zap.Field {
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

func parseFormat(format string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "json":
		return "json", nil
	case "console", "text", "plain", "pretty":
		return "console", nil
	default:
		return "", fmt.Errorf("unsupported log format %q: supported values are json, console, text, plain, pretty", format)
	}
}

func newEncoder(format string, cfg zapcore.EncoderConfig) zapcore.Encoder {
	if format == "console" {
		return zapcore.NewConsoleEncoder(cfg)
	}
	return zapcore.NewJSONEncoder(cfg)
}

func newLeveledLoggers(basePath string, cfg Config) (*leveledLoggers, error) {
	if err := os.MkdirAll(filepath.Dir(basePath), 0o755); err != nil {
		return nil, fmt.Errorf("create log directory: %w", err)
	}

	build := func(path string) *lumberjack.Logger {
		return &lumberjack.Logger{
			Filename:   path,
			MaxSize:    cfg.MaxSizeMB,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAgeDays,
			Compress:   cfg.Compress,
			LocalTime:  true,
		}
	}

	return &leveledLoggers{
		debug: build(appendLevelSuffix(basePath, "debug")),
		info:  build(appendLevelSuffix(basePath, "info")),
		warn:  build(appendLevelSuffix(basePath, "warn")),
		error: build(appendLevelSuffix(basePath, "error")),
	}, nil
}

func resolveLogPath(pathValue, serviceName string) (string, error) {
	cleaned := filepath.Clean(pathValue)
	if cleaned == "." || cleaned == string(filepath.Separator) || filepath.Ext(cleaned) == "" {
		return filepath.Join(cleaned, serviceName+".log"), nil
	}
	return cleaned, nil
}

func appendLevelSuffix(basePath, level string) string {
	ext := filepath.Ext(basePath)
	if ext == "" {
		return basePath + "." + level
	}
	return strings.TrimSuffix(basePath, ext) + "." + level + ext
}

func exactLevelEnabler(global zap.AtomicLevel, target zapcore.Level) zap.LevelEnablerFunc {
	return func(level zapcore.Level) bool {
		return global.Enabled(level) && level == target
	}
}

func errorLevelEnabler(global zap.AtomicLevel) zap.LevelEnablerFunc {
	return func(level zapcore.Level) bool {
		return global.Enabled(level) && level >= zap.ErrorLevel
	}
}

func isIgnorableSyncError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "invalid argument") || strings.Contains(msg, "inappropriate ioctl for device")
}
