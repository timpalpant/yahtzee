// Copyright ©2017 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build !noasm,!appengine

#include "textflag.h"

// func Min(dst, s []float32)
TEXT ·Min(SB), NOSPLIT, $0
	MOVQ    dst_base+0(FP), DI // DI = &dst
	MOVQ    dst_len+8(FP), CX  // CX = len(dst)
	MOVQ    s_base+24(FP), SI  // SI = &s
	CMPQ    s_len+32(FP), CX   // CX = max( CX, len(s) )
	CMOVQLE s_len+32(FP), CX
	CMPQ    CX, $0
	JE      min_end

	XORQ AX, AX
	MOVQ DI, DX
	ANDQ $0xF, DX    // Align on 16-byte boundary for MINPS
	JZ   min_no_trim // if DX == 0 { goto min_no_trim }
	SUBQ $16, DX

min_align: // Trim first value(s) in unaligned buffer  do {
	MOVSS (SI)(AX*4), X0 // X0 = s[i]
	MINSS (DI)(AX*4), X0 // X0 = max(dst[i], X0)
	MOVSS X0, (DI)(AX*4) // dst[i] = X0
	INCQ  AX                // AX++
	DECQ  CX
	JZ    min_end            // if --BX == 0 { return }
	ADDQ  $4, DX
	JNZ   min_align          // } while --DX > 0

min_no_trim:
	MOVQ CX, BX
	ANDQ $0xF, BX      // BX = CX % 16
	SHRQ $4, CX         // CX = floor( LEN / 16 )
	JZ   min_tail4_start // if CX == 0 { goto min_tail4_start }

min_loop: // Loop unrolled 16x  do {
	MOVUPS (SI)(AX*4), X0   // X_i = x[i:i+1]
	MOVUPS 16(SI)(AX*4), X1
	MOVUPS 32(SI)(AX*4), X2
	MOVUPS 48(SI)(AX*4), X3

	MINPS (DI)(AX*4), X0   // X_i *= y[i:i+1]
	MINPS 16(DI)(AX*4), X1
	MINPS 32(DI)(AX*4), X2
	MINPS 48(DI)(AX*4), X3

	MOVUPS X0, (DI)(AX*4)   // dst[i:i+1] = X_i
	MOVUPS X1, 16(DI)(AX*4)
	MOVUPS X2, 32(DI)(AX*4)
	MOVUPS X3, 48(DI)(AX*4)
	
	ADDQ $16, AX // AX += 16
	DECQ CX
	JNZ  min_loop // } while --CX > 0

	CMPQ  BX, $0   // if BX == 0 { return }
	JE    min_end

min_tail4_start: // Reset loop counter for 4-wide tail loop
	MOVQ BX, CX      // CX = floor( BX / 4 )
	SHRQ $2, CX
	JZ   min_tail_start // if CX == 0 { goto min_tail_start }

min_tail4_loop: // Loop unrolled 4x  do {
	MOVUPS (SI)(AX*4), X0 // X_i = x[i:i+1]
	MINPS  (DI)(AX*4), X0 // X_i *= y[i:i+1]
	MOVUPS X0, (DI)(AX*4)   // dst[i:i+1] = X_i
	ADDQ   $4, AX            // i += 4
	DECQ   CX
	JNZ    min_tail4_loop     // } while --CX > 0

min_tail_start: // Reset loop counter for 1-wide tail loop
	ANDQ $3, BX // BX = BX % 4
	JZ   min_end  // if BX == 0 { return }

min_tail: // do {
	MOVSS (SI)(AX*4), X0 // X0 = x[i]
	MINSS (DI)(AX*4), X0 // X_i *= y[i:i+1]
	MOVSS X0, (DI)(AX*4)   // dst[i:i+1] = X_i
	INCQ  AX                // AX++
	DECQ  BX
	JNZ   min_tail           // } while --BX > 0

min_end:
	RET
