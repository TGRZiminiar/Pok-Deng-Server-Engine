package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net"
	"testing"

	"github.com/tgrziminiar/pok-deng-server-engine/p2p"
)

func TestPlayer(t *testing.T) {
	msg := &p2p.Message{
		From:    "hello ",
		Payload: p2p.CommandHelp{},
	}

	fmt.Println("Dialing to :3000")
	conn, err := net.Dial("tcp", "localhost:3000")
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(msg); err != nil {
		t.Fatal(err)
	}

	if _, err := conn.Write(buf.Bytes()); err != nil {
		t.Fatal(err)
	}
}
