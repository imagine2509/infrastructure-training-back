package main

import (
	"encoding/hex"
	"testing"
)

func TestGenerateRequestID(t *testing.T) {
	id1 := generateRequestID()
	id2 := generateRequestID()

	// Check that ID is not empty
	if id1 == "" {
		t.Error("generateRequestID() returned empty string")
	}

	// Check that two calls return different IDs
	if id1 == id2 {
		t.Error("generateRequestID() returned the same ID twice")
	}

	// Check that ID is a valid hex string of expected length (16 chars for 8 bytes)
	if len(id1) != 16 {
		t.Errorf("generateRequestID() returned ID of wrong length: got %d want %d",
			len(id1), 16)
	}

	// Check that ID is valid hex
	_, err := hex.DecodeString(id1)
	if err != nil {
		t.Errorf("generateRequestID() returned invalid hex string: %v", err)
	}
}