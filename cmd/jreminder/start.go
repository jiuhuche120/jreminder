package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"

	"github.com/jiuhuche120/jreminder/internal/app"
	"github.com/jiuhuche120/jreminder/pkg/config"
	"github.com/urfave/cli/v2"
)

var startCmd = &cli.Command{
	Name:   "start",
	Usage:  "start server",
	Action: Start,
}

func Start(ctx *cli.Context) error {
	path, err := config.PathRoot()
	if err != nil {
		return err
	}
	_, err = os.Stat(filepath.Join(path, config.DefaultName))
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("Please initialize config first")
		}
	} else {
		jreminder, err := app.NewJreminder()
		if err != nil {
			return err
		}
		var wg sync.WaitGroup
		wg.Add(1)
		handleShutdown(jreminder, &wg)
		err = jreminder.Start()
		if err != nil {
			return err
		}
		wg.Wait()
	}
	return nil
}

func handleShutdown(server *app.Jreminder, wg *sync.WaitGroup) {
	var stop = make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM)
	signal.Notify(stop, syscall.SIGINT)
	go func() {
		<-stop
		fmt.Println("received interrupt signal, shutting down...")
		server.Stop()
		wg.Done()
		os.Exit(0)
	}()
}
