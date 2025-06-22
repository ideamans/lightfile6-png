package png

import (
	"fmt"
	"math"
	"os"

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
	} else {
		output.Strip = stripResult
		pngData = strippedData
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
	} else {
		psnr, psnrErr := PngPsnr(beforePNGQuant, quantizedData)
		err = psnrErr
		if err != nil {
			return nil, NewDataErrorf(l10n.T("failed to calculate PSNR after quantization: %v"), err)
		}
		output.PNGQuant.PSNR = psnr
		if isAcceptablePSNR(input.Quality, psnr) {
			output.PNGQuant.Applied = true
			pngData = quantizedData
		}
	}
	output.SizeAfterPNGQuant = int64(len(pngData))

	// Calculate final PSNR before building comment
	finalPSNR, err := PngPsnr(originalData, pngData)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to calculate final PSNR: %w"), err)
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
		return &output, nil
	}

	// Write the comment
	commentedData, err := metaManager.WriteComment(pngData, comment)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to write comment: %w"), err)
	}
	pngData = commentedData

	// Re-calculate PSNR after adding comment to ensure it hasn't changed
	finalPSNRAfterComment, err := PngPsnr(originalData, pngData)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to calculate final PSNR after comment: %w"), err)
	}
	output.FinalPSNR = finalPSNRAfterComment

	// Check PSNR threshold (infinity is always acceptable)
	if !math.IsInf(finalPSNR, 1) && finalPSNR < PSNRThreshold {
		output.InspectionFailed = true
		return &output, nil
	}

	// Write the optimized PNG to destination path
	err = os.WriteFile(input.DestPath, pngData, 0600)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to write optimized PNG: %w"), err)
	}

	// Get file size after optimization
	destInfo, err := os.Stat(input.DestPath)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to stat destination file: %w"), err)
	}
	output.AfterSize = destInfo.Size()

	return &output, nil
}
