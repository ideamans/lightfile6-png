package png

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDepth(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	cases := []struct {
		name             string
		file             string
		expectError      bool
		originalBitDepth int    // Expected original bit depth
		description      string // Description of what this test case validates
	}{
		{
			name:             "1-bit depth",
			file:             "depth_1bit.png",
			expectError:      false, // Should work now
			originalBitDepth: 1,
			description:      "1-bit PNG (black and white only)",
		},
		{
			name:             "8-bit depth",
			file:             "depth_8bit.png",
			expectError:      false, // Should succeed
			originalBitDepth: 8,
			description:      "8-bit PNG (standard depth)",
		},
		{
			name:             "16-bit depth",
			file:             "depth_16bit.png",
			expectError:      false, // CantOptimize is not an error
			originalBitDepth: 16,
			description:      "16-bit PNG (high precision)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inputPath := filepath.Join(".", "testdata", "variations", tc.file)
			outputPath := filepath.Join(tempDir, tc.file)

			t.Logf("Testing: %s", tc.description)

			// Check bit depth in original file
			originalBitDepth := checkBitDepth(t, inputPath)
			t.Logf("Original file bit depth: %d", originalBitDepth)

			// Verify our expectation matches reality for the original file
			if originalBitDepth != tc.originalBitDepth {
				t.Errorf("Expected original bit depth %d, but got %d", tc.originalBitDepth, originalBitDepth)
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
					if result != nil {
						t.Logf("Result: BeforeSize=%d, AfterSize=%d, CantOptimize=%v, InspectionFailed=%v",
							result.BeforeSize, result.AfterSize, result.CantOptimize, result.InspectionFailed)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Log optimization details regardless of result
			t.Logf("Optimization result: %d -> %d bytes", result.BeforeSize, result.AfterSize)
			t.Logf("PSNR: %.2f, PNGQuant: %v", result.FinalPSNR, result.PNGQuant.Applied)

			// Check if optimization was aborted
			if result.CantOptimize {
				t.Logf("Cannot optimize: file would be larger")
				return
			}
			if result.InspectionFailed {
				t.Logf("Inspection failed: PSNR too low")
				return
			}

			// Check bit depth handling
			// Check that output file exists
			if _, err := os.Stat(outputPath); os.IsNotExist(err) {
				t.Error("Output file was not created")
				return
			}

			// Check bit depth in optimized file
			optimizedBitDepth := checkBitDepth(t, outputPath)
			t.Logf("Optimized file bit depth: %d", optimizedBitDepth)

			// Log bit depth conversion
			if originalBitDepth != optimizedBitDepth {
				t.Logf("Bit depth conversion: %d -> %d", originalBitDepth, optimizedBitDepth)

				// Analyze the conversion type
				if originalBitDepth > optimizedBitDepth {
					t.Logf("Bit depth reduced (quantization applied)")
				} else {
					t.Logf("Bit depth increased (unexpected)")
				}
			} else {
				t.Logf("Bit depth preserved: %d", originalBitDepth)
			}

			// Log compression details
			if result.BeforeSize > 0 {
				compressionRatio := float64(result.BeforeSize-result.AfterSize) / float64(result.BeforeSize) * 100
				t.Logf("Compression: %.1f%% reduction", compressionRatio)
			}

			// Special logging for high bit depth files
			if originalBitDepth == 16 {
				t.Logf("16-bit processing: PSNR=%.2f (precision retention metric)", result.FinalPSNR)
				if optimizedBitDepth == 8 {
					t.Logf("16-bit to 8-bit conversion applied (libimagequant quantization)")
				}
			}
		})
	}
}
