// Copyright 2024 The Blackout Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"testing"
)

func TestBinomialCoefficient(t *testing.T) {
	binomial := [256][256]Cache{}
	type Test struct {
		n, k uint
		out  float64
	}
	tests := [...]Test{
		{5, 2, 10},
		{10, 5, 252},
		{20, 13, 77520},
		{25, 15, 3268760},
		{27, 17, 8436285},
	}
	for _, test := range tests {
		if a := BinomialCoefficient(&binomial, test.n, test.k); a != test.out {
			t.Fatalf("(%d,%d) != %f is %f", test.n, test.k, test.out, a)
		}
	}
}
