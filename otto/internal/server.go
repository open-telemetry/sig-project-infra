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
			Addr:    fmt.Sprintf(":%v", addr),
			Handler: mux,
		},
		app: app,
	}
	mux.HandleFunc("/webhook", srv.handleWebhook)
	mux.HandleFunc("/healthz", handleHealthz)
	return srv
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte(`ok`))
	if err != nil {
		slog.Error("Failed to write response", "error", err)
	}
}

// handleWebhook verifies signature and decodes GitHub webhook request.
func (s *Server) handleWebhook(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	eventType := github.WebHookType(r)
	ctx, span := StartServerEventSpan(r.Context(), eventType)
	defer span.End()
	IncServerRequest(ctx, "webhook")
	IncServerWebhook(ctx, eventType)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		IncServerError(ctx, "webhook", "readBody")
		RecordServerLatency(ctx, "webhook", float64(time.Since(start).Milliseconds()))
		http.Error(w, "could not read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	sig := r.Header.Get("X-Hub-Signature-256")
	if !s.verifySignature(payload, sig) {
		IncServerError(ctx, "webhook", "badSig")
		RecordServerLatency(ctx, "webhook", float64(time.Since(start).Milliseconds()))
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	eventType = github.WebHookType(r)
	event, err := github.ParseWebHook(eventType, payload)
	if err != nil {
		IncServerError(ctx, "webhook", "parseEvent")
		RecordServerLatency(ctx, "webhook", float64(time.Since(start).Milliseconds()))
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

	RecordServerLatency(ctx, "webhook", float64(time.Since(start).Milliseconds()))
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
