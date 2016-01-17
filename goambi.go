package main

import (
	"encoding/json"
	"fmt"
	"github.com/cenkalti/dominantcolor"
	"image"
	"image/color"
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

	cam := Cam{}
	defer cam.Kill()

	cam.Setup()

	for {
		select {
		case <-time.After(500 * time.Millisecond):
			loop(&cam)
		case s := <-sig:
			log.Println("Got signal", s)
			log.Println("Quitting...")
			return
		}
	}
}

func loop(cam *Cam) {
	cam.Snapshot()

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
	colors := computeDominatorColors(&img)
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
	screenRect := image.Rect(1335, 747, 2184, 1231)
	screen := (*img).(SubImager).SubImage(screenRect)
	return &screen
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

type Cam struct {
	Cmd *exec.Cmd
}

func (c *Cam) Setup() {
	// Start raspistill in signal mode
	log.Println("Initializing raspistill process...")
	c.Cmd = exec.Command("raspistill", "-v", "-n", "-s", "-t", "0", "--thumb", "none", "-o", "pic.jpg")

	err := c.Cmd.Start()
	if err != nil {
		log.Panic(err)
	}
	// Should wait for "Waiting for SIGUSR1" in the stdout - but was unable to get it running :(
	// So we just assume that nothing bad will happen when sending signals to it too early
	// Also, just in case, we wait a sec
	time.Sleep(1000 * time.Millisecond)
}

func (c *Cam) Snapshot() {
	log.Println("Triggering snapshot...")
	err := c.Cmd.Process.Signal(syscall.SIGUSR1)
	if err != nil {
		log.Panic(err)
	}

	// Same problem as in setup - don't know when the picture got taken, just hope it went through faster than 500ms
	time.Sleep(500 * time.Millisecond)
}

func (c *Cam) Kill() {
	err := c.Cmd.Process.Signal(os.Kill)
	if err != nil {
		log.Panic(err)
	}
}
