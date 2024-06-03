package p2p

import (
	"log/slog"
	"net"
)

type TCPTransport struct {
	listenAddr string
	listener   net.Listener
	AddPeer    chan *Peer
}

func NewTCPTransport(addr string) *TCPTransport {
	return &TCPTransport{
		listenAddr: addr,
	}
}

// make a server listen and accept the connection
// and send a channel signal to the read loop to make recieve a new peer
func (t *TCPTransport) ListenAndAccept() error {
	slog.Info("Server listening on port", "listenAddr", t.listenAddr)
	ln, err := net.Listen("tcp", t.listenAddr)
	if err != nil {
		return err
	}

	t.listener = ln

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Error("listener connection accept failed ", "conn", conn)
			continue
		}
		peer := &Peer{
			conn:     conn,
			outbound: false,
		}
		t.AddPeer <- peer
	}

}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}
