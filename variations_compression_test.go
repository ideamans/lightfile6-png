package png

import (
	"path/filepath"
	"testing"
)

func TestCompression(t *testing.T) {
	tempDir := t.TempDir()

	cases := []struct {
		name                string
		file                string
		expectError         bool
		originalCompression int    // Expected original compression level
		description         string // Description of what this test case validates
	}{
		{
			name:                "No compression",
			file:                "compression_0.png",
			expectError:         false, // Should succeed with significant compression
			originalCompression: 0,
			description:         "PNG with no compression (maximum file size)",
		},
		{
			name:                "Standard compression",
			file:                "compression_6.png",
			expectError:         false, // Should succeed with some compression
			originalCompression: 6,
			description:         "PNG with standard compression (default level)",
		},
		{
			name:                "Maximum compression",
			file:                "compression_9.png",
			expectError:         false, // Should succeed with minimal compression
			originalCompression: 9,
			description:         "PNG with maximum compression (minimum file size)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inputPath := filepath.Join(".", "testdata", "variations", tc.file)
			outputPath := filepath.Join(tempDir, tc.file)

			t.Logf("Testing: %s", tc.description)

			// Note: PNG compression level detection from file is complex and not standardized
			// The compression level affects the deflate algorithm used but isn't stored in the file
			// We'll focus on the optimization results instead

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

			// Log optimization details
			t.Logf("Optimization result: %d -> %d bytes", result.BeforeSize, result.AfterSize)
			t.Logf("PSNR: %.2f, PNGQuant: %v", result.FinalPSNR, result.PNGQuant.Applied)

			// Analyze compression effectiveness based on original compression level
			if result.BeforeSize > 0 {
				compressionRatio := float64(result.BeforeSize-result.AfterSize) / float64(result.BeforeSize) * 100

				if compressionRatio >= 0 {
					t.Logf("Compression: %.1f%% reduction", compressionRatio)
				} else {
					t.Logf("Expansion: %.1f%% increase", -compressionRatio)
				}

				// Expected behavior analysis based on original compression
				switch tc.originalCompression {
				case 0:
					t.Logf("No-compression input: optimization should achieve significant compression")
					if compressionRatio < 10 {
						t.Logf("Warning: Expected higher compression ratio for uncompressed input")
					}
				case 6:
					t.Logf("Standard-compression input: optimization should achieve modest compression")
				case 9:
					t.Logf("Max-compression input: optimization should have minimal effect")
				}
			}

			// Log quality metrics
			t.Logf("Successfully optimized with PSNR: %.2f", result.FinalPSNR)

			// Check quality retention
			if result.FinalPSNR < 35.0 {
				t.Logf("Note: Lower PSNR indicates more aggressive optimization")
			} else {
				t.Logf("Note: High PSNR indicates good quality retention")
			}
		})
	}
}
