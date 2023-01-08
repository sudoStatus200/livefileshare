package main

import (
	"os"
	"time"
)

const (
	EventStart = 1
	EventEnd   = 2
	EventFile  = 3
)

type fileHeader struct {
	Name    string
	Path    string
	Perms   os.FileMode
	ModTime time.Time
	IsDir   bool
	IsLink  bool
	Size    int64
}

type controlPacket struct {
	Event int
	File  *fileHeader
}
