package png

import (
	"math"

	"github.com/ideamans/go-l10n"
	pngmetawebstrip "github.com/ideamans/go-png-meta-web-strip"
)

func init() {
	// Register Japanese translations for png.go error messages
	l10n.Register("ja", l10n.LexiconMap{
		"failed to read PNG file: %w":                      "PNGファイルの読み込みに失敗しました: %w",
		"failed to read PNG comment: %w":                   "PNGコメントの読み込みに失敗しました: %w",
		"failed to strip metadata: %v":                     "メタデータの削除に失敗しました: %v",
		"failed to calculate PSNR after quantization: %v":  "量子化後のPSNR計算に失敗しました: %v",
		"failed to calculate final PSNR: %w":               "最終PSNRの計算に失敗しました: %w",
		"failed to build comment: %w":                      "コメントの構築に失敗しました: %w",
		"failed to write comment: %w":                      "コメントの書き込みに失敗しました: %w",
		"failed to calculate final PSNR after comment: %w": "コメント追加後の最終PSNR計算に失敗しました: %w",
		"failed to write optimized PNG: %w":                "最適化されたPNGの書き込みに失敗しました: %w",
		"failed to stat destination file: %w":              "出力ファイルの情報取得に失敗しました: %w",
		// Log messages
		"Starting PNG optimization (quality: %s)": "PNG最適化を開始 (品質: %s)",
		"Already optimized by %s, skipping": "%sによって既に最適化されています、スキップします",
		"Failed to strip metadata: %v": "メタデータの削除に失敗: %v",
		"Stripped metadata - size: %s -> %s": "メタデータを削除 - サイズ: %s -> %s",
		"Failed to quantize: %v": "量子化に失敗: %v",
		"Applied PNGQuant - PSNR: %.2f dB, size: %s": "PNGQuant適用 - PSNR: %.2f dB, サイズ: %s",
		"Rejected PNGQuant - PSNR: %.2f (below threshold for quality: %s)": "PNGQuant却下 - PSNR: %.2f (品質 %s の閾値未満)",
		"Cannot optimize: final size (%s) >= original size (%s)": "最適化不可: 最終サイズ (%s) >= 元のサイズ (%s)",
		"PSNR inspection failed: %.2f dB < %.2f dB": "PSNR検査に失敗: %.2f dB < %.2f dB",
		"Writing optimized PNG": "最適化されたPNGを書き込み中",
		"Optimization completed: %s -> %s (%.1f%% reduction), PSNR: %.2f dB, PNGQuant: %v": "最適化完了: %s -> %s (%.1f%%削減), PSNR: %.2f dB, PNGQuant: %v",
		"Failed to read PNG file: %v": "PNGファイルの読み込みに失敗: %v",
		"Failed to read PNG comment: %v": "PNGコメントの読み込みに失敗: %v",
		"Failed to build comment: %v": "コメントの構築に失敗: %v",
		"Failed to write comment: %v": "コメントの書き込みに失敗: %v",
		"Failed to calculate final PSNR: %v": "最終PSNRの計算に失敗: %v",
		"Failed to write optimized PNG: %v": "最適化されたPNGの書き込みに失敗: %v",
		"Failed to stat destination file: %v": "出力ファイルの情報取得に失敗: %v",
	})
}

var (
	PSNRThreshold = 35.0
)

type OptimizePNGOutput struct {
	BeforeSize         int64
	AlreadyOptimized   bool
	AlreadyOptimizedBy string
	Strip              *pngmetawebstrip.Result
	StripError         error
	SizeAfterStrip     int64
	PNGQuant           struct {
		PSNR    float64
		Applied bool
	}
	SizeAfterPNGQuant int64
	PNGQuantError     error
	CantOptimize      bool
	InspectionFailed  bool
	FinalPSNR         float64
	AfterSize         int64
}

func isAcceptablePSNR(quality string, psnr float64) bool {
	if math.IsInf(psnr, 1) {
		return true
	}

	if quality == "high" {
		return psnr >= 45
	} else if quality == "low" {
		return psnr >= 39
	} else if quality == "force" {
		return true
	} else {
		return psnr >= 42
	}
}

