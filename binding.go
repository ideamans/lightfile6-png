package png

//go:generate git submodule update --init --recursive
//go:generate make

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"unsafe"

	"github.com/ideamans/go-l10n"
)

/*
#cgo LDFLAGS: -Llibimagequant/imagequant-sys -limagequant
#cgo linux LDFLAGS: -lm -ldl
#cgo windows LDFLAGS: -lpthread -lgcc -lwsock32 -lws2_32 -lbcrypt -lntdll -luserenv
#include <libimagequant/imagequant-sys/libimagequant.h>
*/
import "C"

func init() {
	// Register Japanese translations for this file
	l10n.Register("ja", l10n.LexiconMap{
		"png: failed to decode < %v":                   "png: デコードに失敗しました < %v",
		"png: unsupported image type on decoding":      "png: デコード時にサポートされていない画像タイプです",
		"png: failed to decode first in pngquant < %v": "png: pngquantの最初のデコードに失敗しました < %v",
		"png: failed to quantize with %s (code %d)":    "png: quantizeに失敗しました: %s (コード %d)",
		"png: failed to encode pngquant < %v":          "png: pngquantのエンコードに失敗しました < %v",
	})
}

// libimagequantライブラリが返すCGOエラーコード。
// これらの定数は、libimagequant.hで定義されたエラーコードに直接マップされています
const (
	// LIQ_OK は操作成功を示します
	LIQ_OK = 0
	// QualityTooLow は量子化結果が品質要件を満たさないことを示します
	QualityTooLow = 99
	// ValueOutOfRange は無効なパラメータ値が提供されたことを示します
	ValueOutOfRange = 100
	// OutOfMemory はメモリ割り当て失敗を示します
	OutOfMemory = 101
	// Aborted は操作がキャンセルされたことを示します
	Aborted = 102
	// InternalError はライブラリ内部エラーを示します
	InternalError = 103
	// BufferTooSmall は提供されたバッファが不十分であることを示します
	BufferTooSmall = 104
	// InvalidPointer はnullまたは無効なポインタが提供されたことを示します
	InvalidPointer = 105
	// Unsupported は操作がサポートされていないことを示します
	Unsupported = 106
)

// translateError はlibimagequantエラーコードを人間が読める文字列に変換します。
// この関数はエラーレポートとデバッグ目的で使用されます。
func translateError(code int) string {
	switch code {
	case LIQ_OK:
		return "LIQ_OK"
	case QualityTooLow:
		return "QualityTooLow"
	case ValueOutOfRange:
		return "ValueOutOfRange"
	case OutOfMemory:
		return "OutOfMemory"
	case Aborted:
		return "Aborted"
	case InternalError:
		return "InternalError"
	case BufferTooSmall:
		return "BufferTooSmall"
	case InvalidPointer:
		return "InvalidPointer"
	case Unsupported:
		return "Unsupported"
	}
	return "Unknown"
}

// decodeRgbaPng はPNGバイトデータをRGBAビットマップデータにデコードします。
// この関数は、pngquantとの互換性を保証するためにカラーモデル変換を処理します:
//   - RGBA画像は直接処理されます
//   - NRGBA画像はRGBAに変換されます（非事前乗算から事前乗算アルファへ）
//   - パレット画像はnilを返します（すでにインデックスカラー、量子化不要）
//   - その他のカラーモデルはエラーを返します
//
// この関数はpngquant前処理専用に設計されており、
// すべてのPNGカラーモデルを包括的に処理しない可能性があります。
//
// TODO: この関数は、任意のカラーモデルを処理するために描画ベースのアプローチを使用すべきです。
func decodeRgbaPng(data []byte) (*image.RGBA, error) {
	reader := bytes.NewReader(data)

	img, err := png.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("png: failed to decode < %v"), err)
	}

	if _, ok := img.ColorModel().(color.Palette); ok {
		return nil, nil
	} else if nrgba, ok := img.(*image.NRGBA); ok {
		rgba := convertNRGBAToRGBA(nrgba)
		return rgba, nil
	} else if rgba, ok := img.(*image.RGBA); ok {
		return rgba, nil
	}

	return nil, fmt.Errorf(l10n.T("png: unsupported image type on decoding"))
}

// convertNRGBAToRGBA はNRGBAフォーマットの画像をRGBAフォーマットに変換します。
// NRGBAは独立したアルファチャンネルを持ち、RGBAはRGBとアルファが事前乗算されています。
// この変換は、pngquant（libimagequant）がRGBAフォーマットを期待するために必要です。
//
// 各ピクセルの変換公式:
//
//	RGBA.R = (NRGBA.R * NRGBA.A) / 255
//	RGBA.G = (NRGBA.G * NRGBA.A) / 255
//	RGBA.B = (NRGBA.B * NRGBA.A) / 255
//	RGBA.A = NRGBA.A
//
// TODO: この関数は、任意のソースカラーモデルを処理できる、より一般的な描画ベースのアプローチで置き換えるべきです。
func convertNRGBAToRGBA(src *image.NRGBA) *image.RGBA {
	dst := image.NewRGBA(src.Rect)
	for y := src.Rect.Min.Y; y < src.Rect.Max.Y; y++ {
		for x := src.Rect.Min.X; x < src.Rect.Max.X; x++ {
			nrgba := src.NRGBAAt(x, y)
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(uint16(nrgba.R) * uint16(nrgba.A) / 255),
				G: uint8(uint16(nrgba.G) * uint16(nrgba.A) / 255),
				B: uint8(uint16(nrgba.B) * uint16(nrgba.A) / 255),
				A: nrgba.A,
			})
		}
	}

	return dst
}

// Pngquant はCGO経由でlibimagequantライブラリを使用してPNG画像の色量子化を実行します。
// この関数は、Rustベースのpngquant実装のためのGoインターフェースを提供します。
//
// 量子化プロセス:
//  1. 入力PNGをRGBAフォーマットにデコード
//  2. 速度と品質設定でlibimagequantを設定
//  3. RGBAデータからlibimagequant画像オブジェクトを作成
//  4. 色量子化を実行（k-meansクラスタリング）
//  5. ディザリングとガンマ補正を適用
//  6. パレットとインデックス付き画像データを生成
//  7. パレット付きPNGとして結果をエンコード
//
// 品質設定（pngquant CLIデフォルトと一致）:
//   - 速度レベル: 4（品質とパフォーマンスのバランス）
//   - 品質範囲: 0-100（全範囲許可）
//   - ディザリングレベル: 1.0（Floyd-Steinbergディザリング）
//   - ガンマ設定: 0.45455（標準sRGBガンマ）
//
// パレット画像の場合、すでにインデックスカラーフォーマットであるため、
// 関数は単純に入力をそのまま返します。
//
// 量子化が失敗した場合にエラーを返します。
func Pngquant(data []byte) ([]byte, error) {
	sample, err := decodeRgbaPng(data)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("png: failed to decode first in pngquant < %v"), err)
	}

	if sample == nil {
		// すでにインデックスカラーの画像なのでそのまま返す
		return data, nil
	}

	handle := C.liq_attr_create()
	defer C.liq_attr_destroy(handle)

	C.liq_set_speed(handle, 4)
	C.liq_set_quality(handle, 0, 100)

	raw_rgba_pixels := (unsafe.Pointer)(&sample.Pix[0])
	w := C.int(sample.Rect.Dx())
	h := C.int(sample.Rect.Dy())
	input := C.liq_image_create_rgba(handle, raw_rgba_pixels, w, h, 0)
	defer C.liq_image_destroy(input)

	var result *C.liq_result
	quantize_result := C.liq_image_quantize(input, handle, &result)
	if quantize_result != LIQ_OK {
		phrase := translateError(int(quantize_result))
		return nil, fmt.Errorf(l10n.T("png: failed to quantize with %s (code %d)"), phrase, quantize_result)
	}
	defer C.liq_result_destroy(result)

	// pngquantのソースを見ると以下のように設定している
	// https://github.com/kornelski/pngquant/blob/main/pngquant.c#L209
	C.liq_set_dithering_level(result, 1.0)
	C.liq_set_output_gamma(result, 0.45455)

	pixels_size := C.size_t(w * h)
	raw_8bit_pixels := make([]byte, pixels_size)
	C.liq_set_dithering_level(result, 1.0)

	C.liq_write_remapped_image(result, input, (unsafe.Pointer)(&raw_8bit_pixels[0]), pixels_size)
	palette := C.liq_get_palette(result)

	quantizedPalette := make([]color.Color, int(palette.count))
	for i := 0; i < int(palette.count); i++ {
		quantizedPalette[i] = color.RGBA{
			R: uint8(palette.entries[i].r),
			G: uint8(palette.entries[i].g),
			B: uint8(palette.entries[i].b),
			A: uint8(palette.entries[i].a),
		}
	}

	paletted := image.NewPaletted(sample.Rect, quantizedPalette)
	for y := 0; y < sample.Rect.Dy(); y++ {
		for x := 0; x < sample.Rect.Dx(); x++ {
			paletted.SetColorIndex(x, y, raw_8bit_pixels[y*sample.Rect.Dx()+x])
		}
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, paletted)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("png: failed to encode pngquant < %v"), err)
	}

	return buf.Bytes(), nil
}
