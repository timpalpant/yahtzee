// Copyright ©2016 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build !amd64 noasm appengine

package f32

// AxpyUnitaryTo is
//  for i, v := range x {
//  	dst[i] = alpha*v + y[i]
//  }
func AxpyUnitaryTo(dst []float32, alpha float32, x, y []float32) {
	for i, v := range x {
		dst[i] = alpha*v + y[i]
	}
}

// Max is
//  for i, v := range s {
//      if v > dst[i] {
//          dst[i] = v
//      }
//  }
func Max(dst, s []float32) {
	for i, x := range s {
		if x > dst[i] {
			dst[i] = x
		}
	}
}

// Min is
//  for i, v := range s {
//      if v < dst[i] {
//          dst[i] = v
//      }
//  }
func Min(dst, s []float32) {
	for i, x := range s {
		if x < dst[i] {
			dst[i] = x
		}
	}
}
