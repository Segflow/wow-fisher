package robot

import (
	"fmt"
	"image"
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
	screenImage := "../../captures/lookupBobber.png"

	img, err := loadImage(screenImage)
	if err != nil {
		log.Fatal(err)
	}

	centerCropped, _ := cutter.Crop(img, cutter.Config{
		Width:  1000,
		Height: 700,
		Mode:   cutter.Centered,
	})
	saveImage("../../captures/cropped_lookupBobber.png", centerCropped)

	bluefiltered := togglePixels(centerCropped, makeBluePixelFilter(100))
	saveImage("../../captures/filtered_lookupBobber.png", bluefiltered)

	x, y, count := findRegionWithColor(bluefiltered, whiteColor)
	fmt.Printf("white count: %d\n", count)

	croppedImg, err := cutter.Crop(img, cutter.Config{
		Width:  screenCaptureBox,
		Height: screenCaptureBox,
		Anchor: image.Point{x + 5, y + 5},
	})

	p := "../../captures/closest_lookupBobber.png"
	if err := saveImage(p, croppedImg); err != nil {
		log.Fatal(err)
	}

}
