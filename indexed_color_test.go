package png

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

func TestPngquantWithIndexedColor(t *testing.T) {
	tests := []struct {
		name          string
		createFunc    func() ([]byte, error)
		wantQuantized bool
	}{
		{
			name: "Indexed color PNG should not be quantized",
			createFunc: func() ([]byte, error) {
				// Create a palette with 256 colors
				palette := make([]color.Color, 256)
				for i := 0; i < 256; i++ {
					palette[i] = color.RGBA{uint8(i), uint8(i), uint8(i), 255}
				}
				img := image.NewPaletted(image.Rect(0, 0, 16, 16), palette)
				// Fill with some pattern
				for y := 0; y < 16; y++ {
					for x := 0; x < 16; x++ {
						img.SetColorIndex(x, y, uint8((x+y)%256))
					}
				}
				var buf bytes.Buffer
				err := png.Encode(&buf, img)
				return buf.Bytes(), err
			},
			wantQuantized: false,
		},
		{
			name: "RGBA PNG should be quantized",
			createFunc: func() ([]byte, error) {
				img := image.NewRGBA(image.Rect(0, 0, 16, 16))
				// Fill with gradient
				for y := 0; y < 16; y++ {
					for x := 0; x < 16; x++ {
						r := uint8(x * 255 / 16)
						g := uint8(y * 255 / 16)
						b := uint8(128)
						img.Set(x, y, color.RGBA{r, g, b, 255})
					}
				}
				var buf bytes.Buffer
				err := png.Encode(&buf, img)
				return buf.Bytes(), err
			},
			wantQuantized: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test PNG
			pngData, err := tt.createFunc()
			if err != nil {
				t.Fatalf("Failed to create test PNG: %v", err)
			}

			// Call Pngquant
			outputData, wasQuantized, err := Pngquant(pngData)
			if err != nil {
				t.Fatalf("Pngquant failed: %v", err)
			}

			// Check if quantization was applied as expected
			if wasQuantized != tt.wantQuantized {
				t.Errorf("wasQuantized = %v, want %v", wasQuantized, tt.wantQuantized)
			}

			// Verify output is valid PNG
			if len(outputData) == 0 {
				t.Error("Output data is empty")
			}

			// For indexed color images, output should be same as input
			if !tt.wantQuantized {
				if !bytes.Equal(pngData, outputData) {
					t.Error("Expected output to be identical to input for indexed color PNG")
				}
			}
		})
	}
}

func TestOptimizerWithIndexedColor(t *testing.T) {
	tempDir := t.TempDir()

	// Create indexed color PNG
	palette := make([]color.Color, 256)
	for i := 0; i < 256; i++ {
		palette[i] = color.RGBA{uint8(i), 0, 0, 255}
	}
	img := image.NewPaletted(image.Rect(0, 0, 16, 16), palette)
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.SetColorIndex(x, y, uint8((x*16+y)%256))
		}
	}

	// Encode to file
	srcPath := tempDir + "/indexed.png"
	destPath := tempDir + "/indexed_out.png"

	f, err := os.Create(srcPath)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	err = png.Encode(f, img)
	f.Close()
	if err != nil {
		t.Fatalf("Failed to encode PNG: %v", err)
	}

	// Run optimizer
	opt := NewOptimizer("medium")
	output, err := opt.Run(srcPath, destPath)
	if err != nil {
		t.Fatalf("Optimization failed: %v", err)
	}

	// Check that indexed color was detected
	if !output.IsIndexedColor {
		t.Error("Expected IsIndexedColor to be true")
	}

	// Check that PNGQuant was not applied
	if output.PNGQuant.Applied {
		t.Error("Expected PNGQuant.Applied to be false for indexed color image")
	}

	// PSNR should be 0 since PNGQuant didn't run
	if output.PNGQuant.PSNR != 0 {
		t.Errorf("Expected PNGQuant.PSNR to be 0, got %f", output.PNGQuant.PSNR)
	}
}
