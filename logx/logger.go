package logx

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

var (
	globalLogger *zap.SugaredLogger
	globalMu     sync.RWMutex
)

func init() {
	// Initialize with a default production logger
	logger, _ := zap.NewProduction()
	globalLogger = logger.Sugar()
}

// Init initializes the global logger with the given configuration.
// It should be called early in application startup.
func Init(cfg Config) error {
	zapCfg := cfg.toZapConfig()
	logger, err := zapCfg.Build()
	if err != nil {
		return err
	}

	globalMu.Lock()
	globalLogger = logger.Sugar()
	globalMu.Unlock()

	return nil
}

// MustInit initializes the global logger and panics on error.
func MustInit(cfg Config) {
	if err := Init(cfg); err != nil {
		panic(err)
	}
}

// L returns the global sugared logger.
func L() *zap.SugaredLogger {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalLogger
}

// SetLogger sets the global logger.
func SetLogger(logger *zap.SugaredLogger) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalLogger = logger
}

// Sync flushes any buffered log entries.
func Sync() error {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalLogger.Sync()
}

// Debug logs a message at debug level.
func Debug(args ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Debug(args...)
}

// Debugf logs a formatted message at debug level.
func Debugf(template string, args ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Debugf(template, args...)
}

// Debugw logs a message with key-value pairs at debug level.
func Debugw(msg string, keysAndValues ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Debugw(msg, keysAndValues...)
}

// Info logs a message at info level.
func Info(args ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Info(args...)
}

// Infof logs a formatted message at info level.
func Infof(template string, args ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Infof(template, args...)
}

// Infow logs a message with key-value pairs at info level.
func Infow(msg string, keysAndValues ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Infow(msg, keysAndValues...)
}

// Warn logs a message at warn level.
func Warn(args ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Warn(args...)
}

// Warnf logs a formatted message at warn level.
func Warnf(template string, args ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Warnf(template, args...)
}

// Warnw logs a message with key-value pairs at warn level.
func Warnw(msg string, keysAndValues ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Warnw(msg, keysAndValues...)
}

// Error logs a message at error level.
func Error(args ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Error(args...)
}

// Errorf logs a formatted message at error level.
func Errorf(template string, args ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Errorf(template, args...)
}

// Errorw logs a message with key-value pairs at error level.
func Errorw(msg string, keysAndValues ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Errorw(msg, keysAndValues...)
}

// Fatal logs a message at fatal level and then calls os.Exit(1).
func Fatal(args ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Fatal(args...)
}

// Fatalf logs a formatted message at fatal level and then calls os.Exit(1).
func Fatalf(template string, args ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Fatalf(template, args...)
}

// Fatalw logs a message with key-value pairs at fatal level and then calls os.Exit(1).
func Fatalw(msg string, keysAndValues ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Fatalw(msg, keysAndValues...)
}

// Panic logs a message at panic level and then panics.
func Panic(args ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Panic(args...)
}

// Panicf logs a formatted message at panic level and then panics.
func Panicf(template string, args ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Panicf(template, args...)
}

// Panicw logs a message with key-value pairs at panic level and then panics.
func Panicw(msg string, keysAndValues ...interface{}) {
	globalMu.RLock()
	defer globalMu.RUnlock()
	globalLogger.Panicw(msg, keysAndValues...)
}

// With creates a child logger with the given key-value pairs.
func With(keysAndValues ...interface{}) *zap.SugaredLogger {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalLogger.With(keysAndValues...)
}

// Named adds a sub-scope to the logger.
func Named(name string) *zap.SugaredLogger {
	globalMu.RLock()
	defer globalMu.RUnlock()
	return globalLogger.Named(name)
}

// Context key for logger
type ctxKey struct{}

// FromContext extracts a logger from the context.
// Returns the global logger if none is found.
func FromContext(ctx context.Context) *zap.SugaredLogger {
	if logger, ok := ctx.Value(ctxKey{}).(*zap.SugaredLogger); ok {
		return logger
	}
	return L()
}

// WithContext returns a new context with the logger attached.
func WithContext(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}