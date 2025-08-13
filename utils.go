package main

import (
	"crypto/rand"
	"encoding/hex"
)

func generateRequestID() string {
	bytes := make([]byte, 8)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}
