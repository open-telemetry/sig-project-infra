// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
)

type mockModule struct {
	name    string
	handled int32
	eventWG *sync.WaitGroup
}

func (m *mockModule) Name() string { return m.name }
func (m *mockModule) HandleEvent(eventType string, event any, raw json.RawMessage) error {
	atomic.AddInt32(&m.handled, 1)
	if m.eventWG != nil {
		m.eventWG.Done()
	}
	return nil
}

func TestRegisterModuleAndDispatch(t *testing.T) {
	var evWG sync.WaitGroup
	mod := &mockModule{name: "testmod", eventWG: &evWG}

	// Create a test app with a module registry
	app := &App{
		ModuleRegistry: NewModuleRegistry(),
	}

	// Register the module with the app
	app.RegisterModule(mod)

	evWG.Add(1)

	// Use app to dispatch events
	app.DispatchEvent("fake", struct{}{}, nil)

	evWG.Wait()

	if atomic.LoadInt32(&mod.handled) < 1 {
		t.Fatalf("module did not handle the event")
	}
}
