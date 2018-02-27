package f32

// AddScaled performs dst = dst + alpha * s.
// It panics if the lengths of dst and s are not equal.
func AddScaled(dst []float32, alpha float32, s []float32) {
	if len(dst) != len(s) {
		panic("floats: length of destination and source to not match")
	}

	AxpyUnitaryTo(dst, alpha, s, dst)
}
