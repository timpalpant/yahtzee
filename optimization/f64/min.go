// Copyright ©2015 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build !amd64 noasm appengine

package f64

// Min is
//  for i, v := range s {
//      if v < dst[i] {
//          dst[i] = v
//      }
//  }
func Min(dst, s []float64) {
	for i, x := range s {
		if x < dst[i] {
			dst[i] = x
		}
	}
}
