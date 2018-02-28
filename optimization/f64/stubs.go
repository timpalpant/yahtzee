//+build !noasm,!appengine

package f64

// Max is
//  for i, v := range s {
//      if v > dst[i] {
//          dst[i] = v
//      }
//  }
func Max(dst, s []float64)

// Min is
//  for i, v := range s {
//      if v < dst[i] {
//          dst[i] = v
//      }
//  }
func Min(dst, s []float64)
