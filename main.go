package main

import (
	"github.com/sudoStatus200/livefileshare/lib"
	cli "github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		lib.ReceiveCmd,
		lib.SendCmd,
	}

	app.RunAndExitOnError()
}
