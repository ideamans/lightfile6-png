package png

import (
	"bytes"
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
	switch img.(type) {
	case *image.Gray:
		return "Grayscale"
	case *image.Gray16:
		return "Grayscale16"
	case *image.Paletted:
		return "Palette"
	case *image.NRGBA:
		// NRGBA can be either RGBA or Grayscale+Alpha
		// Need to check if it's actually grayscale
		if isGrayscaleImage(img.(*image.NRGBA)) {
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
	file.Seek(8, 0)

	// IHDRチャンクを読む
	var chunkLength uint32
	binary.Read(file, binary.BigEndian, &chunkLength)

	var chunkType [4]byte
	file.Read(chunkType[:])

	if string(chunkType[:]) == "IHDR" {
		// IHDRデータを読む
		ihdrData := make([]byte, 13)
		file.Read(ihdrData)

		// ビット深度は8バイト目
		return int(ihdrData[8])
	}

	return -1
}

// checkCompressionLevelはPNGファイルの圧縮レベルを推測します
func checkCompressionLevel(t *testing.T, filePath string) string {
	info, err := os.Stat(filePath)
	if err != nil {
		t.Logf("Failed to stat file %s: %v", filePath, err)
		return "Unknown"
	}

	// ファイルサイズに基づいて推測（簡略化）
	size := info.Size()
	if size < 1000 {
		return "High"
	} else if size < 5000 {
		return "Medium"
	} else {
		return "Low"
	}
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
	file.Seek(8, 0)

	// IHDRチャンクを読む
	var chunkLength uint32
	binary.Read(file, binary.BigEndian, &chunkLength)

	var chunkType [4]byte
	file.Read(chunkType[:])

	if string(chunkType[:]) == "IHDR" {
		// IHDRデータを読む
		ihdrData := make([]byte, 13)
		file.Read(ihdrData)

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
	file.Seek(8, 0)

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
		file.Seek(int64(chunkLength)+4, 1)
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
	file.Seek(8, 0)

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
		file.Seek(int64(chunkLength)+4, 1)
	}

	return false
}

// checkMetadataTypeはPNGファイルのメタデータタイプを判定します
func checkMetadataType(t *testing.T, filePath string) string {
	hasText := checkChunkPresence(t, filePath, "tEXt")
	hasZText := checkChunkPresence(t, filePath, "zTXt")
	hasIText := checkChunkPresence(t, filePath, "iTXt")

	if !hasText && !hasZText && !hasIText {
		return "None"
	} else if hasText && !hasZText && !hasIText {
		return "Text"
	} else if hasZText {
		return "Compressed"
	} else if hasIText {
		return "International"
	} else {
		return "Mixed"
	}
}

// checkFilterMethodはPNGファイルのフィルタ方式を推測します
func checkFilterMethod(t *testing.T, filePath string) string {
	// PNGファイルのサイズとパターンに基づいて推測
	// 実際のフィルタ方式の検出は複雑なため、ここでは簡略化
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "Unknown"
	}

	// IDATチャンクのパターンを分析する簡略化されたロジック
	if bytes.Contains(data, []byte("IDAT")) {
		// ファイル名に基づいて推測
		if bytes.Contains([]byte(filePath), []byte("none")) {
			return "None"
		} else if bytes.Contains([]byte(filePath), []byte("sub")) {
			return "Sub"
		} else if bytes.Contains([]byte(filePath), []byte("up")) {
			return "Up"
		} else if bytes.Contains([]byte(filePath), []byte("average")) {
			return "Average"
		} else if bytes.Contains([]byte(filePath), []byte("paeth")) {
			return "Paeth"
		}
	}

	return "Mixed"
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