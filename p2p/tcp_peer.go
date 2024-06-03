package p2p

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"net"
)

type Peer struct {
	conn       net.Conn
	listenAddr string
	outbound   bool
}

func (p *Peer) PeerReadLoop(msgch chan *Message) {

	fmt.Println("Running peer read loop")
	defer p.conn.Close()
	scanner := bufio.NewScanner(p.conn)

	for scanner.Scan() {
		text := scanner.Text()
		msg := &Message{
			Payload: text,
			From:    p.conn.RemoteAddr().String(),
		}
		fmt.Println("Sending message to channel:", msg)
		msgch <- msg
	}

	// ### for the encode and decode type with gob
	// fmt.Println("running peer read loop")
	// defer p.conn.Close()
	// decoder := gob.NewDecoder(p.conn)

	// for {
	// 	msg := new(Message)
	// 	if err := decoder.Decode(msg); err != nil {
	// 		log.Printf("decode message error: %v", err)
	// 		break
	// 	}
	// 	fmt.Println("sending message to channel -> ", msg)
	// 	msgch <- msg
	// }
}

func (p *Peer) Send(b []byte) (int, error) {
	return p.conn.Write(b)
}

func init() {
	gob.Register(CommandHelp{})
}
