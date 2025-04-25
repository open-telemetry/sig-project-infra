// SPDX-License-Identifier: Apache-2.0

// Package internal provides shared infrastructure for otto.
//
// telemetry.go sets up OpenTelemetry metrics, traces, and logs, bridging slog.

package internal

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel"
	otelmetric "go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

func InitOttoMetrics() {
	metricsOnce.Do(func() {
		meter := OttoMeter()
		var err error
		serverRequests, err = meter.Int64Counter("otto.server.requests_total", metric.WithDescription("Total HTTP requests"))
		if err != nil {
			panic(err)
		}
		serverWebhooks, err = meter.Int64Counter("otto.server.webhooks_total", metric.WithDescription("Webhooks received"))
		if err != nil {
			panic(err)
		}
		serverErrors, err = meter.Int64Counter("otto.server.errors_total", metric.WithDescription("Server errors"))
		if err != nil {
			panic(err)
		}
		serverLatencyHistogram, err = meter.Float64Histogram("otto.server.request_latency_ms", metric.WithDescription("Request latency (ms)"))
		if err != nil {
			panic(err)
		}
		moduleCommands, err = meter.Int64Counter("otto.module.commands_total", metric.WithDescription("Module command invocations"))
		if err != nil {
			panic(err)
		}
		moduleErrors, err = meter.Int64Counter("otto.module.errors_total", metric.WithDescription("Module errors"))
		if err != nil {
			panic(err)
		}
		moduleAckLatency, err = meter.Float64Histogram("otto.module.ack_latency_ms", metric.WithDescription("Latency from issue to ack (ms)"))
		if err != nil {
			panic(err)
		}
	})
}

// Server metrics helpers
func IncServerRequest(ctx context.Context, handler string) {
	serverRequests.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", handler)))
}
func IncServerWebhook(ctx context.Context, eventType string) {
	serverWebhooks.Add(ctx, 1, metric.WithAttributes(attribute.String("event_type", eventType)))
}
func IncServerError(ctx context.Context, handler string, errType string) {
	serverErrors.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", handler), attribute.String("err_type", errType)))
}
func RecordServerLatency(ctx context.Context, handler string, ms float64) {
	serverLatencyHistogram.Record(ctx, ms, metric.WithAttributes(attribute.String("handler", handler)))
}

// Module metrics helpers
func IncModuleCommand(ctx context.Context, module, command string) {
	moduleCommands.Add(ctx, 1, metric.WithAttributes(attribute.String("module", module), attribute.String("command", command)))
}
func IncModuleError(ctx context.Context, module, errType string) {
	moduleErrors.Add(ctx, 1, metric.WithAttributes(attribute.String("module", module), attribute.String("err_type", errType)))
}
func RecordAckLatency(ctx context.Context, module string, ms float64) {
	moduleAckLatency.Record(ctx, ms, metric.WithAttributes(attribute.String("module", module)))
}

// Tracing helpers
func StartServerEventSpan(ctx context.Context, eventType string) (context.Context, trace.Span) {
	return OttoTracer().Start(ctx, "server.handle_"+eventType)
}
func StartModuleCommandSpan(ctx context.Context, module, command string) (context.Context, trace.Span) {
	return OttoTracer().Start(ctx, "module."+module+"."+command)
}

var (
	otelTracerProvider *sdktrace.TracerProvider
	otelMeterProvider  *sdkmetric.MeterProvider
	otelLoggerProvider *sdklog.LoggerProvider
	rootLogger         *slog.Logger

	metricsOnce            sync.Once
	serverRequests         metric.Int64Counter
	serverWebhooks         metric.Int64Counter
	serverErrors           metric.Int64Counter
	serverLatencyHistogram metric.Float64Histogram

	moduleCommands   metric.Int64Counter
	moduleErrors     metric.Int64Counter
	moduleAckLatency metric.Float64Histogram
)

// InitTelemetry configures global OpenTelemetry providers for Otto,
// including traces, metrics, logs, and slog bridge.
func InitTelemetry(ctx context.Context) error {
	InitOttoMetrics()
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("otto"),
			semconv.ServiceVersion("dev"), // TODO: wire in a build flag for version
		),
	)
	if err != nil {
		return fmt.Errorf("failed to initialize otel resource: %w", err)
	}

	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create otlp trace exporter: %w", err)
	}
	traceProcessor := sdktrace.NewBatchSpanProcessor(traceExporter)

	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create otlp metric exporter: %w", err)
	}
	metricProcessor := sdkmetric.NewPeriodicReader(metricExporter)

	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		return fmt.Errorf("failed to create otlp log exporter: %w", err)
	}
	loggerProcessor := sdklog.NewBatchProcessor(logExporter)

	otelTracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(traceProcessor),
	)
	otelMeterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(metricProcessor),
	)
	otelLoggerProvider = sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(loggerProcessor),
	)
	global.SetLoggerProvider(otelLoggerProvider)

	// Bridge slog to OpenTelemetry logging
	handler := otelslog.NewHandler("otto")
	rootLogger = slog.New(handler)
	slog.SetDefault(rootLogger)

	otel.SetTracerProvider(otelTracerProvider)
	otel.SetMeterProvider(otelMeterProvider)
	slog.Info("[otto] OpenTelemetry (trace, metric, log+slog bridge) initialized")
	return nil
}

// OttoTracer returns the tracer for Otto modules.
func OttoTracer() oteltrace.Tracer {
	return otel.Tracer("otto")
}

// OttoMeter returns the meter for Otto modules.
func OttoMeter() otelmetric.Meter {
	return otel.Meter("otto")
}

// RootSlogLogger returns the bridged *slog.Logger for app use.
func RootSlogLogger() *slog.Logger {
	return rootLogger
}

// ShutdownTelemetry shuts down all providers.
func ShutdownTelemetry(ctx context.Context) error {
	if otelTracerProvider != nil {
		if err := otelTracerProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	if otelMeterProvider != nil {
		if err := otelMeterProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	if otelLoggerProvider != nil {
		if err := otelLoggerProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}
