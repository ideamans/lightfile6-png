package png

import (
	"encoding/binary"
	"image"
	"image/png"
	"os"
	"testing"
)

// checkColorTypeはPNGファイルのカラータイプを判定します
func checkColorType(t *testing.T, filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		t.Logf("Failed to open file %s: %v", filePath, err)
		return "Unknown"
	}
	defer file.Close()

	// PNGをデコードして確認
	img, err := png.Decode(file)
	if err != nil {
		t.Logf("Failed to decode PNG %s: %v", filePath, err)
		return "Unknown"
	}

	// First check concrete image type
	switch img := img.(type) {
	case *image.Gray:
		return "Grayscale"
	case *image.Gray16:
		return "Grayscale16"
	case *image.Paletted:
		return "Palette"
	case *image.NRGBA:
		// NRGBA can be either RGBA or Grayscale+Alpha
		// Need to check if it's actually grayscale
		if isGrayscaleImage(img) {
			return "Grayscale+Alpha"
		}
		return "RGBA"
	case *image.NRGBA64:
		return "RGBA64"
	case *image.RGBA:
		return "RGBA"
	case *image.RGBA64:
		return "RGBA64"
	default:
		return "RGB"
	}
}

// checkBitDepthはPNGファイルのビット深度を判定します
func checkBitDepth(t *testing.T, filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		t.Logf("Failed to open file %s: %v", filePath, err)
		return -1
	}
	defer file.Close()

	// PNG signature をスキップ
	if _, err := file.Seek(8, 0); err != nil {
		return -1
	}

	// IHDRチャンクを読む
	var chunkLength uint32
	if err := binary.Read(file, binary.BigEndian, &chunkLength); err != nil {
		return -1
	}

	var chunkType [4]byte
	if _, err := file.Read(chunkType[:]); err != nil {
		return -1
	}

	if string(chunkType[:]) == "IHDR" {
		// IHDRデータを読む
		ihdrData := make([]byte, 13)
		if _, err := file.Read(ihdrData); err != nil {
			return -1
		}

		// ビット深度は8バイト目
		return int(ihdrData[8])
	}

	return -1
}

// checkInterlaceはPNGファイルがインターレース方式を使用しているかを判定します
func checkInterlace(t *testing.T, filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		t.Logf("Failed to open file %s: %v", filePath, err)
		return "Unknown"
	}
	defer file.Close()

	// PNG signature をスキップ
	if _, err := file.Seek(8, 0); err != nil {
		return "Unknown"
	}

	// IHDRチャンクを読む
	var chunkLength uint32
	if err := binary.Read(file, binary.BigEndian, &chunkLength); err != nil {
		return "Unknown"
	}

	var chunkType [4]byte
	if _, err := file.Read(chunkType[:]); err != nil {
		return "Unknown"
	}

	if string(chunkType[:]) == "IHDR" {
		// IHDRデータを読む
		ihdrData := make([]byte, 13)
		if _, err := file.Read(ihdrData); err != nil {
			return "Unknown"
		}

		// インターレース方式は12バイト目
		interlace := ihdrData[12]
		if interlace == 0 {
			return "None"
		} else if interlace == 1 {
			return "Adam7"
		}
	}

	return "Unknown"
}

// hasTRNSChunkはPNGファイルにtRNS（透明度）チャンクがあるかどうかを確認します
func hasTRNSChunk(t *testing.T, filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		t.Logf("Failed to open file %s: %v", filePath, err)
		return false
	}
	defer file.Close()

	// Skip PNG signature
	if _, err := file.Seek(8, 0); err != nil {
		return false
	}

	// Read chunks until we find tRNS or reach end
	for {
		var chunkLength uint32
		var chunkType [4]byte

		// Read chunk length (4 bytes)
		if err := binary.Read(file, binary.BigEndian, &chunkLength); err != nil {
			break
		}

		// Read chunk type (4 bytes)
		if _, err := file.Read(chunkType[:]); err != nil {
			break
		}

		if string(chunkType[:]) == "tRNS" {
			return true
		}

		// Skip chunk data + CRC (4 bytes)
		if _, err := file.Seek(int64(chunkLength)+4, 1); err != nil {
			break
		}
	}

	return false
}

// checkChunkPresenceは指定されたチャンクタイプがPNGファイルに存在するかを確認します
func checkChunkPresence(t *testing.T, filePath string, targetChunk string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		t.Logf("Failed to open file %s: %v", filePath, err)
		return false
	}
	defer file.Close()

	// Skip PNG signature
	if _, err := file.Seek(8, 0); err != nil {
		return false
	}

	// Read chunks
	for {
		var chunkLength uint32
		var chunkType [4]byte

		// Read chunk length (4 bytes)
		if err := binary.Read(file, binary.BigEndian, &chunkLength); err != nil {
			break
		}

		// Read chunk type (4 bytes)
		if _, err := file.Read(chunkType[:]); err != nil {
			break
		}

		if string(chunkType[:]) == targetChunk {
			return true
		}

		// Skip chunk data + CRC (4 bytes)
		if _, err := file.Seek(int64(chunkLength)+4, 1); err != nil {
			break
		}
	}

	return false
}

// isGrayscaleImage checks if an NRGBA image is actually grayscale
func isGrayscaleImage(img *image.NRGBA) bool {
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			c := img.NRGBAAt(x, y)
			if c.R != c.G || c.G != c.B {
				return false
			}
		}
	}
	return true
}

// runVariationOptimization is a helper function for variation tests
func runVariationOptimization(t *testing.T, inputPath, outputPath, quality string) (*OptimizePNGOutput, error) {
	t.Helper()
	return TestRunOptimization(quality, inputPath, outputPath)
}
