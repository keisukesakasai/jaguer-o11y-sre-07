package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"

	logging "demo-app/internal/log"
	tracing "demo-app/internal/trace"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
)

var (
	tracer = otel.Tracer("jaguar-demo")
)

func main() {
	// OpenTelemetry Traces
	tracerProvider, err := tracing.InitTracer()
	if err != nil {
		log.Fatalf("Error setting up trace provider: %v", err)
	}
	defer func() { _ = tracerProvider.Shutdown(context.Background()) }()

	otelHandler := otelhttp.NewHandler(http.HandlerFunc(mainHandler), "/")
	http.Handle("/", otelHandler)
	log.Fatalln(http.ListenAndServe(":8080", nil))
}

func mainHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	_, span := tracer.Start(ctx, "main handler")
	defer span.End()
	logger := logging.GetLoggerWithTraceID(ctx)
	logger.Infof("main handler")

	time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
	prosessing(ctx)
}

func prosessing(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "processing...")
	logger := logging.GetLoggerWithTraceID(ctx)
	logger.Infof("processing...")
	defer span.End()

	if rand.Float64() < 1.0/100.0 {
		funcAbnormal(ctx)
	} else {
		funcNormal(ctx)
	}
}

func funcNormal(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "funcNormal")
	defer span.End()
	logger := logging.GetLoggerWithTraceID(ctx)
	logger.Infof("funcNormal")
	time.Sleep(10 * time.Millisecond)
}

func funcAbnormal(ctx context.Context) {
	ctx, span := tracer.Start(ctx, "funcAbNormal(Oh...taking a lot of time...)")
	defer span.End()
	logger := logging.GetLoggerWithTraceID(ctx)
	logger.Infof("funcAbNormal(Oh...taking a lot of time...)")
	time.Sleep(3 * time.Second)
}
