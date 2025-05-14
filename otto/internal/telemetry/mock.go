// SPDX-License-Identifier: Apache-2.0

package telemetry

import (
	"context"

	"github.com/open-telemetry/sig-project-infra/otto/internal/logging"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	nooptrace "go.opentelemetry.io/otel/trace/noop"
)

// SpanEvent represents a span event for testing.
type SpanEvent struct {
	SpanName string
	Module   string
	Command  string
}

// MetricEvent represents a metric event for testing.
type MetricEvent struct {
	Name       string
	Value      float64
	Attributes map[string]string
}

// MockProvider implements Provider for testing.
type MockProvider struct {
	Logger          *logging.MockLogger
	Spans           []SpanEvent
	ServerMetrics   []MetricEvent
	ModuleMetrics   []MetricEvent
	LatencyMetrics  []MetricEvent
	TraceAttributes map[string]string
	TracerProvider  *sdktrace.TracerProvider
	SpanRecorder    *tracetest.SpanRecorder
}

// Ensure MockProvider implements Provider.
var _ Provider = (*MockProvider)(nil)

// NewMockProvider creates a new mock telemetry provider.
func NewMockProvider() *MockProvider {
	spanRecorder := tracetest.NewSpanRecorder()
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSpanProcessor(spanRecorder),
	)

	return &MockProvider{
		Logger:          logging.NewMockLogger(),
		Spans:           make([]SpanEvent, 0),
		ServerMetrics:   make([]MetricEvent, 0),
		ModuleMetrics:   make([]MetricEvent, 0),
		LatencyMetrics:  make([]MetricEvent, 0),
		TraceAttributes: make(map[string]string),
		TracerProvider:  tracerProvider,
		SpanRecorder:    spanRecorder,
	}
}

// GetLogger returns the application logger.
func (t *MockProvider) GetLogger() logging.Logger {
	return t.Logger
}

// GetTracer returns the application tracer.
func (t *MockProvider) GetTracer() trace.Tracer {
	if t.TracerProvider != nil {
		return t.TracerProvider.Tracer("mock")
	}
	return nooptrace.NewTracerProvider().Tracer("mock")
}

// GetMeter returns the application meter.
func (t *MockProvider) GetMeter() metric.Meter {
	return noop.NewMeterProvider().Meter("mock")
}

// Shutdown records that shutdown was called.
func (t *MockProvider) Shutdown(ctx context.Context) error {
	if t.TracerProvider != nil {
		return t.TracerProvider.Shutdown(ctx)
	}
	return nil
}

// IncServerRequest records a server request metric.
func (t *MockProvider) IncServerRequest(ctx context.Context, handler string) {
	t.ServerMetrics = append(t.ServerMetrics, MetricEvent{
		Name: "server.request.count",
		Attributes: map[string]string{
			"handler": handler,
		},
	})
}

// IncServerWebhook records a server webhook metric.
func (t *MockProvider) IncServerWebhook(ctx context.Context, eventType string) {
	t.ServerMetrics = append(t.ServerMetrics, MetricEvent{
		Name: "server.webhook.count",
		Attributes: map[string]string{
			"event_type": eventType,
		},
	})
}

// IncServerError records a server error metric.
func (t *MockProvider) IncServerError(ctx context.Context, handler string, errType string) {
	t.ServerMetrics = append(t.ServerMetrics, MetricEvent{
		Name: "server.error.count",
		Attributes: map[string]string{
			"handler":  handler,
			"err_type": errType,
		},
	})
}

// RecordServerLatency records a server latency metric.
func (t *MockProvider) RecordServerLatency(ctx context.Context, handler string, ms float64) {
	t.LatencyMetrics = append(t.LatencyMetrics, MetricEvent{
		Name:  "server.latency",
		Value: ms,
		Attributes: map[string]string{
			"handler": handler,
		},
	})
}

// IncModuleCommand records a module command metric.
func (t *MockProvider) IncModuleCommand(ctx context.Context, module, command string) {
	t.ModuleMetrics = append(t.ModuleMetrics, MetricEvent{
		Name: "module.command.count",
		Attributes: map[string]string{
			"module":  module,
			"command": command,
		},
	})
}

// IncModuleError records a module error metric.
func (t *MockProvider) IncModuleError(ctx context.Context, module, errType string) {
	t.ModuleMetrics = append(t.ModuleMetrics, MetricEvent{
		Name: "module.error.count",
		Attributes: map[string]string{
			"module":   module,
			"err_type": errType,
		},
	})
}

// RecordAckLatency records an acknowledgment latency metric.
func (t *MockProvider) RecordAckLatency(ctx context.Context, module string, ms float64) {
	t.LatencyMetrics = append(t.LatencyMetrics, MetricEvent{
		Name:  "module.ack.latency",
		Value: ms,
		Attributes: map[string]string{
			"module": module,
		},
	})
}

// StartServerEventSpan creates a new tracing span for server event handling.
func (t *MockProvider) StartServerEventSpan(
	ctx context.Context,
	eventType string,
) (context.Context, trace.Span) {
	t.Spans = append(t.Spans, SpanEvent{
		SpanName: "server.handle_" + eventType,
	})

	tracer := t.GetTracer()
	ctx, span := tracer.Start(ctx, "server.handle_"+eventType)
	return ctx, span
}

// StartModuleCommandSpan creates a new tracing span for module command execution.
func (t *MockProvider) StartModuleCommandSpan(
	ctx context.Context,
	module, command string,
) (context.Context, trace.Span) {
	t.Spans = append(t.Spans, SpanEvent{
		SpanName: "module." + module + "." + command,
		Module:   module,
		Command:  command,
	})

	tracer := t.GetTracer()
	ctx, span := tracer.Start(ctx, "module."+module+"."+command)
	return ctx, span
}

// Clear clears all recorded metrics and spans.
func (t *MockProvider) Clear() {
	t.Logger.Clear()
	t.Spans = make([]SpanEvent, 0)
	t.ServerMetrics = make([]MetricEvent, 0)
	t.ModuleMetrics = make([]MetricEvent, 0)
	t.LatencyMetrics = make([]MetricEvent, 0)
	t.TraceAttributes = make(map[string]string)
	if t.SpanRecorder != nil {
		t.SpanRecorder.Reset()
	}
}

// NoopProvider is a telemetry provider that does nothing.
type NoopProvider struct {
	logger logging.Logger
}

// Ensure NoopProvider implements Provider.
var _ Provider = (*NoopProvider)(nil)

// NewNoopProvider creates a new noop telemetry provider.
func NewNoopProvider() Provider {
	return &NoopProvider{
		logger: logging.NewNoopLogger(),
	}
}

// GetLogger returns the application logger.
func (t *NoopProvider) GetLogger() logging.Logger {
	return t.logger
}

// GetTracer returns the application tracer.
func (t *NoopProvider) GetTracer() trace.Tracer {
	return nooptrace.NewTracerProvider().Tracer("noop")
}

// GetMeter returns the application meter.
func (t *NoopProvider) GetMeter() metric.Meter {
	return noop.NewMeterProvider().Meter("noop")
}

// Shutdown gracefully shuts down the telemetry provider.
func (t *NoopProvider) Shutdown(ctx context.Context) error {
	return nil
}

// IncServerRequest does nothing.
func (t *NoopProvider) IncServerRequest(ctx context.Context, handler string) {}

// IncServerWebhook does nothing.
func (t *NoopProvider) IncServerWebhook(ctx context.Context, eventType string) {}

// IncServerError does nothing.
func (t *NoopProvider) IncServerError(ctx context.Context, handler string, errType string) {}

// RecordServerLatency does nothing.
func (t *NoopProvider) RecordServerLatency(ctx context.Context, handler string, ms float64) {}

// IncModuleCommand does nothing.
func (t *NoopProvider) IncModuleCommand(ctx context.Context, module, command string) {}

// IncModuleError does nothing.
func (t *NoopProvider) IncModuleError(ctx context.Context, module, errType string) {}

// RecordAckLatency does nothing.
func (t *NoopProvider) RecordAckLatency(ctx context.Context, module string, ms float64) {}

// StartServerEventSpan does nothing.
func (t *NoopProvider) StartServerEventSpan(
	ctx context.Context,
	eventType string,
) (context.Context, trace.Span) {
	tracer := t.GetTracer()
	ctx, span := tracer.Start(ctx, "server.handle_"+eventType)
	return ctx, span
}

// StartModuleCommandSpan does nothing.
func (t *NoopProvider) StartModuleCommandSpan(
	ctx context.Context,
	module, command string,
) (context.Context, trace.Span) {
	tracer := t.GetTracer()
	ctx, span := tracer.Start(ctx, "module."+module+"."+command)
	return ctx, span
}
