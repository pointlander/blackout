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
	"runtime"
	"strings"

	. "github.com/pointlander/matrix"

	"github.com/nfnt/resize"
)

const (
	// Size is the size of the image
	Size = 16
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

	input = resize.Resize(Size, Size, input, resize.NearestNeighbor)

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
		forms[i] = NewMatrix(Size*Size+1, Size*Size)
	}
	target := NewZeroMatrix(Size*Size+1, 1)
	target.Data[Size*Size] = 1
	index := 0
	bounds := input.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			r, g, b, _ := input.At(x, y).RGBA()
			target.Data[index] = float32(math.Round((float64(r)+float64(g)+float64(b))/(256.0*3) + .5))
			index++
		}
	}
	coords := NewCoord(Size*Size+1, Size*Size)
	optimizer := NewOptimizer(rng, 8, 1, 2, func(samples []Sample, x ...Matrix) {
		done := make(chan bool, 8)
		process := func(seed int64, index int) {
			rng := rand.New(rand.NewSource(seed))
			X := NewZeroMatrix(Size*Size+1, 1)
			X.Data[Size*Size] = 1
			for i := 0; i < 10; i++ {
				x := samples[index].Vars[0][0]
				y := samples[index].Vars[0][1]
				z := samples[index].Vars[0][2]
				Y := Sigmoid(MulT(Add(x, H(y, z)), X))
				xx := samples[index].Vars[1][0]
				yy := samples[index].Vars[1][1]
				zz := samples[index].Vars[1][2]
				hidden := NewZeroMatrix(Size*Size+1, 1)
				copy(hidden.Data, Y.Data)
				hidden.Data[Size*Size] = 1
				Y = MulT(Add(xx, H(yy, zz)), hidden)
				for j := 0; j < Y.Size(); j++ {
					pixel := Y.Data[j] * 255
					if pixel < 0 {
						pixel = 0
					} else if max := 255 - X.Data[j]; pixel > max {
						pixel = max
					}
					d := binomial.Sample(rng, uint(pixel+.5), float64(i)*.1, float64(i+1)*.1)
					X.Data[j] += float32(d)
				}
			}
			cost := Quadratic(target, X)
			samples[index].Cost = float64(cost.Data[0])
			done <- true
		}
		cpus := runtime.NumCPU()
		j, flight := 0, 0
		for j < len(samples) && flight < cpus {
			go process(rng.Int63(), j)
			j++
			flight++
		}
		for j < len(samples) {
			<-done
			flight--

			go process(rng.Int63(), j)
			j++
			flight++
		}
		for j := 0; j < flight; j++ {
			<-done
		}
	}, coords)
	s := optimizer.Optimize(1e-6)
	x := Add(s.Vars[0][0], H(s.Vars[0][1], s.Vars[0][2]))
	_ = x
}
