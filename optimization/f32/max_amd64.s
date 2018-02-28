// Copyright ©2017 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build !noasm,!appengine

#include "textflag.h"

// func Max(dst, s []float32)
TEXT ·Max(SB), NOSPLIT, $0
	MOVQ    dst_base+0(FP), DI // DI = &dst
	MOVQ    dst_len+8(FP), CX  // CX = len(dst)
	MOVQ    s_base+24(FP), SI  // SI = &s
	CMPQ    s_len+32(FP), CX   // CX = max( CX, len(s) )
	CMOVQLE s_len+32(FP), CX
	CMPQ    CX, $0
	JE      max_end

	XORQ AX, AX
	MOVQ DI, DX
	ANDQ $0xF, DX    // Align on 16-byte boundary for MAXPS
	JZ   max_no_trim // if DX == 0 { goto max_no_trim }
	SUBQ $16, DX

max_align: // Trim first value(s) in unaligned buffer  do {
	MOVSS (SI)(AX*4), X0 // X0 = s[i]
	MAXSS (DI)(AX*4), X0 // X0 = max(dst[i], X0)
	MOVSS X0, (DI)(AX*4) // dst[i] = X0
	INCQ  AX                // AX++
	DECQ  CX
	JZ    max_end            // if --BX == 0 { return }
	ADDQ  $4, DX
	JNZ   max_align          // } while --DX > 0

max_no_trim:
	MOVQ CX, BX
	ANDQ $0xF, BX      // BX = CX % 16
	SHRQ $4, CX         // CX = floor( LEN / 16 )
	JZ   max_tail4_start // if CX == 0 { goto max_tail4_start }

max_loop: // Loop unrolled 16x  do {
	MOVUPS (SI)(AX*4), X0   // X_i = x[i:i+1]
	MOVUPS 16(SI)(AX*4), X1
	MOVUPS 32(SI)(AX*4), X2
	MOVUPS 48(SI)(AX*4), X3

	MAXPS (DI)(AX*4), X0   // X_i *= y[i:i+1]
	MAXPS 16(DI)(AX*4), X1
	MAXPS 32(DI)(AX*4), X2
	MAXPS 48(DI)(AX*4), X3

	MOVUPS X0, (DI)(AX*4)   // dst[i:i+1] = X_i
	MOVUPS X1, 16(DI)(AX*4)
	MOVUPS X2, 32(DI)(AX*4)
	MOVUPS X3, 48(DI)(AX*4)
	
	ADDQ $16, AX // AX += 16
	DECQ CX
	JNZ  max_loop // } while --CX > 0

	CMPQ  BX, $0   // if BX == 0 { return }
	JE    max_end

max_tail4_start: // Reset loop counter for 4-wide tail loop
	MOVQ BX, CX      // CX = floor( BX / 4 )
	SHRQ $2, CX
	JZ   max_tail_start // if CX == 0 { goto max_tail_start }

max_tail4_loop: // Loop unrolled 4x  do {
	MOVUPS (SI)(AX*4), X0 // X_i = x[i:i+1]
	MAXPS  (DI)(AX*4), X0 // X_i *= y[i:i+1]
	MOVUPS X0, (DI)(AX*4)   // dst[i:i+1] = X_i
	ADDQ   $4, AX            // i += 4
	DECQ   CX
	JNZ    max_tail4_loop     // } while --CX > 0

max_tail_start: // Reset loop counter for 1-wide tail loop
	ANDQ $3, BX // BX = BX % 4
	JZ   max_end  // if BX == 0 { return }

max_tail: // do {
	MOVSS (SI)(AX*4), X0 // X0 = x[i]
	MAXSS (DI)(AX*4), X0 // X_i *= y[i:i+1]
	MOVSS X0, (DI)(AX*4)   // dst[i:i+1] = X_i
	INCQ  AX                // AX++
	DECQ  BX
	JNZ   max_tail           // } while --BX > 0

max_end:
	RET
