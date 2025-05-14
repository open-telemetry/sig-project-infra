// SPDX-License-Identifier: Apache-2.0

package server

import (
	"context"
	"net/http"
	"sync"
)

// MockServer is a mock implementation of Provider for testing.
type MockServer struct {
	StartCalled       bool
	ShutdownCalled    bool
	Events            []MockEvent
	EventsMutex       sync.RWMutex
	StartFunc         func() error
	ShutdownFunc      func(ctx context.Context) error
	Started           bool
	Stopped           bool
	ReturnStartErr    error
	ReturnShutdownErr error
}

// MockEvent represents a dispatched event for testing.
type MockEvent struct {
	Type  string
	Event any
	Raw   []byte
}

// Ensure MockServer implements Provider.
var _ Provider = (*MockServer)(nil)

// NewMockServer creates a new mock server for testing.
func NewMockServer() *MockServer {
	return &MockServer{
		Events:       make([]MockEvent, 0),
		StartFunc:    func() error { return nil },
		ShutdownFunc: func(ctx context.Context) error { return nil },
	}
}

// Start mocks the server start operation.
func (s *MockServer) Start() error {
	s.StartCalled = true
	s.Started = true
	return s.ReturnStartErr
}

// Shutdown mocks the server shutdown operation.
func (s *MockServer) Shutdown(ctx context.Context) error {
	s.ShutdownCalled = true
	s.Stopped = true
	return s.ReturnShutdownErr
}

// DispatchEvent records an event dispatch for testing.
func (s *MockServer) DispatchEvent(eventType string, event any, raw []byte) {
	s.EventsMutex.Lock()
	defer s.EventsMutex.Unlock()
	s.Events = append(s.Events, MockEvent{
		Type:  eventType,
		Event: event,
		Raw:   raw,
	})
}

// ResetCalls resets the call flags for testing.
func (s *MockServer) ResetCalls() {
	s.StartCalled = false
	s.ShutdownCalled = false
	s.EventsMutex.Lock()
	s.Events = make([]MockEvent, 0)
	s.EventsMutex.Unlock()
}

// GetLastEvent returns the last dispatched event.
func (s *MockServer) GetLastEvent() (MockEvent, bool) {
	s.EventsMutex.RLock()
	defer s.EventsMutex.RUnlock()

	if len(s.Events) == 0 {
		return MockEvent{}, false
	}

	return s.Events[len(s.Events)-1], true
}

// GetEvents returns all dispatched events.
func (s *MockServer) GetEvents() []MockEvent {
	s.EventsMutex.RLock()
	defer s.EventsMutex.RUnlock()

	// Make a copy to avoid race conditions
	events := make([]MockEvent, len(s.Events))
	copy(events, s.Events)

	return events
}

// HandleRequest mocks handling an HTTP request for testing.
func (s *MockServer) HandleRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
