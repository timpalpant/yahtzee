//+build !noasm,!appengine

package f32

// AxpyUnitaryTo is
//  for i, v := range x {
//  	dst[i] = alpha*v + y[i]
//  }
func AxpyUnitaryTo(dst []float32, alpha float32, x, y []float32)

// Max is
//  for i, v := range s {
//      if v > dst[i] {
//          dst[i] = v
//      }
//  }
func Max(dst, s []float32)

// Min is
//  for i, v := range s {
//      if v < dst[i] {
//          dst[i] = v
//      }
//  }
func Min(dst, s []float32)
