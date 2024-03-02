// Copyright 2024 The Blackout Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/nfnt/resize"
)

// Cache is a binomial cache entry
type Cache struct {
	Value float64
	Valid bool
}

// BinomialCoefficient is the binomial coeffcient
func BinomialCoefficient(cache *[256][256]Cache, n, k uint) float64 {
	if k > n {
		return 0
	} else if k == 0 || k == n {
		return 1
	}
	x := 0.0
	if cache[n-1][k-1].Valid {
		x = cache[int(n-1)][int(k-1)].Value
	} else {
		x = BinomialCoefficient(cache, n-1, k-1)
		cache[n-1][k-1].Value = x
		cache[n-1][k-1].Valid = true

	}
	y := 0.0
	if cache[n-1][k].Valid {
		y = cache[int(n-1)][int(k)].Value
	} else {
		y = BinomialCoefficient(cache, n-1, k)
		cache[n-1][k].Value = y
		cache[n-1][k].Valid = true
	}
	return x + y
}

// Gray computes the gray scale version of an image
func Gray(input image.Image) *image.Gray16 {
	bounds := input.Bounds()
	output := image.NewGray16(bounds)
	width, height := bounds.Max.X, bounds.Max.Y
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			r, g, b, _ := input.At(x, y).RGBA()
			output.SetGray16(x, y, color.Gray16{uint16((float64(r)+float64(g)+float64(b))/3 + .5)})
		}
	}
	return output
}

func main() {
	binomial := [256][256]Cache{}
	for n := uint(0); n < 256; n++ {
		for k := uint(0); k <= n; k++ {
			binomial[n][k].Value = BinomialCoefficient(&binomial, n, k)
			binomial[n][k].Valid = true
		}
	}

	file, err := os.Open("test.jpg")
	if err != nil {
		log.Fatal(err)
	}

	info, err := file.Stat()
	if err != nil {
		log.Fatal(err)
	}
	name := info.Name()
	name = name[:strings.Index(name, ".")]

	input, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()

	input = resize.Resize(32, 32, input, resize.NearestNeighbor)

	file, err = os.Create(name + ".png")
	if err != nil {
		log.Fatal(err)
	}

	err = png.Encode(file, input)
	if err != nil {
		log.Fatal(err)
	}
	file.Close()
}
