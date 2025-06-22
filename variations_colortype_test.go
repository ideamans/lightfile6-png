package png

import (
	"os"
	"path/filepath"
	"testing"
)

func TestColortype(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	cases := []struct {
		name              string
		file              string
		expectError       bool
		originalColorType string // Expected original color type
		description       string // Description of what this test case validates
	}{
		{
			name:              "Grayscale",
			file:              "colortype_grayscale.png",
			expectError:       false, // Should work with current implementation
			originalColorType: "Grayscale",
			description:       "Grayscale PNG without alpha channel",
		},
		{
			name:              "Palette",
			file:              "colortype_palette.png",
			expectError:       false, // Should succeed - already palette
			originalColorType: "Palette",
			description:       "Palette-based PNG (indexed color)",
		},
		{
			name:              "RGB",
			file:              "colortype_rgb.png",
			expectError:       false,  // Should succeed with RGB optimization
			originalColorType: "RGBA", // Go decodes as RGBA even without alpha
			description:       "RGB PNG without alpha channel",
		},
		{
			name:              "RGBA",
			file:              "colortype_rgba.png",
			expectError:       false, // Should succeed with RGBA optimization
			originalColorType: "RGBA",
			description:       "RGBA PNG with alpha channel",
		},
		{
			name:              "Grayscale + Alpha",
			file:              "colortype_grayscale_alpha.png",
			expectError:       false,             // Should succeed with grayscale+alpha optimization
			originalColorType: "Grayscale+Alpha", // May be decoded as RGBA or Grayscale+Alpha
			description:       "Grayscale PNG with alpha channel",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inputPath := filepath.Join(".", "testdata", "variations", tc.file)
			outputPath := filepath.Join(tempDir, tc.file)

			t.Logf("Testing: %s", tc.description)

			// Check color type in original file
			originalColorType := checkColorType(t, inputPath)
			t.Logf("Original file color type: %s", originalColorType)

			// Verify our expectation matches reality for the original file
			if originalColorType != tc.originalColorType {
				t.Errorf("Expected original color type %s, but got %s", tc.originalColorType, originalColorType)
			}

			result, err := runVariationOptimization(t, inputPath, outputPath, "force")

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

			// Check color type preservation
			// Check that output file exists
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Error("Output file was not created")
				return
			}

			// Check color type in optimized file
			optimizedColorType := checkColorType(t, outputPath)
			t.Logf("Optimized file color type: %s", optimizedColorType)

			// Log color type conversion
			if originalColorType != optimizedColorType {
				t.Logf("Color type conversion: %s -> %s", originalColorType, optimizedColorType)
			} else {
				t.Logf("Color type preserved: %s", originalColorType)
			}

			// Log compression details
			if result.BeforeSize > 0 && result.AfterSize > 0 {
				compressionRatio := float64(result.BeforeSize-result.AfterSize) / float64(result.BeforeSize) * 100
				t.Logf("Compression: %.1f%% reduction", compressionRatio)
			}
		})
	}
}
