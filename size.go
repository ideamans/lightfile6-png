package png

import "math"

// SizesWithinTolerance は2つのファイルサイズが指定された許容範囲内で一致しているかをチェックします。
// toleranceは0.01で1%の許容範囲を意味します。
func SizesWithinTolerance(actual, expected int64, tolerance float64) bool {
	if expected == 0 {
		return actual == 0
	}
	diff := math.Abs(float64(actual - expected))
	percentDiff := diff / float64(expected)
	return percentDiff <= tolerance
}

// SizesWithin1Percent は2つのファイルサイズが1%の許容範囲内で一致しているかをチェックします。
func SizesWithin1Percent(actual, expected int64) bool {
	return SizesWithinTolerance(actual, expected, 0.01)
}
