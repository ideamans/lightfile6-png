package png

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDebugCI(t *testing.T) {
	// Check working directory
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	t.Logf("Working directory: %s", wd)

	// Check if libimagequant.a exists
	libPath := filepath.Join("libimagequant", "imagequant-sys", "libimagequant.a")
	if _, err := os.Stat(libPath); err != nil {
		t.Logf("libimagequant.a not found at %s: %v", libPath, err)
		// Try absolute path
		absPath := filepath.Join(wd, libPath)
		if _, err := os.Stat(absPath); err != nil {
			t.Logf("libimagequant.a not found at absolute path %s: %v", absPath, err)
		}
	} else {
		t.Logf("libimagequant.a found at %s", libPath)
	}

	// Check testdata directory
	testdataPath := "testdata"
	entries, err := os.ReadDir(testdataPath)
	if err != nil {
		t.Fatalf("Failed to read testdata directory: %v", err)
	}
	t.Logf("testdata directory contains %d entries", len(entries))
	for i, entry := range entries {
		if i < 5 { // Show first 5 entries
			t.Logf("  - %s (dir: %v)", entry.Name(), entry.IsDir())
		}
	}

	// Try to load a test file
	testFile := filepath.Join("testdata", "binding", "psnr-will-50.png")
	if data, err := os.ReadFile(testFile); err != nil {
		t.Errorf("Failed to read test file %s: %v", testFile, err)
	} else {
		t.Logf("Successfully read test file %s (%d bytes)", testFile, len(data))
	}
}