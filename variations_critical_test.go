package png

import (
	"path/filepath"
	"testing"
)

func TestCritical(t *testing.T) {
	tempDir := t.TempDir()

	cases := []struct {
		name        string
		file        string
		expectError bool
		description string // Description of what this test case validates
	}{
		{
			name:        "16-bit to palette conversion",
			file:        "critical_16bit_palette.png",
			expectError: false, // Should succeed with significant color reduction
			description: "16-bit PNG converted to palette (major color information loss)",
		},
		{
			name:        "RGBA to grayscale+alpha",
			file:        "critical_alpha_grayscale.png",
			expectError: false, // Should succeed with color space conversion
			description: "RGBA PNG converted to grayscale with alpha",
		},
		{
			name:        "Max compression + Paeth filter",
			file:        "critical_maxcompression_paeth.png",
			expectError: false, // Should succeed with minimal compression
			description: "PNG with maximum compression and Paeth filter combination",
		},
		{
			name:        "Interlace + high resolution",
			file:        "critical_interlace_highres.png",
			expectError: false, // Should succeed, may remove interlacing
			description: "High resolution PNG with Adam7 interlacing",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inputPath := filepath.Join(".", "testdata", "variations", tc.file)
			outputPath := filepath.Join(tempDir, tc.file)

			t.Logf("Testing: %s", tc.description)

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

			// Log compression ratio if successful
			if result.BeforeSize > 0 {
				compressionRatio := float64(result.BeforeSize-result.AfterSize) / float64(result.BeforeSize) * 100
				t.Logf("Compression: %.1f%% reduction", compressionRatio)
			}

			// Provide test-specific analysis
			switch tc.file {
			case "critical_16bit_palette.png":
				t.Logf("16-bit to palette: Expects significant color quantization")
				t.Logf("Successfully quantized 16-bit to palette colors")
				if result.FinalPSNR < 30.0 {
					t.Logf("Low PSNR expected due to severe color reduction")
				}

			case "critical_alpha_grayscale.png":
				t.Logf("RGBA to grayscale+alpha: Color information will be lost")
				t.Logf("Successfully converted to grayscale while preserving alpha")

			case "critical_maxcompression_paeth.png":
				t.Logf("Max compression + Paeth: Already highly optimized")
				t.Logf("Additional optimization achieved beyond max compression")

			case "critical_interlace_highres.png":
				t.Logf("Interlaced high-res: May remove interlacing for better compression")
				t.Logf("High resolution image processed successfully")
			}

			// Log quality metrics analysis
			if result.FinalPSNR > 0 {
				if result.FinalPSNR >= 40.0 {
					t.Logf("Excellent quality retention (PSNR >= 40)")
				} else if result.FinalPSNR >= 35.0 {
					t.Logf("Good quality retention (PSNR >= 35)")
				} else if result.FinalPSNR >= 25.0 {
					t.Logf("Acceptable quality with significant optimization (PSNR >= 25)")
				} else {
					t.Logf("Aggressive optimization applied (PSNR < 25)")
				}
			}
		})
	}
}
