// Do not edit. Bootstrap copy of /usr/local/go/src/cmd/compile/internal/ssa/opt.go

//line /usr/local/go/src/cmd/compile/internal/ssa/opt.go:1
// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ssa

// machine-independent optimization
func opt(f *Func) {
	applyRewrite(f, rewriteBlockgeneric, rewriteValuegeneric)
}

func dec(f *Func) {
	applyRewrite(f, rewriteBlockdec, rewriteValuedec)
}
