package config

import (
	"encoding/json"
	"os"
)

type ActionDefinition struct {
	Name   string            `json:"name"`
	Params map[string]string `json:"params"`
}

type Config struct {
	ActionDefs []ActionDefinition `json:"actions"`
}

func ReadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	c := new(Config)
	if err := json.NewDecoder(f).Decode(c); err != nil {
		return nil, err
	}

	return c, nil
}
