package png

// TestCreateOptimizer creates an optimizer for testing
func TestCreateOptimizer(quality string) *Optimizer {
	return NewOptimizer(quality)
}

// TestRunOptimization runs optimization and is a helper for tests
func TestRunOptimization(quality, srcPath, destPath string) (*OptimizePNGOutput, error) {
	opt := NewOptimizer(quality)
	return opt.Run(srcPath, destPath)
}