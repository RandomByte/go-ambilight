package main

import (
	"image"
	"image/color"
	"image/jpeg"
	"log"
	"math"
	"os"
	"strconv"
)

func main() {
	file, err := os.Open("_test/pic.jpg")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	screenRect := image.Rect(1350, 663, 2200, 975)
	screen := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(screenRect)

	/*

	   Dividing screen into 6 areas

	              A         B         A
	    ----- -----------------------------
	          |       |           |       |
	    50%   |   0   |     2     |   4   |
	          |       |           |       |
	    ----- -----------------------------
	          |       |           |       |
	    50%   |   1   |     3     |   5   |
	          |       |           |       |
	    ----- -----------------------------
	          |  20%  |    40%    |  20%  |

	*/
	var areas [6]image.Image

	x0 := float64(screenRect.Dx() / 100)
	xA := int(math.Floor(x0 * 20))

	y := int(math.Floor(float64(screenRect.Dy() / 2)))

	bounds := screenRect

	areaRect := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+xA, bounds.Min.Y+y)
	areas[0] = screen.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(areaRect)

	areaRect = image.Rect(bounds.Min.X, bounds.Min.Y+y, bounds.Min.X+xA, bounds.Max.Y)
	areas[1] = screen.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(areaRect)

	areaRect = image.Rect(bounds.Min.X+xA, bounds.Min.Y, bounds.Max.X-xA, bounds.Min.Y+y)
	areas[2] = screen.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(areaRect)

	areaRect = image.Rect(bounds.Min.X+xA, bounds.Min.Y+y, bounds.Max.X-xA, bounds.Max.Y)
	areas[3] = screen.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(areaRect)

	areaRect = image.Rect(bounds.Max.X-xA, bounds.Min.Y, bounds.Max.X, bounds.Min.Y+y)
	areas[4] = screen.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(areaRect)

	areaRect = image.Rect(bounds.Max.X-xA, bounds.Min.Y+y, bounds.Max.X, bounds.Max.Y)
	areas[5] = screen.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(areaRect)

	for i := 0; i < len(areas); i++ {
		color := getAverageColor(areas[i])
		log.Println(i, color)

		out, err := os.Create("./output-" + strconv.Itoa(i) + ".jpg")
		if err != nil {
			log.Fatal(err)
		}
		err = jpeg.Encode(out, areas[i], nil)
		if err != nil {
			log.Fatal(err)
		}
	}

}

func getAverageColor(img image.Image) color.RGBA {
	var (
		r, g, b uint32
	)

	bound := img.Bounds()
	for y := bound.Min.Y; y < bound.Max.Y; y++ {
		for x := bound.Min.X; x < bound.Max.X; x++ {
			color := img.At(x, y)
			pr, pg, pb, _ := color.RGBA()

			r += pr
			g += pg
			b += pb
		}
	}

	count := uint32(bound.Dx() * bound.Dy()) // width * height

	r /= count
	g /= count
	b /= count

	return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
}
