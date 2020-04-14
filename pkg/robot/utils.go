package robot

import (
	"fmt"
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

func findRegionWithColor(img *image.RGBA, c color.RGBA) (int, int, int) {
	highest := 0
	highestX := -1
	highestY := -1

	dxy := 1
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y += dxy {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x += dxy {
			pixel := img.RGBAAt(x, y)

			checkPixels := []color.RGBA{
				pixel,
				img.RGBAAt(x-1, y),
				img.RGBAAt(x+1, y),
				img.RGBAAt(x, y-1),
				img.RGBAAt(x, y+1),
				img.RGBAAt(x-2, y),
				img.RGBAAt(x+2, y),
				img.RGBAAt(x, y-2),
				img.RGBAAt(x, y+2),
			}

			d := 0
			for _, p := range checkPixels {
				if p.R == c.R && p.G == c.G && p.B == c.B {
					d++
				}
			}
			// fmt.Printf("d: %d\n", d)
			if highest == 0 || d > highest {
				highest = d
				highestX = x
				highestY = y
			}
		}
	}

	return highestX, highestY, highest
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

			checkPixels := []color.RGBA{
				pixel,
				img.RGBAAt(w-1, h),
				img.RGBAAt(w+1, h),
				img.RGBAAt(w, h-1),
				img.RGBAAt(w, h+1),
				img.RGBAAt(w-2, h),
				img.RGBAAt(w+2, h),
				img.RGBAAt(w, h-2),
				img.RGBAAt(w, h+2),
				img.RGBAAt(w-3, h),
				img.RGBAAt(w+3, h),
				img.RGBAAt(w, h-3),
				img.RGBAAt(w, h+3),
				img.RGBAAt(w-4, h),
				img.RGBAAt(w+4, h),
				img.RGBAAt(w, h-4),
				img.RGBAAt(w, h+4),
			}

			d := 0
			for _, p := range checkPixels {
				d += pixelDistance(c, p)
			}

			if closest == -1 || d < closest {
				closest = d
				closestW = w
				closestH = h
			}
		}
	}

	return closestW, closestH, closest
}

func convertGray(img image.Image) image.Image {
	grayImg := image.NewGray(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			R, G, B, _ := img.At(x, y).RGBA()
			//Luma: Y = 0.2126*R + 0.7152*G + 0.0722*B
			Y := (0.2126*float64(R) + 0.7152*float64(G) + 0.0722*float64(B)) * (255.0 / 65535)
			grayPix := color.Gray{uint8(Y)}
			grayImg.Set(x, y, grayPix)
		}
	}
	return grayImg
}

func makeBluePixelFilter(blueThreshold uint8) pixelFilterFunc {
	return func(pixel color.RGBA) bool {
		// Blue should be the dominant color.
		if pixel.R > pixel.B || pixel.G > pixel.B {
			return false
		}

		// Red and Green should be low
		if pixel.R > 80 || pixel.G > 80 {
			return false
		}

		// Blue should be above the threshold
		return pixel.B > blueThreshold
	}
}

type pixelFilterFunc func(pixel color.RGBA) bool

// togglePixels sets pixels that matches the filter func to white and the others to black.
func togglePixels(img image.Image, filter pixelFilterFunc) *image.RGBA {
	newImg := image.NewRGBA(img.Bounds())
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {

			pixel := color.RGBA{}
			switch i := img.(type) {
			case *image.RGBA:
				pixel = i.RGBAAt(x, y)
			case *image.NRGBA:
				nrgba := i.NRGBAAt(x, y)
				pixel.A = nrgba.A
				pixel.R = nrgba.R
				pixel.G = nrgba.G
				pixel.B = nrgba.B
			}

			c := blackColor
			if filter(pixel) {
				c = whiteColor
			}
			newImg.Set(x, y, c)
		}
	}
	return newImg
}

func mergeMaps(one, two map[string]string) map[string]string {
	newMap := make(map[string]string)
	for k, v := range one {
		newMap[k] = v
	}

	for k, v := range two {
		newMap[k] = v
	}

	return newMap
}

func imageDiff(image1, image2 string) (int64, error) {
	img1, err := loadImage(image1)
	if err != nil {
		return 0, err
	}
	img2, err := loadImage(image2)
	if err != nil {
		return 0, err
	}

	if img1.Bounds() != img2.Bounds() {
		return 0, fmt.Errorf("image bounds not equal: %+v, %+v", img1.Bounds(), img2.Bounds())
	}

	accumError := int64(0)
	for i := 0; i < len(img1.Pix); i++ {
		accumError += int64(sqDiffUInt8(img1.Pix[i], img2.Pix[i]))
	}

	return int64(math.Sqrt(float64(accumError))), nil
}

func sqDiffUInt8(x, y uint8) uint64 {
	d := uint64(x) - uint64(y)
	return d * d
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
