package robot

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"testing"

	"github.com/oliamb/cutter"
)

type Color struct {
	R, G, B int
}

func TestImageDiff(t *testing.T) {
	image1 := "../../captures/cap30.png"
	image2 := "../../captures/cap177.png"
	d, err := imageDiff(image1, image2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Difference between %q and %q: %d\n", image1, image2, d)
}

func TestFindBobber(t *testing.T) {
	screenImage := "../../locationBobber1.png"

	img, err := loadImage(screenImage)
	if err != nil {
		log.Fatal(err)
	}

	bobberColor := color.RGBA{
		R: 183,
		G: 119,
		B: 88,
	}

	closestW, closestH, distance := closestPixelToColor(img, bobberColor)
	fmt.Printf("Distance: %d\n", distance)

	croppedImg, err := cutter.Crop(img, cutter.Config{
		Width:  screenCaptureBox,
		Height: screenCaptureBox,
		Anchor: image.Point{closestW - 20, closestH - 20},
	})

	p := fmt.Sprintf("../../captures/closest_lookupBobber-%d-%d.png", closestW, closestH)
	if err := saveImage(p, croppedImg); err != nil {
		log.Fatal(err)
	}

}
