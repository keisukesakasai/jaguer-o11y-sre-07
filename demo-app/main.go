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
	tracer = otel.Tracer("jaguer-demo")
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
	processing(ctx, w)
}

func processing(ctx context.Context, w http.ResponseWriter) {
	ctx, span := tracer.Start(ctx, "processing...")
	logger := logging.GetLoggerWithTraceID(ctx)
	logger.Infof("processing...")
	defer span.End()

	if rand.Float64() < 1.0/10.0 {
		funcAbnormal(ctx, w)
	} else {
		funcNormal(ctx, w)
	}
}

func funcNormal(ctx context.Context, w http.ResponseWriter) {
	ctx, span := tracer.Start(ctx, "funcNormal")
	defer span.End()
	logger := logging.GetLoggerWithTraceID(ctx)
	logger.Infof("funcNormal")
	time.Sleep(10 * time.Millisecond)
	w.WriteHeader(http.StatusOK)
}

func funcAbnormal(ctx context.Context, w http.ResponseWriter) {
	ctx, span := tracer.Start(ctx, "funcAbNormal(Oh...taking a lot of time...)")
	defer span.End()
	logger := logging.GetLoggerWithTraceID(ctx)
	logger.Infof("funcAbNormal(Oh...taking a lot of time...)")
	time.Sleep(1 * time.Second)
	logger.Infof("something is happening...")
	time.Sleep(1 * time.Second)
	logger.Errorf("I don’t know anything...Error occurred...")
	time.Sleep(1 * time.Second)
	w.WriteHeader(http.StatusInternalServerError)
}
