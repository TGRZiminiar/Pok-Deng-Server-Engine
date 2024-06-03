package p2p

import (
	"log/slog"
	"sync"
)

type ServerConfig struct {
	ListenAddr string
}

type Server struct {
	ServerConfig
	tcpTransport *TCPTransport
	peerLock     sync.Mutex
	peers        map[string]*Peer
	addPeer      chan *Peer
	delPeer      chan *Peer
	msgCh        chan *Message
}

func NewServer(cfg ServerConfig) *Server {

	serverCfg := ServerConfig{
		ListenAddr: cfg.ListenAddr,
	}

	transport := NewTCPTransport(serverCfg.ListenAddr)

	return &Server{
		ServerConfig: serverCfg,
		peerLock:     sync.Mutex{},
		peers:        make(map[string]*Peer),
		addPeer:      make(chan *Peer),
		delPeer:      make(chan *Peer),
		tcpTransport: transport,
	}
}

// Start will listen to the new connection using tcp transprot
// and also start an infinite read loop to accept the channel of
// adding new peer/ delete peer/  message
func (s *Server) Start() error {

	if err := s.tcpTransport.ListenAndAccept(); err != nil {
		slog.Error("TCPTransport failed to listen and accpet", "error", err)
		return err
	}

	s.loop()

	return nil
}

func (s *Server) loop() {
	defer func() {
		if err := s.tcpTransport.Close(); err != nil {
			slog.Error("TCPTransport failed to close the connection", "error", err)
		}
	}()
	for {
		select {
		case peer := <-s.addPeer:
			_ = peer
		case msg := <-s.msgCh:
			_ = msg
		}
	}
}
