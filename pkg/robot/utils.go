package robot

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"os"
	"time"

	"github.com/go-vgo/robotgo"
	"golang.org/x/exp/rand"
)

const (
	pressDelayMin = 50  // ms
	pressDelayMax = 100 // ms
	wowWindowName = "Wow"
)

func openWow() {
	err := robotgo.ActivePID(12389)
	if err != nil {
		log.Fatal(err)
	}
}

func pixelDistance(a, b color.RGBA) int {
	return int(math.Abs(float64(a.R-b.R)) + math.Abs(float64(a.G-b.G)) + math.Abs(float64(a.B-b.B)))
}

func closestPixelToColor(img *image.RGBA, c color.RGBA) (int, int, int) {
	closest := -1
	closestW := -1
	closestH := -1

	r := img.Bounds()
	width, height := r.Dx(), r.Dy()
	dxy := 1
	for w := 100; w < width; w += dxy {
		for h := 100; h < height; h += dxy {
			pixel := img.RGBAAt(w, h)

			pixelLeft := img.RGBAAt(w-3, h)
			pixelTop := img.RGBAAt(w, h-3)
			d := pixelDistance(c, pixel) + pixelDistance(c, pixelLeft) + pixelDistance(c, pixelTop)
			if closest == -1 || d < closest {
				closest = d
				closestW = w
				closestH = h
			}
		}
	}

	return closestW, closestH, closest
}

func randSleep(minMS, maxMS int) {
	msDelay := minMS + rand.Intn(maxMS-minMS)
	time.Sleep(time.Duration(msDelay) * time.Millisecond)
}

func pressKey(key string) {
	randSleep(pressDelayMin, pressDelayMax)
	robotgo.KeyTap(key)
}

func saveImage(filename string, im image.Image) error {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	return png.Encode(f, im)
}
