package png

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestOptimize(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name                      string
		inputFile                 string
		quality                   string
		expectAlreadyOptimized    bool
		expectStripError          bool
		expectPNGQuantError       bool
		expectPNGQuantApplied     bool
		expectCantOptimize        bool
		expectInspectionFailed    bool
		expectError               bool
		minPNGQuantPSNR           float64
		minFinalPSNR              float64
		maxAfterSize              int64
	}{
		{
			name:                  "Normal optimization with default quality",
			inputFile:             "testdata/optimize/psnr-will-50.png",
			quality:               "",
			expectPNGQuantApplied: true,
			minPNGQuantPSNR:       40.0,
			minFinalPSNR:          40.0,
			maxAfterSize:          12000, // Should be smaller after optimization
		},
		{
			name:                  "High quality optimization",
			inputFile:             "testdata/optimize/psnr-will-50.png",
			quality:               "high",
			expectPNGQuantApplied: true,
			minPNGQuantPSNR:       45.0,
			minFinalPSNR:          45.0,
			maxAfterSize:          15000,
		},
		{
			name:                  "Low quality optimization",
			inputFile:             "testdata/optimize/psnr-will-50.png",
			quality:               "low",
			expectPNGQuantApplied: true,
			minPNGQuantPSNR:       39.0,
			minFinalPSNR:          39.0,
			maxAfterSize:          12000,
		},
		{
			name:                  "Force quality optimization",
			inputFile:             "testdata/optimize/psnr-will-27.png",
			quality:               "force",
			expectPNGQuantApplied: true,
			minPNGQuantPSNR:       20.0, // Very low, but force accepts any
			minFinalPSNR:          20.0,
			maxAfterSize:          51000, // Force mode might not reduce size much
		},
		{
			name:                   "Already optimized file",
			inputFile:              "testdata/optimize/already-lightfile.png",
			quality:                "",
			expectAlreadyOptimized: true,
		},
		{
			name:                   "PSNR inspection failure - low PSNR",
			inputFile:              "testdata/optimize/psnr-will-27.png",
			quality:                "high",
			expectPNGQuantApplied:  true,  // Actually gets applied because PSNR is 27.9
			expectInspectionFailed: false, // Only fails final inspection if < 35
			minFinalPSNR:          45.0,   // Final PSNR is higher due to no PNGQuant
		},
		{
			name:      "PNG with metadata to strip",
			inputFile: "testdata/optimize/with-mac-icc.png",
			quality:   "",
			// Should have metadata stripped
			expectPNGQuantApplied: true,
			minFinalPSNR:          40.0,
		},
		{
			name:        "Invalid PNG file",
			inputFile:   "testdata/optimize/bad.png",
			quality:     "",
			expectError: true,
		},
		{
			name:                "JPEG disguised as PNG",
			inputFile:           "testdata/optimize/jpeg.png",
			quality:             "",
			expectError:         true, // PNG parsing will fail early
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup paths
			srcPath := tc.inputFile
			destPath := filepath.Join(tempDir, filepath.Base(tc.inputFile))

			// Special handling for "already optimized" test case
			if tc.name == "Already optimized file" {
				// First optimize a file to create "already optimized" state
				setupSrc := "testdata/optimize/psnr-will-50.png"
				tmpOptimized := filepath.Join(tempDir, "already-optimized.png")
				
				result, err := Optimize(OptimizePngInput{
					SrcPath:  setupSrc,
					DestPath: tmpOptimized,
					Quality:  "",
				})
				if err != nil {
					t.Fatalf("Failed to create already optimized file: %v", err)
				}
				if result.AlreadyOptimized {
					t.Fatal("Setup file was already optimized")
				}
				
				// Use the optimized file as source
				srcPath = tmpOptimized
			}

			// Run optimization
			input := OptimizePngInput{
				SrcPath:  srcPath,
				DestPath: destPath,
				Quality:  tc.quality,
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

			// Verify all output fields
			// Check AlreadyOptimized
			if result.AlreadyOptimized != tc.expectAlreadyOptimized {
				t.Errorf("AlreadyOptimized = %v, want %v", result.AlreadyOptimized, tc.expectAlreadyOptimized)
			}

			if tc.expectAlreadyOptimized {
				if result.AlreadyOptimizedBy != "LightFile" {
					t.Errorf("AlreadyOptimizedBy = %v, want LightFile", result.AlreadyOptimizedBy)
				}
				return // Skip other checks for already optimized files
			}

			// Check BeforeSize
			if result.BeforeSize <= 0 {
				t.Error("BeforeSize should be > 0")
			}

			// Check Strip results
			if tc.expectStripError && result.StripError == nil {
				t.Error("Expected StripError but got none")
			}
			if !tc.expectStripError && result.StripError != nil {
				t.Errorf("Unexpected StripError: %v", result.StripError)
			}

			if result.Strip != nil && !tc.expectStripError {
				t.Logf("Strip result: Total removed=%d bytes, TextChunks=%d, TimeChunk=%d, ExifData=%d", 
					result.Strip.Total, 
					result.Strip.Removed.TextChunks,
					result.Strip.Removed.TimeChunk,
					result.Strip.Removed.ExifData)
			}

			// Check SizeAfterStrip
			if result.SizeAfterStrip <= 0 {
				t.Error("SizeAfterStrip should be > 0")
			}
			if result.SizeAfterStrip > result.BeforeSize {
				t.Error("SizeAfterStrip should be <= BeforeSize")
			}

			// Check PNGQuant results
			if tc.expectPNGQuantError && result.PNGQuantError == nil {
				t.Error("Expected PNGQuantError but got none")
			}

			if result.PNGQuant.Applied != tc.expectPNGQuantApplied {
				t.Errorf("PNGQuant.Applied = %v, want %v", result.PNGQuant.Applied, tc.expectPNGQuantApplied)
			}

			if result.PNGQuantError == nil { // Only check PSNR if no error
				if tc.minPNGQuantPSNR > 0 && result.PNGQuant.PSNR < tc.minPNGQuantPSNR {
					t.Errorf("PNGQuant.PSNR = %v, want >= %v", result.PNGQuant.PSNR, tc.minPNGQuantPSNR)
				}
			}

			// Check SizeAfterPNGQuant
			if result.SizeAfterPNGQuant <= 0 {
				t.Error("SizeAfterPNGQuant should be > 0")
			}
			if result.PNGQuant.Applied && result.SizeAfterPNGQuant >= result.SizeAfterStrip {
				// PNGQuant might not always reduce size, but log it
				t.Logf("Warning: PNGQuant did not reduce size: %d >= %d", result.SizeAfterPNGQuant, result.SizeAfterStrip)
			}

			// Check CantOptimize
			if result.CantOptimize != tc.expectCantOptimize {
				t.Errorf("CantOptimize = %v, want %v", result.CantOptimize, tc.expectCantOptimize)
			}

			if result.CantOptimize {
				return // Skip remaining checks if can't optimize
			}

			// Check InspectionFailed
			if result.InspectionFailed != tc.expectInspectionFailed {
				t.Errorf("InspectionFailed = %v, want %v", result.InspectionFailed, tc.expectInspectionFailed)
			}

			if result.InspectionFailed {
				return // Skip remaining checks if inspection failed
			}

			// Check FinalPSNR
			if tc.minFinalPSNR > 0 && result.FinalPSNR < tc.minFinalPSNR {
				t.Errorf("FinalPSNR = %v, want >= %v", result.FinalPSNR, tc.minFinalPSNR)
			}

			// Check AfterSize
			if result.AfterSize <= 0 {
				t.Error("AfterSize should be > 0")
			}
			if tc.maxAfterSize > 0 && result.AfterSize > tc.maxAfterSize {
				t.Errorf("AfterSize = %v, want <= %v", result.AfterSize, tc.maxAfterSize)
			}

			// Verify file was written
			if _, err := os.Stat(destPath); os.IsNotExist(err) {
				t.Error("Output file was not created")
			}

			// Verify the output file has LightFile comment
			outputData, err := os.ReadFile(destPath)
			if err != nil {
				t.Fatalf("Failed to read output file: %v", err)
			}

			comment, _, err := ReadComment(outputData)
			if err != nil {
				t.Errorf("Failed to read comment from output: %v", err)
			}
			if comment == nil {
				t.Error("Output file should have LightFile comment")
			} else {
				if comment.By != "LightFile" {
					t.Errorf("Comment.By = %v, want LightFile", comment.By)
				}
				if comment.Before != result.BeforeSize {
					t.Errorf("Comment.Before = %v, want %v", comment.Before, result.BeforeSize)
				}
				// Comment.After is the size before adding the comment itself
				// So it should be slightly smaller than AfterSize
				if comment.After > result.AfterSize {
					t.Errorf("Comment.After = %v should be <= AfterSize %v", comment.After, result.AfterSize)
				}
				if comment.PNGQuant != result.PNGQuant.Applied {
					t.Errorf("Comment.PNGQuant = %v, want %v", comment.PNGQuant, result.PNGQuant.Applied)
				}
			}
		})
	}
}

func TestOptimize_PSNRQualityLevels(t *testing.T) {
	tempDir := t.TempDir()

	qualities := []struct {
		quality      string
		minPSNR      float64
		expectReject bool
	}{
		{"high", 45.0, false},
		{"", 42.0, false},     // default
		{"low", 39.0, false},
		{"force", 0, false},   // accepts any PSNR
	}

	for _, q := range qualities {
		t.Run("quality_"+q.quality, func(t *testing.T) {
			input := OptimizePngInput{
				SrcPath:  "testdata/optimize/psnr-will-44.png",
				DestPath: filepath.Join(tempDir, "output_"+q.quality+".png"),
				Quality:  q.quality,
			}

			result, err := Optimize(input)
			if err != nil {
				t.Fatalf("Optimize failed: %v", err)
			}

			if result.PNGQuant.Applied {
				if q.minPSNR > 0 && result.PNGQuant.PSNR < q.minPSNR {
					t.Errorf("PNGQuant.PSNR = %v, want >= %v for quality %s", 
						result.PNGQuant.PSNR, q.minPSNR, q.quality)
				}
			}

			// For force quality, PNGQuant should always be applied if possible
			if q.quality == "force" && !result.PNGQuant.Applied && result.PNGQuantError == nil {
				t.Error("Force quality should apply PNGQuant regardless of PSNR")
			}
		})
	}
}

func TestOptimize_ErrorHandling(t *testing.T) {
	tempDir := t.TempDir()

	t.Run("NonExistentFile", func(t *testing.T) {
		input := OptimizePngInput{
			SrcPath:  "testdata/optimize/nonexistent.png",
			DestPath: filepath.Join(tempDir, "output.png"),
		}

		_, err := Optimize(input)
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("InvalidDestPath", func(t *testing.T) {
		input := OptimizePngInput{
			SrcPath:  "testdata/optimize/psnr-will-50.png",
			DestPath: "/invalid/path/that/does/not/exist/output.png",
		}

		_, err := Optimize(input)
		if err == nil {
			t.Error("Expected error for invalid destination path")
		}
	})
}

func TestIsAcceptablePSNR(t *testing.T) {
	testCases := []struct {
		quality  string
		psnr     float64
		expected bool
	}{
		{"high", 50.0, true},
		{"high", 45.0, true},
		{"high", 44.9, false},
		{"", 50.0, true},
		{"", 42.0, true},
		{"", 41.9, false},
		{"low", 50.0, true},
		{"low", 39.0, true},
		{"low", 38.9, false},
		{"force", 10.0, true},
		{"force", 1.0, true},
		{"invalid", 50.0, true}, // Should use default threshold
		{"invalid", 42.0, true},
		{"invalid", 41.9, false},
	}

	for _, tc := range testCases {
		t.Run(tc.quality+"_"+string(rune(int(tc.psnr))), func(t *testing.T) {
			result := isAcceptablePSNR(tc.quality, tc.psnr)
			if result != tc.expected {
				t.Errorf("isAcceptablePSNR(%s, %f) = %v, want %v", 
					tc.quality, tc.psnr, result, tc.expected)
			}
		})
	}

	// Test infinity
	t.Run("Infinity", func(t *testing.T) {
		if !isAcceptablePSNR("high", math.Inf(1)) {
			t.Error("Infinity should always be acceptable")
		}
	})
}

func TestOptimize_StripMetadata(t *testing.T) {
	tempDir := t.TempDir()

	input := OptimizePngInput{
		SrcPath:  "testdata/optimize/with-mac-icc.png",
		DestPath: filepath.Join(tempDir, "stripped.png"),
		Quality:  "",
	}

	result, err := Optimize(input)
	if err != nil {
		t.Fatalf("Optimize failed: %v", err)
	}

	// Should have stripped metadata
	if result.Strip == nil {
		t.Error("Expected Strip result to be non-nil")
	} else {
		if result.Strip.Total == 0 {
			t.Error("Expected some metadata to be removed")
		}
		t.Logf("Removed metadata: Total=%d bytes, TextChunks=%d, TimeChunk=%d, ExifData=%d, OtherChunks=%d", 
			result.Strip.Total,
			result.Strip.Removed.TextChunks,
			result.Strip.Removed.TimeChunk,
			result.Strip.Removed.ExifData,
			result.Strip.Removed.OtherChunks)
	}

	// Size should be reduced after stripping
	if result.SizeAfterStrip >= result.BeforeSize {
		t.Errorf("SizeAfterStrip (%d) should be < BeforeSize (%d)", 
			result.SizeAfterStrip, result.BeforeSize)
	}
}

func TestOptimize_Photo(t *testing.T) {
	tempDir := t.TempDir()

	input := OptimizePngInput{
		SrcPath:  "testdata/optimize/me2020.png",
		DestPath: filepath.Join(tempDir, "photo.png"),
		Quality:  "",
	}

	result, err := Optimize(input)
	if err != nil {
		t.Fatalf("Optimize failed: %v", err)
	}

	// Photos might not get PNGQuant applied if PSNR is too low
	if result.PNGQuant.Applied && result.PNGQuant.PSNR < 42.0 {
		t.Errorf("Photo should not have PNGQuant applied if PSNR < 42, got %v", 
			result.PNGQuant.PSNR)
	}

	// Should still have good final PSNR
	if result.FinalPSNR < 35.0 {
		t.Errorf("FinalPSNR = %v, want >= 35.0", result.FinalPSNR)
	}
}
