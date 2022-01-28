package main

import (
	"github.com/YOUSEEBIGGIRL/fakedocke/command"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func init() {
	log.SetFlags(log.Lshortfile | log.Ltime)
}

func main() {
	app := cli.NewApp()
	app.Name = "fakedocker"
	app.Usage = `fakedocker is a simple container runtime implementation.
	The purpose of this project is to learn how docker works and how to write a docker by ourselves
	Enjoy it, just for fun.`

	app.Commands = []*cli.Command{
		command.Run(),
		command.Init(),
	}

	app.Before = func(context *cli.Context) error {
		return nil
	}

	//log.Println("pid: ", os.Getpid())
	//log.Println("os.Args: ", os.Args)
	if err := app.Run(os.Args); err != nil {
		log.Fatalf("[fatal] fatal: %v, the program exit.", err)
	}
}
