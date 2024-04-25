package logging

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	logLevelEnv              = "LOG_LEVEL"
	appVersionEnv            = "APP_VERSION"
	serviceNameEnv           = "SERVICE_NAME"
	OtelCollectorEndpointEnv = "OTEL_COLLECTOR_ENDPOINT"
	traceKey                 = "trace_id"
	spanKey                  = "span_id"
)

type contextKeyLoggerKey int

const (
	contextKeyLogger contextKeyLoggerKey = iota
)

var (
	logLevel    zapcore.Level
	localLogger *zap.SugaredLogger
)

type config struct {
	LogLevel   zapcore.Level
	AppVersion string
	Service    string
}

func NewLogger() *zap.SugaredLogger {
	return localLogger
}

func init() {
	logLevel := os.Getenv(logLevelEnv)
	appVersion := os.Getenv(appVersionEnv)
	serviceName := os.Getenv(serviceNameEnv)

	configure(config{
		LogLevel:   getZapLogLevelFromEnv(logLevel),
		AppVersion: appVersion,
		Service:    serviceName,
	})
}

func configure(config config) {
	zapConfig := defaultZapConfig()

	logger, _ := zapConfig.Build()
	fields := zap.Fields([]zap.Field{
		zap.String(appVersionEnv, config.AppVersion),
		zap.String(serviceNameEnv, config.Service),
	}...)

	localLogger = logger.WithOptions(fields).Sugar()
}

func defaultZapConfig() zap.Config {
	return zap.Config{
		Level:    zap.NewAtomicLevelAt(logLevel),
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:     "message",
			LevelKey:       "severity",
			TimeKey:        "time",
			NameKey:        "logger",
			CallerKey:      "caller",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.RFC3339NanoTimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
}

func WithTrace(ctx context.Context, logger *zap.SugaredLogger) *zap.SugaredLogger {
	spanCtx := trace.SpanContextFromContext(ctx)

	if spanCtx.HasTraceID() {
		traceID := spanCtx.TraceID().String()
		logger = logger.With(traceKey, traceID)
	}

	if spanCtx.HasSpanID() {
		logger = logger.With(spanKey, spanCtx.SpanID().String())
	}
	return logger
}

func GetLoggerWithTraceID(ctx context.Context) *zap.SugaredLogger {
	logger := NewLogger()
	if spanCtx := trace.SpanContextFromContext(ctx); spanCtx.IsValid() {
		logger = logger.With("trace_id", spanCtx.TraceID().String())
		logger = logger.With("span_id", spanCtx.SpanID().String())
	}
	return logger
}

func GetLoggerFromCtx(ctx context.Context) *zap.SugaredLogger {
	logger, ok := ctx.Value(contextKeyLogger).(*zap.SugaredLogger)
	if ok {
		return logger
	}

	return NewLogger()
}

func SetLoggerToCtx(ctx context.Context, logger *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, contextKeyLogger, logger)
}
