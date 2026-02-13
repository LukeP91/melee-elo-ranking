package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseV2Format(t *testing.T) {
	// Use the testdata file
	testFile := filepath.Join("..", "..", "testdata", "matches", "tournament-170676.json")

	// Check if file exists, if not skip
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("testdata file not found")
	}

	parser := New()
	matches, err := parser.ParseFile(testFile, 170676)
	if err != nil {
		t.Fatalf("failed to parse file: %v", err)
	}

	if len(matches) == 0 {
		t.Error("expected matches, got 0")
	}

	// Verify first match
	firstMatch := matches[0]
	if firstMatch.TournamentID != 170676 {
		t.Errorf("expected tournament ID 170676, got %d", firstMatch.TournamentID)
	}
	if firstMatch.RoundNumber == 0 {
		t.Error("expected round number to be set")
	}
	if len(firstMatch.Competitors) != 2 {
		t.Errorf("expected 2 competitors, got %d", len(firstMatch.Competitors))
	}
}

func TestParseInvalidFile(t *testing.T) {
	parser := New()

	// Create temp invalid JSON file
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(invalidFile, []byte("not valid json"), 0644)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}

	_, err = parser.ParseFile(invalidFile, 1)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseNonExistentFile(t *testing.T) {
	parser := New()

	_, err := parser.ParseFile("/nonexistent/file.json", 1)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
