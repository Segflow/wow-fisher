package robot

import (
	"fmt"
	"time"

	"github.com/segflow/wow-fisher/pkg/config"
	"golang.org/x/exp/rand"
)

type Action interface {
	Perform()
}

type Robot struct {
	config  *config.Config
	actions []Action
}

func New(cfg *config.Config) (*Robot, error) {
	actions, err := buildActions(cfg.ActionDefs)
	if err != nil {
		return nil, err
	}

	return &Robot{
		config:  cfg,
		actions: actions,
	}, nil
}

func buildActions(defs []config.ActionDefinition) ([]Action, error) {
	var actions []Action
	for _, def := range defs {
		builder := builderFor(def.Name)
		if builder == nil {
			return nil, fmt.Errorf("No builder for action %q", def.Name)
		}

		action, err := builder.Build(def.Params)
		if err != nil {
			return nil, fmt.Errorf("Cannot build action %q: %w", def.Name, err)
		}
		actions = append(actions, action)
	}
	return actions, nil
}

// Start starts the bot main loop.
func (r *Robot) Start() {
	openWow()
	time.Sleep(2 * time.Second)
	for {
		for _, action := range r.actions {
			action.Perform()
			time.Sleep(100 * time.Millisecond)
		}
	}
}

var (
	probMap = map[string]float32{
		"":       0.0,
		"never":  0.0,
		"rarely": 0.2,
		"often":  0.5,
		"always": 1.1,
	}
)

func shouldPerform(howOften string) bool {
	r := rand.Float32()
	return r < probMap[howOften]
}
