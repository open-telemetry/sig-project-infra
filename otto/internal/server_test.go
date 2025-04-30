// SPDX-License-Identifier: Apache-2.0

package internal

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoints(t *testing.T) {
	// Create a test server with no app (should fail readiness)
	srv := &Server{
		mux: http.NewServeMux(),
	}
	srv.mux.HandleFunc("/check/liveness", srv.handleLivenessCheck)
	srv.mux.HandleFunc("/check/readiness", srv.handleReadinessCheck)

	// Test cases
	tests := []struct {
		name           string
		endpoint       string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Liveness endpoint",
			endpoint:       "/check/liveness",
			expectedStatus: http.StatusOK,
			expectedBody:   `{"status":"UP"}`,
		},
		{
			name:           "Readiness endpoint with no app",
			endpoint:       "/check/readiness",
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   `{"status":"DOWN","details":"App not initialized"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, tc.endpoint, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			srv.mux.ServeHTTP(rr, req)

			// Check status code
			if status := rr.Code; status != tc.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tc.expectedStatus)
			}

			// Check response body
			if rr.Body.String() != tc.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tc.expectedBody)
			}
		})
	}
}

func TestReadinessWithApp(t *testing.T) {
	// Create a database for testing
	testDB, err := OpenDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}
	defer testDB.Close()

	// Create a test app with database
	app := &App{
		Database: &Database{db: testDB},
	}

	// Create a test server with app
	srv := &Server{
		mux: http.NewServeMux(),
		app: app,
	}
	srv.mux.HandleFunc("/check/readiness", srv.handleReadinessCheck)

	// Test readiness endpoint with app
	req, err := http.NewRequest(http.MethodGet, "/check/readiness", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	srv.mux.ServeHTTP(rr, req)

	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Readiness check returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check response body
	expectedResponse := map[string]string{"status": "UP"}
	var actualResponse map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &actualResponse); err != nil {
		t.Fatalf("Failed to parse response JSON: %v", err)
	}

	if actualResponse["status"] != expectedResponse["status"] {
		t.Errorf("Readiness check returned unexpected status: got %v want %v",
			actualResponse["status"], expectedResponse["status"])
	}
}
