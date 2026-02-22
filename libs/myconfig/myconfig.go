package myconfig

import (
	"encoding/json"
	"errors"
	"log"
	"os"

	"github.com/samber/lo"
)

const CONFIG_FILE_PATH = "config.json"

var tryPaths = [...]string{CONFIG_FILE_PATH, "/etc/" + CONFIG_FILE_PATH}

type IModule interface {
	GetType() string
}

type Module struct {
	Type string
}

func (m Module) GetType() string {
	return m.Type
}

type Segment []IModule
type Segments []Segment

func (t *Segment) UnmarshalJSON(data []byte) error {
	var modules []json.RawMessage
	if err := json.Unmarshal(data, &modules); err != nil {
		return err
	}

	for _, module := range modules {
		var mod Module
		if err := json.Unmarshal(module, &mod); err != nil {
			return err
		}

		switch mod.Type {
		case "cam":
			var cam_module CameraModule
			if err := json.Unmarshal(module, &cam_module); err != nil {
				return err
			}
			*t = append(*t, cam_module)
		case "buttons":
			var but_module ButtonModule
			if err := json.Unmarshal(module, &but_module); err != nil {
				return err
			}
			*t = append(*t, but_module)
		default:
			return errors.New("no module named '" + mod.Type + "'")
		}
	}

	return nil
}

type Button struct {
	Name     string
	Pin      uint64
	Default  int8
	IsToggle bool
}

type ButtonModule struct {
	Module
	Buttons []Button
}

type Config struct {
	WebPort       uint16
	Features      Features
	StateFilePath string
	Segments      Segments
}

type Web struct {
	Port uint16
}

type Features struct {
	Camera        bool
	Parallel      bool
	SavePinStatus bool
}

type CameraModule struct {
	Module
	Name       string
	Device     string
	Resolution string
	Fps        uint32
	Format     string
}

// TODO: (c *Config)
func GetButtonCount() int {
	return len(globalConfig.GetAllButton())
}

// TODO: make dirty modifier and cache count
func (c *Config) GetAllButton() []Button {
	buttons := make([]Button, 0)

	for _, segment := range c.Segments {
		for _, module := range segment {
			if v, ok := module.(ButtonModule); ok {
				buttons = append(buttons, v.Buttons...)
			}
		}
	}

	return buttons
}

// TODO: make dirty modifier and cache buttons
func (c *Config) GetButtonByPin(pin int) *Button {
	for _, segment := range c.Segments {
		for _, module := range segment {
			if v, ok := module.(ButtonModule); ok {
				if v, ok := lo.Find(v.Buttons, func(e Button) bool { return int(e.Pin) == pin }); ok {
					return &v
				}
			}
		}
	}

	return nil
}

var defaultConfig = Config{
	WebPort: 8080,
	Features: Features{
		Camera:        true,
		Parallel:      true,
		SavePinStatus: false,
	},
	StateFilePath: "state.json",
	Segments:      Segments{},
}

var ErrConfigNotFound = errors.New("no config files found")

func Load() error {
	var contents []byte
	for _, path := range tryPaths {
		data, err := os.ReadFile(path)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}

		contents = data
		log.Printf("Loading config file at %s", path)
		break
	}

	if contents == nil {
		return ErrConfigNotFound
	}

	config := new(Config)
	err := json.Unmarshal(contents, config)

	globalConfig = config

	return err
}

func Save(config *Config, path string) error {
	data, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data, os.FileMode(0o644))
	if err != nil {
		return err
	}
	log.Printf("Config saved to %s", path)

	return err
}

func LoadOrSaveDefault() error {
	err := Load()
	if errors.Is(err, ErrConfigNotFound) {
		err = Save(&defaultConfig, CONFIG_FILE_PATH)
		if err == nil {
			globalConfig = new(Config)
			*globalConfig = defaultConfig
		}
	}

	return err
}

var globalConfig *Config = nil

func Get() *Config {
	return globalConfig
}
