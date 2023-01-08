package main

import (
	cli "github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Commands = []*cli.Command{}

	app.RunAndExitOnError()
}
