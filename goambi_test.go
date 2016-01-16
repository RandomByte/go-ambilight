package main

import (
	"image"
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
