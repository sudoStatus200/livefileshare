package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
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

	spkt, err := readControlPacket(control)

	if err != nil {
		return nil
	}

	if spkt.Event != EventStart {
		return fmt.Errorf("first event from control packet should be start")
	}

	select {}
	fmt.Println("now waiting on control end")

	spkt, err = readControlPacket(control)
	if err != nil {
		return fmt.Errorf("failed to read control end packet: %w", err)
	}

	fmt.Println("ending...")
	wg.Wait()
	fmt.Println("complete!")

	return nil

}

func handleReceivingFiles(root string, con net.Conn, wg *sync.WaitGroup) error {
	defer wg.Done()
	for {
		dec := json.NewDecoder(con)
		pkt, err := readControlPacket(con)
		if err != nil {
			return fmt.Errorf("failed to read control packet: %w", err)
		}

		r := io.MultiReader(dec.Buffered(), con)

		if pkt.Event != EventFile {
			fmt.Printf("%#v\n", pkt)
			return fmt.Errorf("got non-file event on data socket: %d", pkt.Event)
		}

		fileHandler := pkt.File

		switch {
		case fileHandler.IsDir:
			if err := os.Mkdir(filepath.Join(root, fileHandler.Path, fileHandler.Name), fileHandler.Perms); err != nil {
				return err
			}
		case fileHandler.IsLink:
			lval, err := ioutil.ReadAll(io.LimitReader(r, fileHandler.Size))
			if err != nil {
				return fmt.Errorf("reading link data: %w", err)
			}
			if err := os.Symlink(string(lval), filepath.Join(root, fileHandler.Path, fileHandler.Name)); err != nil {
				return err
			}
		default:
			if err := handleFileTransfer(r, root, fileHandler); err != nil {
				return fmt.Errorf("handling incoming file: %w", err)
			}

		}

	}
}

func handleFileTransfer(r io.Reader, root string, fh *fileHeader) error {
	fi, err := os.OpenFile(filepath.Join(root, fh.Path, fh.Name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, fh.Perms)
	if err != nil {
		return err
	}
	defer fi.Close()
	n, err := io.CopyN(fi, r, fh.Size)
	if err != nil {
		return err
	}

	if n != fh.Size {
		return fmt.Errorf("failed to copy the right amount of bytes: %d != %d", n, fh.Size)
	}

	return nil

}
