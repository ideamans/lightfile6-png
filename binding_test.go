package png

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPngquantNormal(t *testing.T) {
	cases := []struct {
		name       string
		file       string
		beforeSize int64
		afterSize  int64
	}{
		{
			name:       "通常ファイル",
			file:       "psnr-will-50.png",
			beforeSize: 28038,
			afterSize:  10545,
		},
	}

	for _, tc := range cases {
		inputPath := filepath.Join("./testdata/binding", tc.file)

		// ファイルからバイトデータを読み込み
		inputData, err := os.ReadFile(inputPath)
		if err != nil {
			t.Fatalf("os.ReadFile(%s) = %v; want nil", inputPath, err)
		}

		if !SizesWithin1Percent(int64(len(inputData)), tc.beforeSize) {
			t.Errorf("input data size = %d; want %d (within 1%% tolerance)", len(inputData), tc.beforeSize)
		}

		// Pngquantを実行
		outputData, wasQuantized, err := Pngquant(inputData)
		if err != nil {
			t.Errorf("Pngquant(inputData) = %v; want nil", err)
		}

		// 通常のファイルは量子化されるはず
		if !wasQuantized {
			t.Errorf("wasQuantized = false; want true")
		}

		if !SizesWithin1Percent(int64(len(outputData)), tc.afterSize) {
			t.Errorf("output data size = %d; want %d (within 1%% tolerance)", len(outputData), tc.afterSize)
		}

		// 軽量化されたデータをデコードできること
		_, err = decodeRgbaPng(outputData)
		if err != nil {
			t.Errorf("decodeRgbaPng(outputData) = %v; want nil", err)
		}
	}
}

func TestPngquantError(t *testing.T) {
	cases := []struct {
		name         string
		file         string
		errorMessage string
	}{
		{
			name:         "実態がJPEGのファイル",
			file:         "jpeg.png",
			errorMessage: "failed to decode first in pngquant < failed to decode < png: invalid format: not a PNG file",
		},
		{
			name:         "破損したファイル",
			file:         "bad.png",
			errorMessage: "failed to decode first in pngquant < failed to decode < unexpected EOF",
		},
	}

	for _, tc := range cases {
		inputPath := filepath.Join("./testdata/binding", tc.file)

		// ファイルからバイトデータを読み込み
		inputData, err := os.ReadFile(inputPath)
		if err != nil {
			t.Fatalf("os.ReadFile(%s) = %v; want nil", inputPath, err)
		}

		_, _, err = Pngquant(inputData)

		if err == nil {
			t.Fatalf("Pngquant(inputData) = nil; エラーになるはず")
		} else {
			// エラーメッセージ内のバックスラッシュを/に置換してから比較
			actualError := strings.ReplaceAll(err.Error(), "\\", "/")
			if actualError != tc.errorMessage {
				t.Errorf("Pngquant(inputData) = %v; want %s", actualError, tc.errorMessage)
			}
		}
	}
}

func TestNRGBAImage(t *testing.T) {
	cases := []struct {
		name string
		file string
	}{
		{
			name: "NRGBA形式の画像",
			file: "psnr-will-50.png",
		},
	}

	for _, tc := range cases {
		inputPath := filepath.Join("./testdata/binding", tc.file)

		// ファイルからバイトデータを読み込み
		inputData, err := os.ReadFile(inputPath)
		if err != nil {
			t.Fatalf("os.ReadFile(%s) = %v; want nil", inputPath, err)
		}

		outputData, wasQuantized, err := Pngquant(inputData)
		if err != nil {
			t.Errorf("Pngquant(inputData) = %v; want nil", err)
		}

		// NRGBA形式も量子化されるはず
		if !wasQuantized {
			t.Errorf("wasQuantized = false; want true")
		}

		// 軽量化されたデータをデコードできること
		_, err = decodeRgbaPng(outputData)
		if err != nil {
			t.Errorf("decodeRgbaPng(outputData) = %v; want nil", err)
		}
	}
}

func TestAlready8bitPng(t *testing.T) {
	cases := []struct {
		name    string
		file    string
		theSize int64
	}{
		{
			name:    "すでに8bitのPNG",
			file:    "psnr-will-50-fs8.png",
			theSize: 9985,
		},
	}

	for _, tc := range cases {
		inputPath := filepath.Join("./testdata/binding", tc.file)

		// ファイルからバイトデータを読み込み
		inputData, err := os.ReadFile(inputPath)
		if err != nil {
			t.Fatalf("os.ReadFile(%s) = %v; want nil", inputPath, err)
		}

		if !SizesWithin1Percent(int64(len(inputData)), tc.theSize) {
			t.Errorf("input data size = %d; want %d (within 1%% tolerance)", len(inputData), tc.theSize)
		}

		outputData, wasQuantized, err := Pngquant(inputData)
		if err != nil {
			t.Errorf("Pngquant(inputData) = %v; want nil", err)
		}

		// すでに8bitのPNGは量子化されないはず
		if wasQuantized {
			t.Errorf("wasQuantized = true; want false")
		}

		if !SizesWithin1Percent(int64(len(outputData)), tc.theSize) {
			t.Errorf("output data size = %d; want %d (within 1%% tolerance)", len(outputData), tc.theSize)
		}

		// 軽量化されたデータをデコードできること
		_, err = decodeRgbaPng(outputData)
		if err != nil {
			t.Errorf("decodeRgbaPng(outputData) = %v; want nil", err)
		}
	}
}
