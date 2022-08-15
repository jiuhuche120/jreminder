package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "Jreminder"
	app.Usage = "Jreminder is a tool to help manage projects"

	app.Commands = []*cli.Command{
		initCmd,
		startCmd,
		versionCmd,
	}
	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
