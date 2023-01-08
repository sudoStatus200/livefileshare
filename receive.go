package main

import (
	"fmt"
	"net"
	"os"
	"sync"

	cli "github.com/urfave/cli/v2"
)

var receiveCmd = &cli.Command{
	Name: "receive",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "listen",
			Value: ":4900",
		},
	},
	Action: receiveAction,
}

func receiveAction(cctx *cli.Context) error {

	dir := cctx.Args().First()

	if dir == "" {
		curDir, err := os.Getwd()
		if err != nil {
			return err
		}
		dir = curDir
	}

	list, err := net.Listen("tcp", cctx.String("listen"))

	if err != nil {
		return err
	}

	// First connection is the control channel
	control, err := list.Accept()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	go func() {
		defer list.Close()
		for {
			con, err := list.Accept()
			if err != nil {
				fmt.Println("failed to accept new connection: ", err)
				return
			}

			fmt.Println("accepted  a new data connection")
			go func(cc net.Conn) {
				if err := handleReceivingFiles(dir, cc, &wg); err != nil {
					fmt.Println("handleReceivingFiles errored: ", err)
				}

			}(con)
		}
	}()

	return nil

}
