package config

import (
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/mitchellh/go-homedir"
)

const DefaultPath = "~/.jreminder/"
const DefaultName = "config.toml"

type Config struct {
	Log          Log                   `toml:"log"`
	Github       Github                `toml:"github"`
	Repositories map[string]Repository `toml:"repositories"`
	Rules        Rules                 `toml:"rules"`
	Members      map[string]Member     `toml:"members"`
	Webhook      map[string]Webhook    `toml:"webhook"`
}

type Log struct {
	Level string `toml:"level"`
}

type Github struct {
	Token string `toml:"token"`
}

type Repository struct {
	Repository string   `toml:"repository"`
	Project    string   `toml:"project"`
	Rules      []string `toml:"rules"`
	Webhook    []string `toml:"webhook"`
}

type Rules struct {
	CheckMainBranchMerged   map[string]*CheckMainBranchMerged   `toml:"checkMainBranchMerged"`
	CheckPullRequestTimeout map[string]*CheckPullRequestTimeout `toml:"checkPullRequestTimeout"`
}

type CheckMainBranchMerged struct {
	Base string `toml:"base"`
	Head string `toml:"head"`
	Cron string `toml:"cron"`
}

type CheckPullRequestTimeout struct {
	Timeout string `toml:"timeout"`
	Cron    string `toml:"cron"`
}

type Member struct {
	Phone string `toml:"phone"`
}

type Webhook struct {
	Webhook string `toml:"webhook"`
}

func LoadConfig() (*Config, error) {
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
