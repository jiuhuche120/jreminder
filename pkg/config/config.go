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
	Holiday      Holiday               `toml:"holiday"`
	Github       Github                `toml:"github"`
	Account      Account               `toml:"account"`
	Repositories map[string]Repository `toml:"repositories"`
	Teambitions  map[string]Teambition `toml:"teambition"`
	Rules        Rules                 `toml:"rules"`
	Members      map[string]Member     `toml:"members"`
	Webhook      map[string]Webhook    `toml:"webhook"`
}

type Log struct {
	Level        string `toml:"level"`
	ReportCaller bool   `toml:"report_caller"`
}

type Holiday struct {
	Path string `toml:"path"`
}

type Github struct {
	Token string `toml:"token"`
}

type Account struct {
	Email    string `toml:"email"`
	Password string `toml:"password"`
}

type Repository struct {
	Repository string   `toml:"repository"`
	Project    string   `toml:"project"`
	Rules      []string `toml:"rules"`
	Webhook    []string `toml:"webhook"`
}

type Teambition struct {
	Project string   `toml:"project"`
	App     string   `toml:"app"`
	Rules   []string `toml:"rules"`
	Webhook []string `toml:"webhook"`
}

type Rules struct {
	CheckMainBranchMerged   map[string]*CheckMainBranchMerged   `toml:"checkMainBranchMerged"`
	CheckPullRequestTimeout map[string]*CheckPullRequestTimeout `toml:"checkPullRequestTimeout"`
	CheckTeambitionTimeout  map[string]*CheckTeambitionTimeout  `toml:"checkTeambitionTimeout"`
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

type CheckTeambitionTimeout struct {
	Cron string `toml:"cron"`
}

type Member struct {
	Github string `toml:"github"`
	Name   string `toml:"name"`
	Phone  string `toml:"phone"`
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
