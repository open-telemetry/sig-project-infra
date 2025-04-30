// SPDX-License-Identifier: Apache-2.0

// server.go hosts Otto's HTTP webhook endpoint(s) using standard net/http.

package internal

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v71/github"
	"github.com/open-telemetry/sig-project-infra/otto/internal/secrets"
)

type Server struct {
	webhookSecret []byte // from secrets config
	mux           *http.ServeMux
	server        *http.Server
	app           *App // Reference to the app for dispatching events
}

// NewServer creates a new server with the provided webhook secret and address.
func NewServer(addr string, secretsManager secrets.Manager) *Server {
	return NewServerWithApp(addr, secretsManager, nil)
}

// NewServerWithApp creates a server with a reference to the app.
func NewServerWithApp(addr string, secretsManager secrets.Manager, app *App) *Server {
	mux := http.NewServeMux()
	srv := &Server{
		webhookSecret: []byte(secretsManager.GetWebhookSecret()),
		mux:           mux,
		server: &http.Server{
			Addr:              fmt.Sprintf(":%v", addr),
			Handler:           mux,
			ReadHeaderTimeout: 10 * time.Second,
		},
		app: app,
	}
	mux.HandleFunc("/webhook", srv.handleWebhook)

	// Health check endpoints
	mux.HandleFunc("/check/liveness", srv.handleLivenessCheck)   // Kubernetes liveness probe
	mux.HandleFunc("/check/readiness", srv.handleReadinessCheck) // Kubernetes readiness probe

	return srv
}

// handleLivenessCheck implements a Kubernetes liveness probe.
// It returns healthy if the server is running and can accept requests.
func (s *Server) handleLivenessCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`{"status":"UP"}`))
	if err != nil {
		slog.Error("Failed to write liveness response", "error", err)
	}
}

// handleReadinessCheck implements a Kubernetes readiness probe.
// It checks if the server is ready to accept traffic by verifying database connectivity.
func (s *Server) handleReadinessCheck(w http.ResponseWriter, r *http.Request) {
	// Check if app reference exists
	if s.app == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, err := w.Write([]byte(`{"status":"DOWN","details":"App not initialized"}`))
		if err != nil {
			slog.Error("Failed to write readiness failure response", "error", err)
		}
		return
	}

	// Check database connectivity if database exists
	if s.app.Database != nil {
		err := s.app.Database.DB().Ping()
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			_, writeErr := w.Write(
				[]byte(`{"status":"DOWN","details":"Database connection failed"}`),
			)
			if writeErr != nil {
				slog.Error("Failed to write readiness failure response", "error", writeErr)
			}
			return
		}
	}

	// All checks passed
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`{"status":"UP"}`))
	if err != nil {
		slog.Error("Failed to write readiness response", "error", err)
	}
}

// handleWebhook verifies signature and decodes GitHub webhook request.
func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	eventType := github.WebHookType(r)
	ctx, span := s.app.Telemetry.StartServerEventSpan(r.Context(), eventType)
	defer span.End()
	s.app.Telemetry.IncServerRequest(ctx, "webhook")
	s.app.Telemetry.IncServerWebhook(ctx, eventType)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		s.app.Telemetry.IncServerError(ctx, "webhook", "readBody")
		s.app.Telemetry.RecordServerLatency(
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
		s.app.Telemetry.IncServerError(ctx, "webhook", "badSig")
		s.app.Telemetry.RecordServerLatency(
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
		s.app.Telemetry.IncServerError(ctx, "webhook", "parseEvent")
		s.app.Telemetry.RecordServerLatency(
			ctx,
			"webhook",
			float64(time.Since(start).Milliseconds()),
		)
		http.Error(w, "could not parse event", http.StatusBadRequest)
		return
	}

	slog.Info("received event",
		"type", eventType,
		"struct", fmt.Sprintf("%T", event))

	// Dispatch event to all modules
	if s.app != nil {
		s.app.DispatchEvent(eventType, event, payload)
	} else {
		slog.Error("No app reference in server, event dispatch failed")
	}

	s.app.Telemetry.RecordServerLatency(ctx, "webhook", float64(time.Since(start).Milliseconds()))
	w.WriteHeader(http.StatusOK)
}

// verifySignature checks the request payload using the shared secret (GitHub webhook HMAC SHA256).
func (s *Server) verifySignature(payload []byte, sig string) bool {
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
func (s *Server) Start() error {
	slog.Info("starting server", "addr", s.server.Addr)
	return s.server.ListenAndServe()
}

// Shutdown gracefully stops the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
