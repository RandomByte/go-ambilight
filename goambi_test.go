package main

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"os"
	"testing"
)

func TestPix(t *testing.T) {

	file, err := os.Open("_test/pic2.jpg")
	if err != nil {
		t.Fatal("Testpic missing", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		t.Fatal(err)
	}
	bound := img.Bounds()
	m := image.NewRGBA(bound)
	draw.Draw(m, bound, img, bound.Min, draw.Src)

	i := m.PixOffset(1, 0)
	t.Log("==", i)
}

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

	bound := img.Bounds()
	m := image.NewRGBA(bound)
	draw.Draw(m, bound, img, bound.Min, draw.Src)

	var x image.Image
	x = m

	colors := computeDominatorColors(&x)

	if len(colors) == 0 {
		t.Error("No colors returned")
	}
	fmt.Println(colors[0])
	if colors[0].R == 0 && colors[0].G == 0 && colors[0].B == 0 {
		t.Error("Area #0 is black. Expected something colorful")
	}
	fmt.Println(colors[1])
	if colors[1].R < 200 || colors[1].G > 180 || colors[1].B > 180 {
		t.Error("Area #1 should be mainly red")
	}
}

func BenchmarkComputeDominatorColors(b *testing.B) {
	file, err := os.Open("_test/pic2.jpg")
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

func BenchmarkLoadImageAndComputeDominatorColors(b *testing.B) {
	for i := 0; i < b.N; i++ {
		img := loadImage()
		colors := computeDominatorColors(img)
		if len(colors) == 0 {
			b.Error("No colors returned")
		}
	}
}
