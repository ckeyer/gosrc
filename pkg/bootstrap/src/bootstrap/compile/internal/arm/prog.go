// Do not edit. Bootstrap copy of /usr/local/go/src/cmd/compile/internal/arm/prog.go

//line /usr/local/go/src/cmd/compile/internal/arm/prog.go:1
// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package arm

import (
	"bootstrap/compile/internal/gc"
	"bootstrap/internal/obj"
	"bootstrap/internal/obj/arm"
)

const (
	RightRdwr = gc.RightRead | gc.RightWrite
)

// This table gives the basic information about instruction
// generated by the compiler and processed in the optimizer.
// See opt.h for bit definitions.
//
// Instructions not generated need not be listed.
// As an exception to that rule, we typically write down all the
// size variants of an operation even if we just use a subset.
var progtable = [arm.ALAST & obj.AMask]obj.ProgInfo{
	obj.ATYPE:     {Flags: gc.Pseudo | gc.Skip},
	obj.ATEXT:     {Flags: gc.Pseudo},
	obj.AFUNCDATA: {Flags: gc.Pseudo},
	obj.APCDATA:   {Flags: gc.Pseudo},
	obj.AUNDEF:    {Flags: gc.Break},
	obj.AUSEFIELD: {Flags: gc.OK},
	obj.ACHECKNIL: {Flags: gc.LeftRead},
	obj.AVARDEF:   {Flags: gc.Pseudo | gc.RightWrite},
	obj.AVARKILL:  {Flags: gc.Pseudo | gc.RightWrite},
	obj.AVARLIVE:  {Flags: gc.Pseudo | gc.LeftRead},

	// NOP is an internal no-op that also stands
	// for USED and SET annotations, not the Intel opcode.
	obj.ANOP: {Flags: gc.LeftRead | gc.RightWrite},

	// Integer.
	arm.AADC & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.AADD & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.AAND & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.ABIC & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.ACMN & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RightRead},
	arm.ACMP & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RightRead},
	arm.ADIVU & obj.AMask:   {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.ADIV & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.AEOR & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.AMODU & obj.AMask:   {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.AMOD & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.AMULALU & obj.AMask: {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | RightRdwr},
	arm.AMULAL & obj.AMask:  {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | RightRdwr},
	arm.AMULA & obj.AMask:   {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | RightRdwr},
	arm.AMULU & obj.AMask:   {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.AMUL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.AMULL & obj.AMask:   {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.AMULLU & obj.AMask:  {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.AMVN & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite},
	arm.AORR & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.ARSB & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.ARSC & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.ASBC & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.ASLL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.ASRA & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.ASRL & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.ASUB & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RegRead | gc.RightWrite},
	arm.ATEQ & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RightRead},
	arm.ATST & obj.AMask:    {Flags: gc.SizeL | gc.LeftRead | gc.RightRead},

	// Floating point.
	arm.AADDD & obj.AMask:  {Flags: gc.SizeD | gc.LeftRead | RightRdwr},
	arm.AADDF & obj.AMask:  {Flags: gc.SizeF | gc.LeftRead | RightRdwr},
	arm.ACMPD & obj.AMask:  {Flags: gc.SizeD | gc.LeftRead | gc.RightRead},
	arm.ACMPF & obj.AMask:  {Flags: gc.SizeF | gc.LeftRead | gc.RightRead},
	arm.ADIVD & obj.AMask:  {Flags: gc.SizeD | gc.LeftRead | RightRdwr},
	arm.ADIVF & obj.AMask:  {Flags: gc.SizeF | gc.LeftRead | RightRdwr},
	arm.AMULD & obj.AMask:  {Flags: gc.SizeD | gc.LeftRead | RightRdwr},
	arm.AMULF & obj.AMask:  {Flags: gc.SizeF | gc.LeftRead | RightRdwr},
	arm.ASUBD & obj.AMask:  {Flags: gc.SizeD | gc.LeftRead | RightRdwr},
	arm.ASUBF & obj.AMask:  {Flags: gc.SizeF | gc.LeftRead | RightRdwr},
	arm.ASQRTD & obj.AMask: {Flags: gc.SizeD | gc.LeftRead | RightRdwr},

	// Conversions.
	arm.AMOVWD & obj.AMask: {Flags: gc.SizeD | gc.LeftRead | gc.RightWrite | gc.Conv},
	arm.AMOVWF & obj.AMask: {Flags: gc.SizeF | gc.LeftRead | gc.RightWrite | gc.Conv},
	arm.AMOVDF & obj.AMask: {Flags: gc.SizeF | gc.LeftRead | gc.RightWrite | gc.Conv},
	arm.AMOVDW & obj.AMask: {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite | gc.Conv},
	arm.AMOVFD & obj.AMask: {Flags: gc.SizeD | gc.LeftRead | gc.RightWrite | gc.Conv},
	arm.AMOVFW & obj.AMask: {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite | gc.Conv},

	// Moves.
	arm.AMOVB & obj.AMask: {Flags: gc.SizeB | gc.LeftRead | gc.RightWrite | gc.Move},
	arm.AMOVD & obj.AMask: {Flags: gc.SizeD | gc.LeftRead | gc.RightWrite | gc.Move},
	arm.AMOVF & obj.AMask: {Flags: gc.SizeF | gc.LeftRead | gc.RightWrite | gc.Move},
	arm.AMOVH & obj.AMask: {Flags: gc.SizeW | gc.LeftRead | gc.RightWrite | gc.Move},
	arm.AMOVW & obj.AMask: {Flags: gc.SizeL | gc.LeftRead | gc.RightWrite | gc.Move},

	// In addition, duffzero reads R0,R1 and writes R1.  This fact is
	// encoded in peep.c
	obj.ADUFFZERO: {Flags: gc.Call},

	// In addition, duffcopy reads R1,R2 and writes R0,R1,R2.  This fact is
	// encoded in peep.c
	obj.ADUFFCOPY: {Flags: gc.Call},

	// These should be split into the two different conversions instead
	// of overloading the one.
	arm.AMOVBS & obj.AMask: {Flags: gc.SizeB | gc.LeftRead | gc.RightWrite | gc.Conv},
	arm.AMOVBU & obj.AMask: {Flags: gc.SizeB | gc.LeftRead | gc.RightWrite | gc.Conv},
	arm.AMOVHS & obj.AMask: {Flags: gc.SizeW | gc.LeftRead | gc.RightWrite | gc.Conv},
	arm.AMOVHU & obj.AMask: {Flags: gc.SizeW | gc.LeftRead | gc.RightWrite | gc.Conv},

	// Jumps.
	arm.AB & obj.AMask:   {Flags: gc.Jump | gc.Break},
	arm.ABL & obj.AMask:  {Flags: gc.Call},
	arm.ABEQ & obj.AMask: {Flags: gc.Cjmp},
	arm.ABNE & obj.AMask: {Flags: gc.Cjmp},
	arm.ABCS & obj.AMask: {Flags: gc.Cjmp},
	arm.ABHS & obj.AMask: {Flags: gc.Cjmp},
	arm.ABCC & obj.AMask: {Flags: gc.Cjmp},
	arm.ABLO & obj.AMask: {Flags: gc.Cjmp},
	arm.ABMI & obj.AMask: {Flags: gc.Cjmp},
	arm.ABPL & obj.AMask: {Flags: gc.Cjmp},
	arm.ABVS & obj.AMask: {Flags: gc.Cjmp},
	arm.ABVC & obj.AMask: {Flags: gc.Cjmp},
	arm.ABHI & obj.AMask: {Flags: gc.Cjmp},
	arm.ABLS & obj.AMask: {Flags: gc.Cjmp},
	arm.ABGE & obj.AMask: {Flags: gc.Cjmp},
	arm.ABLT & obj.AMask: {Flags: gc.Cjmp},
	arm.ABGT & obj.AMask: {Flags: gc.Cjmp},
	arm.ABLE & obj.AMask: {Flags: gc.Cjmp},
	obj.ARET:             {Flags: gc.Break},
}

func proginfo(p *obj.Prog) {
	info := &p.Info
	*info = progtable[p.As&obj.AMask]
	if info.Flags == 0 {
		gc.Fatalf("unknown instruction %v", p)
	}

	if p.From.Type == obj.TYPE_ADDR && p.From.Sym != nil && (info.Flags&gc.LeftRead != 0) {
		info.Flags &^= gc.LeftRead
		info.Flags |= gc.LeftAddr
	}

	if (info.Flags&gc.RegRead != 0) && p.Reg == 0 {
		info.Flags &^= gc.RegRead
		info.Flags |= gc.CanRegRead | gc.RightRead
	}

	if (p.Scond&arm.C_SCOND != arm.C_SCOND_NONE) && (info.Flags&gc.RightWrite != 0) {
		info.Flags |= gc.RightRead
	}

	switch p.As {
	case arm.ADIV,
		arm.ADIVU,
		arm.AMOD,
		arm.AMODU:
		info.Regset |= RtoB(arm.REG_R12)
	}
}
