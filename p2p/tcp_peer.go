package p2p

import (
	"fmt"
	"log"
	"net"
)

type Peer struct {
	conn       net.Conn
	listenAddr string
	outbound   bool
}

func (p *Peer) PeerReadLoop(msgch chan *Message) {
	fmt.Println("running peer read loop")
	defer p.conn.Close()

	buffer := make([]byte, 1024)
	for {
		n, err := p.conn.Read(buffer)
		if err != nil {
			log.Printf("read message error: %v", err)
			break
		}

		data := buffer[:n]
		msg := &Message{Payload: data, From: p.conn.RemoteAddr().String()}
		msgch <- msg

	}
}

func (p *Peer) Send(b []byte) (int, error) {
	return p.conn.Write(b)
}
