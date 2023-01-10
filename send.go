package main

import (
	"fmt"
	"io"
	"io/fs"
	"net"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"
)

var sendCmf = &cli.Command{
	Name: "send",
	Flags: []cli.Flag{
		&cli.IntFlag{
			Name:  "concurrency",
			Value: 1,
		},
	},
	Action: send,
}

func send(cctx *cli.Context) error {
	if !cctx.Args().Present() {
		return fmt.Errorf("must specify directory to send")
	}

	root := cctx.Args().First()

	dest := cctx.Args().Get(1)

	con, err := net.Dial("tcp", dest)

	if err != nil {
		return err
	}

	if err := writeControlPacket(con, &controlPacket{Event: EventStart}); err != nil {
		return err
	}

	datacon, err := net.Dial("tcp", dest)
	if err != nil {
		return err
	}

	if err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {

		if err != nil {
			return err
		}

		if filepath.Clean(root) == filepath.Clean(path) {
			return nil
		}
		fmt.Printf("sending: %q\n", path)

		rel, err := filepath.Rel(root, path)

		if err != nil {
			return err
		}
		name := filepath.Base(rel)
		dir := filepath.Dir(rel)

		info, err := d.Info()

		if err != nil {
			return err
		}

		islink := (info.Mode() & fs.ModeSymlink) > 0

		if err := writeControlPacket(datacon, &controlPacket{
			Event: EventFile,
			File: &fileHeader{
				Name:    name,
				Path:    dir,
				Perms:   info.Mode().Perm(),
				ModTime: info.ModTime(),
				IsDir:   d.IsDir(),
				IsLink:  islink,
				Size:    info.Size(),
			},
		}); err != nil {
			return err
		}

		if islink {
			return fmt.Errorf("cant handle symlinks right now")
		}

		if d.IsDir() {
			fmt.Println("is dir!")
			return nil
		}
		if err := sendFile(datacon, path, info.Size()); err != nil {
			return fmt.Errorf("sending file: %w", err)
		}

		return nil

	}); err != nil {
		return err
	}
	fmt.Println("terminating...")
	return nil
}

func sendFile(con net.Conn, path string, size int64) error {
	fi, err := os.Open(path)

	if err != nil {
		return err
	}
	defer fi.Close()

	n, err := io.CopyN(con, fi, size)

	if err != nil {
		return fmt.Errorf("sending file data: %w", err)
	}
	if n != size {
		return fmt.Errorf("expected %d bytes, got %d", size, n)
	}

	return nil

}
