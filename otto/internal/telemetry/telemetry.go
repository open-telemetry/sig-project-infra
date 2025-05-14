// SPDX-License-Identifier: Apache-2.0

// Package telemetry provides interfaces and implementations for OpenTelemetry integration.
package telemetry

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/open-telemetry/sig-project-infra/otto/internal/logging"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/metric"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Provider defines the interface for telemetry operations.
type Provider interface {
	// GetLogger returns the application logger.
	GetLogger() logging.Logger
	// GetTracer returns the application tracer.
	GetTracer() trace.Tracer
	// GetMeter returns the application meter.
	GetMeter() metric.Meter
	// Shutdown gracefully shuts down the telemetry provider.
	Shutdown(ctx context.Context) error

	// Server metrics
	IncServerRequest(ctx context.Context, handler string)
	IncServerWebhook(ctx context.Context, eventType string)
	IncServerError(ctx context.Context, handler string, errType string)
	RecordServerLatency(ctx context.Context, handler string, ms float64)

	// Module metrics
	IncModuleCommand(ctx context.Context, module, command string)
	IncModuleError(ctx context.Context, module, errType string)
	RecordAckLatency(ctx context.Context, module string, ms float64)

	// Tracing
	StartServerEventSpan(ctx context.Context, eventType string) (context.Context, trace.Span)
	StartModuleCommandSpan(ctx context.Context, module, command string) (context.Context, trace.Span)
}

// Manager implements the Provider interface using OpenTelemetry.
type Manager struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	loggerProvider *sdklog.LoggerProvider
	slogLogger     *slog.Logger
	logger         logging.Logger

	// Server metrics
	serverRequests         metric.Int64Counter
	serverWebhooks         metric.Int64Counter
	serverErrors           metric.Int64Counter
	serverLatencyHistogram metric.Float64Histogram

	// Module metrics
	moduleCommands   metric.Int64Counter
	moduleErrors     metric.Int64Counter
	moduleAckLatency metric.Float64Histogram

	metricsInitialized bool
}

// Ensure Manager implements Provider.
var _ Provider = (*Manager)(nil)

// NewManager creates a new telemetry manager with OpenTelemetry components.
func NewManager(ctx context.Context) (Provider, error) {
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
	slogLogger := slog.New(handler)
	slog.SetDefault(slogLogger)

	// Create logger adapter
	logger := logging.NewSlogLogger(slogLogger)

	// Create telemetry manager
	telemetry := &Manager{
		tracerProvider: tracerProvider,
		meterProvider:  meterProvider,
		loggerProvider: loggerProvider,
		slogLogger:     slogLogger,
		logger:         logger,
	}

	// Initialize metrics
	if err := telemetry.initMetrics(); err != nil {
		return nil, fmt.Errorf("failed to initialize metrics: %w", err)
	}

	logger.Info("[otto] OpenTelemetry (trace, metric, log+slog bridge) initialized")
	return telemetry, nil
}

// initMetrics initializes all metrics for the Manager.
func (t *Manager) initMetrics() error {
	if t.metricsInitialized {
		return nil
	}

	meter := t.GetMeter()
	var err error

	// Server metrics
	t.serverRequests, err = meter.Int64Counter(
		"otto.server.requests_total",
		metric.WithDescription("Total HTTP requests"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server requests counter: %w", err)
	}

	t.serverWebhooks, err = meter.Int64Counter(
		"otto.server.webhooks_total",
		metric.WithDescription("Webhooks received"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server webhooks counter: %w", err)
	}

	t.serverErrors, err = meter.Int64Counter(
		"otto.server.errors_total",
		metric.WithDescription("Server errors"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server errors counter: %w", err)
	}

	t.serverLatencyHistogram, err = meter.Float64Histogram(
		"otto.server.request_latency_ms",
		metric.WithDescription("Request latency (ms)"),
	)
	if err != nil {
		return fmt.Errorf("failed to create server latency histogram: %w", err)
	}

	// Module metrics
	t.moduleCommands, err = meter.Int64Counter(
		"otto.module.commands_total",
		metric.WithDescription("Module command invocations"),
	)
	if err != nil {
		return fmt.Errorf("failed to create module commands counter: %w", err)
	}

	t.moduleErrors, err = meter.Int64Counter(
		"otto.module.errors_total",
		metric.WithDescription("Module errors"),
	)
	if err != nil {
		return fmt.Errorf("failed to create module errors counter: %w", err)
	}

	t.moduleAckLatency, err = meter.Float64Histogram(
		"otto.module.ack_latency_ms",
		metric.WithDescription("Latency from issue to ack (ms)"),
	)
	if err != nil {
		return fmt.Errorf("failed to create module ack latency histogram: %w", err)
	}

	t.metricsInitialized = true
	return nil
}

// GetLogger returns the application logger.
func (t *Manager) GetLogger() logging.Logger {
	return t.logger
}

// GetTracer returns the application tracer.
func (t *Manager) GetTracer() trace.Tracer {
	return t.tracerProvider.Tracer("otto")
}

// GetMeter returns the application meter.
func (t *Manager) GetMeter() metric.Meter {
	return t.meterProvider.Meter("otto")
}

// Shutdown gracefully shuts down the telemetry provider.
func (t *Manager) Shutdown(ctx context.Context) error {
	if t.tracerProvider != nil {
		if err := t.tracerProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	if t.meterProvider != nil {
		if err := t.meterProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	if t.loggerProvider != nil {
		if err := t.loggerProvider.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}

// IncServerRequest records an HTTP request in server metrics.
func (t *Manager) IncServerRequest(ctx context.Context, handler string) {
	t.serverRequests.Add(ctx, 1, metric.WithAttributes(attribute.String("handler", handler)))
}

// IncServerWebhook records a webhook event in server metrics.
func (t *Manager) IncServerWebhook(ctx context.Context, eventType string) {
	t.serverWebhooks.Add(ctx, 1, metric.WithAttributes(attribute.String("event_type", eventType)))
}

// IncServerError records a server error in metrics.
func (t *Manager) IncServerError(ctx context.Context, handler string, errType string) {
	t.serverErrors.Add(
		ctx,
		1,
		metric.WithAttributes(
			attribute.String("handler", handler),
			attribute.String("err_type", errType),
		),
	)
}

// RecordServerLatency records server request latency.
func (t *Manager) RecordServerLatency(ctx context.Context, handler string, ms float64) {
	t.serverLatencyHistogram.Record(
		ctx,
		ms,
		metric.WithAttributes(attribute.String("handler", handler)),
	)
}

// IncModuleCommand records a module command execution in metrics.
func (t *Manager) IncModuleCommand(ctx context.Context, module, command string) {
	t.moduleCommands.Add(
		ctx,
		1,
		metric.WithAttributes(
			attribute.String("module", module),
			attribute.String("command", command),
		),
	)
}

// IncModuleError records a module error in metrics.
func (t *Manager) IncModuleError(ctx context.Context, module, errType string) {
	t.moduleErrors.Add(
		ctx,
		1,
		metric.WithAttributes(
			attribute.String("module", module),
			attribute.String("err_type", errType),
		),
	)
}

// RecordAckLatency records module acknowledgment latency.
func (t *Manager) RecordAckLatency(ctx context.Context, module string, ms float64) {
	t.moduleAckLatency.Record(ctx, ms, metric.WithAttributes(attribute.String("module", module)))
}

// StartServerEventSpan creates a new tracing span for server event handling.
func (t *Manager) StartServerEventSpan(
	ctx context.Context,
	eventType string,
) (context.Context, trace.Span) {
	return t.GetTracer().Start(ctx, "server.handle_"+eventType)
}

// StartModuleCommandSpan creates a new tracing span for module command execution.
func (t *Manager) StartModuleCommandSpan(
	ctx context.Context,
	module, command string,
) (context.Context, trace.Span) {
	return t.GetTracer().Start(ctx, "module."+module+"."+command)
}
