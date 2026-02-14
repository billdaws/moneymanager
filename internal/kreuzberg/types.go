package kreuzberg

// ExtractionResult represents a single document extraction from the Kreuzberg API.
type ExtractionResult struct {
	Content           string           `json:"content"`
	MimeType          string           `json:"mime_type"`
	Metadata          map[string]any   `json:"metadata"`
	Tables            []Table          `json:"tables"`
	DetectedLanguages []string         `json:"detected_languages"`
	Chunks            []Chunk          `json:"chunks"`
	Images            []Image          `json:"images"`
}

// Table represents an extracted table from a document.
type Table struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
}

// Chunk represents a text chunk from document extraction.
type Chunk struct {
	Content  string         `json:"content"`
	Metadata map[string]any `json:"metadata"`
}

// Image represents an extracted image from a document.
type Image struct {
	ID       string `json:"id"`
	Content  string `json:"content"`
	MimeType string `json:"mime_type"`
}

// HealthResponse represents the Kreuzberg health endpoint response.
type HealthResponse struct {
	Status string `json:"status"`
}
