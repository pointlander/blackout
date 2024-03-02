// Copyright 2024 The Blackout Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"testing"
)

func TestBinomialCoefficient(t *testing.T) {
	if a := BinomialCoefficient(5, 2); a != 10 {
		t.Fatal("(5,2) != 10 is ", a)
	}
}
