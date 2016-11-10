package main

import (
	"flag"
	"fmt"
	"github.com/RandomByte/colorfinder"
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
	"runtime"
	"strconv"
	"syscall"
	"time"
)

var targetAddress string
var lastColors [6]color.RGBA

type SubImager interface {
	SubImage(r image.Rectangle) image.Image
}

func main() {

	ip := flag.String("ip", "", "Target IP")
	port := flag.Int("port", 0, "Target port")
	flag.Parse()

	if *ip == "" || *port == 0 {
		log.Println("Missing target IP or port")
		return
	}

	targetAddress = *ip + ":" + strconv.Itoa(*port)
	log.Println("Target address:", targetAddress)

	log.Println(runtime.GOMAXPROCS(8))

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill, syscall.SIGTERM)
	defer signal.Stop(sig)

	cam := Cam{}
	defer cam.Kill()

	cam.Setup()

	for {
		select {
		case <-time.After(1 * time.Millisecond):
			loop()
		case s := <-sig:
			log.Println("Got signal:", s)
			log.Println("Quitting...")
			return
		}
	}
}

func loop() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered", r)
		}
	}()
	img := loadImage()
	colors := computeDominatorColors(img, lastColors)
	log.Println(colors)
	lastColors = colors

	sendToServer(colors)
}

func loadImage() *image.Image {
	file, err := os.Open("pic.jpg")
	if err != nil {
		log.Panic("No picture found", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Panic(err)
	}
	bound := img.Bounds()
	rgbaImg := image.NewRGBA(bound)
	draw.Draw(rgbaImg, bound, img, bound.Min, draw.Src)

	var ret image.Image = rgbaImg
	return &ret
}

func computeDominatorColors(img *image.Image, lastColors [6]color.RGBA) [6]color.RGBA {
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

	// Ignore area #3
	// areaRect = image.Rect(bounds.Min.X+xA, bounds.Min.Y+y, bounds.Max.X-xA, bounds.Max.Y)
	// areas[3] = screen.(SubImager).SubImage(areaRect)

	areaRect = image.Rect(bounds.Max.X-xA, bounds.Min.Y, bounds.Max.X, bounds.Min.Y+y)
	areas[4] = screen.(SubImager).SubImage(areaRect)

	areaRect = image.Rect(bounds.Max.X-xA, bounds.Min.Y+y, bounds.Max.X, bounds.Max.Y)
	areas[5] = screen.(SubImager).SubImage(areaRect)

	processed := make(chan color.RGBA, 6)

	for i := 0; i < len(areas); i++ {
		if i == 3 { // Ignore area #3
			continue
		}
		go processArea(areas[i], lastColors[i], processed)
	}

	for i := 0; i < len(areas); i++ {
		if i == 3 { // Ignore area #3
			continue
		}
		log.Println("Processed area", i)
		colors[i] = <-processed
	}

	return colors
}

func processArea(area image.Image, lastColor color.RGBA, processed chan color.RGBA) {
	newColor := colorfinder.Find(area.(*image.RGBA))
	if lastColor != (color.RGBA{}) {
		// Compare to last color
		if getDifferenceOfInts(newColor.R, lastColor.R) < 10 {
			// if newColor.R < lastColor.R {
			log.Println("Prioritizing last red")
			newColor.R = lastColor.R
			// }
		}
		if getDifferenceOfInts(newColor.G, lastColor.G) < 10 {
			// if newColor.G < lastColor.G {
			log.Println("Prioritizing last green")
			newColor.G = lastColor.G
			// }
		}
		if getDifferenceOfInts(newColor.B, lastColor.B) < 10 {
			// if newColor.B < lastColor.B {
			log.Println("Prioritizing last blue")
			newColor.B = lastColor.B
			// }
		}
	}
	processed <- newColor
}

func getDifferenceOfInts(x, y uint8) int {
	return int(math.Abs(float64(x) - float64(y)))
}

func getScreen(img *image.Image) *image.Image {
	return img
}

func sendToServer(colors [6]color.RGBA) {
	var payload string

	// Put together our payload. Will look something like this:
	// /0:R070G045B028/1:R068G046B031/2:R066G044B029/3:R064G039B030/4:R064G040B028/5:R070G048B031
	// For six areas, it'll always have a length of 90 chars.
	for i := 0; i < len(colors); i++ {
		if i == 3 { // Ignore area #3
			payload += "/3:R000G000B000"
			continue
		}
		payload += fmt.Sprintf("/%v:R%03dG%03dB%03d", i, colors[i].R, colors[i].G, colors[i].B)
	}

	conn, err := net.Dial("udp", targetAddress)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Fprintf(conn, payload)
	conn.Close()
}

type Cam struct {
	Cmd *exec.Cmd
}

func (c *Cam) Setup() {
	// Start raspistill in timelapse + burst mode
	log.Println("Initializing raspistill process...")
	c.Cmd = exec.Command("raspistill", "-v", "-n", "-t", "0", "--thumb", "none", "-o", "pic.jpg", "-roi", "0.51,0.35,0.33,0.24", "-w", "648", "-h", "486", "-tl", "0", "-bm", "-mm", "spot", "-ex", "spotlight", "-awb", "shade", "-ISO", "100")

	err := c.Cmd.Start()
	if err != nil {
		log.Panic(err)
	}
	// Should wait for something in the stdout - but was unable to get it working :(
	// So we just assume that nothing bad will happen when proceeding too soon
	// Also, just in case, we wait a sec
	time.Sleep(1000 * time.Millisecond)
}

func (c *Cam) Kill() {
	err := c.Cmd.Process.Signal(os.Kill)
	if err != nil {
		log.Panic(err)
	}
}
