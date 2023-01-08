package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"net"
)

func readControlPacket(con net.Conn) (*controlPacket, error) {
	msize := make([]byte, 4)

	_, err := io.ReadFull(con, msize)
	if err != nil {
		return nil, err
	}
	l := binary.BigEndian.Uint32(msize)
	rr := io.LimitReader(con, int64(l))

	var pkt controlPacket
	dec := json.NewDecoder(rr)
	if err := dec.Decode(&pkt); err != nil {
		return nil, err
	}

	return &pkt, nil

}
