// SPDX-License-Identifier: Apache-2.0

// Package internal provides shared infrastructure for otto.
// telemetry.go sets up OpenTelemetry metrics, traces, and logs, bridging slog.
package internal

import (
	"context"
	"fmt"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

// InitMetrics initializes all metrics for the TelemetryManager.
func (t *TelemetryManager) InitMetrics() error {
	if t.metricsInitialized {
		return nil
	}

	meter := t.Meter()
	var err error

	// Server metrics
	t.ServerRequests, err = meter.Int64Counter(
		"otto.server.requests_total",
		metric.WithDescription("Total HTTP requests"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server requests counter: %w", err)
	}

	t.ServerWebhooks, err = meter.Int64Counter(
		"otto.server.webhooks_total",
		metric.WithDescription("Webhooks received"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server webhooks counter: %w", err)
	}

	t.ServerErrors, err = meter.Int64Counter(
		"otto.server.errors_total",
		metric.WithDescription("Server errors"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server errors counter: %w", err)
	}

	t.ServerLatencyHistogram, err = meter.Float64Histogram(
		"otto.server.request_latency_ms",
		metric.WithDescription("Request latency (ms)"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server latency histogram: %w", err)
	}

	// Module metrics
	t.ModuleCommands, err = meter.Int64Counter(
		"otto.module.commands_total",
		metric.WithDescription("Module command invocations"),
	)
	if err != nil {
		return fmt.Errorf("failed to create module commands counter: %w", err)
	}

	t.ModuleErrors, err = meter.Int64Counter(
		"otto.module.errors_total",
		metric.WithDescription("Module errors"),
	)
	if err != nil {
		return fmt.Errorf("failed to create module errors counter: %w", err)
	}

	t.ModuleAckLatency, err = meter.Float64Histogram(
		"otto.module.ack_latency_ms",
		metric.WithDescription("Latency from issue to ack (ms)"),
	)
	if err != nil {
		return fmt.Errorf("failed to create module ack latency histogram: %w", err)
	}

	t.metricsInitialized = true
	return nil
}

// IncServerRequest records an HTTP request in server metrics.
func (t *TelemetryManager) IncServerRequest(ctx context.Context, handler string) {
	t.ServerRequests.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", handler)))
}

// IncServerWebhook records a webhook event in server metrics.
func (t *TelemetryManager) IncServerWebhook(ctx context.Context, eventType string) {
	t.ServerWebhooks.Add(ctx, 1, metric.WithAttributes(attribute.String("event_type", eventType)))
}

// IncServerError records a server error in metrics.
func (t *TelemetryManager) IncServerError(ctx context.Context, handler string, errType string) {
	t.ServerErrors.Add(
		ctx,
		1,
		metric.WithAttributes(
			attribute.String("handler", handler),
			attribute.String("err_type", errType),
		),
	)
}

// RecordServerLatency records server request latency.
func (t *TelemetryManager) RecordServerLatency(ctx context.Context, handler string, ms float64) {
	t.ServerLatencyHistogram.Record(
		ctx,
		ms,
		metric.WithAttributes(attribute.String("handler", handler)),
	)
}

// IncModuleCommand records a module command execution in metrics.
func (t *TelemetryManager) IncModuleCommand(ctx context.Context, module, command string) {
	t.ModuleCommands.Add(
		ctx,
		1,
		metric.WithAttributes(
			attribute.String("module", module),
			attribute.String("command", command),
		),
	)
}

// IncModuleError records a module error in metrics.
func (t *TelemetryManager) IncModuleError(ctx context.Context, module, errType string) {
	t.ModuleErrors.Add(
		ctx,
		1,
		metric.WithAttributes(
			attribute.String("module", module),
			attribute.String("err_type", errType),
		),
	)
}

// RecordAckLatency records module acknowledgment latency.
func (t *TelemetryManager) RecordAckLatency(ctx context.Context, module string, ms float64) {
	t.ModuleAckLatency.Record(ctx, ms, metric.WithAttributes(attribute.String("module", module)))
}

// StartServerEventSpan creates a new tracing span for server event handling.
func (t *TelemetryManager) StartServerEventSpan(
	ctx context.Context,
	eventType string,
) (context.Context, trace.Span) {
	return t.Tracer().Start(ctx, "server.handle_"+eventType)
}

// StartModuleCommandSpan creates a new tracing span for module command execution.
func (t *TelemetryManager) StartModuleCommandSpan(
	ctx context.Context,
	module, command string,
) (context.Context, trace.Span) {
	return t.Tracer().Start(ctx, "module."+module+"."+command)
}

// TelemetryManager encapsulates OpenTelemetry telemetry components.
type TelemetryManager struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
	LoggerProvider *sdklog.LoggerProvider
	Logger         *slog.Logger

	// Server metrics
	ServerRequests         metric.Int64Counter
	ServerWebhooks         metric.Int64Counter
	ServerErrors           metric.Int64Counter
	ServerLatencyHistogram metric.Float64Histogram

	// Module metrics
	ModuleCommands   metric.Int64Counter
	ModuleErrors     metric.Int64Counter
	ModuleAckLatency metric.Float64Histogram

	metricsInitialized bool
}

// NewTelemetryManager creates a new telemetry manager with OpenTelemetry components.
func NewTelemetryManager(ctx context.Context) (*TelemetryManager, error) {
	// Create resource
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("otto"),
			semconv.ServiceVersion("dev"), // TODO: wire in a build flag for version
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize otel resource: %w", err)
	}

	// Create trace components
	traceExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create otlp trace exporter: %w", err)
	}
	traceProcessor := sdktrace.NewBatchSpanProcessor(traceExporter)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(traceProcessor),
	)

	// Create metric components
	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create otlp metric exporter: %w", err)
	}
	metricProcessor := sdkmetric.NewPeriodicReader(metricExporter)
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(metricProcessor),
	)

	// Create log components
	logExporter, err := otlploghttp.New(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create otlp log exporter: %w", err)
	}
	loggerProcessor := sdklog.NewBatchProcessor(logExporter)
	loggerProvider := sdklog.NewLoggerProvider(
		sdklog.WithResource(res),
		sdklog.WithProcessor(loggerProcessor),
	)

	// Use the global provider registry for OpenTelemetry itself
	otel.SetTracerProvider(tracerProvider)
	otel.SetMeterProvider(meterProvider)
	global.SetLoggerProvider(loggerProvider)

	// Create slog bridge
	handler := otelslog.NewHandler("otto")
	logger := slog.New(handler)
	slog.SetDefault(logger)

	// Create telemetry manager
	telemetry := &TelemetryManager{
		TracerProvider: tracerProvider,
		MeterProvider:  meterProvider,
		LoggerProvider: loggerProvider,
		Logger:         logger,
	}

	// Initialize metrics
	if err := telemetry.InitMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	slog.Info("[otto] OpenTelemetry (trace, metric, log+slog bridge) initialized")
	return telemetry, nil
}

// Tracer returns the tracer for Otto modules.
func (t *TelemetryManager) Tracer() trace.Tracer {
	return t.TracerProvider.Tracer("otto")
}

// Meter returns the meter for Otto modules.
func (t *TelemetryManager) Meter() metric.Meter {
	return t.MeterProvider.Meter("otto")
}

// Shutdown shuts down all telemetry providers.
func (t *TelemetryManager) Shutdown(ctx context.Context) error {
	if t.TracerProvider != nil {
		if err := t.TracerProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	if t.MeterProvider != nil {
		if err := t.MeterProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	if t.LoggerProvider != nil {
		if err := t.LoggerProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}
