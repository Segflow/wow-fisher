package robot

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/go-vgo/robotgo"
	"github.com/oliamb/cutter"
	"github.com/sirupsen/logrus"
)

const (
	defFishingSpellKey     = "f6"
	defToggleBlueThreshold = "100"
	defCheckCatchPeriod    = "150"
	fishingMaxDuration     = 21 * time.Second // 20 second + cast time + pick animation time
	screenCaptureBox       = 30
)

var (
	fishingDefParams = map[string]string{
		"happens":               "always",
		"spell_key":             defFishingSpellKey,
		"captures_dir":          "./captures",
		"catch_threshold":       "2000",
		"toggle_blue_threshold": defToggleBlueThreshold,
		"check_catch_period":    defCheckCatchPeriod,
	}

	blackColor = color.RGBA{
		R: 0,
		G: 0,
		B: 0,
		A: 255,
	}

	whiteColor = color.RGBA{
		R: 255,
		G: 255,
		B: 255,
		A: 255,
	}
)

type FishAction struct {
	spellKey            string
	happens             string
	capturesDir         string
	catchThreshold      int
	toggleBlueThreshold uint8
	checkCatchPeriod    time.Duration
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

	ticker := time.NewTicker(f.checkCatchPeriod * time.Millisecond)
	i := 0
FOR:
	for {
		select {
		case <-timeout:
			logrus.Info("Action fishing: TIMEOUT")
			break FOR
		case <-ticker.C:
			i++
			capturePath := path.Join(f.capturesDir, fmt.Sprintf("cap%d.png", i))
			saveMouseBoxCapture(capturePath, screenCaptureBox)
			if f.isFishCatched(i) {
				fmt.Printf("Fish caught at %d\n", i)
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

	// TODO: make crop option configurable
	centerCropped, _ := cutter.Crop(img, cutter.Config{
		Width:  1000,
		Height: 700,
		Mode:   cutter.Centered,
	})

	filtered := togglePixels(centerCropped, makeBluePixelFilter(f.toggleBlueThreshold))

	x, y, _ := findRegionWithColor(filtered, whiteColor)
	robotgo.MoveMouseSmooth(x+10, y+10)
	return nil
}

func (f *FishAction) isFishCatched(index int) bool {
	if index < 3 {
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

	toggleBlueThreshold, err := strconv.Atoi(p["toggle_blue_threshold"])
	if err != nil {
		return nil, err
	}

	checkCatchPeriod, err := strconv.Atoi(p["check_catch_period"])
	if err != nil {
		return nil, err
	}

	return &FishAction{
		spellKey:            p["spell_key"],
		happens:             p["happens"],
		capturesDir:         p["captures_dir"],
		catchThreshold:      threshold,
		toggleBlueThreshold: uint8(toggleBlueThreshold),
		checkCatchPeriod:    time.Duration(checkCatchPeriod),
	}, nil
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
