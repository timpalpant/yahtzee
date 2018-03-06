package f32

import "fmt"

// AddScaled performs dst = dst + alpha * s.
// It panics if the lengths of dst and s are not equal.
func AddScaled(dst []float32, alpha float32, s []float32) {
	if len(dst) != len(s) {
		panic(fmt.Errorf("length of destination and source to not match: %v != %v", len(dst), len(s)))
	}

	AxpyUnitaryTo(dst, alpha, s, dst)
}
