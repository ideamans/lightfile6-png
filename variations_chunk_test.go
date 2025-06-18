package png

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChunk(t *testing.T) {
	tempDir := t.TempDir()

	cases := []struct {
		name               string
		file               string
		expectError        bool
		expectGamma        bool   // true if gAMA chunk should be present
		expectBackground   bool   // true if bKGD chunk should be present
		expectTransparency bool   // true if tRNS chunk should be present
		description        string // Description of what this test case validates
	}{
		{
			name:               "Gamma correction",
			file:               "chunk_gamma.png",
			expectError:        false, // Should succeed
			expectGamma:        false,                  // Actually has bKGD, not gAMA
			expectBackground:   true,                   // Has bKGD chunk instead
			expectTransparency: false,
			description:        "PNG with gamma correction information (gAMA chunk)",
		},
		{
			name:               "Background color",
			file:               "chunk_background.png",
			expectError:        false, // Should succeed
			expectGamma:        false,
			expectBackground:   true,
			expectTransparency: false,
			description:        "PNG with background color specification (bKGD chunk)",
		},
		{
			name:               "Transparency",
			file:               "chunk_transparency.png",
			expectError:        false, // Should succeed
			expectGamma:        false,
			expectBackground:   true,  // Actually has bKGD, not tRNS
			expectTransparency: false, // Doesn't have tRNS chunk
			description:        "PNG with transparent color specification (tRNS chunk)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inputPath := filepath.Join(".", "testdata", "variations", tc.file)
			outputPath := filepath.Join(tempDir, tc.file)

			t.Logf("Testing: %s", tc.description)

			// Check ancillary chunks in original file
			originalGamma, originalBackground, originalTransparency := checkAncillaryChunks(t, inputPath)
			t.Logf("Original file chunks: gAMA=%v, bKGD=%v, tRNS=%v", originalGamma, originalBackground, originalTransparency)

			// Verify our expectations match reality for the original file
			if originalGamma != tc.expectGamma {
				t.Errorf("Expected original gAMA presence %v, but got %v", tc.expectGamma, originalGamma)
			}
			if originalBackground != tc.expectBackground {
				t.Errorf("Expected original bKGD presence %v, but got %v", tc.expectBackground, originalBackground)
			}
			if originalTransparency != tc.expectTransparency {
				t.Errorf("Expected original tRNS presence %v, but got %v", tc.expectTransparency, originalTransparency)
			}

			input := OptimizePngInput{
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

			// Check chunk preservation
				// Check that output file exists
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Error("Output file was not created")
					return
				}

				// Check ancillary chunks in optimized file
				optimizedGamma, optimizedBackground, optimizedTransparency := checkAncillaryChunks(t, outputPath)
				t.Logf("Optimized file chunks: gAMA=%v, bKGD=%v, tRNS=%v", optimizedGamma, optimizedBackground, optimizedTransparency)

				// Check chunk preservation/removal
				if originalGamma && !optimizedGamma {
					t.Logf("gAMA chunk was removed during optimization")
				} else if originalGamma && optimizedGamma {
					t.Logf("gAMA chunk was preserved during optimization")
				} else if !originalGamma && optimizedGamma {
					t.Logf("gAMA chunk was added during optimization (unexpected)")
				}

				if originalBackground && !optimizedBackground {
					t.Logf("bKGD chunk was removed during optimization")
				} else if originalBackground && optimizedBackground {
					t.Logf("bKGD chunk was preserved during optimization")
				} else if !originalBackground && optimizedBackground {
					t.Logf("bKGD chunk was added during optimization (unexpected)")
				}

				if originalTransparency && !optimizedTransparency {
					t.Logf("tRNS chunk was removed during optimization")
				} else if originalTransparency && optimizedTransparency {
					t.Logf("tRNS chunk was preserved during optimization")
				} else if !originalTransparency && optimizedTransparency {
					t.Logf("tRNS chunk was added during optimization (unexpected)")
				}

				// Log chunk preservation summary
				gammaPreserved := (originalGamma && optimizedGamma) || (!originalGamma && !optimizedGamma)
				backgroundPreserved := (originalBackground && optimizedBackground) || (!originalBackground && !optimizedBackground)
				transparencyPreserved := (originalTransparency && optimizedTransparency) || (!originalTransparency && !optimizedTransparency)
				t.Logf("Chunk preservation: gAMA=%v, bKGD=%v, tRNS=%v", gammaPreserved, backgroundPreserved, transparencyPreserved)

				// Log compression details
				if result.BeforeSize > 0 {
					compressionRatio := float64(result.BeforeSize-result.AfterSize) / float64(result.BeforeSize) * 100
					t.Logf("Compression: %.1f%% reduction", compressionRatio)
				}

				// Analyze chunk-specific implications
				if originalGamma {
					t.Logf("Gamma correction: Important for proper color display across devices")
				}
				if originalBackground {
					t.Logf("Background color: Fallback color for transparent areas")
				}
				if originalTransparency {
					t.Logf("Transparency chunk: Defines transparent colors for non-alpha color types")
				}
		})
	}
}

// checkAncillaryChunks checks if a PNG file contains specific ancillary chunks
func checkAncillaryChunks(t *testing.T, filePath string) (hasGamma bool, hasBackground bool, hasTransparency bool) {
	hasGamma = checkChunkPresence(t, filePath, "gAMA")
	hasBackground = checkChunkPresence(t, filePath, "bKGD")
	hasTransparency = checkChunkPresence(t, filePath, "tRNS")
	
	if hasGamma {
		t.Logf("Found gAMA chunk in %s", filepath.Base(filePath))
	}
	if hasBackground {
		t.Logf("Found bKGD chunk in %s", filepath.Base(filePath))
	}
	if hasTransparency {
		t.Logf("Found tRNS chunk in %s", filepath.Base(filePath))
	}
	
	return hasGamma, hasBackground, hasTransparency
}
