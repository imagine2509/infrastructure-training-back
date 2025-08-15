package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResponseWriter(t *testing.T) {
	// Create a mock response writer
	recorder := httptest.NewRecorder()
	rw := &responseWriter{
		ResponseWriter: recorder,
		statusCode:     0,
		size:          0,
	}

	// Test WriteHeader
	rw.WriteHeader(http.StatusOK)
	if rw.statusCode != http.StatusOK {
		t.Errorf("WriteHeader() statusCode = %v, want %v", rw.statusCode, http.StatusOK)
	}

	// Test Write
	testData := []byte("test response")
	n, err := rw.Write(testData)
	if err != nil {
		t.Errorf("Write() error = %v", err)
	}
	if n != len(testData) {
		t.Errorf("Write() returned %d bytes, want %d", n, len(testData))
	}
	if rw.size != len(testData) {
		t.Errorf("Write() size = %d, want %d", rw.size, len(testData))
	}

	// Test that data was actually written
	if recorder.Body.String() != string(testData) {
		t.Errorf("Write() body = %v, want %v", recorder.Body.String(), string(testData))
	}
}