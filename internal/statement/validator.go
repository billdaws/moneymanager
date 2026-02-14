package statement

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"slices"
)

// ValidateFile checks that the file data is within size limits and has an allowed MIME type.
// It returns the detected MIME type.
func ValidateFile(data []byte, maxSizeMB int, allowedTypes []string) (string, error) {
	maxBytes := int64(maxSizeMB) * 1024 * 1024
	if int64(len(data)) > maxBytes {
		return "", fmt.Errorf("file size %d bytes exceeds maximum %d MB", len(data), maxSizeMB)
	}

	if len(data) == 0 {
		return "", fmt.Errorf("file is empty")
	}

	mimeType := http.DetectContentType(data)

	// http.DetectContentType returns "application/octet-stream" for PDFs,
	// so also check for the PDF magic bytes.
	if len(data) >= 5 && string(data[:5]) == "%PDF-" {
		mimeType = "application/pdf"
	}

	if slices.Contains(allowedTypes, mimeType) {
		return mimeType, nil
	}

	// Also accept text/plain as CSV (DetectContentType returns text/plain for CSV files).
	if mimeType == "text/plain; charset=utf-8" || mimeType == "text/plain" {
		if slices.Contains(allowedTypes, "text/csv") {
			return "text/csv", nil
		}
	}

	return "", fmt.Errorf("file type %q is not allowed", mimeType)
}

// HashFile returns the hex-encoded SHA256 hash of the data.
func HashFile(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
