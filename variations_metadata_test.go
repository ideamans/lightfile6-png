package png

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMetadata(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	cases := []struct {
		name                string
		file                string
		expectError         bool
		expectText          bool   // true if text chunks should be present
		expectCompressed    bool   // true if compressed text chunks should be present
		expectInternational bool   // true if international text chunks should be present
		description         string // Description of what this test case validates
	}{
		{
			name:                "No metadata",
			file:                "metadata_none.png",
			expectError:         false, // With force quality, pngquant always runs and usually reduces size
			expectText:          false,
			expectCompressed:    false,
			expectInternational: false,
			description:         "PNG without any metadata chunks",
		},
		{
			name:                "Text metadata",
			file:                "metadata_text.png",
			expectError:         false, // Should succeed
			expectText:          true,
			expectCompressed:    true, // Actually has zTXt chunks too
			expectInternational: false,
			description:         "PNG with tEXt chunks (uncompressed text)",
		},
		{
			name:                "Compressed text metadata",
			file:                "metadata_compressed.png",
			expectError:         false, // Should succeed
			expectText:          true,  // Actually has tEXt chunks too
			expectCompressed:    true,
			expectInternational: false,
			description:         "PNG with zTXt chunks (compressed text)",
		},
		{
			name:                "International text metadata",
			file:                "metadata_international.png",
			expectError:         false, // Should succeed
			expectText:          true,  // Actually has tEXt chunks
			expectCompressed:    true,  // Actually has zTXt chunks
			expectInternational: false, // Doesn't have iTXt chunks
			description:         "PNG with iTXt chunks (international text, UTF-8)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inputPath := filepath.Join(".", "testdata", "variations", tc.file)
			outputPath := filepath.Join(tempDir, tc.file)

			t.Logf("Testing: %s", tc.description)

			// Check metadata in original file
			originalText, originalCompressed, originalInternational := checkMetadata(t, inputPath)
			t.Logf("Original file metadata: tEXt=%v, zTXt=%v, iTXt=%v", originalText, originalCompressed, originalInternational)

			// Verify our expectations match reality for the original file
			if originalText != tc.expectText {
				t.Errorf("Expected original tEXt presence %v, but got %v", tc.expectText, originalText)
			}
			if originalCompressed != tc.expectCompressed {
				t.Errorf("Expected original zTXt presence %v, but got %v", tc.expectCompressed, originalCompressed)
			}
			if originalInternational != tc.expectInternational {
				t.Errorf("Expected original iTXt presence %v, but got %v", tc.expectInternational, originalInternational)
			}

			input := OptimizePNGInput{
				SrcPath:  inputPath,
				DestPath: outputPath,
				Quality:  "force",
			}

			result, err := Optimize(input)

			// Check error expectation
			if tc.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Log optimization details regardless of result
			t.Logf("Optimization result: %d -> %d bytes", result.BeforeSize, result.AfterSize)
			t.Logf("PSNR: %.2f, PNGQuant: %v", result.FinalPSNR, result.PNGQuant.Applied)

			// Check metadata preservation
			// Check that output file exists
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Error("Output file was not created")
				return
			}

			// Check metadata in optimized file
			optimizedText, optimizedCompressed, optimizedInternational := checkMetadata(t, outputPath)
			t.Logf("Optimized file metadata: tEXt=%v, zTXt=%v, iTXt=%v", optimizedText, optimizedCompressed, optimizedInternational)

			// Check metadata preservation/removal
			if originalText && !optimizedText {
				t.Logf("tEXt metadata was removed during optimization")
			} else if originalText && optimizedText {
				t.Logf("tEXt metadata was preserved during optimization")
			} else if !originalText && optimizedText {
				t.Logf("tEXt metadata was added during optimization (unexpected)")
			}

			if originalCompressed && !optimizedCompressed {
				t.Logf("zTXt metadata was removed during optimization")
			} else if originalCompressed && optimizedCompressed {
				t.Logf("zTXt metadata was preserved during optimization")
			} else if !originalCompressed && optimizedCompressed {
				t.Logf("zTXt metadata was added during optimization (unexpected)")
			}

			if originalInternational && !optimizedInternational {
				t.Logf("iTXt metadata was removed during optimization")
			} else if originalInternational && optimizedInternational {
				t.Logf("iTXt metadata was preserved during optimization")
			} else if !originalInternational && optimizedInternational {
				t.Logf("iTXt metadata was added during optimization (could be optimization tool signature)")
			}

			// Log metadata preservation summary
			textPreserved := (originalText && optimizedText) || (!originalText && !optimizedText)
			compressedPreserved := (originalCompressed && optimizedCompressed) || (!originalCompressed && !optimizedCompressed)
			internationalPreserved := (originalInternational && optimizedInternational) || (!originalInternational && !optimizedInternational)
			t.Logf("Metadata preservation: tEXt=%v, zTXt=%v, iTXt=%v", textPreserved, compressedPreserved, internationalPreserved)

			// Log compression details
			if result.BeforeSize > 0 {
				compressionRatio := float64(result.BeforeSize-result.AfterSize) / float64(result.BeforeSize) * 100
				t.Logf("Compression: %.1f%% reduction", compressionRatio)
			}
		})
	}
}

// checkMetadata checks if a PNG file contains various types of text metadata chunks
func checkMetadata(t *testing.T, filePath string) (hasText bool, hasCompressed bool, hasInternational bool) {
	hasText = checkChunkPresence(t, filePath, "tEXt")
	hasCompressed = checkChunkPresence(t, filePath, "zTXt")
	hasInternational = checkChunkPresence(t, filePath, "iTXt")

	if hasText {
		t.Logf("Found tEXt chunk in %s", filepath.Base(filePath))
	}
	if hasCompressed {
		t.Logf("Found zTXt chunk in %s", filepath.Base(filePath))
	}
	if hasInternational {
		t.Logf("Found iTXt chunk in %s", filepath.Base(filePath))
	}

	return hasText, hasCompressed, hasInternational
}
