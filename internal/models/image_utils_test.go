package models_test

import (
	"testing"

	"github.com/uragamarco/proyecto-balistica/internal/models"
)

func TestGenerateImageHash(t *testing.T) {
	// Test con imagen de prueba
	hash, err := models.GenerateImageHash("testdata/images/test_case.png")
	if err != nil {
		t.Fatalf("Error generating hash: %v", err)
	}

	if len(hash) != 64 { // SHA-256 siempre produce 64 caracteres hex
		t.Errorf("Invalid hash length: got %d, want 64", len(hash))
	}

	// Test de consistencia
	hash2, _ := models.GenerateImageHash("testdata/images/test_case.png")
	if hash != hash2 {
		t.Error("Hashes should be identical for same input")
	}
}
