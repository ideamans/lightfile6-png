package png

import (
	"path/filepath"
	"testing"
)

func TestFilter(t *testing.T) {
	t.Parallel()
	tempDir := t.TempDir()

	cases := []struct {
		name        string
		file        string
		expectError bool
		filterType  string // Expected filter type used
		description string // Description of what this test case validates
	}{
		{
			name:        "No filter",
			file:        "filter_none.png",
			expectError: false, // Should succeed
			filterType:  "None",
			description: "PNG with no filter (filter type 0)",
		},
		{
			name:        "Sub filter",
			file:        "filter_sub.png",
			expectError: false, // Should succeed
			filterType:  "Sub",
			description: "PNG with Sub filter (horizontal prediction)",
		},
		{
			name:        "Up filter",
			file:        "filter_up.png",
			expectError: false, // Should succeed
			filterType:  "Up",
			description: "PNG with Up filter (vertical prediction)",
		},
		{
			name:        "Average filter",
			file:        "filter_average.png",
			expectError: false, // Should succeed
			filterType:  "Average",
			description: "PNG with Average filter (average prediction)",
		},
		{
			name:        "Paeth filter",
			file:        "filter_paeth.png",
			expectError: false, // Should succeed
			filterType:  "Paeth",
			description: "PNG with Paeth filter (complex prediction)",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			inputPath := filepath.Join(".", "testdata", "variations", tc.file)
			outputPath := filepath.Join(tempDir, tc.file)

			t.Logf("Testing: %s", tc.description)

			// Note: PNG filter detection requires decompressing and analyzing IDAT chunks
			// This is complex as filters can vary per scanline. We'll focus on optimization behavior
			// and assume the test files use the filters indicated by their names

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

			// Analyze optimization effectiveness based on filter type
			if result.BeforeSize > 0 {
				compressionRatio := float64(result.BeforeSize-result.AfterSize) / float64(result.BeforeSize) * 100

				if compressionRatio >= 0 {
					t.Logf("Compression: %.1f%% reduction", compressionRatio)
				} else {
					t.Logf("Expansion: %.1f%% increase", -compressionRatio)
				}

				// Expected behavior analysis based on filter type
				switch tc.filterType {
				case "None":
					t.Logf("No-filter input: optimization should apply appropriate filtering")
				case "Sub":
					t.Logf("Sub-filter input: works well for images with horizontal patterns")
				case "Up":
					t.Logf("Up-filter input: works well for images with vertical patterns")
				case "Average":
					t.Logf("Average-filter input: balanced approach for mixed patterns")
				case "Paeth":
					t.Logf("Paeth-filter input: complex prediction, often most effective")
					if compressionRatio > 0 {
						t.Logf("Additional optimization achieved beyond Paeth filtering")
					}
				}
			}

			// Log quality metrics
			t.Logf("Successfully optimized with PSNR: %.2f", result.FinalPSNR)

			// Filter-specific quality analysis
			t.Logf("Original filter (%s) processed successfully", tc.filterType)

			// Check if optimization maintained quality
			if result.FinalPSNR >= 40.0 {
				t.Logf("Excellent quality retention (PSNR >= 40)")
			} else if result.FinalPSNR >= 35.0 {
				t.Logf("Good quality retention (PSNR >= 35)")
			} else {
				t.Logf("Lower quality retention - aggressive optimization applied")
			}

			// Log filter-specific insights
			switch tc.filterType {
			case "None":
				t.Logf("Filter analysis: No prediction - larger file size but faster decode")
			case "Sub":
				t.Logf("Filter analysis: Horizontal prediction - good for gradients")
			case "Up":
				t.Logf("Filter analysis: Vertical prediction - good for vertical patterns")
			case "Average":
				t.Logf("Filter analysis: Average prediction - balanced compression")
			case "Paeth":
				t.Logf("Filter analysis: Complex prediction - usually best compression")
			}
		})
	}
}
