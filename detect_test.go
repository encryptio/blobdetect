// Copyright (c) 2014, Jack Christopher Kastorff <encryptio@gmail.com>
// Use of this source code is governed by an ISC-style license that
// can be found in the LICENSE.txt file.

package blobdetect

import (
	"image"
	_ "image/png"
	"math"
	"os"
	"testing"
)

func floatEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.000001
}

func blobListEqual(as, bs []Blob) bool {
	if len(as) != len(bs) {
		return false
	}

	for i, a := range as {
		b := bs[i]
		if !(floatEqual(a.X, b.X) && floatEqual(a.Y, b.Y) && floatEqual(a.Xdev, b.Xdev) && floatEqual(a.Ydev, b.Ydev) && floatEqual(a.Weight, b.Weight)) {
			return false
		}
	}

	return true
}

func TestDetectBasic(t *testing.T) {
	third := math.Sqrt((1.0 / 3) - (1.0 / 9))

	tests := []struct {
		img   image.Image
		blobs []Blob
		name  string
	}{
		{&image.Gray{[]uint8{0, 0, 0, 0}, 2, image.Rect(0, 0, 2, 2)}, nil, "empty 2x2"},
		{&image.Gray{[]uint8{1, 0, 0, 0}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{0, 0, 0, 0, 1}}, "upper left 2x2"},
		{&image.Gray{[]uint8{0, 1, 0, 0}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{1, 0, 0, 0, 1}}, "upper right 2x2"},
		{&image.Gray{[]uint8{1, 1, 0, 0}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{0.5, 0, 0.5, 0, 2}}, "upper row 2x2"},
		{&image.Gray{[]uint8{0, 0, 1, 0}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{0, 1, 0, 0, 1}}, "lower left 2x2"},
		{&image.Gray{[]uint8{1, 0, 1, 0}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{0, 0.5, 0, 0.5, 2}}, "left column 2x2"},
		{&image.Gray{[]uint8{0, 1, 1, 0}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{1, 0, 0, 0, 1}, Blob{0, 1, 0, 0, 1}}, "diagonal up 2x2"},
		{&image.Gray{[]uint8{1, 1, 1, 0}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{1.0 / 3, 1.0 / 3, third, third, 3}}, "all but lower right 2x2"},
		{&image.Gray{[]uint8{0, 0, 0, 1}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{1, 1, 0, 0, 1}}, "lower right 2x2"},
		{&image.Gray{[]uint8{1, 0, 0, 1}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{0, 0, 0, 0, 1}, Blob{1, 1, 0, 0, 1}}, "diagonal down 2x2"},
		{&image.Gray{[]uint8{0, 1, 0, 1}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{1, 0.5, 0, 0.5, 2}}, "right column 2x2"},
		{&image.Gray{[]uint8{1, 1, 0, 1}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{2.0 / 3, 1.0 / 3, third, third, 3}}, "all but lower left 2x2"},
		{&image.Gray{[]uint8{0, 0, 1, 1}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{0.5, 1, 0.5, 0, 2}}, "bottom row 2x2"},
		{&image.Gray{[]uint8{1, 0, 1, 1}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{1.0 / 3, 2.0 / 3, third, third, 3}}, "all but upper right 2x2"},
		{&image.Gray{[]uint8{0, 1, 1, 1}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{2.0 / 3, 2.0 / 3, third, third, 3}}, "all but upper left 2x2"},
		{&image.Gray{[]uint8{1, 1, 1, 1}, 2, image.Rect(0, 0, 2, 2)}, []Blob{Blob{0.5, 0.5, 0.5, 0.5, 4}}, "full 2x2"},
		{&image.Gray{[]uint8{1, 2}, 2, image.Rect(0, 0, 2, 1)}, []Blob{Blob{2.0 / 3, 0, third, 0, 3}}, "weighted 2x1"},
	}

	for _, test := range tests {
		out := Detect(test.img, nil)

		if !blobListEqual(out, test.blobs) {
			t.Errorf("%s: Detect(%v) = %v, wanted %v", test.name, test.img, out, test.blobs)
		}
	}
}

func TestDetectReuse(t *testing.T) {
	img := &image.Gray{[]uint8{1, 0, 0, 1}, 2, image.Rect(0, 0, 2, 2)}
	expect := []Blob{Blob{0, 0, 0, 0, 1}, Blob{1, 1, 0, 0, 1}}
	have := []Blob{Blob{9, 9, 9, 9, 9}, Blob{4, 4, 4, 4, 4}}

	have = Detect(img, have)

	if !blobListEqual(have, expect) {
		t.Errorf("Detect(%v) = %v, wanted %v", img, have, expect)
	}
}

func benchFile(filename string, b *testing.B) {
	r, err := os.Open(filename)
	if err != nil {
		b.Fatal(err)
	}

	img, _, err := image.Decode(r)
	if err != nil {
		b.Fatal(err)
	}

	gray := convertToGray(img)

	b.ResetTimer()
	var out []Blob
	for i := 0; i < b.N; i++ {
		out = Detect(gray, out)
	}
}

func BenchmarkRandomCrap(b *testing.B) {
	benchFile("randomcrap.png", b)
}

func BenchmarkChunky(b *testing.B) {
	benchFile("chunky.png", b)
}
