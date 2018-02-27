package f64

import "testing"

func TestMax(t *testing.T) {
	cases := []struct {
		a        []float64
		b        []float64
		expected []float64
	}{
		{
			a:        []float64{1},
			b:        []float64{2},
			expected: []float64{2},
		},
		{
			a:        []float64{1, 4},
			b:        []float64{2, 3},
			expected: []float64{2, 4},
		},
		{
			a:        []float64{1, 2, 3, 4},
			b:        []float64{2, 2, 1, 5},
			expected: []float64{2, 2, 3, 5},
		},
		{
			a:        []float64{1, 2, 3, 4, 5},
			b:        []float64{2, 2, 1, 5, 1},
			expected: []float64{2, 2, 3, 5, 5},
		},
		{
			a:        []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			b:        []float64{2, 2, 1, 5, 1, 1, 2, 3, 4, 15, 16},
			expected: []float64{2, 2, 3, 5, 5, 6, 7, 8, 9, 15, 16},
		},
		{
			a:        []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
			b:        []float64{2, 2, 1, 5, 1, 1, 2, 3, 4, 15, 16, 17, 16, 15, 14, 13},
			expected: []float64{2, 2, 3, 5, 5, 6, 7, 8, 9, 15, 16, 17, 16, 15, 15, 16},
		},
		{
			a:        []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17},
			b:        []float64{2, 2, 1, 5, 1, 1, 2, 3, 4, 15, 16, 17, 16, 15, 14, 13, 12},
			expected: []float64{2, 2, 3, 5, 5, 6, 7, 8, 9, 15, 16, 17, 16, 15, 15, 16, 17},
		},
	}

	for _, tc := range cases {
		Max(tc.a, tc.b)
		for i, x := range tc.expected {
			if x != tc.a[i] {
				t.Errorf("result[%v]: got %v, expected %v",
					i, tc.a[i], x)
			}
		}
	}
}

func BenchmarkMax(b *testing.B) {
	x := make([]float64, 10000)
	for i := 0; i < len(x); i++ {
		x[i] = float64(i)
	}

	y := make([]float64, 10000)
	for i := 0; i < len(y); i++ {
		y[i] = float64(b.N - i)
	}

	for i := 0; i < b.N; i++ {
		Max(x, y)
	}
}
