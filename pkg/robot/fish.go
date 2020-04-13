package robot

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/sirupsen/logrus"
)

const (
	defFishingSpellKey = "f6"
	fishingMaxDuration = 21 * time.Second // 20 second + cast time + pick animation time
	screenCaptureBox   = 30
)

var (
	fishingDefParams = map[string]string{
		"happens":         "always",
		"spell_key":       defFishingSpellKey,
		"captures_dir":    "./captures",
		"catch_threshold": "2000",
	}

	bobberColor = color.RGBA{
		R: 183,
		G: 119,
		B: 88,
	}
)

type FishAction struct {
	spellKey       string
	happens        string
	capturesDir    string
	catchThreshold int
}

// Perform performs the fishing action
func (f *FishAction) Perform() {
	if !shouldPerform(f.happens) {
		logrus.Info("Skipping action: fishing")
		return
	}
	logrus.Info("Doing action: fishing")
	pressKey(f.spellKey)
	timeout := time.After(fishingMaxDuration)

	// TODO: move the cursor to the right position.
	f.moveMouseToCatchPosition()

	ticker := time.NewTicker(250 * time.Millisecond)
	i := 0
FOR:
	for {
		select {
		case <-timeout:
			logrus.Info("Action fishing: TIMEOUT")
			return
		case <-ticker.C:
			i++
			capturePath := path.Join(f.capturesDir, fmt.Sprintf("cap%d.png", i))
			saveMouseBoxCapture(capturePath, screenCaptureBox)
			if f.isFishCatched(i) {
				robotgo.MouseClick("right", false)
				break FOR
			}
		}
	}
	randSleep(1000, 2000) // Sleep between 1 to 2 second after each catch.
}

func saveMouseBoxCapture(path string, box int) {
	mx, my := robotgo.GetMousePos()

	x, y := mx-(box/2), my-(box/2)
	w, h := box, box

	robotgo.SaveCapture(path, x, y, w, h)
}

func (f *FishAction) moveMouseToCatchPosition() error {
	randSleep(2000, 2500) // Leave some time for the animation to finist
	capturePath := path.Join(f.capturesDir, "lookupBobber.png")
	robotgo.SaveCapture(capturePath)

	img, err := loadImage(capturePath)
	if err != nil {
		return err
	}

	x, y, _ := closestPixelToColor(img, bobberColor)
	robotgo.MoveMouseSmooth(x, y)
	return nil
}

func (f *FishAction) isFishCatched(index int) bool {
	if index == 1 {
		return false
	}
	capturePath1 := path.Join(f.capturesDir, fmt.Sprintf("cap%d.png", 1))
	capturePath2 := path.Join(f.capturesDir, fmt.Sprintf("cap%d.png", index))
	diff, err := imageDiff(capturePath1, capturePath2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Diff %d-%d: %d\n", 1, index, diff)
	return diff >= int64(f.catchThreshold)
}

func buildFishAction(params map[string]string) (Action, error) {
	p := mergeMaps(fishingDefParams, params)

	threshold, err := strconv.Atoi(p["catch_threshold"])
	if err != nil {
		return nil, err
	}

	return &FishAction{
		spellKey:       p["spell_key"],
		happens:        p["happens"],
		capturesDir:    p["captures_dir"],
		catchThreshold: threshold,
	}, nil
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

func init() {
	RegisterBuilderFunc("fish", buildFishAction)
}

func loadImage(filename string) (*image.RGBA, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	i, _, err := image.Decode(f)
	p, ok := i.(*image.RGBA)
	if !ok {
		return nil, fmt.Errorf("not a png file")
	}
	return p, nil
}
