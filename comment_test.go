package png

import (
	"os"
	"testing"
)

func TestReadComment_ValidPNG(t *testing.T) {
	// Load a test PNG file
	data, err := os.ReadFile("testdata/variations/metadata_text.png")
	if err != nil {
		t.Skipf("Test PNG file not found: %v", err)
	}

	comment, rawComment, err := ReadComment(data)
	if err != nil {
		t.Fatalf("ReadComment failed: %v", err)
	}

	// Should not find LightFile comment in regular PNG
	if comment != nil {
		t.Errorf("Expected no LightFile comment, got: %+v", comment)
	}
	if rawComment != "" {
		t.Errorf("Expected empty raw comment, got: %s", rawComment)
	}
}

func TestReadComment_InvalidJSON(t *testing.T) {
	// Create a PNG with LightFile tEXt chunk but invalid JSON
	originalData, err := os.ReadFile("testdata/variations/colortype_rgb.png")
	if err != nil {
		t.Skipf("Test PNG file not found: %v", err)
	}

	// Write invalid JSON as LightFile comment
	invalidJSON := "not valid json"
	modifiedData, err := WriteComment(originalData, invalidJSON)
	if err != nil {
		t.Fatalf("WriteComment failed: %v", err)
	}

	// Read comment back
	comment, rawComment, err := ReadComment(modifiedData)
	if err != nil {
		t.Fatalf("ReadComment should not fail for invalid JSON: %v", err)
	}

	// Should return nil comment but raw text
	if comment != nil {
		t.Errorf("Expected nil comment for invalid JSON, got: %+v", comment)
	}

	if rawComment != invalidJSON {
		t.Errorf("Expected raw comment %q, got %q", invalidJSON, rawComment)
	}
}

func TestBuildComment(t *testing.T) {
	comment := &LightFileComment{
		By:       "lightfile6-png",
		Before:   1000,
		After:    800,
		PNGQuant: true,
	}

	jsonStr, size, err := BuildComment(comment)
	if err != nil {
		t.Fatalf("BuildComment failed: %v", err)
	}

	if jsonStr == "" {
		t.Error("Expected non-empty JSON string")
	}

	if size != len(jsonStr) {
		t.Errorf("Size mismatch: expected %d, got %d", len(jsonStr), size)
	}

	expectedFields := []string{
		`"by":"lightfile6-png"`,
		`"before":1000`,
		`"after":800`,
		`"pngquant":true`,
	}

	for _, field := range expectedFields {
		if !contains(jsonStr, field) {
			t.Errorf("JSON missing expected field: %s\nFull JSON: %s", field, jsonStr)
		}
	}
}

func TestWriteAndReadComment_RoundTrip(t *testing.T) {
	// Load a test PNG file
	originalData, err := os.ReadFile("testdata/variations/colortype_rgb.png")
	if err != nil {
		t.Skipf("Test PNG file not found: %v", err)
	}

	// Create a test comment
	testComment := &LightFileComment{
		By:       "lightfile6-png-test",
		Before:   2048,
		After:    1536,
		PNGQuant: false,
	}

	// Build comment JSON
	commentJSON, _, err := BuildComment(testComment)
	if err != nil {
		t.Fatalf("BuildComment failed: %v", err)
	}

	// Write comment to PNG
	modifiedData, err := WriteComment(originalData, commentJSON)
	if err != nil {
		t.Fatalf("WriteComment failed: %v", err)
	}

	// Verify the modified data is different
	if len(modifiedData) <= len(originalData) {
		t.Error("Modified PNG should be larger than original")
	}

	// Read comment back
	readComment, rawComment, err := ReadComment(modifiedData)
	if err != nil {
		t.Fatalf("ReadComment failed: %v", err)
	}

	// Verify comment was read correctly
	if readComment == nil {
		t.Fatal("Expected to read LightFile comment, got nil")
	}

	if readComment.By != testComment.By {
		t.Errorf("By field mismatch: expected %s, got %s", testComment.By, readComment.By)
	}

	if readComment.Before != testComment.Before {
		t.Errorf("Before field mismatch: expected %d, got %d", testComment.Before, readComment.Before)
	}

	if readComment.After != testComment.After {
		t.Errorf("After field mismatch: expected %d, got %d", testComment.After, readComment.After)
	}

	if readComment.PNGQuant != testComment.PNGQuant {
		t.Errorf("PNGQuant field mismatch: expected %t, got %t", testComment.PNGQuant, readComment.PNGQuant)
	}

	if rawComment != commentJSON {
		t.Errorf("Raw comment mismatch:\nExpected: %s\nGot: %s", commentJSON, rawComment)
	}
}

func TestReadComment_InvalidPNG(t *testing.T) {
	invalidData := []byte("not a png file")

	comment, rawComment, err := ReadComment(invalidData)
	if err == nil {
		t.Error("Expected error for invalid PNG data")
	}

	if comment != nil {
		t.Error("Expected nil comment for invalid PNG")
	}

	if rawComment != "" {
		t.Error("Expected empty raw comment for invalid PNG")
	}

	// Should be a DataError
	if AsDataError(err) == nil {
		t.Error("Expected DataError for invalid PNG structure")
	}
}

func TestWriteComment_InvalidPNG(t *testing.T) {
	invalidData := []byte("not a png file")
	comment := `{"by":"test","before":100,"after":80,"pngquant":false}`

	result, err := WriteComment(invalidData, comment)
	if err == nil {
		t.Error("Expected error for invalid PNG data")
	}

	if result != nil {
		t.Error("Expected nil result for invalid PNG")
	}

	// Should be a DataError
	if AsDataError(err) == nil {
		t.Error("Expected DataError for invalid PNG structure")
	}
}

func TestReadComment_EmptyPNG(t *testing.T) {
	emptyData := []byte{}

	comment, rawComment, err := ReadComment(emptyData)
	if err == nil {
		t.Error("Expected error for empty PNG data")
	}

	if comment != nil {
		t.Error("Expected nil comment for empty PNG")
	}

	if rawComment != "" {
		t.Error("Expected empty raw comment for empty PNG")
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsAt(s, substr, 1)))
}

func containsAt(s, substr string, start int) bool {
	if start >= len(s) {
		return false
	}
	if start+len(substr) <= len(s) && s[start:start+len(substr)] == substr {
		return true
	}
	return containsAt(s, substr, start+1)
}
