package myconfig

import (
	"bytes"
	"errors"
	"log"
	"os"

	"github.com/pelletier/go-toml"
)

const DEFAULT_PATH = "rcrs.toml"
var tryPaths = [...]string{DEFAULT_PATH, "/etc/rcrs.toml"}

type Config struct {
	Web Web
	Peripheral Peripheral
	Camera []Camera
	Parallel Parallel
}

type Web struct {
	Port uint16
}

type Peripheral struct {
	Camera bool
	Parallel bool
}

type Camera struct {
	Name string
	Device string
	Resolution string
	Fps uint32
	Format string
}

type Parallel struct {
	Config string
}

var defaultConfig = Config{
	Web: Web{
		Port: 8080,
	},
	Peripheral: Peripheral{
		Camera: true,
		Parallel: true,
	},
	Camera: []Camera{},
	Parallel: Parallel{Config: "pins.txt"},
}

var ErrConfigNotFound = errors.New("No config files found");

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
		log.Printf("Loading config file at %s", path);
		break
	}

	if contents == nil {
		return ErrConfigNotFound
	}

	config := new(Config)
	err := toml.Unmarshal(contents, config)

	globalConfig = config

	return err
}

func Save(config *Config, path string) error {
	var data bytes.Buffer
	encoder := toml.NewEncoder(&data)
	encoder.Order(toml.OrderPreserve)
	encoder.Indentation("")

	err := encoder.Encode(config)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, data.Bytes(), os.FileMode(0o644))
	if err != nil {
		return err
	}
	log.Printf("Config saved to %s", path)

	return err
}

func LoadOrSaveDefault() error {
	err := Load()
	if errors.Is(err, ErrConfigNotFound) {
		err = Save(&defaultConfig, DEFAULT_PATH)
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
