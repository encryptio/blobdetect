// Copyright (c) 2014, Jack Christopher Kastorff <encryptio@gmail.com>
// Use of this source code is governed by an ISC-style license that
// can be found in the LICENSE.txt file.

package blobdetect

import (
	"image"
	"image/draw"
	"math"
)

type Blob struct {
	// average location
	X, Y float64

	// standard deviation
	Xdev, Ydev float64

	// sum of pixel values
	Weight float64
}

func convertToGray(i image.Image) *image.Gray {
	g, ok := i.(*image.Gray)
	if ok {
		return g
	}

	g = image.NewGray(i.Bounds())
	draw.Draw(g, i.Bounds(), i, image.Point{}, draw.Over)
	return g
}

// Detect calculates the list of connected blobs in the image, defined as
// 4-connected blobs of non-zero value in an image.Gray. (The image given
// is converted to an image.Gray if it is not one already.)
//
// If given its second argument, it will reuse the memory in the slice for
// its return value (if possible) and thus, in the common case, cause zero
// allocations.
func Detect(source image.Image, out []Blob) []Blob {
	gray := convertToGray(source)
	bounds := gray.Bounds()

	// out will contain blobs with:
	// X, Y: running weighted sum of X and Y locations
	// Xdev, Ydev: running weighted sum of squared X and Y locations
	// Weight: running sum of weights
	if len(out) > 0 {
		out = out[0:0]
	}

	// At any given time in the x loop, the indicies[:x-1] contains the blob
	// indicies of the pixels for the current row, and indicies[x:] contains
	// the blob indicies of the pixels for the previous row (if any). If there
	// is no blob associated with that pixel, the value is -1.
	indicies := getIntArray(bounds.Dx())

	for i := range indicies {
		indicies[i] = -1
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		rowData := gray.Pix[gray.PixOffset(bounds.Min.X, y):gray.PixOffset(bounds.Max.X, y)]

		for x, v := range rowData {
			if v == 0 {
				// not in any blob
				indicies[x] = -1
				continue
			}

			upBlob := indicies[x]
			var leftBlob int
			if x > 0 {
				leftBlob = indicies[x-1]
			} else {
				leftBlob = -1
			}

			vf := float64(v)
			xf := float64(x)
			yf := float64(y)
			xf2 := xf * xf
			yf2 := yf * yf

			// 5 cases:
			//     .
			//   . x -> create new blob
			//
			//     a     .     a
			//   . x   a x   a x -> add to a
			//
			//     a
			//   b x -> merge a and b, then add

			if upBlob == -1 && leftBlob == -1 {
				// new blob
				out = append(out, Blob{xf * vf, yf * vf, xf2 * vf, yf2 * vf, vf})
				indicies[x] = len(out) - 1
				continue
			}

			addTo := -1

			if upBlob != -1 && leftBlob == -1 {
				addTo = upBlob
			} else if upBlob == -1 && leftBlob != -1 {
				addTo = leftBlob
			} else {
				// upBlob != -1 && leftBlob != -1
				if upBlob == leftBlob {
					addTo = upBlob
				} else {
					// merge upBlob and leftBlob

					out[upBlob].X += out[leftBlob].X
					out[upBlob].Y += out[leftBlob].Y
					out[upBlob].Xdev += out[leftBlob].Xdev
					out[upBlob].Ydev += out[leftBlob].Ydev
					out[upBlob].Weight += out[leftBlob].Weight

					for i := range indicies {
						if indicies[i] == leftBlob {
							indicies[i] = upBlob
						}
					}

					out[leftBlob].Weight = -1 // mark this output for later deletion

					addTo = upBlob
				}
			}

			out[addTo].X += xf * vf
			out[addTo].Y += yf * vf
			out[addTo].Xdev += xf2 * vf
			out[addTo].Ydev += yf2 * vf
			out[addTo].Weight += vf

			indicies[x] = addTo
		}
	}

	releaseIntArray(indicies)

	// finalize output array
	dead := 0
	for i := range out {
		if out[i].Weight == -1 {
			// sentinel, delete
			dead++
			continue
		}

		// convert running sums to mean and standard deviation
		out[i].X /= out[i].Weight
		out[i].Y /= out[i].Weight
		out[i].Xdev = math.Sqrt(out[i].Xdev/out[i].Weight - out[i].X*out[i].X)
		out[i].Ydev = math.Sqrt(out[i].Ydev/out[i].Weight - out[i].Y*out[i].Y)

		out[i-dead] = out[i]
	}

	return out[:len(out)-dead]
}
