package models

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
)

// generateImageHash genera un hash SHA-256 de un archivo de imagen
func GenerateImageHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// GenerateImageHashFromBytes genera hash directamente desde bytes en memoria
func GenerateImageHashFromBytes(imgBytes []byte) string {
	hasher := sha256.New()
	hasher.Write(imgBytes)
	return hex.EncodeToString(hasher.Sum(nil))
}
