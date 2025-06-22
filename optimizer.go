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

// Optimizer is the main interface for PNG optimization
type Optimizer struct {
	Quality string
	Logger  Logger
}

// NewOptimizer creates a new PNG optimizer with the specified quality setting
func NewOptimizer(quality string) *Optimizer {
	opt := &Optimizer{
		Quality: quality,
	}
	return opt
}

// SetLogger sets the logger for this optimizer
func (o *Optimizer) SetLogger(logger Logger) {
	o.Logger = logger
}

// logDebug logs debug messages if logger is set
func (o *Optimizer) logDebug(format string, args ...interface{}) {
	if o.Logger != nil {
		o.Logger.Debug(format, args...)
	}
}

// logInfo logs info messages if logger is set
func (o *Optimizer) logInfo(format string, args ...interface{}) {
	if o.Logger != nil {
		o.Logger.Info(format, args...)
	}
}

// logWarn logs warning messages if logger is set
func (o *Optimizer) logWarn(format string, args ...interface{}) {
	if o.Logger != nil {
		o.Logger.Warn(format, args...)
	}
}

// logError logs error messages if logger is set
func (o *Optimizer) logError(format string, args ...interface{}) {
	if o.Logger != nil {
		o.Logger.Error(format, args...)
	}
}

// Run performs PNG optimization from srcPath to destPath
func (o *Optimizer) Run(srcPath, destPath string) (*OptimizePNGOutput, error) {
	o.logInfo("Starting PNG optimization (quality: %s)", o.Quality)
	output := OptimizePNGOutput{}

	// Read PNG file
	pngData, err := os.ReadFile(srcPath)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to read PNG file: %w"), err)
	}
	output.BeforeSize = int64(len(pngData))

	// Create metadata manager
	metaManager := &PNGMetaManager{}

	// Check if already optimized using ReadComment
	comment, _, err := metaManager.ReadComment(pngData)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to read PNG comment: %w"), err)
	}

	// If already optimized, return early
	if comment != nil && comment.By != "" {
		output.AlreadyOptimized = true
		output.AlreadyOptimizedBy = comment.By
		o.logInfo("Already optimized by %s, skipping", comment.By)
		return &output, nil
	}

	// Keep original data for PSNR comparison
	originalData := make([]byte, len(pngData))
	copy(originalData, pngData)

	// Strip metadata using pngmetawebstrip
	o.logDebug("Stripping metadata")
	strippedData, stripResult, err := pngmetawebstrip.Strip(pngData)
	if err != nil {
		// stripは外部パッケージで行うのでデータエラーの区別がない
		// しかし本質的にオンメモリのデータ処理だけなのでデータエラーとして扱う
		output.StripError = NewDataErrorf(l10n.T("failed to strip metadata: %v"), err)
		o.logWarn("Failed to strip metadata: %v", err)
	} else {
		output.Strip = stripResult
		pngData = strippedData
		o.logDebug("Stripped metadata - size: %s -> %s", humanize.Bytes(uint64(output.BeforeSize)), humanize.Bytes(uint64(len(pngData))))
	}
	output.SizeAfterStrip = int64(len(pngData))

	// PngquantはPSNRにより棄却する可能性がある
	beforePNGQuant := make([]byte, len(pngData))
	copy(beforePNGQuant, pngData)

	// Perform PNG quantization using Pngquant
	quantizedData, err := Pngquant(pngData)
	if err != nil {
		// Set quantize error and continue with stripped data
		output.PNGQuantError = err
		o.logWarn("Failed to quantize: %v", err)
	} else {
		// Calculate PSNR between before and after quantization
		psnrValue, err := psnr.Compute(beforePNGQuant, quantizedData)
		if err != nil {
			output.PNGQuantError = NewDataErrorf(l10n.T("failed to calculate PSNR after PNGQuant: %w"), err)
			o.logWarn("Failed to calculate PSNR after PNGQuant: %v", err)
		} else {
			output.PNGQuant.PSNR = psnrValue
			// Apply PNGQuant only if PSNR is acceptable
			if isAcceptablePSNR(o.Quality, psnrValue) {
				output.PNGQuant.Applied = true
				pngData = quantizedData
				o.logDebug("PNGQuant applied - PSNR: %.2f dB, size: %s", psnrValue, humanize.Bytes(uint64(len(pngData))))
			} else {
				o.logDebug("PNGQuant rejected - PSNR: %.2f dB below threshold", psnrValue)
			}
		}
	}
	output.SizeAfterPNGQuant = int64(len(pngData))

	// Calculate final PSNR between original and final
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
		PSNR:     MaybeInf(finalPSNR),
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
		o.logInfo("Cannot optimize: final size (%s) >= original size (%s)", 
			humanize.Bytes(uint64(finalSizeWithComment)), humanize.Bytes(uint64(output.BeforeSize)))
		return &output, nil
	}

	// Write the comment
	commentedData, err := metaManager.WriteComment(pngData, comment)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to write comment: %w"), err)
	}
	pngData = commentedData

	// Calculate PSNR for quality inspection
	output.FinalPSNR = finalPSNR

	// Check PSNR threshold (infinity is always acceptable)
	if !math.IsInf(finalPSNR, 1) && finalPSNR < PSNRThreshold {
		output.InspectionFailed = true
		o.logWarn("PSNR inspection failed: %.2f dB < %.2f dB", finalPSNR, PSNRThreshold)
		return &output, nil
	}

	// Write the optimized PNG to destination path
	err = os.WriteFile(destPath, pngData, 0644)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to write optimized PNG: %w"), err)
	}

	// Get file size after optimization
	destInfo, err := os.Stat(destPath)
	if err != nil {
		return nil, fmt.Errorf(l10n.T("failed to stat destination file: %w"), err)
	}
	output.AfterSize = destInfo.Size()

	o.logInfo("Optimization completed: %s -> %s (%.1f%% reduction), PSNR: %.2f dB",
		humanize.Bytes(uint64(output.BeforeSize)), humanize.Bytes(uint64(output.AfterSize)),
		float64(output.BeforeSize-output.AfterSize)/float64(output.BeforeSize)*100,
		finalPSNR)

	return &output, nil
}