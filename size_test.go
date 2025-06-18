package png

import "testing"

func TestSizesWithinTolerance(t *testing.T) {
	cases := []struct {
		name      string
		actual    int64
		expected  int64
		tolerance float64
		want      bool
	}{
		{
			name:      "完全一致",
			actual:    1000,
			expected:  1000,
			tolerance: 0.01,
			want:      true,
		},
		{
			name:      "1%以内の差 (999)",
			actual:    999,
			expected:  1000,
			tolerance: 0.01,
			want:      true,
		},
		{
			name:      "1%以内の差 (1010)",
			actual:    1010,
			expected:  1000,
			tolerance: 0.01,
			want:      true,
		},
		{
			name:      "1%を超える差 (989)",
			actual:    989,
			expected:  1000,
			tolerance: 0.01,
			want:      false,
		},
		{
			name:      "1%を超える差 (1011)",
			actual:    1011,
			expected:  1000,
			tolerance: 0.01,
			want:      false,
		},
		{
			name:      "両方ゼロ",
			actual:    0,
			expected:  0,
			tolerance: 0.01,
			want:      true,
		},
		{
			name:      "期待値ゼロ、実際値非ゼロ",
			actual:    100,
			expected:  0,
			tolerance: 0.01,
			want:      false,
		},
		{
			name:      "大きな値での1%以内",
			actual:    323992,
			expected:  324046,
			tolerance: 0.01,
			want:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SizesWithinTolerance(tc.actual, tc.expected, tc.tolerance)
			if got != tc.want {
				t.Errorf("SizesWithinTolerance(%d, %d, %f) = %v, want %v",
					tc.actual, tc.expected, tc.tolerance, got, tc.want)
			}
		})
	}
}

func TestSizesWithin1Percent(t *testing.T) {
	cases := []struct {
		name     string
		actual   int64
		expected int64
		want     bool
	}{
		{
			name:     "1%以内",
			actual:   1005,
			expected: 1000,
			want:     true,
		},
		{
			name:     "1%を超える",
			actual:   1015,
			expected: 1000,
			want:     false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SizesWithin1Percent(tc.actual, tc.expected)
			if got != tc.want {
				t.Errorf("SizesWithin1Percent(%d, %d) = %v, want %v",
					tc.actual, tc.expected, got, tc.want)
			}
		})
	}
}
