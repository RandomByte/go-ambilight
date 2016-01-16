package main

import (
	"bytes"
	dominantcolor "github.com/cenkalti/dominantcolor"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"os/exec"
	"syscall"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func main() {
	cam := Cam{}
	cam.Setup()
}

func loop() {
	file, err := os.Open("_test/pic.jpg")
	if err != nil {
		log.Fatal("Testpic missing", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	colors := computeDominatorColors(&img)
	log.Println(colors)
}

func computeDominatorColors(img *image.Image) [6]color.RGBA {
	screen := *getScreen(img)

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
	var (
		areas  [6]image.Image
		colors [6]color.RGBA
	)

	bounds := screen.Bounds()
	log.Println(bounds)

	x0 := float64(bounds.Dx() / 100)
	xA := int(math.Floor(x0 * 20))

	y := int(math.Floor(float64(bounds.Dy() / 2)))

	areaRect := image.Rect(bounds.Min.X, bounds.Min.Y, bounds.Min.X+xA, bounds.Min.Y+y)
	areas[0] = screen.(SubImager).SubImage(areaRect)

	areaRect = image.Rect(bounds.Min.X, bounds.Min.Y+y, bounds.Min.X+xA, bounds.Max.Y)
	areas[1] = screen.(SubImager).SubImage(areaRect)

	areaRect = image.Rect(bounds.Min.X+xA, bounds.Min.Y, bounds.Max.X-xA, bounds.Min.Y+y)
	areas[2] = screen.(SubImager).SubImage(areaRect)

	areaRect = image.Rect(bounds.Min.X+xA, bounds.Min.Y+y, bounds.Max.X-xA, bounds.Max.Y)
	areas[3] = screen.(SubImager).SubImage(areaRect)

	areaRect = image.Rect(bounds.Max.X-xA, bounds.Min.Y, bounds.Max.X, bounds.Min.Y+y)
	areas[4] = screen.(SubImager).SubImage(areaRect)

	areaRect = image.Rect(bounds.Max.X-xA, bounds.Min.Y+y, bounds.Max.X, bounds.Max.Y)
	areas[5] = screen.(SubImager).SubImage(areaRect)

	for i := 0; i < len(areas); i++ {
		colors[i] = dominantcolor.Find(areas[i])
		log.Println(i, colors[i])
	}

	return colors
}

func getScreen(img *image.Image) *image.Image {
	screenRect := image.Rect(1350, 663, 2200, 975)
	screen := (*img).(SubImager).SubImage(screenRect)
	return &screen
}

type Cam struct {
	Cmd *exec.Cmd
}

func (c *Cam) Setup() {
	// Start raspistill in signal mode
	c.Cmd = exec.Command("raspistill", "-n", "-s", "-t 0", "--thumb none", "-o pic.jpg")
	out := &bytes.Buffer{}
	c.Cmd.Stdout = out
	err := c.Cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Cam) Snapshot() {
	err := c.Cmd.Process.Signal(syscall.SIGUSR1)
	if err != nil {
		log.Fatal(err)
	}
}

func (c *Cam) Kill() {
	err := c.Cmd.Process.Signal(os.Kill)
	if err != nil {
		log.Fatal(err)
	}
}
