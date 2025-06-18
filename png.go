package png

import (
	"fmt"
	"math"
	"os"

	pngmetawebstrip "github.com/ideamans/go-png-meta-web-strip"
)

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
		return nil, fmt.Errorf("failed to read PNG file: %w", err)
	}
	output.BeforeSize = int64(len(pngData))

	// Check if already optimized using ReadComment
	comment, _, err := ReadComment(pngData)
	if err != nil {
		return nil, fmt.Errorf("failed to read PNG comment: %w", err)
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
		output.StripError = NewDataErrorf("failed to strip metadata: %v", err)
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
			return nil, NewDataErrorf("failed to calculate PSNR after quantization: %v", err)
		}
		output.PNGQuant.PSNR = psnr
		if isAcceptablePSNR(input.Quality, psnr) {
			output.PNGQuant.Applied = true
			pngData = quantizedData
		}
	}
	output.SizeAfterPNGQuant = int64(len(pngData))

	// Build comment with optimization information
	comment = &LightFileComment{
		By:       "LightFile",
		Before:   output.BeforeSize,
		After:    int64(len(pngData)),
		PNGQuant: output.PNGQuant.Applied,
	}

	// Calculate comment size and check if final size would exceed original
	commentJSON, commentSizeIncrease, err := BuildComment(comment)
	if err != nil {
		return nil, fmt.Errorf("failed to build comment: %w", err)
	}

	// Check if adding comment would make file larger than original
	currentSize := int64(len(pngData))
	finalSizeWithComment := currentSize + int64(commentSizeIncrease)
	if finalSizeWithComment >= output.BeforeSize {
		output.CantOptimize = true
		return &output, nil
	}

	// Write the comment
	commentedData, err := WriteComment(pngData, commentJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to write comment: %w", err)
	}
	pngData = commentedData

	// Calculate PSNR for quality inspection
	finalPSNR, err := PngPsnr(originalData, pngData)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate final PSNR: %w", err)
	}
	output.FinalPSNR = finalPSNR

	// Check PSNR threshold (infinity is always acceptable)
	if !math.IsInf(finalPSNR, 1) && finalPSNR < PSNRThreshold {
		output.InspectionFailed = true
		return &output, nil
	}

	// Write the optimized PNG to destination path
	err = os.WriteFile(input.DestPath, pngData, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to write optimized PNG: %w", err)
	}

	// Get file size after optimization
	destInfo, err := os.Stat(input.DestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat destination file: %w", err)
	}
	output.AfterSize = destInfo.Size()

	return &output, nil
}
