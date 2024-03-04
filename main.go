// Copyright 2024 The Blackout Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"log"
	"math"
	"math/rand"
	"os"
	"strings"

	. "github.com/pointlander/matrix"

	"github.com/nfnt/resize"
)

// Cache is a binomial cache entry
type Cache struct {
	Value float64
	Valid bool
}

// ByteCache is a byte binomial cache
type ByteCache [256][256]Cache

// BinomialCoefficient is the binomial coeffcient
func (b *ByteCache) BinomialCoefficient(n, k uint) float64 {
	if k > n {
		return 0
	} else if k == 0 || k == n {
		return 1
	}
	x := 0.0
	if b[n-1][k-1].Valid {
		x = b[int(n-1)][int(k-1)].Value
	} else {
		x = b.BinomialCoefficient(n-1, k-1)
		b[n-1][k-1].Value = x
		b[n-1][k-1].Valid = true

	}
	y := 0.0
	if b[n-1][k].Valid {
		y = b[int(n-1)][int(k)].Value
	} else {
		y = b.BinomialCoefficient(n-1, k)
		b[n-1][k].Value = y
		b[n-1][k].Valid = true
	}
	return x + y
}

// Probability computes the probability
func Probability(t1, t2 float64) float64 {
	return (math.Exp(-t1) - math.Exp(-t2)) / (1 - math.Exp(-t2))
}

// Sample samples the binomial distribution
func (b *ByteCache) Sample(rng *rand.Rand, n uint, l1, l2 float64) uint {
	sum := 0.0
	sample := rng.Float64()
	p := Probability(l1, l2)
	for k := uint(0); k <= n; k++ {
		f := b[n][k].Value * math.Pow(p, float64(k)) * math.Pow(1-p, float64(n-k))
		sum += f
		if sum > sample {
			return k
		}
	}
	return 0
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
	binomial := ByteCache{}
	for n := uint(0); n < 256; n++ {
		for k := uint(0); k <= n; k++ {
			binomial[n][k].Value = binomial.BinomialCoefficient(n, k)
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

	rng := rand.New(rand.NewSource(1))
	for i := 0; i < 16; i++ {
		s := binomial.Sample(rng, 128, .1, .5)
		fmt.Println(s)
	}

	forms := make([]Matrix, 10)
	for i := range forms {
		forms[i] = NewMatrix(32*32, 32*32)
	}
}
