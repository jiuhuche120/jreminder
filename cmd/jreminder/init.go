package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/gobuffalo/packr/v2"
	"github.com/jiuhuche120/jreminder/pkg/config"
	"github.com/urfave/cli/v2"
)

const DefaultConfig = "config.toml"

var initCmd = &cli.Command{
	Name:   "init",
	Usage:  "init config home for Jreminder",
	Action: Initialize,
}

func Initialize(ctx *cli.Context) error {
	box := packr.New("Jreminder Config", "../../config")
	path, err := config.PathRoot()
	if err != nil {
		return err
	}
	_, err = os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.Mkdir(path, 0755)
			if err != nil {
				return err
			}
		}
	}
	_, err = os.Stat(filepath.Join(path, config.DefaultName))
	if err != nil {
		if os.IsNotExist(err) {
			data, err := box.Find(DefaultConfig)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(filepath.Join(path, config.DefaultName), data, 0755)
			if err != nil {
				return err
			}
		}
	} else {
		fmt.Println("Jreminder configuration file already exists")
		fmt.Println("reinitializing would overwrite your configuration, Y/N?")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		if input.Text() == "Y" || input.Text() == "y" {
			data, err := box.Find(DefaultConfig)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(filepath.Join(path, config.DefaultName), data, 0755)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
