package png

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInterlace(t *testing.T) {
	tempDir := t.TempDir()

	cases := []struct {
		name              string
		file              string
		expectError       bool
		originalInterlace int    // Expected original interlace method
		description       string // Description of what this test case validates
	}{
		{
			name:              "No interlace",
			file:              "interlace_none.png",
			expectError: false, // Should succeed
			originalInterlace: 0,
			description:       "PNG without interlacing (standard)",
		},
		{
			name:              "Adam7 interlace",
			file:              "interlace_adam7.png",
			expectError: false, // Should succeed
			originalInterlace: 1,
			description:       "PNG with Adam7 interlacing (progressive display)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inputPath := filepath.Join(".", "testdata", "variations", tc.file)
			outputPath := filepath.Join(tempDir, tc.file)

			t.Logf("Testing: %s", tc.description)

			// Check interlace method in original file
			originalInterlaceStr := checkInterlace(t, inputPath)
			originalInterlace := 0
			if originalInterlaceStr == "Adam7" {
				originalInterlace = 1
			}
			t.Logf("Original file interlace method: %d", originalInterlace)

			// Verify our expectation matches reality for the original file
			if originalInterlace != tc.originalInterlace {
				t.Errorf("Expected original interlace method %d, but got %d", tc.originalInterlace, originalInterlace)
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

			// Log optimization details
			t.Logf("Optimization result: %d -> %d bytes", result.BeforeSize, result.AfterSize)
			t.Logf("PSNR: %.2f, PNGQuant: %v", result.FinalPSNR, result.PNGQuant.Applied)

			// Check interlace handling
				// Check that output file exists
				if _, err := os.Stat(outputPath); os.IsNotExist(err) {
					t.Error("Output file was not created")
					return
				}

				// Check interlace method in optimized file
				optimizedInterlaceStr := checkInterlace(t, outputPath)
				optimizedInterlace := 0
				if optimizedInterlaceStr == "Adam7" {
					optimizedInterlace = 1
				}
				t.Logf("Optimized file interlace method: %d", optimizedInterlace)

				// Log interlace method conversion
				if originalInterlace != optimizedInterlace {
					t.Logf("Interlace method changed: %d -> %d", originalInterlace, optimizedInterlace)

					if originalInterlace == 1 && optimizedInterlace == 0 {
						t.Logf("Adam7 interlacing removed (optimization typically removes interlacing)")
					} else if originalInterlace == 0 && optimizedInterlace == 1 {
						t.Logf("Interlacing added (unexpected)")
					}
				} else {
					t.Logf("Interlace method preserved: %d", originalInterlace)
				}

				// Log compression details
				if result.BeforeSize > 0 {
					compressionRatio := float64(result.BeforeSize-result.AfterSize) / float64(result.BeforeSize) * 100
					t.Logf("Compression: %.1f%% reduction", compressionRatio)
				}

				// Interlace-specific analysis
				switch originalInterlace {
				case 0:
					t.Logf("Non-interlaced optimization: standard progressive scan processing")
				case 1:
					t.Logf("Adam7 interlaced optimization: 7-pass progressive image processing")
					if optimizedInterlace == 0 {
						t.Logf("Interlacing removed for better compression (common optimization)")
					}
				}

			// Log interlace method implications
			switch tc.originalInterlace {
			case 0:
				t.Logf("Interlace analysis: Standard scanline order - faster decode, smaller files")
			case 1:
				t.Logf("Interlace analysis: Adam7 - progressive display, larger files")
			}
		})
	}
}
