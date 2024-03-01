// Copyright 2024 The Blackout Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"image"
	_ "image/jpeg"
	"image/png"
	"log"
	"os"
	"strings"

	"github.com/nfnt/resize"
)

func main() {
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
