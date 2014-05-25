// Copyright (c) 2014, Jack Christopher Kastorff <encryptio@gmail.com>
// Use of this source code is governed by an ISC-style license that
// can be found in the LICENSE.txt file.

package main

import (
	"git.encryptio.com/blobdetect"
	"image"
	_ "image/png"
	"os"
	"fmt"
	"math/rand"
	"log"
)

func randomSVGColor() string {
	r := rand.Float64()
	g := rand.Float64()
	b := rand.Float64()

	min := r
	if min > g { min = g }
	if min > b { min = b }

	r -= min
	g -= min
	b -= min

	max := r
	if max < g { max = g }
	if max < b { max = b }

	r /= max
	g /= max
	b /= max

	ri := byte(r*255)
	gi := byte(g*255)
	bi := byte(b*255)

	return fmt.Sprintf("#%x", []byte{ri,gi,bi})
}

func main() {
	filename := os.Args[1]

	r, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.Decode(r)
	if err != nil {
		log.Fatal(err)
	}

	blobs := blobdetect.Detect(img, nil)

	fmt.Printf(`<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">`)
	fmt.Printf("\n")
	fmt.Printf("<svg width=\"%v\" height=\"%v\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\">\n", img.Bounds().Dx(), img.Bounds().Dy())
	fmt.Printf("<image x=\"0\" y=\"0\" width=\"%v\" height=\"%v\" xlink:href=\"%s\" />\n", img.Bounds().Dx(), img.Bounds().Dy(), filename)
	for _, blob := range blobs {
		fmt.Printf("<ellipse cx=\"%v\" cy=\"%v\" rx=\"%v\" ry=\"%v\" style=\"stroke:%v;stroke-width:2;fill:none\" />\n", blob.X+0.5, blob.Y+0.5, blob.Xdev, blob.Ydev, randomSVGColor())
	}
	fmt.Printf("</svg>\n")
}
