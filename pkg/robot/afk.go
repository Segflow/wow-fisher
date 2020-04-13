package robot

import (
	"time"

	"github.com/sirupsen/logrus"
)

var (
	afkDefParams = map[string]string{
		"happens":  "often",
		"duration": "5s",
	}
)

type AFKAction struct {
	spellKey string
	happens  string
}

func (f *AFKAction) Perform() {
	if !shouldPerform(f.happens) {
		logrus.Info("Skipping action: AFK")
		return
	}
	logrus.Info("Doing action: AFK")
	// TODO: make this random between X and Y user params.
	time.Sleep(1 * time.Second)
}

func buildAFKAction(params map[string]string) (Action, error) {
	p := mergeMaps(afkDefParams, params)
	return &AFKAction{
		happens: p["happens"],
	}, nil
}

func init() {
	RegisterBuilderFunc("afk", buildAFKAction)
}
