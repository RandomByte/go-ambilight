package main

import (
	"image"
	"image/draw"
	_ "image/jpeg"
	"os"
	"testing"
)

func TestComputeDominatorColors(t *testing.T) {

	file, err := os.Open("_test/pic.jpg")
	if err != nil {
		t.Fatal("Testpic missing", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatal(err)
	}

	colors := computeDominatorColors(&img)
	if len(colors) == 0 {
		t.Error("No colors returned")
	}
}

func BenchmarkComputeDominatorColors(b *testing.B) {

	file, err := os.Open("_test/pic.jpg")
	if err != nil {
		b.Fatal("Testpic missing", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		b.Fatal(err)
	}

	bound := img.Bounds()
	m := image.NewRGBA(bound)
	draw.Draw(m, bound, img, bound.Min, draw.Src)

	var x image.Image
	x = m

	for i := 0; i < b.N; i++ {
		colors := computeDominatorColors(&x)
		if len(colors) == 0 {
			b.Error("No colors returned")
		}
	}

}
