// Copyright 2024 The Blackout Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"testing"
)

func TestBinomialCoefficient(t *testing.T) {
	binomial := [256][256]Cache{}
	if a := BinomialCoefficient(&binomial, 5, 2); a != 10 {
		t.Fatal("(5,2) != 10 is ", a)
	}
}
