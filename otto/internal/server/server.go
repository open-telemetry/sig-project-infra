// SPDX-License-Identifier: Apache-2.0

// Package server provides HTTP server functionality for Otto.
package server

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v71/github"
	"github.com/open-telemetry/sig-project-infra/otto/internal/database"
	"github.com/open-telemetry/sig-project-infra/otto/internal/telemetry"
)

// Provider defines the interface for HTTP server operations.
type Provider interface {
	// Start runs the HTTP server (blocking).
	Start() error
	// Shutdown gracefully stops the server.
	Shutdown(ctx context.Context) error
}

// EventDispatcher defines the interface for dispatching events to modules.
type EventDispatcher interface {
	// DispatchEvent dispatches an event to all modules.
	DispatchEvent(eventType string, event any, raw []byte)
}

// Config contains configuration for the HTTP server.
type Config struct {
	// Address to listen on (e.g., "8080")
	Address string
	// WebhookSecret used to verify GitHub webhook signatures
	WebhookSecret string
}

// HTTPServer implements Provider for HTTP.
type HTTPServer struct {
	config        Config
	mux           *http.ServeMux
	server        *http.Server
	dispatcher    EventDispatcher
	telemetry     telemetry.Provider
	database      database.Provider
	webhookSecret []byte
}

// Ensure HTTPServer implements Provider.
var _ Provider = (*HTTPServer)(nil)

// NewHTTPServer creates a new HTTP server.
func NewHTTPServer(
	config Config,
	telemetryProvider telemetry.Provider,
	dbProvider database.Provider,
	dispatcher EventDispatcher,
) Provider {
	mux := http.NewServeMux()

	srv := &HTTPServer{
		config:        config,
		webhookSecret: []byte(config.WebhookSecret),
		mux:           mux,
		server: &http.Server{
			Addr:              fmt.Sprintf(":%v", config.Address),
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		},
		dispatcher: dispatcher,
		telemetry:  telemetryProvider,
		database:   dbProvider,
	}

	// Register handlers
	mux.HandleFunc("/webhook", srv.handleWebhook)

	// Health check endpoints
	mux.HandleFunc("/check/liveness", srv.handleLivenessCheck)   // Kubernetes liveness probe
	mux.HandleFunc("/check/readiness", srv.handleReadinessCheck) // Kubernetes readiness probe

	return srv
}

// handleLivenessCheck implements a Kubernetes liveness probe.
// It returns healthy if the server is running and can accept requests.
func (s *HTTPServer) handleLivenessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`{"status":"UP"}`))
	if err != nil {
		s.telemetry.GetLogger().Error("Failed to write liveness response", "error", err)
	}
}

// handleReadinessCheck implements a Kubernetes readiness probe.
// It checks if the server is ready to accept traffic by verifying database connectivity.
func (s *HTTPServer) handleReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// Check database connectivity
	if s.database != nil {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		err := s.database.Ping(ctx)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, writeErr := w.Write(
				[]byte(`{"status":"DOWN","details":"Database connection failed"}`),
			)
			if writeErr != nil {
				s.telemetry.GetLogger().Error("Failed to write readiness failure response", "error", writeErr)
			}
			return
		}
	}

	// All checks passed
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`{"status":"UP"}`))
	if err != nil {
		s.telemetry.GetLogger().Error("Failed to write readiness response", "error", err)
	}
}

// handleWebhook verifies signature and decodes GitHub webhook request.
func (s *HTTPServer) handleWebhook(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	eventType := github.WebHookType(r)
	ctx, span := s.telemetry.StartServerEventSpan(r.Context(), eventType)
	defer span.End()
	s.telemetry.IncServerRequest(ctx, "webhook")
	s.telemetry.IncServerWebhook(ctx, eventType)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		s.telemetry.IncServerError(ctx, "webhook", "readBody")
		s.telemetry.RecordServerLatency(
			ctx,
			"webhook",
			float64(time.Since(start).Milliseconds()),
		)
		http.Error(w, "could not read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	sig := r.Header.Get("X-Hub-Signature-256")
	if !s.verifySignature(payload, sig) {
		s.telemetry.IncServerError(ctx, "webhook", "badSig")
		s.telemetry.RecordServerLatency(
			ctx,
			"webhook",
			float64(time.Since(start).Milliseconds()),
		)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	eventType = github.WebHookType(r)
	event, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		s.telemetry.IncServerError(ctx, "webhook", "parseEvent")
		s.telemetry.RecordServerLatency(
			ctx,
			"webhook",
			float64(time.Since(start).Milliseconds()),
		)
		http.Error(w, "could not parse event", http.StatusBadRequest)
		return
	}

	s.telemetry.GetLogger().Info("received event",
		"type", eventType,
		"struct", fmt.Sprintf("%T", event))

	// Dispatch event to all modules
	if s.dispatcher != nil {
		s.dispatcher.DispatchEvent(eventType, event, payload)
	} else {
		s.telemetry.GetLogger().Error("No dispatcher in server, event dispatch failed")
	}

	s.telemetry.RecordServerLatency(ctx, "webhook", float64(time.Since(start).Milliseconds()))
	w.WriteHeader(http.StatusOK)
}

// verifySignature checks the request payload using the shared secret (GitHub webhook HMAC SHA256).
func (s *HTTPServer) verifySignature(payload []byte, sig string) bool {
	if !strings.HasPrefix(sig, "sha256=") {
		return false
	}
	sig = strings.TrimPrefix(sig, "sha256=")
	mac := hmac.New(sha256.New, s.webhookSecret)
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	receivedMAC, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(receivedMAC, expectedMAC) == 1
}

// Start runs the HTTP server (blocking).
func (s *HTTPServer) Start() error {
	s.telemetry.GetLogger().Info("starting server", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
