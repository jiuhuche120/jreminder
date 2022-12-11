package config

import (
	"path/filepath"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/go-homedir"
)

const DefaultPath = "~/.jreminder/"
const DefaultName = "config.toml"

var globalConfig *Config
var once sync.Once

func GetConfig() (*Config, error) {
	var err error
	if globalConfig == nil {
		once.Do(func() {
			globalConfig, err = loadConfig()
		})
	}
	if err != nil {
		return nil, err
	}
	return globalConfig, nil
}

func loadConfig() (*Config, error) {
	path, err := DefaultConfigPath()
	if err != nil {
		return nil, err
	}
	config := Config{}
	if _, err := toml.DecodeFile(path, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func PathRoot() (string, error) {
	return homedir.Expand(DefaultPath)
}

func DefaultConfigPath() (string, error) {
	path, err := homedir.Expand(DefaultPath)
	if err != nil {
		return "", err
	}
	return filepath.Join(path, DefaultName), nil
}
