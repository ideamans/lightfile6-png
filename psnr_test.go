package png

import (
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestPsnr(t *testing.T) {
	cases := []struct {
		name   string
		input1 string
		input2 string
		want   float64
	}{
		{
			name:   "PSNR 50",
			input1: "psnr-will-50.png",
			input2: "psnr-will-50-fs8.png",
			want:   50.329482262403424,
		},
		{
			name:   "PSNR 48",
			input1: "psnr-will-48.png",
			input2: "psnr-will-48-fs8.png",
			want:   48.3450745165523,
		},
		{
			name:   "PSNR 44",
			input1: "psnr-will-44.png",
			input2: "psnr-will-44-fs8.png",
			want:   44.414379903161375,
		},
		{
			name:   "PSNR 27",
			input1: "psnr-will-27.png",
			input2: "psnr-will-27-fs8.png",
			want:   27.905236383078247,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Load PNG files as byte data
			data1, err := os.ReadFile(filepath.Join("./testdata/psnr", tc.input1))
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tc.input1, err)
			}

			data2, err := os.ReadFile(filepath.Join("./testdata/psnr", tc.input2))
			if err != nil {
				t.Fatalf("Failed to read %s: %v", tc.input2, err)
			}

			psnr, err := PngPsnr(data1, data2)
			if err != nil {
				t.Errorf("CalculatePsnr(%s, %s) = %v", tc.input1, tc.input2, err)
			}

			if math.Abs(psnr-tc.want) > 0.1 {
				t.Errorf("CalculatePsnr(%s, %s) = %v, want %v", tc.input1, tc.input2, psnr, tc.want)
			}
		})
	}
}
