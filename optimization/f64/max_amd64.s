// Copyright ©2016 The Gonum Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build !noasm,!appengine

#include "textflag.h"

// func Max(dst, s []float64)
TEXT ·Max(SB), NOSPLIT, $0
	MOVQ    dst_base+0(FP), DI // DI = &dst
	MOVQ    dst_len+8(FP), CX  // CX = len(dst)
	MOVQ    s_base+24(FP), SI  // SI = &s
	CMPQ    s_len+32(FP), CX   // CX = max( CX, len(s) )
	CMOVQLE s_len+32(FP), CX
	CMPQ    CX, $0             // if CX == 0 { return }
	JE      max_end
	XORQ    AX, AX
	MOVQ    DI, BX
	ANDQ    $0x0F, BX          // BX = &dst & 15
	JZ      max_no_trim        // if BX == 0 { goto max_no_trim }

	// Align on 16-bit boundary
	MOVSD (SI)(AX*8), X0 // X0 = s[i]
	MAXSD (DI)(AX*8), X0 // X0 += dst[i]
	MOVSD X0, (DI)(AX*8) // dst[i] = X0
	INCQ  AX             // i++
	DECQ  CX             // --CX
	JE    max_end        // if CX == 0 { return  }

max_no_trim:
	MOVQ CX, BX
	ANDQ $7, BX         // BX = len(dst) % 8
	SHRQ $3, CX         // CX = floor( len(dst) / 8 )
	JZ   max_tail_start // if CX == 0 { goto max_tail_start }

max_loop: // Loop unrolled 8x   do {
	MOVUPS (SI)(AX*8), X0   // X_i = s[i:i+1]
	MOVUPS 16(SI)(AX*8), X1
	MOVUPS 32(SI)(AX*8), X2
	MOVUPS 48(SI)(AX*8), X3
	MAXPD  (DI)(AX*8), X0   // X_i += dst[i:i+1]
	MAXPD  16(DI)(AX*8), X1
	MAXPD  32(DI)(AX*8), X2
	MAXPD  48(DI)(AX*8), X3
	MOVUPS X0, (DI)(AX*8)   // dst[i:i+1] = X_i
	MOVUPS X1, 16(DI)(AX*8)
	MOVUPS X2, 32(DI)(AX*8)
	MOVUPS X3, 48(DI)(AX*8)
	ADDQ   $8, AX           // i += 8
	LOOP   max_loop         // } while --CX > 0
	CMPQ   BX, $0           // if BX == 0 { return }
	JE     max_end

max_tail_start: // Reset loop registers
	MOVQ BX, CX // Loop counter: CX = BX

max_tail: // do {
	MOVSD (SI)(AX*8), X0 // X0 = s[i]
	MAXSD (DI)(AX*8), X0 // X0 += dst[i]
	MOVSD X0, (DI)(AX*8) // dst[i] = X0
	INCQ  AX             // ++i
	LOOP  max_tail       // } while --CX > 0

max_end:
	RET
