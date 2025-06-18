package png

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"math"

	"github.com/ideamans/go-l10n"
)

func init() {
	// Register Japanese translations for this file
	l10n.Register("ja", l10n.LexiconMap{
		"png: failed to decode as png < %v": "png: PNGとしてデコードできませんでした < %v",
	})
}

// loadPngFromBytes はバイトデータからPNG画像をデコードします。
// この内部関数は、PSNR計算のためのPNGデコードを処理します。
// 異なるカラーモデルの柔軟な処理のために汎用的なimage.Imageインターフェースを返します。
func loadPngFromBytes(data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)

	img, err := png.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("png: failed to decode as png < %v"), err)
	}

	return img, nil
}

// PngPsnr は2つのPNG画像間のピーク信号対雑音比を計算します。
// PSNRは画像間の類似性を比較する品質測定であり、
// 高い値ほど良好な品質/類似性を示します。
//
// アルゴリズム:
//  1. 両方の画像を読み込み、同じ寸法であることを確認
//  2. RGB値をピクセル単位で比較（アルファチャンネルは無視）
//  3. 差分から平均二乗誤差（MSE）を計算
//  4. 公式を使用してPSNRを計算: 10 * log10(255² / MSE)
//  5. 画像が同一の場合（MSE = 0）は正の無限大を返す
//
// この関数は、異なるビット深度間での一貫した比較のために、
// 16ビットカラー値を右シフトで8ビットに変換して処理します。
//
// 戻り値:
//   - float64: デシベル（dB）単位のPSNR値
//   - error: 任意のI/Oまたは画像処理エラー
func PngPsnr(data1, data2 []byte) (float64, error) {
	var sum int64

	img1, err := loadPngFromBytes(data1)
	if err != nil {
		return 0, fmt.Errorf(l10n.T("png: failed to decode as png < %v"), err)
	}

	img2, err := loadPngFromBytes(data2)
	if err != nil {
		return 0, fmt.Errorf(l10n.T("png: failed to decode as png < %v"), err)
	}

	bounds := img1.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r1, g1, b1, _ := img1.At(x, y).RGBA()
			r2, g2, b2, _ := img2.At(x, y).RGBA()

			r := int64(r1) - int64(r2)
			g := int64(g1) - int64(g2)
			b := int64(b1) - int64(b2)

			// 各チャンネル16ビットから8ビットに変換
			sum += (r*r)>>16 + (g*g)>>16 + (b*b)>>16
		}
	}

	if sum == 0 {
		return math.Inf(1), nil
	}

	mse256 := float64(sum) / float64(bounds.Dx()*bounds.Dy()*3)
	maxValue := float64(255)

	psnr := 10 * math.Log10(maxValue*maxValue/mse256)

	return psnr, nil
}
