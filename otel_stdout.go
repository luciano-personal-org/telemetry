// Package telemetry provides functionality for OpenTelemetry tracing.
package telemetry

import (
	"context"
	"time"

	"github.com/luciano-personal-org/config"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
)

// setupOTelSDK bootstraps the OpenTelemetry pipeline.
// If it does not return an error, make sure to call shutdown for proper cleanup.
func SetupOTelSDKStdout(ctx context.Context, configuration config.Config) (shutdown func(context.Context) error, err error) {

	res, err := resource.New(ctx,
		resource.WithAttributes(
			// The service name used to display traces in backends
			attribute.String("SERVICE_NAME", configuration.Get("APP_NAME")),
		),
	)
	if err != nil {
		handleErr(err, ctx)
		return
	}

	// Set up propagator.
	prop := newPropagator()
	otel.SetTextMapPropagator(prop)

	// Set up trace provider.
	tracerProvider, err := newTraceProvider(res)
	if err != nil {
		handleErr(err, ctx)
		return
	}
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			handleErr(err, ctx)
		}
	}()
	otel.SetTracerProvider(tracerProvider)

	// Set up meter provider.
	meterProvider, err := newMeterProvider(res)
	if err != nil {
		handleErr(err, ctx)
		return
	}
	defer func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			handleErr(err, ctx)
		}
	}()
	otel.SetMeterProvider(meterProvider)

	// Set up logger provider.
	// loggerProvider, err := newLoggerProvider(res)
	// if err != nil {
	// 	handleErr(err, ctx)
	// 	return
	// }
	// defer func() {
	// 	if err := loggerProvider.Shutdown(ctx); err != nil {
	// 		handleErr(err, ctx)
	// 	}
	// }()
	// global.SetLoggerProvider(loggerProvider)

	return
}

// newPropagator returns a new propagator.
func newPropagator() propagation.TextMapPropagator {
	return propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
}

// newTraceProvider returns a new trace provider.
func newTraceProvider(res *resource.Resource) (*trace.TracerProvider, error) {
	traceExporter, err := stdouttrace.New(
		stdouttrace.WithPrettyPrint())
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter,
			// Default is 5s. Set to 1s for demonstrative purposes.
			trace.WithBatchTimeout(time.Second)),
		trace.WithResource(res),
	)
	return traceProvider, nil
}

// newMeterProvider returns a new meter provider.
func newMeterProvider(res *resource.Resource) (*metric.MeterProvider, error) {
	metricExporter, err := stdoutmetric.New()
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(metric.NewPeriodicReader(metricExporter,
			// Default is 1m. Set to 3s for demonstrative purposes.
			metric.WithInterval(3*time.Second))),
		metric.WithResource(res),
	)
	return meterProvider, nil
}

// newLoggerProvider returns a new logger provider.
func newLoggerProvider(res *resource.Resource) (*log.LoggerProvider, error) {
	logExporter, err := stdoutlog.New()
	if err != nil {
		return nil, err
	}

	loggerProvider := log.NewLoggerProvider(
		log.WithProcessor(log.NewBatchProcessor(logExporter)),
		log.WithResource(res),
	)
	return loggerProvider, nil
}
