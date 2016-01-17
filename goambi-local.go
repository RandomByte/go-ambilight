package main

import (
	"encoding/json"
	"fmt"
	"github.com/RandomByte/dominantcolor"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	"log"
	"math"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer signal.Stop(sig)

	for {
		select {
		case <-time.After(1 * time.Millisecond):
			loop()
		case s := <-sig:
			log.Println("Got signal", s)
			log.Println("Quitting...")
			return
		}
	}
}

func loop() {
	// cam.Snapshot()
	transmitImg()

	file, err := os.Open("pic.jpg")
	if err != nil {
		log.Println("No picture found - skipping loop", err)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Panic(err)
	}
	bound := img.Bounds()
	m := image.NewRGBA(bound)
	draw.Draw(m, bound, img, bound.Min, draw.Src)

	var x image.Image
	x = m

	colors := computeDominatorColors(&x)
	log.Println(colors)

	sendToServer(colors)
}

func computeDominatorColors(img *image.Image) [6]color.RGBA {
	log.Println("Computing dominant colors...")
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

	processed := make(chan color.RGBA, 6)

	for i := 0; i < len(areas); i++ {
		go processArea(areas[i], processed)
	}

	for i := 0; i < len(areas); i++ {
		log.Println("Processed", i)
		colors[i] = <-processed
	}

	return colors
}

func processArea(area image.Image, processed chan color.RGBA) {
	color := dominantcolor.Find(area)
	processed <- color
}

func getScreen(img *image.Image) *image.Image {
	return img
	// Full res is 2592 x 1944  e.g. 1 / 2592 * x
	// screenRect := image.Rect(1335, 747, 2184, 1231)
	// screen := (*img).(SubImager).SubImage(screenRect)
	// return &screen
}

func sendToServer(colors [6]color.RGBA) {
	conn, err := net.Dial("udp", "192.168.2.6:64001")
	if err != nil {
		fmt.Println(err)
		return
	}
	json, err := json.Marshal(colors)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Fprintf(conn, string(json))
	conn.Close()
}

func transmitImg() {
	cmd := exec.Command("scp", "serpens:~/pic.jpg", "./")
	_, err := cmd.Output()
	if err != nil {
		log.Panic(err)
	}
}
