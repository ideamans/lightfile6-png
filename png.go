package png

import (
	"fmt"
	"math"
	"os"

	"github.com/dustin/go-humanize"
	"github.com/ideamans/go-l10n"
	pngmetawebstrip "github.com/ideamans/go-png-meta-web-strip"
	"github.com/ideamans/go-psnr"
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

type OptimizePngInput struct {
	SrcPath  string
	DestPath string
	Quality  string
}

var (
	PSNRThreshold = 35.0
)

type OptimizePngOutput struct {
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

func Optimize(input OptimizePngInput) (*OptimizePngOutput, error) {
	logInfo("Starting PNG optimization (quality: %s)", input.Quality)
	output := OptimizePngOutput{}

	// Read PNG file
	pngData, err := os.ReadFile(input.SrcPath)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to read PNG file: %w"), err)
	}
	output.BeforeSize = int64(len(pngData))

	// Create metadata manager
	metaManager := &PngMetaManager{}

	// Check if already optimized using ReadComment
	comment, _, err := metaManager.ReadComment(pngData)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to read PNG comment: %w"), err)
	}

	// If already optimized, return early
	if comment != nil && comment.By != "" {
		output.AlreadyOptimized = true
		output.AlreadyOptimizedBy = comment.By
		logInfo("Already optimized by %s, skipping", comment.By)
		return &output, nil
	}

	// Keep original data for PSNR comparison
	originalData := make([]byte, len(pngData))
	copy(originalData, pngData)

	// Strip metadata using pngmetawebstrip
	strippedData, stripResult, err := pngmetawebstrip.Strip(pngData)
	if err != nil {
		// stripは外部パッケージで行うのでデータエラーの区別がない
		// しかし本質的にオンメモリのデータ処理だけなのでデータエラーとして扱う
		output.StripError = NewDataErrorf(l10n.T("failed to strip metadata: %v"), err)
		logWarn("Failed to strip metadata: %v", err)
	} else {
		output.Strip = stripResult
		pngData = strippedData
		logDebug("Stripped metadata - size: %s -> %s", humanize.Bytes(uint64(output.BeforeSize)), humanize.Bytes(uint64(len(pngData))))
	}
	output.SizeAfterStrip = int64(len(pngData))

	// PngquantはPsnrにより棄却する可能性がある
	beforePNGQuant := make([]byte, len(pngData))
	copy(beforePNGQuant, pngData)

	// Perform PNG quantization using Pngquant
	quantizedData, err := Pngquant(pngData)
	if err != nil {
		// Set quantize error and continue with stripped data
		output.PNGQuantError = err
		logWarn("Failed to quantize: %v", err)
	} else {
		psnrValue, psnrErr := psnr.Compute(beforePNGQuant, quantizedData)
		err = psnrErr
		if err != nil {
			return nil, NewDataErrorf(l10n.T("failed to calculate PSNR after quantization: %v"), err)
		}
		output.PNGQuant.PSNR = psnrValue
		if isAcceptablePSNR(input.Quality, psnrValue) {
			output.PNGQuant.Applied = true
			pngData = quantizedData
			logDebug("Applied PNGQuant - PSNR: %.2f dB, size: %s", psnrValue, humanize.Bytes(uint64(len(quantizedData))))
		} else {
			logDebug("Rejected PNGQuant - PSNR: %f (below threshold for quality: %s)", psnrValue, input.Quality)
		}
	}
	output.SizeAfterPNGQuant = int64(len(pngData))

	// Calculate final PSNR before building comment
	finalPSNR, err := psnr.Compute(originalData, pngData)
	if err != nil {
		return nil, NewDataErrorf(l10n.T("failed to calculate final PSNR: %w"), err)
	}

	// Build comment with optimization information
	comment = &LightFileComment{
		By:       "LightFile",
		Before:   output.BeforeSize,
		After:    int64(len(pngData)),
		PNGQuant: output.PNGQuant.Applied,
		Psnr:     MaybeInf(finalPSNR),
	}

	// Calculate comment size and check if final size would exceed original
	_, commentSizeIncrease, err := metaManager.BuildComment(comment)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to build comment: %w"), err)
	}

	// Check if adding comment would make file larger than original
	currentSize := int64(len(pngData))
	finalSizeWithComment := currentSize + int64(commentSizeIncrease)
	if finalSizeWithComment >= output.BeforeSize {
		output.CantOptimize = true
		logInfo("Cannot optimize: final size (%s) >= original size (%s)", humanize.Bytes(uint64(finalSizeWithComment)), humanize.Bytes(uint64(output.BeforeSize)))
		return &output, nil
	}

	// Write the comment
	commentedData, err := metaManager.WriteComment(pngData, comment)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to write comment: %w"), err)
	}
	pngData = commentedData

	// Re-calculate PSNR after adding comment to ensure it hasn't changed
	finalPSNRAfterComment, err := psnr.Compute(originalData, pngData)
	if err != nil {
		return nil, NewDataErrorf(l10n.T("failed to calculate final PSNR after comment: %w"), err)
	}
	output.FinalPSNR = finalPSNRAfterComment

	// Check PSNR threshold (infinity is always acceptable)
	if !math.IsInf(finalPSNR, 1) && finalPSNR < PSNRThreshold {
		output.InspectionFailed = true
		logWarn("PSNR inspection failed: %.2f dB < %.2f dB", finalPSNR, PSNRThreshold)
		return &output, nil
	}

	// Write the optimized PNG to destination path
	logDebug("Writing optimized PNG")
	err = os.WriteFile(input.DestPath, pngData, 0600)
	if err != nil {
		logError(l10n.T("Failed to write optimized PNG: %v"), err)
		return nil, fmt.Errorf(l10n.T("failed to write optimized PNG: %w"), err)
	}

	// Get file size after optimization
	destInfo, err := os.Stat(input.DestPath)
	if err != nil {
		logError(l10n.T("Failed to stat destination file: %v"), err)
		return nil, fmt.Errorf(l10n.T("failed to stat destination file: %w"), err)
	}
	output.AfterSize = destInfo.Size()

	logInfo("Optimization completed: %s -> %s (%.1f%% reduction), PSNR: %.2f dB, PNGQuant: %v",
		humanize.Bytes(uint64(output.BeforeSize)), humanize.Bytes(uint64(output.AfterSize)),
		float64(output.BeforeSize-output.AfterSize)/float64(output.BeforeSize)*100,
		finalPSNRAfterComment,
		output.PNGQuant.Applied)

	return &output, nil
}
