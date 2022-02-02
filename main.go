package main

import (
	"github.com/YOUSEEBIGGIRL/fakedocke/zlog"
	"github.com/urfave/cli/v2"
	"go.uber.org/zap"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "fakedocker"
	app.Usage = `fakedocker is a simple container runtime implementation.
	The purpose of this project is to learn how docker works and how to write a docker by ourselves
	Enjoy it, just for fun.`

	app.Commands = []*cli.Command{
		run,
		init_,
	}

	app.Before = func(context *cli.Context) error {
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		zlog.New().Error("fatal, the program exit.", zap.Error(err))
		os.Exit(-1)
	}
}
